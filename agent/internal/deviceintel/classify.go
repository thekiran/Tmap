package deviceintel

import (
	"fmt"
	"sort"
	"strings"

	"github.com/thekiran/iad/internal/discovery"
	"github.com/thekiran/iad/pkg/models"
)

func classifyAll(store *EvidenceStore) {
	classifyAllWithEvidence(store, nil)
}

func classifyAllWithEvidence(store *EvidenceStore, evidence []models.Evidence) {
	for _, d := range store.Devices {
		classifyDevice(d, evidence)
	}
}

func classifyDevice(d *models.DeviceIntelDevice, evidence []models.Evidence) {
	addServiceRoles(d)
	addProtocolInfo(d)
	applyMobileFingerprint(d, evidence)
	addDeviceTypeCandidates(d)
	pickPrimaryType(d)
	guessOS(d)
	addSecurityFindings(d)
	finalizeConfidence(d)
}

func applyMobileFingerprint(d *models.DeviceIntelDevice, evidence []models.Evidence) {
	engine := discovery.NewMobileDeviceFingerprintEngine(nil)
	fp := engine.FingerprintDeviceIntel(*d, evidence)
	d.MobileFingerprint = &fp
	d.OSHint = discovery.MobileOSHint(fp.Classification)
	d.OSConfidence = fp.Confidence
	d.DeviceTypeHint = discovery.MobileDeviceTypeHint(fp, d.Hostnames)
	if d.OSHint != models.MobileOSHintUnknown {
		d.ClassificationExplanation = appendUnique(d.ClassificationExplanation, fp.WhyThisClassification)
		d.UndeterminedWithoutOptIn = appendUnique(d.UndeterminedWithoutOptIn, fp.WhyNotCertain)
	}
}

func addServiceRoles(d *models.DeviceIntelDevice) {
	if d.Topology.IsGateway {
		d.Roles = appendUnique(d.Roles, models.RoleGateway, models.RoleRouter)
	}
	if d.Topology.IsAgent {
		d.Roles = appendUnique(d.Roles, models.RoleAgent)
	}
	if d.Topology.IsUpstreamGatewayCandidate {
		d.Roles = appendUnique(d.Roles, models.DeviceRoleUpstreamGateway, models.DeviceRolePossibleCPE)
	}
	if hasOpenPort(d, 53) {
		d.Roles = appendUnique(d.Roles, models.DeviceRoleNameResolution)
	}
	if hasAnyPort(d, 80, 443, 8080, 8443, 7547) && (d.Topology.IsGateway || d.Topology.IsUpstreamGatewayCandidate) {
		d.Roles = appendUnique(d.Roles, models.DeviceRoleManagement)
	}
	if hasAnyPort(d, 515, 631, 9100) {
		d.Roles = appendUnique(d.Roles, models.DeviceRolePrinter)
	}
	if hasAnyPort(d, 5000, 5001) {
		d.Roles = appendUnique(d.Roles, models.DeviceRoleStorage)
	}
	if hasAnyPort(d, 32400, 8008, 8009, 554) {
		d.Roles = appendUnique(d.Roles, models.DeviceRoleMedia)
	}
	if len(d.Roles) == 0 {
		d.Roles = appendUnique(d.Roles, models.RoleHost)
	}
}

func addProtocolInfo(d *models.DeviceIntelDevice) {
	if hasAnyPort(d, 445, 139) || len(d.NBNSRecords) > 0 {
		info := &models.SMBInfo{OSFamily: "unknown", EvidenceIDs: evidenceIDsForPorts(d, 445, 139)}
		if len(d.NBNSRecords) > 0 {
			info.NetBIOSName = d.NBNSRecords[0].Name
			info.Workgroup = d.NBNSRecords[0].Workgroup
			info.EvidenceIDs = sortedUnique(append(info.EvidenceIDs, d.NBNSRecords[0].EvidenceIDs...))
		}
		d.SMBInfo = info
	}

	protocols := printerProtocols(d)
	if len(protocols) > 0 {
		d.PrinterInfo = &models.PrinterInfo{Detected: true, Protocols: protocols, EvidenceIDs: printerEvidence(d)}
	}

	if hasOpenPort(d, 1900) || ssdpHas(d, "InternetGatewayDevice") || ssdpHas(d, "MediaRenderer") {
		info := &models.UPnPInfo{Detected: true, EvidenceIDs: ssdpEvidence(d)}
		for _, rec := range d.SSDPRecords {
			if rec.ST != "" {
				info.DeviceType = rec.ST
				break
			}
		}
		if ssdpHas(d, "InternetGatewayDevice") {
			info.IGD = true
		}
		d.UPnPInfo = info
	}
}

