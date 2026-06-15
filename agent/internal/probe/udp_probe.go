package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type UDPProbe struct{}

func (UDPProbe) Name() string            { return "udp_probe" }
func (UDPProbe) SafeModeAllowed() bool   { return false }
func (UDPProbe) NormalModeAllowed() bool { return true }
func (UDPProbe) DeepModeAllowed() bool   { return true }

func (p UDPProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	_ = ctx
	if devices, _ := input.Metadata["udp_observations"].([]model.Device); len(devices) > 0 {
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Devices: devices}, nil
	}
	return skippedResult(p.Name(), "no UDP discovery adapter configured"), nil
}
