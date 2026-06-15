package detection

import "github.com/thekiran/iad/pkg/models"

// classify ranks the scores and returns the leading type plus up to three
// alternatives (excluding the leader). It is a thin convenience wrapper over
// rankAll; the engine's decision layer may still override the leader to
// "Unknown" when the evidence is too weak to commit. Ties are broken by name for
// stable output.
func classify(scores map[string]float64) (primary string, alternatives []models.TypeScore) {
	ranked := rankAll(scores)
	if len(ranked) == 0 {
		return "", nil
	}
	return ranked[0].Type, capN(ranked[1:], 3)
}
