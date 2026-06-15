package detection

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestFiberConfirmedByOpticalInterface(t *testing.T) {
	res := analyzeSynthetic(t, physicalProbe("tr181_interface_stack_probe",
		[]string{models.TypeGPON, models.TypeFTTH, models.TypeFiber},
		"Device.Optical.Interface GPON active ONT optical rx power -18.2 dBm"))

	if res.Classification.PrimaryType != models.TypeFiber || res.Classification.State != "confirmed" {
		t.Fatalf("classification = %#v, want confirmed Fiber", res.Classification)
	}
	if !res.Classification.SafeToDisplayAsFinal {
		t.Fatalf("fiber optical evidence should be final-safe: %#v", res.Classification)
	}
}

func TestFiberConfirmedByGPONFingerprintAndActiveOpticalWAN(t *testing.T) {
	res := analyzeSynthetic(t,
		modelOnlyHTTPFingerprint("Huawei", "HG8245H", []string{models.TypeFiber, models.TypeFTTH, models.TypeGPON}),
		physicalProbe("tr181_interface_stack_probe",
			[]string{models.TypeGPON, models.TypeFTTH, models.TypeFiber},
			"GPON ONT Device.Optical.Interface active"))

	if res.Classification.PrimaryType != models.TypeFiber || res.Classification.State != "confirmed" {
		t.Fatalf("classification = %#v, want confirmed Fiber", res.Classification)
	}
}

func TestVDSLConfirmedByVDSL2LineMIB(t *testing.T) {
	res := analyzeSynthetic(t, physicalProbe("snmp_probe_opt_in",
		[]string{models.TypeVDSL2, models.TypeVDSL, models.TypeDSL},
		"VDSL2-LINE-MIB ifName ptm0 VDSL2 Profile 17a line rate"))

	if res.Classification.PrimaryType != models.TypeVDSL || res.Classification.State != "confirmed" {
		t.Fatalf("classification = %#v, want confirmed VDSL", res.Classification)
	}
	if res.Classification.Subtype == nil || *res.Classification.Subtype == "" {
		t.Fatalf("expected VDSL subtype, got %#v", res.Classification)
	}
}

func TestVDSLConfirmedByTR181DSLLineAndPTMLink(t *testing.T) {
	res := analyzeSynthetic(t, physicalProbe("tr181_interface_stack_probe",
		[]string{models.TypeVDSL2, models.TypeVDSL, models.TypeDSL},
		"Device.DSL.Line active Device.PTM.Link active VDSL2"))

	if res.Classification.PrimaryType != models.TypeVDSL || res.Classification.State != "confirmed" {
		t.Fatalf("classification = %#v, want confirmed VDSL", res.Classification)
	}
}

func TestADSLConfirmedByIfTypeADSL94(t *testing.T) {
	res := analyzeSynthetic(t, physicalProbe("snmp_probe_opt_in",
		[]string{models.TypeADSL, models.TypeDSL},
		"IF-MIB ifType adsl(94) ADSL-LINE-MIB atm0 ADSL2+"))

	if res.Classification.PrimaryType != models.TypeADSL || res.Classification.State != "confirmed" {
		t.Fatalf("classification = %#v, want confirmed ADSL", res.Classification)
	}
}

func TestDOCSISConfirmedByDOCSISMIB(t *testing.T) {
	res := analyzeSynthetic(t, physicalProbe("snmp_probe_opt_in",
		[]string{models.TypeDOCSIS, models.TypeCable},
		"DOCSIS MIB DOCS-IF DOCSIS 3.1 OFDM OFDMA cable modem"))

	if res.Classification.PrimaryType != models.TypeCable || res.Classification.State != "confirmed" {
		t.Fatalf("classification = %#v, want confirmed Cable", res.Classification)
	}
}

func TestFWAConfirmedByLTENRInterface(t *testing.T) {
	res := analyzeSynthetic(t, physicalProbe("tr181_interface_stack_probe",
		[]string{models.TypeFWA, models.TypeLTE, models.TypeMobile},
		"Device.Cellular.Interface NR5G LTE WWAN active"))

	if res.Classification.PrimaryType != models.TypeFWA || res.Classification.State != "confirmed" {
		t.Fatalf("classification = %#v, want confirmed FWA", res.Classification)
	}
}

