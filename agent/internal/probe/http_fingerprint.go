package probe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/thekiran/iad/internal/model"
	"github.com/thekiran/iad/internal/safety"
)

type HTTPFingerprintProbe struct {
	Timeout time.Duration
}

func (HTTPFingerprintProbe) Name() string            { return "http_fingerprint_probe" }
func (HTTPFingerprintProbe) SafeModeAllowed() bool   { return false }
func (HTTPFingerprintProbe) NormalModeAllowed() bool { return true }
func (HTTPFingerprintProbe) DeepModeAllowed() bool   { return true }

func (p HTTPFingerprintProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 1200 * time.Millisecond
	}
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, // read-only local fingerprinting
	}
	result := model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess}
	for _, ip := range input.CandidateIPs {
		if !safety.IsPrivateIPString(ip) {
			continue
		}
		for _, schemePort := range []string{"http://%s", "https://%s", "http://%s:8080", "https://%s:8443"} {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			default:
			}
			url := fmt.Sprintf(schemePort, ip)
			req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
			if err != nil {
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			_ = resp.Body.Close()
			raw := map[string]any{
				"url":              url,
				"status_code":      resp.StatusCode,
				"server":           resp.Header.Get("Server"),
				"www_authenticate": resp.Header.Get("WWW-Authenticate"),
				"location":         resp.Header.Get("Location"),
			}
			reason := "Private-host HTTP metadata was read without authentication or login attempts."
			ev := baseEvidence(p.Name(), ip, model.EvidenceMedium, 0.65, reason, raw)
			result.Evidence = append(result.Evidence, ev)
			dev := model.Device{
				ID:          "ip_" + ip,
				IPAddresses: []string{ip},
				DeviceType:  inferHTTPDeviceType(raw),
				OpenPorts:   []model.PortInfo{{Port: httpPort(url), Protocol: "tcp", State: "open"}},
				Services:    []model.ServiceInfo{{Name: "http", Port: httpPort(url), Protocol: "tcp", Raw: raw}},
				Confidence:  0.65,
				Evidence:    []model.Evidence{ev},
			}
			result.Devices = append(result.Devices, dev)
			break
		}
	}
	if len(result.Evidence) == 0 {
		result.Status = model.ProbeStatusPartial
	}
	return result, nil
}

func inferHTTPDeviceType(raw map[string]any) model.DeviceType {
	text := strings.ToLower(fmt.Sprint(raw["server"], " ", raw["www_authenticate"], " ", raw["location"]))
	switch {
	case strings.Contains(text, "printer"), strings.Contains(text, "ipp"):
		return model.DeviceTypePrinter
	case strings.Contains(text, "synology"), strings.Contains(text, "qnap"), strings.Contains(text, "nas"):
		return model.DeviceTypeNAS
	case strings.Contains(text, "camera"), strings.Contains(text, "rtsp"):
		return model.DeviceTypeCamera
	case strings.Contains(text, "router"), strings.Contains(text, "gateway"), strings.Contains(text, "openwrt"):
		return model.DeviceTypeRouter
	default:
		return model.DeviceTypeUnknown
	}
}

func httpPort(url string) int {
	switch {
	case strings.HasPrefix(url, "https://") && strings.Contains(url, ":8443"):
		return 8443
	case strings.HasPrefix(url, "https://"):
		return 443
	case strings.Contains(url, ":8080"):
		return 8080
	default:
		return 80
	}
}
