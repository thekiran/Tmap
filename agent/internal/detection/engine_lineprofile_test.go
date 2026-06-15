package detection

import (
	"strings"
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

// TestAnalyzeParsesVDSL2Profile verifies the full engine reads authorized CPE
// telemetry into a fine-grained LineProfile (VDSL2 profile 35b + line stats),
// commits a DSL-family verdict on that physical evidence, and explains it.
func TestAnalyzeParsesVDSL2Profile(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, loadFixture(t, "scan_vdsl2_35b.json"))

	if res.Category != models.CatDSL {
		t.Errorf("category = %q, want DSL (primary=%q)", res.Category, res.PrimaryType)
	}
	if !contains([]string{models.TypeDSL, models.TypeVDSL, models.TypeVDSL2}, res.PrimaryType) {
		t.Errorf("primary = %q, want a DSL-family type", res.PrimaryType)
	}
	if res.DecisionQuality == "low" {
		t.Errorf("decision quality = low, want medium/high given physical line evidence")
	}

	nc := res.DetectedNetworkContext
	if nc == nil || nc.LineProfile == nil || nc.LineProfile.DSL == nil {
		t.Fatalf("expected a DSL line profile in the network context, got %+v", nc)
	}
	lp := nc.LineProfile
	if lp.DSL.Profile != "35b" {
		t.Errorf("line profile = %q, want 35b", lp.DSL.Profile)
	}
	if !lp.DSL.Vectoring {
		t.Error("expected vectoring=true for 35b")
	}
	if lp.DSL.SNRMarginDownDB != 6.1 {
		t.Errorf("snr margin = %v, want 6.1", lp.DSL.SNRMarginDownDB)
	}
	if lp.DSL.SyncDownKbps != 294000 {
		t.Errorf("sync down = %d, want 294000", lp.DSL.SyncDownKbps)
	}

	if !explanationMentions(res.Explanation, "35b") {
		t.Errorf("explanation should mention the 35b profile, got %v", res.Explanation)
	}
}

// TestAnalyzeParsesDOCSISProfile verifies DOCSIS 3.1 telemetry commits a Cable
// verdict and records the version + OFDM/OFDMA.
func TestAnalyzeParsesDOCSISProfile(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, loadFixture(t, "scan_docsis31.json"))

	if res.Category != models.CatCable {
		t.Errorf("category = %q, want Cable (primary=%q)", res.Category, res.PrimaryType)
	}
	nc := res.DetectedNetworkContext
	if nc == nil || nc.LineProfile == nil || nc.LineProfile.DOCSIS == nil {
		t.Fatalf("expected a DOCSIS line profile, got %+v", nc)
	}
	if nc.LineProfile.DOCSIS.Version != "3.1" {
		t.Errorf("docsis version = %q, want 3.1", nc.LineProfile.DOCSIS.Version)
	}
	if !nc.LineProfile.DOCSIS.OFDM || !nc.LineProfile.DOCSIS.OFDMA {
		t.Error("expected OFDM and OFDMA to be detected")
	}
}

func explanationMentions(lines []string, sub string) bool {
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}
