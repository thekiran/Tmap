package deviceintel

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

type DialFunc func(ctx context.Context, network, address string) error

type SafeTCPProbe struct {
	Ports   []int
	Timeout time.Duration
	Dial    DialFunc
}

func (p SafeTCPProbe) Name() string { return "device_intel_tcp_sweep" }

func (p SafeTCPProbe) Run(ctx context.Context, scope ScanScope, store *EvidenceStore) ProbeResult {
	start := time.Now().UTC()
	result := ProbeResult{ProbeName: p.Name(), Status: StatusSuccess, StartedAt: start}
	ports := p.Ports
	if len(ports) == 0 {
		ports = SafeTCPPorts()
	}
	timeout := p.Timeout
	if timeout == 0 {
		timeout = 750 * time.Millisecond
	}
	dial := p.Dial
	if dial == nil {
		dial = func(ctx context.Context, network, address string) error {
			var d net.Dialer
			conn, err := d.DialContext(ctx, network, address)
			if err != nil {
				return err
			}
			return conn.Close()
		}
	}
	for _, ip := range scope.Targets {
		if !ShouldProbeTarget(scope, ip) {
			result.SkippedTargets = append(result.SkippedTargets, ip)
			continue
		}
		for _, port := range ports {
			attemptCtx, cancel := context.WithTimeout(ctx, timeout)
			err := dial(attemptCtx, "tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)))
			cancel()
			if err != nil {
				store.AddFailedAttempt(ip, models.ProbeAttempt{
					Source:   p.Name(),
					Target:   ip,
					Protocol: "tcp",
					Port:     port,
					Error:    err.Error(),
					Timeout:  strings.Contains(strings.ToLower(err.Error()), "timeout") || attemptCtx.Err() == context.DeadlineExceeded,
				})
				continue
			}
			evID := store.AddObservation(ip, p.Name(), "tcp_connect",
				map[string]any{"ip": ip, "port": port, "protocol": "tcp"},
				map[string]any{"state": "open", "service": serviceName(port)}, 0.75, "")
			store.AddService(ip, models.DeviceIntelService{
				Port:        port,
				Protocol:    "tcp",
				State:       "open",
				Name:        serviceName(port),
				Confidence:  0.75,
				EvidenceIDs: []string{evID},
			})
			result.Observations = append(result.Observations, Observation{
				ID:            evID,
				DeviceID:      deviceID(ip),
				IP:            ip,
				SourceProbe:   p.Name(),
				Kind:          "tcp_connect",
				Confidence:    0.75,
				Timestamp:     start,
				SafeToDisplay: true,
			})
		}
	}
	result.FinishedAt = time.Now().UTC()
	if len(result.Observations) == 0 && len(result.SkippedTargets) > 0 {
		result.Status = StatusSkipped
	}
	return result
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type HTTPFingerprintProbe struct {
	Client  HTTPDoer
	Timeout time.Duration
}

func (p HTTPFingerprintProbe) Name() string { return "device_intel_http_fingerprint" }

func (p HTTPFingerprintProbe) Run(ctx context.Context, scope ScanScope, store *EvidenceStore) ProbeResult {
	start := time.Now().UTC()
	result := ProbeResult{ProbeName: p.Name(), Status: StatusSuccess, StartedAt: start}
	client := p.Client
	if client == nil {
		timeout := p.Timeout
		if timeout == 0 {
			timeout = 2 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}
	for _, ip := range scope.Targets {
		if !ShouldProbeTarget(scope, ip) {
			result.SkippedTargets = append(result.SkippedTargets, ip)
			continue
		}
		for _, port := range []int{80, 443, 8080, 8443} {
			scheme := "http"
			if port == 443 || port == 8443 {
				scheme = "https"
			}
			url := fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(ip, fmt.Sprintf("%d", port)))
			obs, err := httpProbeOnce(ctx, client, url)
			if err != nil {
				store.AddFailedAttempt(ip, models.ProbeAttempt{Source: p.Name(), Target: ip, Protocol: "tcp", Port: port, URL: url, Method: "HEAD", Error: err.Error(), Timeout: strings.Contains(strings.ToLower(err.Error()), "timeout")})
				continue
			}
			obs.Source = p.Name()
			obs.URL = url
			obs.EvidenceID = store.AddObservation(ip, p.Name(), "http_fingerprint",
				map[string]any{"url": url, "status_code": obs.StatusCode, "server": obs.ServerHeader},
				map[string]any{"title": obs.Title, "server": obs.ServerHeader, "realm": obs.WWWAuthRealm}, 0.80, "")
			store.AddHTTPObservation(ip, obs, true, models.ProbeAttempt{})
			result.Observations = append(result.Observations, Observation{ID: obs.EvidenceID, DeviceID: deviceID(ip), IP: ip, SourceProbe: p.Name(), Kind: "http_fingerprint", Confidence: 0.80, Timestamp: start, SafeToDisplay: true})
		}
	}
	result.FinishedAt = time.Now().UTC()
	if len(result.Observations) == 0 && len(result.SkippedTargets) > 0 {
		result.Status = StatusSkipped
	}
	return result
}

func httpProbeOnce(ctx context.Context, client HTTPDoer, rawURL string) (models.HTTPObservation, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return models.HTTPObservation{}, err
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode == http.StatusMethodNotAllowed {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if reqErr != nil {
			return models.HTTPObservation{}, reqErr
		}
		resp, err = client.Do(req)
	}
	if err != nil {
		return models.HTTPObservation{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	obs := models.HTTPObservation{
		Method:           resp.Request.Method,
		StatusCode:       resp.StatusCode,
		ServerHeader:     resp.Header.Get("Server"),
		WWWAuthenticate:  resp.Header.Get("WWW-Authenticate"),
		WWWAuthRealm:     authRealm(resp.Header.Get("WWW-Authenticate")),
		RedirectLocation: resp.Header.Get("Location"),
		Title:            htmlTitle(string(body)),
	}
	return obs, nil
}

type TLSFingerprintProbe struct {
	Timeout time.Duration
	DialTLS func(ctx context.Context, network, address string, config *tls.Config) (*tls.Conn, error)
}

func (p TLSFingerprintProbe) Name() string { return "device_intel_tls_fingerprint" }

func (p TLSFingerprintProbe) Run(ctx context.Context, scope ScanScope, store *EvidenceStore) ProbeResult {
	start := time.Now().UTC()
	result := ProbeResult{ProbeName: p.Name(), Status: StatusSuccess, StartedAt: start}
	timeout := p.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	dialTLS := p.DialTLS
	if dialTLS == nil {
		dialTLS = func(ctx context.Context, network, address string, config *tls.Config) (*tls.Conn, error) {
			var d net.Dialer
			raw, err := d.DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			conn := tls.Client(raw, config)
			if err := conn.HandshakeContext(ctx); err != nil {
				raw.Close()
				return nil, err
			}
			return conn, nil
		}
	}
	for _, ip := range scope.Targets {
		if !ShouldProbeTarget(scope, ip) {
			result.SkippedTargets = append(result.SkippedTargets, ip)
			continue
		}
		for _, port := range []int{443, 8443} {
			attemptCtx, cancel := context.WithTimeout(ctx, timeout)
			conn, err := dialTLS(attemptCtx, "tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)), &tls.Config{InsecureSkipVerify: true})
			cancel()
			if err != nil {
				store.AddFailedAttempt(ip, models.ProbeAttempt{Source: p.Name(), Target: ip, Protocol: "tcp", Port: port, Error: err.Error(), Timeout: strings.Contains(strings.ToLower(err.Error()), "timeout")})
				continue
			}
			state := conn.ConnectionState()
			_ = conn.Close()
			if len(state.PeerCertificates) == 0 {
				continue
			}
			cert := state.PeerCertificates[0]
			obs := models.TLSObservation{Source: p.Name(), IP: ip, Port: port, CN: cert.Subject.CommonName, SANs: cert.DNSNames, Issuer: cert.Issuer.CommonName}
			obs.EvidenceID = store.AddObservation(ip, p.Name(), "tls_fingerprint",
				map[string]any{"ip": ip, "port": port, "cn": obs.CN, "issuer": obs.Issuer},
				map[string]any{"sans": strings.Join(obs.SANs, ",")}, 0.80, "")
			store.AddTLSObservation(ip, obs)
			result.Observations = append(result.Observations, Observation{ID: obs.EvidenceID, DeviceID: deviceID(ip), IP: ip, SourceProbe: p.Name(), Kind: "tls_fingerprint", Confidence: 0.80, Timestamp: start, SafeToDisplay: true})
		}
	}
	result.FinishedAt = time.Now().UTC()
	if len(result.Observations) == 0 && len(result.SkippedTargets) > 0 {
		result.Status = StatusSkipped
	}
	return result
}

type OptInProbe struct {
	ProbeName string
	Kind      string
	Allowed   func(ScanScope) bool
}

func (p OptInProbe) Name() string { return p.ProbeName }

func (p OptInProbe) Run(ctx context.Context, scope ScanScope, store *EvidenceStore) ProbeResult {
	_ = ctx
	_ = store
	if p.Allowed == nil || !p.Allowed(scope) {
		return ProbeResult{
			ProbeName:      p.Name(),
			Status:         StatusSkipped,
			SkippedTargets: append([]string(nil), scope.Targets...),
			Errors:         []string{p.Kind + " requires explicit opt-in credentials."},
			StartedAt:      time.Now().UTC(),
			FinishedAt:     time.Now().UTC(),
		}
	}
	return ProbeResult{
		ProbeName:  p.Name(),
		Status:     StatusSkipped,
		Errors:     []string{p.Kind + " opt-in transport is not wired in this build."},
		StartedAt:  time.Now().UTC(),
		FinishedAt: time.Now().UTC(),
	}
}

func SNMPOptInProbe() Probe {
	return OptInProbe{ProbeName: "device_intel_snmp_optin", Kind: "SNMP", Allowed: SNMPAllowed}
}

func SSHBannerOptInProbe() Probe {
	return OptInProbe{ProbeName: "device_intel_ssh_optin", Kind: "SSH banner/authenticated SSH", Allowed: SSHAllowed}
}

func TR064OptInProbe() Probe {
	return OptInProbe{ProbeName: "device_intel_tr064_optin", Kind: "TR-064", Allowed: TR064Allowed}
}

func TR181OptInProbe() Probe {
	return OptInProbe{ProbeName: "device_intel_tr181_optin", Kind: "TR-181", Allowed: TR181Allowed}
}

func RouterAPIOptInProbe() Probe {
	return OptInProbe{ProbeName: "device_intel_router_api_optin", Kind: "router API", Allowed: RouterAPIAllowed}
}

func htmlTitle(body string) string {
	lower := strings.ToLower(body)
	start := strings.Index(lower, "<title")
	if start < 0 {
		return ""
	}
	start = strings.Index(lower[start:], ">")
	if start < 0 {
		return ""
	}
	startAbs := strings.Index(lower, "<title") + start + 1
	end := strings.Index(lower[startAbs:], "</title>")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(body[startAbs : startAbs+end])
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
