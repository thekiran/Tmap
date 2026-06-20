package topology

import (
	"strings"
	"testing"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

func TestBuildRichTopologyMergesRootAndDeviceIntel(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	report := models.ScanReport{
		CreatedAt: now,
		Devices: []models.Device{
			{
				ID: "dev-agent", Hostname: "pc", IsAgent: true, Confidence: 1,
				Addresses:   []models.IPAddress{{IP: "192.168.31.10", Version: 4}},
				EvidenceIDs: []string{"ev-agent"},
			},
			{
				ID: "dev-gw", IsGateway: true, Roles: []string{models.RoleGateway, models.RoleRouter}, Confidence: 0.9,
				Addresses:   []models.IPAddress{{IP: "192.168.31.1", Version: 4}},
				EvidenceIDs: []string{"ev-gw"},
			},
		},
		Edges: []models.TopologyEdge{
			{
				ID: "edge-gw", Source: "dev-agent", Target: "dev-gw", Type: models.EdgeGatewayDefault,
				Relationship: "default_gateway", Confidence: 0.55, EvidenceIDs: []string{"ev-gw"},
			},
		},
		Evidence: []models.Evidence{
			{ID: "ev-agent", Kind: "interface", Source: "interface_probe", Summary: "agent interface", Timestamp: now},
			{ID: "ev-gw", Kind: "gateway_route", Source: "gateway_probe", Summary: "default gateway", Timestamp: now},
			{ID: "ev-ap", Kind: "client_table", Source: "ap_api", Summary: "AP reported associated client", Timestamp: now},
		},
		DeviceIntel: &models.DeviceIntelReport{
			Devices: []models.DeviceIntelDevice{
				{
					ID: "dev-upstream", IPAddresses: []string{"192.168.1.1"},
					Roles:      []string{models.DeviceRolePossibleCPE, models.DeviceRoleUpstreamGateway},
					DeviceType: models.DeviceTypeGuess{Primary: models.DeviceTypeUpstreamCPE},
					Topology:   models.DeviceTopologyFacts{IsUpstreamGatewayCandidate: true},
					Confidence: 0.65,
				},
				{
					ID: "dev-client", IPAddresses: []string{"192.168.31.44"},
					DeviceType: models.DeviceTypeGuess{Primary: models.DeviceTypeUnknown},
					Confidence: 0.50,
				},
			},
			Edges: []models.DeviceIntelEdge{
				{
					ID: "edge-upstream", Source: "dev-gw", Target: "dev-upstream",
					Type: models.DeviceEdgeUpstreamPrivate, Confidence: 0.65, Inferred: true,
					Reason: "private upstream gateway observed",
				},
				{
					ID: "edge-ap-client", Source: "dev-gw", Target: "dev-client",
					Type: models.DeviceEdgeWiFiAssociation, Confidence: 0.90, EvidenceIDs: []string{"ev-ap"},
					Reason: "AP client table reported station",
				},
			},
		},
	}

	rich := BuildRichTopology(report)
	if rich.SchemaVersion != models.RichTopologySchema {
		t.Fatalf("schema = %q", rich.SchemaVersion)
	}
	if rich.UI.RootNodeID != "dev-gw" {
		t.Fatalf("root = %q, want dev-gw", rich.UI.RootNodeID)
	}
	if len(rich.Nodes) != 4 {
		t.Fatalf("nodes = %d, want 4: %#v", len(rich.Nodes), rich.Nodes)
	}
	if node := findRichNode(rich, "dev-upstream"); node == nil || node.DeviceRole != models.NodeRoleRouter {
		t.Fatalf("upstream node not classified as router/CPE: %#v", node)
	}
	if edge := findRichEdge(rich, "edge-gw"); edge == nil || edge.Type != models.RichEdgeGatewayLink {
		t.Fatalf("gateway edge wrong: %#v", edge)
	}
	if edge := findRichEdge(rich, "edge-upstream"); edge == nil || edge.Type != models.RichEdgeWeakInferred {
		t.Fatalf("upstream edge should be weak inferred, got %#v", edge)
	}
	if edge := findRichEdge(rich, "edge-ap-client"); edge == nil || edge.Type != models.RichEdgeReportedByAP {
		t.Fatalf("AP-reported edge should be reported-by-ap, got %#v", edge)
	}
	if len(findRichEdge(rich, "edge-gw").Evidence) == 0 {
		t.Fatalf("gateway edge missing embedded evidence")
	}
}

func TestBuildTopologyV2AddsEdgeExplanationsWarningsAndRawObservations(t *testing.T) {
	now := time.Unix(200, 0).UTC()
	report := models.ScanReport{
		SchemaVersion: models.TopologyReportSchema,
		CreatedAt:     now,
		Devices: []models.Device{
			{
				ID: "dev-192.168.1.10", Hostname: "pc", IsAgent: true, Confidence: 1,
				Addresses:   []models.IPAddress{{IP: "192.168.1.10", Version: 4}},
				EvidenceIDs: []string{"ev-agent"},
			},
			{
				ID: "dev-192.168.1.1", IsGateway: true, Confidence: 0.9,
				Addresses:   []models.IPAddress{{IP: "192.168.1.1", Version: 4}},
				EvidenceIDs: []string{"ev-gw"},
			},
			{
				ID: "dev-192.168.1.22", Confidence: 0.5,
				Addresses:   []models.IPAddress{{IP: "192.168.1.22", Version: 4}},
				EvidenceIDs: []string{"ev-arp"},
			},
		},
		Edges: []models.TopologyEdge{
			{
				ID: "edge-l2", Source: "dev-192.168.1.1", Target: "dev-192.168.1.22",
				Type: models.EdgeInferredL2, Layer: "L2", Relationship: "same_subnet",
				Confidence: 0.40, EvidenceIDs: []string{"ev-arp"}, Reason: "same subnet",
			},
		},
		Evidence: []models.Evidence{
			{ID: "ev-agent", Kind: "interface", Source: "interface_probe", Summary: "agent", Timestamp: now},
			{ID: "ev-gw", Kind: "gateway_route", Source: "gateway_probe", Summary: "gateway", Timestamp: now, Data: map[string]string{"ip": "192.168.1.1"}},
			{ID: "ev-arp", Kind: "arp_table", Source: "arp_table", Summary: "arp", Timestamp: now, Data: map[string]string{"ip": "192.168.1.22", "mac": "aa:bb:cc:dd:ee:ff", "payload": "must-not-appear"}},
		},
	}

	rich := BuildRichTopology(report)
	v2 := BuildTopologyV2(report, rich)
	if v2.SchemaVersion != models.TopologyReportSchema {
		t.Fatalf("schema = %q, want %q", v2.SchemaVersion, models.TopologyReportSchema)
	}
	edge := findV2Edge(v2, "edge-l2")
	if edge == nil {
		t.Fatal("missing v2 edge")
	}
	if edge.Type != models.RichEdgeSubnetInferred || edge.Explanation == "" || len(edge.Warnings) == 0 {
		t.Fatalf("edge not enriched with inferred explanation/warnings: %#v", edge)
	}
	raw := BuildRawObservations(report)
	if len(raw) != len(report.Evidence) {
		t.Fatalf("raw observations = %d, want %d", len(raw), len(report.Evidence))
	}
	for _, obs := range raw {
		if _, ok := obs.Metadata["payload"]; ok {
			t.Fatalf("raw observation stored payload metadata: %#v", obs)
		}
	}
}

func TestStableDeviceIDsAreEvidenceKeysNotIndexes(t *testing.T) {
	report := models.ScanReport{
		Devices: []models.Device{
			{ID: "dev-192.168.1.20", Addresses: []models.IPAddress{{IP: "192.168.1.20", Version: 4}}},
			{ID: "dev-192.168.1.1", Addresses: []models.IPAddress{{IP: "192.168.1.1", Version: 4}}, IsGateway: true},
		},
	}
	rich := BuildRichTopology(report)
	for _, node := range rich.Nodes {
		if node.ID == "0" || node.ID == "1" || node.ID == "" {
			t.Fatalf("unstable node id generated: %#v", node)
		}
		if !strings.HasPrefix(node.ID, "dev-") {
			t.Fatalf("node id = %q, want deterministic device key", node.ID)
		}
	}
}

func findRichNode(r models.RichTopologyModel, id string) *models.RichTopologyNode {
	for i := range r.Nodes {
		if r.Nodes[i].ID == id {
			return &r.Nodes[i]
		}
	}
	return nil
}

func findRichEdge(r models.RichTopologyModel, id string) *models.RichTopologyEdge {
	for i := range r.Edges {
		if r.Edges[i].ID == id {
			return &r.Edges[i]
		}
	}
	return nil
}

func findV2Edge(r models.TopologyV2Model, id string) *models.TopologyV2Edge {
	for i := range r.Edges {
		if r.Edges[i].ID == id {
			return &r.Edges[i]
		}
	}
	return nil
}