func TestEthernetWANDoesNotBecomeFiber(t *testing.T) {
	res := analyzeSynthetic(t, models.ProbeResult{
		ProbeName: "upnp_igd_probe",
		Status:    models.StatusSuccess,
		Hints:     []string{models.TypeEthernetWAN},
		Evidence: map[string]any{
			"igd_wan_common_found":   true,
			"wan_access_type":        "Ethernet",
			"access_confidence":      0.85,
			"device_confidence":      0.55,
			"strong_access_evidence": true,
			"wan_signals": []models.WANSignal{{
				Source: "upnp_igd_probe", Type: "wan_common_interface", Value: "Ethernet",
				Strength: string(models.EvidencePhysical), Confidence: 0.85,
			}},
		},
	})

	if res.Classification.PrimaryType != models.TypeEthernetWAN {
		t.Fatalf("classification = %#v, want EthernetWAN", res.Classification)
	}
	if res.Classification.State != "possible" || res.Classification.SafeToDisplayAsFinal {
		t.Fatalf("Ethernet WAN should be possible and not final-safe: %#v", res.Classification)
	}
	if res.Category == models.CatFiber || res.Scores[models.TypeFiber] > 0 {
		t.Fatalf("Ethernet WAN must not imply Fiber: category=%s scores=%v", res.Category, res.Scores)
	}
}

func TestWeakLatencyOnlyEvidenceReturnsUnknown(t *testing.T) {
	res := analyzeSynthetic(t, models.ProbeResult{
		ProbeName: "latency_probe",
		Status:    models.StatusSuccess,
		Evidence:  map[string]any{"avg_ms": 4.0, "jitter_ms": 1.0},
	})

	if res.PrimaryType != "Unknown" || res.Classification.State != "insufficient_evidence" {
		t.Fatalf("latency-only classification = %#v primary=%s", res.Classification, res.PrimaryType)
	}
	if res.Classification.Confidence > 0.40 {
		t.Fatalf("latency-only confidence = %v, want <= 0.40", res.Classification.Confidence)
	}
}

func TestASNPTROnlyEvidenceReturnsUnknown(t *testing.T) {
	res := analyzeSynthetic(t, models.ProbeResult{
		ProbeName: "asn_probe",
		Status:    models.StatusSuccess,
		Evidence:  map[string]any{"ptr": "95.15.1.1.dynamic.ttnet.com.tr", "org": "TurkTelekom"},
		Hints:     []string{models.TypeFiber, models.TypeVDSL},
	})

	if res.PrimaryType != "Unknown" {
		t.Fatalf("ASN/PTR-only primary=%s, want Unknown", res.PrimaryType)
	}
}

func TestFiberVDSLTieReturnsUnknown(t *testing.T) {
	scores := map[string]float64{models.TypeFiber: 0.50, models.TypeVDSL: 0.45}
	unknown, _ := shouldReturnUnknown(scores, 0.70, evidenceBag{RouterModel: "Ambiguous CPE", DeviceEvidence: 0.70}, false)
	if !unknown {
		t.Fatal("Fiber/VDSL close race without Tier A evidence must return Unknown")
	}
}

func TestNoTierABCapsConfidenceAt040(t *testing.T) {
	res := analyzeSynthetic(t,
		models.ProbeResult{
			ProbeName: "latency_probe",
			Status:    models.StatusSuccess,
			Evidence:  map[string]any{"avg_ms": 5.0, "jitter_ms": 1.0},
		},
		models.ProbeResult{
			ProbeName: "asn_probe",
			Status:    models.StatusSuccess,
			Evidence:  map[string]any{"org": "Example ISP"},
			Hints:     []string{models.TypeFiber, models.TypeVDSL},
		},
		models.ProbeResult{
			ProbeName: "gateway_probe",
			Status:    models.StatusSuccess,
			Evidence:  map[string]any{"gateway": "192.168.1.1"},
		})

	if res.Classification.Confidence > 0.40 {
		t.Fatalf("confidence=%v, want <=0.40 without Tier A/B", res.Classification.Confidence)
	}
}

func TestPerformanceRegionalOnlyCapsConfidenceAt025(t *testing.T) {
	res := analyzeSynthetic(t,
		models.ProbeResult{
			ProbeName: "latency_probe",
			Status:    models.StatusSuccess,
			Evidence:  map[string]any{"avg_ms": 5.0, "jitter_ms": 1.0},
		},
		models.ProbeResult{
			ProbeName: "asn_probe",
			Status:    models.StatusSuccess,
			Evidence:  map[string]any{"org": "Example ISP"},
			Hints:     []string{models.TypeFiber, models.TypeVDSL},
		})

	if res.Classification.Confidence > 0.25 {
		t.Fatalf("confidence=%v, want <=0.25 for performance+regional only", res.Classification.Confidence)
	}
}

func TestConflictingGatewayReachabilityDowngradesConfidence(t *testing.T) {
	res := analyzeSynthetic(t,
		modelOnlyHTTPFingerprint("Zyxel", "VMG3312-B10B", []string{models.TypeVDSL, models.TypeDSL}),
		models.ProbeResult{
			ProbeName: "gateway_reachability_diagnostics_probe",
			Status:    models.StatusSuccess,
			Evidence: map[string]any{
				"gateway_ip":           "192.168.1.1",
				"management_reachable": true,
				"tcp_ports_reachable":  []string{"80"},
				"network_confidence":   0.35,
				"route_present":        true,
			},
		})

	if !res.DataQuality.HasConflicts || len(res.Conflicts) == 0 {
		t.Fatalf("expected reachability conflict, got %#v", res.DataQuality)
	}
	if res.Classification.Confidence > 0.35 {
		t.Fatalf("confidence = %v, want high conflict cap <= 0.35", res.Classification.Confidence)
	}
}

