package detection

import (
	"sort"
	"strings"

	"github.com/thekiran/iad/internal/scoring"
	"github.com/thekiran/iad/pkg/models"
)

// ---------------------------------------------------------------------------
// Confidence split: classification vs context (spec §A.4)
// ---------------------------------------------------------------------------

// contextConfidence is confidence in the surrounding network situation (ISP,
// NAT, IPv6, performance). It can be medium/high even when the physical access
// type is Unknown, because those signals are genuinely known.
func contextConfidence(bag evidenceBag) float64 {
	c := bag.NetworkEvidence
	c = maxFloat(c, 0.85*bag.PerformanceEvidence)
	c = maxFloat(c, 0.6*bag.DeviceEvidence)

	signals := 0
	if bag.Org != "" || bag.PTR != "" {
		signals++
	}
	if hasUsefulNATContext(bag.NATTopology) {
		signals++
	}
	if bag.IPv6Context != nil {
		signals++
	}
	if bag.HasLatency || bag.PerformanceProfile != nil {
		signals++
	}
	if bag.PublicIP != "" {
		signals++
	}
	if signals > 0 {
		c = maxFloat(c, clamp01(0.2+0.1*float64(signals)))
	}
	return clamp01(c)
}

func hasUsefulNATContext(n *models.NATTopology) bool {
	if n == nil {
		return false
	}
	if n.PublicIP != "" || n.STUNPublicIP != "" || n.CGNAT || n.InternalDoubleNATPossible || n.DoubleNAT {
		return true
	}
	if n.PCPReachable || n.NATPMPReachable || n.GatewayNATControlReachable {
		return true
	}
	return n.Topology != "" && n.Topology != "unknown"
}

// ---------------------------------------------------------------------------
// Score contribution audit (spec §A.5)
// ---------------------------------------------------------------------------

// buildContributions reconstructs the traceable list of score additions from the
// same inputs the scoreboard used, so every point is attributable.
func buildContributions(bag evidenceBag, fp *Fingerprint, matched bool, rules []scoring.Rule, fired []string) []models.ScoreContribution {
	var out []models.ScoreContribution
	add := func(target string, amount float64, class, strength, probe, reason string) {
		if amount == 0 {
			return
		}
		candidate := candidateForType(target)
		out = append(out, models.ScoreContribution{
			Target: target, Category: candidate.Category, Type: candidate.Type, Subtype: candidate.Subtype,
			Amount: amount, EvidenceClass: class,
			Strength: strength, ProbeName: probe, Reason: reason,
		})
	}

	if matched && fp != nil {
		for _, h := range fp.AccessHints {
			add(h, scoring.WeightFingerprintHint, "device", "strong", "fingerprint",
				"Known CPE model maps to this access family.")
		}
		for _, s := range fp.Supports {
			add(s, scoring.WeightFingerprintSupp, "device", "medium", "fingerprint",
				"CPE model advertises support for this technology.")
		}
	}
	for _, h := range bag.Hints {
		add(h, scoring.WeightProbeHint, "network", "weak", "probe_hints",
			"Probe hint (contextual, not physical-layer proof).")
	}
	for _, h := range bag.StrongAccessHints {
		add(h, scoring.WeightStrongAccessHint, "physical", "strong", "cpe_wan",
			"CPE exposed physical WAN access evidence.")
	}
	for _, lc := range latencyContributions(bag) {
		out = append(out, lc)
	}

	ruleByID := map[string]scoring.Rule{}
	for _, r := range rules {
		ruleByID[r.ID] = r
	}
	for _, id := range fired {
		r, ok := ruleByID[id]
		if !ok {
			continue
		}
		class, strength := ruleEvidenceClass(id)
		for target, amount := range r.Then.AddScore {
			add(target, amount, class, strength, "rule:"+id,
				"Matched rule "+id+".")
		}
	}
	return out
}

