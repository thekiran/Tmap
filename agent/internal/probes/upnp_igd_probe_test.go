package probes

import (
	"context"
	"testing"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

func TestUPnPIGDReturnsDSLAccessHint(t *testing.T) {
	p := fakeUPnPIGDProbe([]upnpIGDService{{
		ServiceType: "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1",
		ControlURL:  "/upnp/control/WANCommonIFC1",
	}}, map[string]string{
		"NewWANAccessType":              "DSL",
		"NewPhysicalLinkStatus":         "Up",
		"NewLayer1UpstreamMaxBitRate":   "20480000",
		"NewLayer1DownstreamMaxBitRate": "102400000",
	})
	res, err := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !res.Evidence["strong_access_evidence"].(bool) {
		t.Fatalf("strong_access_evidence = %v, want true", res.Evidence["strong_access_evidence"])
	}
	if len(res.Hints) == 0 {
		t.Fatalf("hints empty, want DSL hint")
	}
	if res.Evidence["wan_access_type"] != "DSL" {
		t.Fatalf("wan_access_type = %v, want DSL", res.Evidence["wan_access_type"])
	}
}

func TestUPnPIGDGenericNATDoesNotProduceAccessHint(t *testing.T) {
	p := fakeUPnPIGDProbe([]upnpIGDService{{
		ServiceType: "urn:schemas-upnp-org:service:WANIPConnection:1",
		ControlURL:  "/upnp/control/WANIPConn1",
	}}, nil)
	res, err := p.Run(context.Background(), models.ScanInput{Mode: models.ModeQuick})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Hints) != 0 {
		t.Fatalf("hints = %v, want empty", res.Hints)
	}
	if getProbeBool(res.Evidence, "strong_access_evidence") {
		t.Fatal("generic IGD/NAT service must not be strong access evidence")
	}
}

func fakeUPnPIGDProbe(services []upnpIGDService, soapValues map[string]string) UPnPIGDProbe {
	return UPnPIGDProbe{funcs: upnpIGDProbeFuncs{
		ssdpSearch: func(ctx context.Context, wait time.Duration) ([]string, error) {
			return []string{"http://192.168.1.1:1900/rootDesc.xml"}, nil
		},
		fetchDescription: func(ctx context.Context, loc string) (*upnpIGDDescription, error) {
			return &upnpIGDDescription{Device: upnpIGDDevice{
				DeviceType: "urn:schemas-upnp-org:device:InternetGatewayDevice:1",
				Services:   services,
			}}, nil
		},
		soapAction: func(ctx context.Context, controlURL, serviceType, action string) (map[string]string, error) {
			return soapValues, nil
		},
	}}
}

func getProbeBool(ev map[string]any, key string) bool {
	v, _ := ev[key].(bool)
	return v
}

