package modemcollector

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestWANCommonInterfaceProducesWANPhysicalEvidence(t *testing.T) {
	result := models.ScanResult{
		Classification: models.Classification{
			PrimaryType: "Unknown",
			Category:    models.CatUnknown,
		},
		DetectedNetworkContext: &models.NetworkContext{
			Gateway: "192.168.1.1",
			GatewayDevices: []models.GatewayDevice{
				{
					IP:                      "192.168.1.1",
					Role:                    "default_gateway",
					WANCommonInterfaceFound: true,
					WANAccessType:           "Cable Up",
					PhysicalLinkStatus:      "Up",
					AccessConfidence:        0.91,
					Confidence:              0.91,
					EvidenceIDs:             []string{"ev-wan-common"},
				},
			},
		},
	}

	collection := Build(BuildInput{Result: result})
	if len(collection.CPECandidates) == 0 {
		t.Fatal("expected a CPE candidate")
	}
	ev := collection.CPECandidates[0].WANPhysicalEvidence
	if ev.Status != "present" {
		t.Fatalf("wan physical evidence status = %q, want present", ev.Status)
	}
	if ev.Type == nil || *ev.Type != "Cable Up" {
		t.Fatalf("wan physical evidence type = %#v, want Cable Up", ev.Type)
	}
	if ev.Confidence < 0.90 {
		t.Fatalf("wan physical evidence confidence = %.2f, want >= 0.90", ev.Confidence)
	}
}
