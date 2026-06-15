package topology

import (
	"fmt"

	"github.com/thekiran/iad/pkg/models"
)

// newEdge constructs a TopologyEdge with a deterministic ID, the type's confidence
// (optionally overridden, but always clamped to the type's ceiling), a derived
// label, deduped evidence IDs, and a mandatory human reason. A panic-free
// contract: every edge that leaves this constructor is well-formed.
func newEdge(source, target, edgeType, reason string, confidenceOverride float64, evidenceIDs []string) models.TopologyEdge {
	conf := BaseConfidence(edgeType)
	if confidenceOverride > 0 {
		conf = confidenceOverride
	}
	conf = clampConfidence(edgeType, conf)
	return models.TopologyEdge{
		ID:              edgeID(source, target, edgeType),
		Source:          source,
		Target:          target,
		Type:            edgeType,
		Confidence:      conf,
		ConfidenceLabel: ConfidenceLabel(conf),
		EvidenceIDs:     dedupSorted(evidenceIDs),
		Reason:          reason,
	}
}

// edgeID is stable for a given (source, target, type) triple so re-running a scan
// over identical input yields identical edge IDs.
func edgeID(source, target, edgeType string) string {
	return fmt.Sprintf("edge-%s--%s--%s", source, target, shortEdgeType(edgeType))
}

func shortEdgeType(edgeType string) string {
	switch edgeType {
	case models.EdgeDirectLLDP:
		return "lldp"
	case models.EdgeDirectCDP:
		return "cdp"
	case models.EdgeSNMPBridge:
		return "snmp"
	case models.EdgeRouteHop:
		return "route"
	case models.EdgeInferredL2:
		return "l2"
	case models.EdgeGatewayDefault:
		return "gw"
	default:
		return "x"
	}
}
