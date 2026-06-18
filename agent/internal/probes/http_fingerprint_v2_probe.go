package probes

import (
	"context"
	"crypto/tls"
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

type HTTPFingerprintV2Probe struct {
	funcs httpFingerprintV2Funcs
}

type httpFingerprintV2Funcs struct {
	gateway    func() (net.IP, error)
	traceroute func(context.Context, string, int) ([]string, error)
	fetch      func(context.Context, string) (httpFingerprintV2Result, error)
	tlsInfo    func(context.Context, string) (string, []string, string)
}

type httpFingerprintV2Result struct {
	StatusCode        int
	Title             string
	Server            string
	WWWAuthenticate   string
	WWWAuthRealm      string
	RedirectPath      string
	RedirectLocation  string
	FaviconHash       string
	HTMLMetaGenerator string
	LoginLabels       []string
	Body              string
}

func (HTTPFingerprintV2Probe) Name() string { return "http_fingerprint_v2" }

func (p HTTPFingerprintV2Probe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	f := p.withDefaults()
	candidates := p.candidates(ctx, in, f)
	res.Evidence["gateway_candidates"] = candidates
	if len(candidates) == 0 {
		res.Confidence = 0
		return res, nil
	}

	var devices []models.GatewayDevice
	var hints []string
	maxConf := 0.0
	for i, ip := range candidates {
		role := roleDefaultGateway
		if i > 0 {
			role = roleUpstreamGateway
		}
		dev := models.GatewayDevice{IP: ip, Role: role}
		var textParts []string
		for _, endpoint := range gatewayHTTPURLs(ip) {
			fp, err := f.fetch(ctx, endpoint)
			if err != nil {
				dev.FailedAttempts = append(dev.FailedAttempts, failedHTTPAttempt(p.Name(), ip, endpoint, http.MethodGet, err))
				continue
			}
			dev.Reachable = true
			dev.ReachableState = models.ReachableTrue
			if port := endpointPort(endpoint); port != 0 {
				dev.OpenPorts = appendUniqueInts(dev.OpenPorts, port)
			}
			dev.HTTPObservations = append(dev.HTTPObservations, httpObservationFromFingerprint(p.Name(), endpoint, http.MethodGet, fp))
			if dev.HTTPTitle == "" {
				dev.HTTPTitle = fp.Title
			}
			if dev.ServerHeader == "" {
				dev.ServerHeader = fp.Server
			}
			if dev.WWWAuthenticate == "" {
				dev.WWWAuthenticate = fp.WWWAuthenticate
			}
			if dev.WWWAuthRealm == "" {
				dev.WWWAuthRealm = fp.WWWAuthRealm
			}
			if dev.RedirectPath == "" {
				dev.RedirectPath = fp.RedirectPath
			}
			if dev.RedirectLocation == "" {
				dev.RedirectLocation = fp.RedirectLocation
			}
			if dev.FaviconHash == "" {
				dev.FaviconHash = fp.FaviconHash
			}
			if dev.HTMLMetaGenerator == "" {
				dev.HTMLMetaGenerator = fp.HTMLMetaGenerator
			}
			dev.LoginLabels = appendUniqueStrings(dev.LoginLabels, fp.LoginLabels...)
			textParts = append(textParts, fp.Title, fp.WWWAuthRealm, fp.HTMLMetaGenerator, strings.Join(fp.LoginLabels, " "))
			if !isGenericHTTPTitle(fp.Title) {
				textParts = append(textParts, fp.Body)
			}
			break
		}
		if dev.TLSCertCN == "" {
			cn, sans, issuer := f.tlsInfo(ctx, ip)
			dev.TLSCertCN = cn
			dev.TLSCertSANs = sans
			dev.TLSCertIssuer = issuer
			dev.TLSServerName = cn
			if cn != "" || len(sans) > 0 || issuer != "" {
				dev.TLSObservations = append(dev.TLSObservations, models.TLSObservation{
					Source: p.Name(), IP: ip, Port: 443, CN: cn, SANs: sans, Issuer: issuer, ServerName: cn,
					EvidenceID: p.Name() + ":" + ip + ":tls:443",
				})
			}
			textParts = append(textParts, cn, strings.Join(sans, " "))
		}
		allText := strings.Join(textParts, " ")
		dev.Manufacturer = inferManufacturer(allText)
		dev.Model = inferModel(dev.HTTPTitle)
		if dev.Model == "" {
			dev.Model = inferModel(dev.WWWAuthRealm)
		}
		dev.CPEModelGuess = strings.TrimSpace(strings.Join([]string{dev.Manufacturer, dev.Model}, " "))
		accessText := strings.Join([]string{dev.Manufacturer, dev.Model, allText}, " ")
		dev.AccessHints = inferAccessHints(accessText)
		dev.DeviceConfidence = gatewayDeviceConfidence(dev)
		dev.AccessConfidence = gatewayDeviceAccessConfidence(dev, accessText)
		dev.Confidence = dev.DeviceConfidence
		dev.EvidenceIDs = appendUniqueStrings(dev.EvidenceIDs, p.Name()+":"+ip)
		if dev.AccessConfidence > 0 {
			dev.AccessEvidence = append(dev.AccessEvidence, models.GatewayAccessEvidence{
				Source: p.Name(), Type: "device_text", Value: accessText, Strength: "medium",
				Confidence: dev.AccessConfidence, Hints: dev.AccessHints, EvidenceID: p.Name() + ":" + ip + ":access",
			})
			for _, h := range dev.AccessHints {
				hints = appendUnique(hints, h)
			}
		}
		maxConf = maxFloat(maxConf, dev.Confidence)
		devices = append(devices, dev)
	}
	res.Evidence["gateway_devices"] = devices
	res.Evidence["device_confidence"] = maxConf
	res.Evidence["access_confidence"] = maxDeviceAccessConfidence(devices)
	res.Evidence["strong_access_evidence"] = len(hints) > 0
	res.Hints = hints
	res.Confidence = maxConf
	return res, nil
}

func (p HTTPFingerprintV2Probe) withDefaults() httpFingerprintV2Funcs {
	f := p.funcs
	if f.gateway == nil {
		f.gateway = network.Gateway
	}
	if f.traceroute == nil {
		f.traceroute = system.Traceroute
	}
	if f.fetch == nil {
		f.fetch = fetchHTTPFingerprintV2
	}
	if f.tlsInfo == nil {
		f.tlsInfo = fetchTLSInfo
	}
	return f
}

func (p HTTPFingerprintV2Probe) candidates(ctx context.Context, in models.ScanInput, f httpFingerprintV2Funcs) []string {
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

func fetchHTTPFingerprintV2(ctx context.Context, endpoint string) (httpFingerprintV2Result, error) {
	client := &http.Client{
		Timeout:   2 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return httpFingerprintV2Result{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return httpFingerprintV2Result{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return httpFingerprintV2Result{}, err
	}
	text := string(body)
	authHeader := resp.Header.Get("WWW-Authenticate")
	redirectLocation := resp.Header.Get("Location")
	return httpFingerprintV2Result{
		Title:             extractTitle(text),
		StatusCode:        resp.StatusCode,
		Server:            resp.Header.Get("Server"),
		WWWAuthenticate:   authHeader,
		WWWAuthRealm:      parseAuthRealm(authHeader),
		RedirectPath:      redirectPath(resp),
		RedirectLocation:  redirectLocation,
		FaviconHash:       fetchFaviconHash(ctx, client, endpoint),
		HTMLMetaGenerator: extractMetaGenerator(text),
		LoginLabels:       extractLoginLabels(text),
		Body:              text,
	}, nil
}

func fetchTLSInfo(ctx context.Context, ip string) (string, []string, string) {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 1200 * time.Millisecond},
		Config:    &tls.Config{InsecureSkipVerify: true, ServerName: ip},
	}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, "443"))
	if err != nil {
		return "", nil, ""
	}
	defer conn.Close()
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return "", nil, ""
	}
	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return "", nil, ""
	}
	cert := state.PeerCertificates[0]
	return cert.Subject.CommonName, append([]string{}, cert.DNSNames...), cert.Issuer.String()
}

