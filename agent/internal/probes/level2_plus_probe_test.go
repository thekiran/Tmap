package probes

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

func TestUPnPIGDDeepWANAccessTypeDSLProducesPhysicalEvidence(t *testing.T) {
	p := fakeUPnPIGDDeepProbe(map[string]string{"NewWANAccessType": "DSL", "NewPhysicalLinkStatus": "Up"})
	res, err := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ProbeName != "upnp_igd_deep_probe" {
		t.Fatalf("probe_name = %q", res.ProbeName)
	}
	if !res.Evidence["strong_access_evidence"].(bool) {
		t.Fatal("DSL WANAccessType should be physical evidence")
	}
}

func TestUPnPIGDDeepEthernetWANDoesNotClassifyAsFiber(t *testing.T) {
	p := fakeUPnPIGDDeepProbe(map[string]string{"NewWANAccessType": "Ethernet", "NewPhysicalLinkStatus": "Up"})
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep})
	if !containsString(res.Hints, models.TypeEthernetWAN) {
		t.Fatalf("hints = %v, want EthernetWAN", res.Hints)
	}
	if containsString(res.Hints, models.TypeFiber) {
		t.Fatalf("Ethernet WAN must not imply Fiber: %v", res.Hints)
	}
}

func TestTR064AuthRequiredDoesNotBruteForce(t *testing.T) {
	p := fakeTR064Probe("192.168.1.1", nil, nil)
	p.funcs.fetchDescription = func(ctx context.Context, loc string) (*upnpIGDDescription, error) {
		return nil, fmt.Errorf("401 auth required")
	}
	res, err := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Evidence["auth_required"].(bool) {
		t.Fatalf("auth_required = %v, want true", res.Evidence["auth_required"])
	}
	if len(res.Hints) != 0 {
		t.Fatalf("auth-only TR-064 must not hint: %v", res.Hints)
	}
}

func TestTR064WANKeywordProducesStrongPhysicalEvidence(t *testing.T) {
	p := fakeTR064Probe("192.168.1.1", []upnpIGDService{{
		ServiceType: "urn:dslforum-org:service:WANDSLInterfaceConfig:1",
		ControlURL:  "/dsl",
	}}, map[string]string{"NewDataPath": "PTM", "NewModulationType": "VDSL2"})
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick})
	if !res.Evidence["strong_access_evidence"].(bool) || len(res.Hints) == 0 {
		t.Fatalf("TR-064 DSL keywords should produce physical hints: evidence=%v hints=%v", res.Evidence, res.Hints)
	}
}

func TestHTTPFingerprintV2GenericServerDoesNotHint(t *testing.T) {
	p := fakeHTTPFingerprintV2Probe("192.168.1.1", httpFingerprintV2Result{Title: "Router Login", Server: "nginx"})
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick})
	if len(res.Hints) != 0 {
		t.Fatalf("hints = %v, want empty", res.Hints)
	}
	devices := res.Evidence["gateway_devices"].([]models.GatewayDevice)
	if devices[0].AccessConfidence != 0 {
		t.Fatalf("access_confidence = %v, want 0", devices[0].AccessConfidence)
	}
}

func TestHTTPFingerprintV2ModelMatchAllowsAccessHint(t *testing.T) {
	p := fakeHTTPFingerprintV2Probe("192.168.1.1", httpFingerprintV2Result{Title: "Zyxel VMG3312-B10B VDSL"})
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick})
	if len(res.Hints) == 0 {
		t.Fatal("model fingerprint text should allow access hint")
	}
}

func TestIPv6TransitionProducesContextOnly(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("2001:db8::1/64")
	p := IPv6TransitionProbe{funcs: ipv6TransitionFuncs{
		interfaces: func() ([]net.Interface, error) {
			return []net.Interface{{Name: "Ethernet", Flags: net.FlagUp}}, nil
		},
		addrs: func(ifc net.Interface) ([]net.Addr, error) {
			return []net.Addr{ipnet}, nil
		},
	}}
	res, _ := p.Run(context.Background(), models.ScanInput{})
	if len(res.Hints) != 0 {
		t.Fatalf("IPv6 context must not classify access type: %v", res.Hints)
	}
	if res.Evidence["ipv6_context"] == nil {
		t.Fatal("missing ipv6_context")
	}
}

func TestSTUNPCPNATDoesNotInflateFiberOrDSL(t *testing.T) {
	p := STUNPCPNATProbe{funcs: stunPCPNATFuncs{
		publicIP: func(ctx context.Context) (string, error) { return "95.15.1.1", nil },
		gateway:  func() (net.IP, error) { return net.ParseIP("192.168.1.1"), nil },
		stun:     func(ctx context.Context, server string) (string, int, error) { return "95.15.1.1", 40000, nil },
		udpProbe: func(ctx context.Context, ip, port string, payload []byte) bool { return false },
	}}
	res, _ := p.Run(context.Background(), models.ScanInput{Online: true})
	if len(res.Hints) != 0 {
		t.Fatalf("NAT context must not classify access type: %v", res.Hints)
	}
}

