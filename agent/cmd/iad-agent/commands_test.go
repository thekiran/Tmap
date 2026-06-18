package main

import (
	"testing"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

func TestApplyFullScanDefaultsUsesFullProfile(t *testing.T) {
	profile := "quick"
	classify := false
	useNmap := false
	timeout := 30 * time.Second

	applyFullScanDefaults(false, false, &profile, &classify, &useNmap, &timeout)

	if profile != "full" {
		t.Fatalf("profile = %q, want full", profile)
	}
	if !classify {
		t.Fatal("full mode must enable access classification")
	}
	if !useNmap {
		t.Fatal("full mode must request optional nmap discovery")
	}
	if timeout != 60*time.Second {
		t.Fatalf("timeout = %s, want 60s", timeout)
	}
}

func TestApplyFullScanDefaultsPreservesExplicitProfileAndTimeout(t *testing.T) {
	profile := "standard"
	classify := false
	useNmap := false
	timeout := 10 * time.Second

	applyFullScanDefaults(true, true, &profile, &classify, &useNmap, &timeout)

	if profile != "standard" {
		t.Fatalf("profile = %q, want explicit standard", profile)
	}
	if timeout != 10*time.Second {
		t.Fatalf("timeout = %s, want explicit 10s", timeout)
	}
	if !classify || !useNmap {
		t.Fatal("full mode should still enable classify and nmap")
	}
}

func TestEvidenceRegistryCoversReferencedEvidenceIDs(t *testing.T) {
	report := models.ScanReport{
		Evidence: []models.Evidence{
			{ID: "ev-root", Kind: "interface", Source: "interface_probe", Summary: "selected interface"},
		},
		Devices: []models.Device{
			{
				ID:          "device-1",
				EvidenceIDs: []string{"ev-root", "ev-device-only"},
				Services: []models.Service{
					{Port: 80, Protocol: "tcp", EvidenceIDs: []string{"ev-service-only"}},
				},
			},
		},
		Edges: []models.TopologyEdge{
			{ID: "edge-1", EvidenceIDs: []string{"ev-edge-only"}},
		},
		AccessClassification: &models.ScanResult{
			DetectedNetworkContext: &models.NetworkContext{
				GatewayDevices: []models.GatewayDevice{
					{
						IP:          "192.168.1.1",
						EvidenceIDs: []string{"ev-gateway-only"},
						FailedAttempts: []models.ProbeAttempt{
							{EvidenceID: "ev-failed-attempt"},
						},
					},
				},
			},
		},
	}

	registry := buildEvidenceRegistry(report)
	known := map[string]bool{}
	for _, item := range registry {
		known[item.ID] = true
	}
	for _, id := range []string{"ev-root", "ev-device-only", "ev-service-only", "ev-edge-only", "ev-gateway-only", "ev-failed-attempt"} {
		if !known[id] {
			t.Fatalf("registry does not contain referenced evidence id %q", id)
		}
	}
}