func parseAuthRealm(header string) string {
	l := strings.ToLower(header)
	idx := strings.Index(l, "realm=")
	if idx < 0 {
		return ""
	}
	realm := strings.TrimSpace(header[idx+6:])
	realm = strings.Trim(realm, `"`)
	if comma := strings.Index(realm, ","); comma >= 0 {
		realm = strings.TrimSpace(realm[:comma])
	}
	return realm
}

func redirectPath(resp *http.Response) string {
	loc := resp.Header.Get("Location")
	if loc == "" {
		return ""
	}
	u, err := url.Parse(loc)
	if err != nil {
		return loc
	}
	if u.Path == "" {
		return loc
	}
	return u.Path
}

var metaGeneratorRe = regexp.MustCompile(`(?is)<meta[^>]+name=["']generator["'][^>]+content=["']([^"']+)["']`)
var labelRe = regexp.MustCompile(`(?is)<label[^>]*>(.*?)</label>`)

func extractMetaGenerator(body string) string {
	m := metaGeneratorRe.FindStringSubmatch(body)
	if len(m) < 2 {
		return ""
	}
	return strings.Join(strings.Fields(htmlEntityReplacer.Replace(m[1])), " ")
}

func extractLoginLabels(body string) []string {
	var labels []string
	for _, m := range labelRe.FindAllStringSubmatch(body, 8) {
		if len(m) < 2 {
			continue
		}
		label := strings.Join(strings.Fields(htmlEntityReplacer.Replace(m[1])), " ")
		l := strings.ToLower(label)
		if strings.Contains(l, "user") || strings.Contains(l, "password") || strings.Contains(l, "login") {
			labels = appendUniqueStrings(labels, label)
		}
	}
	return labels
}

