package topology

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

// BuildRichTopology folds all currently available evidence-backed report
// sections into a frontend-ready topology graph. It is intentionally conservative:
// inferred route/subnet/wireless correlations stay inferred, and only
// LLDP/CDP/SNMP/router/AP table evidence can produce high-confidence adjacency.
func BuildRichTopology(report models.ScanReport) models.RichTopologyModel {
	evidenceByID := indexEvidence(report)
	nodes := map[string]*models.RichTopologyNode{}
	edges := map[string]*models.RichTopologyEdge{}

	for _, d := range report.Devices {
		upsertRootDevice(nodes, d, evidenceByID)
	}
	if report.DeviceIntel != nil {
		for _, d := range report.DeviceIntel.Devices {
			upsertIntelDevice(nodes, d, evidenceByID)
		}
	}

	for _, e := range report.Edges {
		upsertRichEdge(edges, edgeFromRoot(e, evidenceByID))
	}
	if report.DeviceIntel != nil {
		for _, e := range report.DeviceIntel.Edges {
			upsertRichEdge(edges, edgeFromIntel(e, evidenceByID))
		}
	}

	outNodes := make([]models.RichTopologyNode, 0, len(nodes))
	for _, n := range nodes {
		sortNode(n)
		outNodes = append(outNodes, *n)
	}
	sort.Slice(outNodes, func(i, j int) bool {
		if outNodes[i].DeviceRole != outNodes[j].DeviceRole {
			return roleRank(outNodes[i].DeviceRole) < roleRank(outNodes[j].DeviceRole)
		}
		return outNodes[i].ID < outNodes[j].ID
	})

	outEdges := make([]models.RichTopologyEdge, 0, len(edges))
	inferredOnly := true
	physicalCount := 0
	for _, e := range edges {
		sortEdge(e)
		if e.Type == models.RichEdgeReportedByRouter || e.Type == models.RichEdgeReportedByAP || e.Type == models.RichEdgeWirelessAssociated {
			inferredOnly = false
		}
		if medium := strings.ToLower(e.Medium); medium == "l2" || medium == "wireless" {
			if e.Confidence >= 0.80 && (e.Type == models.RichEdgeReportedByAP || e.Type == models.RichEdgeReportedByRouter) {
				physicalCount++
			}
		}
		outEdges = append(outEdges, *e)
	}
	sort.Slice(outEdges, func(i, j int) bool {
		if outEdges[i].Confidence != outEdges[j].Confidence {
			return outEdges[i].Confidence > outEdges[j].Confidence
		}
		return outEdges[i].ID < outEdges[j].ID
	})

	warnings := richWarnings(report, outEdges)
	return models.RichTopologyModel{
		SchemaVersion: models.RichTopologySchema,
		GeneratedAt:   generatedAt(report),
		Nodes:         outNodes,
		Edges:         outEdges,
		Warnings:      warnings,
		Capabilities:  richCapabilities(report),
		UI: models.RichTopologyUI{
			RootNodeID:        rootNodeID(outNodes),
			InferredOnly:      inferredOnly,
			PhysicalEdgeCount: physicalCount,
			Warnings:          warningMessages(warnings),
		},
	}
}

// BuildTopologyV2 exposes the same evidence graph at /topology using the
// iad.topology/v2 contract requested by the desktop map. It is deliberately
// derived from BuildRichTopology so legacy /devices, /edges, and /rich_topology
// cannot drift from the primary frontend graph.
func BuildTopologyV2(report models.ScanReport, rich models.RichTopologyModel) models.TopologyV2Model {
	edges := make([]models.TopologyV2Edge, 0, len(rich.Edges))
	for _, edge := range rich.Edges {
		edges = append(edges, topologyV2Edge(edge))
	}
	warnings := topologyV2Warnings(rich.Warnings)
	return models.TopologyV2Model{
		SchemaVersion: models.TopologyReportSchema,
		GeneratedAt:   rich.GeneratedAt,
		RootID:        rich.UI.RootNodeID,
		Nodes:         append([]models.RichTopologyNode{}, rich.Nodes...),
		Edges:         edges,
		Warnings:      warnings,
		Capabilities:  rich.Capabilities,
		UI:            rich.UI,
	}
}

func topologyV2Edge(edge models.RichTopologyEdge) models.TopologyV2Edge {
	explanation := edgeExplanation(edge)
	warnings := edgeWarnings(edge)
	return models.TopologyV2Edge{
		ID:          edge.ID,
		Source:      edge.Source,
		Target:      edge.Target,
		Type:        edge.Type,
		Relation:    edge.Relation,
		Medium:      edge.Medium,
		Confidence:  edge.Confidence,
		Evidence:    append([]models.RichEvidence{}, edge.Evidence...),
		Explanation: explanation,
		Warnings:    warnings,
		FirstSeen:   edge.FirstSeen,
		LastSeen:    edge.LastSeen,
		UI:          mergeUI(copyUI(edge.UI), map[string]any{"warnings": warnings, "explanation": explanation}),
	}
}

