package detection

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestCanonicalGatewayChainStateTraceroutePrivateHopsDoubleNAT(t *testing.T) {
	res := analyzeSynthetic(t,
		models.ProbeResult{
			ProbeName: "gateway_probe",
			Status:    models.StatusSuccess,
			Evidence:  map[string]any{"gateway": "192.168.31.1"},
		},
		models.ProbeResult{
			ProbeName:  "traceroute_probe",
			Status:     models.StatusSuccess,
			Confidence: 0.25,
			Evidence: map[string]any{
				"hops": []string{"192.168.31.1", "192.168.1.1", "95.15.180.1"},
			},
		})

	ctx := res.DetectedNetworkContext
	if ctx == nil || ctx.GatewayChainState == nil {
		t.Fatalf("missing gateway_chain_state: %#v", ctx)
	}
	if !ctx.GatewayChainState.InternalDoubleNATPossible || !ctx.DoubleNATPossible {
		t.Fatalf("double NAT not set: %#v", ctx.GatewayChainState)
	}
	if got := ctx.GatewayChainState.Chain; len(got) != 2 || got[0] != "192.168.31.1" || got[1] != "192.168.1.1" {
		t.Fatalf("chain = %v, want [192.168.31.1 192.168.1.1]", got)
	}
}

func TestPublicIPMatchDoesNotErasePrivateUpstreamDoubleNAT(t *testing.T) {
	res := analyzeSynthetic(t,
		models.ProbeResult{
			ProbeName: "public_ip_probe",
			Status:    models.StatusSuccess,
			Evidence:  map[string]any{"public_ip": "95.15.182.146", "cgnat": false},
		},
		models.ProbeResult{
			ProbeName: "traceroute_probe",
			Status:    models.StatusSuccess,
			Evidence:  map[string]any{"hops": []string{"192.168.31.1", "192.168.1.1", "95.15.180.1"}},
		},
		models.ProbeResult{
			ProbeName: "stun_pcp_nat_probe",
			Status:    models.StatusSuccess,
			Evidence: map[string]any{
				"nat_topology": models.NATTopology{
					PublicIP: "95.15.182.146", STUNPublicIP: "95.15.182.146",
					PublicIPMatches: true, ExternalPublicIPConsistent: true,
					DoubleNAT: false, InternalDoubleNATPossible: false,
				},
				"network_confidence": 0.35,
			},
		})

	nat := res.DetectedNetworkContext.NATTopology
	if nat == nil || !nat.InternalDoubleNATPossible || !nat.DoubleNAT {
		t.Fatalf("private upstream hop must survive STUN/public-IP match: %#v", nat)
	}
}

func TestHTTPSuccessThenHTTPFailureKeepsGatewayReachable(t *testing.T) {
	res := analyzeSynthetic(t,
		models.ProbeResult{
			ProbeName: "http_fingerprint_v2",
			Status:    models.StatusSuccess,
			Evidence: map[string]any{"gateway_devices": []models.GatewayDevice{{
				IP: "192.168.31.1", Role: "default_gateway", Reachable: true,
				ReachableState: models.ReachableTrue, ServerHeader: "nginx", FaviconHash: "252c9ce330c5f06d",
				DeviceConfidence: 0.50, Confidence: 0.50,
			}}},
		},
		models.ProbeResult{
			ProbeName: "http_fingerprint_v3",
			Status:    models.StatusSuccess,
			Evidence: map[string]any{"gateway_devices": []models.GatewayDevice{{
				IP: "192.168.31.1", Role: "default_gateway", ReachableState: models.ReachableUnknown,
				FailedAttempts: []models.ProbeAttempt{{
					Source: "http_fingerprint_v3", Target: "192.168.31.1", URL: "http://192.168.31.1/",
					Method: "GET", Error: "i/o timeout", Timeout: true,
				}},
			}}},
		})

	devs := res.DetectedNetworkContext.GatewayDevices
	if len(devs) != 1 {
		t.Fatalf("gateway_devices len = %d, want 1: %#v", len(devs), devs)
	}
	if !devs[0].Reachable || devs[0].ReachableState != models.ReachableTrue {
		t.Fatalf("reachable was overwritten by later failure: %#v", devs[0])
	}
	if len(devs[0].FailedAttempts) == 0 {
		t.Fatalf("failed attempt was not preserved: %#v", devs[0])
	}
}

func TestGenericNginxGatewayIsNotPhysicalAccessEvidence(t *testing.T) {
	res := analyzeSynthetic(t, models.ProbeResult{
		ProbeName: "http_fingerprint_v2",
		Status:    models.StatusSuccess,
		Evidence: map[string]any{"gateway_devices": []models.GatewayDevice{{
			IP: "192.168.31.1", Role: "default_gateway", Reachable: true,
			ReachableState: models.ReachableTrue, ServerHeader: "nginx",
			DeviceConfidence: 0.25, Confidence: 0.25,
		}}},
	})

	if res.PrimaryType != "Unknown" {
		t.Fatalf("generic nginx gateway classified access: primary=%s scores=%v", res.PrimaryType, res.Scores)
	}
	if res.EvidenceTiers.DirectPhysical.Present {
		t.Fatalf("generic nginx must not be direct physical evidence: %#v", res.EvidenceTiers.DirectPhysical)
	}
}

func TestTurkTelekomPTRDoesNotProvePhysicalAccess(t *testing.T) {
	res := analyzeSynthetic(t, models.ProbeResult{
		ProbeName: "asn_probe",
		Status:    models.StatusSuccess,
		Evidence:  map[string]any{"ptr": "95.15.182.146.static.ttnet.com.tr", "org": "TurkTelekom"},
		Hints:     []string{models.TypeFiber, models.TypeVDSL},
	})
	if res.PrimaryType != "Unknown" {
		t.Fatalf("PTR/provider-only evidence classified access: primary=%s scores=%v", res.PrimaryType, res.Scores)
	}
}

func TestThreeMillisecondLatencyDoesNotProveFiber(t *testing.T) {
	res := analyzeSynthetic(t, models.ProbeResult{
		ProbeName: "latency_probe",
		Status:    models.StatusSuccess,
		Evidence:  map[string]any{"avg_ms": 3.0, "jitter_ms": 0.3},
	})
	if res.PrimaryType != "Unknown" || res.Classification.PrimaryType != "Unknown" {
		t.Fatalf("latency-only classified access: %#v scores=%v", res.Classification, res.Scores)
	}
}

func TestFiberVDSLTiedWeakScoresReturnUnknown(t *testing.T) {
	res := analyzeSynthetic(t,
		models.ProbeResult{
			ProbeName: "latency_probe",
			Status:    models.StatusSuccess,
			Evidence:  map[string]any{"avg_ms": 3.0, "jitter_ms": 0.3},
		},
		models.ProbeResult{
			ProbeName: "asn_probe",
			Status:    models.StatusSuccess,
			Evidence:  map[string]any{"ptr": "dynamic.ttnet.com.tr", "org": "TurkTelekom"},
			Hints:     []string{models.TypeFiber, models.TypeVDSL},
		})
	if res.PrimaryType != "Unknown" {
		t.Fatalf("weak Fiber/VDSL race classified access: primary=%s scores=%v", res.PrimaryType, res.Scores)
	}
	if res.Classification.SafeToDisplayAsFinal {
		t.Fatalf("weak tied candidates must not be final-safe: %#v", res.Classification)
	}
}