func maxDeviceAccessConfidence(devices []models.GatewayDevice) float64 {
	max := 0.0
	for _, d := range devices {
		max = maxFloat(max, d.AccessConfidence)
	}
	return max
}

func appendUniqueStrings(s []string, values ...string) []string {
	for _, v := range values {
		if v == "" {
			continue
		}
		found := false
		for _, e := range s {
			if e == v {
				found = true
				break
			}
		}
		if !found {
			s = append(s, v)
		}
	}
	return s
}

func httpObservationFromFingerprint(source, endpoint, method string, fp httpFingerprintV2Result) models.HTTPObservation {
	return models.HTTPObservation{
		Source: source, URL: endpoint, Method: method, StatusCode: fp.StatusCode,
		Title: fp.Title, ServerHeader: fp.Server, WWWAuthenticate: fp.WWWAuthenticate,
		WWWAuthRealm: fp.WWWAuthRealm, RedirectPath: fp.RedirectPath, RedirectLocation: fp.RedirectLocation,
		FaviconHash: fp.FaviconHash, HTMLMetaGenerator: fp.HTMLMetaGenerator, LoginLabels: fp.LoginLabels,
		EvidenceID: source + ":" + endpoint,
	}
}

func failedHTTPAttempt(source, ip, endpoint, method string, err error) models.ProbeAttempt {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return models.ProbeAttempt{
		Source: source, Target: ip, Protocol: "http", URL: endpoint, Method: method,
		Error: msg, Timeout: strings.Contains(strings.ToLower(msg), "timeout"),
		EvidenceID: source + ":" + endpoint + ":failed",
	}
}

func endpointPort(endpoint string) int {
	u, err := url.Parse(endpoint)
	if err != nil {
		return 0
	}
	if port := u.Port(); port != "" {
		switch port {
		case "80":
			return 80
		case "443":
			return 443
		case "8080":
			return 8080
		case "8443":
			return 8443
		case "7547":
			return 7547
		}
	}
	if u.Scheme == "https" {
		return 443
	}
	if u.Scheme == "http" {
		return 80
	}
	return 0
}

func appendUniqueInts(s []int, values ...int) []int {
	for _, v := range values {
		if v == 0 {
			continue
		}
		found := false
		for _, e := range s {
			if e == v {
				found = true
				break
			}
		}
		if !found {
			s = append(s, v)
		}
	}
	return s
}
