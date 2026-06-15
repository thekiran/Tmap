package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
	"github.com/thekiran/iad/internal/safety"
)

type SNMPOptInProbe struct{}

func (SNMPOptInProbe) Name() string            { return "snmp_optin_probe" }
func (SNMPOptInProbe) SafeModeAllowed() bool   { return false }
func (SNMPOptInProbe) NormalModeAllowed() bool { return false }
func (SNMPOptInProbe) DeepModeAllowed() bool   { return true }

func (p SNMPOptInProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	select {
	case <-ctx.Done():
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusFailed}, ctx.Err()
	default:
	}
	if input.SNMP == nil || (input.SNMP.Community == "" && input.SNMP.Username == "") {
		return skippedResult(p.Name(), "SNMP disabled: explicit read-only credentials were not provided; no community guessing or brute force attempted"), nil
	}
	for _, ip := range input.CandidateIPs {
		if !safety.IsPrivateIPString(ip) {
			continue
		}
		ev := baseEvidence(p.Name(), ip, model.EvidenceWeak, 0.10, "SNMP is enabled, but no SNMP backend is configured in this build.", map[string]any{"target": ip, "read_only": true})
		return model.ProbeResult{
			ProbeName:  p.Name(),
			Status:     model.ProbeStatusSkipped,
			Confidence: 0.10,
			Evidence:   []model.Evidence{ev},
			Errors:     []string{"SNMP backend not configured; no community guessing or brute force attempted"},
		}, nil
	}
	return skippedResult(p.Name(), "no private SNMP targets"), nil
}
