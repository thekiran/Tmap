package modemcollector

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestEvidenceStoreHTTPTimeoutDoesNotOverrideReachable(t *testing.T) {
	store := NewEvidenceStore()
	store.MergeDevice(models.GatewayDevice{
		IP: "192.168.31.1", Role: "default_gateway", Reachable: true,
		ReachableState: string(models.TriTrue), OpenPorts: []int{80},
		HTTPObservations: []models.HTTPObservation{{Source: "http_v2", URL: "http://192.168.31.1/", StatusCode: 200}},
	})
	store.MergeDevice(models.GatewayDevice{
		IP: "192.168.31.1", Role: "default_gateway", ReachableState: string(models.TriUnknown),
		FailedAttempts: []models.ProbeAttempt{{Source: "http_v3", Target: "192.168.31.1", URL: "http://192.168.31.1/", Error: "i/o timeout", Timeout: true}},
	})

	devs := store.Devices()
	if len(devs) != 1 {
		t.Fatalf("devices len = %d, want 1", len(devs))
	}
	if !devs[0].Reachable || devs[0].ReachableState != string(models.TriTrue) {
		t.Fatalf("reachable overwritten: %#v", devs[0])
	}
	if len(devs[0].FailedAttempts) != 1 {
		t.Fatalf("failed attempts not preserved: %#v", devs[0])
	}
}

func TestCandidateBuilderPrivateOnlyAndPrioritizesUpstreamCPE(t *testing.T) {
	builder := CandidateBuilder{}
	devs := builder.Build(CandidateInput{
		DefaultGateway: "192.168.31.1",
		AgentIP:        "192.168.31.147",
		ChainState: &models.GatewayChainState{
			PrivateHops: []models.GatewayHop{
				{IP: "192.168.31.1", Role: "default_gateway", Source: "traceroute_probe"},
				{IP: "192.168.1.1", Role: "upstream_private_gateway", Source: "traceroute_probe"},
				{IP: "8.8.8.8", Role: "public_hop", Source: "traceroute_probe"},
			},
			Confidence: 0.65,
		},
		ManualTargets: []string{"192.168.31.147", "8.8.4.4"},
	})

	if len(devs) != 2 {
		t.Fatalf("devices = %#v, want default + upstream private only", devs)
	}
	var upstream models.GatewayDevice
	for _, d := range devs {
		if d.IP == "192.168.1.1" {
			upstream = d
		}
		if d.IP == "8.8.8.8" || d.IP == "192.168.31.147" {
			t.Fatalf("excluded target included: %#v", devs)
		}
	}
	if upstream.Role != "upstream_private_gateway" || upstream.Confidence < 0.55 {
		t.Fatalf("upstream not prioritized: %#v", upstream)
	}
}

