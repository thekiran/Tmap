package models

import "time"

// DeviceIntelSchema is the normalized LAN device intelligence contract used by
// topology-map UIs and inventory exports.
const DeviceIntelSchema = "iad.device_intel/v1"

// Device intelligence edge types. These are UI-facing names that make inferred
// paths distinct from physical adjacency.
const (
	DeviceEdgeDefaultGatewayRoute  = "default_gateway_route"
	DeviceEdgeSameSubnetInferred   = "same_subnet_inferred"
	DeviceEdgeARPNeighbor          = "arp_neighbor"
	DeviceEdgeTracerouteHop        = "traceroute_hop"
	DeviceEdgeLLDPPhysicalNeighbor = "lldp_physical_neighbor"
	DeviceEdgeCDPPhysicalNeighbor  = "cdp_physical_neighbor"
	DeviceEdgeSNMPBridgeFDB        = "snmp_bridge_fdb"
	DeviceEdgeWiFiAssociation      = "wifi_association_unknown"
	DeviceEdgeUpstreamPrivate      = "upstream_private_gateway"
	DeviceEdgePossibleCPEPath      = "possible_cpe_path"
)

// Common device intelligence roles and type labels.
const (
	DeviceTypeUnknown         = "unknown"
	DeviceTypeGatewayRouter   = "gateway_router"
	DeviceTypeUpstreamCPE     = "upstream_gateway_cpe_candidate"
	DeviceTypeAccessPoint     = "access_point"
	DeviceTypeSwitch          = "switch"
	DeviceTypeWindowsHost     = "windows_pc_or_server"
	DeviceTypeLinuxHost       = "linux_host"
	DeviceTypeAppleDevice     = "macos_ios_device"
	DeviceTypeAndroidDevice   = "android_device"
	DeviceTypePrinter         = "printer"
	DeviceTypeIPCamera        = "ip_camera"
	DeviceTypeMediaDevice     = "smart_tv_or_media_device"
	DeviceTypeNAS             = "nas"
	DeviceTypeIoT             = "iot_device"
	DeviceTypeGameConsole     = "game_console"
	DeviceTypeServer          = "server"
	DeviceTypeVirtualMachine  = "virtual_machine"
	DeviceRoleUpstreamGateway = "upstream_private_gateway"
	DeviceRolePossibleCPE     = "possible_cpe"
	DeviceRolePrinter         = "printer"
	DeviceRoleMedia           = "media"
	DeviceRoleStorage         = "storage"
	DeviceRoleVirtual         = "virtual"
	DeviceRoleUnknownHost     = "unknown_host"
	DeviceRoleManagement      = "management_interface"
	DeviceRoleNameResolution  = "name_resolution"
)

type DeviceIntelReport struct {
	SchemaVersion string                `json:"schema_version"`
	ScanID        string                `json:"scan_id"`
	CreatedAt     time.Time             `json:"created_at"`
	Scope         DeviceIntelScope      `json:"scope"`
	Summary       DeviceIntelSummary    `json:"summary"`
	Devices       []DeviceIntelDevice   `json:"devices"`
	Edges         []DeviceIntelEdge     `json:"edges"`
	Evidence      []DeviceIntelEvidence `json:"evidence"`
	Conflicts     []DataConflict        `json:"conflicts,omitempty"`
	Warnings      []Warning             `json:"warnings,omitempty"`
	UI            DeviceIntelUI         `json:"ui"`
	SecurityNotes []string              `json:"security_notes,omitempty"`
	Undetermined  []string              `json:"undetermined_without_credentials_or_physical_evidence,omitempty"`
}

type DeviceIntelScope struct {
	CIDR           string   `json:"cidr"`
	Interface      string   `json:"interface,omitempty"`
	AgentIP        string   `json:"agent_ip,omitempty"`
	DefaultGateway string   `json:"default_gateway,omitempty"`
	PrivateOnly    bool     `json:"private_only"`
	PublicAllowed  bool     `json:"public_allowed"`
	Profile        string   `json:"profile"`
	Targets        []string `json:"targets,omitempty"`
	OptInProbes    []string `json:"opt_in_probes,omitempty"`
}

type DeviceIntelSummary struct {
	DeviceCount           int `json:"device_count"`
	GatewayCount          int `json:"gateway_count"`
	UnknownDeviceCount    int `json:"unknown_device_count"`
	ServiceCount          int `json:"service_count"`
	HighConfidenceDevices int `json:"high_confidence_devices"`
	InferredEdges         int `json:"inferred_edges"`
	PhysicalEdges         int `json:"physical_edges"`
	SecurityFindingCount  int `json:"security_finding_count"`
}

