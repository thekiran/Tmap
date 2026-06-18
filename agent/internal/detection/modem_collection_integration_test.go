package detection

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestCurrentDoubleNATFixtureBuildsModemCollection(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep, Online: true}, loadFixture(t, "scan_current_double_nat_unknown.json"))

	if res.Classification.PrimaryType != "Unknown" || res.Classification.SafeToDisplayAsFinal {
		t.Fatalf("classification = %#v", res.Classification)
	}
	if res.ModemCollection == nil {
		t.Fatal("missing modem_collection")
	}
	mc := res.ModemCollection
	if mc.NormalizedGatewayChain.InternalDoubleNATPossible != models.TriTrue {
		t.Fatalf("gateway chain = %#v", mc.NormalizedGatewayChain)
	}
	if mc.NAT.PublicIPMatches != models.TriTrue || mc.NAT.InternalDoubleNATPossible != models.TriTrue {
		t.Fatalf("nat state = %#v", mc.NAT)
	}
	if len(mc.CPECandidates) < 2 {
		t.Fatalf("candidates = %#v", mc.CPECandidates)
	}
	var def, upstream *models.CPECandidate
	for i := range mc.CPECandidates {
		switch mc.CPECandidates[i].IP {
		case "192.168.31.1":
			def = &mc.CPECandidates[i]
		case "192.168.1.1":
			upstream = &mc.CPECandidates[i]
		}
	}
	if def == nil || def.ReachableState != models.TriTrue {
		t.Fatalf("default gateway candidate = %#v", def)
	}
	if upstream == nil || upstream.Priority != "high" || upstream.WANPhysicalEvidence.Status != "missing" {
		t.Fatalf("upstream candidate = %#v", upstream)
	}
	if mc.AccessClassification.PrimaryType != "Unknown" || mc.AccessClassification.SafeToDisplayAsFinal {
		t.Fatalf("modem access classification = %#v", mc.AccessClassification)
	}
}
