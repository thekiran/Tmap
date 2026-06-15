package detection

import "github.com/thekiran/iad/pkg/models"

// computeConfidence turns the score distribution and supporting context into a
// single 0..1 confidence. It is deliberately *not* just the top score (doc §5.5):
//
//   - top category score — how strong the leading category is on its own,
//   - separation         — how far the leader is ahead of the next *category*
//     (two types in the same category being close is not ambiguity),
//   - corroboration      — how many independent probes contributed evidence,
//   - fingerprint bonus  — a known modem model is a uniquely strong signal.
//
// Comparing at the category level (DSL vs Fiber vs Cable) rather than the raw
// type level prevents an artificial confidence drop when, say, DSL and VDSL tie.
func computeConfidence(scores map[string]float64, sources int, matched bool) float64 {
	if len(scores) == 0 {
		return 0
	}
	pair := topTwo(categoryScores(scores))
	top := pair.FirstScore
	if top <= 0 {
		return 0
	}

	separation := clamp01(pair.Margin / top)
	corroboration := clamp01(float64(sources) / 4.0)

	conf := 0.6*top + 0.25*separation + 0.15*corroboration
	if matched {
		conf += 0.1
	}
	return clamp01(conf)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func computeConfidenceBreakdown(scores map[string]float64, bag evidenceBag, matched bool) models.ConfidenceBreakdown {
	device := bag.DeviceEvidence
	if matched {
		device = maxFloat(device, 0.80)
	}
	physical := bag.PhysicalEvidence
	if len(bag.StrongAccessHints) > 0 {
		physical = maxFloat(physical, 0.80)
	}
	breakdown := models.ConfidenceBreakdown{
		Physical:    clamp01(physical),
		Device:      clamp01(device),
		Network:     clamp01(bag.NetworkEvidence),
		Performance: clamp01(bag.PerformanceEvidence),
	}
	if !hasStrongPhysicalEvidence(bag, matched) {
		breakdown.Penalty = 0.35
	}
	return breakdown
}

func maxFloat(a, b float64) float64 {
	if b > a {
		return b
	}
	return a
}