func latencyContributions(bag evidenceBag) []models.ScoreContribution {
	if !bag.HasLatency {
		return nil
	}
	mk := func(target string, amount float64) models.ScoreContribution {
		candidate := candidateForType(target)
		return models.ScoreContribution{
			Target: target, Category: candidate.Category, Type: candidate.Type, Subtype: candidate.Subtype,
			Amount: amount, EvidenceClass: "performance", Strength: "weak",
			ProbeName: "latency_probe", Reason: "Latency band is compatible with this type (not conclusive).",
		}
	}
	switch ms := bag.AvgMS; {
	case ms <= 8:
		return []models.ScoreContribution{mk(models.TypeFiber, scoring.LatLowFiber), mk(models.TypeVDSL, scoring.LatLowVDSL)}
	case ms <= 25:
		return []models.ScoreContribution{mk(models.TypeDSL, scoring.LatMidGeneric), mk(models.TypeVDSL, scoring.LatMidGeneric), mk(models.TypeFiber, scoring.LatMidGeneric), mk(models.TypeCable, scoring.LatMidGeneric)}
	case ms <= 80:
		return []models.ScoreContribution{mk(models.TypeDSL, scoring.LatDSLFWA), mk(models.TypeFWA, scoring.LatDSLFWA)}
	case ms <= 200:
		return []models.ScoreContribution{mk(models.TypeMobile, scoring.LatMobile), mk(models.TypeFWA, scoring.LatMobile)}
	case ms >= 500:
		return []models.ScoreContribution{mk(models.TypeSatellite, scoring.LatSatellite)}
	}
	return nil
}

// ruleEvidenceClass classifies a rule's contribution by its id prefix. Provider
// and PTR rules are weak regional/network evidence; interface rules are strong
// physical; model rules are medium device evidence.
func ruleEvidenceClass(id string) (class, strength string) {
	switch {
	case strings.HasPrefix(id, "iface_"):
		return "physical", "strong"
	case strings.HasPrefix(id, "ptr_"), strings.HasPrefix(id, "isp_"), strings.HasPrefix(id, "provider_"):
		return "regional", "weak"
	default:
		return "device", "medium"
	}
}

// ---------------------------------------------------------------------------
// Candidates (category/type/subtype) (spec §C)
// ---------------------------------------------------------------------------

// subtypeParent maps a fine-grained subtype to its parent type, so a Fiber/GPON
// score is expressed as category=Fiber, type=FTTH, subtype=GPON rather than as
// separate competing entries.
var subtypeParent = map[string]string{
	models.TypeGPON:   models.TypeFTTH,
	"EPON":            models.TypeFTTH,
	"XG-PON":          models.TypeFTTH,
	"XGS-PON":         models.TypeFTTH,
	"10G-EPON":        models.TypeFTTH,
	models.TypeVDSL2:  models.TypeVDSL,
	models.TypeADSL2:  models.TypeADSL,
	"ADSL2":           models.TypeADSL,
	models.TypeDOCSIS: models.TypeCable,
}

// buildCandidates groups scores by category and emits one candidate per
// category (no parent/subtype duplication), highest score first.
func buildCandidates(scores map[string]float64, bag evidenceBag, matched bool) []models.AccessCandidate {
	tiers := buildEvidenceTiers(bag, matched)
	strength := tiers.DirectPhysical.Strength
	if strength == "none" {
		strength = tiers.DeviceModel.Strength
	}
	if strength == "none" {
		strength = "weak"
	}
	byCat := map[string][]models.TypeScore{}
	for t, s := range scores {
		if s <= 0 {
			continue
		}
		c := models.CategoryFor(t)
		if c == models.CatUnknown {
			c = t
		}
		byCat[c] = append(byCat[c], models.TypeScore{Type: t, Score: s})
	}

	var out []models.AccessCandidate
	for cat, list := range byCat {
		sort.Slice(list, func(i, j int) bool {
			if list[i].Score != list[j].Score {
				return list[i].Score > list[j].Score
			}
			return list[i].Type < list[j].Type
		})
		cand := models.AccessCandidate{Category: cat, Score: list[0].Score, Confidence: list[0].Score, EvidenceStrength: strength}
		// Pick the most specific subtype present; its parent becomes the type.
		for _, ts := range list {
			if parent, ok := subtypeParent[ts.Type]; ok {
				cand.Subtype = ts.Type
				cand.Type = parent
				break
			}
		}
		if cand.Type == "" {
			// No subtype: use the top type-level key (skip the bare category name
			// unless it is the only entry).
			cand.Type = list[0].Type
			if cand.Type == cat && len(list) > 1 {
				cand.Type = list[1].Type
			}
		}
		cand.SupportingEvidence = candidateSupportingEvidence(tiers)
		cand.MissingEvidence = candidateMissingEvidence(cand, tiers)
		out = append(out, cand)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Category < out[j].Category
	})
	return out
}

