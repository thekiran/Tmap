package deviceintel

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/thekiran/iad/internal/safety"
	"github.com/thekiran/iad/pkg/models"
)

func Build(report models.ScanReport) models.DeviceIntelReport {
	store := NewEvidenceStore(func() time.Time {
		if !report.CreatedAt.IsZero() {
			return report.CreatedAt
		}
		return time.Now()
	})
	evidenceByID := evidenceIndex(report.Evidence)
	scope := buildScope(report)

	ingestTopologyDevices(report, store, evidenceByID, scope)
	ingestEvidenceRecords(report, store, evidenceByID, scope)
	ingestAccessContext(report, store, evidenceByID, scope)
	edges := convertEdges(report, store)
	classifyAllWithEvidence(store, report.Evidence)

	devices := store.DeviceList()
	out := models.DeviceIntelReport{
		SchemaVersion: models.DeviceIntelSchema,
		ScanID:        report.ScanID,
		CreatedAt:     report.CreatedAt,
		Scope:         scope,
		Devices:       devices,
		Edges:         edges,
		Evidence:      store.Observations,
		Conflicts:     collectConflicts(report, store),
		Warnings:      report.Warnings,
		SecurityNotes: []string{
			"Default device intelligence probes are read-only and limited to private/local targets.",
			"SNMP, SSH, router APIs, TR-064, TR-181, and authenticated APIs require explicit user opt-in credentials.",
			"No login forms, credential guessing, exploit checks, configuration changes, or print jobs are attempted.",
		},
		Undetermined: []string{
			"Exact WAN access type requires direct physical-layer CPE evidence or opt-in router/modem telemetry.",
			"Physical cable or Wi-Fi association topology requires LLDP/CDP/SNMP bridge/FDB or controller data.",
			"Device model and OS remain Unknown when only generic banners, MAC OUI, latency, PTR, or open ports are observed.",
		},
	}
	out.Summary = summarize(out)
	out.UI = buildUI(out)
	return out
}

func buildScope(report models.ScanReport) models.DeviceIntelScope {
	agentIP := selectedAgentIP(report.Agent.Interfaces)
	gateway := report.Agent.Gateway
	if report.AccessClassification != nil && report.AccessClassification.DetectedNetworkContext != nil {
		if gateway == "" {
			gateway = report.AccessClassification.DetectedNetworkContext.Gateway
		}
	}
	profile := report.Scope.Profile
	if profile == "" {
		profile = "device_intel"
	} else {
		profile += "_device_intel"
	}
	return models.DeviceIntelScope{
		CIDR:           report.Scope.CIDR,
		Interface:      report.Scope.Interface,
		AgentIP:        agentIP,
		DefaultGateway: gateway,
		PrivateOnly:    !report.Scope.PublicAllowed,
		PublicAllowed:  report.Scope.PublicAllowed,
		Profile:        profile,
	}
}

