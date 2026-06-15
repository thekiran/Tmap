package probes

import (
	"context"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

// DNSProbe lists the configured DNS servers. This is contextual evidence (it
// can corroborate an ISP) and runs without contacting any external service.
type DNSProbe struct{}

func (DNSProbe) Name() string { return "dns_probe" }

func (p DNSProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	servers, err := network.DNSServers(ctx)
	if err != nil {
		return res, err
	}
	res.Evidence["servers"] = servers
	res.Confidence = 0.2
	return res, nil
}