// ---------------------------------------------------------------------------
// Evidence strength summary (spec §C)
// ---------------------------------------------------------------------------

func strengthLabel(v float64) string {
	switch {
	case v <= 0:
		return "none"
	case v < 0.40:
		return "weak"
	case v < 0.70:
		return "medium"
	default:
		return "strong"
	}
}

func buildEvidenceStrengthSummary(bag evidenceBag, matched bool) *models.EvidenceStrengthSummary {
	physical := bag.PhysicalEvidence
	if len(bag.StrongAccessHints) > 0 {
		physical = maxFloat(physical, 0.80)
	}
	device := bag.DeviceEvidence
	if matched {
		device = maxFloat(device, 0.80)
	}
	regional := 0.0
	if bag.Org != "" || bag.PTR != "" {
		regional = 0.30
	}
	return &models.EvidenceStrengthSummary{
		Physical:    strengthLabel(physical),
		Device:      strengthLabel(device),
		Network:     strengthLabel(bag.NetworkEvidence),
		Performance: strengthLabel(bag.PerformanceEvidence),
		Regional:    strengthLabel(regional),
	}
}

func candidateSupportingEvidence(tiers models.EvidenceTiers) []models.EvidenceItem {
	var out []models.EvidenceItem
	if tiers.DirectPhysical.Present {
		out = append(out, tiers.DirectPhysical.Items...)
	}
	if tiers.DeviceModel.Present {
		out = append(out, tiers.DeviceModel.Items...)
	}
	if tiers.Topology.Present {
		out = append(out, tiers.Topology.Items...)
	}
	if tiers.Performance.Present {
		out = append(out, tiers.Performance.Items...)
	}
	if tiers.Regional.Present {
		out = append(out, tiers.Regional.Items...)
	}
	return out
}

func candidateMissingEvidence(cand models.AccessCandidate, tiers models.EvidenceTiers) []string {
	if tiers.DirectPhysical.Present {
		return nil
	}
	switch cand.Type {
	case models.TypeFiber, models.TypeFTTH:
		return []string{
			"No optical interface evidence.",
			"No ONT/GPON/EPON/XGS-PON model with active optical WAN.",
			"No UPnP WANAccessType proving optical access.",
			"No TR-064 WAN optical data.",
			"No SNMP optical interface evidence.",
		}
	case models.TypeVDSL:
		return []string{
			"No DSL.Line evidence.",
			"No PTM/ATM evidence.",
			"No VDSL2-LINE-MIB evidence.",
			"No TR-064 DSL service data.",
			"No active DSL CPE interface detected.",
		}
	case models.TypeADSL:
		return []string{
			"No ADSL-LINE-MIB evidence.",
			"No ifType adsl(94) evidence.",
			"No TR-181 ADSL/ADSL2/ADSL2+ mode evidence.",
			"No TR-064 DSL service data.",
		}
	case models.TypeCable:
		return []string{
			"No DOCSIS MIB evidence.",
			"No docsCable interface evidence.",
			"No active cable WAN interface evidence.",
		}
	case models.TypeFWA, models.TypeMobile:
		return []string{
			"No LTE/NR/WWAN CPE interface evidence.",
			"No active cellular WAN metadata.",
		}
	case models.TypeEthernetWAN:
		return []string{
			"No explicit WANAccessType Ethernet evidence.",
			"No TR-181 Ethernet physical WAN stack evidence.",
		}
	default:
		return []string{
			"No direct physical WAN evidence.",
			"No mapped CPE model evidence.",
			"No TR-181/TR-064/SNMP/UPnP WAN physical-layer evidence.",
		}
	}
}
