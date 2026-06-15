package evidence

type EvidenceClass string

const (
	EvidencePhysical    EvidenceClass = "physical"
	EvidenceDevice      EvidenceClass = "device"
	EvidenceNetwork     EvidenceClass = "network"
	EvidencePerformance EvidenceClass = "performance"
	EvidenceRegional    EvidenceClass = "regional"
)

type EvidenceStrength string

const (
	StrengthNone   EvidenceStrength = "none"
	StrengthWeak   EvidenceStrength = "weak"
	StrengthMedium EvidenceStrength = "medium"
	StrengthStrong EvidenceStrength = "strong"
)

type NormalizedEvidence struct {
	ID            string           `json:"id"`
	Class         EvidenceClass    `json:"class"`
	Strength      EvidenceStrength `json:"strength"`
	SourceProbe   string           `json:"source_probe"`
	SourceField   string           `json:"source_field,omitempty"`
	TargetCategory string          `json:"target_category,omitempty"`
	TargetType     string          `json:"target_type,omitempty"`
	TargetSubtype  string          `json:"target_subtype,omitempty"`
	Confidence     float64         `json:"confidence"`
	RawValue       string          `json:"raw_value,omitempty"`
	Reason         string          `json:"reason"`
}