func addDeviceTypeCandidates(d *models.DeviceIntelDevice) {
	add := func(c models.DeviceTypeCandidate) {
		d.DeviceType.Candidates = mergeCandidate(d.DeviceType.Candidates, c)
	}

	if d.Topology.IsGateway && (hasOpenPort(d, 53) || hasAnyPort(d, 80, 443, 8080, 8443)) {
		add(models.DeviceTypeCandidate{
			Type:            models.DeviceTypeGatewayRouter,
			Confidence:      0.82,
			SupportingFacts: []string{"Default route points to this device and router-like services are open."},
			MissingEvidence: []string{"No authenticated router telemetry or physical WAN interface evidence was collected."},
			EvidenceIDs:     d.EvidenceIDs,
		})
	} else if d.Topology.IsGateway {
		add(models.DeviceTypeCandidate{
			Type:            models.DeviceTypeGatewayRouter,
			Confidence:      0.70,
			SupportingFacts: []string{"Default route points to this device."},
			MissingEvidence: []string{"Router services, UPnP IGD, LLDP/CDP, or SNMP bridge data were not observed."},
			EvidenceIDs:     d.EvidenceIDs,
		})
	}

	if d.Topology.IsUpstreamGatewayCandidate {
		add(models.DeviceTypeCandidate{
			Type:            models.DeviceTypeUpstreamCPE,
			Confidence:      0.60,
			SupportingFacts: []string{"A later RFC1918 hop was observed after the default gateway."},
			MissingEvidence: []string{"No direct CPE login, SNMP, TR-064/TR-181, UPnP WAN, DSL, DOCSIS, or optical evidence was collected."},
			EvidenceIDs:     d.EvidenceIDs,
		})
	}

	if hasAnyPort(d, 631, 9100, 515) || mdnsHas(d, "_ipp._tcp") || mdnsHas(d, "_printer") {
		conf := 0.65
		fact := "Printer-related TCP port is open."
		if mdnsHas(d, "_ipp._tcp") || mdnsHas(d, "_printer") {
			conf = 0.78
			fact = "mDNS advertised a printer or IPP service."
		}
		add(models.DeviceTypeCandidate{
			Type:            models.DeviceTypePrinter,
			Confidence:      conf,
			SupportingFacts: []string{fact},
			MissingEvidence: []string{"Printer model/manufacturer were not read unless the device exposed them unauthenticated."},
			EvidenceIDs:     printerEvidence(d),
		})
	}

	if hasOpenPort(d, 445) {
		conf := 0.60
		fact := "SMB TCP/445 is open."
		if hasAnyPort(d, 135, 139, 5357) || len(d.NBNSRecords) > 0 {
			conf = 0.74
			fact = "SMB/NetBIOS/WSD Windows-style services are open."
		}
		add(models.DeviceTypeCandidate{
			Type:            models.DeviceTypeWindowsHost,
			Confidence:      conf,
			SupportingFacts: []string{fact},
			MissingEvidence: []string{"Authenticated SMB details and shares were not enumerated."},
			EvidenceIDs:     evidenceIDsForPorts(d, 135, 139, 445, 5357),
		})
	}

	if hasAnyPort(d, 5000, 5001) {
		add(models.DeviceTypeCandidate{
			Type:            models.DeviceTypeNAS,
			Confidence:      0.62,
			SupportingFacts: []string{"NAS/media-management ports 5000/5001 are open."},
			MissingEvidence: []string{"Vendor-specific web UI or mDNS/SSDP storage identity was not proven."},
			EvidenceIDs:     evidenceIDsForPorts(d, 5000, 5001),
		})
	}

	if hasOpenPort(d, 32400) || hasAnyPort(d, 8008, 8009) || ssdpHas(d, "MediaRenderer") {
		add(models.DeviceTypeCandidate{
			Type:            models.DeviceTypeMediaDevice,
			Confidence:      0.68,
			SupportingFacts: []string{"Media service or media renderer signal was observed."},
			MissingEvidence: []string{"Device brand/model was not proven by a specific mDNS/SSDP/HTTP fingerprint."},
			EvidenceIDs:     sortedUnique(append(evidenceIDsForPorts(d, 32400, 8008, 8009), ssdpEvidence(d)...)),
		})
	}

	if hasOpenPort(d, 554) {
		add(models.DeviceTypeCandidate{
			Type:            models.DeviceTypeIPCamera,
			Confidence:      0.55,
			SupportingFacts: []string{"RTSP TCP/554 is open."},
			MissingEvidence: []string{"Camera brand, ONVIF, or video metadata was not queried."},
			EvidenceIDs:     evidenceIDsForPorts(d, 554),
		})
	}

	if hasOpenPort(d, 1883) {
		add(models.DeviceTypeCandidate{
			Type:            models.DeviceTypeIoT,
			Confidence:      0.58,
			SupportingFacts: []string{"MQTT TCP/1883 is open."},
			MissingEvidence: []string{"MQTT authentication, topics, and payloads were not accessed."},
			EvidenceIDs:     evidenceIDsForPorts(d, 1883),
		})
	}

	if hasOpenPort(d, 8123) {
		add(models.DeviceTypeCandidate{
			Type:            "home_automation_server",
			Confidence:      0.72,
			SupportingFacts: []string{"Home Assistant default TCP/8123 is open."},
			MissingEvidence: []string{"No login or API call was attempted."},
			EvidenceIDs:     evidenceIDsForPorts(d, 8123),
		})
	}

	if d.MobileFingerprint != nil && d.OSHint != models.MobileOSHintUnknown {
		typ := models.DeviceTypeAndroidDevice
		if d.OSHint == models.MobileOSHintIOS || d.OSHint == models.MobileOSHintIPadOS {
			typ = models.DeviceTypeAppleDevice
		}
		conf := d.MobileFingerprint.Confidence
		if conf < 0.45 {
			conf = 0.45
		}
		add(models.DeviceTypeCandidate{
			Type:            typ,
			Confidence:      conf,
			SupportingFacts: []string{d.MobileFingerprint.WhyThisClassification},
			MissingEvidence: []string{d.MobileFingerprint.WhyNotCertain},
			EvidenceIDs:     d.EvidenceIDs,
		})
	}

	if hasOpenPort(d, 22) && hasAnyPort(d, 80, 443, 8080, 8443) && !d.Topology.IsGateway {
		add(models.DeviceTypeCandidate{
			Type:            models.DeviceTypeServer,
			Confidence:      0.52,
			SupportingFacts: []string{"SSH and web services are open."},
			MissingEvidence: []string{"No authenticated service banner, package, or OS data was collected."},
			EvidenceIDs:     evidenceIDsForPorts(d, 22, 80, 443, 8080, 8443),
		})
	}

	for _, obs := range d.HTTPFingerprints {
		if strings.EqualFold(obs.ServerHeader, "nginx") || strings.HasPrefix(strings.ToLower(obs.ServerHeader), "nginx") {
			d.ClassificationExplanation = appendUnique(d.ClassificationExplanation, "Generic nginx Server header is web-server evidence only; it does not identify router model, WAN type, or physical access.")
			d.SecurityPosture.Notes = appendUnique(d.SecurityPosture.Notes, "Generic nginx does not prove device model.")
		}
	}
}

