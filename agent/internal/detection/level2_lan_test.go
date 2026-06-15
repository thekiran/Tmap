package detection

import (
	"testing"

	"github.com/thekiran/iad/internal/scoring"
	"github.com/thekiran/iad/pkg/models"
)

func TestUPnPIGDWANAccessTypeCanCommitDSL(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeQuick}, []models.ProbeResult{{
		ProbeName:  "upnp_igd_probe",
		Status:     models.StatusSuccess,
		Confidence: 0.80,
		Hints:      []string{models.TypeDSL, models.TypeVDSL},
		Evidence: map[string]any{
			"igd_wan_common_found": true,
			"wan_access_type":      "DSL",
			"access_confidence":    0.80,
			"device_confidence":    0.55,
			"strong_access_evidence": true,
			"wan_signals": []models.WANSignal{{
				Source:     "upnp_igd_probe",
				IP:         "192.168.1.1",
				Type:       "wan_common_interface",
				Value:      "DSL",
				Strength:   scoring.EvidenceStrong,
				Confidence: 0.80,
			}},
		},
	}})
	if res.PrimaryType == "Unknown" {
		t.Fatalf("explicit UPnP WANAccessType should commit, reasons: %v scores=%v", res.UncertaintyReasons, res.Scores)
	}
	if res.Category != models.CatDSL {
		t.Fatalf("category = %q, want DSL", res.Category)
	}
	if res.DetectedNetworkContext == nil || len(res.DetectedNetworkContext.WANSignals) == 0 {
		t.Fatalf("WANSignals missing from context: %#v", res.DetectedNetworkContext)
	}
}

func TestUPnPIGDGenericServiceKeepsUnknown(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeQuick}, []models.ProbeResult{{
		ProbeName:  "upnp_igd_probe",
		Status:     models.StatusSuccess,
		Confidence: 0.35,
		Evidence: map[string]any{
			"igd_wan_common_found": false,
			"cpe_services":         []string{"urn:schemas-upnp-org:service:WANIPConnection:1"},
			"device_confidence":    0.35,
			"access_confidence":    0.0,
		},
	}})
	if res.PrimaryType != "Unknown" {
		t.Fatalf("primary = %q, want Unknown", res.PrimaryType)
	}
	if res.Confidence > 0.35 {
		t.Fatalf("confidence = %v, want <= 0.35", res.Confidence)
	}
}

func TestTR064WANSignalCanCommitVDSL(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeQuick}, []models.ProbeResult{{
		ProbeName:  "tr064_probe",
		Status:     models.StatusSuccess,
		Confidence: 0.80,
		Hints:      []string{models.TypeDSL, models.TypeVDSL},
		Evidence: map[string]any{
			"tr064_found":            true,
			"physical_link_status":   "Up",
			"access_confidence":      0.80,
			"device_confidence":      0.55,
			"strong_access_evidence": true,
			"wan_signals": []models.WANSignal{{
				Source:     "tr064_probe",
				IP:         "192.168.1.1",
				Type:       "tr064_service",
				Value:      "WANDSLInterfaceConfig PTM VDSL2",
				Strength:   scoring.EvidenceStrong,
				Confidence: 0.80,
			}},
		},
	}})
	if res.PrimaryType == "Unknown" {
		t.Fatalf("TR-064 DSL/PTM should commit, reasons: %v scores=%v", res.UncertaintyReasons, res.Scores)
	}
	if res.Category != models.CatDSL {
		t.Fatalf("category = %q, want DSL", res.Category)
	}
}

func TestSTUNContextDoesNotClassifyAccessType(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeQuick}, []models.ProbeResult{{
		ProbeName:  "stun_pcp_nat_probe",
		Status:     models.StatusSuccess,
		Confidence: 0.40,
		Evidence: map[string]any{
			"network_confidence": 0.35,
			"nat_topology": models.NATTopology{
				PublicIP:        "95.15.182.146",
				STUNPublicIP:    "95.15.182.146",
				PublicIPMatches: true,
				Topology:        "stun_observed",
			},
		},
	}})
	if res.PrimaryType != "Unknown" {
		t.Fatalf("primary = %q, want Unknown", res.PrimaryType)
	}
}

func TestConfidenceBreakdownReflectsEvidenceClasses(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, []models.ProbeResult{
		{
			ProbeName: "performance_profile_probe",
			Status:    models.StatusSuccess,
			Evidence: map[string]any{
				"performance_confidence": 0.35,
				"performance_profile": models.PerformanceProfile{
					Target: "1.1.1.1", Method: "fake", IdleLatencyMS: 5,
				},
			},
		},
		{
			ProbeName: "upnp_igd_deep_probe",
			Status:    models.StatusSuccess,
			Hints:     []string{models.TypeDSL, models.TypeVDSL},
			Evidence: map[string]any{
				"strong_access_evidence": true,
				"access_confidence":      0.80,
				"device_confidence":      0.55,
				"wan_signals": []models.WANSignal{{
					Source: "upnp_igd_deep_probe", Type: "wan_common_interface", Value: "DSL", Strength: scoring.EvidenceStrong, Confidence: 0.80,
				}},
			},
		},
	})
	if res.ConfidenceBreakdown.Physical == 0 || res.ConfidenceBreakdown.Performance == 0 {
		t.Fatalf("breakdown missing evidence classes: %#v", res.ConfidenceBreakdown)
	}
}

func TestNextBestProbeSuggestsTR064OrUPnPWhenPhysicalMissing(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, []models.ProbeResult{{
		ProbeName: "performance_profile_probe",
		Status:    models.StatusSuccess,
		Evidence: map[string]any{
			"performance_confidence": 0.35,
			"performance_profile": models.PerformanceProfile{Target: "1.1.1.1", Method: "fake", IdleLatencyMS: 4},
		},
	}})
	if res.PrimaryType != "Unknown" {
		t.Fatalf("primary = %q, want Unknown", res.PrimaryType)
	}
	if len(res.NextBestProbes) == 0 {
		t.Fatal("next_best_probes should be populated when physical evidence is missing")
	}
}

func TestPerformanceOnlyKeepsUnknownLowConfidence(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, []models.ProbeResult{{
		ProbeName: "performance_profile_probe",
		Status:    models.StatusSuccess,
		Evidence: map[string]any{
			"performance_confidence": 0.35,
			"performance_profile": models.PerformanceProfile{Target: "1.1.1.1", Method: "fake", IdleLatencyMS: 3},
		},
	}})
	if res.PrimaryType != "Unknown" || res.DecisionQuality != "low" {
		t.Fatalf("performance-only should stay Unknown/low, got %s/%s", res.PrimaryType, res.DecisionQuality)
	}
	if res.Confidence > 0.35 {
		t.Fatalf("confidence = %v, want <= 0.35", res.Confidence)
	}
}
