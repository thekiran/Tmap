package probes

import (
	"context"

	"github.com/thekiran/iad/pkg/models"
)

type SNMPProbeOptIn struct{}

func (SNMPProbeOptIn) Name() string { return "snmp_probe_opt_in" }

func (p SNMPProbeOptIn) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	res.Status = models.StatusSkipped
	res.Evidence["enabled"] = false
	res.Evidence["reason"] = "SNMP is disabled by default and requires explicit user-provided read-only credentials."
	res.Evidence["safety"] = "No community guessing, brute force, credential collection, or public IP probing is performed."
	res.Confidence = 0
	return res, nil
}

