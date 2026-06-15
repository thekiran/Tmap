package detection

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

// Paths are relative to this package directory (agent/internal/detection).
const (
	rulesDir    = "../../../rules"
	fixturesDir = "../../../tests/fixtures"
)

func loadFixture(t *testing.T, name string) []models.ProbeResult {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixturesDir, name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	var results []models.ProbeResult
	if err := json.Unmarshal(data, &results); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	return results
}

// TestAnalyzeFixtures runs the full engine over canned probe outputs and checks
// the classification. It is fully deterministic and needs no network, so it
// behaves identically on every OS.
func TestAnalyzeFixtures(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	if len(engine.Rules) == 0 {
		t.Fatal("no rules loaded; check rulesDir")
	}
	if len(engine.Fingerprints) == 0 {
		t.Fatal("no fingerprints loaded; check rulesDir")
	}

	cases := []struct {
		fixture     string
		wantCat     string
		wantPrimary []string // any one of these is acceptable as the top verdict
	}{
		{"scan_vdsl_vr400.json", models.CatDSL, []string{models.TypeDSL, models.TypeVDSL, models.TypeVDSL2}},
		{"scan_fiber_gpon.json", models.CatFiber, []string{models.TypeFiber, models.TypeFTTH, models.TypeGPON}},
		{"scan_mobile_lte.json", models.CatMobile, []string{models.TypeMobile, models.TypeFWA, models.TypeLTE}},
	}

	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			results := loadFixture(t, tc.fixture)
			res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, results)

			if res.Category != tc.wantCat {
				t.Errorf("category = %q, want %q (primary=%q, scores=%v)",
					res.Category, tc.wantCat, res.PrimaryType, res.Scores)
			}
			if !contains(tc.wantPrimary, res.PrimaryType) {
				t.Errorf("primary = %q, want one of %v", res.PrimaryType, tc.wantPrimary)
			}
			if res.Confidence <= 0 || res.Confidence > 1 {
				t.Errorf("confidence = %v, want in (0,1]", res.Confidence)
			}
			if len(res.Explanation) == 0 {
				t.Error("expected a non-empty explanation")
			}
		})
	}
}

// TestAnalyzeNoEvidence verifies the engine degrades gracefully to "Unknown"
// rather than panicking or inventing a verdict.
func TestAnalyzeNoEvidence(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeQuick}, nil)
	if res.PrimaryType != "Unknown" {
		t.Errorf("primary = %q, want Unknown", res.PrimaryType)
	}
	if res.Confidence != 0 {
		t.Errorf("confidence = %v, want 0", res.Confidence)
	}
	if res.ContextConfidence != 0 {
		t.Errorf("context confidence = %v, want 0 with no evidence", res.ContextConfidence)
	}
}

func contains(s []string, v string) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}
