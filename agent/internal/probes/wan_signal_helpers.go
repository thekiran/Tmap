package probes

import (
	"strconv"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

func wanHintsFromText(text string) []string {
	hints := inferAccessHints(text)
	l := strings.ToLower(text)
	if strings.Contains(l, "pots") {
		hints = appendUnique(hints, models.TypeDSL)
		hints = appendUnique(hints, models.TypeADSL)
	}
	if strings.Contains(l, "cable") {
		hints = appendUnique(hints, models.TypeCable)
	}
	if strings.Contains(l, "fiber") || strings.Contains(l, "fibre") || strings.Contains(l, "ftth") {
		hints = appendUnique(hints, models.TypeFiber)
		hints = appendUnique(hints, models.TypeFTTH)
	}
	return hints
}

func wanSignal(source, ip, typ, value, detail string, hints []string) models.WANSignal {
	strength := string(models.EvidenceNetwork)
	conf := 0.45
	if len(hints) > 0 {
		strength = string(models.EvidencePhysical)
		conf = 0.80
	}
	return models.WANSignal{
		Source:     source,
		IP:         ip,
		Type:       typ,
		Value:      strings.TrimSpace(value),
		Strength:   strength,
		Detail:     strings.TrimSpace(detail),
		Confidence: conf,
	}
}

func accessConfidenceFromHints(hints []string, foundCPE bool) float64 {
	if len(hints) == 0 {
		return 0
	}
	conf := 0.60
	if foundCPE {
		conf += 0.20
	}
	if conf > 1 {
		return 1
	}
	return conf
}

func parseBitrate(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
