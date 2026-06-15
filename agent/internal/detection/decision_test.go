package detection

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

// weakBag is an evidence bag that carries only contextual network evidence.
func weakBag() evidenceBag {
	return evidenceBag{PTR: "x.dynamic.ttnet.com.tr", Org: "TurkTelekom", UPnPFound: false}
}

func TestShouldReturnUnknown_LowConfidence(t *testing.T) {
	// Decent scores/margin but very low confidence -> Unknown.
	scores := map[string]float64{"DSL": 0.5, "Fiber": 0.2}
	unknown, reasons := shouldReturnUnknown(scores, 0.25, weakBag(), false)
	if !unknown {
		t.Fatal("expected Unknown for confidence 0.25")
	}
	if !hasReason(reasons, "Classification confidence is low.") {
		t.Errorf("expected low-confidence reason, got %v", reasons)
	}
}

func TestShouldReturnUnknown_LowMargin(t *testing.T) {
	// Two different categories essentially tied -> Unknown even if confidence ok.
	scores := map[string]float64{"DSL": 0.40, "Fiber": 0.37}
	unknown, _ := shouldReturnUnknown(scores, 0.6, weakBag(), false)
	if !unknown {
		t.Fatal("expected Unknown for category margin 0.03")
	}
}

func TestShouldReturnUnknown_LowTopScore(t *testing.T) {
	scores := map[string]float64{"DSL": 0.25, "VDSL": 0.18, "Fiber": 0.15}
	unknown, _ := shouldReturnUnknown(scores, 0.6, weakBag(), false)
	if !unknown {
		t.Fatal("expected Unknown for top score 0.25")
	}
}

func TestShouldReturnUnknown_NoStrongEvidence(t *testing.T) {
	// Even a high, clear, confident score is downgraded when nothing physically
	// proves the access type.
	scores := map[string]float64{"DSL": 0.9}
	unknown, reasons := shouldReturnUnknown(scores, 0.9, weakBag(), false)
	if !unknown {
		t.Fatal("expected Unknown without strong physical evidence")
	}
	if !hasReason(reasons, "No strong physical-layer evidence of the access type was found.") {
		t.Errorf("expected no-strong-evidence reason, got %v", reasons)
	}
}

func TestShouldReturnUnknown_StrongEvidenceCommits(t *testing.T) {
	// Fingerprint matched (strong) + good numbers -> a real verdict, not Unknown.
	scores := map[string]float64{"VDSL": 0.9, "DSL": 0.85}
	unknown, _ := shouldReturnUnknown(scores, 0.8, evidenceBag{RouterModel: "TP-Link Archer VR400"}, true)
	if unknown {
		t.Fatal("did not expect Unknown with a fingerprint match and strong scores")
	}
}

func TestDecisionQuality(t *testing.T) {
	if got := decisionQuality(0.9, 0.5, 0.9, true); got != "high" {
		t.Errorf("strong/clear case = %q, want high", got)
	}
	if got := decisionQuality(0.6, 0.2, 0.5, false); got != "low" {
		t.Errorf("solid-margin without strong evidence = %q, want low", got)
	}
	if got := decisionQuality(0.6, 0.2, 0.5, true); got != "medium" {
		t.Errorf("solid-margin with strong evidence = %q, want medium", got)
	}
	if got := decisionQuality(0.25, 0.03, 0.25, false); got != "low" {
		t.Errorf("weak case = %q, want low", got)
	}
}

func TestHasStrongPhysicalEvidence(t *testing.T) {
	if !hasStrongPhysicalEvidence(evidenceBag{}, true) {
		t.Error("fingerprint match must count as strong")
	}
	if !hasStrongPhysicalEvidence(evidenceBag{Text: "Huawei HG8245H GPON ONT"}, false) {
		t.Error("GPON/ONT marker must count as strong")
	}
	if hasStrongPhysicalEvidence(weakBag(), false) {
		t.Error("PTR/ASN alone must NOT count as strong")
	}
}

func TestCategoryScoresMaxPerCategory(t *testing.T) {
	cat := categoryScores(map[string]float64{"DSL": 0.4, "VDSL": 0.7, "Fiber": 0.3})
	if cat[models.CatDSL] != 0.7 {
		t.Errorf("DSL category = %v, want 0.7 (max of DSL/VDSL)", cat[models.CatDSL])
	}
	if cat[models.CatFiber] != 0.3 {
		t.Errorf("Fiber category = %v, want 0.3", cat[models.CatFiber])
	}
}

func hasReason(reasons []string, want string) bool {
	for _, r := range reasons {
		if r == want {
			return true
		}
	}
	return false
}
