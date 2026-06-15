package detection

import (
	"sort"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

// Decision thresholds. The classifier is allowed to show a probable/possible
// result from model evidence, but only Tier A direct physical evidence can become
// a final confirmed result.
const (
	minConfidence   = 0.40 // below this -> Unknown
	minTopScore     = 0.35 // top category score below this -> Unknown
	minCategoryGap  = 0.12 // category margin below this -> Unknown unless Tier A resolves it
	highConfidence  = 0.85
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

// topTwoScores returns the leading pair of raw type scores.
func topTwoScores(scores map[string]float64) CandidatePair {
	return topTwo(scores)
}

// categoryScores collapses per-type scores to per-category scores using the max
// type score in each category. This prevents DSL vs VDSL subtype closeness from
// looking like cross-medium ambiguity.
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

var directEvidenceTokens = []string{
	"ptm0", "atm0", "dsl0", "pppoe-wan", "eth-wan", "gpon", "epon", "ont",
	"docsis", "cable modem", "vdsl2", "adsl2+", "line rate", "line attenuation",
	"snr margin", "wwan", "lte0", "wan dsl", "wan gpon", "wan ethernet",
	"wandslinterfaceconfig", "dsl.line", "dsl.channel", "ptm.link", "atm.link",
	"optical.interface", "cellular.interface", "vdsl2-line", "adsl-line",
	"docs-if", "docsis mib", "iftype adsl",
}

// genericServerTokens are HTTP Server-header values that identify only the web
// stack, never the physical access type.
var genericServerTokens = []string{
	"nginx", "apache", "lighttpd", "openresty", "caddy", "go", "microsoft-iis",
}

// hasDirectPhysicalEvidence reports Tier A evidence only. Known model names,
// PTR/ASN, public IP, latency, and local Ethernet/Wi-Fi adapter names are not
// direct proof of the WAN access medium.
func hasDirectPhysicalEvidence(bag evidenceBag) bool {
	if bag.LineProfile != nil && bag.LineProfile.Confidence > 0 {
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
	text := strings.ToLower(strings.TrimSpace(bag.WANSignalText + " " + bag.CPEText))
	for _, tok := range directEvidenceTokens {
		if strings.Contains(text, tok) {
			return true
		}
	}
	return false
}

// hasStrongPhysicalEvidence is retained for older call sites. In the new model,
// "strong physical" means Tier A direct physical WAN evidence only.
func hasStrongPhysicalEvidence(bag evidenceBag, matched bool) bool {
	return hasDirectPhysicalEvidence(bag)
}

func hasDeviceModelEvidence(bag evidenceBag, matched bool) bool {
	if matched {
		return true
	}
	if strings.TrimSpace(bag.RouterModel) != "" && bag.DeviceEvidence > 0 {
		return true
	}
	for _, d := range bag.GatewayDevices {
		if (d.Model != "" || d.Manufacturer != "" || d.FingerprintID != "") && d.DeviceConfidence >= 0.40 {
			return true
		}
	}
	return false
}

// shouldReturnUnknown applies the decision gate. It returns whether the verdict
// must be downgraded to Unknown, plus human-readable reasons.
func shouldReturnUnknown(scores map[string]float64, confidence float64, bag evidenceBag, matched bool) (bool, []string) {
	cat := topTwo(categoryScores(scores))
	direct := hasDirectPhysicalEvidence(bag)
	device := hasDeviceModelEvidence(bag, matched)

	var reasons []string
	if confidence < minConfidence {
		reasons = append(reasons, "Classification confidence is low.")
	}
	if cat.FirstScore < minTopScore {
		reasons = append(reasons, "The top score is low.")
	}
	if cat.Margin < minCategoryGap && !direct {
		reasons = append(reasons, "The leading category scores are too close to call.")
	}
	if !direct {
		reasons = append(reasons, "No strong physical-layer evidence of the access type was found.")
	}
	if !direct && !device {
		reasons = append(reasons, "Only topology, performance, or operator hints were available.")
	}
	if !matched && !bag.UPnPFound {
		reasons = append(reasons, "No UPnP modem model was discovered.")
	}
	if bag.PTR != "" && !direct {
		reasons = append(reasons, "The PTR record is not conclusive evidence of the access type.")
	}

	unknown := confidence < minConfidence ||
		cat.FirstScore < minTopScore ||
		(cat.Margin < minCategoryGap && !direct) ||
		(!direct && !device)

	return unknown, reasons
}

// decisionQuality grades the verdict's trustworthiness. Only direct physical
// evidence can be high quality; strong model evidence tops out at medium.
func decisionQuality(confidence, margin, topScore float64, direct, device bool) string {
	if direct && confidence >= highConfidence && topScore >= highTopScore && margin >= highCategoryGap {
		return "high"
	}
	if (direct || device) && confidence >= minConfidence && margin >= minCategoryGap && topScore >= minTopScore {
		return "medium"
	}
	return "low"
}
