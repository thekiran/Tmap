package detection

import (
	"strings"

	"github.com/thekiran/iad/internal/scoring"
	"github.com/thekiran/iad/pkg/models"
)

func addHintsWithAudit(board *scoring.Board, hints []string, weight float64, class, strength, probe, reason string, out *[]models.ScoreContribution) {
	board.AddHints(hints, weight)
	for _, h := range hints {
		if h == "" || weight == 0 {
			continue
		}
		*out = append(*out, scoreContribution(h, weight, class, strength, probe, reason))
	}
}

func addScoresWithAudit(board *scoring.Board, scores map[string]float64, class, strength, probe, reason string, out *[]models.ScoreContribution) {
	board.Add(scores)
	for target, amount := range scores {
		if amount == 0 {
			continue
		}
		*out = append(*out, scoreContribution(target, amount, class, strength, probe, reason))
	}
}

func scoreContribution(target string, amount float64, class, strength, probe, reason string) models.ScoreContribution {
	candidate := candidateForType(target)
	return models.ScoreContribution{
		Target:        target,
		Category:      candidate.Category,
		Type:          candidate.Type,
		Subtype:       candidate.Subtype,
		Amount:        amount,
		EvidenceClass: class,
		Strength:      strength,
		ProbeName:     probe,
		Reason:        reason,
	}
}

func classifyRuleEvidence(ruleID string) (class, strength string) {
	id := strings.ToLower(ruleID)
	switch {
	case strings.Contains(id, "cpe_wan"), strings.Contains(id, "docsis"):
		return "physical", "strong"
	case strings.Contains(id, "model"), strings.Contains(id, "router"), strings.Contains(id, "wisp"):
		return "device", "medium"
	case strings.Contains(id, "cgnat"), strings.Contains(id, "isp"), strings.Contains(id, "ptr"):
		return "regional", "weak"
	default:
		return "device", "medium"
	}
}

