package modemcollector

import (
	"sort"

	"github.com/thekiran/iad/pkg/models"
)

func Build(in BuildInput) models.ModemCollection {
	r := in.Result
	nc := r.DetectedNetworkContext
	collection := models.ModemCollection{
		Status:   "completed",
		SafeMode: true,
		Scope: models.ModemCollectionScope{
			PrivateOnly: true,
		},
		AccessClassification: models.ModemAccessClassification{
			PrimaryType:             r.Classification.PrimaryType,
			Subtype:                 r.Classification.Subtype,
			Confidence:              r.Classification.Confidence,
			DecisionQuality:         r.Classification.DecisionQuality,
			SafeToDisplayAsFinal:    r.Classification.SafeToDisplayAsFinal,
			Candidates:              r.Candidates,
			MissingRequiredEvidence: missingRequiredEvidence(r),
			Reason:                  append([]string{}, r.Explanation...),
		},
		DataQuality: r.DataQuality,
		UI:          modemUI(r),
		SecurityNotes: []string{
			"Private/local targets only by default.",
			"Read-only probes only; no form submission, login attempt, credential guessing, exploitation, or public range scanning.",
			"SNMP and authenticated TR-064/TR-181 collection require explicit user-provided read-only credentials.",
		},
		Undetermined: []string{
			"Physical WAN layer cannot be proven without direct CPE telemetry such as TR-064/TR-181/UPnP WANAccessType/SNMP interface data or a strong model fingerprint.",
			"Ethernet LAN, low latency, public IP, no CGNAT, PTR, ASN, and traceroute context do not prove DSL/VDSL/Fiber/GPON/LTE/WISP.",
			"Authenticated CPE details such as line stats, optical levels, full TR-181 InterfaceStack, and some model fields remain unavailable without user credentials.",
		},
	}
	if nc == nil {
		collection.Status = "partial"
		return collection
	}

	chain := nc.GatewayChainState
	if chain != nil {
		collection.NormalizedGatewayChain = models.NormalizedGatewayChain{
			Hops:                      append([]models.GatewayHop{}, chain.PrivateHops...),
			InternalDoubleNATPossible: triFromBool(chain.InternalDoubleNATPossible),
			EvidenceSources:           chainSources(chain),
		}
	} else {
		collection.NormalizedGatewayChain.InternalDoubleNATPossible = models.TriUnknown
	}
	collection.NAT = natState(nc)

	builder := CandidateBuilder{}
	devices := builder.Build(CandidateInput{
		DefaultGateway: nc.Gateway,
		ChainState:     chain,
		Devices:        nc.GatewayDevices,
	})
	for _, d := range devices {
		c := candidateFromDevice(d)
		collection.CPECandidates = append(collection.CPECandidates, c)
		collection.Scope.Targets = appendUniqueStrings(collection.Scope.Targets, c.IP)
	}
	sort.SliceStable(collection.CPECandidates, func(i, j int) bool {
		if priorityRank(collection.CPECandidates[i].Priority) != priorityRank(collection.CPECandidates[j].Priority) {
			return priorityRank(collection.CPECandidates[i].Priority) > priorityRank(collection.CPECandidates[j].Priority)
		}
		return rolePriority(collection.CPECandidates[i].Role) > rolePriority(collection.CPECandidates[j].Role)
	})
	return collection
}

func natState(nc *models.NetworkContext) models.ModemNATState {
	out := models.ModemNATState{
		CGNAT: models.TriUnknown, DoubleNAT: models.TriUnknown, InternalDoubleNATPossible: models.TriUnknown,
		PublicIPMatches: models.TriUnknown, ExternalPublicIPConsistent: models.TriUnknown,
		PCPReachable: models.TriUnknown, NATPMPReachable: models.TriUnknown,
	}
	if nc == nil {
		return out
	}
	if nc.CGNAT {
		out.CGNAT = models.TriTrue
	}
	if nc.DoubleNATPossible {
		out.DoubleNAT = models.TriTrue
		out.InternalDoubleNATPossible = models.TriTrue
	}
	if nc.NATTopology == nil {
		return out
	}
	nat := nc.NATTopology
	if nat.CGNAT {
		out.CGNAT = models.TriTrue
	}
	if nat.DoubleNAT {
		out.DoubleNAT = models.TriTrue
	}
	if nat.InternalDoubleNATPossible {
		out.InternalDoubleNATPossible = models.TriTrue
	}
	if nat.PublicIP != "" && nat.STUNPublicIP != "" {
		out.PublicIPMatches = triFromBool(nat.PublicIPMatches)
		out.ExternalPublicIPConsistent = triFromBool(nat.ExternalPublicIPConsistent)
	}
	if nat.PCPReachable {
		out.PCPReachable = models.TriTrue
	}
	if nat.NATPMPReachable {
		out.NATPMPReachable = models.TriTrue
	}
	return out
}