type DeviceIntelDevice struct {
	ID                        string               `json:"id"`
	IPAddresses               []string             `json:"ip_addresses,omitempty"`
	MACAddresses              []string             `json:"mac_addresses,omitempty"`
	Hostnames                 []string             `json:"hostnames,omitempty"`
	Vendor                    DeviceVendor         `json:"vendor"`
	MobileFingerprint         *MobileFingerprint   `json:"mobileFingerprint,omitempty"`
	DeviceTypeHint            string               `json:"deviceTypeHint,omitempty"`
	OSHint                    string               `json:"osHint,omitempty"`
	OSConfidence              float64              `json:"osConfidence,omitempty"`
	Roles                     []string             `json:"roles,omitempty"`
	DeviceType                DeviceTypeGuess      `json:"device_type"`
	OSGuess                   OSGuess              `json:"os_guess"`
	Services                  []DeviceIntelService `json:"services,omitempty"`
	HTTPFingerprints          []HTTPObservation    `json:"http_fingerprints,omitempty"`
	TLSFingerprints           []TLSObservation     `json:"tls_fingerprints,omitempty"`
	SMBInfo                   *SMBInfo             `json:"smb_info"`
	MDNSRecords               []MDNSRecord         `json:"mdns_records,omitempty"`
	SSDPRecords               []SSDPRecord         `json:"ssdp_records,omitempty"`
	NBNSRecords               []NBNSRecord         `json:"nbns_records,omitempty"`
	LLMNRRecords              []LLMNRRecord        `json:"llmnr_records,omitempty"`
	PrinterInfo               *PrinterInfo         `json:"printer_info"`
	UPnPInfo                  *UPnPInfo            `json:"upnp_info"`
	SNMPInfo                  *SNMPInfo            `json:"snmp_info"`
	LLDPCDPInfo               *LLDPCDPInfo         `json:"lldp_cdp_info"`
	SecurityPosture           SecurityPosture      `json:"security_posture"`
	Topology                  DeviceTopologyFacts  `json:"topology"`
	Confidence                float64              `json:"confidence"`
	EvidenceIDs               []string             `json:"evidence_ids,omitempty"`
	Conflicts                 []DataConflict       `json:"conflicts,omitempty"`
	FailedAttempts            []ProbeAttempt       `json:"failed_attempts,omitempty"`
	LastSeen                  string               `json:"last_seen,omitempty"`
	ClassificationExplanation []string             `json:"classification_explanation,omitempty"`
	UndeterminedWithoutOptIn  []string             `json:"undetermined_without_credentials_or_physical_evidence,omitempty"`

	// Upstream-gateway enrichment (see internal/upstream). These are populated
	// for gateway/upstream/CPE candidates by the dedicated read-only enrichment
	// phase; they stay nil/empty for ordinary LAN devices.
	Reachability       *DeviceReachability `json:"reachability,omitempty"`
	RoutingEvidence    *RoutingEvidence    `json:"routing_evidence,omitempty"`
	ClassificationTags []string            `json:"classification_tags,omitempty"`
	IntelEvidence      []IntelEvidenceItem `json:"intel_evidence,omitempty"`
	EnrichmentWarnings []string            `json:"enrichment_warnings,omitempty"`
}

// DeviceReachability captures how (and whether) a device answered safe,
// read-only reachability probes. ICMP fields come from a short system ping;
// when ICMP is blocked the values fall back to TCP-connect timing.
type DeviceReachability struct {
	ICMP              bool     `json:"icmp"`
	TCPReachable      bool     `json:"tcp_reachable"`
	AvgLatencyMs      *float64 `json:"avg_latency_ms,omitempty"`
	MinLatencyMs      *float64 `json:"min_latency_ms,omitempty"`
	MaxLatencyMs      *float64 `json:"max_latency_ms,omitempty"`
	PacketLoss        *float64 `json:"packet_loss,omitempty"`
	TTL               *int     `json:"ttl,omitempty"`
	HopDistance       *int     `json:"hop_distance,omitempty"`
	DirectlyReachable bool     `json:"directly_reachable"`
	Method            string   `json:"method"` // icmp_ping | tcp_connect | none
	Note              string   `json:"note,omitempty"`
}

