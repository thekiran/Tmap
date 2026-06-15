package probes

import (
	"context"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

func TestGatewayChainCreatesDeviceCandidates(t *testing.T) {
	p := fakeGatewayChainProbe("192.168.31.1", []string{"192.168.31.1", "192.168.1.1", "95.15.180.1"})
	res, err := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep, Online: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	chain := res.Evidence["gateway_chain"].([]string)
	if len(chain) != 2 || chain[0] != "192.168.31.1" || chain[1] != "192.168.1.1" {
		t.Fatalf("gateway_chain = %v, want [192.168.31.1 192.168.1.1]", chain)
	}
	devices := res.Evidence["gateway_devices"].([]models.GatewayDevice)
	if len(devices) != 2 {
		t.Fatalf("gateway_devices len = %d, want 2", len(devices))
	}
}

func TestOnlyPrivateGatewayIPsAreProbed(t *testing.T) {
	probed := map[string]bool{}
	p := fakeGatewayChainProbe("192.168.31.1", []string{"192.168.31.1", "192.168.1.1", "8.8.8.8"})
	p.funcs.httpGet = func(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
		host, _, _ := net.SplitHostPort(endpointHost(endpoint))
		if host == "" {
			host = endpointHost(endpoint)
		}
		probed[host] = true
		return gatewayHTTPResult{}, errFakeHTTP
	}
	_, _ = p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep, Online: true})
	if !probed["192.168.31.1"] || !probed["192.168.1.1"] {
		t.Fatalf("private gateways were not probed: %v", probed)
	}
	if probed["8.8.8.8"] {
		t.Fatal("public hop must not be probed")
	}
}

func TestPublicIPsAreNotProbed(t *testing.T) {
	probed := false
	p := fakeGatewayChainProbe("8.8.8.8", []string{"8.8.8.8"})
	p.funcs.httpGet = func(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
		probed = true
		return gatewayHTTPResult{}, nil
	}
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep, Online: true})
	if probed {
		t.Fatal("public IP must not be probed")
	}
	if chain := res.Evidence["gateway_chain"].([]string); len(chain) != 0 {
		t.Fatalf("gateway_chain = %v, want empty", chain)
	}
}

func TestDefaultGatewayRoleDetected(t *testing.T) {
	p := fakeGatewayChainProbe("192.168.31.1", nil)
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick, Online: false})
	devices := res.Evidence["gateway_devices"].([]models.GatewayDevice)
	if devices[0].Role != roleDefaultGateway {
		t.Fatalf("role = %q, want %q", devices[0].Role, roleDefaultGateway)
	}
}

func TestUpstreamGatewayRoleDetected(t *testing.T) {
	p := fakeGatewayChainProbe("192.168.31.1", []string{"192.168.31.1", "192.168.1.1", "95.15.180.1"})
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep, Online: true})
	devices := res.Evidence["gateway_devices"].([]models.GatewayDevice)
	if devices[1].Role != roleUpstreamGateway && devices[1].Role != rolePossibleModem {
		t.Fatalf("upstream role = %q", devices[1].Role)
	}
}

func TestLikelyModemSelectedFromUpstream(t *testing.T) {
	p := fakeGatewayChainProbe("192.168.31.1", []string{"192.168.31.1", "192.168.1.1", "95.15.180.1"})
	p.funcs.httpGet = func(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
		if stringsContains(endpoint, "192.168.1.1") {
			return gatewayHTTPResult{Title: "Zyxel VMG3312-B10B VDSL"}, nil
		}
		return gatewayHTTPResult{Title: "Mi Router"}, nil
	}
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep, Online: true})
	if got := res.Evidence["likely_modem_ip"]; got != "192.168.1.1" {
		t.Fatalf("likely_modem_ip = %v, want 192.168.1.1", got)
	}
}

func TestNginxHeaderDoesNotProduceFiberHint(t *testing.T) {
	p := fakeGatewayChainProbe("192.168.31.1", nil)
	p.funcs.httpGet = func(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
		return gatewayHTTPResult{Server: "nginx"}, nil
	}
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick, Online: false})
	devices := res.Evidence["gateway_devices"].([]models.GatewayDevice)
	if len(devices[0].AccessHints) != 0 {
		t.Fatalf("access_hints = %v, want empty", devices[0].AccessHints)
	}
	if devices[0].AccessConfidence != 0 {
		t.Fatalf("access_confidence = %v, want 0", devices[0].AccessConfidence)
	}
}

