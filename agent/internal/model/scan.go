package model

import "time"

type ScanMode string

const (
	ScanModeSafe   ScanMode = "safe"
	ScanModeNormal ScanMode = "normal"
	ScanModeDeep   ScanMode = "deep"
)

type ScanScope struct {
	LocalSubnets   []string `json:"local_subnets"`
	ScannedRanges  []string `json:"scanned_ranges"`
	ExcludedRanges []string `json:"excluded_ranges"`
	PublicScanning bool     `json:"public_scanning"`
}

type NetworkContext struct {
	DefaultGateway string   `json:"default_gateway"`
	DHCPServer     string   `json:"dhcp_server"`
	DNSServers     []string `json:"dns_servers"`
	PublicIP       string   `json:"public_ip"`
	ISP            string   `json:"isp"`
	ASN            string   `json:"asn"`
	CGNAT          *bool    `json:"cgnat"`
	IPv6Available  bool     `json:"ipv6_available"`
}

type RouteHop struct {
	Index      int        `json:"index"`
	IP         string     `json:"ip"`
	Hostname   string     `json:"hostname,omitempty"`
	ASN        string     `json:"asn,omitempty"`
	Org        string     `json:"org,omitempty"`
	LatencyMS  float64    `json:"latency_ms,omitempty"`
	Private    bool       `json:"private"`
	Evidence   []Evidence `json:"evidence,omitempty"`
	Confidence float64    `json:"confidence"`
}

type ISPPath struct {
	PublicIP       string     `json:"public_ip"`
	ASN            string     `json:"asn"`
	Organization   string     `json:"organization"`
	CGNAT          bool       `json:"cgnat"`
	FirstPublicHop string     `json:"first_public_hop"`
	PrivateHops    []RouteHop `json:"private_hops"`
	PublicHops     []RouteHop `json:"public_hops"`
	Confidence     float64    `json:"confidence"`
	Warning        string     `json:"warning"`
}

type ConflictSeverity string

const (
	ConflictLow    ConflictSeverity = "low"
	ConflictMedium ConflictSeverity = "medium"
	ConflictHigh   ConflictSeverity = "high"
)

type Conflict struct {
	Type        string           `json:"type"`
	Severity    ConflictSeverity `json:"severity"`
	Devices     []string         `json:"devices,omitempty"`
	Description string           `json:"description"`
	Effect      string           `json:"effect"`
	Resolution  string           `json:"resolution"`
	Evidence    []Evidence       `json:"evidence,omitempty"`
}

type DataQuality struct {
	HasConflicts bool       `json:"has_conflicts"`
	Conflicts    []Conflict `json:"conflicts"`
}

type Summary struct {
	DeviceCount      int `json:"device_count"`
	ConfirmedDevices int `json:"confirmed_devices"`
	InferredDevices  int `json:"inferred_devices"`
	Routers          int `json:"routers"`
	Switches         int `json:"switches"`
	AccessPoints     int `json:"access_points"`
	UnknownDevices   int `json:"unknown_devices"`
}

type NextBestProbe struct {
	ProbeName        string   `json:"probe_name"`
	Reason           string   `json:"reason"`
	ExpectedEvidence []string `json:"expected_evidence"`
	Safety           string   `json:"safety"`
}

type UIOutput struct {
	Headline string   `json:"headline"`
	Summary  string   `json:"summary"`
	Warnings []string `json:"warnings"`
	Badges   []string `json:"badges"`
}

type ScanOutput struct {
	ScanID         string          `json:"scan_id"`
	CreatedAt      time.Time       `json:"created_at"`
	Mode           ScanMode        `json:"mode"`
	Scope          ScanScope       `json:"scope"`
	LocalHost      Device          `json:"local_host"`
	NetworkContext NetworkContext  `json:"network_context"`
	Devices        []Device        `json:"devices"`
	Topology       Topology        `json:"topology"`
	ISPPath        ISPPath         `json:"isp_path"`
	Evidence       []Evidence      `json:"evidence"`
	DataQuality    DataQuality     `json:"data_quality"`
	Summary        Summary         `json:"summary"`
	NextBestProbes []NextBestProbe `json:"next_best_probes"`
	UI             UIOutput        `json:"ui"`
}

func Clamp01(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}