func candidateFromDevice(d models.GatewayDevice) models.CPECandidate {
	httpReachable := models.TriUnknown
	if len(d.HTTPObservations) > 0 || d.HTTPTitle != "" || d.ServerHeader != "" || d.FaviconHash != "" || d.WWWAuthenticate != "" {
		httpReachable = models.TriTrue
	}
	tlsReachable := models.TriUnknown
	if len(d.TLSObservations) > 0 || d.TLSCertCN != "" || len(d.TLSCertSANs) > 0 || d.TLSCertIssuer != "" {
		tlsReachable = models.TriTrue
	}
	modelConf := 0.0
	if d.Manufacturer != "" || d.Model != "" || d.CPEModelGuess != "" || d.FingerprintID != "" {
		modelConf = d.DeviceConfidence
	}
	c := models.CPECandidate{
		IP:             d.IP,
		Role:           d.Role,
		Source:         append([]string{}, d.EvidenceIDs...),
		Priority:       priorityForDevice(d),
		Private:        isRFC1918(d.IP),
		ReachableState: triFromString(d.ReachableState),
		OpenPorts:      append([]int{}, d.OpenPorts...),
		HTTP: models.CPEHTTPState{
			Reachable:    httpReachable,
			Observations: append([]models.HTTPObservation{}, d.HTTPObservations...),
		},
		TLS: models.CPETLSState{
			Reachable:    tlsReachable,
			Certificates: append([]models.TLSObservation{}, d.TLSObservations...),
		},
		UPnP: models.CPEUPnPState{
			Found:                      triFromBool(d.UPnPFound),
			IGDFound:                   triFromBool(d.UPnPIGDFound),
			WANCommonInterfaceFound:    triFromBool(d.WANCommonInterfaceFound),
			WANAccessType:              ptrString(d.WANAccessType),
			Layer1UpstreamMaxBitRate:   ptrInt64(d.Layer1UpstreamMaxBitRate),
			Layer1DownstreamMaxBitRate: ptrInt64(d.Layer1DownstreamMaxBitRate),
			PhysicalLinkStatus:         ptrString(d.PhysicalLinkStatus),
		},
		TR064: models.CPETR064State{
			Found:          triFromBool(d.TR064Found || d.TR064AuthRequired),
			AuthRequired:   triFromBool(d.TR064AuthRequired),
			DataAccessible: tr064DataAccessible(d),
			Services:       append([]string{}, d.TR064Services...),
		},
		TR181: models.CPETR181State{
			Available: models.TriUnknown,
		},
		SNMP: models.CPESNMPState{
			Enabled: false,
			Status:  "skipped",
			Reason:  "requires explicit user credentials",
		},
		ModelFingerprint: models.CPEModelFingerprint{
			Vendor:     ptrString(d.Manufacturer),
			Model:      ptrString(firstNonEmpty(d.Model, d.CPEModelGuess)),
			Confidence: modelConf,
		},
		WANPhysicalEvidence: physicalEvidence(d),
		FailedAttempts:      append([]models.ProbeAttempt{}, d.FailedAttempts...),
		Confidence:          d.Confidence,
		EvidenceIDs:         append([]string{}, d.EvidenceIDs...),
	}
	if c.ReachableState == models.TriUnknown && (httpReachable == models.TriTrue || tlsReachable == models.TriTrue || len(d.OpenPorts) > 0 || d.Reachable) {
		c.ReachableState = models.TriTrue
	}
	return c
}

func tr064DataAccessible(d models.GatewayDevice) models.TriState {
	if len(d.TR064Services) > 0 && !d.TR064AuthRequired {
		return models.TriTrue
	}
	if d.TR064AuthRequired {
		return models.TriFalse
	}
	return models.TriUnknown
}