func TestGenericHTTPDoesNotProduceAccessHint(t *testing.T) {
	p := fakeGatewayChainProbe("192.168.31.1", nil)
	p.funcs.httpGet = func(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
		return gatewayHTTPResult{Title: "Router Login", Server: "Apache", Body: "<html>login</html>"}, nil
	}
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick, Online: false})
	devices := res.Evidence["gateway_devices"].([]models.GatewayDevice)
	if len(devices[0].AccessHints) != 0 {
		t.Fatalf("access_hints = %v, want empty", devices[0].AccessHints)
	}
}

func TestGatewayReachableDoesNotIncreaseAccessConfidence(t *testing.T) {
	p := fakeGatewayChainProbe("192.168.31.1", nil)
	p.funcs.httpGet = func(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
		return gatewayHTTPResult{Server: "nginx"}, nil
	}
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick, Online: false})
	devices := res.Evidence["gateway_devices"].([]models.GatewayDevice)
	if !devices[0].Reachable || devices[0].DeviceConfidence == 0 {
		t.Fatalf("device should be reachable with device confidence: %#v", devices[0])
	}
	if devices[0].AccessConfidence != 0 {
		t.Fatalf("access_confidence = %v, want 0", devices[0].AccessConfidence)
	}
}

func TestGenericFaviconDoesNotProduceAccessHint(t *testing.T) {
	p := fakeGatewayChainProbe("192.168.31.1", nil)
	p.funcs.httpGet = func(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
		return gatewayHTTPResult{Title: "Router", FaviconHash: "abcdef1234567890"}, nil
	}
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick, Online: false})
	devices := res.Evidence["gateway_devices"].([]models.GatewayDevice)
	if devices[0].FaviconHash == "" || devices[0].DeviceConfidence == 0 {
		t.Fatalf("favicon should contribute device identity: %#v", devices[0])
	}
	if len(devices[0].AccessHints) != 0 || devices[0].AccessConfidence != 0 {
		t.Fatalf("favicon must not classify access: %#v", devices[0])
	}
}

func TestUpstreamGatewayUnreachableDoesNotSelectDefaultAsModem(t *testing.T) {
	p := fakeGatewayChainProbe("192.168.31.1", []string{"192.168.31.1", "192.168.1.1", "95.15.180.1"})
	p.funcs.httpGet = func(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
		if stringsContains(endpoint, "192.168.1.1") {
			return gatewayHTTPResult{}, errFakeHTTP
		}
		return gatewayHTTPResult{Server: "nginx"}, nil
	}
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep, Online: true})
	if _, ok := res.Evidence["likely_modem_ip"]; ok {
		t.Fatalf("likely_modem_ip must be omitted for generic default + unreachable upstream: %v", res.Evidence["likely_modem_ip"])
	}
	devices := res.Evidence["gateway_devices"].([]models.GatewayDevice)
	if devices[0].Role != roleDefaultGateway {
		t.Fatalf("default role = %q, want default_gateway", devices[0].Role)
	}
}

func TestDoubleNATPrefersUpstreamAsModemCandidate(t *testing.T) {
	p := fakeGatewayChainProbe("192.168.31.1", []string{"192.168.31.1", "192.168.1.1", "95.15.180.1"})
	p.funcs.httpGet = func(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
		if stringsContains(endpoint, "192.168.1.1") {
			return gatewayHTTPResult{Title: "Zyxel VMG3312-B10B VDSL"}, nil
		}
		return gatewayHTTPResult{Server: "nginx"}, nil
	}
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep, Online: true})
	if got := res.Evidence["likely_modem_ip"]; got != "192.168.1.1" {
		t.Fatalf("likely_modem_ip = %v, want 192.168.1.1", got)
	}
}

type fakeHTTPError struct{}

func (fakeHTTPError) Error() string { return "fake http error" }

var errFakeHTTP fakeHTTPError

func fakeGatewayChainProbe(gateway string, hops []string) GatewayChainProbe {
	return GatewayChainProbe{funcs: gatewayChainFuncs{
		gateway: func() (net.IP, error) {
			return net.ParseIP(gateway), nil
		},
		traceroute: func(ctx context.Context, host string, maxHops int) ([]string, error) {
			return hops, nil
		},
		httpGet: func(ctx context.Context, endpoint string) (gatewayHTTPResult, error) {
			return gatewayHTTPResult{Title: "Router"}, nil
		},
		checkTCP: func(ctx context.Context, ip, port string) bool {
			return false
		},
		ssdpSearch: func(ctx context.Context, wait time.Duration) ([]string, error) {
			return nil, nil
		},
		fetchDevice: func(ctx context.Context, loc string) (*upnpDevice, error) {
			return nil, errFakeHTTP
		},
	}}
}

func endpointHost(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	return u.Host
}

func stringsContains(s, sub string) bool {
	return strings.Contains(s, sub)
}
