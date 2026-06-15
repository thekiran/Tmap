package detection

import (
	"strings"
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestGatewayDeviceTextIncludedInFingerprint(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, gatewayDeviceResults("Zyxel", "VMG3312-B10B", []string{models.TypeDSL, models.TypeVDSL}))
	if res.DetectedNetworkContext == nil || !res.DetectedNetworkContext.FingerprintMatched {
		t.Fatalf("gateway device text did not produce a fingerprint match: %#v", res.DetectedNetworkContext)
	}
	if res.DetectedNetworkContext.LikelyModemIP != "192.168.1.1" {
		t.Fatalf("likely modem = %q, want 192.168.1.1", res.DetectedNetworkContext.LikelyModemIP)
	}
}

func TestGatewayFingerprintCanRaiseConfidence(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	weak := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, loadFixture(t, "scan_uncertain_ttnet.json"))
	strong := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, gatewayDeviceResults("Zyxel", "VMG3312-B10B", []string{models.TypeDSL, models.TypeVDSL}))
	if strong.PrimaryType == "Unknown" {
		t.Fatalf("gateway fingerprint should allow a committed verdict, got Unknown with reasons %v", strong.UncertaintyReasons)
	}
	if strong.Confidence <= weak.Confidence {
		t.Fatalf("confidence = %v, want > weak %v", strong.Confidence, weak.Confidence)
	}
	if strong.DecisionQuality != "medium" && strong.DecisionQuality != "high" {
		t.Fatalf("decision_quality = %q, want medium/high", strong.DecisionQuality)
	}
}

func TestNoGatewayDeviceKeepsUnknown(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, loadFixture(t, "scan_uncertain_ttnet.json"))
	if res.PrimaryType != "Unknown" {
		t.Fatalf("primary = %q, want Unknown", res.PrimaryType)
	}
}

func TestEqualFiberVDSLScoresExplanationDoesNotPickOnlyFiber(t *testing.T) {
	lines := buildUncertainExplanation("Fiber", map[string]float64{
		models.TypeFiber: 0.50,
		models.TypeVDSL:  0.50,
	}, evidenceBag{}, false, nil)
	text := strings.Join(lines, "\n")
	if strings.Contains(text, "Leading candidate: Fiber") {
		t.Fatalf("explanation picked only Fiber: %s", text)
	}
	if !strings.Contains(text, "Leading candidates: Fiber, VDSL") {
		t.Fatalf("missing tied-candidates explanation: %s", text)
	}
}

func TestGenericGatewayKeepsUnknown(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, genericGatewayResults())
	if res.PrimaryType != "Unknown" {
		t.Fatalf("primary = %q, want Unknown", res.PrimaryType)
	}
	if res.DecisionQuality != "low" {
		t.Fatalf("decision_quality = %q, want low", res.DecisionQuality)
	}
}

func TestNoStrongEvidenceKeepsConfidenceLow(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, genericGatewayResults())
	if res.Confidence > 0.35 {
		t.Fatalf("confidence = %v, want <= 0.35", res.Confidence)
	}
}

func TestFiberScoreNotInflatedByNginx(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, genericGatewayResults())
	if res.Scores[models.TypeFiber] > 0.15 {
		t.Fatalf("Fiber score = %v, want <= 0.15", res.Scores[models.TypeFiber])
	}
}

func TestExplanationMatchesReachabilityState(t *testing.T) {
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	res := engine.Analyze(models.ScanInput{Mode: models.ModeDeep}, genericGatewayResults())
	text := strings.Join(res.Explanation, "\n")
	if !strings.Contains(text, "An upstream gateway 192.168.1.1 was detected but could not be reached") {
		t.Fatalf("missing unreachable upstream explanation: %s", text)
	}
	if strings.Contains(text, "may be reachable, but the modem model could not be verified") {
		t.Fatalf("explanation contradicts reachability: %s", text)
	}
}

func TestFiberFTTHNotDuplicatedAsIndependentStrongCandidates(t *testing.T) {
	ranked := collapseParentSubtypeCandidates(rankAll(map[string]float64{
		models.TypeFiber: 0.50,
		models.TypeFTTH:  0.45,
		models.TypeVDSL:  0.20,
	}))
	for _, ts := range ranked {
		if ts.Type == models.TypeFiber {
			t.Fatalf("Fiber parent should be collapsed when FTTH subtype is present: %v", ranked)
		}
	}
}

func gatewayDeviceResults(manufacturer, model string, hints []string) []models.ProbeResult {
	return []models.ProbeResult{
		{ProbeName: "gateway_probe", Status: models.StatusSuccess, Evidence: map[string]any{"gateway": "192.168.31.1"}},
		{
			ProbeName: "gateway_chain_probe",
			Status:    models.StatusSuccess,
			Confidence: 0.85,
			Hints: hints,
			Evidence: map[string]any{
				"gateway_chain":       []string{"192.168.31.1", "192.168.1.1"},
				"double_nat_possible": true,
				"likely_modem_ip":     "192.168.1.1",
				"gateway_devices": []models.GatewayDevice{
					{IP: "192.168.31.1", Role: "default_gateway", Reachable: true, HTTPTitle: "Mi Router", Manufacturer: "Xiaomi", DeviceConfidence: 0.40, Confidence: 0.40},
					{IP: "192.168.1.1", Role: "possible_modem", Reachable: true, HTTPTitle: model + " VDSL", Manufacturer: manufacturer, Model: model, AccessHints: hints, DeviceConfidence: 0.85, AccessConfidence: 0.85, Confidence: 0.85},
				},
			},
		},
	}
}

func genericGatewayResults() []models.ProbeResult {
	return []models.ProbeResult{
		{ProbeName: "gateway_probe", Status: models.StatusSuccess, Evidence: map[string]any{"gateway": "192.168.31.1"}},
		{ProbeName: "latency_probe", Status: models.StatusSuccess, Evidence: map[string]any{"avg_ms": 3.8, "jitter_ms": 2.0}},
		{ProbeName: "asn_probe", Status: models.StatusSuccess, Evidence: map[string]any{"ptr": "95.15.182.146.dynamic.ttnet.com.tr", "org": "TurkTelekom"}},
		{
			ProbeName: "gateway_chain_probe",
			Status:    models.StatusSuccess,
			Confidence: 0.60,
			Evidence: map[string]any{
				"gateway_chain":       []string{"192.168.31.1", "192.168.1.1"},
				"double_nat_possible": true,
				"gateway_devices": []models.GatewayDevice{
					{IP: "192.168.31.1", Role: "default_gateway", Reachable: true, ServerHeader: "nginx", DeviceConfidence: 0.60, AccessConfidence: 0, Confidence: 0.60},
					{IP: "192.168.1.1", Role: "upstream_private_gateway", Reachable: false, DeviceConfidence: 0, AccessConfidence: 0, Confidence: 0},
				},
			},
		},
	}
}