func pickPrimaryType(d *models.DeviceIntelDevice) {
	sort.SliceStable(d.DeviceType.Candidates, func(i, j int) bool {
		if d.DeviceType.Candidates[i].Confidence != d.DeviceType.Candidates[j].Confidence {
			return d.DeviceType.Candidates[i].Confidence > d.DeviceType.Candidates[j].Confidence
		}
		return d.DeviceType.Candidates[i].Type < d.DeviceType.Candidates[j].Type
	})
	if len(d.DeviceType.Candidates) == 0 {
		d.DeviceType.Primary = models.DeviceTypeUnknown
		d.DeviceType.Confidence = 0
		d.DeviceType.MissingEvidence = []string{"Only IP/MAC or generic reachability was observed; no service, OS, model, or protocol identity evidence was conclusive."}
		d.DeviceType.Alternatives = nil
		d.Roles = appendUnique(d.Roles, models.DeviceRoleUnknownHost)
		return
	}
	top := d.DeviceType.Candidates[0]
	if top.Confidence < 0.50 {
		d.DeviceType.Primary = models.DeviceTypeUnknown
		d.DeviceType.Confidence = top.Confidence
		d.DeviceType.MissingEvidence = appendUnique(top.MissingEvidence, "The strongest candidate is below the display threshold.")
		d.DeviceType.Alternatives = append([]models.DeviceTypeCandidate(nil), d.DeviceType.Candidates...)
		return
	}
	d.DeviceType.Primary = top.Type
	d.DeviceType.Confidence = top.Confidence
	d.DeviceType.EvidenceIDs = sortedUnique(top.EvidenceIDs)
	d.DeviceType.MissingEvidence = sortedUnique(top.MissingEvidence)
	if len(d.DeviceType.Candidates) > 1 {
		d.DeviceType.Alternatives = append([]models.DeviceTypeCandidate(nil), d.DeviceType.Candidates[1:]...)
	}
	d.ClassificationExplanation = appendUnique(d.ClassificationExplanation, top.SupportingFacts...)
}

