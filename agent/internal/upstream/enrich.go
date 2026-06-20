package upstream

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/thekiran/iad/internal/safety"
	"github.com/thekiran/iad/pkg/models"
)

// safeUpstreamPorts are the management/service ports probed with a plain TCP
// connect (open/closed only — no login, no banner grab beyond HTTP/TLS). 161
// (SNMP) is intentionally excluded; it stays opt-in elsewhere.
var safeUpstreamPorts = []int{80, 443, 8080, 8443, 7547, 53, 22, 23}

// Options tunes the enrichment run. Zero values fall back to safe defaults.
type Options struct {
	PerProbeTimeout time.Duration
	EnablePing      bool
	MaxConcurrency  int
	Now             func() time.Time
}

func (o Options) withDefaults() Options {
	if o.PerProbeTimeout <= 0 {
		o.PerProbeTimeout = 2 * time.Second
	}
	if o.MaxConcurrency <= 0 {
		o.MaxConcurrency = 3
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	return o
}

// IntelResult is the enrichment outcome for a single upstream device.
type IntelResult struct {
	IP             string
	Reachability   models.DeviceReachability
	Classification Classification
	Routing        models.RoutingEvidence
	Warnings       []string
	Evidence       []models.Evidence
}

type target struct {
	ip               string
	isDefaultGateway bool
	inGatewayChain   bool
	hopIndex         int
	hopDistance      int
	doubleNATHint    bool
	macVendor        string
	hasUPnPRoot      bool
}

// EnrichReport runs the upstream enrichment phase over every gateway/upstream/
// CPE candidate it can find in the report. It is fully read-only and only ever
// probes private/local targets. A failure on one probe or target is recorded as
// a warning and never aborts the scan.
func EnrichReport(ctx context.Context, report models.ScanReport, opts Options) map[string]IntelResult {
	opts = opts.withDefaults()
	targets := collectTargets(report)
	if len(targets) == 0 {
		return nil
	}
	subnet := agentSubnet(report)

	results := make(map[string]IntelResult, len(targets))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, opts.MaxConcurrency)

	for _, t := range targets {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			res := enrichOne(ctx, t, subnet, opts)
			mu.Lock()
			results[t.ip] = res
			mu.Unlock()
		}()
	}
	wg.Wait()
	return results
}