func ingestTopologyDevices(report models.ScanReport, store *EvidenceStore, evidenceByID map[string]models.Evidence, scope models.DeviceIntelScope) {
	for _, src := range report.Devices {
		ip := devicePrimaryIP(src)
		if ip == "" || !targetAllowed(scope, ip) {
			continue
		}
		dst := store.UpsertDevice(ip)
		dst.ID = src.ID
		dst.IPAddresses = appendUnique(dst.IPAddresses, deviceIPs(src)...)
		if src.Hostname != "" {
			store.AddHostname(ip, src.Hostname, "topology_device", firstEvidence(src.EvidenceIDs), 0.55)
		}
		dst.Roles = appendUnique(dst.Roles, src.Roles...)
		if src.MobileFingerprint != nil {
			fp := *src.MobileFingerprint
			dst.MobileFingerprint = &fp
			dst.DeviceTypeHint = src.DeviceTypeHint
			dst.OSHint = src.OSHint
			dst.OSConfidence = src.OSConfidence
		}
		dst.Topology.IsAgent = src.IsAgent
		dst.Topology.IsGateway = src.IsGateway
		dst.Confidence = maxFloat(dst.Confidence, src.Confidence)
		dst.EvidenceIDs = sortedUnique(append(dst.EvidenceIDs, src.EvidenceIDs...))
		registerEvidenceIDs(store, evidenceByID, dst.ID, src.EvidenceIDs, src.Confidence)

		for _, iface := range src.Interfaces {
			store.AddMAC(ip, iface.MAC, firstNonEmpty(iface.Vendor, src.Vendor), firstEvidence(src.EvidenceIDs), 0.60)
			if iface.Vendor != "" {
				dst.Vendor.OUIVendor = ptrString(iface.Vendor)
				dst.Vendor.Confidence = maxFloat(dst.Vendor.Confidence, 0.55)
			}
		}
		if src.Vendor != "" {
			dst.Vendor.OUIVendor = ptrString(src.Vendor)
			dst.Vendor.Confidence = maxFloat(dst.Vendor.Confidence, 0.55)
		}
		for _, svc := range src.Services {
			store.AddService(ip, models.DeviceIntelService{
				Port:        svc.Port,
				Protocol:    firstNonEmpty(svc.Protocol, "tcp"),
				State:       firstNonEmpty(svc.State, "open"),
				Name:        firstNonEmpty(svc.Name, serviceName(svc.Port)),
				Product:     svc.Product,
				Confidence:  0.75,
				EvidenceIDs: sortedUnique(svc.EvidenceIDs),
			})
			registerEvidenceIDs(store, evidenceByID, dst.ID, svc.EvidenceIDs, 0.75)
		}
	}
}

func ingestEvidenceRecords(report models.ScanReport, store *EvidenceStore, evidenceByID map[string]models.Evidence, scope models.DeviceIntelScope) {
	for _, ev := range report.Evidence {
		ip := evidenceIP(ev)
		if ip == "" || !targetAllowed(scope, ip) {
			continue
		}
		d := store.UpsertDevice(ip)
		store.RegisterEvidence(ev, d.ID, evidenceConfidence(ev))
		d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, ev.ID))
		switch strings.ToLower(ev.Kind) {
		case "arp_table":
			store.AddMAC(ip, ev.Data["mac"], ev.Data["vendor"], ev.ID, 0.65)
		case "reverse_dns":
			store.AddHostname(ip, ev.Data["hostname"], ev.Source, ev.ID, 0.45)
		case "tcp_connect", "nmap":
			for _, port := range parsePorts(ev.Data["ports"]) {
				store.AddService(ip, models.DeviceIntelService{
					Port:        port,
					Protocol:    "tcp",
					State:       "open",
					Name:        serviceName(port),
					Confidence:  0.75,
					EvidenceIDs: []string{ev.ID},
				})
			}
		case "icmp_echo":
			if strings.EqualFold(ev.Data["status"], "timeout") {
				store.AddFailedAttempt(ip, models.ProbeAttempt{Source: ev.Source, Target: ip, Protocol: "icmp", Error: "timeout", Timeout: true, EvidenceID: ev.ID})
			}
		case "http_fingerprint", "http", "http_fingerprint_v2", "http_fingerprint_v3":
			ingestHTTPEvidence(ip, ev, store)
		case "tls_fingerprint", "tls":
			ingestTLSEvidence(ip, ev, store)
		case "mdns":
			store.AddMDNS(ip, models.MDNSRecord{Name: ev.Data["name"], Service: ev.Data["service"], Target: ev.Data["target"], Text: splitList(ev.Data["txt"]), EvidenceIDs: []string{ev.ID}})
		case "ssdp":
			store.AddSSDP(ip, models.SSDPRecord{USN: ev.Data["usn"], ST: ev.Data["st"], Server: ev.Data["server"], Location: ev.Data["location"], EvidenceIDs: []string{ev.ID}})
		case "nbns", "netbios":
			store.AddNBNS(ip, models.NBNSRecord{Name: ev.Data["name"], Workgroup: ev.Data["workgroup"], EvidenceIDs: []string{ev.ID}})
		case "llmnr":
			store.AddLLMNR(ip, models.LLMNRRecord{Name: ev.Data["name"], EvidenceIDs: []string{ev.ID}})
		case "printer", "ipp", "wsd":
			addPrinterEvidence(ip, ev, store)
		}
		_ = evidenceByID
	}
}