func TestDoubleNATConflictDowngradesConfidence(t *testing.T) {
	res := analyzeSynthetic(t,
		modelOnlyHTTPFingerprint("Zyxel", "VMG3312-B10B", []string{models.TypeVDSL, models.TypeDSL}),
		models.ProbeResult{
			ProbeName: "stun_pcp_nat_probe",
			Status:    models.StatusSuccess,
			Evidence: map[string]any{
				"network_confidence": 0.35,
				"nat_topology": models.NATTopology{
					DoubleNAT:                 false,
					InternalDoubleNATPossible: false,
					Topology:                  "single_nat",
				},
			},
		},
		models.ProbeResult{
			ProbeName: "gateway_chain_probe",
			Status:    models.StatusSuccess,
			Evidence: map[string]any{
				"gateway_chain":       []string{"192.168.31.1", "192.168.1.1"},
				"double_nat_possible": true,
			},
		})

	if !res.DataQuality.HasConflicts {
		t.Fatalf("expected double NAT conflict, got %#v", res.DataQuality)
	}
	found := false
	for _, c := range res.Conflicts {
		if c.Field == "nat_topology.double_nat" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing double NAT conflict: %#v", res.Conflicts)
	}
	if res.Classification.Confidence >= 0.80 {
		t.Fatalf("confidence=%v, want downgraded", res.Classification.Confidence)
	}
}

func TestHTTPModelFingerprintGivesOnlyMediumConfidence(t *testing.T) {
	res := analyzeSynthetic(t, modelOnlyHTTPFingerprint("Zyxel", "VMG3312-B10B", []string{models.TypeVDSL, models.TypeDSL}))

	if res.PrimaryType == "Unknown" {
		t.Fatalf("expected probable model-based classification, reasons=%v", res.UncertaintyReasons)
	}
	if res.EvidenceTiers.DirectPhysical.Present {
		t.Fatalf("HTTP model fingerprint must not be direct physical evidence: %#v", res.EvidenceTiers)
	}
	if res.Classification.State != "probable" || res.Classification.SafeToDisplayAsFinal {
		t.Fatalf("classification = %#v, want probable and not final-safe", res.Classification)
	}
	if res.Classification.Confidence > 0.75 {
		t.Fatalf("model confidence = %v, want <= 0.75", res.Classification.Confidence)
	}
}

func TestTR064AuthRequiredWithoutCredentialsAddsNoDirectEvidence(t *testing.T) {
	res := analyzeSynthetic(t, models.ProbeResult{
		ProbeName: "tr064_probe",
		Status:    models.StatusSuccess,
		Evidence: map[string]any{
			"tr064_found":       true,
			"auth_required":     true,
			"device_confidence": 0.45,
			"access_confidence": 0.0,
		},
	})

	if res.PrimaryType != "Unknown" {
		t.Fatalf("TR-064 auth-required without credentials primary=%s, want Unknown", res.PrimaryType)
	}
	if res.EvidenceTiers.DirectPhysical.Present {
		t.Fatalf("auth-required TR-064 must not produce direct evidence: %#v", res.EvidenceTiers.DirectPhysical)
	}
}

func analyzeSynthetic(t *testing.T, results ...models.ProbeResult) models.ScanResult {
	t.Helper()
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	return engine.Analyze(models.ScanInput{Mode: models.ModeDeep, Online: true}, results)
}

func physicalProbe(name string, hints []string, text string) models.ProbeResult {
	return models.ProbeResult{
		ProbeName: name,
		Status:    models.StatusSuccess,
		Hints:     hints,
		Evidence: map[string]any{
			"direct_physical_evidence": true,
			"strong_access_evidence":   true,
			"access_confidence":        0.92,
			"device_confidence":        0.60,
			"cpe_text":                 text,
			"wan_signals": []models.WANSignal{{
				Source: name, Type: "physical_interface", Value: text,
				Strength: string(models.EvidencePhysical), Confidence: 0.92,
			}},
		},
	}
}

func modelOnlyHTTPFingerprint(manufacturer, model string, hints []string) models.ProbeResult {
	return models.ProbeResult{
		ProbeName: "http_fingerprint_v2",
		Status:    models.StatusSuccess,
		Hints:     hints,
		Evidence: map[string]any{
			"gateway_devices": []models.GatewayDevice{{
				IP: "192.168.1.1", Role: "possible_cpe", Reachable: false,
				Manufacturer: manufacturer, Model: model, AccessHints: hints,
				DeviceConfidence: 0.80, AccessConfidence: 0.80, Confidence: 0.80,
			}},
			"device_confidence":      0.80,
			"access_confidence":      0.80,
			"strong_access_evidence": true,
		},
	}
}