func enrichOne(ctx context.Context, t target, subnet *net.IPNet, opts Options) IntelResult {
	now := opts.Now().UTC()
	res := IntelResult{IP: t.ip}
	idx := 0
	evID := func(kind string) string {
		idx++
		return fmt.Sprintf("up-ev-%s-%s-%d", kind, strings.NewReplacer(".", "-", ":", "-").Replace(t.ip), idx)
	}
	addEv := func(kind string, data map[string]string) {
		res.Evidence = append(res.Evidence, models.Evidence{ID: evID(kind), Kind: kind, Source: "upstream_enrichment", Data: data, Timestamp: now})
	}

	sameSubnet := subnet != nil && subnet.Contains(net.ParseIP(t.ip))

	// 1) Reachability — short system ping, falling back to TCP-connect timing.
	reach := models.DeviceReachability{DirectlyReachable: sameSubnet, Method: "none"}
	if opts.EnablePing {
		pctx, cancel := context.WithTimeout(ctx, opts.PerProbeTimeout+1500*time.Millisecond)
		pr, err := Ping(pctx, t.ip)
		cancel()
		if err != nil && !pr.Reachable {
			res.Warnings = append(res.Warnings, fmt.Sprintf("ICMP ping to %s failed or was blocked: %v — falling back to TCP reachability.", t.ip, err))
		}
		if pr.Reachable {
			reach.ICMP = true
			reach.Method = "icmp_ping"
			reach.AvgLatencyMs, reach.MinLatencyMs, reach.MaxLatencyMs = pr.AvgMs, pr.MinMs, pr.MaxMs
			reach.TTL, reach.PacketLoss = pr.TTL, pr.LossPct
			status := "reply"
			addEv("icmp_echo", pruneEmpty(map[string]string{
				"ip": t.ip, "status": status, "ttl": itoaPtr(pr.TTL), "latency_ms": ftoaPtr(pr.AvgMs), "packet_loss": ftoaPtr(pr.LossPct),
			}))
		} else {
			addEv("icmp_echo", map[string]string{"ip": t.ip, "status": "timeout"})
		}
	}

	// 2) Safe TCP connect sweep (open/closed only).
	var openPorts []int
	var bestLatency *float64
	for _, port := range safeUpstreamPorts {
		open, latency := tcpConnect(ctx, t.ip, port, opts.PerProbeTimeout)
		if !open {
			continue
		}
		openPorts = append(openPorts, port)
		if bestLatency == nil || latency < *bestLatency {
			l := latency
			bestLatency = &l
		}
	}
	sort.Ints(openPorts)
	if len(openPorts) > 0 {
		reach.TCPReachable = true
		addEv("tcp_connect", map[string]string{"ip": t.ip, "ports": joinInts(openPorts)})
		// If ICMP didn't answer, derive latency/method from TCP timing.
		if !reach.ICMP {
			reach.Method = "tcp_connect"
			reach.AvgLatencyMs = bestLatency
		}
	}

	// 3) HTTP/HTTPS fingerprint on any open web port.
	var vendor VendorResult
	for _, port := range []int{80, 8080, 443, 8443} {
		if !containsInt(openPorts, port) {
			continue
		}
		scheme := "http"
		if port == 443 || port == 8443 {
			scheme = "https"
		}
		obs, err := httpFingerprint(ctx, t.ip, port, scheme, opts.PerProbeTimeout)
		if err != nil || obs == nil {
			continue
		}
		v := AnalyzeHTTP(obs.ServerHeader, obs.Title, obs.WWWAuthRealm, obs.Body)
		if obs.StatusCode == 401 || obs.WWWAuthenticate != "" {
			v.AdminPanel = true
		}
		vendor = mergeVendor(vendor, v)
		addEv("http_fingerprint", pruneEmpty(map[string]string{
			"ip": t.ip, "url": obs.URL, "status_code": fmt.Sprintf("%d", obs.StatusCode),
			"server": obs.ServerHeader, "title": obs.Title, "www_authenticate": obs.WWWAuthenticate,
			"realm": obs.WWWAuthRealm, "location": obs.RedirectLocation,
		}))
	}

	// 4) TLS fingerprint on HTTPS ports.
	for _, port := range []int{443, 8443} {
		if !containsInt(openPorts, port) {
			continue
		}
		tobs, err := tlsFingerprint(ctx, t.ip, port, opts.PerProbeTimeout)
		if err != nil || tobs == nil {
			continue
		}
		if v := AnalyzeHTTP("", tobs.CN+" "+strings.Join(tobs.SANs, " "), "", ""); v.Vendor != "" {
			vendor = mergeVendor(vendor, v)
		}
		addEv("tls_fingerprint", pruneEmpty(map[string]string{
			"ip": t.ip, "port": fmt.Sprintf("%d", port), "cn": tobs.CN,
			"issuer": tobs.Issuer, "sans": strings.Join(tobs.SANs, ","),
		}))
	}

	// 5) Assemble facts and classify.
	f := Facts{
		IP:                t.ip,
		IsPrivate:         safety.IsPrivateIPString(t.ip),
		SameSubnetAsAgent: sameSubnet,
		IsDefaultGateway:  t.isDefaultGateway,
		InGatewayChain:    t.inGatewayChain,
		HopIndex:          t.hopIndex,
		HopDistance:       t.hopDistance,
		DoubleNATHint:     t.doubleNATHint,
		ReachableICMP:     reach.ICMP,
		ReachableTCP:      reach.TCPReachable,
		OpenPorts:         openPorts,
		HTTPVendor:        vendor.Vendor,
		HTTPAdminPanel:    vendor.AdminPanel,
		RouterLikeHTTP:    vendor.RouterLike,
		HasUPnPRoot:       t.hasUPnPRoot,
		HasDNSService:     containsInt(openPorts, 53),
		HasCWMP:           containsInt(openPorts, 7547),
		MACVendor:         t.macVendor,
		ONTHint:           vendor.ONT,
		ModemHint:         vendor.Modem,
		Now:               now,
	}
	if reach.HopDistance == nil && t.hopDistance > 0 {
		hd := t.hopDistance
		reach.HopDistance = &hd
	}
	res.Classification = Classify(f)
	res.Routing = AnalyzeRouting(f)
	res.Reachability = reach
	res.Warnings = append(res.Warnings, res.Classification.Warnings...)
	return res
}

