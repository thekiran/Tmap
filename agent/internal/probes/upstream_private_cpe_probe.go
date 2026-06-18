package probes

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/internal/system"
	"github.com/thekiran/iad/pkg/models"
)

type UpstreamPrivateCPEProbe struct {
	funcs upstreamPrivateCPEFuncs
}

type upstreamPrivateCPEFuncs struct {
	gateway    func() (net.IP, error)
	traceroute func(context.Context, string, int) ([]string, error)
	checkTCP   func(context.Context, string, string) bool
	fetch      func(context.Context, string, string) (httpFingerprintV2Result, error)
	tlsInfo    func(context.Context, string, string) (string, []string, string)
}

func (UpstreamPrivateCPEProbe) Name() string { return "upstream_private_cpe_probe" }

func (p UpstreamPrivateCPEProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	f := p.withDefaults()
	candidates := p.candidates(ctx, in, f)
	res.Evidence["gateway_candidates"] = candidates
	res.Evidence["gateway_chain"] = candidates
	res.Evidence["double_nat_possible"] = len(candidates) >= 2
	if len(candidates) == 0 {
		return res, nil
	}

	var devices []models.GatewayDevice
	maxConf := 0.0
	for i, ip := range candidates {
		role := roleDefaultGateway
		if i > 0 {
			role = roleUpstreamGateway
		}
		dev := p.probeCandidate(ctx, f, ip, role)
		maxConf = maxFloat(maxConf, dev.Confidence)
		devices = append(devices, dev)
	}
	res.Evidence["gateway_devices"] = devices
	res.Evidence["device_confidence"] = maxConf
	res.Confidence = maxConf
	return res, nil
}

func (p UpstreamPrivateCPEProbe) withDefaults() upstreamPrivateCPEFuncs {
	f := p.funcs
	if f.gateway == nil {
		f.gateway = network.Gateway
	}
	if f.traceroute == nil {
		f.traceroute = system.Traceroute
	}
	if f.checkTCP == nil {
		f.checkTCP = checkTCPPort
	}
	if f.fetch == nil {
		f.fetch = fetchHTTPFingerprintV2Method
	}
	if f.tlsInfo == nil {
		f.tlsInfo = fetchTLSInfoForPort
	}
	return f
}

func (p UpstreamPrivateCPEProbe) candidates(ctx context.Context, in models.ScanInput, f upstreamPrivateCPEFuncs) []string {
	var out []string
	if gw, err := f.gateway(); err == nil && isRFC1918IPv4(gw.String()) {
		out = appendUnique(out, gw.String())
	}
	if in.Mode == models.ModeDeep && in.Online {
		if hops, err := f.traceroute(ctx, traceTarget, 8); err == nil {
			for _, h := range leadingPrivateHops(hops) {
				out = appendUnique(out, h)
			}
		}
	}
	return out
}