func ingestAccessContext(report models.ScanReport, store *EvidenceStore, evidenceByID map[string]models.Evidence, scope models.DeviceIntelScope) {
	if report.AccessClassification == nil {
		return
	}
	ctx := report.AccessClassification.DetectedNetworkContext
	if ctx != nil {
		if ctx.GatewayChainState != nil {
			for _, hop := range ctx.GatewayChainState.PrivateHops {
				if !targetAllowed(scope, hop.IP) {
					continue
				}
				d := store.UpsertDevice(hop.IP)
				d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, hop.EvidenceID))
				d.Confidence = maxFloat(d.Confidence, ctx.GatewayChainState.Confidence)
				if hop.Role == "default_gateway" {
					d.Topology.IsGateway = true
					d.Roles = appendUnique(d.Roles, models.RoleGateway, models.RoleRouter)
				}
				if hop.Role == "upstream_private_gateway" {
					d.Topology.IsUpstreamGatewayCandidate = true
					d.Roles = appendUnique(d.Roles, models.DeviceRoleUpstreamGateway, models.DeviceRolePossibleCPE)
				}
			}
		}
		for _, gd := range ctx.GatewayDevices {
			ingestGatewayDevice(gd, store, evidenceByID, scope)
		}
	}
	if report.AccessClassification.ModemCollection != nil {
		for _, cpe := range report.AccessClassification.ModemCollection.CPECandidates {
			ingestCPECandidate(cpe, store, evidenceByID, scope)
		}
	}
}

func ingestGatewayDevice(gd models.GatewayDevice, store *EvidenceStore, evidenceByID map[string]models.Evidence, scope models.DeviceIntelScope) {
	if gd.IP == "" || !targetAllowed(scope, gd.IP) {
		return
	}
	d := store.UpsertDevice(gd.IP)
	switch gd.Role {
	case "default_gateway":
		d.Topology.IsGateway = true
		d.Roles = appendUnique(d.Roles, models.RoleGateway, models.RoleRouter)
	case "upstream_private_gateway", "possible_cpe":
		d.Topology.IsUpstreamGatewayCandidate = true
		d.Roles = appendUnique(d.Roles, models.DeviceRoleUpstreamGateway, models.DeviceRolePossibleCPE)
	}
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, gd.EvidenceIDs...))
	registerEvidenceIDs(store, evidenceByID, d.ID, gd.EvidenceIDs, gd.Confidence)
	for _, port := range gd.OpenPorts {
		store.AddService(gd.IP, models.DeviceIntelService{Port: port, Protocol: "tcp", State: "open", Name: serviceName(port), Confidence: 0.80, EvidenceIDs: gd.EvidenceIDs})
	}
	for _, obs := range gd.HTTPObservations {
		store.AddHTTPObservation(gd.IP, obs, true, models.ProbeAttempt{})
	}
	for _, obs := range gd.TLSObservations {
		store.AddTLSObservation(gd.IP, obs)
	}
	for _, attempt := range gd.FailedAttempts {
		store.AddFailedAttempt(gd.IP, attempt)
	}
	if gd.MACVendor != "" {
		d.Vendor.OUIVendor = ptrString(gd.MACVendor)
		d.Vendor.Confidence = maxFloat(d.Vendor.Confidence, 0.55)
	}
	if gd.Manufacturer != "" {
		d.Vendor.FingerprintVendor = ptrString(gd.Manufacturer)
		d.Vendor.Confidence = maxFloat(d.Vendor.Confidence, 0.70)
	}
	if gd.UPnPFound || gd.UPnPIGDFound {
		d.UPnPInfo = &models.UPnPInfo{Detected: true, IGD: gd.UPnPIGDFound, EvidenceIDs: gd.EvidenceIDs}
	}
	if gd.SNMPState != "" {
		d.SNMPInfo = &models.SNMPInfo{Enabled: gd.SNMPState == "enabled", Status: gd.SNMPState, EvidenceIDs: gd.EvidenceIDs}
	}
	d.Confidence = maxFloat(d.Confidence, gd.Confidence)
}

