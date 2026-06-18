package detection

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestCandidateEvidenceDoesNotLeakAcrossAccessTypes(t *testing.T) {
	bag := evidenceBag{
		PhysicalEvidence: 0.92,
		WANSignals: []models.WANSignal{
			{
				Source:     "upnp_wan_common_interface",
				Type:       "WANAccessType",
				Value:      "DOCSIS Cable Up",
				Strength:   string(models.EvidencePhysical),
				Confidence: 0.92,
			},
		},
	}
	candidates := buildCandidates(map[string]float64{
		models.TypeCable: 0.92,
		models.TypeFiber: 0.70,
		models.TypeVDSL:  0.65,
	}, bag, false)

	cable := mustCandidate(t, candidates, models.CatCable)
	if len(cable.SupportingEvidence) == 0 {
		t.Fatal("Cable candidate should carry DOCSIS/Cable supporting evidence")
	}

	fiber := mustCandidate(t, candidates, models.CatFiber)
	assertNoCableSupport(t, fiber)
	if len(fiber.ContradictingEvidence) == 0 {
		t.Fatal("Fiber candidate should list Cable direct evidence as contradicting evidence")
	}

	dsl := mustCandidate(t, candidates, models.CatDSL)
	assertNoCableSupport(t, dsl)
	if len(dsl.ContradictingEvidence) == 0 {
		t.Fatal("DSL/VDSL candidate should list Cable direct evidence as contradicting evidence")
	}
}

func mustCandidate(t *testing.T, candidates []models.AccessCandidate, category string) models.AccessCandidate {
	t.Helper()
	for _, c := range candidates {
		if c.Category == category {
			return c
		}
	}
	t.Fatalf("candidate category %q not found in %#v", category, candidates)
	return models.AccessCandidate{}
}

func assertNoCableSupport(t *testing.T, candidate models.AccessCandidate) {
	t.Helper()
	for _, ev := range candidate.SupportingEvidence {
		target := publicTypeForTarget(ev.TargetType)
		if target == models.TypeCable || target == models.TypeDOCSIS {
			t.Fatalf("%s candidate leaked cable evidence as support: %#v", candidate.Category, ev)
		}
	}
}
