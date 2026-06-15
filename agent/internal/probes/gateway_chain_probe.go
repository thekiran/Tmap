package probes

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/internal/system"
	"github.com/thekiran/iad/pkg/models"
)

const (
	roleDefaultGateway  = "default_gateway"
	roleUpstreamGateway = "upstream_private_gateway"
	rolePossibleModem   = "possible_modem"
)

// GatewayChainProbe inspects only already observed RFC1918 gateway IPs. It does
// not scan subnets or public addresses; deep+online mode adds a short traceroute
// solely to discover leading private hops in the local router chain.
type GatewayChainProbe struct {
	funcs gatewayChainFuncs
}

type gatewayChainFuncs struct {
	gateway     func() (net.IP, error)
	traceroute  func(context.Context, string, int) ([]string, error)
	httpGet     func(context.Context, string) (gatewayHTTPResult, error)
	checkTCP    func(context.Context, string, string) bool
	ssdpSearch  func(context.Context, time.Duration) ([]string, error)
	fetchDevice func(context.Context, string) (*upnpDevice, error)
}

type gatewayHTTPResult struct {
	Title  string
	Server string
	Body   string
	FaviconHash string
}

func (GatewayChainProbe) Name() string { return "gateway_chain_probe" }

func (p GatewayChainProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	f := p.withDefaults()

	candidates := p.candidates(ctx, in, f)
	res.Evidence["gateway_chain"] = candidates
	res.Evidence["double_nat_possible"] = len(candidates) >= 2
	if len(candidates) == 0 {
		res.Confidence = 0
		return res, nil
	}

	locations := p.ssdpLocations(ctx, f)
	devices := make([]models.GatewayDevice, 0, len(candidates))
	for i, ip := range candidates {
		role := roleDefaultGateway
		if i > 0 {
			role = roleUpstreamGateway
		}
		dev := p.discoverDevice(ctx, f, ip, role, locations)
		devices = append(devices, dev)
	}

	likely := likelyModemIP(devices)
	if likely != "" {
		for i := range devices {
			if devices[i].IP == likely {
				devices[i].Role = rolePossibleModem
				break
			}
		}
		res.Evidence["likely_modem_ip"] = likely
	}

	var hints []string
	maxConf := 0.0
	for _, d := range devices {
		for _, h := range d.AccessHints {
			hints = appendUnique(hints, h)
		}
		if d.Confidence > maxConf {
			maxConf = d.Confidence
		}
	}
	res.Evidence["gateway_devices"] = devices
	res.Hints = hints
	res.Confidence = maxConf
	return res, nil
}

func (p GatewayChainProbe) withDefaults() gatewayChainFuncs {
	f := p.funcs
	if f.gateway == nil {
		f.gateway = network.Gateway
	}
	if f.traceroute == nil {
		f.traceroute = system.Traceroute
	}
	if f.httpGet == nil {
		f.httpGet = fetchGatewayHTTP
	}
	if f.checkTCP == nil {
		f.checkTCP = checkTCPPort
	}
	if f.ssdpSearch == nil {
		f.ssdpSearch = ssdpSearch
	}
	if f.fetchDevice == nil {
		f.fetchDevice = func(ctx context.Context, loc string) (*upnpDevice, error) {
			client := &http.Client{Timeout: 2 * time.Second}
			return fetchDescription(ctx, client, loc)
		}
	}
	return f
}

func (p GatewayChainProbe) candidates(ctx context.Context, in models.ScanInput, f gatewayChainFuncs) []string {
	var out []string
	if gw, err := f.gateway(); err == nil && isRFC1918IPv4(gw.String()) {
		out = appendUnique(out, gw.String())
	}
	if in.Mode == models.ModeDeep && in.Online {
		if hops, err := f.traceroute(ctx, traceTarget, 6); err == nil {
			for _, h := range leadingPrivateHops(hops) {
				out = appendUnique(out, h)
			}
		}
	}
	return out
}

