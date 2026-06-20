package models

import "time"

// RichTopologySchema identifies the frontend-ready, evidence-rich topology view.
// It is additive to ScanReport so older consumers can keep reading /devices and
// /edges while newer UIs can use /rich_topology for AP/repeater/wireless context.
const RichTopologySchema = "iad.rich_topology/v1"

// Evidence source names supported by the rich topology model.
const (
	EvidenceSourceARP                   = "arp"
	EvidenceSourceNeighborTable         = "neighbor_table"
	EvidenceSourceICMP                  = "icmp"
	EvidenceSourceTCPProbe              = "tcp_probe"
	EvidenceSourceUDPProbe              = "udp_probe"
	EvidenceSourceMDNS                  = "mdns"
	EvidenceSourceSSDP                  = "ssdp"
	EvidenceSourceLLMNR                 = "llmnr"
	EvidenceSourceNBNS                  = "nbns"
	EvidenceSourceDHCP                  = "dhcp"
	EvidenceSourceDNS                   = "dns"
	EvidenceSourceTLS                   = "tls"
	EvidenceSourceHTTP                  = "http"
	EvidenceSourceUPnP                  = "upnp"
	EvidenceSourceSNMP                  = "snmp"
	EvidenceSourceRouterAPI             = "router_api"
	EvidenceSourceAPAPI                 = "ap_api"
	EvidenceSourcePassiveLAN            = "passive_lan"
	EvidenceSourcePassiveWiFi           = "passive_wifi"
	EvidenceSourceWirelessBeacon        = "wireless_beacon"
	EvidenceSourceWirelessProbeResponse = "wireless_probe_response"
	EvidenceSourceWirelessAssociation   = "wireless_association"
	EvidenceSourceManual                = "manual"
)

// Rich topology node categories and roles.
const (
	NodeCategoryNetwork  = "network"
	NodeCategoryDevice   = "device"
	NodeCategoryWireless = "wireless"
	NodeCategoryUnknown  = "unknown"

	NodeRoleGateway        = "gateway"
	NodeRoleRouter         = "router"
	NodeRoleAccessPoint    = "access_point"
	NodeRoleMeshNode       = "mesh_node"
	NodeRoleRepeater       = "repeater"
	NodeRoleSwitch         = "switch"
	NodeRoleWiredClient    = "wired_client"
	NodeRoleWirelessClient = "wireless_client"
	NodeRoleLocalAgent     = "local_agent"
	NodeRoleUnknown        = "unknown"
)

// Rich topology edge types. Values intentionally match the requested JSON names.
const (
	RichEdgeGatewayLink        = "gateway-link"
	RichEdgeSubnetInferred     = "subnet-inferred"
	RichEdgeARPNeighbor        = "arp-neighbor"
	RichEdgeWirelessAssociated = "wireless-associated"
	RichEdgeWirelessObserved   = "wireless-observed"
	RichEdgeMeshBackhaul       = "mesh-backhaul"
	RichEdgeRepeaterUplink     = "repeater-uplink"
	RichEdgeSwitchUplink       = "switch-uplink"
	RichEdgeReportedByRouter   = "reported-by-router"
	RichEdgeReportedByAP       = "reported-by-ap"
	RichEdgeWeakInferred       = "weak-inferred"
	RichEdgeUnknown            = "unknown"
)

// RichEvidence is a UI-facing evidence entry embedded directly on nodes/edges.
type RichEvidence struct {
	Source     string    `json:"source"`
	Value      string    `json:"value"`
	Confidence float64   `json:"confidence"`
	Timestamp  time.Time `json:"timestamp"`
	Interface  string    `json:"interface,omitempty"`
	Notes      string    `json:"notes,omitempty"`
}

type RichTopologyModel struct {
	SchemaVersion string             `json:"schema_version"`
	GeneratedAt   time.Time          `json:"generated_at"`
	Nodes         []RichTopologyNode `json:"nodes"`
	Edges         []RichTopologyEdge `json:"edges"`
	Warnings      []Warning          `json:"warnings,omitempty"`
	Capabilities  []ReportCapability `json:"capabilities,omitempty"`
	UI            RichTopologyUI     `json:"ui"`
}

