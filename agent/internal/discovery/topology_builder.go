package discovery

import (
	"fmt"
	"strings"
	"time"

	"github.com/thekiran/iad/internal/model"
	"github.com/thekiran/iad/internal/safety"
)

func BuildTopology(devices []model.Device, evidence []model.Evidence) model.Topology {
	nodes := make([]model.TopologyNode, 0, len(devices)+1)
	for _, d := range devices {
		d = ClassifyDevice(d, d.Evidence)
		nodes = append(nodes, nodeFromDevice(d))
	}

	var edges []model.TopologyEdge
	addEdge := func(source, target string, rel model.EdgeType, confidence float64, ev []model.Evidence, inferred bool) {
		if source == "" || target == "" || source == target {
			return
		}
		if !nodeExists(nodes, source) || !nodeExists(nodes, target) {
			return
		}
		edge := model.TopologyEdge{
			ID:           edgeID(source, target, rel),
			Source:       source,
			Target:       target,
			Relationship: rel,
			Confidence:   topologyEdgeConfidence(rel, confidence, inferred),
			Evidence:     ev,
			Inferred:     inferred,
		}
		edges = upsertEdge(edges, edge)
	}

	localID := findDeviceID(devices, func(d model.Device) bool {
		return d.DeviceType == model.DeviceTypeLocalHost || d.HasRole(model.RoleLocalHost)
	})
	gatewayID := findDeviceID(devices, func(d model.Device) bool {
		return d.HasRole(model.RoleDefaultGateway)
	})
	apID := findDeviceID(devices, func(d model.Device) bool {
		return d.DeviceType == model.DeviceTypeAccessPoint || d.HasRole(model.RoleWiFiAP)
	})

	activeMedium := activeInterfaceMedium(evidence)
	if localID != "" && gatewayID != "" {
		target := gatewayID
		rel := model.EdgeTypeUnknownLink
		conf := 0.40
		inferred := true
		if activeMedium == "wifi" {
			rel = model.EdgeTypeWiFiLink
			conf = 0.85
			inferred = false
			if apID != "" {
				target = apID
			}
		} else if activeMedium == "ethernet" {
			rel = model.EdgeTypeEthernetLink
			conf = 0.70
		}
		addEdge(localID, target, rel, conf, matchingEvidence(evidence, "os_interface_probe", "route_table_probe", "wifi_probe"), inferred)
		if target != gatewayID && target != "" {
			addEdge(target, gatewayID, model.EdgeTypeInferredL2, 0.55, matchingEvidence(evidence, "wifi_probe", "route_table_probe"), true)
		}
	}

	// Explicit neighbor evidence wins over inference.
	for _, ev := range evidence {
		source, _ := ev.Raw["source"].(string)
		target, _ := ev.Raw["target"].(string)
		if source == "" {
			source, _ = ev.Raw["edge_source"].(string)
		}
		if target == "" {
			target, _ = ev.Raw["edge_target"].(string)
		}
		if source == "" || target == "" {
			continue
		}
		rel := edgeTypeFromRaw(ev)
		conf := ev.Confidence
		inferred := ev.Strength != model.EvidenceStrong
		if strings.Contains(strings.ToLower(ev.Source), "lldp") || strings.Contains(strings.ToLower(ev.Source), "cdp") {
			conf = max(conf, 0.90)
			inferred = false
		}
		if strings.Contains(strings.ToLower(ev.Source), "snmp") && strings.Contains(strings.ToLower(ev.Reason), "bridge") {
			conf = max(conf, 0.80)
			inferred = false
		}
		addEdge(source, target, rel, conf, []model.Evidence{ev}, inferred)
	}

	privateRoute := routeDevices(devices, true)
	publicRoute := routeDevices(devices, false)
	for i := 0; i+1 < len(privateRoute); i++ {
		addEdge(privateRoute[i].ID, privateRoute[i+1].ID, model.EdgeTypeGatewayChain, 0.60, privateRoute[i].Evidence, false)
	}
	if gatewayID != "" && len(privateRoute) > 0 {
		addEdge(gatewayID, privateRoute[0].ID, model.EdgeTypeUpstreamNAT, 0.55, privateRoute[0].Evidence, false)
	}
	prev := ""
	if len(privateRoute) > 0 {
		prev = privateRoute[len(privateRoute)-1].ID
	} else {
		prev = gatewayID
	}
	for _, hop := range publicRoute {
		if prev != "" {
			addEdge(prev, hop.ID, model.EdgeTypeISPRouteHop, 0.50, hop.Evidence, false)
		}
		prev = hop.ID
	}

	if shouldInferSwitch(devices, activeMedium) {
		switchID := "inferred_switch_lan_1"
		ev := model.NewEvidence("topology_builder", switchID, model.EvidenceWeak, 0.50, "Multiple L2 peers are present and no managed switch was directly observed.", nil, latestEvidenceTime(evidence))
		nodes = append(nodes, model.TopologyNode{
			ID:         switchID,
			Label:      "Inferred switch",
			DeviceType: model.DeviceTypeInferredSwitch,
			Roles:      []model.DeviceRole{model.RoleSwitchingDevice},
			Confidence: 0.50,
			Evidence:   []model.Evidence{ev},
			Inferred:   true,
		})
		if localID != "" {
			addEdge(localID, switchID, model.EdgeTypeInferredL2, 0.50, []model.Evidence{ev}, true)
		}
		for _, d := range devices {
			if d.ID == localID || d.ID == gatewayID || d.DeviceType == model.DeviceTypeISPHop {
				continue
			}
			if hasPrivateIP(d) {
				addEdge(switchID, d.ID, model.EdgeTypeInferredL2, 0.45, []model.Evidence{ev}, true)
			}
		}
	}

	return model.Topology{Nodes: nodes, Edges: edges}
}