// collectTargets gathers private gateway/upstream/CPE IPs from the report.
func collectTargets(report models.ScanReport) []target {
	seen := map[string]*target{}
	add := func(ip string, mutate func(*target)) {
		ip = strings.TrimSpace(ip)
		if ip == "" || !safety.IsPrivateIPString(ip) {
			return
		}
		t := seen[ip]
		if t == nil {
			t = &target{ip: ip}
			seen[ip] = t
		}
		if mutate != nil {
			mutate(t)
		}
	}

	defaultGW := report.Agent.Gateway
	ssdpByIP := ssdpTargets(report)

	if report.AccessClassification != nil && report.AccessClassification.DetectedNetworkContext != nil {
		ctx := report.AccessClassification.DetectedNetworkContext
		if defaultGW == "" {
			defaultGW = ctx.Gateway
		}
		if st := ctx.GatewayChainState; st != nil {
			if defaultGW == "" {
				defaultGW = st.DefaultGateway
			}
			for _, hop := range st.PrivateHops {
				h := hop
				add(h.IP, func(t *target) {
					t.inGatewayChain = true
					t.hopIndex = h.Order
					t.hopDistance = h.Order
					t.doubleNATHint = st.InternalDoubleNATPossible
					if h.Role == "default_gateway" || h.IP == defaultGW {
						t.isDefaultGateway = true
					}
				})
			}
		}
		for _, gd := range ctx.GatewayDevices {
			g := gd
			add(g.IP, func(t *target) {
				if g.MACVendor != "" {
					t.macVendor = g.MACVendor
				}
				if g.UPnPFound || g.UPnPIGDFound {
					t.hasUPnPRoot = true
				}
				if g.Role == "default_gateway" || g.IP == defaultGW {
					t.isDefaultGateway = true
				}
			})
		}
	}
	if defaultGW != "" {
		add(defaultGW, func(t *target) { t.isDefaultGateway = true })
	}
	for ip := range ssdpByIP {
		add(ip, func(t *target) { t.hasUPnPRoot = true })
	}

	out := make([]target, 0, len(seen))
	for _, t := range seen {
		out = append(out, *t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ip < out[j].ip })
	return out
}

func ssdpTargets(report models.ScanReport) map[string]bool {
	out := map[string]bool{}
	for _, ev := range report.Evidence {
		if strings.EqualFold(ev.Kind, "ssdp") {
			if ip := strings.TrimSpace(ev.Data["ip"]); ip != "" {
				out[ip] = true
			}
		}
	}
	return out
}

func agentSubnet(report models.ScanReport) *net.IPNet {
	if cidr := strings.TrimSpace(report.Scope.CIDR); cidr != "" {
		if _, n, err := net.ParseCIDR(cidr); err == nil {
			return n
		}
	}
	return nil
}

// --- safe native probes ---

