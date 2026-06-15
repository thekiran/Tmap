package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type RouteTableProbe struct{}

func (RouteTableProbe) Name() string            { return "route_table_probe" }
func (RouteTableProbe) SafeModeAllowed() bool   { return true }
func (RouteTableProbe) NormalModeAllowed() bool { return true }
func (RouteTableProbe) DeepModeAllowed() bool   { return true }

func (p RouteTableProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	select {
	case <-ctx.Done():
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusFailed}, ctx.Err()
	default:
	}
	gw, _ := input.Metadata["default_gateway"].(string)
	if gw == "" {
		return skippedResult(p.Name(), "no default gateway supplied by platform adapter"), nil
	}
	ev := baseEvidence(p.Name(), gw, model.EvidenceStrong, 0.65, "Default route points to this private gateway.", map[string]any{"default_gateway": gw})
	dev := model.Device{
		ID:          "ip_" + gw,
		IPAddresses: []string{gw},
		DeviceType:  model.DeviceTypeRouter,
		Roles:       []model.DeviceRole{model.RoleDefaultGateway},
		Confidence:  0.65,
		Evidence:    []model.Evidence{ev},
	}
	return model.ProbeResult{
		ProbeName:  p.Name(),
		Status:     model.ProbeStatusSuccess,
		Confidence: 0.65,
		Evidence:   []model.Evidence{ev},
		Devices:    []model.Device{dev},
		Raw:        map[string]any{"default_gateway": gw},
	}, nil
}
