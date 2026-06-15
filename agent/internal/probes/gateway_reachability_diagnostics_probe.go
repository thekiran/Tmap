package probes

import (
	"context"
	"net"
	"time"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

type GatewayReachabilityDiagnosticsProbe struct {
	funcs gatewayReachabilityFuncs
}

type gatewayReachabilityFuncs struct {
	gateway  func() (net.IP, error)
	checkTCP func(context.Context, string, string) bool
}

func (GatewayReachabilityDiagnosticsProbe) Name() string {
	return "gateway_reachability_diagnostics_probe"
}

func (p GatewayReachabilityDiagnosticsProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	f := p.withDefaults()
	gw, err := f.gateway()
	if err != nil || gw == nil || !isRFC1918IPv4(gw.String()) {
		res.Status = models.StatusSkipped
		return res, nil
	}
	ip := gw.String()
	ports := []string{}
	for _, port := range []string{"80", "443", "8080", "49000"} {
		if f.checkTCP(ctx, ip, port) {
			ports = append(ports, port)
		}
	}
	res.Evidence["gateway_ip"] = ip
	res.Evidence["route_present"] = true
	res.Evidence["tcp_ports_reachable"] = ports
	res.Evidence["management_reachable"] = len(ports) > 0
	if len(ports) == 0 {
		res.Evidence["reason_if_unreachable"] = "The gateway is present in the route table, but no tested management TCP port responded. This does not prove the gateway is unreachable."
	}
	res.Evidence["network_confidence"] = 0.35
	res.Confidence = 0.35
	return res, nil
}

func (p GatewayReachabilityDiagnosticsProbe) withDefaults() gatewayReachabilityFuncs {
	f := p.funcs
	if f.gateway == nil {
		f.gateway = network.Gateway
	}
	if f.checkTCP == nil {
		f.checkTCP = func(ctx context.Context, ip, port string) bool {
			d := net.Dialer{Timeout: 600 * time.Millisecond}
			conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(ip, port))
			if err != nil {
				return false
			}
			conn.Close()
			return true
		}
	}
	return f
}

