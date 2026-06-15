// Package models holds the shared data contracts exchanged between the probe
// layer, the detection engine and the report/CLI layers. Keeping them in one
// place means the future UI, local API and SQLite storage can all consume the
// exact same JSON shapes without reimplementing types.
package models

import "time"

// Probe status values.
const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// Scan modes.
const (
	ModeQuick = "quick"
	ModeDeep  = "deep"
)

// ProbeResult is the single output format produced by every probe (doc §10).
// A uniform shape is what lets new probes drop in without touching the engine
// or the UI.
type ProbeResult struct {
	ProbeName  string         `json:"probe_name"`
	Status     string         `json:"status"`
	Confidence float64        `json:"confidence"`
	Evidence   map[string]any `json:"evidence,omitempty"`
	Hints      []string       `json:"hints,omitempty"`
	Errors     []string       `json:"errors,omitempty"`
}

// ScanInput configures a single scan run.
type ScanInput struct {
	Mode     string // ModeQuick | ModeDeep
	Online   bool   // when false, no probe may contact an external service
	RulesDir string // directory holding the YAML rule/fingerprint files
}

// TypeScore is one access type with its normalized (0..1) score.
type TypeScore struct {
	Type  string  `json:"type"`
	Score float64 `json:"score"`
}

// GatewayDevice is the local discovery result for one observed private gateway
// IP. It is intentionally factual: the detection engine may use the text as
// evidence, but the report also shows what was actually reachable.
type GatewayDevice struct {
	IP                         string   `json:"ip"`
	Role                       string   `json:"role,omitempty"`
	Reachable                  bool     `json:"reachable"`
	HTTPTitle                  string   `json:"http_title,omitempty"`
	ServerHeader               string   `json:"server_header,omitempty"`
	WWWAuthenticate            string   `json:"www_authenticate,omitempty"`
	FaviconHash                string   `json:"favicon_hash,omitempty"`
	WWWAuthRealm               string   `json:"www_authenticate_realm,omitempty"`
	RedirectPath               string   `json:"redirect_path,omitempty"`
	RedirectLocation           string   `json:"redirect_location,omitempty"`
	TLSCertCN                  string   `json:"tls_cert_cn,omitempty"`
	TLSCertSANs                []string `json:"tls_cert_sans,omitempty"`
	TLSServerName              string   `json:"tls_server_name,omitempty"`
	HTMLMetaGenerator          string   `json:"html_meta_generator,omitempty"`
	LoginLabels                []string `json:"login_labels,omitempty"`
	UPnPFound                  bool     `json:"upnp_found"`
	UPnPIGDFound               bool     `json:"upnp_igd_found"`
	WANCommonInterfaceFound    bool     `json:"wan_common_interface_found"`
	WANAccessType              string   `json:"wan_access_type,omitempty"`
	PhysicalLinkStatus         string   `json:"physical_link_status,omitempty"`
	Layer1UpstreamMaxBitRate   int64    `json:"layer1_upstream_max_bitrate,omitempty"`
	Layer1DownstreamMaxBitRate int64    `json:"layer1_downstream_max_bitrate,omitempty"`
	TR064Found                 bool     `json:"tr064_found"`
	TR064AuthRequired          bool     `json:"tr064_auth_required"`
	TR064Services              []string `json:"tr064_services,omitempty"`
	MACVendor                  string   `json:"mac_vendor,omitempty"`
	Model                      string   `json:"model,omitempty"`
	Manufacturer               string   `json:"manufacturer,omitempty"`
	FingerprintID              string   `json:"fingerprint_id,omitempty"`
	AccessHints                []string `json:"access_hints,omitempty"`
	PhysicalHints              []string `json:"physical_hints,omitempty"`
	Notes                      []string `json:"notes,omitempty"`
	DeviceConfidence           float64  `json:"device_confidence"`
	AccessConfidence           float64  `json:"access_confidence"`
	Confidence                 float64  `json:"confidence"`
}

// WANSignal is a confirmed WAN-side CPE signal. It is kept separate from
// generic gateway reachability so the classifier can distinguish "a web UI was
// reachable" from "the CPE explicitly exposed DSL/GPON/DOCSIS/LTE evidence".
type WANSignal struct {
	Source     string  `json:"source"`
	IP         string  `json:"ip,omitempty"`
	Type       string  `json:"type,omitempty"`
	Value      string  `json:"value,omitempty"`
	Strength   string  `json:"strength,omitempty"`
	Detail     string  `json:"detail,omitempty"`
	Confidence float64 `json:"confidence"`
}