func (p UpstreamPrivateCPEProbe) probeCandidate(ctx context.Context, f upstreamPrivateCPEFuncs, ip, role string) models.GatewayDevice {
	dev := models.GatewayDevice{
		IP: ip, Role: role, ReachableState: models.ReachableUnknown,
		UPnPState: "not_probed", TR064State: "not_probed", SNMPState: "not_probed",
		EvidenceIDs: []string{p.Name() + ":" + ip},
	}
	if !isRFC1918IPv4(ip) {
		dev.FailedAttempts = append(dev.FailedAttempts, models.ProbeAttempt{
			Source: p.Name(), Target: ip, Protocol: "safety", Error: "public/non-private candidate skipped",
			EvidenceID: p.Name() + ":" + ip + ":skipped",
		})
		return dev
	}

	for _, port := range []string{"80", "443", "8080", "8443", "7547"} {
		if f.checkTCP(ctx, ip, port) {
			dev.Reachable = true
			dev.ReachableState = models.ReachableTrue
			dev.OpenPorts = appendUniqueInts(dev.OpenPorts, atoi(port))
		} else {
			dev.FailedAttempts = append(dev.FailedAttempts, models.ProbeAttempt{
				Source: p.Name(), Target: ip, Protocol: "tcp", Port: atoi(port),
				Error: "tcp connect failed", EvidenceID: p.Name() + ":" + ip + ":tcp:" + port + ":failed",
			})
		}
	}

	var textParts []string
	for _, endpoint := range gatewayHTTPURLs(ip) {
		if !isRFC1918IPv4(endpointIP(endpoint)) {
			continue
		}
		for _, method := range []string{http.MethodHead, http.MethodGet} {
			fp, err := f.fetch(ctx, method, endpoint)
			if err != nil {
				dev.FailedAttempts = append(dev.FailedAttempts, failedHTTPAttempt(p.Name(), ip, endpoint, method, err))
				if method == http.MethodHead {
					continue
				}
				break
			}
			dev.Reachable = true
			dev.ReachableState = models.ReachableTrue
			if port := endpointPort(endpoint); port != 0 {
				dev.OpenPorts = appendUniqueInts(dev.OpenPorts, port)
			}
			dev.HTTPObservations = append(dev.HTTPObservations, httpObservationFromFingerprint(p.Name(), endpoint, method, fp))
			dev.HTTPTitle = firstNonEmpty(dev.HTTPTitle, fp.Title)
			dev.ServerHeader = firstNonEmpty(dev.ServerHeader, fp.Server)
			dev.WWWAuthenticate = firstNonEmpty(dev.WWWAuthenticate, fp.WWWAuthenticate)
			dev.WWWAuthRealm = firstNonEmpty(dev.WWWAuthRealm, fp.WWWAuthRealm)
			dev.RedirectPath = firstNonEmpty(dev.RedirectPath, fp.RedirectPath)
			dev.RedirectLocation = firstNonEmpty(dev.RedirectLocation, fp.RedirectLocation)
			dev.FaviconHash = firstNonEmpty(dev.FaviconHash, fp.FaviconHash)
			dev.HTMLMetaGenerator = firstNonEmpty(dev.HTMLMetaGenerator, fp.HTMLMetaGenerator)
			dev.LoginLabels = appendUniqueStrings(dev.LoginLabels, fp.LoginLabels...)
			textParts = append(textParts, fp.Title, fp.Server, fp.WWWAuthRealm, fp.RedirectLocation, fp.HTMLMetaGenerator, strings.Join(fp.LoginLabels, " "))
			if method == http.MethodGet && !isGenericHTTPTitle(fp.Title) {
				textParts = append(textParts, fp.Body)
			}
		}
	}

	for _, port := range []string{"443", "8443"} {
		if !containsInt(dev.OpenPorts, atoi(port)) {
			continue
		}
		cn, sans, issuer := f.tlsInfo(ctx, ip, port)
		if cn == "" && len(sans) == 0 && issuer == "" {
			continue
		}
		dev.TLSCertCN = firstNonEmpty(dev.TLSCertCN, cn)
		dev.TLSCertSANs = appendUniqueStrings(dev.TLSCertSANs, sans...)
		dev.TLSCertIssuer = firstNonEmpty(dev.TLSCertIssuer, issuer)
		dev.TLSServerName = firstNonEmpty(dev.TLSServerName, cn)
		dev.TLSObservations = append(dev.TLSObservations, models.TLSObservation{
			Source: p.Name(), IP: ip, Port: atoi(port), CN: cn, SANs: sans, Issuer: issuer, ServerName: cn,
			EvidenceID: p.Name() + ":" + ip + ":tls:" + port,
		})
		textParts = append(textParts, cn, strings.Join(sans, " "), issuer)
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
	if len(dev.AccessHints) > 0 {
		dev.AccessEvidence = append(dev.AccessEvidence, models.GatewayAccessEvidence{
			Source: p.Name(), Type: "device_text", Value: accessText, Strength: "medium",
			Confidence: dev.AccessConfidence, Hints: dev.AccessHints, EvidenceID: p.Name() + ":" + ip + ":access",
		})
	}
	return dev
}

func fetchHTTPFingerprintV2Method(ctx context.Context, method, endpoint string) (httpFingerprintV2Result, error) {
	client := &http.Client{
		Timeout:   2 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return httpFingerprintV2Result{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return httpFingerprintV2Result{}, err
	}
	defer resp.Body.Close()
	var text string
	if method != http.MethodHead {
		body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		if err != nil {
			return httpFingerprintV2Result{}, err
		}
		text = string(body)
	}
	authHeader := resp.Header.Get("WWW-Authenticate")
	redirectLocation := resp.Header.Get("Location")
	return httpFingerprintV2Result{
		StatusCode: resp.StatusCode, Title: extractTitle(text), Server: resp.Header.Get("Server"),
		WWWAuthenticate: authHeader, WWWAuthRealm: parseAuthRealm(authHeader),
		RedirectPath: redirectPath(resp), RedirectLocation: redirectLocation,
		FaviconHash: fetchFaviconHash(ctx, client, endpoint), HTMLMetaGenerator: extractMetaGenerator(text),
		LoginLabels: extractLoginLabels(text), Body: text,
	}, nil
}

func fetchTLSInfoForPort(ctx context.Context, ip, port string) (string, []string, string) {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 1200 * time.Millisecond},
		Config:    &tls.Config{InsecureSkipVerify: true, ServerName: ip},
	}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
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

func endpointIP(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func containsInt(values []int, want int) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