func TestBuildModemCollectionCurrentScanShape(t *testing.T) {
	res := models.ScanResult{
		Classification: models.Classification{
			PrimaryType: "Unknown", Confidence: 0.09, DecisionQuality: "low", SafeToDisplayAsFinal: false,
		},
		Scores: map[string]float64{models.TypeFiber: 0.07, models.TypeVDSL: 0.07},
		Candidates: []models.AccessCandidate{
			{Category: models.CatFiber, Type: models.TypeFiber, Score: 0.07},
			{Category: models.CatDSL, Type: models.TypeVDSL, Score: 0.07},
		},
		UncertaintyReasons: []string{"No strong physical-layer evidence of the access type was found."},
		DetectedNetworkContext: &models.NetworkContext{
			Gateway:           "192.168.31.1",
			DoubleNATPossible: true,
			GatewayChainState: &models.GatewayChainState{
				PrivateHops: []models.GatewayHop{
					{IP: "192.168.31.1", Role: "default_gateway", Source: "traceroute_probe"},
					{IP: "192.168.1.1", Role: "upstream_private_gateway", Source: "traceroute_probe"},
				},
				InternalDoubleNATPossible: true,
				Sources: []models.GatewayChainSource{{
					Source: "traceroute_probe", Chain: []string{"192.168.31.1", "192.168.1.1"}, InternalDoubleNATPossible: true,
				}},
				Confidence: 0.65,
			},
			GatewayDevices: []models.GatewayDevice{
				{
					IP: "192.168.31.1", Role: "default_gateway", Reachable: true, ReachableState: string(models.TriTrue),
					OpenPorts: []int{80, 443, 8443}, ServerHeader: "nginx", FaviconHash: "252c9ce330c5f06d",
					DeviceConfidence: 0.50, Confidence: 0.50,
				},
				{
					IP: "192.168.1.1", Role: "upstream_private_gateway", ReachableState: string(models.TriUnknown),
					FailedAttempts: []models.ProbeAttempt{{Source: "upstream_private_cpe_probe", Target: "192.168.1.1", Port: 80, Error: "tcp connect failed"}},
				},
			},
		},
	}

	got := Build(BuildInput{Result: res})
	if got.AccessClassification.PrimaryType != "Unknown" || got.AccessClassification.SafeToDisplayAsFinal {
		t.Fatalf("classification = %#v", got.AccessClassification)
	}
	if got.NormalizedGatewayChain.InternalDoubleNATPossible != models.TriTrue {
		t.Fatalf("double NAT = %#v", got.NormalizedGatewayChain)
	}
	if len(got.CPECandidates) != 2 {
		t.Fatalf("candidates = %#v", got.CPECandidates)
	}
	upstream := got.CPECandidates[0]
	if upstream.IP != "192.168.1.1" || upstream.Priority != "high" {
		t.Fatalf("upstream candidate not first/high priority: %#v", got.CPECandidates)
	}
	if upstream.WANPhysicalEvidence.Status != "missing" {
		t.Fatalf("failed upstream probe produced WAN evidence: %#v", upstream.WANPhysicalEvidence)
	}
}

func TestGenericNginxAndUnknownFaviconRemainDeviceEvidenceOnly(t *testing.T) {
	got := Build(BuildInput{Result: models.ScanResult{
		Classification: models.Classification{PrimaryType: "Unknown"},
		DetectedNetworkContext: &models.NetworkContext{
			Gateway: "192.168.31.1",
			GatewayDevices: []models.GatewayDevice{{
				IP: "192.168.31.1", Role: "default_gateway", Reachable: true, ReachableState: string(models.TriTrue),
				ServerHeader: "nginx", FaviconHash: "unknownhash", DeviceConfidence: 0.30, Confidence: 0.30,
			}},
		},
	}})

	if len(got.CPECandidates) != 1 {
		t.Fatalf("candidates = %#v", got.CPECandidates)
	}
	c := got.CPECandidates[0]
	if c.WANPhysicalEvidence.Status != "missing" || c.ModelFingerprint.Confidence != 0 {
		t.Fatalf("generic web evidence was misclassified: %#v", c)
	}
	if c.SNMP.Status != "skipped" || c.SNMP.Enabled {
		t.Fatalf("SNMP default state = %#v", c.SNMP)
	}
}

func TestMissingNATEvidenceStaysUnknownNotFalse(t *testing.T) {
	got := Build(BuildInput{Result: models.ScanResult{
		Classification:           models.Classification{PrimaryType: "Unknown"},
		DetectedNetworkContext:   &models.NetworkContext{Gateway: "192.168.31.1"},
		UncertaintyReasons:       []string{"public IP probe failed"},
		ClassificationConfidence: 0,
	}})

	if got.NAT.CGNAT != models.TriUnknown ||
		got.NAT.PublicIPMatches != models.TriUnknown ||
		got.NAT.InternalDoubleNATPossible != models.TriUnknown {
		t.Fatalf("missing NAT evidence must remain unknown: %#v", got.NAT)
	}
}
