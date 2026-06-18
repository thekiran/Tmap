package topology

import (
	"github.com/thekiran/iad/pkg/models"
)

// RouteHop is an adjacent pair of devices observed in a traceroute: traffic from
// FromID was seen to reach ToID at the next hop.
type RouteHop struct {
	FromID      string
	ToID        string
	EvidenceIDs []string
}

// Neighbor is a link discovered from real link-layer protocol evidence
// (LLDP/CDP) or an SNMP bridge/FDB table. Kind must be one of the physical edge
// types; the builder refuses anything else so inference can never be smuggled in
// as a proven neighbour.
type Neighbor struct {
	FromID      string
	ToID        string
	Kind        string // models.EdgeDirectLLDP | EdgeDirectCDP | EdgeSNMPBridge
	Confidence  float64
	Reason      string
	EvidenceIDs []string
}

// BuildInput is the evidence-backed view the builder turns into a graph. Discovery
// (or any other producer) is responsible for only populating fields it can prove.
type BuildInput struct {
	AgentID   string
	GatewayID string // agent's default gateway device id ("" when unknown)
	Devices   []models.Device

	// GatewayRouteEvidenceIDs back the agent→gateway gateway_default edge.
	GatewayRouteEvidenceIDs []string
	// RouteHops are adjacent traceroute hops (medium confidence).
	RouteHops []RouteHop
	// L2Peers are device ids observed on the agent's own subnet (ARP/ping). They
	// yield inferred_l2 edges — adjacency is inferred from shared subnet, not proven.
	L2Peers []string
	// Neighbors are LLDP/CDP/SNMP-proven adjacencies (only when such evidence exists).
	Neighbors []Neighbor
}

// BuildResult is the assembled topology: devices with roles applied, edges, and
// summary flags.
type BuildResult struct {
	Devices             []models.Device
	Edges               []models.TopologyEdge
	InferredOnly        bool
	HighConfidenceEdges int
}

// Build constructs the topology graph. It emits, in order of trust:
//   - direct_lldp / direct_cdp / snmp_bridge from proven neighbours,
//   - route_hop from adjacent traceroute hops,
//   - gateway_default from the agent's default route,
//   - inferred_l2 from same-subnet peers (lowest, clearly labelled as inferred).
//
// Devices get evidence-based roles (gateway only when it is the proven default
// gateway; router only when a hop beyond it was observed).
func Build(in BuildInput) BuildResult {
	ids := make([]string, 0, len(in.Devices))
	for _, d := range in.Devices {
		ids = append(ids, d.ID)
	}
	g := NewGraph(ids)

	// Which devices forward traffic (are the source of a route hop)? Used for the
	// "router" role.
	forwards := map[string]bool{}
	for _, h := range in.RouteHops {
		forwards[h.FromID] = true
	}
	if in.GatewayID != "" && len(in.RouteHops) > 0 {
		// The default gateway forwards if any hop chain exists.
		forwards[in.GatewayID] = true
	}

	// 1) Proven neighbours (LLDP/CDP/SNMP) — the only edges allowed to claim
	// physical adjacency.
	for _, n := range in.Neighbors {
		if !IsPhysicalEvidenceEdge(n.Kind) {
			continue // refuse to record a "proven" neighbour without a physical kind
		}
		if len(n.EvidenceIDs) == 0 {
			continue // never fake LLDP/CDP/SNMP: no evidence → no edge
		}
		reason := n.Reason
		if reason == "" {
			reason = neighborReason(n.Kind)
		}
		g.AddEdge(newEdge(n.FromID, n.ToID, n.Kind, reason, n.Confidence, n.EvidenceIDs))
	}

	// 2) Route hops (medium).
	for _, h := range in.RouteHops {
		g.AddEdge(newEdge(h.FromID, h.ToID, models.EdgeRouteHop,
			"Adjacent hops observed in a traceroute; routed (L3) adjacency, not necessarily a direct physical link.",
			0, h.EvidenceIDs))
	}

	// 3) Gateway default (fallback link from the agent to its gateway).
	if in.AgentID != "" && in.GatewayID != "" {
		g.AddEdge(newEdge(in.AgentID, in.GatewayID, models.EdgeGatewayDefault,
			"The agent's default route points at this gateway; this is the route to the rest of the network, not proof of a direct cable.",
			0, in.GatewayRouteEvidenceIDs))
	}

	// 4) Inferred L2 (same subnet) — lowest confidence, explicitly inferred.
	// Anchor same-subnet peers at the gateway when one is known (in a flat home
	// LAN the router/gateway is the common L2 point, so the map reads as a star
	// around it); fall back to the agent when there is no gateway. This is still
	// an inferred relationship — never a claimed physical cable.
	hub := in.GatewayID
	if hub == "" {
		hub = in.AgentID
	}
	for _, peer := range in.L2Peers {
		if peer == in.AgentID || peer == in.GatewayID {
			continue // gateway already linked via gateway_default; skip self
		}
		g.AddEdge(newEdge(hub, peer, models.EdgeInferredL2,
			"On the same subnet (observed via ARP/ping); L2 adjacency is inferred from the shared broadcast domain and anchored at the gateway, not proven.",
			0, nil))
	}

	edges := g.Edges()

	// Apply evidence-based roles to the devices.
	devices := make([]models.Device, len(in.Devices))
	copy(devices, in.Devices)
	for i := range devices {
		d := &devices[i]
		isGW := d.ID == in.GatewayID
		d.IsGateway = isGW
		d.IsAgent = d.ID == in.AgentID
		d.Roles = ClassifyRoles(RoleInput{
			IsAgent:         d.IsAgent,
			IsDefaultGW:     isGW,
			ForwardsTraffic: forwards[d.ID],
		})
	}

	inferredOnly := true
	high := 0
	for _, e := range edges {
		if IsPhysicalEvidenceEdge(e.Type) {
			inferredOnly = false
		}
		if e.Confidence >= 0.70 {
			high++
		}
	}

	return BuildResult{
		Devices:             devices,
		Edges:               edges,
		InferredOnly:        inferredOnly,
		HighConfidenceEdges: high,
	}
}

func neighborReason(kind string) string {
	switch kind {
	case models.EdgeDirectLLDP:
		return "Neighbour reported via LLDP; this is a directly attached link-layer neighbour."
	case models.EdgeDirectCDP:
		return "Neighbour reported via CDP; this is a directly attached link-layer neighbour."
	case models.EdgeSNMPBridge:
		return "Forwarding/bridge (FDB) table entry read via SNMP indicates a link-layer association."
	default:
		return "Link-layer neighbour."
	}
}
