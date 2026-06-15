// Package topology turns evidence-bearing discovery results into a device graph.
// It separates inference from proof: an edge's type and confidence record *how*
// the link was established, and the builder never emits a physical-adjacency edge
// (LLDP/CDP/SNMP) without the matching evidence.
package topology

import "github.com/thekiran/iad/pkg/models"

// Confidence bands for each edge type. These follow the project's evidence
// hierarchy: protocol-reported adjacency (LLDP/CDP) is near-certain; a bridge
// table is strong; a route hop is moderate; same-subnet inference is weak; a
// default-gateway link is a fallback only. Callers may adjust within the band
// (e.g. snmp_bridge with multiple corroborating FDB entries → higher), but never
// promote an inferred edge into "very high".
const (
	confDirectLLDP     = 0.95
	confDirectCDP      = 0.95
	confSNMPBridgeHigh = 0.80
	confSNMPBridgeLow  = 0.60
	confRouteHop       = 0.60
	confInferredL2High = 0.50
	confInferredL2Low  = 0.40
	confGatewayDefault = 0.30
)

// BaseConfidence returns the default confidence for an edge type. It is the
// starting point; the builder may refine it using corroborating evidence but is
// bounded by the type (an inferred edge can never reach the LLDP/CDP band).
func BaseConfidence(edgeType string) float64 {
	switch edgeType {
	case models.EdgeDirectLLDP:
		return confDirectLLDP
	case models.EdgeDirectCDP:
		return confDirectCDP
	case models.EdgeSNMPBridge:
		return confSNMPBridgeHigh
	case models.EdgeRouteHop:
		return confRouteHop
	case models.EdgeInferredL2:
		return confInferredL2Low
	case models.EdgeGatewayDefault:
		return confGatewayDefault
	default:
		return 0.0
	}
}

// MaxConfidence is the ceiling an edge type may reach regardless of
// corroboration. Inference can never masquerade as proof.
func MaxConfidence(edgeType string) float64 {
	switch edgeType {
	case models.EdgeDirectLLDP, models.EdgeDirectCDP:
		return 0.98
	case models.EdgeSNMPBridge:
		return confSNMPBridgeHigh
	case models.EdgeRouteHop:
		return confRouteHop
	case models.EdgeInferredL2:
		return confInferredL2High
	case models.EdgeGatewayDefault:
		return confGatewayDefault
	default:
		return 0.0
	}
}

// clampConfidence keeps a confidence within [0, MaxConfidence(edgeType)].
func clampConfidence(edgeType string, c float64) float64 {
	max := MaxConfidence(edgeType)
	switch {
	case c < 0:
		return 0
	case c > max:
		return max
	default:
		return c
	}
}

// ConfidenceLabel buckets a numeric confidence into a human-facing label.
func ConfidenceLabel(c float64) string {
	switch {
	case c >= 0.90:
		return models.ConfVeryHigh
	case c >= 0.70:
		return models.ConfHigh
	case c >= 0.45:
		return models.ConfMedium
	case c >= 0.25:
		return models.ConfLow
	default:
		return models.ConfVeryLow
	}
}

// IsPhysicalEvidenceEdge reports whether an edge type represents proven physical
// adjacency (as opposed to inference). Used to set the report's InferredOnly flag.
func IsPhysicalEvidenceEdge(edgeType string) bool {
	switch edgeType {
	case models.EdgeDirectLLDP, models.EdgeDirectCDP, models.EdgeSNMPBridge:
		return true
	default:
		return false
	}
}