func leadingPrivateHops(hops []string) []string {
	var out []string
	for _, h := range hops {
		if h == "*" || h == "" {
			continue
		}
		if !isRFC1918IPv4(h) {
			break
		}
		out = appendUnique(out, h)
	}
	return out
}

func (p GatewayChainProbe) ssdpLocations(ctx context.Context, f gatewayChainFuncs) []string {
	locations, err := f.ssdpSearch(ctx, 1500*time.Millisecond)
	if err != nil {
		return nil
	}
	return locations
}

func (p GatewayChainProbe) discoverDevice(ctx context.Context, f gatewayChainFuncs, ip, role string, locations []string) models.GatewayDevice {
	dev := models.GatewayDevice{IP: ip, Role: role}
	var textParts []string
	var accessParts []string

	for _, endpoint := range gatewayHTTPURLs(ip) {
		hr, err := f.httpGet(ctx, endpoint)
		if err != nil {
			continue
		}
		dev.Reachable = true
		if dev.HTTPTitle == "" {
			dev.HTTPTitle = hr.Title
		}
		if dev.ServerHeader == "" {
			dev.ServerHeader = hr.Server
		}
		if dev.FaviconHash == "" {
			dev.FaviconHash = hr.FaviconHash
		}
		textParts = append(textParts, hr.Title, hr.Server, hr.Body)
		accessParts = append(accessParts, accessRelevantHTTPText(hr)...)
		if dev.HTTPTitle != "" && dev.ServerHeader != "" {
			break
		}
	}

	for _, port := range []string{"80", "443", "8080", "7547", "49000"} {
		if f.checkTCP(ctx, ip, port) {
			dev.Reachable = true
			if port == "49000" {
				dev.TR064Found = true
			}
		}
	}

	if d := p.upnpForIP(ctx, f, ip, locations); d != nil {
		dev.Reachable = true
		dev.UPnPFound = true
		dev.Manufacturer = strings.TrimSpace(d.Manufacturer)
		dev.Model = strings.TrimSpace(d.ModelName + " " + d.ModelNumber)
		textParts = append(textParts, d.Manufacturer, d.ModelName, d.ModelNumber, d.ModelDescription, d.FriendlyName, d.DeviceType)
		accessParts = append(accessParts, d.Manufacturer, d.ModelName, d.ModelNumber, d.ModelDescription, d.FriendlyName, d.DeviceType)
	}

	if dev.TR064Found {
		for _, path := range []string{"/tr64desc.xml", "/igddesc.xml"} {
			if d, err := f.fetchDevice(ctx, "http://"+net.JoinHostPort(ip, "49000")+path); err == nil && d != nil {
				if dev.Manufacturer == "" {
					dev.Manufacturer = strings.TrimSpace(d.Manufacturer)
				}
				if dev.Model == "" {
					dev.Model = strings.TrimSpace(d.ModelName + " " + d.ModelNumber)
				}
				textParts = append(textParts, d.Manufacturer, d.ModelName, d.ModelNumber, d.ModelDescription, d.FriendlyName, d.DeviceType)
				accessParts = append(accessParts, d.Manufacturer, d.ModelName, d.ModelNumber, d.ModelDescription, d.FriendlyName, d.DeviceType)
				break
			}
		}
	}

	allText := strings.Join(textParts, " ")
	if dev.Manufacturer == "" {
		dev.Manufacturer = inferManufacturer(allText)
	}
	if dev.Model == "" {
		dev.Model = inferModel(dev.HTTPTitle)
	}
	accessText := strings.Join(accessParts, " ") + " " + dev.Manufacturer + " " + dev.Model
	dev.AccessHints = inferAccessHints(accessText)
	dev.DeviceConfidence = gatewayDeviceConfidence(dev)
	dev.AccessConfidence = gatewayDeviceAccessConfidence(dev, accessText)
	dev.Confidence = dev.DeviceConfidence
	return dev
}

