package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type ARPNeighborProbe struct{}

func (ARPNeighborProbe) Name() string            { return "arp_neighbor_probe" }
func (ARPNeighborProbe) SafeModeAllowed() bool   { return true }
func (ARPNeighborProbe) NormalModeAllowed() bool { return true }
func (ARPNeighborProbe) DeepModeAllowed() bool   { return true }

func (p ARPNeighborProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	select {
	case <-ctx.Done():
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusFailed}, ctx.Err()
	default:
	}
	entries, _ := input.Metadata["arp_entries"].([]model.Device)
	if len(entries) == 0 {
		return skippedResult(p.Name(), "no platform ARP adapter supplied entries"), nil
	}
	result := model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Devices: entries}
	for _, d := range entries {
		ev := baseEvidence(p.Name(), d.ID, model.EvidenceWeak, 0.35, "ARP table observed this local IPv4 neighbor.", map[string]any{"device": d.ID})
		result.Evidence = append(result.Evidence, ev)
	}
	return result, nil
}