func physicalEvidence(d models.GatewayDevice) models.CPEWANPhysicalEvidence {
	out := models.CPEWANPhysicalEvidence{Status: "missing"}
	if d.WANCommonInterfaceFound && d.WANAccessType != "" {
		out.Status = "present"
		out.Type = ptrString(d.WANAccessType)
		out.Confidence = max(out.Confidence, d.AccessConfidence)
		if out.Confidence == 0 {
			out.Confidence = d.Confidence
		}
		out.Source = ptrString("upnp_wan_common_interface")
		out.Evidence = appendUniqueStrings(out.Evidence, d.EvidenceIDs...)
		out.Evidence = appendUniqueStrings(out.Evidence, d.WANAccessType, d.PhysicalLinkStatus)
	}
	for _, ev := range d.AccessEvidence {
		if ev.Strength == string(models.EvidencePhysical) || ev.Strength == "strong" {
			out.Status = "present"
			out.Type = ptrString(firstNonEmpty(ev.Type, firstHint(ev.Hints)))
			out.Confidence = ev.Confidence
			out.Source = ptrString(ev.Source)
			out.Evidence = appendUniqueStrings(out.Evidence, ev.EvidenceID, ev.Value)
		}
	}
	if len(d.PhysicalHints) > 0 {
		out.Status = "present"
		out.Type = ptrString(firstHint(d.PhysicalHints))
		out.Confidence = max(out.Confidence, d.AccessConfidence)
		out.Source = ptrString("gateway_device")
		out.Evidence = appendUniqueStrings(out.Evidence, d.PhysicalHints...)
	}
	if out.Status == "missing" {
		out.Type = nil
		out.Subtype = nil
		out.Source = nil
	}
	return out
}

func missingRequiredEvidence(r models.ScanResult) []string {
	var out []string
	out = appendUniqueStrings(out, r.UncertaintyReasons...)
	if r.Classification.PrimaryType == "Unknown" {
		out = appendUniqueStrings(out,
			"No direct WAN physical-layer evidence.",
			"No TR-064/TR-181/SNMP/UPnP WAN physical-layer proof.",
			"Fiber and VDSL can be shown as candidates only when supported by weak scores.",
		)
	}
	return out
}

func modemUI(r models.ScanResult) models.UIOutput {
	ui := r.UI
	if ui.Headline == "" {
		ui.Headline = "Access type unknown"
	}
	if ui.Summary == "" && r.Classification.PrimaryType == "Unknown" {
		ui.Summary = "Erisim turu kesin belirlenemedi. Ozel gateway zinciri ve CPE adaylari bulundu, ancak DSL, VDSL veya Fiber oldugunu kanitlayan dogrudan WAN fiziksel katman verisi yok."
	}
	ui.Badges = appendUniqueStrings(ui.Badges,
		"Default gateway found",
		"SNMP skipped by default",
		"Classification not final",
	)
	if r.DetectedNetworkContext != nil && r.DetectedNetworkContext.DoubleNATPossible {
		ui.Badges = appendUniqueStrings(ui.Badges, "Upstream private gateway detected", "Possible double NAT")
	}
	if r.Classification.PrimaryType == "Unknown" {
		ui.Badges = appendUniqueStrings(ui.Badges, "No direct WAN evidence")
		ui.Warnings = appendUniqueStrings(ui.Warnings,
			"Do not treat Ethernet LAN as Fiber.",
			"Do not treat latency as Fiber.",
			"Do not treat PTR/ASN as access type proof.",
			"Physical WAN evidence is required.",
		)
	}
	return ui
}

func priorityForDevice(d models.GatewayDevice) string {
	switch d.Role {
	case "upstream_private_gateway", "possible_cpe", "possible_modem", "possible_modem_or_ont", "cpe_management_endpoint":
		return "high"
	case "default_gateway":
		return "medium"
	default:
		return "low"
	}
}

func priorityRank(p string) int {
	switch p {
	case "high":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}

func chainSources(chain *models.GatewayChainState) []string {
	var out []string
	for _, src := range chain.Sources {
		out = appendUniqueStrings(out, src.Source)
	}
	return out
}

func firstHint(hints []string) string {
	for _, h := range hints {
		if h != "" {
			return h
		}
	}
	return ""
}
