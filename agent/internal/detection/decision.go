package detection

import (
	"sort"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

// Decision thresholds. These encode the project's core stance: this is
// probabilistic, evidence-based analysis — not certain detection. When the
// evidence is weak, contested, or lacks anything that physically proves the
// access type, the verdict is Unknown (but the candidate scores are kept).
const (
	minConfidence   = 0.45 // below this → Unknown
	minTopScore     = 0.35 // top category score below this → Unknown
	minCategoryGap  = 0.12 // category margin below this → Unknown (too close)
	highConfidence  = 0.70
	highTopScore    = 0.60
	highCategoryGap = 0.20
)

// CandidatePair holds the two leading entries of a score map and their margin.
type CandidatePair struct {
	First       string
	FirstScore  float64
	Second      string
	SecondScore float64
	Margin      float64
}

// topTwo returns the two highest-scoring keys of m and their gap.
func topTwo(m map[string]float64) CandidatePair {
	var p CandidatePair
	for k, v := range m {
		switch {
		case v > p.FirstScore:
			p.Second, p.SecondScore = p.First, p.FirstScore
			p.First, p.FirstScore = k, v
		case v > p.SecondScore:
			p.Second, p.SecondScore = k, v
		}
	}
	p.Margin = p.FirstScore - p.SecondScore
	return p
}

// topTwoScores returns the leading pair of raw type scores (helper from the spec).
func topTwoScores(scores map[string]float64) CandidatePair {
	return topTwo(scores)
}

// categoryScores collapses per-type scores to per-category scores using the max
// type score in each category. This is what the decision layer compares, so that
// two types in the *same* category (e.g. DSL vs VDSL) being close does not look
// like ambiguity — only competition *between* categories (DSL vs Fiber vs Cable)
// counts as a close, uncertain race.
func categoryScores(scores map[string]float64) map[string]float64 {
	cat := map[string]float64{}
	for t, s := range scores {
		c := models.CategoryFor(t)
		if c == models.CatUnknown {
			c = t
		}
		if s > cat[c] {
			cat[c] = s
		}
	}
	return cat
}

// rankAll returns every positive-scoring type, highest first, ties broken by
// name for stable output.
func rankAll(scores map[string]float64) []models.TypeScore {
	out := make([]models.TypeScore, 0, len(scores))
	for t, s := range scores {
		if s > 0 {
			out = append(out, models.TypeScore{Type: t, Score: s})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Type < out[j].Type
	})
	return out
}

func collapseParentSubtypeCandidates(ranked []models.TypeScore) []models.TypeScore {
	parent := map[string]bool{
		models.TypeDSL:   true,
		models.TypeFiber: true,
		models.TypeCable: true,
	}
	hasSubtype := map[string]bool{}
	for _, ts := range ranked {
		cat := models.CategoryFor(ts.Type)
		if cat != ts.Type && ts.Score > 0 {
			hasSubtype[cat] = true
		}
	}
	out := make([]models.TypeScore, 0, len(ranked))
	for _, ts := range ranked {
		if parent[ts.Type] && hasSubtype[models.CategoryFor(ts.Type)] {
			continue
		}
		out = append(out, ts)
	}
	return out
}

// strongEvidenceTokens are substrings that, if present in the evidence text or a
// (router-side) interface name, prove the physical access medium.
var strongEvidenceTokens = []string{
	"ptm0", "atm0", "dsl0", "pppoe-wan", "eth-wan", "gpon", "epon", "ont",
	"docsis", "cable modem", "vdsl2", "adsl2+", "line rate", "line attenuation",
	"snr margin", "wwan", "lte0",
}

var gatewayStrongEvidenceTokens = []string{
	"vdsl", "adsl", "dsl", "ptm", "atm", "gpon", "epon", "ont", "docsis",
	"cable modem", "lte", "5g cpe", "wan dsl", "wan gpon",
}

// genericServerTokens are HTTP Server-header values that identify only the web
// stack, never the physical access type. They must never produce an access hint
// on their own (spec §A, §B).
var genericServerTokens = []string{
	"nginx", "apache", "lighttpd", "openresty", "caddy", "go", "microsoft-iis",
}

// hasStrongPhysicalEvidence reports whether anything in the evidence actually
// proves the access type. A modem fingerprint match counts; so do WAN-side
// interface names and DSL/GPON/DOCSIS markers. PTR, ASN, public IP, latency and
// local Ethernet/Wi-Fi names do NOT — they are weak/contextual only.
func hasStrongPhysicalEvidence(bag evidenceBag, matched bool) bool {
	if matched {
		return true
	}
	if bag.PhysicalEvidence > 0 {
		return true
	}
	if len(bag.StrongAccessHints) > 0 {
		return true
	}
	for _, sig := range bag.WANSignals {
		if sig.Strength == string(models.EvidencePhysical) || strings.EqualFold(sig.Strength, "strong") || sig.Confidence >= 0.70 {
			return true
		}
	}
	text := strings.ToLower(bag.Text)
	for _, tok := range strongEvidenceTokens {
		if strings.Contains(text, tok) {
			return true
		}
	}
	gatewayText := strings.ToLower(bag.GatewayDeviceText)
	for _, tok := range gatewayStrongEvidenceTokens {
		if strings.Contains(gatewayText, tok) {
			return true
		}
	}
	for _, ifc := range bag.Interfaces {
		l := strings.ToLower(ifc)
		for _, tok := range strongEvidenceTokens {
			if strings.Contains(l, tok) {
				return true
			}
		}
	}
	return false
}

// shouldReturnUnknown applies the decision gate. It returns whether the verdict
// must be downgraded to Unknown, plus the human-readable reasons (which always
// reflect every condition that contributed, for explainability).
func shouldReturnUnknown(scores map[string]float64, confidence float64, bag evidenceBag, matched bool) (bool, []string) {
	cat := topTwo(categoryScores(scores))
	strong := hasStrongPhysicalEvidence(bag, matched)

	var reasons []string
	if confidence < minConfidence {
		reasons = append(reasons, "Classification confidence is low.")
	}
	if cat.FirstScore < minTopScore {
		reasons = append(reasons, "The top score is low.")
	}
	if cat.Margin < minCategoryGap {
		reasons = append(reasons, "The leading category scores are too close to call.")
	}
	if !strong {
		reasons = append(reasons, "No strong physical-layer evidence of the access type was found.")
	}
	if !matched && !bag.UPnPFound {
		reasons = append(reasons, "No UPnP modem model was discovered.")
	}
	if bag.PTR != "" && !strong {
		reasons = append(reasons, "The PTR record is not conclusive evidence of the access type.")
	}

	unknown := confidence < minConfidence ||
		cat.FirstScore < minTopScore ||
		cat.Margin < minCategoryGap ||
		!strong

	return unknown, reasons
}

// decisionQuality grades the verdict's trustworthiness. Strong physical evidence
// with high confidence and a clear lead is "high"; strong evidence OR a solid
// statistical margin is "medium"; everything else is "low".
func decisionQuality(confidence, margin, topScore float64, strong bool) string {
	if !strong {
		return "low"
	}
	if strong && confidence >= highConfidence && topScore >= highTopScore && margin >= highCategoryGap {
		return "high"
	}
	if confidence >= minConfidence && margin >= minCategoryGap && topScore >= minTopScore {
		return "medium"
	}
	return "low"
}