func guessOS(d *models.DeviceIntelDevice) {
	if d.MobileFingerprint != nil && d.OSHint != models.MobileOSHintUnknown {
		name := displayMobileOSName(d.OSHint)
		d.OSGuess = models.OSGuess{
			Family:      d.OSHint,
			Name:        ptrString(name),
			Confidence:  d.MobileFingerprint.Confidence,
			Evidence:    mobileEvidenceExplanations(d.MobileFingerprint.Evidence),
			EvidenceIDs: d.EvidenceIDs,
		}
		return
	}
	if hasOpenPort(d, 445) && (hasAnyPort(d, 135, 139, 5357) || len(d.NBNSRecords) > 0) {
		d.OSGuess = models.OSGuess{
			Family:      "windows",
			Name:        ptrString("Windows candidate"),
			Confidence:  0.72,
			Evidence:    []string{"SMB/NetBIOS/WSD Windows-style services were observed."},
			EvidenceIDs: evidenceIDsForPorts(d, 135, 139, 445, 5357),
		}
		return
	}
	d.OSGuess = models.OSGuess{
		Family:      "unknown",
		Confidence:  0,
		Evidence:    []string{"No safe OS-specific protocol evidence was conclusive."},
		EvidenceIDs: nil,
	}
}

func displayMobileOSName(osHint string) string {
	switch osHint {
	case models.MobileOSHintIOS:
		return "iOS candidate"
	case models.MobileOSHintIPadOS:
		return "iPadOS candidate"
	case models.MobileOSHintAndroid:
		return "Android candidate"
	default:
		return "Mobile OS candidate"
	}
}

func mobileEvidenceExplanations(items []models.MobileEvidenceItem) []string {
	var out []string
	for _, item := range items {
		out = appendUnique(out, item.Explanation)
	}
	return out
}

func addSecurityFindings(d *models.DeviceIntelDevice) {
	var findings []models.SecurityFinding
	add := func(id, severity, title, desc, rec string, evIDs []string) {
		findings = append(findings, models.SecurityFinding{
			ID:                 id,
			Severity:           severity,
			Title:              title,
			Description:        desc,
			SafeRecommendation: rec,
			EvidenceIDs:        sortedUnique(evIDs),
		})
	}

	if hasOpenPort(d, 23) {
		add("open_telnet", "high", "Telnet port open", "TCP/23 is open on this LAN host. The collector only detected the port and did not attempt login.", "Disable Telnet if it is not required; prefer SSH with explicit authorization.", evidenceIDsForPorts(d, 23))
	}
	if hasOpenPort(d, 445) {
		add("open_smb", "medium", "SMB port open", "TCP/445 is open on this LAN host.", "Keep file sharing disabled if not needed and restrict it to trusted LAN profiles.", evidenceIDsForPorts(d, 445))
	}
	if hasAnyPort(d, 80, 8080) && (d.Topology.IsGateway || d.Topology.IsUpstreamGatewayCandidate) {
		add("open_http_management", "medium", "HTTP management interface possible", "A router/CPE candidate exposes an unencrypted HTTP service.", "Use HTTPS for management where possible and keep router firmware current.", evidenceIDsForPorts(d, 80, 8080))
	}
	if hasOpenPort(d, 9100) {
		add("printer_raw_9100", "medium", "Raw printer port open", "TCP/9100 is open. The collector did not send print jobs.", "Disable raw printing if it is not needed or restrict it to trusted hosts.", evidenceIDsForPorts(d, 9100))
	}
	if d.UPnPInfo != nil && d.UPnPInfo.Detected {
		add("upnp_exposed_lan", "low", "UPnP visible on LAN", "UPnP/SSDP information is exposed to local devices.", "Disable UPnP if you do not need automatic port mapping.", d.UPnPInfo.EvidenceIDs)
	}
	if d.DeviceType.Primary == models.DeviceTypeUnknown {
		add("unknown_device", "low", "Unknown device on LAN", "The device was observed but did not expose enough safe identity evidence for classification.", "Confirm the device belongs on this network.", d.EvidenceIDs)
	}
	if d.Topology.IsAgent && len(d.Services) >= 5 {
		add("many_services_on_agent", "info", "Many services open on agent host", fmt.Sprintf("%d open services were observed on the scanning host.", len(d.Services)), "Review local services and disable unused listeners.", d.EvidenceIDs)
	}

	d.SecurityPosture.Findings = mergeFindings(d.SecurityPosture.Findings, findings)
	d.SecurityPosture.RiskLevel = riskLevel(d.SecurityPosture.Findings)
}

