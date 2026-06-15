package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type LLDPCDPPassiveProbe struct{}

func (LLDPCDPPassiveProbe) Name() string            { return "lldp_cdp_passive_probe" }
func (LLDPCDPPassiveProbe) SafeModeAllowed() bool   { return false }
func (LLDPCDPPassiveProbe) NormalModeAllowed() bool { return false }
func (LLDPCDPPassiveProbe) DeepModeAllowed() bool   { return true }

func (p LLDPCDPPassiveProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	_ = ctx
	edges, _ := input.Metadata["lldp_edges"].([]model.TopologyEdge)
	if len(edges) == 0 {
		edges, _ = input.Metadata["cdp_edges"].([]model.TopologyEdge)
	}
	if len(edges) > 0 {
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Edges: edges}, nil
	}
	return skippedResult(p.Name(), "no passive LLDP/CDP adapter supplied frames"), nil
}
