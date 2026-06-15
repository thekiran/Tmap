package probes

import (
	"context"
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestSNMPProbeDisabledByDefault(t *testing.T) {
	res, err := (SNMPProbeOptIn{}).Run(context.Background(), models.ScanInput{})
	if err != nil {
		t.Fatalf("SNMP probe returned error: %v", err)
	}
	if res.Status != models.StatusSkipped {
		t.Fatalf("status = %q, want skipped", res.Status)
	}
	if enabled, _ := res.Evidence["enabled"].(bool); enabled {
		t.Fatalf("SNMP should be disabled by default: %#v", res.Evidence)
	}
	if res.Confidence != 0 {
		t.Fatalf("confidence = %v, want 0", res.Confidence)
	}
}