func finalizeConfidence(d *models.DeviceIntelDevice) {
	conf := d.Confidence
	if d.Topology.IsAgent {
		conf = maxFloat(conf, 0.95)
	}
	if d.Topology.IsGateway {
		conf = maxFloat(conf, 0.90)
	}
	if d.Topology.IsUpstreamGatewayCandidate {
		conf = maxFloat(conf, 0.60)
	}
	if len(d.MACAddresses) > 0 {
		conf = maxFloat(conf, 0.55)
	}
	if len(d.Services) > 0 || len(d.HTTPFingerprints) > 0 || len(d.TLSFingerprints) > 0 {
		conf = maxFloat(conf, 0.75)
	}
	if d.DeviceType.Primary == models.DeviceTypeUnknown && len(d.Services) == 0 && !d.Topology.IsGateway && !d.Topology.IsUpstreamGatewayCandidate {
		conf = minFloat(maxFloat(conf, 0.45), 0.60)
	}
	d.Confidence = clamp01(conf)
	if len(d.ClassificationExplanation) == 0 {
		d.ClassificationExplanation = []string{"Insufficient strong device-specific evidence; classification remains conservative."}
	}
	if len(d.UndeterminedWithoutOptIn) == 0 && (d.Topology.IsGateway || d.Topology.IsUpstreamGatewayCandidate) {
		d.UndeterminedWithoutOptIn = []string{
			"WAN access type cannot be known without direct physical-layer evidence or opt-in CPE telemetry.",
			"Router model cannot be trusted from generic HTTP server headers alone.",
		}
	}
}

func printerProtocols(d *models.DeviceIntelDevice) []string {
	var out []string
	if hasOpenPort(d, 631) {
		out = append(out, "ipp")
	}
	if hasOpenPort(d, 9100) {
		out = append(out, "jetdirect")
	}
	if hasOpenPort(d, 515) {
		out = append(out, "lpd")
	}
	if mdnsHas(d, "_ipp._tcp") {
		out = append(out, "mdns_ipp")
	}
	return sortedUnique(out)
}

func printerEvidence(d *models.DeviceIntelDevice) []string {
	ids := evidenceIDsForPorts(d, 631, 9100, 515)
	for _, rec := range d.MDNSRecords {
		if strings.Contains(strings.ToLower(rec.Service), "ipp") || strings.Contains(strings.ToLower(rec.Name), "printer") {
			ids = append(ids, rec.EvidenceIDs...)
		}
	}
	return sortedUnique(ids)
}

func mdnsHas(d *models.DeviceIntelDevice, needle string) bool {
	needle = strings.ToLower(needle)
	for _, rec := range d.MDNSRecords {
		if strings.Contains(strings.ToLower(rec.Service), needle) || strings.Contains(strings.ToLower(rec.Name), needle) {
			return true
		}
	}
	return false
}

func ssdpHas(d *models.DeviceIntelDevice, needle string) bool {
	needle = strings.ToLower(needle)
	for _, rec := range d.SSDPRecords {
		if strings.Contains(strings.ToLower(rec.ST), needle) || strings.Contains(strings.ToLower(rec.USN), needle) || strings.Contains(strings.ToLower(rec.Server), needle) {
			return true
		}
	}
	return false
}

func ssdpEvidence(d *models.DeviceIntelDevice) []string {
	var ids []string
	for _, rec := range d.SSDPRecords {
		ids = append(ids, rec.EvidenceIDs...)
	}
	if hasOpenPort(d, 1900) {
		ids = append(ids, evidenceIDsForPorts(d, 1900)...)
	}
	return sortedUnique(ids)
}

func mergeFindings(existing, next []models.SecurityFinding) []models.SecurityFinding {
	byID := map[string]int{}
	out := append([]models.SecurityFinding(nil), existing...)
	for i, f := range out {
		byID[f.ID] = i
	}
	for _, f := range next {
		if idx, ok := byID[f.ID]; ok {
			out[idx].EvidenceIDs = sortedUnique(append(out[idx].EvidenceIDs, f.EvidenceIDs...))
			continue
		}
		byID[f.ID] = len(out)
		out = append(out, f)
	}
	return out
}

func riskLevel(findings []models.SecurityFinding) string {
	if len(findings) == 0 {
		return "unknown"
	}
	level := "low"
	for _, f := range findings {
		switch f.Severity {
		case "high":
			return "high"
		case "medium":
			level = "medium"
		}
	}
	return level
}