func edgeExplanation(edge models.RichTopologyEdge) string {
	if reason, _ := edge.UI["reason"].(string); strings.TrimSpace(reason) != "" {
		return reason
	}
	switch edge.Type {
	case models.RichEdgeGatewayLink:
		return "Default route evidence links the local agent to the known gateway."
	case models.RichEdgeReportedByRouter:
		return "A router-reported client or route table supplied this relationship."
	case models.RichEdgeReportedByAP:
		return "An authorized AP client table supplied this relationship."
	case models.RichEdgeSwitchUplink:
		return "Switch/router evidence such as LLDP, CDP, SNMP, or bridge data supports this uplink."
	case models.RichEdgeWirelessAssociated:
		return "Wireless association evidence links the station to the BSSID/AP."
	case models.RichEdgeWirelessObserved:
		return "Wireless metadata was observed, but the AP/station relationship is not fully proven."
	case models.RichEdgeMeshBackhaul:
		return "Mesh/backhaul metadata suggests a mesh relationship."
	case models.RichEdgeRepeaterUplink:
		return "Repeater/extender metadata suggests an uplink relationship."
	case models.RichEdgeSubnetInferred:
		return "Both devices are known in the same local subnet, so the relationship is inferred through the LAN."
	case models.RichEdgeARPNeighbor:
		return "ARP/neighbor evidence shows both endpoints on the local broadcast domain."
	case models.RichEdgeWeakInferred:
		return "Only weak route, timing, or correlation evidence is available; this is not a proven physical link."
	default:
		return "The relationship is present in the source topology, but evidence is insufficient to classify it strongly."
	}
}

func edgeWarnings(edge models.RichTopologyEdge) []string {
	switch edge.Type {
	case models.RichEdgeSubnetInferred:
		return []string{"Inferred from subnet membership; not a proven physical connection."}
	case models.RichEdgeWeakInferred:
		return []string{"Weak inferred relationship; do not treat as a proven AP, switch, or cable link."}
	case models.RichEdgeWirelessObserved:
		return []string{"Wireless metadata observed without association/client-table proof."}
	case models.RichEdgeARPNeighbor:
		return []string{"ARP proves local neighbor visibility, not physical switch-port adjacency."}
	}
	if physical, _ := edge.UI["physical"].(bool); !physical && edge.Medium != "" {
		return []string{"Physical path is not proven by this evidence."}
	}
	return nil
}

func topologyV2Warnings(existing []models.Warning) []models.Warning {
	out := append([]models.Warning{}, existing...)
	out = append(out,
		models.Warning{
			Code:     "topology_physical_links_not_assumed",
			Severity: models.SeverityInfo,
			Message:  "Topology inferred olabilir, fiziksel linkler her zaman kanıtlı değildir.",
		},
		models.Warning{
			Code:     "passive_wifi_unsupported_by_default",
			Severity: models.SeverityInfo,
			Message:  "Passive Wi-Fi metadata collection is optional and remains disabled/unsupported unless the OS, adapter, driver, and permissions support monitor metadata.",
		},
	)
	return dedupeWarnings(out)
}

func BuildWirelessMetadata(rich models.RichTopologyModel) []models.WirelessMetadata {
	var out []models.WirelessMetadata
	for _, node := range rich.Nodes {
		if node.Wireless == nil {
			continue
		}
		out = append(out, models.WirelessMetadata{
			SSID:             node.Wireless.SSID,
			BSSID:            node.Wireless.BSSID,
			Channel:          node.Wireless.Channel,
			Frequency:        node.Wireless.Frequency,
			Band:             node.Wireless.Band,
			RSSI:             node.Wireless.RSSI,
			Security:         node.Wireless.Security,
			APMAC:            node.Wireless.BSSID,
			StationMAC:       firstString(node.MACAddresses),
			FirstSeen:        node.FirstSeen,
			LastSeen:         node.LastSeen,
			ObservationCount: node.Wireless.ObservationCount,
			Confidence:       node.Wireless.Confidence,
			Source:           "topology_inference",
		})
	}
	return out
}

