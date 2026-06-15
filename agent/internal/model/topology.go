package model

type EdgeType string

const (
	EdgeTypeEthernetLink EdgeType = "ethernet_link"
	EdgeTypeWiFiLink     EdgeType = "wifi_link"
	EdgeTypeRoutedHop    EdgeType = "routed_hop"
	EdgeTypeInferredL2   EdgeType = "inferred_l2_link"
	EdgeTypeUpstreamNAT  EdgeType = "upstream_nat"
	EdgeTypeGatewayChain EdgeType = "gateway_chain"
	EdgeTypeISPRouteHop  EdgeType = "isp_route_hop"
	EdgeTypeUnknownLink  EdgeType = "unknown_link"
)

type Topology struct {
	Nodes []TopologyNode `json:"nodes"`
	Edges []TopologyEdge `json:"edges"`
}

type TopologyNode struct {
	ID         string         `json:"id"`
	Label      string         `json:"label,omitempty"`
	DeviceType DeviceType     `json:"device_type"`
	Roles      []DeviceRole   `json:"roles,omitempty"`
	Confidence float64        `json:"confidence"`
	Evidence   []Evidence     `json:"evidence,omitempty"`
	Inferred   bool           `json:"inferred"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type TopologyEdge struct {
	ID           string     `json:"id"`
	Source       string     `json:"source"`
	Target       string     `json:"target"`
	Relationship EdgeType   `json:"relationship"`
	Confidence   float64    `json:"confidence"`
	Evidence     []Evidence `json:"evidence,omitempty"`
	Inferred     bool       `json:"inferred"`
}
