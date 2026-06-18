package main

import (
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

type reportEnrichmentOptions struct {
	Full                    bool
	Profile                 string
	ClassificationRequested bool
	NmapRequested           bool
	NmapAvailable           bool
	IncludeVirtual          bool
	RedactionMode           string
	MaskPublicIP            bool
	MaskMAC                 bool
	MaskHostnames           bool
}

func enrichReport(report *models.ScanReport, opts reportEnrichmentOptions) {
	enrichInterfaceIdentity(report)
	report.Capabilities = buildReportCapabilities(*report, reportCapabilityOptions{
		Full:                    opts.Full,
		Profile:                 opts.Profile,
		ClassificationRequested: opts.ClassificationRequested,
		NmapRequested:           opts.NmapRequested,
		NmapAvailable:           opts.NmapAvailable,
	})
	report.InterfaceSelection = buildInterfaceSelection(*report, opts)
	report.ProbeInventory = buildProbeInventory(*report, opts)
	report.EvidenceRegistry = buildEvidenceRegistry(*report)
	report.UI = buildReportUI(*report)
	report.RedactionMode = opts.RedactionMode
	report.Privacy = models.PrivacyOptions{
		MaskPublicIP:  opts.MaskPublicIP,
		MaskMAC:       opts.MaskMAC,
		MaskHostnames: opts.MaskHostnames,
	}
	report.SafeToShare = models.SafeShareReport{
		Enabled: opts.RedactionMode == "safe_to_share" || opts.MaskPublicIP || opts.MaskMAC || opts.MaskHostnames,
		Mode:    opts.RedactionMode,
		Applied: report.Privacy,
		Notes: []string{
			"Default probes are read-only and scoped to private/local targets.",
			"Review public IPs, MAC addresses, hostnames, and service banners before sharing externally.",
		},
	}
	if report.SafeToShare.Enabled {
		applyReportRedaction(report)
	}
}

func buildEvidenceRegistry(report models.ScanReport) []models.EvidenceRegistryItem {
	reg := map[string]models.EvidenceRegistryItem{}
	add := func(item models.EvidenceRegistryItem) {
		if item.ID == "" {
			return
		}
		if item.Source == "" {
			item.Source = "unknown"
		}
		if item.Kind == "" {
			item.Kind = "unknown"
		}
		item.SafeToDisplay = true
		reg[item.ID] = item
	}

	for _, ev := range report.Evidence {
		data := map[string]any{}
		for k, v := range ev.Data {
			data[k] = v
		}
		add(models.EvidenceRegistryItem{
			ID:            ev.ID,
			Kind:          ev.Kind,
			Source:        ev.Source,
			Summary:       ev.Summary,
			Data:          data,
			Timestamp:     ev.Timestamp,
			Confidence:    evidenceConfidenceForKind(ev.Kind),
			SafeToDisplay: true,
		})
	}
	if report.DeviceIntel != nil {
		for _, ev := range report.DeviceIntel.Evidence {
			add(models.EvidenceRegistryItem{
				ID:            ev.ID,
				Kind:          ev.Kind,
				Source:        ev.SourceProbe,
				Summary:       firstNonEmptyString(ev.Error, ev.Kind),
				Data:          mergeAnyMaps(ev.Raw, ev.Normalized),
				Timestamp:     ev.Timestamp,
				Confidence:    ev.Confidence,
				SafeToDisplay: ev.SafeToDisplay,
			})
		}
	}

	for i, pr := range accessProbeResults(report) {
		id := fmt.Sprintf("probe-%02d-%s", i, sanitizeID(pr.ProbeName))
		add(models.EvidenceRegistryItem{
			ID:            id,
			Kind:          "probe_result",
			Source:        pr.ProbeName,
			Summary:       fmt.Sprintf("%s finished with status %s.", pr.ProbeName, pr.Status),
			Data:          copyAnyMap(pr.Evidence),
			Confidence:    pr.Confidence,
			SafeToDisplay: true,
		})
	}

	for _, id := range collectReferencedEvidenceIDs(report) {
		if _, ok := reg[id]; ok {
			continue
		}
		add(models.EvidenceRegistryItem{
			ID:            id,
			Kind:          "referenced_evidence",
			Source:        "reference_index",
			Summary:       "Evidence ID referenced by a report object; original producer did not emit a standalone root evidence record.",
			SafeToDisplay: true,
		})
	}

	out := make([]models.EvidenceRegistryItem, 0, len(reg))
	for _, item := range reg {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func buildProbeInventory(report models.ScanReport, opts reportEnrichmentOptions) []models.ProbeInventoryItem {
	items := []models.ProbeInventoryItem{
		{
			Name:                  "interface_probe",
			Category:              "topology",
			Status:                "completed",
			ProducedEvidenceCount: countEvidenceBySource(report.Evidence, "interface_probe"),
			SafetyMode:            "local_os_read_only",
			OutputPath:            "/agent/interfaces",
		},
		{
			Name:                  "gateway_probe",
			Category:              "topology",
			Status:                "completed",
			ProducedEvidenceCount: countEvidenceBySource(report.Evidence, "gateway_probe"),
			SafetyMode:            "local_os_read_only",
			OutputPath:            "/agent/gateway",
		},
		{
			Name:                  "tcp_lan_sweep",
			Category:              "topology",
			Status:                "completed",
			DurationMS:            report.Summary.DurationMS,
			ProducedEvidenceCount: len(report.Devices),
			SafetyMode:            "private_scope_read_only_tcp_connect",
			OutputPath:            "/devices",
		},
		{
			Name:                  "arp_table_probe",
			Category:              "topology",
			Status:                "completed",
			ProducedEvidenceCount: countEvidenceKind(report.Evidence, "arp_table"),
			SafetyMode:            "local_os_read_only",
			OutputPath:            "/evidence",
		},
	}
	nmap := models.ProbeInventoryItem{
		Name:       "nmap_service_discovery",
		Category:   "optional_tool",
		Status:     "skipped",
		SafetyMode: "private_scope_read_only_tcp_connect",
		OutputPath: "/devices/services",
	}
	if opts.NmapRequested && opts.NmapAvailable {
		nmap.Status = "completed"
		nmap.ProducedEvidenceCount = serviceCount(report.Devices)
	} else if opts.NmapRequested {
		nmap.Status = "unavailable"
		nmap.SkippedReason = "nmap binary was not found"
	}
	items = append(items, nmap)

	for i, pr := range accessProbeResults(report) {
		item := models.ProbeInventoryItem{
			Name:                  pr.ProbeName,
			Category:              "access_probe",
			Status:                probeStatus(pr.Status),
			DurationMS:            pr.DurationMS,
			ProducedEvidenceCount: producedEvidenceCount(pr),
			SafetyMode:            probeSafetyMode(pr.ProbeName),
			OutputPath:            fmt.Sprintf("/access_classification/evidence/%d", i),
		}
		if len(pr.Errors) > 0 {
			item.ErrorClass = classifyProbeError(strings.Join(pr.Errors, "; "))
			item.Timeout = item.ErrorClass == "timeout"
			item.SkippedReason = strings.Join(pr.Errors, "; ")
		}
		items = append(items, item)
	}
	return items
}

func buildInterfaceSelection(report models.ScanReport, opts reportEnrichmentOptions) models.InterfaceSelectionDiagnostics {
	out := models.InterfaceSelectionDiagnostics{
		GatewayIP:      report.Agent.Gateway,
		IncludeVirtual: opts.IncludeVirtual,
	}
	for _, ifc := range report.Agent.Interfaces {
		score, reasons, ignored := interfaceSelectionScore(ifc, report.Agent.Gateway, opts.IncludeVirtual)
		c := models.InterfaceSelectionCandidate{
			Name:     ifc.Name,
			Selected: ifc.Selected,
			Score:    score,
			Reasons:  reasons,
			Ignored:  ignored,
		}
		if ifc.Selected {
			out.SelectedInterface = ifc.Name
			out.Reason = strings.Join(reasons, "; ")
		}
		out.Candidates = append(out.Candidates, c)
	}
	return out
}

func buildReportUI(report models.ScanReport) models.ReportUI {
	ui := models.ReportUI{
		Graph: buildUIGraph(report),
		Panels: models.UIPanels{
			Summary: map[string]any{
				"scan_id":        report.ScanID,
				"profile":        report.Scope.Profile,
				"device_count":   report.Summary.DeviceCount,
				"edge_count":     report.Summary.EdgeCount,
				"evidence_count": len(report.EvidenceRegistry),
				"inferred_only":  report.Summary.InferredOnly,
			},
			DeviceDetails:    buildUIDeviceDetails(report),
			EvidenceTimeline: buildEvidenceTimeline(report.EvidenceRegistry),
		},
		Badges:   []string{"read-only", "private-scope", "no-bruteforce"},
		Warnings: warningMessages(report.Warnings),
	}
	if report.DeviceIntel != nil {
		ui.Badges = appendUniqueStringsLocal(ui.Badges, report.DeviceIntel.UI.Badges...)
		ui.Warnings = appendUniqueStringsLocal(ui.Warnings, report.DeviceIntel.SecurityNotes...)
	}
	if report.AccessClassification != nil {
		ui.Badges = appendUniqueStringsLocal(ui.Badges, report.AccessClassification.UI.Badges...)
		ui.Warnings = appendUniqueStringsLocal(ui.Warnings, report.AccessClassification.UI.Warnings...)
		ui.NextActions = report.AccessClassification.NextBestProbes
	}
	return ui
}

func buildUIGraph(report models.ScanReport) models.UIGraph {
	graph := models.UIGraph{
		Nodes: make([]models.UIGraphNode, 0, len(report.Devices)),
		Edges: make([]models.UIGraphEdge, 0, len(report.Edges)),
	}
	for _, d := range report.Devices {
		graph.Nodes = append(graph.Nodes, models.UIGraphNode{
			ID:         d.ID,
			Label:      deviceLabel(d),
			Type:       deviceTypeForUI(d),
			Confidence: d.Confidence,
			Inferred:   false,
			Badges:     append([]string{}, d.Roles...),
			Metadata: map[string]any{
				"is_agent":     d.IsAgent,
				"is_gateway":   d.IsGateway,
				"addresses":    d.Addresses,
				"evidence_ids": d.EvidenceIDs,
			},
		})
	}
	for _, e := range report.Edges {
		graph.Edges = append(graph.Edges, models.UIGraphEdge{
			ID:           e.ID,
			Source:       e.Source,
			Target:       e.Target,
			Layer:        e.Layer,
			Relationship: e.Relationship,
			Physical:     e.Physical,
			Inferred:     e.Inferred,
			Confidence:   e.Confidence,
			ProofSource:  e.ProofSource,
			LineStyle:    e.UILineStyle,
		})
	}
	return graph
}

func buildUIDeviceDetails(report models.ScanReport) []map[string]any {
	out := make([]map[string]any, 0, len(report.Devices))
	for _, d := range report.Devices {
		out = append(out, map[string]any{
			"id":           d.ID,
			"label":        deviceLabel(d),
			"addresses":    d.Addresses,
			"roles":        d.Roles,
			"services":     d.Services,
			"confidence":   d.Confidence,
			"evidence_ids": d.EvidenceIDs,
		})
	}
	return out
}

func buildEvidenceTimeline(reg []models.EvidenceRegistryItem) []models.EvidenceTimelineEntry {
	out := make([]models.EvidenceTimelineEntry, 0, len(reg))
	for _, ev := range reg {
		out = append(out, models.EvidenceTimelineEntry{
			EvidenceID: ev.ID,
			Source:     ev.Source,
			Kind:       ev.Kind,
			Summary:    ev.Summary,
			Timestamp:  ev.Timestamp,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Timestamp.Equal(out[j].Timestamp) {
			return out[i].EvidenceID < out[j].EvidenceID
		}
		return out[i].Timestamp.Before(out[j].Timestamp)
	})
	return out
}

func collectReferencedEvidenceIDs(report models.ScanReport) []string {
	seen := map[string]bool{}
	var walk func(reflect.Value)
	walk = func(v reflect.Value) {
		if !v.IsValid() {
			return
		}
		if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
			if v.IsNil() {
				return
			}
			walk(v.Elem())
			return
		}
		switch v.Kind() {
		case reflect.Struct:
			if v.Type().PkgPath() == "time" && v.Type().Name() == "Time" {
				return
			}
			t := v.Type()
			for i := 0; i < v.NumField(); i++ {
				name := t.Field(i).Name
				f := v.Field(i)
				if name == "EvidenceID" && f.Kind() == reflect.String {
					if id := f.String(); id != "" {
						seen[id] = true
					}
					continue
				}
				if name == "EvidenceIDs" && f.Kind() == reflect.Slice {
					for j := 0; j < f.Len(); j++ {
						if f.Index(j).Kind() == reflect.String {
							if id := f.Index(j).String(); id != "" {
								seen[id] = true
							}
						}
					}
					continue
				}
				walk(f)
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i))
			}
		case reflect.Map:
			for _, key := range v.MapKeys() {
				walk(v.MapIndex(key))
			}
		}
	}
	walk(reflect.ValueOf(report))
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func enrichInterfaceIdentity(report *models.ScanReport) {
	for i := range report.Agent.Interfaces {
		enrichInterfaceInfo(&report.Agent.Interfaces[i])
	}
}

func enrichInterfaceInfo(ifc *models.InterfaceInfo) {
	if ifc.MAC == "" {
		return
	}
	hw, err := net.ParseMAC(ifc.MAC)
	if err != nil || len(hw) == 0 {
		return
	}
	ifc.LocallyAdministeredMAC = hw[0]&0x02 != 0
	ifc.RandomizedMACLikely = ifc.LocallyAdministeredMAC && !ifc.Virtual
	if ifc.LocallyAdministeredMAC {
		ifc.OUIVendorDBVersion = "local-admin-bit"
		return
	}
	ifc.OUIVendorDBVersion = "builtin-lite-2026-06"
	ifc.OUIVendor = ouiVendor(hw)
}

func ouiVendor(hw net.HardwareAddr) string {
	if len(hw) < 3 {
		return ""
	}
	prefix := strings.ToUpper(fmt.Sprintf("%02X:%02X:%02X", hw[0], hw[1], hw[2]))
	switch prefix {
	case "00:50:56", "00:0C:29", "00:05:69":
		return "VMware"
	case "08:00:27":
		return "Oracle VirtualBox"
	case "00:15:5D":
		return "Microsoft Hyper-V"
	case "F4:F5:D8", "74:D4:DD", "D8:BB:C1":
		return "Intel"
	case "A4:83:E7", "E8:9C:25":
		return "Apple"
	default:
		return ""
	}
}

func interfaceSelectionScore(ifc models.InterfaceInfo, gatewayIP string, includeVirtual bool) (int, []string, []string) {
	score := 0
	var reasons []string
	var ignored []string
	if !ifc.Up {
		ignored = append(ignored, "interface is down")
	}
	if ifc.Loopback {
		ignored = append(ignored, "loopback interface")
	}
	if ifc.Virtual && !includeVirtual {
		ignored = append(ignored, "virtual or overlay adapter excluded")
	}
	if ifc.CIDR == "" {
		ignored = append(ignored, "no usable IPv4 CIDR")
	}
	if len(ignored) > 0 {
		return score, reasons, ignored
	}
	if hasRoutablePrivateIPv4Local(ifc) {
		score += 4
		reasons = append(reasons, "has routable RFC1918 IPv4 network")
	}
	if hasLinkLocalIPv4Local(ifc) {
		score += 1
		reasons = append(reasons, "has APIPA/link-local IPv4 only; last-resort candidate")
	}
	if gatewayIP != "" && cidrContainsLocal(ifc.CIDR, gatewayIP) {
		score += 10
		reasons = append(reasons, "contains default gateway")
	}
	if ifc.Virtual {
		reasons = append(reasons, "virtual adapter included by explicit option")
	}
	return score, reasons, ignored
}

func applyReportRedaction(report *models.ScanReport) {
	if report.Privacy.MaskMAC {
		for i := range report.Agent.Interfaces {
			if report.Agent.Interfaces[i].MAC != "" {
				report.Agent.Interfaces[i].MAC = "xx:xx:xx:xx:xx:xx"
			}
		}
		for i := range report.Devices {
			for j := range report.Devices[i].Interfaces {
				if report.Devices[i].Interfaces[j].MAC != "" {
					report.Devices[i].Interfaces[j].MAC = "xx:xx:xx:xx:xx:xx"
				}
			}
		}
	}
	if report.Privacy.MaskHostnames {
		report.Agent.Hostname = "redacted-host"
		for i := range report.Devices {
			if report.Devices[i].Hostname != "" {
				report.Devices[i].Hostname = "redacted-host"
			}
		}
	}
	if report.Privacy.MaskPublicIP && report.AccessClassification != nil && report.AccessClassification.DetectedNetworkContext != nil {
		ctx := report.AccessClassification.DetectedNetworkContext
		if ctx.PublicIP != "" {
			ctx.PublicIP = "x.x.x.x"
		}
		if ctx.NATTopology != nil {
			if ctx.NATTopology.PublicIP != "" {
				ctx.NATTopology.PublicIP = "x.x.x.x"
			}
			if ctx.NATTopology.STUNPublicIP != "" {
				ctx.NATTopology.STUNPublicIP = "x.x.x.x"
			}
		}
	}
}

func accessProbeResults(report models.ScanReport) []models.ProbeResult {
	if report.AccessClassification == nil {
		return nil
	}
	return report.AccessClassification.Evidence
}

func countEvidenceBySource(evidence []models.Evidence, source string) int {
	n := 0
	for _, ev := range evidence {
		if ev.Source == source {
			n++
		}
	}
	return n
}

func countEvidenceKind(evidence []models.Evidence, kind string) int {
	n := 0
	for _, ev := range evidence {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

func serviceCount(devices []models.Device) int {
	n := 0
	for _, d := range devices {
		n += len(d.Services)
	}
	return n
}

func producedEvidenceCount(pr models.ProbeResult) int {
	n := len(pr.Evidence)
	for _, v := range pr.Evidence {
		switch vv := v.(type) {
		case []models.GatewayDevice:
			n += len(vv)
		case []models.WANSignal:
			n += len(vv)
		case []string:
			n += len(vv)
		}
	}
	return n
}

func probeStatus(status string) string {
	if status == "" {
		return "unknown"
	}
	if status == models.StatusSuccess {
		return "completed"
	}
	return status
}

func probeSafetyMode(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "snmp"), strings.Contains(lower, "tr064"), strings.Contains(lower, "tr181"):
		return "private_scope_read_only_opt_in_for_credentials"
	case strings.Contains(lower, "public_ip"), strings.Contains(lower, "asn"):
		return "online_context_probe_no_scope_scan"
	case strings.Contains(lower, "traceroute"):
		return "bounded_route_observation"
	case strings.Contains(lower, "http"), strings.Contains(lower, "upnp"), strings.Contains(lower, "gateway"):
		return "private_scope_read_only_no_login"
	default:
		return "read_only"
	}
}

func classifyProbeError(text string) string {
	l := strings.ToLower(text)
	switch {
	case strings.Contains(l, "timeout"), strings.Contains(l, "deadline"):
		return "timeout"
	case strings.Contains(l, "permission"), strings.Contains(l, "access denied"):
		return "permission"
	case strings.Contains(l, "not found"):
		return "unavailable"
	default:
		return "runtime_error"
	}
}

func evidenceConfidenceForKind(kind string) float64 {
	switch kind {
	case "gateway_route":
		return 0.85
	case "interface":
		return 0.80
	case "arp_table":
		return 0.65
	case "tcp_connect", "nmap":
		return 0.75
	default:
		return 0.50
	}
}

func mergeAnyMaps(a, b map[string]any) map[string]any {
	out := copyAnyMap(a)
	for k, v := range b {
		out[k] = v
	}
	return out
}

func copyAnyMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func sanitizeID(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	repl := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ":", "-")
	return repl.Replace(s)
}

func deviceLabel(d models.Device) string {
	if d.Hostname != "" {
		return d.Hostname
	}
	if len(d.Addresses) > 0 {
		return d.Addresses[0].IP
	}
	return d.ID
}

func deviceTypeForUI(d models.Device) string {
	if d.IsAgent {
		return "agent"
	}
	if d.IsGateway {
		return "gateway"
	}
	if len(d.Services) > 0 {
		return "host_with_services"
	}
	return "host"
}

func warningMessages(warnings []models.Warning) []string {
	out := make([]string, 0, len(warnings))
	for _, w := range warnings {
		out = append(out, w.Message)
	}
	return out
}

func appendUniqueStringsLocal(values []string, more ...string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values)+len(more))
	for _, v := range append(values, more...) {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func hasRoutablePrivateIPv4Local(ifc models.InterfaceInfo) bool {
	for _, a := range ifc.Addresses {
		ip := net.ParseIP(a.IP)
		if ip == nil || ip.To4() == nil {
			continue
		}
		if isPrivateIPv4Local(ip) && !ip.IsLinkLocalUnicast() {
			return true
		}
	}
	return false
}

func hasLinkLocalIPv4Local(ifc models.InterfaceInfo) bool {
	for _, a := range ifc.Addresses {
		if ip := net.ParseIP(a.IP); ip != nil && ip.To4() != nil && ip.IsLinkLocalUnicast() {
			return true
		}
	}
	return false
}

func cidrContainsLocal(cidr, ipStr string) bool {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(ipStr)
	return ip != nil && ipnet.Contains(ip)
}

func isPrivateIPv4Local(ip net.IP) bool {
	v4 := ip.To4()
	if v4 == nil {
		return false
	}
	return v4[0] == 10 || (v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31) || (v4[0] == 192 && v4[1] == 168)
}
