package output

import "github.com/thekiran/iad/internal/model"

type UIGraph struct {
	Nodes []UINode `json:"nodes"`
	Edges []UIEdge `json:"edges"`
}

type UINode struct {
	ID         string         `json:"id"`
	Label      string         `json:"label"`
	Type       string         `json:"type"`
	Icon       string         `json:"icon"`
	Confidence float64        `json:"confidence"`
	Inferred   bool           `json:"inferred"`
	Badges     []string       `json:"badges,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type UIEdge struct {
	ID         string  `json:"id"`
	Source     string  `json:"source"`
	Target     string  `json:"target"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
	Inferred   bool    `json:"inferred"`
	Style      string  `json:"style"`
}

func GraphAdapter(topology model.Topology) UIGraph {
	graph := UIGraph{
		Nodes: make([]UINode, 0, len(topology.Nodes)),
		Edges: make([]UIEdge, 0, len(topology.Edges)),
	}
	for _, node := range topology.Nodes {
		metadata := node.Metadata
		if metadata == nil {
			metadata = map[string]any{}
		}
		if node.DeviceType == model.DeviceTypeISPHop {
			metadata["section"] = "isp_path"
		} else {
			metadata["section"] = "lan"
		}
		graph.Nodes = append(graph.Nodes, UINode{
			ID:         node.ID,
			Label:      node.Label,
			Type:       string(node.DeviceType),
			Icon:       iconForType(node.DeviceType),
			Confidence: node.Confidence,
			Inferred:   node.Inferred,
			Badges:     badgesForNode(node),
			Metadata:   metadata,
		})
	}
	for _, edge := range topology.Edges {
		style := "solid"
		if edge.Inferred || edge.Confidence < 0.70 {
			style = "dashed"
		}
		graph.Edges = append(graph.Edges, UIEdge{
			ID:         edge.ID,
			Source:     edge.Source,
			Target:     edge.Target,
			Type:       string(edge.Relationship),
			Confidence: edge.Confidence,
			Inferred:   edge.Inferred,
			Style:      style,
		})
	}
	return graph
}

func iconForType(t model.DeviceType) string {
	switch t {
	case model.DeviceTypeRouter:
		return "router"
	case model.DeviceTypeModem:
		return "modem"
	case model.DeviceTypeONT:
		return "fiber"
	case model.DeviceTypeAccessPoint:
		return "wifi"
	case model.DeviceTypeSwitch, model.DeviceTypeManagedSwitch, model.DeviceTypeInferredSwitch:
		return "switch"
	case model.DeviceTypePrinter:
		return "printer"
	case model.DeviceTypeNAS:
		return "database"
	case model.DeviceTypeCamera:
		return "camera"
	case model.DeviceTypePhone:
		return "smartphone"
	case model.DeviceTypeLaptop:
		return "laptop"
	case model.DeviceTypeDesktop:
		return "monitor"
	case model.DeviceTypeISPHop:
		return "route"
	default:
		return "device"
	}
}

func badgesForNode(node model.TopologyNode) []string {
	var badges []string
	for _, role := range node.Roles {
		badges = append(badges, string(role))
	}
	if node.Inferred {
		badges = append(badges, "inferred")
	}
	if node.DeviceType == model.DeviceTypeISPHop {
		badges = append(badges, "observed route hop")
	}
	return badges
}