// RoutingEvidence is the topology/routing interpretation of an upstream device.
type RoutingEvidence struct {
	// Kind is one of: default_gateway, upstream_private_gateway,
	// double_nat_upstream, isp_cpe, bridged, stale_route, virtual_or_docker,
	// unreachable_inferred, unknown.
	Kind              string   `json:"kind"`
	DoubleNAT         bool     `json:"double_nat"`
	PrivateUpstream   bool     `json:"private_upstream"`
	SameSubnetAsAgent bool     `json:"same_subnet_as_agent"`
	HopDistance       *int     `json:"hop_distance,omitempty"`
	Notes             []string `json:"notes,omitempty"`
}

// IntelEvidenceItem is a single weighted signal that fed the upstream
// classification. The sum of ConfidenceImpact values (clamped to [0,1]) is the
// device confidence. Named distinctly from the access-detection EvidenceItem.
type IntelEvidenceItem struct {
	Type             string  `json:"type"`
	Value            string  `json:"value"`
	Source           string  `json:"source"`
	ConfidenceImpact float64 `json:"confidence_impact"`
	Timestamp        string  `json:"timestamp"`
}

type DeviceVendor struct {
	OUIVendor         *string `json:"oui_vendor"`
	FingerprintVendor *string `json:"fingerprint_vendor"`
	Confidence        float64 `json:"confidence"`
}

type DeviceTypeGuess struct {
	Primary         string                `json:"primary"`
	Candidates      []DeviceTypeCandidate `json:"candidates,omitempty"`
	Alternatives    []DeviceTypeCandidate `json:"alternatives,omitempty"`
	MissingEvidence []string              `json:"missing_evidence,omitempty"`
	Confidence      float64               `json:"confidence"`
	EvidenceIDs     []string              `json:"evidence_ids,omitempty"`
}

type DeviceTypeCandidate struct {
	Type            string   `json:"type"`
	Confidence      float64  `json:"confidence"`
	SupportingFacts []string `json:"supporting_facts,omitempty"`
	MissingEvidence []string `json:"missing_evidence,omitempty"`
	EvidenceIDs     []string `json:"evidence_ids,omitempty"`
}

type OSGuess struct {
	Family      string   `json:"family"`
	Name        *string  `json:"name"`
	Version     *string  `json:"version"`
	Confidence  float64  `json:"confidence"`
	Evidence    []string `json:"evidence,omitempty"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"`
}

type DeviceIntelService struct {
	Port        int      `json:"port"`
	Protocol    string   `json:"protocol"`
	State       string   `json:"state"`
	Name        string   `json:"name,omitempty"`
	Product     string   `json:"product,omitempty"`
	Version     string   `json:"version,omitempty"`
	Banner      string   `json:"banner,omitempty"`
	Confidence  float64  `json:"confidence"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"`
}

type SMBInfo struct {
	NetBIOSName     string   `json:"netbios_name,omitempty"`
	Workgroup       string   `json:"workgroup,omitempty"`
	OSFamily        string   `json:"os_family,omitempty"`
	SigningRequired *bool    `json:"signing_required,omitempty"`
	EvidenceIDs     []string `json:"evidence_ids,omitempty"`
}

type MDNSRecord struct {
	Name        string   `json:"name,omitempty"`
	Service     string   `json:"service,omitempty"`
	Target      string   `json:"target,omitempty"`
	Text        []string `json:"txt,omitempty"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"`
}

type SSDPRecord struct {
	USN         string   `json:"usn,omitempty"`
	ST          string   `json:"st,omitempty"`
	Server      string   `json:"server,omitempty"`
	Location    string   `json:"location,omitempty"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"`
}

type NBNSRecord struct {
	Name        string   `json:"name,omitempty"`
	Workgroup   string   `json:"workgroup,omitempty"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"`
}

type LLMNRRecord struct {
	Name        string   `json:"name,omitempty"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"`
}

type PrinterInfo struct {
	Detected     bool     `json:"detected"`
	Protocols    []string `json:"protocols,omitempty"`
	Manufacturer string   `json:"manufacturer,omitempty"`
	Model        string   `json:"model,omitempty"`
	EvidenceIDs  []string `json:"evidence_ids,omitempty"`
}

