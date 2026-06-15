package detection

import (
	"fmt"
	"strings"

	evnorm "github.com/thekiran/iad/internal/detection/evidence"
	"github.com/thekiran/iad/pkg/models"
)

func normalizeProbeResult(r models.ProbeResult) []evnorm.NormalizedEvidence {
	var out []evnorm.NormalizedEvidence
	add := func(class evnorm.EvidenceClass, strength evnorm.EvidenceStrength, field, target, raw, reason string, conf float64) {
		category := ""
		if target != "" {
			category = models.CategoryFor(target)
		}
		out = append(out, evnorm.NormalizedEvidence{
			ID:             fmt.Sprintf("%s:%s:%d", r.ProbeName, field, len(out)+1),
			Class:          class,
			Strength:       strength,
			SourceProbe:    r.ProbeName,
			SourceField:    field,
			TargetCategory: category,
			TargetType:     target,
			Confidence:     conf,
			RawValue:       raw,
			Reason:         reason,
		})
	}

	if getBool(r.Evidence, "strong_access_evidence") || getFloat(r.Evidence, "access_confidence") > 0 {
		for _, h := range r.Hints {
			add(evnorm.EvidencePhysical, evnorm.StrengthStrong, "access_confidence", h, strings.Join(r.Hints, ","), "CPE/WAN evidence indicates a physical access family.", getFloat(r.Evidence, "access_confidence"))
		}
	}
	if getFloat(r.Evidence, "device_confidence") > 0 {
		add(evnorm.EvidenceDevice, evnorm.StrengthMedium, "device_confidence", "", "", "Device identity or reachability evidence was observed.", getFloat(r.Evidence, "device_confidence"))
	}
	if getFloat(r.Evidence, "network_confidence") > 0 ||
		r.ProbeName == "asn_probe" ||
		strings.HasPrefix(r.ProbeName, "stun_") ||
		strings.HasPrefix(r.ProbeName, "ipv6_transition_probe") ||
		strings.HasPrefix(r.ProbeName, "gateway_reachability_diagnostics") {
		add(evnorm.EvidenceNetwork, evnorm.StrengthMedium, "network_context", "", "", "Network context was observed; it is not physical access proof.", getFloat(r.Evidence, "network_confidence"))
	}
	if getFloat(r.Evidence, "performance_confidence") > 0 || r.ProbeName == "latency_probe" || strings.HasPrefix(r.ProbeName, "performance_profile_probe") {
		add(evnorm.EvidencePerformance, evnorm.StrengthWeak, "performance_profile", "", "", "Performance evidence is contextual and not conclusive by itself.", getFloat(r.Evidence, "performance_confidence"))
	}
	if r.ProbeName == "asn_probe" && (getString(r.Evidence, "ptr") != "" || getString(r.Evidence, "org") != "") {
		add(evnorm.EvidenceRegional, evnorm.StrengthWeak, "provider_context", "", getString(r.Evidence, "ptr")+" "+getString(r.Evidence, "org"), "Provider/PTR evidence is regional context only.", 0.30)
	}
	return out
}
