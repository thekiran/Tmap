// Package ippool implements the Continuous IP Discovery Pool: a controlled,
// read-only, rate-limited service that grows a pool of reachable private/local
// devices from seed IPs and emits live updates to the Wails frontend.
//
// SAFETY CONTRACT (enforced by ScopeGuard + the scheduler):
//   - Only RFC1918 / link-local / RFC4193 targets are ever probed.
//   - Probes are read-only reachability checks (system ping + TCP connect).
//   - No brute force, logins, exploits, config changes, or destructive traffic.
//   - Strict concurrency + rate limits with jitter; no infinite tight loops.
//   - Subnets larger than /24 require explicit user confirmation.
//
// It lives in the desktop module (not the agent) because it is a long-running
// service that emits Wails events; the agent is a single-shot CLI.
package ippool

import "time"

// DeviceStatus is the lifecycle state of a pool entry.
type DeviceStatus string

const (
	StatusCandidate    DeviceStatus = "candidate"
	StatusActive       DeviceStatus = "active"
	StatusRecentlySeen DeviceStatus = "recently_seen"
	StatusStale        DeviceStatus = "stale"
	StatusUnreachable  DeviceStatus = "unreachable"
)

// Source records how an IP first entered the pool.
type Source string

const (
	SourceSeed      Source = "seed"
	SourceGateway   Source = "gateway"
	SourceARP       Source = "arp"
	SourceRoute     Source = "route"
	SourceDNS       Source = "dns"
	SourceCandidate Source = "candidate"
	SourcePassive   Source = "passive_packet"
)

// EvidenceStrength grades a single evidence item.
type EvidenceStrength string

const (
	StrengthConfirmed EvidenceStrength = "confirmed"
	StrengthInferred  EvidenceStrength = "inferred"
	StrengthWeak      EvidenceStrength = "weak"
)

// EvidenceItem records why an IP exists in the pool.
type EvidenceItem struct {
	Type             string           `json:"type"`
	Source           string           `json:"source"`
	Value            string           `json:"value"`
	Timestamp        string           `json:"timestamp"`
	ConfidenceImpact float64          `json:"confidenceImpact"`
	Strength         EvidenceStrength `json:"strength"`
}

// MobileEvidenceItem is one normalized, UI-safe reason used by the mobile OS
// fingerprint. Values intentionally avoid raw DNS browsing history or payloads.
type MobileEvidenceItem struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Value            string `json:"value"`
	OSHint           string `json:"osHint"`
	ConfidenceImpact int    `json:"confidenceImpact"`
	Strength         string `json:"strength"`
	Source           string `json:"source"`
	Timestamp        string `json:"timestamp"`
	Explanation      string `json:"explanation"`
}

// MobileConflictItem preserves contradictory Apple/Android evidence instead of
// hiding it behind a false single-answer classification.
type MobileConflictItem struct {
	Reason             string   `json:"reason"`
	IOSEvidenceIDs     []string `json:"iosEvidenceIds,omitempty"`
	AndroidEvidenceIDs []string `json:"androidEvidenceIds,omitempty"`
	Severity           string   `json:"severity"`
	ResolutionHint     string   `json:"resolutionHint"`
}

// MobileFingerprint is the live-registry copy of the agent report shape. Keep
// JSON names aligned with the frontend and final scan reports.
type MobileFingerprint struct {
	Classification        string               `json:"classification"`
	IOSScore              int                  `json:"iosScore"`
	AndroidScore          int                  `json:"androidScore"`
	IPadScore             int                  `json:"ipadScore"`
	Confidence            float64              `json:"confidence"`
	Evidence              []MobileEvidenceItem `json:"evidence,omitempty"`
	Conflicts             []MobileConflictItem `json:"conflicts,omitempty"`
	Warnings              []string             `json:"warnings,omitempty"`
	LastUpdatedAt         string               `json:"lastUpdatedAt,omitempty"`
	WhyThisClassification string               `json:"whyThisClassification,omitempty"`
	WhyNotCertain         string               `json:"whyNotCertain,omitempty"`
}

// DevicePoolEntry is one tracked device.
type DevicePoolEntry struct {
	ID                string             `json:"id"`
	IP                string             `json:"ip"`
	MAC               string             `json:"mac,omitempty"`
	Hostname          string             `json:"hostname,omitempty"`
	Vendor            string             `json:"vendor,omitempty"`
	FirstSeen         string             `json:"firstSeen"`
	LastSeen          string             `json:"lastSeen,omitempty"`
	LastProbeAt       string             `json:"lastProbeAt,omitempty"`
	Status            DeviceStatus       `json:"status"`
	ResponseCount     int                `json:"responseCount"`
	FailureCount      int                `json:"failureCount"`
	AvgLatencyMs      *float64           `json:"avgLatencyMs,omitempty"`
	TTL               *int               `json:"ttl,omitempty"`
	Source            Source             `json:"source"`
	Evidence          []EvidenceItem     `json:"evidence,omitempty"`
	MobileFingerprint *MobileFingerprint `json:"mobileFingerprint,omitempty"`
	DeviceTypeHint    string             `json:"deviceTypeHint,omitempty"`
	OSHint            string             `json:"osHint,omitempty"`
	OSConfidence      float64            `json:"osConfidence,omitempty"`
	OSEvidenceSummary []string           `json:"osEvidenceSummary,omitempty"`
}

