package models

import "time"

// This file holds the stable data contracts for authorized local network
// discovery and topology mapping. They are intentionally separate from the
// access-type ScanResult: a ScanReport describes *who is on the network and how
// they connect*, with every fact backed by Evidence. JSON field names are part
// of the contract — add fields, never rename or repurpose them.

// TopologyReportSchema is the schema identifier embedded in every ScanReport so
// consumers can detect the format/version.
const TopologyReportSchema = "iad.topology/v1"

// Edge types. An edge is only as trustworthy as the evidence behind it, so the
// type records *how* the link was established. Physical adjacency (LLDP/CDP/SNMP)
// is distinguished from inference (route hops, same-subnet) so the agent never
// claims real topology it cannot prove.
const (
	EdgeDirectLLDP     = "direct_lldp"     // neighbour reported via LLDP (very high)
	EdgeDirectCDP      = "direct_cdp"       // neighbour reported via CDP (very high)
	EdgeSNMPBridge     = "snmp_bridge"      // forwarding/bridge (FDB) table via SNMP (high/medium)
	EdgeRouteHop       = "route_hop"        // adjacent hop observed in a traceroute (medium)
	EdgeInferredL2     = "inferred_l2"      // same subnet / ARP-observed; L2 adjacency inferred (medium/low)
	EdgeGatewayDefault = "gateway_default"  // the agent's default route to its gateway (fallback)
)

// Confidence labels (human-facing bucketing of the numeric confidence).
const (
	ConfVeryHigh = "very_high"
	ConfHigh     = "high"
	ConfMedium   = "medium"
	ConfLow      = "low"
	ConfVeryLow  = "very_low"
)

// Warning severities.
const (
	SeverityInfo    = "info"
	SeverityWarning = "warning"
	SeverityDanger  = "danger"
)

// Device roles (evidence-based; "host" is the default when nothing more specific
// is proven).
const (
	RoleAgent   = "agent"
	RoleGateway = "gateway"
	RoleRouter  = "router"
	RoleHost    = "host"
)

// ScanReport is the full, serializable result of a topology scan.
type ScanReport struct {
	SchemaVersion string         `json:"schema_version"`
	ScanID        string         `json:"scan_id"`
	CreatedAt     time.Time      `json:"created_at"`
	Agent         AgentInfo      `json:"agent"`
	Scope         ScanScope      `json:"scope"`
	Devices       []Device       `json:"devices"`
	Edges         []TopologyEdge `json:"edges"`
	Evidence      []Evidence     `json:"evidence"`
	Warnings      []Warning      `json:"warnings"`
	Summary       ScanSummary    `json:"summary"`
	// AccessClassification optionally carries the access-type verdict from the
	// existing detection engine, when the user runs scan with classification on.
	AccessClassification *ScanResult `json:"access_classification,omitempty"`
}

// AgentInfo describes the host running the scan.
type AgentInfo struct {
	Version    string          `json:"version"`
	Hostname   string          `json:"hostname,omitempty"`
	OS         string          `json:"os,omitempty"`
	Gateway    string          `json:"gateway,omitempty"`
	DNSServers []string        `json:"dns_servers,omitempty"`
	Interfaces []InterfaceInfo `json:"interfaces,omitempty"`
}

// InterfaceInfo is one local network interface and how it was classified.
type InterfaceInfo struct {
	Name      string      `json:"name"`
	MAC       string      `json:"mac,omitempty"`
	Up        bool        `json:"up"`
	Loopback  bool        `json:"loopback"`
	Virtual   bool        `json:"virtual"`
	Selected  bool        `json:"selected"`
	CIDR      string      `json:"cidr,omitempty"` // primary IPv4 network in CIDR form
	Addresses []IPAddress `json:"addresses,omitempty"`
}