func ingestCPECandidate(cpe models.CPECandidate, store *EvidenceStore, evidenceByID map[string]models.Evidence, scope models.DeviceIntelScope) {
	if cpe.IP == "" || !targetAllowed(scope, cpe.IP) {
		return
	}
	d := store.UpsertDevice(cpe.IP)
	if cpe.Role == "default_gateway" {
		d.Topology.IsGateway = true
		d.Roles = appendUnique(d.Roles, models.RoleGateway, models.RoleRouter)
	} else {
		d.Topology.IsUpstreamGatewayCandidate = true
		d.Roles = appendUnique(d.Roles, models.DeviceRoleUpstreamGateway, models.DeviceRolePossibleCPE)
	}
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, cpe.EvidenceIDs...))
	registerEvidenceIDs(store, evidenceByID, d.ID, cpe.EvidenceIDs, cpe.Confidence)
	for _, port := range cpe.OpenPorts {
		store.AddService(cpe.IP, models.DeviceIntelService{Port: port, Protocol: "tcp", State: "open", Name: serviceName(port), Confidence: 0.75, EvidenceIDs: cpe.EvidenceIDs})
	}
	for _, obs := range cpe.HTTP.Observations {
		store.AddHTTPObservation(cpe.IP, obs, true, models.ProbeAttempt{})
	}
	for _, cert := range cpe.TLS.Certificates {
		store.AddTLSObservation(cpe.IP, cert)
	}
	for _, attempt := range cpe.FailedAttempts {
		store.AddFailedAttempt(cpe.IP, attempt)
	}
	if cpe.ModelFingerprint.Vendor != nil {
		d.Vendor.FingerprintVendor = cpe.ModelFingerprint.Vendor
		d.Vendor.Confidence = maxFloat(d.Vendor.Confidence, cpe.ModelFingerprint.Confidence)
	}
	if cpe.UPnP.Found == models.TriTrue || cpe.UPnP.IGDFound == models.TriTrue {
		d.UPnPInfo = &models.UPnPInfo{Detected: true, IGD: cpe.UPnP.IGDFound == models.TriTrue, EvidenceIDs: cpe.EvidenceIDs}
	}
	if cpe.SNMP.Status != "" {
		d.SNMPInfo = &models.SNMPInfo{Enabled: cpe.SNMP.Enabled, Status: cpe.SNMP.Status, Reason: cpe.SNMP.Reason, EvidenceIDs: cpe.EvidenceIDs}
	}
	d.Confidence = maxFloat(d.Confidence, cpe.Confidence)
}

func ingestHTTPEvidence(ip string, ev models.Evidence, store *EvidenceStore) {
	status, _ := strconv.Atoi(ev.Data["status_code"])
	if strings.EqualFold(ev.Data["status"], "timeout") || strings.EqualFold(ev.Data["success"], "false") || ev.Data["error"] != "" {
		store.AddFailedAttempt(ip, models.ProbeAttempt{
			Source:     ev.Source,
			Target:     ip,
			Protocol:   "tcp",
			Port:       parseInt(ev.Data["port"]),
			URL:        ev.Data["url"],
			Method:     ev.Data["method"],
			Error:      firstNonEmpty(ev.Data["error"], "http probe failed"),
			Timeout:    strings.Contains(strings.ToLower(ev.Data["error"]), "timeout") || strings.EqualFold(ev.Data["status"], "timeout"),
			EvidenceID: ev.ID,
		})
		return
	}
	store.AddHTTPObservation(ip, models.HTTPObservation{
		Source:           ev.Source,
		URL:              ev.Data["url"],
		Method:           firstNonEmpty(ev.Data["method"], "HEAD"),
		StatusCode:       status,
		Title:            ev.Data["title"],
		ServerHeader:     ev.Data["server"],
		WWWAuthenticate:  ev.Data["www_authenticate"],
		WWWAuthRealm:     ev.Data["realm"],
		RedirectLocation: ev.Data["location"],
		FaviconHash:      ev.Data["favicon_hash"],
		EvidenceID:       ev.ID,
	}, true, models.ProbeAttempt{})
}

func ingestTLSEvidence(ip string, ev models.Evidence, store *EvidenceStore) {
	store.AddTLSObservation(ip, models.TLSObservation{
		Source:     ev.Source,
		IP:         ip,
		Port:       parseInt(ev.Data["port"]),
		CN:         ev.Data["cn"],
		SANs:       splitList(ev.Data["sans"]),
		Issuer:     ev.Data["issuer"],
		ServerName: ev.Data["server_name"],
		EvidenceID: ev.ID,
	})
}

