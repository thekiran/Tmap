package model

import "time"

// EvidenceStrength describes how directly an observation supports a device,
// edge, or classification. Weak evidence is useful context, but must not create
// high-confidence conclusions by itself.
type EvidenceStrength string

const (
	EvidenceStrong EvidenceStrength = "strong"
	EvidenceMedium EvidenceStrength = "medium"
	EvidenceWeak   EvidenceStrength = "weak"
	EvidenceNone   EvidenceStrength = "none"
)

// Evidence is one observed, read-only fact. Every device, edge, and inference in
// the topology contract should be traceable to one or more Evidence items.
type Evidence struct {
	Source     string           `json:"source"`
	Target     string           `json:"target"`
	Strength   EvidenceStrength `json:"strength"`
	Confidence float64          `json:"confidence"`
	Reason     string           `json:"reason"`
	Raw        map[string]any   `json:"raw,omitempty"`
	Timestamp  time.Time        `json:"timestamp"`
}

func NewEvidence(source, target string, strength EvidenceStrength, confidence float64, reason string, raw map[string]any, ts time.Time) Evidence {
	return Evidence{
		Source:     source,
		Target:     target,
		Strength:   strength,
		Confidence: Clamp01(confidence),
		Reason:     reason,
		Raw:        raw,
		Timestamp:  ts,
	}
}