func BuildRawObservations(report models.ScanReport) []models.RawObservation {
	out := make([]models.RawObservation, 0, len(report.Evidence))
	for _, ev := range report.Evidence {
		obs := models.RawObservation{
			Source:         richEvidenceSource(ev),
			Kind:           ev.Kind,
			Interface:      ev.Data["interface"],
			SourceMAC:      ev.Data["source_mac"],
			DestinationMAC: ev.Data["destination_mac"],
			SourceIP:       ev.Data["source_ip"],
			DestinationIP:  ev.Data["destination_ip"],
			Protocol:       ev.Data["protocol"],
			EtherType:      ev.Data["ether_type"],
			FirstSeen:      ev.Timestamp,
			LastSeen:       ev.Timestamp,
			Count:          1,
			Metadata:       sanitizedObservationMetadata(ev.Data),
		}
		if obs.SourceIP == "" {
			obs.SourceIP = ev.Data["ip"]
		}
		if obs.SourceMAC == "" {
			obs.SourceMAC = ev.Data["mac"]
		}
		out = append(out, obs)
	}
	return out
}

func sanitizedObservationMetadata(data map[string]string) map[string]any {
	if len(data) == 0 {
		return nil
	}
	blocked := map[string]bool{
		"payload": true, "body": true, "cookie": true, "cookies": true,
		"token": true, "password": true, "secret": true, "credential": true,
	}
	out := map[string]any{}
	for k, v := range data {
		if blocked[strings.ToLower(k)] {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func upsertRootDevice(nodes map[string]*models.RichTopologyNode, d models.Device, evidenceByID map[string]models.Evidence) {
	n := ensureNode(nodes, d.ID)
	n.Label = firstNonEmpty(d.Hostname, firstIP(d.Addresses), d.ID)
	n.Type = nodeTypeFromRoot(d)
	n.Category = models.NodeCategoryDevice
	n.DeviceRole = roleFromRoot(d)
	n.IPAddresses = appendUniqueStrings(n.IPAddresses, deviceIPs(d)...)
	n.MACAddresses = appendUniqueStrings(n.MACAddresses, d.MAC)
	for _, iface := range d.Interfaces {
		n.MACAddresses = appendUniqueStrings(n.MACAddresses, iface.MAC)
		n.Interfaces = append(n.Interfaces, models.RichInterface{MAC: iface.MAC, Vendor: iface.Vendor, IPs: append([]string{}, iface.IPs...)})
	}
	n.Vendor = firstNonEmpty(n.Vendor, d.Vendor, d.OUIVendor)
	n.Hostname = firstNonEmpty(n.Hostname, d.Hostname)
	if d.MobileFingerprint != nil {
		fp := *d.MobileFingerprint
		n.MobileFingerprint = &fp
	}
	n.DeviceTypeHint = firstNonEmpty(n.DeviceTypeHint, d.DeviceTypeHint)
	n.MobileOSHint = firstNonEmpty(n.MobileOSHint, d.OSHint)
	n.OSConfidence = maxFloat(n.OSConfidence, d.OSConfidence)
	n.Services = append(n.Services, rootServices(d.Services)...)
	n.FirstSeen = earliest(n.FirstSeen, d.FirstSeen)
	n.LastSeen = latest(n.LastSeen, d.LastSeen)
	n.Confidence = maxFloat(n.Confidence, d.Confidence)
	n.RawSources = appendUniqueStrings(n.RawSources, d.DiscoverySources...)
	n.Evidence = mergeRichEvidence(n.Evidence, richEvidenceForIDs(d.EvidenceIDs, evidenceByID)...)
	n.UI = mergeUI(n.UI, map[string]any{"is_gateway": d.IsGateway, "is_agent": d.IsAgent, "mobile_label": mobileNodeLabel(d.MobileFingerprint)})
}

func upsertIntelDevice(nodes map[string]*models.RichTopologyNode, d models.DeviceIntelDevice, evidenceByID map[string]models.Evidence) {
	n := ensureNode(nodes, d.ID)
	n.Label = firstNonEmpty(firstString(d.Hostnames), firstString(d.IPAddresses), d.ID)
	n.Type = firstNonEmpty(d.DeviceType.Primary, n.Type, models.DeviceTypeUnknown)
	n.Category = categoryForIntel(d)
	n.DeviceRole = roleFromIntel(d)
	n.IPAddresses = appendUniqueStrings(n.IPAddresses, d.IPAddresses...)
	n.MACAddresses = appendUniqueStrings(n.MACAddresses, d.MACAddresses...)
	n.Hostname = firstNonEmpty(n.Hostname, firstString(d.Hostnames))
	n.Vendor = firstNonEmpty(n.Vendor, ptrValue(d.Vendor.FingerprintVendor), ptrValue(d.Vendor.OUIVendor))
	n.OSHint = firstNonEmpty(n.OSHint, d.OSGuess.Family)
	if d.MobileFingerprint != nil {
		fp := *d.MobileFingerprint
		n.MobileFingerprint = &fp
	}
	n.DeviceTypeHint = firstNonEmpty(n.DeviceTypeHint, d.DeviceTypeHint)
	n.MobileOSHint = firstNonEmpty(n.MobileOSHint, d.OSHint)
	n.OSConfidence = maxFloat(n.OSConfidence, d.OSConfidence)
	n.Services = append(n.Services, intelServices(d.Services)...)
	n.Confidence = maxFloat(n.Confidence, d.Confidence)
	n.LastSeen = latest(n.LastSeen, parseTime(d.LastSeen))
	n.Evidence = mergeRichEvidence(n.Evidence, richEvidenceForIDs(d.EvidenceIDs, evidenceByID)...)
	n.RawSources = appendUniqueStrings(n.RawSources, evidenceSourcesFromIntel(d)...)
	n.RiskFlags = appendUniqueStrings(n.RiskFlags, riskFlags(d.SecurityPosture.Findings)...)
	if n.Wireless == nil {
		n.Wireless = wirelessFromIntel(d)
	} else {
		mergeWireless(n.Wireless, wirelessFromIntel(d))
	}
	n.UI = mergeUI(n.UI, map[string]any{
		"is_gateway":      d.Topology.IsGateway,
		"is_agent":        d.Topology.IsAgent,
		"is_upstream_cpe": d.Topology.IsUpstreamGatewayCandidate,
		"classification":  d.ClassificationExplanation,
		"mobile_label":    mobileNodeLabel(d.MobileFingerprint),
	})
}

func mobileNodeLabel(fp *models.MobileFingerprint) string {
	if fp == nil {
		return ""
	}
	switch fp.Classification {
	case models.MobileClassificationConfirmedIOS:
		return "Confirmed iPhone"
	case models.MobileClassificationProbableIOS:
		return "Probable iPhone"
	case models.MobileClassificationPossibleIOS:
		return "Possible iPhone"
	case models.MobileClassificationConfirmedIPadOS:
		return "Confirmed iPad"
	case models.MobileClassificationProbableIPadOS:
		return "Probable iPad"
	case models.MobileClassificationPossibleIPadOS:
		return "Possible iPad"
	case models.MobileClassificationConfirmedAndroid:
		return "Confirmed Android"
	case models.MobileClassificationProbableAndroid:
		return "Probable Android"
	case models.MobileClassificationPossibleAndroid:
		return "Possible Android"
	case models.MobileClassificationConflict:
		return "Conflicting OS evidence"
	case models.MobileClassificationUnknownMobile:
		return "Unknown mobile"
	default:
		return ""
	}
}

func ensureNode(nodes map[string]*models.RichTopologyNode, id string) *models.RichTopologyNode {
	if id == "" {
		id = "unknown"
	}
	n := nodes[id]
	if n == nil {
		n = &models.RichTopologyNode{
			ID:         id,
			Label:      id,
			Type:       models.DeviceTypeUnknown,
			Category:   models.NodeCategoryUnknown,
			DeviceRole: models.NodeRoleUnknown,
			UI:         map[string]any{},
		}
		nodes[id] = n
	}
	return n
}

func edgeFromRoot(e models.TopologyEdge, evidenceByID map[string]models.Evidence) models.RichTopologyEdge {
	typ, medium := richEdgeTypeFromRoot(e)
	if reportedType, reportedMedium := reportedEdgeType(e.EvidenceIDs, evidenceByID); reportedType != "" {
		typ, medium = reportedType, reportedMedium
	}
	return models.RichTopologyEdge{
		ID:         firstNonEmpty(e.ID, fmt.Sprintf("edge-%s-%s-%s", e.Source, e.Target, typ)),
		Source:     e.Source,
		Target:     e.Target,
		Type:       typ,
		Relation:   firstNonEmpty(e.Relationship, e.Type),
		Medium:     medium,
		Confidence: e.Confidence,
		Evidence:   richEvidenceForIDs(e.EvidenceIDs, evidenceByID),
		UI: map[string]any{
			"line_style": e.UILineStyle,
			"physical":   e.Physical,
			"inferred":   e.Inferred,
			"reason":     e.Reason,
		},
	}
}

func edgeFromIntel(e models.DeviceIntelEdge, evidenceByID map[string]models.Evidence) models.RichTopologyEdge {
	typ, medium := richEdgeTypeFromIntel(e)
	if reportedType, reportedMedium := reportedEdgeType(e.EvidenceIDs, evidenceByID); reportedType != "" {
		typ, medium = reportedType, reportedMedium
	}
	return models.RichTopologyEdge{
		ID:         firstNonEmpty(e.ID, fmt.Sprintf("di-edge-%s-%s-%s", e.Source, e.Target, typ)),
		Source:     e.Source,
		Target:     e.Target,
		Type:       typ,
		Relation:   firstNonEmpty(e.Type, typ),
		Medium:     medium,
		Confidence: e.Confidence,
		Evidence:   richEvidenceForIDs(e.EvidenceIDs, evidenceByID),
		UI: map[string]any{
			"physical":         e.Physical,
			"inferred":         e.Inferred,
			"reason":           e.Reason,
			"confidence_label": e.ConfidenceLabel,
		},
	}
}

func reportedEdgeType(ids []string, evidenceByID map[string]models.Evidence) (string, string) {
	for _, id := range ids {
		source := richEvidenceSource(evidenceByID[id])
		switch source {
		case models.EvidenceSourceAPAPI:
			return models.RichEdgeReportedByAP, "wireless"
		case models.EvidenceSourceRouterAPI:
			return models.RichEdgeReportedByRouter, "l3"
		}
	}
	return "", ""
}

func upsertRichEdge(edges map[string]*models.RichTopologyEdge, e models.RichTopologyEdge) {
	if e.Source == "" || e.Target == "" || e.Source == e.Target {
		return
	}
	if e.ID == "" {
		e.ID = fmt.Sprintf("edge-%s-%s-%s", e.Source, e.Target, e.Type)
	}
	existing := edges[e.ID]
	if existing == nil || e.Confidence > existing.Confidence {
		cp := e
		edges[e.ID] = &cp
		return
	}
	existing.Evidence = mergeRichEvidence(existing.Evidence, e.Evidence...)
}

func richEdgeTypeFromRoot(e models.TopologyEdge) (string, string) {
	switch e.Type {
	case models.EdgeGatewayDefault:
		return models.RichEdgeGatewayLink, "l3"
	case models.EdgeInferredL2:
		return models.RichEdgeSubnetInferred, "l2"
	case models.EdgeRouteHop:
		return models.RichEdgeWeakInferred, "l3"
	case models.EdgeDirectLLDP, models.EdgeDirectCDP, models.EdgeSNMPBridge:
		return models.RichEdgeSwitchUplink, "l2"
	default:
		return models.RichEdgeUnknown, strings.ToLower(e.Layer)
	}
}

func richEdgeTypeFromIntel(e models.DeviceIntelEdge) (string, string) {
	switch e.Type {
	case models.DeviceEdgeDefaultGatewayRoute:
		return models.RichEdgeGatewayLink, "l3"
	case models.DeviceEdgeSameSubnetInferred:
		return models.RichEdgeSubnetInferred, "l2"
	case models.DeviceEdgeARPNeighbor:
		return models.RichEdgeARPNeighbor, "l2"
	case models.DeviceEdgeWiFiAssociation:
		if e.Confidence >= 0.80 && !e.Inferred {
			return models.RichEdgeWirelessAssociated, "wireless"
		}
		return models.RichEdgeWirelessObserved, "wireless"
	case models.DeviceEdgeLLDPPhysicalNeighbor, models.DeviceEdgeCDPPhysicalNeighbor, models.DeviceEdgeSNMPBridgeFDB:
		return models.RichEdgeSwitchUplink, "l2"
	case models.DeviceEdgeUpstreamPrivate, models.DeviceEdgePossibleCPEPath, models.DeviceEdgeTracerouteHop:
		return models.RichEdgeWeakInferred, "l3"
	default:
		return models.RichEdgeUnknown, "unknown"
	}
}

func indexEvidence(report models.ScanReport) map[string]models.Evidence {
	out := map[string]models.Evidence{}
	for _, ev := range report.Evidence {
		out[ev.ID] = ev
	}
	return out
}

func richEvidenceForIDs(ids []string, evidenceByID map[string]models.Evidence) []models.RichEvidence {
	out := make([]models.RichEvidence, 0, len(ids))
	for _, id := range ids {
		ev, ok := evidenceByID[id]
		if !ok {
			out = append(out, models.RichEvidence{Source: "manual", Value: id, Confidence: 0.2, Timestamp: time.Time{}, Notes: "referenced evidence id was not present in root evidence"})
			continue
		}
		out = append(out, models.RichEvidence{
			Source:     richEvidenceSource(ev),
			Value:      firstNonEmpty(ev.Summary, ev.ID),
			Confidence: evidenceConfidence(ev.Kind),
			Timestamp:  ev.Timestamp,
			Interface:  ev.Data["interface"],
			Notes:      ev.Data["note"],
		})
	}
	return out
}

func richEvidenceSource(ev models.Evidence) string {
	kind := strings.ToLower(ev.Kind + " " + ev.Source)
	switch {
	case strings.Contains(kind, "arp"):
		return models.EvidenceSourceARP
	case strings.Contains(kind, "neighbor"):
		return models.EvidenceSourceNeighborTable
	case strings.Contains(kind, "icmp"):
		return models.EvidenceSourceICMP
	case strings.Contains(kind, "tcp"), strings.Contains(kind, "nmap"):
		return models.EvidenceSourceTCPProbe
	case strings.Contains(kind, "udp"):
		return models.EvidenceSourceUDPProbe
	case strings.Contains(kind, "mdns"):
		return models.EvidenceSourceMDNS
	case strings.Contains(kind, "ssdp"), strings.Contains(kind, "upnp"):
		return models.EvidenceSourceSSDP
	case strings.Contains(kind, "llmnr"):
		return models.EvidenceSourceLLMNR
	case strings.Contains(kind, "nbns"), strings.Contains(kind, "netbios"):
		return models.EvidenceSourceNBNS
	case strings.Contains(kind, "dhcp"):
		return models.EvidenceSourceDHCP
	case strings.Contains(kind, "dns"):
		return models.EvidenceSourceDNS
	case strings.Contains(kind, "tls"):
		return models.EvidenceSourceTLS
	case strings.Contains(kind, "http"):
		return models.EvidenceSourceHTTP
	case strings.Contains(kind, "snmp"):
		return models.EvidenceSourceSNMP
	case strings.Contains(kind, "router_api"):
		return models.EvidenceSourceRouterAPI
	case strings.Contains(kind, "ap_api"):
		return models.EvidenceSourceAPAPI
	case strings.Contains(kind, "wifi"), strings.Contains(kind, "wireless"):
		return models.EvidenceSourcePassiveWiFi
	default:
		return models.EvidenceSourceManual
	}
}

func evidenceConfidence(kind string) float64 {
	switch strings.ToLower(kind) {
	case "gateway_route", "interface":
		return 0.80
	case "arp_table", "tcp_connect", "nmap":
		return 0.70
	case "lldp", "cdp":
		return 0.95
	case "snmp_bridge":
		return 0.85
	default:
		return 0.50
	}
}

func roleFromRoot(d models.Device) string {
	switch {
	case d.IsAgent:
		return models.NodeRoleLocalAgent
	case d.IsGateway:
		return models.NodeRoleGateway
	case hasRole(d.Roles, models.RoleRouter):
		return models.NodeRoleRouter
	case hasAnyRole(d.Roles, "access_point", "wifi_ap"):
		return models.NodeRoleAccessPoint
	case hasAnyRole(d.Roles, "switch", "switching_device"):
		return models.NodeRoleSwitch
	default:
		return models.NodeRoleWiredClient
	}
}

func roleFromIntel(d models.DeviceIntelDevice) string {
	roles := d.Roles
	switch {
	case d.Topology.IsAgent:
		return models.NodeRoleLocalAgent
	case d.Topology.IsGateway:
		return models.NodeRoleGateway
	case hasAnyRole(roles, models.DeviceRoleUpstreamGateway, models.DeviceRolePossibleCPE):
		return models.NodeRoleRouter
	case strings.Contains(d.DeviceType.Primary, "access_point"):
		return models.NodeRoleAccessPoint
	case strings.Contains(d.DeviceType.Primary, "mesh"):
		return models.NodeRoleMeshNode
	case strings.Contains(d.DeviceType.Primary, "repeater"):
		return models.NodeRoleRepeater
	case strings.Contains(d.DeviceType.Primary, "switch"):
		return models.NodeRoleSwitch
	default:
		return models.NodeRoleWiredClient
	}
}

func nodeTypeFromRoot(d models.Device) string {
	switch {
	case d.IsAgent:
		return "local_host"
	case d.IsGateway:
		return models.DeviceTypeGatewayRouter
	default:
		return "host"
	}
}

func categoryForIntel(d models.DeviceIntelDevice) string {
	if d.Topology.IsGateway || d.Topology.IsUpstreamGatewayCandidate {
		return models.NodeCategoryNetwork
	}
	if d.LLDPCDPInfo != nil {
		return models.NodeCategoryNetwork
	}
	return models.NodeCategoryDevice
}

func wirelessFromIntel(d models.DeviceIntelDevice) *models.RichWireless {
	if d.LLDPCDPInfo == nil && !strings.Contains(d.DeviceType.Primary, "access_point") && !strings.Contains(strings.Join(d.Roles, " "), "wifi") {
		return nil
	}
	return &models.RichWireless{
		IsAP:             strings.Contains(d.DeviceType.Primary, "access_point"),
		IsStation:        !strings.Contains(d.DeviceType.Primary, "access_point"),
		IsMesh:           strings.Contains(d.DeviceType.Primary, "mesh"),
		IsRepeaterHint:   strings.Contains(d.DeviceType.Primary, "repeater"),
		ObservationCount: len(d.EvidenceIDs),
		Confidence:       d.Confidence,
	}
}

func mergeWireless(dst *models.RichWireless, src *models.RichWireless) {
	if dst == nil || src == nil {
		return
	}
	dst.SSID = firstNonEmpty(dst.SSID, src.SSID)
	dst.BSSID = firstNonEmpty(dst.BSSID, src.BSSID)
	dst.AssociatedBSSID = firstNonEmpty(dst.AssociatedBSSID, src.AssociatedBSSID)
	dst.Channel = firstNonZero(dst.Channel, src.Channel)
	dst.Frequency = firstNonZero(dst.Frequency, src.Frequency)
	dst.Band = firstNonEmpty(dst.Band, src.Band)
	dst.RSSI = firstNonZero(dst.RSSI, src.RSSI)
	dst.Noise = firstNonZero(dst.Noise, src.Noise)
	dst.Security = firstNonEmpty(dst.Security, src.Security)
	dst.PHY = firstNonEmpty(dst.PHY, src.PHY)
	dst.Capabilities = appendUniqueStrings(dst.Capabilities, src.Capabilities...)
	dst.IsAP = dst.IsAP || src.IsAP
	dst.IsStation = dst.IsStation || src.IsStation
	dst.IsMesh = dst.IsMesh || src.IsMesh
	dst.IsRepeaterHint = dst.IsRepeaterHint || src.IsRepeaterHint
	dst.ObservationCount += src.ObservationCount
	dst.Confidence = maxFloat(dst.Confidence, src.Confidence)
}

func rootServices(in []models.Service) []models.RichService {
	out := make([]models.RichService, 0, len(in))
	for _, svc := range in {
		out = append(out, models.RichService{
			Port: svc.Port, Protocol: svc.Protocol, State: svc.State,
			Name: svc.Name, Product: svc.Product, EvidenceIDs: append([]string{}, svc.EvidenceIDs...),
		})
	}
	return out
}

func intelServices(in []models.DeviceIntelService) []models.RichService {
	out := make([]models.RichService, 0, len(in))
	for _, svc := range in {
		out = append(out, models.RichService{
			Port: svc.Port, Protocol: svc.Protocol, State: svc.State,
			Name: svc.Name, Product: firstNonEmpty(svc.Product, svc.Version), EvidenceIDs: append([]string{}, svc.EvidenceIDs...),
		})
	}
	return out
}

func richWarnings(report models.ScanReport, edges []models.RichTopologyEdge) []models.Warning {
	out := append([]models.Warning{}, report.Warnings...)
	for _, edge := range edges {
		if edge.Type == models.RichEdgeWeakInferred || edge.Type == models.RichEdgeSubnetInferred {
			out = append(out, models.Warning{Code: "rich_topology_inferred_edges", Severity: models.SeverityInfo, Message: "Some topology relationships are inferred from subnet, route, timing, or path evidence and are not proven physical links."})
			break
		}
	}
	return dedupeWarnings(out)
}

func richCapabilities(report models.ScanReport) []models.ReportCapability {
	caps := append([]models.ReportCapability{}, report.Capabilities...)
	caps = append(caps,
		models.ReportCapability{Name: "rich_topology_graph", Category: "topology", Status: "completed", OutputPath: "/rich_topology", Description: "Frontend-ready graph combining LAN discovery, device intelligence, and evidence-backed topology inference."},
		models.ReportCapability{Name: "passive_lan_observer", Category: "passive_observation", Status: "available", OutputPath: "/rich_topology", Description: "Metadata-only passive LAN observation abstraction is available; capture backend must be explicitly configured."},
		models.ReportCapability{Name: "passive_wifi_observer", Category: "wireless", Status: "unsupported", OutputPath: "/rich_topology/capabilities", Reason: "No monitor/radiotap backend is configured by default; unsupported is expected on many Windows adapters."},
	)
	return caps
}

func rootNodeID(nodes []models.RichTopologyNode) string {
	for _, n := range nodes {
		if n.DeviceRole == models.NodeRoleGateway {
			return n.ID
		}
	}
	for _, n := range nodes {
		if n.DeviceRole == models.NodeRoleLocalAgent {
			return n.ID
		}
	}
	if len(nodes) > 0 {
		return nodes[0].ID
	}
	return ""
}

func sortNode(n *models.RichTopologyNode) {
	sort.Strings(n.IPAddresses)
	sort.Strings(n.MACAddresses)
	sort.Strings(n.RawSources)
	sort.Strings(n.RiskFlags)
	sort.Slice(n.Evidence, func(i, j int) bool { return n.Evidence[i].Value < n.Evidence[j].Value })
}

func sortEdge(e *models.RichTopologyEdge) {
	sort.Slice(e.Evidence, func(i, j int) bool { return e.Evidence[i].Value < e.Evidence[j].Value })
	if e.FirstSeen.IsZero() || e.LastSeen.IsZero() {
		for _, ev := range e.Evidence {
			e.FirstSeen = earliest(e.FirstSeen, ev.Timestamp)
			e.LastSeen = latest(e.LastSeen, ev.Timestamp)
		}
	}
}

func mergeRichEvidence(existing []models.RichEvidence, more ...models.RichEvidence) []models.RichEvidence {
	seen := map[string]bool{}
	out := make([]models.RichEvidence, 0, len(existing)+len(more))
	for _, ev := range append(existing, more...) {
		key := ev.Source + "|" + ev.Value + "|" + ev.Timestamp.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, ev)
	}
	return out
}

func mergeUI(existing map[string]any, more map[string]any) map[string]any {
	if existing == nil {
		existing = map[string]any{}
	}
	for k, v := range more {
		existing[k] = v
	}
	return existing
}

func copyUI(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func deviceIPs(d models.Device) []string {
	var out []string
	for _, a := range d.Addresses {
		out = appendUniqueStrings(out, a.IP)
	}
	for _, iface := range d.Interfaces {
		out = appendUniqueStrings(out, iface.IPs...)
	}
	return out
}

func firstIP(values []models.IPAddress) string {
	for _, value := range values {
		if value.IP != "" {
			return value.IP
		}
	}
	return ""
}

func firstString(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func ptrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func appendUniqueStrings(values []string, more ...string) []string {
	for _, value := range more {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		found := false
		for _, existing := range values {
			if existing == value {
				found = true
				break
			}
		}
		if !found {
			values = append(values, value)
		}
	}
	return values
}

func hasRole(roles []string, want string) bool { return hasAnyRole(roles, want) }

func hasAnyRole(roles []string, wants ...string) bool {
	for _, role := range roles {
		for _, want := range wants {
			if strings.EqualFold(role, want) {
				return true
			}
		}
	}
	return false
}

func maxFloat(a, b float64) float64 {
	if b > a {
		return b
	}
	return a
}

func earliest(a, b time.Time) time.Time {
	if b.IsZero() {
		return a
	}
	if a.IsZero() || b.Before(a) {
		return b
	}
	return a
}

func latest(a, b time.Time) time.Time {
	if b.After(a) {
		return b
	}
	return a
}

func parseTime(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	if ts, err := time.Parse(time.RFC3339, raw); err == nil {
		return ts
	}
	return time.Time{}
}

func generatedAt(report models.ScanReport) time.Time {
	if !report.CreatedAt.IsZero() {
		return report.CreatedAt
	}
	return time.Now().UTC()
}

func evidenceSourcesFromIntel(d models.DeviceIntelDevice) []string {
	var out []string
	for _, svc := range d.Services {
		out = appendUniqueStrings(out, svc.Name)
	}
	if d.UPnPInfo != nil && d.UPnPInfo.Detected {
		out = appendUniqueStrings(out, "upnp")
	}
	if d.SNMPInfo != nil && d.SNMPInfo.Status != "" {
		out = appendUniqueStrings(out, "snmp")
	}
	if len(d.MDNSRecords) > 0 {
		out = appendUniqueStrings(out, "mdns")
	}
	if len(d.SSDPRecords) > 0 {
		out = appendUniqueStrings(out, "ssdp")
	}
	return out
}

func riskFlags(findings []models.SecurityFinding) []string {
	var out []string
	for _, finding := range findings {
		out = appendUniqueStrings(out, finding.ID, finding.Severity)
	}
	return out
}

func roleRank(role string) int {
	switch role {
	case models.NodeRoleGateway:
		return 0
	case models.NodeRoleRouter:
		return 1
	case models.NodeRoleSwitch:
		return 2
	case models.NodeRoleAccessPoint, models.NodeRoleMeshNode, models.NodeRoleRepeater:
		return 3
	case models.NodeRoleLocalAgent:
		return 4
	default:
		return 5
	}
}

func dedupeWarnings(warnings []models.Warning) []models.Warning {
	seen := map[string]bool{}
	var out []models.Warning
	for _, warning := range warnings {
		key := warning.Code + "|" + warning.Message
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, warning)
	}
	return out
}

func warningMessages(warnings []models.Warning) []string {
	out := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		out = append(out, warning.Message)
	}
	return out
}
