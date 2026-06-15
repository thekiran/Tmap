package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type TracerouteISPPathProbe struct{}

func (TracerouteISPPathProbe) Name() string            { return "traceroute_isp_path_probe" }
func (TracerouteISPPathProbe) SafeModeAllowed() bool   { return false }
func (TracerouteISPPathProbe) NormalModeAllowed() bool { return false }
func (TracerouteISPPathProbe) DeepModeAllowed() bool   { return true }

func (p TracerouteISPPathProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	select {
	case <-ctx.Done():
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusFailed}, ctx.Err()
	default:
	}
	hops, _ := input.Metadata["route_hops"].([]model.RouteHop)
	if len(hops) == 0 {
		return skippedResult(p.Name(), "no traceroute adapter supplied route hops"), nil
	}
	result := model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Raw: map[string]any{"route_hops": hops}}
	for _, hop := range hops {
		target := hop.IP
		ev := baseEvidence(p.Name(), target, model.EvidenceMedium, 0.50, "Traceroute observed a route hop; this is not proof of physical ISP infrastructure.", map[string]any{"hop": hop})
		result.Evidence = append(result.Evidence, ev)
		dtype := model.DeviceTypeISPHop
		if hop.Private {
			dtype = model.DeviceTypeRouter
		}
		result.Devices = append(result.Devices, model.Device{ID: "hop_" + target, IPAddresses: []string{target}, DeviceType: dtype, Confidence: 0.50, Evidence: []model.Evidence{ev}, Inferred: false})
	}
	return result, nil
}
