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
		Layer:           edgeLayer(edgeType),
		Relationship:    edgeRelationship(edgeType),
		Physical:        IsPhysicalEvidenceEdge(edgeType),
		Inferred:        edgeInferred(edgeType),
		Confidence:      conf,
		ConfidenceLabel: ConfidenceLabel(conf),
		ProofSource:     edgeProofSource(edgeType),
		UILineStyle:     edgeLineStyle(edgeType, conf),
		EvidenceIDs:     dedupSorted(evidenceIDs),
		Reason:          reason,
	}
}

// edgeID is stable for a given (source, target, type) triple so re-running a scan
// over identical input yields identical edge IDs.
func edgeID(source, target, edgeType string) string {
	return fmt.Sprintf("edge-%s--%s--%s", source, target, shortEdgeType(edgeType))
}

func edgeLayer(edgeType string) string {
	switch edgeType {
	case models.EdgeDirectLLDP, models.EdgeDirectCDP, models.EdgeSNMPBridge, models.EdgeInferredL2:
		return "L2"
	case models.EdgeRouteHop, models.EdgeGatewayDefault:
		return "L3"
	default:
		return "unknown"
	}
}

func edgeRelationship(edgeType string) string {
	switch edgeType {
	case models.EdgeDirectLLDP:
		return "lldp_neighbor"
	case models.EdgeDirectCDP:
		return "cdp_neighbor"
	case models.EdgeSNMPBridge:
		return "snmp_bridge_fdb"
	case models.EdgeRouteHop:
		return "traceroute_hop"
	case models.EdgeGatewayDefault:
		return "default_gateway"
	case models.EdgeInferredL2:
		return "same_subnet"
	default:
		return "unknown"
	}
}

func edgeInferred(edgeType string) bool {
	switch edgeType {
	case models.EdgeDirectLLDP, models.EdgeDirectCDP, models.EdgeSNMPBridge, models.EdgeRouteHop:
		return false
	default:
		return true
	}
}

func edgeProofSource(edgeType string) string {
	switch edgeType {
	case models.EdgeDirectLLDP:
		return "lldp"
	case models.EdgeDirectCDP:
		return "cdp"
	case models.EdgeSNMPBridge:
		return "snmp_bridge_fdb"
	case models.EdgeRouteHop:
		return "traceroute"
	case models.EdgeGatewayDefault:
		return "route_table"
	case models.EdgeInferredL2:
		return "same_subnet_arp_ping"
	default:
		return "unknown"
	}
}

func edgeLineStyle(edgeType string, confidence float64) string {
	if IsPhysicalEvidenceEdge(edgeType) {
		return "solid"
	}
	if edgeType == models.EdgeInferredL2 || confidence < 0.50 {
		return "dotted"
	}
	return "dashed"
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
