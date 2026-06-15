package probes

import (
	"context"
	"net"
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestTR064DSLServiceProducesStrongEvidence(t *testing.T) {
	p := fakeTR064Probe("192.168.1.1", []upnpIGDService{{
		ServiceType: "urn:dslforum-org:service:WANDSLInterfaceConfig:1",
		ControlURL:  "/upnp/control/wandslifconfig1",
	}}, map[string]string{
		"NewStatus":             "Up",
		"NewDataPath":           "PTM",
		"NewModulationType":     "VDSL2",
		"NewUpstreamCurrRate":   "20480",
		"NewDownstreamCurrRate": "102400",
	})
	res, err := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Evidence["tr064_found"].(bool) {
		t.Fatal("tr064_found = false, want true")
	}
	if !res.Evidence["strong_access_evidence"].(bool) {
		t.Fatal("DSL/PTM service should be strong access evidence")
	}
	if len(res.Hints) == 0 {
		t.Fatal("hints empty, want DSL/VDSL")
	}
}

func TestTR064OpenPortWithoutWANDataIsDeviceConfidenceOnly(t *testing.T) {
	p := fakeTR064Probe("192.168.1.1", nil, nil)
	res, err := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Hints) != 0 {
		t.Fatalf("hints = %v, want empty", res.Hints)
	}
	if res.Evidence["access_confidence"].(float64) != 0 {
		t.Fatalf("access_confidence = %v, want 0", res.Evidence["access_confidence"])
	}
	if res.Evidence["device_confidence"].(float64) == 0 {
		t.Fatal("device_confidence should reflect open TR-064 port")
	}
}

func TestTR064DoesNotProbePublicGateway(t *testing.T) {
	probed := false
	p := fakeTR064Probe("8.8.8.8", nil, nil)
	p.funcs.checkTCP = func(ctx context.Context, ip, port string) bool {
		probed = true
		return true
	}
	res, _ := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep, Online: true})
	if probed {
		t.Fatal("public gateway must not be probed")
	}
	if got := res.Evidence["gateway_candidates"].([]string); len(got) != 0 {
		t.Fatalf("gateway_candidates = %v, want empty", got)
	}
}

func fakeTR064Probe(gateway string, services []upnpIGDService, soapValues map[string]string) TR064Probe {
	return TR064Probe{funcs: tr064ProbeFuncs{
		gateway: func() (net.IP, error) {
			return net.ParseIP(gateway), nil
		},
		traceroute: func(ctx context.Context, host string, maxHops int) ([]string, error) {
			return nil, nil
		},
		checkTCP: func(ctx context.Context, ip, port string) bool {
			return true
		},
		fetchDescription: func(ctx context.Context, loc string) (*upnpIGDDescription, error) {
			manufacturer := ""
			model := ""
			if len(services) > 0 {
				manufacturer = "Zyxel"
				model = "VMG3312-B10B"
			}
			return &upnpIGDDescription{Device: upnpIGDDevice{
				DeviceType:   "urn:dslforum-org:device:InternetGatewayDevice:1",
				Manufacturer: manufacturer,
				ModelName:    model,
				Services:     services,
			}}, nil
		},
		soapAction: func(ctx context.Context, controlURL, serviceType, action string) (map[string]string, error) {
			return soapValues, nil
		},
	}}
}
