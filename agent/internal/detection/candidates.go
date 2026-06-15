package detection

import (
	"sort"

	"github.com/thekiran/iad/pkg/models"
)

func buildAccessCandidates(scores map[string]float64, bag evidenceBag, matched bool) []models.AccessCandidate {
	ranked := collapseParentSubtypeCandidates(rankAll(scores))
	out := make([]models.AccessCandidate, 0, len(ranked))
	strength := "weak"
	if hasStrongPhysicalEvidence(bag, matched) {
		strength = "strong"
	} else if bag.DeviceEvidence > 0 {
		strength = "medium"
	}
	for _, ts := range ranked {
		c := candidateForType(ts.Type)
		c.Score = ts.Score
		c.Confidence = ts.Score
		c.EvidenceStrength = strength
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		return out[i].Subtype < out[j].Subtype
	})
	return out
}

func candidateForType(t string) models.AccessCandidate {
	switch t {
	case models.TypeVDSL2:
		return models.AccessCandidate{Category: models.CatDSL, Type: models.TypeVDSL, Subtype: models.TypeVDSL2}
	case models.TypeVDSL:
		return models.AccessCandidate{Category: models.CatDSL, Type: models.TypeVDSL}
	case models.TypeADSL2:
		return models.AccessCandidate{Category: models.CatDSL, Type: models.TypeADSL, Subtype: models.TypeADSL2}
	case models.TypeADSL:
		return models.AccessCandidate{Category: models.CatDSL, Type: models.TypeADSL}
	case models.TypeGPON:
		return models.AccessCandidate{Category: models.CatFiber, Type: models.TypeFTTH, Subtype: models.TypeGPON}
	case models.TypeFTTH:
		return models.AccessCandidate{Category: models.CatFiber, Type: models.TypeFTTH}
	case models.TypeDOCSIS:
		return models.AccessCandidate{Category: models.CatCable, Type: models.TypeCable, Subtype: models.TypeDOCSIS}
	case models.TypeFWA:
		return models.AccessCandidate{Category: models.CatMobile, Type: models.TypeFWA}
	case models.TypeLTE:
		return models.AccessCandidate{Category: models.CatMobile, Type: models.TypeMobile, Subtype: models.TypeLTE}
	default:
		return models.AccessCandidate{Category: models.CategoryFor(t), Type: t}
	}
}