func nodeFromDevice(d model.Device) model.TopologyNode {
	label := d.ID
	if len(d.Hostnames) > 0 {
		label = d.Hostnames[0]
	} else if len(d.IPAddresses) > 0 {
		label = d.IPAddresses[0]
	} else if d.Model != "" {
		label = d.Model
	}
	return model.TopologyNode{
		ID:         d.ID,
		Label:      label,
		DeviceType: d.DeviceType,
		Roles:      d.Roles,
		Confidence: d.Confidence,
		Evidence:   d.Evidence,
		Inferred:   d.Inferred,
	}
}

func edgeID(source, target string, rel model.EdgeType) string {
	return fmt.Sprintf("edge_%s_%s_%s", source, target, rel)
}

func upsertEdge(edges []model.TopologyEdge, edge model.TopologyEdge) []model.TopologyEdge {
	for i := range edges {
		if edges[i].ID == edge.ID {
			if edge.Confidence > edges[i].Confidence {
				edges[i] = edge
			}
			return edges
		}
	}
	return append(edges, edge)
}

func findDeviceID(devices []model.Device, pred func(model.Device) bool) string {
	for _, d := range devices {
		if pred(d) {
			return d.ID
		}
	}
	return ""
}

func nodeExists(nodes []model.TopologyNode, id string) bool {
	for _, n := range nodes {
		if n.ID == id {
			return true
		}
	}
	return false
}

func activeInterfaceMedium(evidence []model.Evidence) string {
	for _, ev := range evidence {
		for _, key := range []string{"active_interface", "adapter_type", "medium"} {
			if v, _ := ev.Raw[key].(string); v != "" {
				lower := strings.ToLower(v)
				if strings.Contains(lower, "wifi") || strings.Contains(lower, "wi-fi") || strings.Contains(lower, "wireless") {
					return "wifi"
				}
				if strings.Contains(lower, "ethernet") {
					return "ethernet"
				}
			}
		}
	}
	return ""
}

func matchingEvidence(evidence []model.Evidence, sources ...string) []model.Evidence {
	var out []model.Evidence
	for _, ev := range evidence {
		for _, source := range sources {
			if strings.Contains(ev.Source, source) {
				out = append(out, ev)
				break
			}
		}
	}
	return out
}

func edgeTypeFromRaw(ev model.Evidence) model.EdgeType {
	if rawType, _ := ev.Raw["edge_type"].(string); rawType != "" {
		return model.EdgeType(rawType)
	}
	source := strings.ToLower(ev.Source + " " + ev.Reason)
	switch {
	case strings.Contains(source, "wifi"):
		return model.EdgeTypeWiFiLink
	case strings.Contains(source, "lldp"), strings.Contains(source, "cdp"), strings.Contains(source, "snmp"):
		return model.EdgeTypeEthernetLink
	case strings.Contains(source, "traceroute"), strings.Contains(source, "route"):
		return model.EdgeTypeRoutedHop
	default:
		return model.EdgeTypeUnknownLink
	}
}

func routeDevices(devices []model.Device, private bool) []model.Device {
	var out []model.Device
	for _, d := range devices {
		if d.DeviceType == model.DeviceTypeISPHop {
			if !private {
				out = append(out, d)
			}
			continue
		}
		if private && d.HasRole(model.RoleUpstreamGateway) {
			out = append(out, d)
		}
	}
	return out
}

func shouldInferSwitch(devices []model.Device, activeMedium string) bool {
	if activeMedium == "wifi" {
		return false
	}
	peers := 0
	for _, d := range devices {
		if d.DeviceType == model.DeviceTypeManagedSwitch || d.DeviceType == model.DeviceTypeSwitch {
			return false
		}
		if d.DeviceType == model.DeviceTypeISPHop || d.DeviceType == model.DeviceTypeLocalHost || d.HasRole(model.RoleDefaultGateway) {
			continue
		}
		if hasPrivateIP(d) {
			peers++
		}
	}
	return peers >= 2
}

func hasPrivateIP(d model.Device) bool {
	for _, ip := range d.IPAddresses {
		if safety.IsPrivateIPString(ip) {
			return true
		}
	}
	return false
}

func latestEvidenceTime(evidence []model.Evidence) time.Time {
	var latest time.Time
	for _, ev := range evidence {
		if ev.Timestamp.After(latest) {
			latest = ev.Timestamp
		}
	}
	return latest
}
