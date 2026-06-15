package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type PingSweepProbe struct{}

func (PingSweepProbe) Name() string            { return "ping_sweep_probe" }
func (PingSweepProbe) SafeModeAllowed() bool   { return false }
func (PingSweepProbe) NormalModeAllowed() bool { return true }
func (PingSweepProbe) DeepModeAllowed() bool   { return true }

func (p PingSweepProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	_ = ctx
	if hits, _ := input.Metadata["ping_hits"].([]model.Device); len(hits) > 0 {
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Devices: hits}, nil
	}
	return skippedResult(p.Name(), "no ICMP sweeper configured; private-only ping sweep skipped"), nil
}