// ScanScope is the validated boundary of the scan. The agent only scans inside
// this scope; anything outside is refused.
type ScanScope struct {
	Requested     string `json:"requested"`      // what the user asked for ("auto" or a CIDR)
	CIDR          string `json:"cidr"`           // resolved network
	Interface     string `json:"interface,omitempty"`
	HostCount     int    `json:"host_count"`     // addressable hosts in CIDR
	Private       bool   `json:"private"`        // RFC1918 / ULA / link-local
	PublicAllowed bool   `json:"public_allowed"` // true only when the dangerous flag is set
	Profile       string `json:"profile"`        // quick | standard | deep
}

// Device is a normalized host discovered on the network. Every non-trivial field
// is traceable to one or more Evidence records via EvidenceIDs.
type Device struct {
	ID          string            `json:"id"`
	Hostname    string            `json:"hostname,omitempty"`
	Vendor      string            `json:"vendor,omitempty"`
	Roles       []string          `json:"roles,omitempty"`
	IsAgent     bool              `json:"is_agent,omitempty"`
	IsGateway   bool              `json:"is_gateway,omitempty"`
	Addresses   []IPAddress       `json:"addresses"`
	Interfaces  []DeviceInterface `json:"interfaces,omitempty"`
	Services    []Service         `json:"services,omitempty"`
	Confidence  float64           `json:"confidence"`
	EvidenceIDs []string          `json:"evidence_ids,omitempty"`
}

// DeviceInterface is a link-layer interface of a device (MAC + bound IPs).
type DeviceInterface struct {
	MAC    string   `json:"mac,omitempty"`
	Vendor string   `json:"vendor,omitempty"` // OUI vendor, only when known
	IPs    []string `json:"ips,omitempty"`
}

// IPAddress is a single address with its IP version and optional CIDR.
type IPAddress struct {
	IP      string `json:"ip"`
	Version int    `json:"version"` // 4 or 6
	CIDR    string `json:"cidr,omitempty"`
}

// Service is an observed open service on a device. State is always evidence-based
// (e.g. a successful TCP connect or an Nmap "open" state).
type Service struct {
	Port        int      `json:"port"`
	Protocol    string   `json:"protocol"` // tcp | udp
	State       string   `json:"state"`    // open
	Name        string   `json:"name,omitempty"`
	Product     string   `json:"product,omitempty"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"`
}

// TopologyEdge is a connection between two devices. The type and confidence make
// explicit how strongly the link is supported by evidence.
type TopologyEdge struct {
	ID              string   `json:"id"`
	Source          string   `json:"source"` // device ID
	Target          string   `json:"target"` // device ID
	Type            string   `json:"type"`
	Confidence      float64  `json:"confidence"`
	ConfidenceLabel string   `json:"confidence_label"`
	EvidenceIDs     []string `json:"evidence_ids,omitempty"`
	Reason          string   `json:"reason"`
}

// Evidence is a single observed fact. Devices, services and edges reference
// Evidence by ID so the whole report is auditable back to what was measured.
type Evidence struct {
	ID        string            `json:"id"`
	Kind      string            `json:"kind"`   // arp_table | reverse_dns | tcp_connect | icmp_echo | gateway_route | nmap | lldp | cdp | snmp_bridge | interface
	Source    string            `json:"source"` // tool/probe that produced it
	Summary   string            `json:"summary"`
	Data      map[string]string `json:"data,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// Warning is a scope/safety/quality note surfaced to the user.
type Warning struct {
	Code     string `json:"code"`
	Severity string `json:"severity"` // info | warning | danger
	Message  string `json:"message"`
}

// ScanSummary is the at-a-glance roll-up.
type ScanSummary struct {
	DeviceCount         int    `json:"device_count"`
	EdgeCount           int    `json:"edge_count"`
	EvidenceCount       int    `json:"evidence_count"`
	HighConfidenceEdges int    `json:"high_confidence_edges"`
	// InferredOnly is true when the topology rests entirely on inference (no
	// LLDP/CDP/SNMP evidence). It is the honest "this is not proven physical
	// topology" flag.
	InferredOnly bool   `json:"inferred_only"`
	Profile      string `json:"profile"`
	DurationMS   int64  `json:"duration_ms"`
}