// EvidenceStrength describes the evidence class used by the decision layer and
// confidence breakdown.
type EvidenceStrength string

// Evidence strength classes used by the decision layer and confidence
// breakdown. They are intentionally broad: access classification depends on
// Physical evidence, while the other classes explain context and confidence.
const (
	EvidencePhysical    EvidenceStrength = "Physical"
	EvidenceDevice      EvidenceStrength = "Device"
	EvidenceNetwork     EvidenceStrength = "Network"
	EvidencePerformance EvidenceStrength = "Performance"
)

type AccessArchitecture struct {
	LocalMedium    string `json:"local_medium,omitempty"`
	WANMedium      string `json:"wan_medium,omitempty"`
	IPArchitecture string `json:"ip_architecture,omitempty"`
	NATTopology    string `json:"nat_topology,omitempty"`
	LikelyCPERole  string `json:"likely_cpe_role,omitempty"`
}

type ConfidenceBreakdown struct {
	// Classification is the confidence in the physical access-type verdict; it is
	// capped low when no strong physical evidence exists. Context is the
	// confidence in the surrounding network situation (ISP, NAT, IPv6,
	// performance) which can be higher even when the type is Unknown.
	Classification float64 `json:"classification"`
	Context        float64 `json:"context"`
	Physical       float64 `json:"physical"`
	Device         float64 `json:"device"`
	Network        float64 `json:"network"`
	Performance    float64 `json:"performance"`
	Regional       float64 `json:"regional,omitempty"`
	Penalty        float64 `json:"penalty"`
}

// EvidenceStrengthSummary reports the strongest evidence class observed per tier
// ("none" | "weak" | "medium" | "strong"). Physical drives classification; the
// others explain context.
type EvidenceStrengthSummary struct {
	Physical    string `json:"physical"`
	Device      string `json:"device"`
	Network     string `json:"network"`
	Performance string `json:"performance"`
	Regional    string `json:"regional,omitempty"`
}

// AccessCandidate is one ranked access possibility expressed as a
// category/type/subtype tree, so a parent (Fiber) and its child (FTTH/GPON) are
// a single candidate instead of competing flat scores.
type AccessCandidate struct {
	Category           string         `json:"category,omitempty"`
	Type               string         `json:"type"`
	Subtype            string         `json:"subtype,omitempty"`
	Score              float64        `json:"score"`
	Confidence         float64        `json:"confidence,omitempty"`
	EvidenceStrength   string         `json:"evidence_strength,omitempty"`
	SupportingEvidence []EvidenceItem `json:"supporting_evidence,omitempty"`
	MissingEvidence    []string       `json:"missing_evidence,omitempty"`
}

// Classification is the UI/API-facing verdict. PrimaryType is intentionally a
// conservative coarse value from the public contract; subtype carries details
// such as GPON, VDSL2, or DOCSIS 3.1 when the evidence supports them.
type Classification struct {
	PrimaryType          string  `json:"primary_type"`
	Subtype              *string `json:"subtype"`
	Confidence           float64 `json:"confidence"`
	DecisionQuality      string  `json:"decision_quality"`
	State                string  `json:"state"`
	SafeToDisplayAsFinal bool    `json:"safe_to_display_as_final"`
}

type EvidenceItem struct {
	Source     string         `json:"source"`
	TargetType string         `json:"target_type"`
	Strength   string         `json:"strength"`
	Confidence float64        `json:"confidence"`
	Reason     string         `json:"reason"`
	Raw        map[string]any `json:"raw,omitempty"`
}

type EvidenceTier struct {
	Status string         `json:"status"`
	Items  []EvidenceItem `json:"items"`

	// Non-serialized convenience fields used by the engine.
	Present    bool    `json:"-"`
	Strength   string  `json:"-"`
	Confidence float64 `json:"-"`
}

type EvidenceTiers struct {
	DirectPhysical EvidenceTier `json:"direct_physical"`
	DeviceModel    EvidenceTier `json:"device_model"`
	Topology       EvidenceTier `json:"topology"`
	Performance    EvidenceTier `json:"performance"`
	Regional       EvidenceTier `json:"regional"`
}