// Wails event names emitted by the manager.
const (
	EvtStarted         = "ip_pool:started"
	EvtSeedAdded       = "ip_pool:seed_added"
	EvtCandidateAdded  = "ip_pool:candidate_added"
	EvtProbeResult     = "ip_pool:probe_result"
	EvtDeviceFound     = "ip_pool:device_found"
	EvtDeviceUpdated   = "ip_pool:device_updated"
	EvtDeviceStale     = "ip_pool:device_stale"
	EvtTopologyUpdated = "ip_pool:topology_updated"
	EvtWarning         = "ip_pool:warning"
	EvtStopped         = "ip_pool:stopped"
	EvtStatus          = "ip_pool:status"

	EvtDiscoveryDeviceUpdated         = "discovery:device_updated"
	EvtDiscoveryTopologyUpdated       = "discovery:topology_updated"
	EvtDeviceMobileFingerprintUpdated = "discovery:device_mobile_fingerprint_updated"
)

// MobileFingerprintUpdatedPayload is emitted only when the classification or
// scoring materially changes.
type MobileFingerprintUpdatedPayload struct {
	DeviceID          string            `json:"deviceId"`
	IPAddresses       []string          `json:"ipAddresses"`
	Hostname          string            `json:"hostname,omitempty"`
	MobileFingerprint MobileFingerprint `json:"mobileFingerprint"`
	UpdatedAt         string            `json:"updatedAt"`
}

// Emitter decouples the manager from Wails so it can be unit-tested. In the app
// it is backed by runtime.EventsEmit; in tests by a recording fake.
type Emitter interface {
	Emit(event string, data any)
}

// ProbeResultPayload is emitted (batched) per probe.
type ProbeResultPayload struct {
	IP        string  `json:"ip"`
	Reachable bool    `json:"reachable"`
	LatencyMs float64 `json:"latencyMs,omitempty"`
	TTL       int     `json:"ttl,omitempty"`
	Method    string  `json:"method"`
	Timestamp int64   `json:"timestamp"`
}

// StatusSnapshot is the live monitoring summary for the frontend panel.
type StatusSnapshot struct {
	Running        bool     `json:"running"`
	Seeds          []string `json:"seeds"`
	ActiveCount    int      `json:"activeCount"`
	RecentlyCount  int      `json:"recentlyCount"`
	StaleCount     int      `json:"staleCount"`
	CandidateCount int      `json:"candidateCount"`
	CandidateQueue int      `json:"candidateQueue"`
	Subnets        []string `json:"subnets"`
	LastProbeAt    int64    `json:"lastProbeAt,omitempty"`
	ProbesPerSec   float64  `json:"probesPerSec"`
	Concurrency    int      `json:"concurrency"`
	Warnings       []string `json:"warnings,omitempty"`
}

// Config holds all the rate / timeout / scope knobs with safe defaults.
type Config struct {
	MaxConcurrency        int           // worker pool size
	MaxProbesPerSec       float64       // global rate limit
	PingTimeout           time.Duration // per-probe timeout
	PingCount             int           // echoes per probe
	ActiveInterval        time.Duration // active-pool retest cadence (15–60s)
	StaleInterval         time.Duration // stale-pool retest cadence (2–5m)
	CandidateRate         float64       // candidate probes per second (low, background)
	BatchInterval         time.Duration // event batching/coalescing window
	JitterFraction        float64       // 0..1 random jitter on intervals
	RecentlyAfter         int           // consecutive failures: active -> recently_seen
	StaleAfter            int           // consecutive failures: recently_seen -> stale
	CandidateFailMax      int           // candidate failures before it stops being retried
	MaxAutoPrefix         int           // largest auto subnet (24)
	WarnPrefix            int           // warn at /23
	BlockPrefix           int           // block /16 (and larger) without confirmation
	MaxPoolSize           int           // memory bound on tracked devices
	PerformanceMode       bool          // reduce concurrency/rate when set
	MobileRefreshInterval time.Duration // low-frequency mobile re-scoring cadence
}

// DefaultConfig returns the conservative, network-friendly defaults.
func DefaultConfig() Config {
	return Config{
		MaxConcurrency:        16,
		MaxProbesPerSec:       8,
		PingTimeout:           time.Second,
		PingCount:             1,
		ActiveInterval:        20 * time.Second,
		StaleInterval:         3 * time.Minute,
		CandidateRate:         4,
		BatchInterval:         750 * time.Millisecond,
		JitterFraction:        0.25,
		RecentlyAfter:         1,
		StaleAfter:            3,
		CandidateFailMax:      2,
		MaxAutoPrefix:         24,
		WarnPrefix:            23,
		BlockPrefix:           16,
		MaxPoolSize:           4096,
		PerformanceMode:       false,
		MobileRefreshInterval: time.Minute,
	}
}

// normalized applies performance mode and clamps to safe minimums.
func (c Config) normalized() Config {
	if c.MaxConcurrency <= 0 {
		c.MaxConcurrency = 16
	}
	if c.PerformanceMode {
		if c.MaxConcurrency > 6 {
			c.MaxConcurrency = 6
		}
		if c.MaxProbesPerSec > 4 {
			c.MaxProbesPerSec = 4
		}
	}
	if c.MaxProbesPerSec <= 0 {
		c.MaxProbesPerSec = 8
	}
	if c.PingTimeout <= 0 {
		c.PingTimeout = time.Second
	}
	if c.BatchInterval <= 0 {
		c.BatchInterval = 750 * time.Millisecond
	}
	if c.MaxPoolSize <= 0 {
		c.MaxPoolSize = 4096
	}
	if c.MobileRefreshInterval <= 0 {
		c.MobileRefreshInterval = time.Minute
	}
	return c
}
