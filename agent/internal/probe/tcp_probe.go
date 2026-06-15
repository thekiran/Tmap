package probe

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/thekiran/iad/internal/model"
	"github.com/thekiran/iad/internal/safety"
)

var defaultSafeTCPPorts = []int{22, 23, 53, 80, 443, 445, 554, 631, 1900, 5000, 5001, 5353, 8080, 8443, 9100}

type TCPConnectProbe struct {
	Ports   []int
	Timeout time.Duration
}

func (TCPConnectProbe) Name() string            { return "tcp_connect_probe" }
func (TCPConnectProbe) SafeModeAllowed() bool   { return false }
func (TCPConnectProbe) NormalModeAllowed() bool { return true }
func (TCPConnectProbe) DeepModeAllowed() bool   { return true }

func (p TCPConnectProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	ports := p.Ports
	if len(ports) == 0 {
		ports = defaultSafeTCPPorts
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 700 * time.Millisecond
	}
	result := model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess}
	for _, ip := range input.CandidateIPs {
		if !safety.IsPrivateIPString(ip) {
			continue
		}
		var openPorts []model.PortInfo
		var services []model.ServiceInfo
		for _, port := range ports {
			select {
			case <-ctx.Done():
				return result, ctx.Err()
			default:
			}
			target := net.JoinHostPort(ip, strconv.Itoa(port))
			dialer := net.Dialer{Timeout: timeout}
			conn, err := dialer.DialContext(ctx, "tcp", target)
			if err != nil {
				continue
			}
			_ = conn.Close()
			reason := fmt.Sprintf("TCP connect to safe local port %d succeeded.", port)
			ev := baseEvidence(p.Name(), ip, model.EvidenceMedium, 0.50, reason, map[string]any{"ip": ip, "port": port, "protocol": "tcp"})
			result.Evidence = append(result.Evidence, ev)
			openPorts = append(openPorts, model.PortInfo{Port: port, Protocol: "tcp", State: "open", Reason: reason})
			services = append(services, model.ServiceInfo{Name: serviceName(port), Port: port, Protocol: "tcp"})
		}
		if len(openPorts) > 0 {
			result.Devices = append(result.Devices, model.Device{
				ID:          "ip_" + ip,
				IPAddresses: []string{ip},
				DeviceType:  model.DeviceTypeUnknown,
				OpenPorts:   openPorts,
				Services:    services,
				Confidence:  0.45,
				Evidence:    result.Evidence,
			})
		}
	}
	if len(result.Evidence) == 0 {
		result.Status = model.ProbeStatusPartial
	}
	return result, nil
}
