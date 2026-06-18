package deviceintel

import (
	"context"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// Probe is the extension point for safe, read-only device intelligence probes.
// Default probes must stay unauthenticated and non-mutating. Credentialed
// probes (SNMP, SSH, TR-064/TR-181, router APIs) should return skipped unless
// the caller supplied explicit opt-in credentials in ScanScope.
type Probe interface {
	Name() string
	Run(ctx context.Context, scope ScanScope, store *EvidenceStore) ProbeResult
}

type ScanScope struct {
	CIDR                 string
	Interface            string
	AgentIP              string
	DefaultGateway       string
	PrivateOnly          bool
	PublicAllowed        bool
	Profile              string
	Targets              []string
	OptInProbes          []string
	SNMPCredentials      bool
	SSHCredentials       bool
	RouterAPICredentials bool
	TR064Credentials     bool
	TR181Credentials     bool
	Now                  func() time.Time
}

type ProbeResult struct {
	ProbeName      string        `json:"probe_name"`
	Status         string        `json:"status"`
	Observations   []Observation `json:"observations,omitempty"`
	SkippedTargets []string      `json:"skipped_targets,omitempty"`
	Errors         []string      `json:"errors,omitempty"`
	StartedAt      time.Time     `json:"started_at,omitempty"`
	FinishedAt     time.Time     `json:"finished_at,omitempty"`
}

type Observation struct {
	ID            string         `json:"id"`
	DeviceID      string         `json:"device_id,omitempty"`
	IP            string         `json:"ip,omitempty"`
	SourceProbe   string         `json:"source_probe"`
	Kind          string         `json:"kind"`
	Raw           map[string]any `json:"raw,omitempty"`
	Normalized    map[string]any `json:"normalized,omitempty"`
	Confidence    float64        `json:"confidence"`
	Timestamp     time.Time      `json:"timestamp"`
	TTL           time.Duration  `json:"ttl,omitempty"`
	Error         string         `json:"error,omitempty"`
	SafeToDisplay bool           `json:"safe_to_display"`
	EvidenceIDs   []string       `json:"evidence_ids,omitempty"`
}

type ProbeTarget struct {
	IP       string
	Ports    []int
	Private  bool
	DeviceID string
}

func toDeviceIntelEvidence(obs Observation) models.DeviceIntelEvidence {
	return models.DeviceIntelEvidence{
		ID:            obs.ID,
		DeviceID:      obs.DeviceID,
		SourceProbe:   obs.SourceProbe,
		Kind:          obs.Kind,
		Raw:           obs.Raw,
		Normalized:    obs.Normalized,
		Confidence:    obs.Confidence,
		Timestamp:     obs.Timestamp.UTC(),
		TTLSeconds:    int64(obs.TTL.Seconds()),
		Error:         obs.Error,
		SafeToDisplay: obs.SafeToDisplay,
	}
}
