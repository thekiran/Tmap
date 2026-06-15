package detection

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

// TestAnalyzeUncertainTTNet is the end-to-end check for the real-world case the
// decision layer was built for: TTNet dynamic IP, no UPnP modem, double NAT,
// low latency, scores clustered across DSL/VDSL/Fiber. The engine must NOT
// commit, but must keep the candidates and report the factual context.
func TestAnalyzeUncertainTTNet(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	results := loadFixture(t, "scan_uncertain_ttnet.json")
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, results)

	if res.PrimaryType != "Unknown" {
		t.Errorf("primary = %q, want Unknown (scores=%v)", res.PrimaryType, res.Scores)
	}
	if res.Category != models.CatUnknown {
		t.Errorf("category = %q, want Unknown", res.Category)
	}
	if res.DecisionQuality != "low" {
		t.Errorf("decision_quality = %q, want low", res.DecisionQuality)
	}
	if res.Confidence <= 0 || res.Confidence >= minConfidence {
		t.Errorf("confidence = %v, want in (0, %v)", res.Confidence, minConfidence)
	}
	if len(res.Alternatives) == 0 {
		t.Error("alternatives must still be present even when Unknown")
	}
	if len(res.UncertaintyReasons) == 0 {
		t.Error("uncertainty reasons must be populated")
	}

	nc := res.DetectedNetworkContext
	if nc == nil {
		t.Fatal("network context must be reported")
	}
	if !nc.DoubleNATPossible {
		t.Error("double NAT must be detected from the traceroute hops")
	}
	if nc.UPnPFound {
		t.Error("UPnP should be reported as not found")
	}
	if nc.FingerprintMatched {
		t.Error("no fingerprint should match")
	}
	if nc.MainAdapter != "Ethernet" {
		t.Errorf("main adapter = %q, want Ethernet (virtual/APIPA filtered)", nc.MainAdapter)
	}
	if nc.LocalAccess != "Ethernet" {
		t.Errorf("local access = %q, want Ethernet", nc.LocalAccess)
	}
	if nc.ISP != "TurkTelekom" {
		t.Errorf("ISP = %q, want TurkTelekom", nc.ISP)
	}
}

// TestLatencyAndPTRDoNotDecideAlone verifies that weak signals (latency + PTR),
// with no modem model, cannot by themselves produce a committed verdict.
func TestLatencyAndPTRDoNotDecideAlone(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	results := []models.ProbeResult{
		{ProbeName: "latency_probe", Status: models.StatusSuccess, Evidence: map[string]any{"avg_ms": 3.8, "jitter_ms": 2.0}},
		{ProbeName: "asn_probe", Status: models.StatusSuccess, Evidence: map[string]any{"ptr": "x.dynamic.ttnet.com.tr", "org": "TurkTelekom"}},
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, results)
	if res.PrimaryType != "Unknown" {
		t.Errorf("primary = %q, want Unknown (latency+PTR must not decide alone)", res.PrimaryType)
	}
}
