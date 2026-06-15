package detection

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestMergeGatewayDevicesPreservesStrongRoleAndEvidence(t *testing.T) {
	devices := MergeGatewayDevices([]models.GatewayDevice{
		{
			IP: "192.168.1.1", Role: "default_gateway", Reachable: true,
			ServerHeader: "nginx", DeviceConfidence: 0.30,
		},
		{
			IP: "192.168.1.1", Role: "possible_cpe", TR064Found: true,
			Model: "VMG3312-B10B", AccessHints: []string{models.TypeVDSL},
			TR064Services: []string{"WANDSLInterfaceConfig"}, AccessConfidence: 0.80,
		},
	})
	if len(devices) != 1 {
		t.Fatalf("devices = %d, want 1", len(devices))
	}
	d := devices[0]
	if d.Role != "possible_cpe" || !d.Reachable || !d.TR064Found {
		t.Fatalf("merged device lost role/reachability/TR-064 evidence: %#v", d)
	}
	if d.DeviceConfidence != 0.30 || d.AccessConfidence != 0.80 {
		t.Fatalf("merged confidence mismatch: %#v", d)
	}
	if !containsTestString(d.AccessHints, models.TypeVDSL) || !containsTestString(d.TR064Services, "WANDSLInterfaceConfig") {
		t.Fatalf("merged slices missing evidence: %#v", d)
	}
}

func TestResolveNATTopologyKeepsInternalDoubleNATWithMatchingSTUN(t *testing.T) {
	nat := ResolveNATTopology(evidenceBag{
		PublicIP:          "95.15.1.1",
		DoubleNATPossible: true,
		NATTopology: &models.NATTopology{
			STUNPublicIP: "95.15.1.1",
		},
	})
	if !nat.InternalDoubleNATPossible || !nat.DoubleNAT {
		t.Fatalf("internal double NAT should survive STUN/public-IP agreement: %#v", nat)
	}
	if !nat.ExternalPublicIPConsistent || !nat.PublicIPMatches {
		t.Fatalf("STUN/public-IP agreement should be recorded: %#v", nat)
	}
	if nat.Topology != "internal_double_nat_public_ipv4" {
		t.Fatalf("topology = %q, want internal_double_nat_public_ipv4", nat.Topology)
	}
}

func TestScoreContributionsExposeCategoryTypeSubtype(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeQuick}, []models.ProbeResult{{
		ProbeName: "tr064_probe",
		Status:    models.StatusSuccess,
		Hints:     []string{models.TypeDSL, models.TypeVDSL},
		Evidence: map[string]any{
			"tr064_found":            true,
			"access_confidence":      0.80,
			"device_confidence":      0.55,
			"strong_access_evidence": true,
			"wan_signals": []models.WANSignal{{
				Source: "tr064_probe", Type: "tr064_service", Value: "WANDSLInterfaceConfig PTM VDSL2", Strength: "strong", Confidence: 0.80,
			}},
		},
	}})
	found := false
	for _, c := range res.ScoreContributions {
		if c.Target == models.TypeVDSL && c.Category == models.CatDSL && c.Type == models.TypeVDSL {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing auditable VDSL contribution: %#v", res.ScoreContributions)
	}
}

func TestContextConfidenceCanExceedClassificationWithoutPhysical(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, []models.ProbeResult{{
		ProbeName: "performance_profile_probe",
		Status:    models.StatusSuccess,
		Evidence: map[string]any{
			"performance_confidence": 0.35,
			"performance_profile":    models.PerformanceProfile{Target: "1.1.1.1", Method: "fake", IdleLatencyMS: 4},
		},
	}})
	if res.PrimaryType != "Unknown" {
		t.Fatalf("primary = %q, want Unknown", res.PrimaryType)
	}
	if res.ContextConfidence <= res.ClassificationConfidence {
		t.Fatalf("context confidence should exceed classification confidence without physical evidence: %#v", res.ConfidenceBreakdown)
	}
}

func TestTR064FoundIsCarriedIntoNetworkContext(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeQuick}, []models.ProbeResult{{
		ProbeName: "tr064_probe",
		Status:    models.StatusSuccess,
		Evidence: map[string]any{
			"tr064_found":       true,
			"device_confidence": 0.45,
		},
	}})
	if res.DetectedNetworkContext == nil || !res.DetectedNetworkContext.TR064Found {
		t.Fatalf("TR-064 flag missing from context: %#v", res.DetectedNetworkContext)
	}
}

func containsTestString(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
