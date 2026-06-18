package deviceintel

import (
	"fmt"
	"sort"

	"github.com/thekiran/iad/pkg/models"
)

func convertEdges(report models.ScanReport, store *EvidenceStore) []models.DeviceIntelEdge {
	var out []models.DeviceIntelEdge
	for _, edge := range report.Edges {
		di := models.DeviceIntelEdge{
			ID:              edge.ID,
			Source:          edge.Source,
			Target:          edge.Target,
			Type:            mapEdgeType(edge.Type),
			Confidence:      edgeConfidence(edge),
			ConfidenceLabel: confidenceLabel(edgeConfidence(edge)),
			EvidenceIDs:     sortedUnique(edge.EvidenceIDs),
			Reason:          edge.Reason,
			Inferred:        isInferredEdge(edge.Type),
			Physical:        isPhysicalEdge(edge.Type),
		}
		out = append(out, di)
		addEdgeHint(store, di)
	}
	out = append(out, gatewayChainEdges(report, store, existingEdgeKeys(out))...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		if out[i].Target != out[j].Target {
			return out[i].Target < out[j].Target
		}
		return out[i].Type < out[j].Type
	})
	return out
}

func mapEdgeType(t string) string {
	switch t {
	case models.EdgeGatewayDefault:
		return models.DeviceEdgeDefaultGatewayRoute
	case models.EdgeInferredL2:
		return models.DeviceEdgeSameSubnetInferred
	case models.EdgeRouteHop:
		return models.DeviceEdgeTracerouteHop
	case models.EdgeDirectLLDP:
		return models.DeviceEdgeLLDPPhysicalNeighbor
	case models.EdgeDirectCDP:
		return models.DeviceEdgeCDPPhysicalNeighbor
	case models.EdgeSNMPBridge:
		return models.DeviceEdgeSNMPBridgeFDB
	default:
		return t
	}
}

func edgeConfidence(edge models.TopologyEdge) float64 {
	switch edge.Type {
	case models.EdgeInferredL2:
		return minFloat(edge.Confidence, 0.35)
	case models.EdgeGatewayDefault:
		return minFloat(maxFloat(edge.Confidence, 0.45), 0.60)
	case models.EdgeRouteHop:
		return minFloat(maxFloat(edge.Confidence, 0.55), 0.70)
	default:
		return edge.Confidence
	}
}

func isInferredEdge(t string) bool {
	switch t {
	case models.EdgeGatewayDefault, models.EdgeInferredL2, models.EdgeRouteHop:
		return true
	default:
		return false
	}
}

func isPhysicalEdge(t string) bool {
	switch t {
	case models.EdgeDirectLLDP, models.EdgeDirectCDP, models.EdgeSNMPBridge:
		return true
	default:
		return false
	}
}

func addEdgeHint(store *EvidenceStore, edge models.DeviceIntelEdge) {
	for _, id := range []string{edge.Source, edge.Target} {
		d := store.Devices[id]
		if d == nil {
			continue
		}
		peer := edge.Target
		if id == edge.Target {
			peer = edge.Source
		}
		d.Topology.EdgeHints = append(d.Topology.EdgeHints, models.DeviceEdgeHint{
			EdgeID:          edge.ID,
			Type:            edge.Type,
			Peer:            peer,
			Confidence:      edge.Confidence,
			ConfidenceLabel: edge.ConfidenceLabel,
			EvidenceIDs:     edge.EvidenceIDs,
		})
		if edge.Physical {
			d.Topology.PhysicalAdjacencyProven = true
			d.Topology.InferredOnly = false
		} else if len(d.Topology.EdgeHints) > 0 && !d.Topology.PhysicalAdjacencyProven {
			d.Topology.InferredOnly = true
		}
	}
}

func gatewayChainEdges(report models.ScanReport, store *EvidenceStore, existing map[string]bool) []models.DeviceIntelEdge {
	state := gatewayChainState(report)
	if state == nil || len(state.PrivateHops) < 2 {
		return nil
	}
	var out []models.DeviceIntelEdge
	for i := 0; i+1 < len(state.PrivateHops); i++ {
		from := deviceID(state.PrivateHops[i].IP)
		to := deviceID(state.PrivateHops[i+1].IP)
		key := edgeKey(from, to, models.DeviceEdgeUpstreamPrivate)
		if existing[key] {
			continue
		}
		id := fmt.Sprintf("di-edge-upstream-%d", i+1)
		evIDs := sortedUnique([]string{state.PrivateHops[i].EvidenceID, state.PrivateHops[i+1].EvidenceID})
		edge := models.DeviceIntelEdge{
			ID:              id,
			Source:          from,
			Target:          to,
			Type:            models.DeviceEdgeUpstreamPrivate,
			Confidence:      maxFloat(state.Confidence, 0.60),
			ConfidenceLabel: confidenceLabel(maxFloat(state.Confidence, 0.60)),
			EvidenceIDs:     evIDs,
			Reason:          "Traceroute observed a later private gateway after the default gateway; this is a path hint, not physical cabling proof.",
			Inferred:        true,
			Physical:        false,
		}
		out = append(out, edge)
		addEdgeHint(store, edge)
		existing[key] = true
	}
	return out
}

func existingEdgeKeys(edges []models.DeviceIntelEdge) map[string]bool {
	out := map[string]bool{}
	for _, edge := range edges {
		out[edgeKey(edge.Source, edge.Target, edge.Type)] = true
	}
	return out
}

func edgeKey(source, target, typ string) string {
	return source + "|" + target + "|" + typ
}

func gatewayChainState(report models.ScanReport) *models.GatewayChainState {
	if report.AccessClassification == nil || report.AccessClassification.DetectedNetworkContext == nil {
		return nil
	}
	return report.AccessClassification.DetectedNetworkContext.GatewayChainState
}
