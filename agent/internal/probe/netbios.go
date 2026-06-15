package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type NetBIOSLLMNRProbe struct{}

func (NetBIOSLLMNRProbe) Name() string            { return "netbios_llmnr_probe" }
func (NetBIOSLLMNRProbe) SafeModeAllowed() bool   { return false }
func (NetBIOSLLMNRProbe) NormalModeAllowed() bool { return true }
func (NetBIOSLLMNRProbe) DeepModeAllowed() bool   { return true }

func (p NetBIOSLLMNRProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	_ = ctx
	if devices, _ := input.Metadata["netbios_devices"].([]model.Device); len(devices) > 0 {
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Devices: devices}, nil
	}
	return skippedResult(p.Name(), "no NetBIOS/LLMNR adapter supplied names"), nil
}