func TestOSInterfaceWifiEthernetRemainLocalMediumOnly(t *testing.T) {
	p := OSInterfaceProbeV2{funcs: osInterfaceV2Funcs{adapters: func() ([]network.Adapter, error) {
		return []network.Adapter{{Name: "Wi-Fi", Up: true, Addrs: []string{"192.168.1.10/24"}}}, nil
	}}}
	res, _ := p.Run(context.Background(), models.ScanInput{})
	if len(res.Hints) != 0 {
		t.Fatalf("Wi-Fi must not classify WAN access: %v", res.Hints)
	}
}

func TestOSInterfaceCellularAddsMobileLocalEvidence(t *testing.T) {
	p := OSInterfaceProbeV2{funcs: osInterfaceV2Funcs{adapters: func() ([]network.Adapter, error) {
		return []network.Adapter{{Name: "Mobile Broadband LTE", Up: true, Addrs: []string{"10.0.0.2/24"}}}, nil
	}}}
	res, _ := p.Run(context.Background(), models.ScanInput{})
	if !containsString(res.Hints, models.TypeMobile) {
		t.Fatalf("cellular adapter should hint Mobile: %v", res.Hints)
	}
}

func TestLLDPCDPKeywordlessPacketDoesNotHint(t *testing.T) {
	p := LLDPCDPPassiveProbe{funcs: lldpCDPFuncs{read: func(ctx context.Context, wait time.Duration) ([]LLDPCDPFrame, error) {
		return []LLDPCDPFrame{{Protocol: "LLDP", SystemName: "switch01", SystemDescription: "managed switch"}}, nil
	}}}
	res, _ := p.Run(context.Background(), models.ScanInput{})
	if len(res.Hints) != 0 {
		t.Fatalf("keywordless LLDP must not hint: %v", res.Hints)
	}
}

func TestLLDPCDPGenericFiberSwitchDoesNotHint(t *testing.T) {
	p := LLDPCDPPassiveProbe{funcs: lldpCDPFuncs{read: func(ctx context.Context, wait time.Duration) ([]LLDPCDPFrame, error) {
		return []LLDPCDPFrame{{Protocol: "LLDP", SystemName: "switch01", SystemDescription: "fiber uplink switch"}}, nil
	}}}
	res, _ := p.Run(context.Background(), models.ScanInput{})
	if len(res.Hints) != 0 {
		t.Fatalf("generic fiber switch text must not hint WAN access: %v", res.Hints)
	}
}

func TestPerformanceProbeProducesContextOnly(t *testing.T) {
	p := PerformanceProfileProbe{funcs: performanceProfileFuncs{
		gateway: func() (string, error) { return "192.168.1.1", nil },
		measure: func(ctx context.Context, target string) (time.Duration, time.Duration, string, error) {
			return 5 * time.Millisecond, time.Millisecond, "fake", nil
		},
	}}
	res, _ := p.Run(context.Background(), models.ScanInput{Online: false})
	if len(res.Hints) != 0 {
		t.Fatalf("performance must not classify: %v", res.Hints)
	}
	if res.Evidence["performance_profile"] == nil {
		t.Fatal("missing performance_profile")
	}
}

func fakeUPnPIGDDeepProbe(values map[string]string) UPnPIGDDeepProbe {
	return UPnPIGDDeepProbe{funcs: upnpIGDProbeFuncs{
		ssdpSearch: func(ctx context.Context, wait time.Duration) ([]string, error) {
			return []string{"http://192.168.1.1:1900/rootDesc.xml"}, nil
		},
		fetchDescription: func(ctx context.Context, loc string) (*upnpIGDDescription, error) {
			return &upnpIGDDescription{Device: upnpIGDDevice{
				DeviceType: "urn:schemas-upnp-org:device:InternetGatewayDevice:1",
				Services: []upnpIGDService{{
					ServiceType: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
					ControlURL:  "/wanCommon",
				}},
			}}, nil
		},
		soapAction: func(ctx context.Context, controlURL, serviceType, action string) (map[string]string, error) {
			return values, nil
		},
	}}
}

func fakeHTTPFingerprintV2Probe(gateway string, result httpFingerprintV2Result) HTTPFingerprintV2Probe {
	return HTTPFingerprintV2Probe{funcs: httpFingerprintV2Funcs{
		gateway: func() (net.IP, error) { return net.ParseIP(gateway), nil },
		traceroute: func(ctx context.Context, host string, maxHops int) ([]string, error) {
			return nil, nil
		},
		fetch: func(ctx context.Context, endpoint string) (httpFingerprintV2Result, error) {
			return result, nil
		},
		tlsInfo: func(ctx context.Context, ip string) (string, []string, string) {
			return "", nil, ""
		},
	}}
}

func containsString(values []string, want string) bool {
	for _, v := range values {
		if strings.EqualFold(v, want) {
			return true
		}
	}
	return false
}
