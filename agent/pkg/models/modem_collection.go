package models

// TriState is used where "not observed" must stay distinct from a verified
// negative result. Failed probes and timeouts should normally become unknown.
type TriState string

const (
	TriTrue    TriState = "true"
	TriFalse   TriState = "false"
	TriUnknown TriState = "unknown"
)

type ModemCollection struct {
	Status                 string                    `json:"status"`
	SafeMode               bool                      `json:"safe_mode"`
	Scope                  ModemCollectionScope      `json:"scope"`
	CPECandidates          []CPECandidate            `json:"cpe_candidates"`
	NormalizedGatewayChain NormalizedGatewayChain    `json:"normalized_gateway_chain"`
	NAT                    ModemNATState             `json:"nat"`
	AccessClassification   ModemAccessClassification `json:"access_classification"`
	DataQuality            DataQuality               `json:"data_quality"`
	UI                     UIOutput                  `json:"ui,omitempty"`
	SecurityNotes          []string                  `json:"security_notes,omitempty"`
	Undetermined           []string                  `json:"undetermined_without_credentials_or_physical_evidence,omitempty"`
}

type ModemNATState struct {
	CGNAT                      TriState `json:"cgnat"`
	DoubleNAT                  TriState `json:"double_nat"`
	InternalDoubleNATPossible  TriState `json:"internal_double_nat_possible"`
	PublicIPMatches            TriState `json:"public_ip_matches"`
	ExternalPublicIPConsistent TriState `json:"external_public_ip_consistent"`
	PCPReachable               TriState `json:"pcp_reachable"`
	NATPMPReachable            TriState `json:"nat_pmp_reachable"`
}

type ModemCollectionScope struct {
	PrivateOnly bool     `json:"private_only"`
	Targets     []string `json:"targets"`
}

type CPECandidate struct {
	IP                  string                  `json:"ip"`
	Role                string                  `json:"role"`
	Source              []string                `json:"source,omitempty"`
	Priority            string                  `json:"priority"`
	Private             bool                    `json:"private"`
	ReachableState      TriState                `json:"reachable_state"`
	OpenPorts           []int                   `json:"open_ports,omitempty"`
	HTTP                CPEHTTPState            `json:"http"`
	TLS                 CPETLSState             `json:"tls"`
	UPnP                CPEUPnPState            `json:"upnp"`
	TR064               CPETR064State           `json:"tr064"`
	TR181               CPETR181State           `json:"tr181"`
	SNMP                CPESNMPState            `json:"snmp"`
	ModelFingerprint    CPEModelFingerprint     `json:"model_fingerprint"`
	WANPhysicalEvidence CPEWANPhysicalEvidence  `json:"wan_physical_evidence"`
	FailedAttempts      []ProbeAttempt          `json:"failed_attempts,omitempty"`
	Conflicts           []DataConflict          `json:"conflicts,omitempty"`
	Confidence          float64                 `json:"confidence"`
	EvidenceIDs         []string                `json:"evidence_ids,omitempty"`
	HTTPObservations    []HTTPObservation       `json:"-"`
	TLSObservations     []TLSObservation        `json:"-"`
	AccessEvidence      []GatewayAccessEvidence `json:"-"`
}

type CPEHTTPState struct {
	Reachable    TriState          `json:"reachable"`
	Observations []HTTPObservation `json:"observations,omitempty"`
}

type CPETLSState struct {
	Reachable    TriState         `json:"reachable"`
	Certificates []TLSObservation `json:"certificates,omitempty"`
}

type CPEUPnPState struct {
	Found                      TriState `json:"found"`
	IGDFound                   TriState `json:"igd_found"`
	WANCommonInterfaceFound    TriState `json:"wan_common_interface_found"`
	WANAccessType              *string  `json:"wan_access_type"`
	Layer1UpstreamMaxBitRate   *int64   `json:"layer1_upstream_max_bitrate"`
	Layer1DownstreamMaxBitRate *int64   `json:"layer1_downstream_max_bitrate"`
	PhysicalLinkStatus         *string  `json:"physical_link_status"`
}

type CPETR064State struct {
	Found          TriState `json:"found"`
	AuthRequired   TriState `json:"auth_required"`
	DataAccessible TriState `json:"data_accessible"`
	Services       []string `json:"services,omitempty"`
}

type CPETR181State struct {
	Available      TriState `json:"available"`
	InterfaceStack []string `json:"interface_stack,omitempty"`
	PhysicalLayers []string `json:"physical_layers,omitempty"`
}

type CPESNMPState struct {
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
}

type CPEModelFingerprint struct {
	Vendor     *string        `json:"vendor"`
	Model      *string        `json:"model"`
	Confidence float64        `json:"confidence"`
	Evidence   []EvidenceItem `json:"evidence,omitempty"`
}

type CPEWANPhysicalEvidence struct {
	Status     string   `json:"status"`
	Type       *string  `json:"type"`
	Subtype    *string  `json:"subtype"`
	Confidence float64  `json:"confidence"`
	Source     *string  `json:"source"`
	Evidence   []string `json:"evidence,omitempty"`
}

type NormalizedGatewayChain struct {
	Hops                      []GatewayHop `json:"hops,omitempty"`
	InternalDoubleNATPossible TriState     `json:"internal_double_nat_possible"`
	EvidenceSources           []string     `json:"evidence_sources,omitempty"`
}

type ModemAccessClassification struct {
	PrimaryType             string            `json:"primary_type"`
	Subtype                 *string           `json:"subtype"`
	Confidence              float64           `json:"confidence"`
	DecisionQuality         string            `json:"decision_quality"`
	SafeToDisplayAsFinal    bool              `json:"safe_to_display_as_final"`
	Candidates              []AccessCandidate `json:"candidates,omitempty"`
	MissingRequiredEvidence []string          `json:"missing_required_evidence,omitempty"`
	Reason                  []string          `json:"reason,omitempty"`
}