func tcpConnect(ctx context.Context, ip string, port int, timeout time.Duration) (bool, float64) {
	dctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var d net.Dialer
	start := time.Now()
	conn, err := d.DialContext(dctx, "tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)))
	if err != nil {
		return false, 0
	}
	latency := float64(time.Since(start).Microseconds()) / 1000.0
	_ = conn.Close()
	return true, latency
}

type httpObs struct {
	URL, Method, ServerHeader, WWWAuthenticate, WWWAuthRealm, RedirectLocation, Title, Body string
	StatusCode                                                                              int
}

func httpFingerprint(ctx context.Context, ip string, port int, scheme string, timeout time.Duration) (*httpObs, error) {
	client := &http.Client{
		Timeout: timeout,
		// Do NOT follow redirects — capture the Location instead (no risky hops).
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport:     &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	url := fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(ip, fmt.Sprintf("%d", port)))
	method := http.MethodHead
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode == http.StatusMethodNotAllowed {
		method = http.MethodGet
		if req2, e := http.NewRequestWithContext(ctx, method, url, nil); e == nil {
			if resp != nil {
				resp.Body.Close()
			}
			resp, err = client.Do(req2)
		}
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	obs := &httpObs{
		URL: url, Method: resp.Request.Method, StatusCode: resp.StatusCode,
		ServerHeader: resp.Header.Get("Server"), WWWAuthenticate: resp.Header.Get("WWW-Authenticate"),
		WWWAuthRealm: authRealm(resp.Header.Get("WWW-Authenticate")), RedirectLocation: resp.Header.Get("Location"),
		Title: htmlTitle(string(body)), Body: string(body),
	}
	return obs, nil
}

type tlsObs struct {
	CN, Issuer string
	SANs       []string
}

func tlsFingerprint(ctx context.Context, ip string, port int, timeout time.Duration) (*tlsObs, error) {
	dctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var d net.Dialer
	raw, err := d.DialContext(dctx, "tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)))
	if err != nil {
		return nil, err
	}
	conn := tls.Client(raw, &tls.Config{InsecureSkipVerify: true})
	if err := conn.HandshakeContext(dctx); err != nil {
		raw.Close()
		return nil, err
	}
	state := conn.ConnectionState()
	_ = conn.Close()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no peer certificate")
	}
	cert := state.PeerCertificates[0]
	return &tlsObs{CN: cert.Subject.CommonName, Issuer: cert.Issuer.CommonName, SANs: cert.DNSNames}, nil
}

// --- small helpers ---

func mergeVendor(a, b VendorResult) VendorResult {
	if a.Vendor == "" {
		a.Vendor = b.Vendor
	}
	a.RouterLike = a.RouterLike || b.RouterLike
	a.AdminPanel = a.AdminPanel || b.AdminPanel
	a.ONT = a.ONT || b.ONT
	a.Modem = a.Modem || b.Modem
	a.Hints = appendUnique(a.Hints, b.Hints...)
	return a
}

func containsInt(values []int, n int) bool {
	return slices.Contains(values, n)
}

func htmlTitle(body string) string {
	lower := strings.ToLower(body)
	open := strings.Index(lower, "<title")
	if open < 0 {
		return ""
	}
	gt := strings.Index(lower[open:], ">")
	if gt < 0 {
		return ""
	}
	start := open + gt + 1
	end := strings.Index(lower[start:], "</title>")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(body[start : start+end])
}

func authRealm(header string) string {
	lower := strings.ToLower(header)
	idx := strings.Index(lower, "realm=")
	if idx < 0 {
		return ""
	}
	value := strings.TrimSpace(header[idx+len("realm="):])
	value = strings.Trim(value, `"`)
	if comma := strings.Index(value, ","); comma >= 0 {
		value = value[:comma]
	}
	return strings.Trim(value, `"`)
}

func pruneEmpty(m map[string]string) map[string]string {
	for k, v := range m {
		if strings.TrimSpace(v) == "" {
			delete(m, k)
		}
	}
	return m
}

func itoaPtr(v *int) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
}

func ftoaPtr(v *float64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.1f", *v)
}
