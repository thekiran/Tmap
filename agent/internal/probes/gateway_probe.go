package probes

import (
	"context"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

// GatewayProbe discovers the default gateway IP. On its own this is weak
// evidence, but the address is recorded for the report and reused by other
// probes (e.g. offline latency targets the gateway).
type GatewayProbe struct{}

func (GatewayProbe) Name() string { return "gateway_probe" }

func (p GatewayProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	gw, err := network.Gateway()
	if err != nil {
		return res, err
	}
	res.Evidence["gateway"] = gw.String()
	res.Confidence = 0.3
	return res, nil
}