func (p GatewayChainProbe) upnpForIP(ctx context.Context, f gatewayChainFuncs, ip string, locations []string) *upnpDevice {
	for _, loc := range locations {
		u, err := url.Parse(loc)
		if err != nil || u.Hostname() != ip {
			continue
		}
		dev, err := f.fetchDevice(ctx, loc)
		if err == nil && dev != nil {
			return dev
		}
	}
	return nil
}

func gatewayHTTPURLs(ip string) []string {
	return []string{
		"http://" + ip + "/",
		"http://" + net.JoinHostPort(ip, "80") + "/",
		"http://" + net.JoinHostPort(ip, "8080") + "/",
		"https://" + ip + "/",
	}
}

func fetchGatewayHTTP(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // local router UIs often use self-signed certs
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return gatewayHTTPResult{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return gatewayHTTPResult{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return gatewayHTTPResult{}, err
	}
	text := string(body)
	return gatewayHTTPResult{
		Title:  extractTitle(text),
		Server: resp.Header.Get("Server"),
		Body:   text,
		FaviconHash: fetchFaviconHash(ctx, client, endpoint),
	}, nil
}

func fetchFaviconHash(ctx context.Context, client *http.Client, endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	u.Path = "/favicon.ico"
	u.RawQuery = ""
	u.Fragment = ""
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ""
	}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil || len(body) == 0 {
		return ""
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:8])
}

