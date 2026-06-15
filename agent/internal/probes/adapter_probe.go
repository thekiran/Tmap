package probes

import (
	"context"
	"strings"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

// AdapterProbe enumerates the host's network adapters. A directly-attached
// cellular adapter (the host *is* the modem) is a strong mobile hint; ordinary
// Wi-Fi/Ethernet to a router tells us nothing about the WAN, so it is recorded
// as evidence but emits no hint.
type AdapterProbe struct{}

func (AdapterProbe) Name() string { return "adapter_probe" }

func (p AdapterProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	adapters, err := network.Adapters()
	if err != nil {
		return res, err
	}
	res.Evidence["adapters"] = adapters

	var active []string
	for _, a := range adapters {
		if a.Up && len(a.Addrs) > 0 {
			active = append(active, a.Name)
		}
		if isCellularName(a.Name) && a.Up {
			res.Hints = appendUnique(res.Hints, models.TypeMobile)
		}
	}
	res.Evidence["active"] = active
	res.Confidence = 0.4
	return res, nil
}

func isCellularName(name string) bool {
	n := strings.ToLower(name)
	for _, kw := range []string{"cellular", "wwan", "mobile broadband", "lte", "5g"} {
		if strings.Contains(n, kw) {
			return true
		}
	}
	return false
}
