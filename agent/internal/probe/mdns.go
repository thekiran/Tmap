package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type MDNSProbe struct{}

func (MDNSProbe) Name() string            { return "mdns_probe" }
func (MDNSProbe) SafeModeAllowed() bool   { return false }
func (MDNSProbe) NormalModeAllowed() bool { return true }
func (MDNSProbe) DeepModeAllowed() bool   { return true }

func (p MDNSProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	_ = ctx
	if devices, _ := input.Metadata["mdns_devices"].([]model.Device); len(devices) > 0 {
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Devices: devices}, nil
	}
	return skippedResult(p.Name(), "no mDNS adapter supplied service records"), nil
}