type ConflictValue struct {
	Source string `json:"source"`
	Value  any    `json:"value"`
}

type DataConflict struct {
	Field      string          `json:"field"`
	Values     []ConflictValue `json:"values"`
	Severity   string          `json:"severity"`
	Effect     string          `json:"effect"`
	Resolution string          `json:"resolution,omitempty"`
}

type DataQuality struct {
	HasConflicts bool           `json:"has_conflicts"`
	Conflicts    []DataConflict `json:"conflicts"`
}

type UIOutput struct {
	Headline string   `json:"headline"`
	Summary  string   `json:"summary"`
	Badges   []string `json:"badges,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ScoreContribution records a single traceable addition to the score, so every
// point in the verdict can be audited back to a probe and evidence class.
type ScoreContribution struct {
	Target        string  `json:"target"`
	Category      string  `json:"category,omitempty"`
	Type          string  `json:"type,omitempty"`
	Subtype       string  `json:"subtype,omitempty"`
	Amount        float64 `json:"amount"`
	EvidenceClass string  `json:"evidence_class"`
	Strength      string  `json:"strength"`
	ProbeName     string  `json:"probe_name"`
	Reason        string  `json:"reason"`
}

type NextBestProbe struct {
	ProbeName        string   `json:"probe_name"`
	Reason           string   `json:"reason"`
	ExpectedEvidence []string `json:"expected_evidence"`
	Safety           string   `json:"safety"`
}

type IPv6Context struct {
	IPv6Available   bool     `json:"ipv6_available"`
	GlobalIPv6      bool     `json:"global_ipv6"`
	DefaultRoute    string   `json:"default_route,omitempty"`
	DNS64NAT64      bool     `json:"dns64_nat64"`
	TransitionHints []string `json:"transition_hints,omitempty"`
}

type NATTopology struct {
	PublicIP        string `json:"public_ip,omitempty"`
	STUNPublicIP    string `json:"stun_public_ip,omitempty"`
	STUNPublicPort  int    `json:"stun_public_port,omitempty"`
	PublicIPMatches bool   `json:"public_ip_matches"`
	CGNAT           bool   `json:"cgnat"`
	// DoubleNAT is retained for backward compatibility; it mirrors
	// InternalDoubleNATPossible. New consumers should read the explicit fields
	// below, which the NAT topology resolver keeps internally consistent.
	DoubleNAT                  bool     `json:"double_nat"`
	InternalDoubleNATPossible  bool     `json:"internal_double_nat_possible"`
	ExternalPublicIPConsistent bool     `json:"external_public_ip_consistent"`
	GatewayNATControlReachable bool     `json:"gateway_nat_control_reachable"`
	PCPReachable               bool     `json:"pcp_reachable"`
	NATPMPReachable            bool     `json:"nat_pmp_reachable"`
	Topology                   string   `json:"topology,omitempty"`
	Notes                      []string `json:"notes,omitempty"`
}

type PerformanceProfile struct {
	Target          string  `json:"target,omitempty"`
	Method          string  `json:"method,omitempty"`
	IdleLatencyMS   float64 `json:"idle_latency_ms,omitempty"`
	JitterMS        float64 `json:"jitter_ms,omitempty"`
	PacketLossPct   float64 `json:"packet_loss_pct,omitempty"`
	LoadedLatencyMS float64 `json:"loaded_latency_ms,omitempty"`
}

// ScanResult is the full, serializable result of a scan: the verdict plus all
// the evidence that produced it, so a user can always see *why*.
type ScanResult struct {
	ScanID         string         `json:"scan_id"`
	CreatedAt      time.Time      `json:"created_at"`
	Status         string         `json:"status"`
	Mode           string         `json:"mode"`
	Classification Classification `json:"classification"`
	PrimaryType    string         `json:"primary_type"`
	Category       string         `json:"category"`
	// Confidence is the headline (classification) confidence, kept for backward
	// compatibility. ClassificationConfidence and ContextConfidence split it: the
	// former is the confidence in the physical access type (capped low without
	// strong physical evidence), the latter the confidence in the network context.
	Confidence               float64 `json:"confidence"`
	ClassificationConfidence float64 `json:"classification_confidence"`
	ContextConfidence        float64 `json:"context_confidence"`

	// DecisionQuality grades how trustworthy the verdict is ("low" | "medium" |
	// "high"), independent of which type ranked first. A low quality result with
	// a populated Scores map means "here are the candidates, but the evidence is
	// too weak to commit."
	DecisionQuality string `json:"decision_quality,omitempty"`
	// UncertaintyReasons lists, in plain language, why a definite verdict was not
	// made. Empty when the decision is confident.
	UncertaintyReasons []string `json:"uncertainty_reasons,omitempty"`
	// DetectedNetworkContext carries the factual network situation (ISP, gateway
	// chain, double-NAT, local access medium, ...) regardless of the verdict.
	DetectedNetworkContext *NetworkContext `json:"detected_network_context,omitempty"`

	Scores map[string]float64 `json:"scores"`
	// Candidates is the category/type/subtype view of the scores (preferred over
	// the flat Scores map, which is retained for backward compatibility).
	Candidates          []AccessCandidate   `json:"candidates,omitempty"`
	Alternatives        []TypeScore         `json:"alternatives"`
	Explanation         []string            `json:"explanation"`
	ConfidenceBreakdown ConfidenceBreakdown `json:"confidence_breakdown,omitempty"`
	// ScoreContributions audits every point added to the scores back to a probe
	// and evidence class, making false positives debuggable.
	ScoreContributions []ScoreContribution `json:"score_contributions,omitempty"`
	NextBestProbes     []NextBestProbe     `json:"next_best_probes,omitempty"`
	EvidenceTiers      EvidenceTiers       `json:"evidence_tiers"`
	DataQuality        DataQuality         `json:"data_quality"`
	Conflicts          []DataConflict      `json:"conflicts,omitempty"`
	UI                 UIOutput            `json:"ui"`
	Evidence           []ProbeResult       `json:"evidence"`
}

// NetworkContext is the observed, factual network situation. It is reported even
// when the access type is Unknown, so the user always sees what *was* detected
// (ISP, addresses, topology) versus what could not be concluded (the type).
type NetworkContext struct {
	ISP                string          `json:"isp,omitempty"`
	Country            string          `json:"country,omitempty"`
	Region             string          `json:"region,omitempty"`
	PublicIP           string          `json:"public_ip,omitempty"`
	PTR                string          `json:"ptr,omitempty"`
	ASN                string          `json:"asn,omitempty"`
	BGPOrg             string          `json:"bgp_org,omitempty"`
	CGNAT              bool            `json:"cgnat"`
	Gateway            string          `json:"gateway,omitempty"`
	GatewayChain       []string        `json:"gateway_chain,omitempty"`
	DoubleNATPossible  bool            `json:"double_nat_possible"`
	LocalAccess        string          `json:"local_access,omitempty"`
	MainAdapter        string          `json:"main_adapter,omitempty"`
	AdapterType        string          `json:"adapter_type,omitempty"`
	LinkSpeedMbps      int             `json:"link_speed_mbps,omitempty"`
	RouterModel        string          `json:"router_model,omitempty"`
	FingerprintMatched bool            `json:"fingerprint_matched"`
	UPnPFound          bool            `json:"upnp_found"`
	TR064Found         bool            `json:"tr064_found"`
	GatewayDevices     []GatewayDevice `json:"gateway_devices,omitempty"`
	// LikelyCPEIP is the preferred name; LikelyModemIP is retained for backward
	// compatibility and mirrors it.
	LikelyCPEIP   string      `json:"likely_cpe_ip,omitempty"`
	LikelyModemIP string      `json:"likely_modem_ip,omitempty"`
	WANSignals    []WANSignal `json:"wan_signals,omitempty"`
	// LineProfile is the fine-grained physical-layer line reading (VDSL2 profile,
	// DOCSIS version, PON optical power, line stats) when authorized CPE telemetry
	// exposed it. Nil when the line could not be read.
	LineProfile        *LineProfile             `json:"line_profile,omitempty"`
	AccessArchitecture AccessArchitecture       `json:"access_architecture,omitempty"`
	IPv6Context        *IPv6Context             `json:"ipv6_context,omitempty"`
	NATTopology        *NATTopology             `json:"nat_topology,omitempty"`
	PerformanceProfile *PerformanceProfile      `json:"performance_profile,omitempty"`
	EvidenceStrength   *EvidenceStrengthSummary `json:"evidence_strength,omitempty"`
}