func addPrinterEvidence(ip string, ev models.Evidence, store *EvidenceStore) {
	if port := parseInt(ev.Data["port"]); port > 0 {
		store.AddService(ip, models.DeviceIntelService{Port: port, Protocol: "tcp", State: "open", Name: serviceName(port), Confidence: 0.75, EvidenceIDs: []string{ev.ID}})
	}
	d := store.UpsertDevice(ip)
	if d.PrinterInfo == nil {
		d.PrinterInfo = &models.PrinterInfo{Detected: true}
	}
	d.PrinterInfo.Protocols = appendUnique(d.PrinterInfo.Protocols, ev.Data["protocol"])
	d.PrinterInfo.Manufacturer = ev.Data["manufacturer"]
	d.PrinterInfo.Model = ev.Data["model"]
	d.PrinterInfo.EvidenceIDs = sortedUnique(append(d.PrinterInfo.EvidenceIDs, ev.ID))
}

func evidenceIndex(evidence []models.Evidence) map[string]models.Evidence {
	out := make(map[string]models.Evidence, len(evidence))
	for _, ev := range evidence {
		out[ev.ID] = ev
	}
	return out
}

func registerEvidenceIDs(store *EvidenceStore, evidenceByID map[string]models.Evidence, deviceID string, ids []string, confidence float64) {
	for _, id := range ids {
		if ev, ok := evidenceByID[id]; ok {
			store.RegisterEvidence(ev, deviceID, confidence)
		}
	}
}

func collectConflicts(report models.ScanReport, store *EvidenceStore) []models.DataConflict {
	var out []models.DataConflict
	out = append(out, store.Conflicts...)
	if report.AccessClassification != nil {
		out = append(out, report.AccessClassification.Conflicts...)
		out = append(out, report.AccessClassification.DataQuality.Conflicts...)
		if ctx := report.AccessClassification.DetectedNetworkContext; ctx != nil && ctx.GatewayChainState != nil {
			out = append(out, ctx.GatewayChainState.Conflicts...)
		}
	}
	return out
}

func summarize(report models.DeviceIntelReport) models.DeviceIntelSummary {
	var s models.DeviceIntelSummary
	s.DeviceCount = len(report.Devices)
	for _, d := range report.Devices {
		if d.Topology.IsGateway || containsString(d.Roles, models.RoleGateway) || containsString(d.Roles, models.DeviceRoleUpstreamGateway) {
			s.GatewayCount++
		}
		if d.DeviceType.Primary == models.DeviceTypeUnknown {
			s.UnknownDeviceCount++
		}
		s.ServiceCount += len(d.Services)
		if d.Confidence >= 0.75 {
			s.HighConfidenceDevices++
		}
		s.SecurityFindingCount += len(d.SecurityPosture.Findings)
	}
	for _, e := range report.Edges {
		if e.Physical {
			s.PhysicalEdges++
		}
		if e.Inferred {
			s.InferredEdges++
		}
	}
	return s
}

func buildUI(report models.DeviceIntelReport) models.DeviceIntelUI {
	ui := models.DeviceIntelUI{
		Headline: fmt.Sprintf("%d private/local devices observed", report.Summary.DeviceCount),
		Badges:   []string{"read-only", "private-scope", "no-login", "no-bruteforce"},
		TopologyNotes: []string{
			"Same-subnet and default-route links are inferred unless LLDP/CDP/SNMP bridge evidence is present.",
			"Upstream private gateways are path candidates from traceroute/private-hop evidence, not physical WAN proof.",
		},
	}
	for _, d := range report.Devices {
		card := models.DeviceCard{
			DeviceID:         d.ID,
			Title:            deviceTitle(d),
			Role:             strings.Join(d.Roles, " / "),
			Confidence:       d.Confidence,
			Hostnames:        d.Hostnames,
			MobileOSHint:     d.OSHint,
			MobileConfidence: d.OSConfidence,
			OpenServices:     serviceLabels(d.Services),
			DeviceType:       d.DeviceType.Primary,
			OSGuess:          d.OSGuess.Family,
			LastSeen:         d.LastSeen,
			EvidenceSources:  evidenceSources(report.Evidence, d.EvidenceIDs),
			RiskNotes:        riskNotes(d.SecurityPosture.Findings),
			Explanation:      d.ClassificationExplanation,
		}
		if d.Vendor.OUIVendor != nil {
			card.MACVendor = *d.Vendor.OUIVendor
		}
		if d.MobileFingerprint != nil {
			card.MobileLabel = mobileCardLabel(d.MobileFingerprint.Classification)
		}
		ui.DeviceCards = append(ui.DeviceCards, card)
	}
	return ui
}

