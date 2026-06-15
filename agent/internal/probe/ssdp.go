package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type SSDPUPnPProbe struct{}

func (SSDPUPnPProbe) Name() string            { return "ssdp_upnp_probe" }
func (SSDPUPnPProbe) SafeModeAllowed() bool   { return false }
func (SSDPUPnPProbe) NormalModeAllowed() bool { return true }
func (SSDPUPnPProbe) DeepModeAllowed() bool   { return true }

func (p SSDPUPnPProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	_ = ctx
	if devices, _ := input.Metadata["ssdp_devices"].([]model.Device); len(devices) > 0 {
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Devices: devices}, nil
	}
	return skippedResult(p.Name(), "no SSDP adapter supplied responses"), nil
}
