package probes

import (
	"context"
	"strings"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

type OSInterfaceProbeV2 struct {
	funcs osInterfaceV2Funcs
}

type osInterfaceV2Funcs struct {
	adapters func() ([]network.Adapter, error)
}

func (OSInterfaceProbeV2) Name() string { return "os_interface_probe_v2" }

func (p OSInterfaceProbeV2) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	f := p.withDefaults()
	adapters, err := f.adapters()
	if err != nil {
		return res, err
	}
	localMedium := "Unknown"
	var active []string
	var cellular bool
	for _, a := range adapters {
		if !a.Up || len(a.Addrs) == 0 {
			continue
		}
		active = append(active, a.Name)
		medium := classifyLocalMedium(a.Name)
		if localMedium == "Unknown" {
			localMedium = medium
		}
		if medium == "Cellular" {
			cellular = true
			localMedium = medium
		}
	}
	arch := models.AccessArchitecture{LocalMedium: localMedium}
	res.Evidence["adapters"] = adapters
	res.Evidence["active"] = active
	res.Evidence["access_architecture"] = arch
	res.Evidence["device_confidence"] = 0.25
	if cellular {
		res.Hints = appendUnique(res.Hints, models.TypeMobile)
		res.Hints = appendUnique(res.Hints, models.TypeFWA)
		res.Evidence["local_cellular_evidence"] = true
		res.Evidence["access_confidence"] = 0.45
	}
	res.Confidence = 0.35
	return res, nil
}

func (p OSInterfaceProbeV2) withDefaults() osInterfaceV2Funcs {
	f := p.funcs
	if f.adapters == nil {
		f.adapters = network.Adapters
	}
	return f
}

func classifyLocalMedium(name string) string {
	n := strings.ToLower(name)
	switch {
	case isCellularName(name), strings.Contains(n, "wwan"), strings.Contains(n, "modemmanager"):
		return "Cellular"
	case strings.Contains(n, "wi-fi"), strings.Contains(n, "wifi"), strings.Contains(n, "wireless"), strings.Contains(n, "wlan"):
		return "Wi-Fi"
	case strings.Contains(n, "ethernet"), strings.Contains(n, "eth"), strings.Contains(n, "en"):
		return "Ethernet"
	default:
		return "Unknown"
	}
}