type UPnPInfo struct {
	Detected     bool     `json:"detected"`
	DeviceType   string   `json:"device_type,omitempty"`
	FriendlyName string   `json:"friendly_name,omitempty"`
	IGD          bool     `json:"igd,omitempty"`
	EvidenceIDs  []string `json:"evidence_ids,omitempty"`
}

type SNMPInfo struct {
	Enabled     bool     `json:"enabled"`
	Status      string   `json:"status"`
	Reason      string   `json:"reason,omitempty"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"`
}

type LLDPCDPInfo struct {
	Protocol     string   `json:"protocol,omitempty"`
	SystemName   string   `json:"system_name,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	PortID       string   `json:"port_id,omitempty"`
	EvidenceIDs  []string `json:"evidence_ids,omitempty"`
}

type SecurityPosture struct {
	RiskLevel string            `json:"risk_level"`
	Findings  []SecurityFinding `json:"findings,omitempty"`
	Notes     []string          `json:"notes,omitempty"`
}

type SecurityFinding struct {
	ID                 string   `json:"id"`
	Severity           string   `json:"severity"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	EvidenceIDs        []string `json:"evidence_ids,omitempty"`
	SafeRecommendation string   `json:"safe_recommendation,omitempty"`
}

type DeviceTopologyFacts struct {
	EdgeHints                  []DeviceEdgeHint `json:"edge_hints,omitempty"`
	IsGateway                  bool             `json:"is_gateway"`
	IsAgent                    bool             `json:"is_agent"`
	IsUpstreamGatewayCandidate bool             `json:"is_upstream_gateway_candidate"`
	PhysicalAdjacencyProven    bool             `json:"physical_adjacency_proven"`
	InferredOnly               bool             `json:"inferred_only"`
}

type DeviceEdgeHint struct {
	EdgeID          string   `json:"edge_id,omitempty"`
	Type            string   `json:"type"`
	Peer            string   `json:"peer,omitempty"`
	Confidence      float64  `json:"confidence"`
	ConfidenceLabel string   `json:"confidence_label,omitempty"`
	EvidenceIDs     []string `json:"evidence_ids,omitempty"`
}

type DeviceIntelEdge struct {
	ID              string   `json:"id"`
	Source          string   `json:"source"`
	Target          string   `json:"target"`
	Type            string   `json:"type"`
	Confidence      float64  `json:"confidence"`
	ConfidenceLabel string   `json:"confidence_label"`
	EvidenceIDs     []string `json:"evidence_ids,omitempty"`
	Reason          string   `json:"reason"`
	Inferred        bool     `json:"inferred"`
	Physical        bool     `json:"physical"`
}

type DeviceIntelEvidence struct {
	ID            string         `json:"id"`
	DeviceID      string         `json:"device_id,omitempty"`
	SourceProbe   string         `json:"source_probe"`
	Kind          string         `json:"kind"`
	Raw           map[string]any `json:"raw,omitempty"`
	Normalized    map[string]any `json:"normalized,omitempty"`
	Confidence    float64        `json:"confidence"`
	Timestamp     time.Time      `json:"timestamp"`
	TTLSeconds    int64          `json:"ttl_seconds,omitempty"`
	Error         string         `json:"error,omitempty"`
	SafeToDisplay bool           `json:"safe_to_display"`
}

type DeviceIntelUI struct {
	Headline      string       `json:"headline"`
	Badges        []string     `json:"badges,omitempty"`
	DeviceCards   []DeviceCard `json:"device_cards,omitempty"`
	TopologyNotes []string     `json:"topology_notes,omitempty"`
}

type DeviceCard struct {
	DeviceID         string   `json:"device_id"`
	Title            string   `json:"title"`
	Role             string   `json:"role"`
	Confidence       float64  `json:"confidence"`
	MACVendor        string   `json:"mac_vendor,omitempty"`
	Hostnames        []string `json:"hostnames,omitempty"`
	MobileLabel      string   `json:"mobile_label,omitempty"`
	MobileOSHint     string   `json:"mobile_os_hint,omitempty"`
	MobileConfidence float64  `json:"mobile_confidence,omitempty"`
	OpenServices     []string `json:"open_services,omitempty"`
	DeviceType       string   `json:"device_type"`
	OSGuess          string   `json:"os_guess"`
	LastSeen         string   `json:"last_seen,omitempty"`
	EvidenceSources  []string `json:"evidence_sources,omitempty"`
	RiskNotes        []string `json:"risk_notes,omitempty"`
	Explanation      []string `json:"why_this_classification,omitempty"`
}