func checkTCPPort(ctx context.Context, ip, port string) bool {
	d := net.Dialer{Timeout: 800 * time.Millisecond}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

var titleRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

func extractTitle(body string) string {
	m := titleRe.FindStringSubmatch(body)
	if len(m) < 2 {
		return ""
	}
	return strings.Join(strings.Fields(htmlEntityReplacer.Replace(m[1])), " ")
}

var htmlEntityReplacer = strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&#45;", "-")

func inferManufacturer(text string) string {
	l := strings.ToLower(text)
	for _, c := range []struct {
		token string
		name  string
	}{
		{"tp-link", "TP-Link"},
		{"tplink", "TP-Link"},
		{"zyxel", "Zyxel"},
		{"zte", "ZTE"},
		{"huawei", "Huawei"},
		{"keenetic", "Keenetic"},
		{"xiaomi", "Xiaomi"},
		{"mi router", "Xiaomi"},
		{"mikrotik", "MikroTik"},
		{"arris", "Arris"},
		{"technicolor", "Technicolor"},
	} {
		if strings.Contains(l, c.token) {
			return c.name
		}
	}
	return ""
}

func inferModel(title string) string {
	if isGenericHTTPTitle(title) {
		return ""
	}
	fields := strings.Fields(title)
	for _, f := range fields {
		if hasDigit(f) && len(f) >= 4 {
			return strings.Trim(f, " -_|")
		}
	}
	return strings.TrimSpace(title)
}

func hasDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func inferAccessHints(text string) []string {
	l := scrubGenericHTTPText(strings.ToLower(text))
	var hints []string
	for _, c := range []struct {
		tokens []string
		hint   string
	}{
		{[]string{"vdsl", "vdsl2", "ptm", "wan dsl", "zyxel vmg", "vmg3312", "vmg3625", "tp-link td-", "keenetic dsl", "zte h168a", "zxhn h168"}, models.TypeVDSL},
		{[]string{"adsl", "adsl2+", "atm"}, models.TypeADSL},
		{[]string{" dsl ", "snr", "attenuation", "line rate"}, models.TypeDSL},
		{[]string{"gpon", "epon", "ont", "onu", "ftth", "wan gpon", "fiberhome", "echolife ont", "nokia ont", "zte f660", "zte f601", "zte f6"}, models.TypeFiber},
		{[]string{"gpon"}, models.TypeGPON},
		{[]string{"docsis", "cable modem", "eurodocsis", "arris", "technicolor"}, models.TypeCable},
		{[]string{"lte", "4g", "5g", "nr5g", "cpe lte", "wwan", "cellular"}, models.TypeMobile},
	} {
		for _, token := range c.tokens {
			if strings.Contains(l, token) {
				hints = appendUnique(hints, c.hint)
				break
			}
		}
	}
	return hints
}

func accessRelevantHTTPText(hr gatewayHTTPResult) []string {
	var parts []string
	if !isGenericHTTPTitle(hr.Title) {
		parts = append(parts, hr.Title)
	}
	body := strings.ToLower(hr.Body)
	for _, token := range []string{"snr", "attenuation", "line rate", "wan dsl", "wan gpon"} {
		if strings.Contains(body, token) {
			parts = append(parts, hr.Body)
			break
		}
	}
	return parts
}

func scrubGenericHTTPText(text string) string {
	text = " " + strings.Join(strings.Fields(text), " ") + " "
	for _, token := range []string{
		" nginx ", " apache ", " microsoft-iis ", " lighttpd ", " caddy ", " openresty ",
		"generic http 200", "generic login page", "generic router login",
		"router login", "login page", "web interface",
	} {
		text = strings.ReplaceAll(text, token, " ")
	}
	return " " + strings.Join(strings.Fields(text), " ") + " "
}

func isGenericHTTPTitle(title string) bool {
	t := strings.TrimSpace(strings.ToLower(title))
	if t == "" {
		return true
	}
	for _, token := range []string{
		"router login", "login page", "web interface",
	} {
		if strings.Contains(t, token) {
			return true
		}
	}
	for _, exact := range []string{"router", "mi router", "gateway", "home gateway", "index", "welcome"} {
		if t == exact {
			return true
		}
	}
	return false
}

func gatewayDeviceConfidence(d models.GatewayDevice) float64 {
	conf := 0.0
	if d.Reachable {
		conf += 0.25
	}
	if d.HTTPTitle != "" || d.ServerHeader != "" {
		conf += 0.15
	}
	if d.FaviconHash != "" {
		conf += 0.10
	}
	if d.UPnPFound || d.TR064Found {
		conf += 0.25
	}
	if d.Manufacturer != "" || d.Model != "" {
		conf += 0.15
	}
	if len(d.AccessHints) > 0 {
		conf += 0.20
	}
	if conf > 1 {
		return 1
	}
	return conf
}

func gatewayDeviceAccessConfidence(d models.GatewayDevice, accessText string) float64 {
	if len(d.AccessHints) == 0 {
		return 0
	}
	conf := 0.35
	if d.UPnPFound || d.TR064Found {
		conf += 0.25
	}
	if d.Model != "" || d.Manufacturer != "" {
		conf += 0.20
	}
	if containsStrongAccessToken(accessText) {
		conf += 0.20
	}
	if conf > 1 {
		return 1
	}
	return conf
}

func containsStrongAccessToken(text string) bool {
	return len(inferAccessHints(text)) > 0
}

func likelyModemIP(devices []models.GatewayDevice) string {
	bestIdx := -1
	bestScore := -1.0
	hasUpstream := false
	for _, d := range devices {
		if d.Role == roleUpstreamGateway {
			hasUpstream = true
			break
		}
	}
	for i, d := range devices {
		hasAccessIdentity := d.AccessConfidence > 0 || len(d.AccessHints) > 0
		hasRouterIdentity := (d.UPnPFound || d.TR064Found) && (d.Model != "" || d.Manufacturer != "") && hasAccessIdentity
		if !hasAccessIdentity && !hasRouterIdentity {
			continue
		}
		if hasUpstream && d.Role == roleDefaultGateway && !hasAccessIdentity {
			continue
		}
		score := d.AccessConfidence
		if d.Role == roleUpstreamGateway {
			score += 0.10
		}
		if len(d.AccessHints) > 0 {
			score += 0.20
		}
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return ""
	}
	return devices[bestIdx].IP
}

func isRFC1918IPv4(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	v4 := parsed.To4()
	if v4 == nil {
		return false
	}
	switch {
	case v4[0] == 10:
		return true
	case v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31:
		return true
	case v4[0] == 192 && v4[1] == 168:
		return true
	default:
		return false
	}
}
