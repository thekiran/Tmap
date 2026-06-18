package topology

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestTopologyEdgesAreNotPhysicalUnlessProven(t *testing.T) {
	result := Build(BuildInput{
		AgentID:                 "dev-agent",
		GatewayID:               "dev-gw",
		GatewayRouteEvidenceIDs: []string{"ev-gw"},
		Devices: []models.Device{
			deviceForEdgeTest("dev-agent", "192.168.1.10"),
			deviceForEdgeTest("dev-gw", "192.168.1.1"),
			deviceForEdgeTest("dev-peer", "192.168.1.20"),
		},
		L2Peers: []string{"dev-peer"},
		RouteHops: []RouteHop{
			{FromID: "dev-gw", ToID: "dev-peer", EvidenceIDs: []string{"ev-route"}},
		},
	})

	for _, edge := range result.Edges {
		if edge.Type == models.EdgeDirectLLDP || edge.Type == models.EdgeDirectCDP || edge.Type == models.EdgeSNMPBridge {
			continue
		}
		if edge.Physical {
			t.Fatalf("edge %s type %s marked physical=true without physical proof", edge.ID, edge.Type)
		}
		if edge.Layer == "" || edge.Relationship == "" || edge.UILineStyle == "" {
			t.Fatalf("edge %s missing frontend contract fields: %#v", edge.ID, edge)
		}
	}
}

func TestProvenNeighborEdgeIsPhysical(t *testing.T) {
	result := Build(BuildInput{
		AgentID:   "dev-agent",
		GatewayID: "dev-gw",
		Devices: []models.Device{
			deviceForEdgeTest("dev-agent", "192.168.1.10"),
			deviceForEdgeTest("dev-gw", "192.168.1.1"),
		},
		Neighbors: []Neighbor{
			{FromID: "dev-agent", ToID: "dev-gw", Kind: models.EdgeDirectLLDP, EvidenceIDs: []string{"ev-lldp"}, Confidence: 0.95},
		},
	})
	if len(result.Edges) == 0 {
		t.Fatal("expected a proven neighbor edge")
	}
	edge := result.Edges[0]
	if !edge.Physical {
		t.Fatalf("LLDP edge physical = false: %#v", edge)
	}
	if edge.Relationship != "lldp_neighbor" || edge.Layer != "L2" || edge.UILineStyle != "solid" {
		t.Fatalf("unexpected LLDP edge contract fields: %#v", edge)
	}
}

func deviceForEdgeTest(id, ip string) models.Device {
	return models.Device{
		ID:        id,
		Addresses: []models.IPAddress{{IP: ip, Version: 4}},
	}
}