func mobileCardLabel(classification string) string {
	switch classification {
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
		return "Conflicting mobile OS evidence"
	case models.MobileClassificationUnknownMobile:
		return "Unknown mobile"
	default:
		return ""
	}
}

func deviceTitle(d models.DeviceIntelDevice) string {
	if len(d.Hostnames) > 0 {
		return d.Hostnames[0]
	}
	if len(d.IPAddresses) > 0 {
		return d.IPAddresses[0]
	}
	return d.ID
}

func serviceLabels(services []models.DeviceIntelService) []string {
	out := make([]string, 0, len(services))
	for _, svc := range services {
		out = append(out, serviceLabel(svc))
	}
	sort.Strings(out)
	return out
}

func evidenceSources(evidence []models.DeviceIntelEvidence, ids []string) []string {
	wanted := map[string]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	var out []string
	for _, ev := range evidence {
		if wanted[ev.ID] {
			out = appendUnique(out, ev.SourceProbe)
		}
	}
	return sortedUnique(out)
}

func riskNotes(findings []models.SecurityFinding) []string {
	out := make([]string, 0, len(findings))
	for _, f := range findings {
		out = append(out, f.Title)
	}
	sort.Strings(out)
	return out
}

func evidenceIP(ev models.Evidence) string {
	for _, key := range []string{"ip", "target", "gateway", "host"} {
		if v := strings.TrimSpace(ev.Data[key]); v != "" {
			return v
		}
	}
	return ""
}

func evidenceConfidence(ev models.Evidence) float64 {
	switch ev.Kind {
	case "gateway_route":
		return 0.85
	case "arp_table":
		return 0.65
	case "tcp_connect", "nmap":
		return 0.75
	case "reverse_dns":
		return 0.45
	case "lldp", "cdp":
		return 0.95
	case "snmp_bridge":
		return 0.80
	default:
		return 0.50
	}
}

func targetAllowed(scope models.DeviceIntelScope, ip string) bool {
	if safety.IsPrivateIPString(ip) {
		return true
	}
	return scope.PublicAllowed && !scope.PrivateOnly
}

func devicePrimaryIP(d models.Device) string {
	ips := deviceIPs(d)
	if len(ips) == 0 {
		return ""
	}
	sort.Slice(ips, func(i, j int) bool { return ipLess(ips[i], ips[j]) })
	return ips[0]
}

func deviceIPs(d models.Device) []string {
	var ips []string
	for _, a := range d.Addresses {
		if a.IP != "" {
			ips = append(ips, a.IP)
		}
	}
	for _, iface := range d.Interfaces {
		ips = append(ips, iface.IPs...)
	}
	return sortedUnique(ips)
}

func selectedAgentIP(ifaces []models.InterfaceInfo) string {
	for _, iface := range ifaces {
		if !iface.Selected {
			continue
		}
		for _, addr := range iface.Addresses {
			if addr.Version == 4 && addr.IP != "" {
				return addr.IP
			}
		}
	}
	return ""
}

func firstEvidence(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

func parsePorts(raw string) []int {
	raw = strings.ReplaceAll(raw, ";", ",")
	raw = strings.ReplaceAll(raw, " ", ",")
	parts := strings.Split(raw, ",")
	var ports []int
	for _, part := range parts {
		p := parseInt(part)
		if p > 0 {
			ports = append(ports, p)
		}
	}
	return sortedUniqueInts(ports)
}

func parseInt(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, _ := strconv.Atoi(raw)
	return n
}

func splitList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	raw = strings.ReplaceAll(raw, ";", ",")
	parts := strings.Split(raw, ",")
	return sortedUnique(parts)
}
