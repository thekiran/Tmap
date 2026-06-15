package scoring

// Board accumulates per-access-type scores in "points" and records which rules
// fired, so the engine can both rank types and explain the verdict.
type Board struct {
	scores map[string]float64
	fired  []string
}

// NewBoard returns an empty scoreboard.
func NewBoard() *Board {
	return &Board{scores: map[string]float64{}}
}

// Add applies a set of score deltas (e.g. a fired rule's add_score).
func (b *Board) Add(deltas map[string]float64) {
	for t, v := range deltas {
		b.scores[t] += v
	}
}

// AddHints adds the same weight to every listed type/category key. Used for
// fingerprint access hints and probe hints.
func (b *Board) AddHints(hints []string, weight float64) {
	for _, h := range hints {
		b.scores[h] += weight
	}
}

// MarkFired records a rule id that contributed to the score.
func (b *Board) MarkFired(id string) {
	b.fired = append(b.fired, id)
}

// Fired returns the ids of the rules that fired.
func (b *Board) Fired() []string {
	return b.fired
}

// Raw returns the accumulated point scores (unnormalized).
func (b *Board) Raw() map[string]float64 {
	return b.scores
}

// Normalize maps raw points to a 0..1 confidence-like scale by treating 100
// points as "full" and clamping. This matches the doc's convention where a raw
// score of 82 surfaces as 0.82.
func (b *Board) Normalize() map[string]float64 {
	out := make(map[string]float64, len(b.scores))
	for t, v := range b.scores {
		n := v / 100.0
		if n > 1 {
			n = 1
		}
		if n < 0 {
			n = 0
		}
		out[t] = n
	}
	return out
}
