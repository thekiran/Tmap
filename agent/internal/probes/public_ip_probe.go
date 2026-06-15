package probes

import (
	"context"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

// PublicIPProbe fetches the host's public IP and flags carrier-grade NAT. It is
// only added to the probe set when the user opted into online probes, so it
// never runs in offline mode.
type PublicIPProbe struct{}

func (PublicIPProbe) Name() string { return "public_ip_probe" }

func (p PublicIPProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	ip, err := network.PublicIP(ctx)
	if err != nil {
		return res, err
	}
	res.Evidence["public_ip"] = ip
	if network.IsCGNAT(ip) {
		res.Evidence["cgnat"] = true
		res.Hints = appendUnique(res.Hints, models.TypeMobile)
		res.Hints = appendUnique(res.Hints, models.TypeFWA)
		res.Confidence = 0.6
	} else {
		res.Evidence["cgnat"] = false
		res.Confidence = 0.3
	}
	return res, nil
}
