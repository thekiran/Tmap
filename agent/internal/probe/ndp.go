package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type NDPNeighborProbe struct{}

func (NDPNeighborProbe) Name() string            { return "ndp_neighbor_probe" }
func (NDPNeighborProbe) SafeModeAllowed() bool   { return true }
func (NDPNeighborProbe) NormalModeAllowed() bool { return true }
func (NDPNeighborProbe) DeepModeAllowed() bool   { return true }

func (p NDPNeighborProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	_ = ctx
	if entries, _ := input.Metadata["ndp_entries"].([]model.Device); len(entries) > 0 {
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Devices: entries}, nil
	}
	return skippedResult(p.Name(), "no platform NDP adapter supplied entries"), nil
}