type RichTopologyNode struct {
	ID                string             `json:"id"`
	Label             string             `json:"label"`
	Type              string             `json:"type"`
	Category          string             `json:"category"`
	DeviceRole        string             `json:"device_role"`
	IPAddresses       []string           `json:"ip_addresses,omitempty"`
	MACAddresses      []string           `json:"mac_addresses,omitempty"`
	Vendor            string             `json:"vendor,omitempty"`
	Hostname          string             `json:"hostname,omitempty"`
	OSHint            string             `json:"os_hint,omitempty"`
	MobileFingerprint *MobileFingerprint `json:"mobileFingerprint,omitempty"`
	DeviceTypeHint    string             `json:"deviceTypeHint,omitempty"`
	MobileOSHint      string             `json:"osHint,omitempty"`
	OSConfidence      float64            `json:"osConfidence,omitempty"`
	Services          []RichService      `json:"services,omitempty"`
	Interfaces        []RichInterface    `json:"interfaces,omitempty"`
	Wireless          *RichWireless      `json:"wireless,omitempty"`
	FirstSeen         time.Time          `json:"first_seen,omitempty"`
	LastSeen          time.Time          `json:"last_seen,omitempty"`
	Confidence        float64            `json:"confidence"`
	RiskFlags         []string           `json:"risk_flags,omitempty"`
	Evidence          []RichEvidence     `json:"evidence,omitempty"`
	RawSources        []string           `json:"raw_sources,omitempty"`
	UI                map[string]any     `json:"ui,omitempty"`
}

type RichService struct {
	Port        int      `json:"port,omitempty"`
	Protocol    string   `json:"protocol,omitempty"`
	State       string   `json:"state,omitempty"`
	Name        string   `json:"name,omitempty"`
	Product     string   `json:"product,omitempty"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"`
}

type RichInterface struct {
	Name   string   `json:"name,omitempty"`
	MAC    string   `json:"mac,omitempty"`
	Vendor string   `json:"vendor,omitempty"`
	IPs    []string `json:"ips,omitempty"`
}

type RichWireless struct {
	SSID             string   `json:"ssid,omitempty"`
	BSSID            string   `json:"bssid,omitempty"`
	AssociatedBSSID  string   `json:"associated_bssid,omitempty"`
	Channel          int      `json:"channel,omitempty"`
	Frequency        int      `json:"frequency,omitempty"`
	Band             string   `json:"band,omitempty"`
	RSSI             int      `json:"rssi,omitempty"`
	Noise            int      `json:"noise,omitempty"`
	Security         string   `json:"security,omitempty"`
	PHY              string   `json:"phy,omitempty"`
	Capabilities     []string `json:"capabilities,omitempty"`
	IsAP             bool     `json:"is_ap"`
	IsStation        bool     `json:"is_station"`
	IsMesh           bool     `json:"is_mesh"`
	IsRepeaterHint   bool     `json:"is_repeater_hint"`
	ObservationCount int      `json:"observation_count"`
	Confidence       float64  `json:"confidence"`
}

type RichTopologyEdge struct {
	ID         string         `json:"id"`
	Source     string         `json:"source"`
	Target     string         `json:"target"`
	Type       string         `json:"type"`
	Relation   string         `json:"relation"`
	Medium     string         `json:"medium"`
	Confidence float64        `json:"confidence"`
	Evidence   []RichEvidence `json:"evidence,omitempty"`
	FirstSeen  time.Time      `json:"first_seen,omitempty"`
	LastSeen   time.Time      `json:"last_seen,omitempty"`
	UI         map[string]any `json:"ui,omitempty"`
}

type RichTopologyUI struct {
	RootNodeID        string   `json:"root_node_id,omitempty"`
	InferredOnly      bool     `json:"inferred_only"`
	PhysicalEdgeCount int      `json:"physical_edge_count"`
	Warnings          []string `json:"warnings,omitempty"`
}
