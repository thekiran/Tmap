package detection

import (
	"fmt"
	"strings"

	"github.com/thekiran/iad/internal/linestats"
	"github.com/thekiran/iad/pkg/models"
)

func buildEvidenceTiers(bag evidenceBag, matched bool) models.EvidenceTiers {
	var directItems []models.EvidenceItem
	if bag.LineProfile != nil {
		target := lineProfileTargetType(bag.LineProfile)
		directItems = append(directItems, models.EvidenceItem{
			Source:     firstNonEmpty(bag.LineProfile.Source, "line_profile"),
			TargetType: target,
			Strength:   "strong",
			Confidence: bag.LineProfile.Confidence,
			Reason:     strings.Join(linestats.Summary(bag.LineProfile), " "),
			Raw: map[string]any{
				"medium":     bag.LineProfile.Medium,
				"technology": bag.LineProfile.Technology,
				"subtype":    bag.LineProfile.Subtype,
			},
		})
	}
	for _, sig := range bag.WANSignals {
		if sig.Strength == string(models.EvidencePhysical) || strings.EqualFold(sig.Strength, "strong") || sig.Confidence >= 0.70 {
			directItems = append(directItems, models.EvidenceItem{
				Source:     firstNonEmpty(sig.Source, "wan_signal"),
				TargetType: inferTargetType(sig.Value + " " + sig.Detail),
				Strength:   "strong",
				Confidence: sig.Confidence,
				Reason:     strings.TrimSpace(sig.Type + " " + sig.Value + " " + sig.Detail),
				Raw: map[string]any{
					"ip":     sig.IP,
					"type":   sig.Type,
					"value":  sig.Value,
					"detail": sig.Detail,
				},
			})
		}
	}
	if len(bag.StrongAccessHints) > 0 {
		for _, h := range bag.StrongAccessHints {
			directItems = append(directItems, models.EvidenceItem{
				Source:     "cpe_wan",
				TargetType: publicTypeForTarget(h),
				Strength:   "strong",
				Confidence: maxFloat(bag.PhysicalEvidence, 0.80),
				Reason:     "CPE exposed direct physical WAN evidence for " + h + ".",
				Raw:        map[string]any{"hint": h},
			})
		}
	}

	var deviceItems []models.EvidenceItem
	if matched {
		deviceItems = append(deviceItems, models.EvidenceItem{
			Source:     "fingerprint",
			TargetType: inferTargetType(strings.Join(bag.Hints, " ") + " " + bag.RouterModel),
			Strength:   "medium",
			Confidence: 0.72,
			Reason:     "Known CPE model fingerprint matched the observed gateway identity.",
			Raw:        map[string]any{"router_model": bag.RouterModel},
		})
	}
	if bag.RouterModel != "" {
		deviceItems = append(deviceItems, models.EvidenceItem{
			Source:     "gateway_identity",
			TargetType: inferTargetType(bag.RouterModel + " " + strings.Join(bag.Hints, " ")),
			Strength:   strengthLabel(bag.DeviceEvidence),
			Confidence: bag.DeviceEvidence,
			Reason:     "Gateway model was identified: " + bag.RouterModel + ".",
			Raw:        map[string]any{"router_model": bag.RouterModel},
		})
	}
	for _, d := range bag.GatewayDevices {
		if d.Manufacturer != "" || d.Model != "" || d.FingerprintID != "" {
			deviceItems = append(deviceItems, models.EvidenceItem{
				Source:     "gateway_device:" + d.IP,
				TargetType: firstHintTarget(d.AccessHints),
				Strength:   strengthLabel(d.DeviceConfidence),
				Confidence: d.DeviceConfidence,
				Reason:     "Gateway device model evidence was identified.",
				Raw: map[string]any{
					"ip":             d.IP,
					"manufacturer":   d.Manufacturer,
					"model":          d.Model,
					"fingerprint_id": d.FingerprintID,
					"access_hints":   d.AccessHints,
				},
			})
		}
	}

	var topologyItems []models.EvidenceItem
	if bag.Gateway != "" {
		topologyItems = append(topologyItems, models.EvidenceItem{
			Source:     "gateway_probe",
			TargetType: models.CatUnknown,
			Strength:   "weak",
			Confidence: 0.25,
			Reason:     "Default gateway was detected; this is topology context, not WAN physical proof.",
			Raw:        map[string]any{"gateway": bag.Gateway},
		})
	}
	if len(bag.GatewayChain) > 0 {
		topologyItems = append(topologyItems, models.EvidenceItem{
			Source:     "gateway_chain_probe",
			TargetType: models.CatUnknown,
			Strength:   "weak",
			Confidence: 0.35,
			Reason:     "Private gateway chain was detected; this can explain router/CPE placement but not access type.",
			Raw:        map[string]any{"gateway_chain": bag.GatewayChain},
		})
	}
	if hasUsefulNATContext(bag.NATTopology) {
		topologyItems = append(topologyItems, models.EvidenceItem{
			Source:     "nat_topology",
			TargetType: models.CatUnknown,
			Strength:   "weak",
			Confidence: 0.35,
			Reason:     "NAT topology is network context only.",
			Raw:        map[string]any{"topology": bag.NATTopology.Topology, "double_nat": bag.NATTopology.DoubleNAT},
		})
	}

	var perfItems []models.EvidenceItem
	if bag.PerformanceProfile != nil {
		perfItems = append(perfItems, models.EvidenceItem{
			Source:     "performance_profile_probe",
			TargetType: models.CatUnknown,
			Strength:   "weak",
			Confidence: maxFloat(bag.PerformanceEvidence, 0.25),
			Reason:     fmt.Sprintf("Idle latency %.1f ms and jitter %.1f ms are contextual only.", bag.PerformanceProfile.IdleLatencyMS, bag.PerformanceProfile.JitterMS),
			Raw:        map[string]any{"idle_latency_ms": bag.PerformanceProfile.IdleLatencyMS, "jitter_ms": bag.PerformanceProfile.JitterMS},
		})
	} else if bag.HasLatency {
		perfItems = append(perfItems, models.EvidenceItem{
			Source:     "latency_probe",
			TargetType: models.CatUnknown,
			Strength:   "weak",
			Confidence: 0.20,
			Reason:     fmt.Sprintf("Latency %.1f ms is compatible with several access types and is not conclusive.", bag.AvgMS),
			Raw:        map[string]any{"avg_ms": bag.AvgMS, "jitter_ms": bag.JitterMS},
		})
	}

	var regionalItems []models.EvidenceItem
	if bag.Org != "" {
		regionalItems = append(regionalItems, models.EvidenceItem{
			Source:     "asn_probe",
			TargetType: models.CatUnknown,
			Strength:   "weak",
			Confidence: 0.25,
			Reason:     "Operator/ASN context does not prove the physical access type.",
			Raw:        map[string]any{"org": bag.Org},
		})
	}
	if bag.PTR != "" {
		regionalItems = append(regionalItems, models.EvidenceItem{
			Source:     "asn_probe",
			TargetType: models.CatUnknown,
			Strength:   "weak",
			Confidence: 0.20,
			Reason:     "PTR/reverse DNS is regional context only.",
			Raw:        map[string]any{"ptr": bag.PTR},
		})
	}

	return models.EvidenceTiers{
		DirectPhysical: tier("direct_physical", directItems),
		DeviceModel:    tier("device_model", deviceItems),
		Topology:       tier("topology", topologyItems),
		Performance:    tier("performance", perfItems),
		Regional:       tier("regional", regionalItems),
	}
}

func tier(name string, items []models.EvidenceItem) models.EvidenceTier {
	strength := "none"
	conf := 0.0
	for i := range items {
		items[i].Confidence = clamp01(items[i].Confidence)
		if items[i].TargetType == "" {
			items[i].TargetType = models.CatUnknown
		}
		if stronger(items[i].Strength, strength) {
			strength = items[i].Strength
		}
		conf = maxFloat(conf, items[i].Confidence)
	}
	present := len(items) > 0
	return models.EvidenceTier{
		Present:    present,
		Status:     tierStatus(name, strength, present),
		Items:      items,
		Strength:   strength,
		Confidence: clamp01(conf),
	}
}

func tierStatus(name, strength string, present bool) string {
	if !present {
		return "missing"
	}
	if name == "topology" && strength == "weak" {
		return "present_but_weak"
	}
	return strength
}

func stronger(a, b string) bool {
	return strengthRank(a) > strengthRank(b)
}

func strengthRank(s string) int {
	switch s {
	case "strong":
		return 3
	case "medium":
		return 2
	case "weak":
		return 1
	default:
		return 0
	}
}

func compactStrings(values []string) []string {
	var out []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			out = appendUniqueStrings(out, v)
		}
	}
	return out
}

func lineProfileTargetType(lp *models.LineProfile) string {
	if lp == nil {
		return models.CatUnknown
	}
	for _, h := range linestats.AccessHints(lp) {
		return publicTypeForTarget(h)
	}
	return inferTargetType(lp.Medium + " " + lp.Technology + " " + lp.Subtype)
}

func firstHintTarget(hints []string) string {
	for _, h := range hints {
		if h != "" {
			return publicTypeForTarget(h)
		}
	}
	return models.CatUnknown
}

func inferTargetType(text string) string {
	l := strings.ToLower(text)
	switch {
	case strings.Contains(l, "xgs-pon"), strings.Contains(l, "xg-pon"), strings.Contains(l, "gpon"), strings.Contains(l, "epon"), strings.Contains(l, "ont"), strings.Contains(l, "optical"):
		return models.TypeFiber
	case strings.Contains(l, "vdsl"):
		return models.TypeVDSL
	case strings.Contains(l, "adsl"), strings.Contains(l, "iftype adsl"):
		return models.TypeADSL
	case strings.Contains(l, "dsl"), strings.Contains(l, "ptm"), strings.Contains(l, "atm"):
		return models.TypeDSL
	case strings.Contains(l, "docsis"), strings.Contains(l, "cable"):
		return models.TypeCable
	case strings.Contains(l, "lte"), strings.Contains(l, "nr"), strings.Contains(l, "wwan"), strings.Contains(l, "cellular"), strings.Contains(l, "5g"):
		return models.TypeFWA
	case strings.Contains(l, "satellite"), strings.Contains(l, "starlink"), strings.Contains(l, "vsat"):
		return models.TypeSatellite
	case strings.Contains(l, "wisp"), strings.Contains(l, "radio"), strings.Contains(l, "wireless cpe"):
		return models.TypeWISP
	case strings.Contains(l, "ethernet"):
		return models.TypeEthernetWAN
	default:
		return models.CatUnknown
	}
}

func publicTypeForTarget(target string) string {
	switch models.CategoryFor(target) {
	case models.CatFiber:
		return models.TypeFiber
	case models.CatDSL:
		switch target {
		case models.TypeVDSL, models.TypeVDSL2:
			return models.TypeVDSL
		case models.TypeADSL, models.TypeADSL2:
			return models.TypeADSL
		default:
			return models.TypeDSL
		}
	case models.CatCable:
		return models.TypeCable
	case models.CatMobile:
		return models.TypeFWA
	case models.CatWireless:
		return models.TypeWISP
	case models.CatSatellite:
		return models.TypeSatellite
	case models.CatEthernetWAN:
		return models.TypeEthernetWAN
	default:
		if target == "" {
			return models.CatUnknown
		}
		return target
	}
}

func directProbeHasPhysicalEvidence(r models.ProbeResult) bool {
	if getBool(r.Evidence, "direct_physical_evidence") {
		return true
	}
	if wanAccess := strings.ToLower(getString(r.Evidence, "wan_access_type")); wanAccess != "" &&
		wanAccess != "unknown" && wanAccess != "other" {
		return true
	}
	for _, sig := range getWANSignals(r.Evidence, "wan_signals") {
		if sig.Strength == string(models.EvidencePhysical) || strings.EqualFold(sig.Strength, "strong") || sig.Confidence >= 0.70 {
			return true
		}
	}
	if lp := linestats.Parse(getStringMap(r.Evidence, "cpe_kv"), getString(r.Evidence, "cpe_text"), r.ProbeName); lp != nil {
		return true
	}
	text := strings.ToLower(strings.Join(getStrings(r.Evidence, "cpe_services"), " ") + " " + getString(r.Evidence, "cpe_text"))
	for _, tok := range directEvidenceTokens {
		if strings.Contains(text, tok) {
			return true
		}
	}
	return false
}

func gatewayDeviceHasDirectPhysicalEvidence(d models.GatewayDevice) bool {
	if d.WANCommonInterfaceFound && d.WANAccessType != "" {
		return true
	}
	if len(d.PhysicalHints) > 0 {
		return true
	}
	text := strings.ToLower(strings.Join(append([]string{d.WANAccessType, d.PhysicalLinkStatus}, d.TR064Services...), " "))
	for _, tok := range directEvidenceTokens {
		if strings.Contains(text, tok) {
			return true
		}
	}
	return false
}

func gatewayDevicePhysicalHints(d models.GatewayDevice) []string {
	if len(d.PhysicalHints) > 0 {
		return d.PhysicalHints
	}
	return d.AccessHints
}

func detectConflicts(results []models.ProbeResult, bag evidenceBag, scores map[string]float64) []models.DataConflict {
	var conflicts []models.DataConflict

	deviceReach := map[string]bool{}
	for _, d := range bag.GatewayDevices {
		if d.IP != "" {
			deviceReach[d.IP] = d.Reachable
		}
	}
	for _, r := range results {
		if r.Status != models.StatusSuccess {
			continue
		}
		if r.ProbeName == "gateway_reachability_diagnostics_probe" || r.ProbeName == "gateway_reachability_probe" {
			ip := firstNonEmpty(getString(r.Evidence, "gateway_ip"), getString(r.Evidence, "ip"))
			reachable := getBool(r.Evidence, "management_reachable")
			if ip != "" {
				if prev, ok := deviceReach[ip]; ok && prev != reachable {
					conflicts = append(conflicts, models.DataConflict{
						Field:      "gateway_devices[" + ip + "].reachable",
						Severity:   "high",
						Effect:     "downgrade_confidence",
						Resolution: "Prefer active reachability probe over stale gateway summary field.",
						Values: []models.ConflictValue{
							{Source: "gateway_devices", Value: prev},
							{Source: r.ProbeName, Value: reachable},
						},
					})
				}
			}
		}
	}

	if len(bag.GatewayChain) > 0 {
		for _, r := range results {
			if r.Status != models.StatusSuccess || r.ProbeName != "traceroute_probe" {
				continue
			}
			chain := detectGatewayChain(getStrings(r.Evidence, "hops")).Chain
			if len(chain) > 0 && strings.Join(chain, ",") != strings.Join(bag.GatewayChain, ",") {
				conflicts = append(conflicts, models.DataConflict{
					Field:      "gateway_chain",
					Severity:   "medium",
					Effect:     "downgrade_confidence",
					Resolution: "Prefer the latest private-hop traceroute when it contains a longer supported private chain.",
					Values: []models.ConflictValue{
						{Source: "traceroute_probe", Value: chain},
						{Source: "gateway_chain_probe", Value: bag.GatewayChain},
					},
				})
			}
		}
	}

	conflicts = append(conflicts, detectDoubleNATConflicts(results, bag)...)
	conflicts = append(conflicts, detectWANTypeConflicts(results)...)

	if hasEthernetWANSignal(bag) && (scores[models.TypeDSL] > 0.35 || scores[models.TypeVDSL] > 0.35 || scores[models.TypeADSL] > 0.35 || scores[models.TypeFiber] > 0.35) {
		conflicts = append(conflicts, models.DataConflict{
			Field:      "classification.wan_access_type",
			Severity:   "high",
			Effect:     "set_conflicting_evidence",
			Resolution: "Direct physical evidence wins over weaker model or score hints; Ethernet WAN is not Fiber.",
			Values: []models.ConflictValue{
				{Source: "direct_wan_evidence", Value: models.TypeEthernetWAN},
				{Source: "weaker_scores", Value: categoryScores(scores)},
			},
		})
	}

	return conflicts
}

func detectDoubleNATConflicts(results []models.ProbeResult, bag evidenceBag) []models.DataConflict {
	finalDouble := bag.DoubleNATPossible || (bag.NATTopology != nil && (bag.NATTopology.DoubleNAT || bag.NATTopology.InternalDoubleNATPossible))
	var conflicts []models.DataConflict
	for _, r := range results {
		if r.Status != models.StatusSuccess {
			continue
		}
		sourceDouble, ok := doubleNATValue(r)
		if !ok || sourceDouble == finalDouble {
			continue
		}
		severity := "medium"
		if r.ProbeName == "gateway_chain_probe" || r.ProbeName == "traceroute_probe" {
			severity = "high"
		}
		conflicts = append(conflicts, models.DataConflict{
			Field:      "nat_topology.double_nat",
			Severity:   severity,
			Effect:     "downgrade_confidence",
			Resolution: "Final double-NAT state must be derived from supported gateway-chain or NAT evidence.",
			Values: []models.ConflictValue{
				{Source: r.ProbeName, Value: sourceDouble},
				{Source: "final_nat_topology", Value: finalDouble},
			},
		})
	}
	return conflicts
}

func doubleNATValue(r models.ProbeResult) (bool, bool) {
	if _, ok := r.Evidence["double_nat_possible"]; ok {
		return getBool(r.Evidence, "double_nat_possible"), true
	}
	if _, ok := r.Evidence["double_nat"]; ok {
		return getBool(r.Evidence, "double_nat"), true
	}
	if nat := getNATTopology(r.Evidence, "nat_topology"); nat != nil {
		return nat.DoubleNAT || nat.InternalDoubleNATPossible, true
	}
	return false, false
}

func detectWANTypeConflicts(results []models.ProbeResult) []models.DataConflict {
	type sourceType struct {
		source string
		value  string
	}
	var values []sourceType
	for _, r := range results {
		if r.Status != models.StatusSuccess {
			continue
		}
		if v := getString(r.Evidence, "wan_access_type"); v != "" {
			values = append(values, sourceType{source: r.ProbeName, value: conflictFamily(publicTypeForTarget(inferTargetType(v)))})
		}
		for _, sig := range getWANSignals(r.Evidence, "wan_signals") {
			if sig.Value == "" {
				continue
			}
			values = append(values, sourceType{source: firstNonEmpty(sig.Source, r.ProbeName), value: conflictFamily(publicTypeForTarget(inferTargetType(sig.Value + " " + sig.Detail)))})
		}
		if getBool(r.Evidence, "direct_physical_evidence") {
			if target := firstHintTarget(r.Hints); target != models.CatUnknown {
				values = append(values, sourceType{source: r.ProbeName, value: conflictFamily(target)})
			}
		}
	}
	if len(values) < 2 {
		return nil
	}
	first := values[0]
	for _, v := range values[1:] {
		if v.value == "" || v.value == models.CatUnknown || v.value == first.value {
			continue
		}
		return []models.DataConflict{{
			Field:      "wan_access_type",
			Severity:   "high",
			Effect:     "set_conflicting_evidence",
			Resolution: "Prefer explicit active physical interface evidence over generic Ethernet WAN or model hints.",
			Values: []models.ConflictValue{
				{Source: first.source, Value: first.value},
				{Source: v.source, Value: v.value},
			},
		}}
	}
	return nil
}

func conflictFamily(target string) string {
	switch target {
	case models.TypeVDSL, models.TypeADSL, models.TypeDSL, models.TypeVDSL2, models.TypeADSL2:
		return models.TypeDSL
	case models.TypeFiber, models.TypeFTTH, models.TypeGPON, models.TypeEPON, models.TypeXGPON, models.TypeXGSPON:
		return models.TypeFiber
	case models.TypeCable, models.TypeDOCSIS:
		return models.TypeCable
	case models.TypeFWA, models.TypeMobile, models.TypeLTE, models.TypeNR5G:
		return models.TypeFWA
	case models.TypeEthernetWAN:
		return models.TypeEthernetWAN
	default:
		return target
	}
}

func hasEthernetWANSignal(bag evidenceBag) bool {
	for _, sig := range bag.WANSignals {
		text := strings.ToLower(sig.Value + " " + sig.Detail)
		if strings.Contains(text, "ethernet") {
			return true
		}
	}
	return strings.Contains(strings.ToLower(strings.Join(bag.StrongAccessHints, " ")), strings.ToLower(models.TypeEthernetWAN))
}

func ethernetWANOnlyEvidence(bag evidenceBag) bool {
	if !hasEthernetWANSignal(bag) {
		return false
	}
	for _, h := range bag.StrongAccessHints {
		if h != models.TypeEthernetWAN {
			return false
		}
	}
	for _, sig := range bag.WANSignals {
		target := inferTargetType(sig.Value + " " + sig.Detail)
		if target != models.TypeEthernetWAN && target != models.CatUnknown {
			return false
		}
	}
	return true
}

func hasHighSeverityConflict(conflicts []models.DataConflict) bool {
	for _, c := range conflicts {
		if c.Severity == "high" {
			return true
		}
	}
	return false
}

func hasMediumSeverityConflict(conflicts []models.DataConflict) bool {
	for _, c := range conflicts {
		if c.Severity == "medium" {
			return true
		}
	}
	return false
}

func hasWANClassificationConflict(conflicts []models.DataConflict) bool {
	for _, c := range conflicts {
		if strings.Contains(c.Field, "wan_access_type") || strings.Contains(c.Field, "classification") {
			return true
		}
	}
	return false
}

func conflictPenalty(conflicts []models.DataConflict) float64 {
	if hasHighSeverityConflict(conflicts) {
		return 0.25
	}
	if len(conflicts) > 0 {
		return 0.10
	}
	return 0
}

func populateOutputContract(result *models.ScanResult, bag evidenceBag, matched bool, leading string) {
	primary, subtype := publicClassification(result.PrimaryType, leading, bag)
	direct := hasDirectPhysicalEvidence(bag)
	state := classificationState(primary, result.Confidence, direct, result.Conflicts)
	safe := primary != "Unknown" && state == "confirmed" && direct && len(result.Conflicts) == 0

	result.Classification = models.Classification{
		PrimaryType:          primary,
		Subtype:              subtype,
		Confidence:           result.Confidence,
		DecisionQuality:      result.DecisionQuality,
		State:                state,
		SafeToDisplayAsFinal: safe,
	}
	result.UI = buildUIOutput(*result, bag, primary, subtype, state)
}

func publicClassification(primary, leading string, bag evidenceBag) (string, *string) {
	if primary == "" || primary == "Unknown" {
		return "Unknown", nil
	}
	raw := primary
	subtype := ""
	if bag.LineProfile != nil && bag.LineProfile.Subtype != "" {
		subtype = bag.LineProfile.Subtype
	}
	switch raw {
	case models.TypeVDSL2:
		if subtype == "" {
			subtype = models.TypeVDSL2
		}
		return models.TypeVDSL, stringPtr(subtype)
	case models.TypeVDSL:
		return models.TypeVDSL, optionalSubtype(subtype)
	case models.TypeADSL2:
		if subtype == "" {
			subtype = models.TypeADSL2
		}
		return models.TypeADSL, stringPtr(subtype)
	case models.TypeADSL:
		return models.TypeADSL, optionalSubtype(subtype)
	case models.TypeDSL:
		return models.TypeDSL, optionalSubtype(subtype)
	case models.TypeGPON, models.TypeEPON, models.TypeXGPON, models.TypeXGSPON, models.TypeTenGEPON:
		if subtype == "" {
			subtype = raw
		}
		return models.TypeFiber, stringPtr(subtype)
	case models.TypeFTTH, models.TypeFiber:
		return models.TypeFiber, optionalSubtype(subtype)
	case models.TypeDOCSIS:
		if subtype == "" {
			subtype = models.TypeDOCSIS
		}
		return models.TypeCable, stringPtr(subtype)
	case models.TypeCable:
		return models.TypeCable, optionalSubtype(subtype)
	case models.TypeLTE, models.TypeNR5G, models.TypeFWA, models.TypeFWA5G, models.TypeMobile, models.TypeCellular:
		if raw != models.TypeFWA && subtype == "" {
			subtype = raw
		}
		return models.TypeFWA, optionalSubtype(subtype)
	case models.TypeWISP, models.TypeFixedWireless:
		return models.TypeWISP, optionalSubtype(subtype)
	case models.TypeSatellite, models.TypeLEOSatellite, models.TypeGEOSatellite, models.TypeMEOSatellite, models.TypeVSAT:
		if raw != models.TypeSatellite && subtype == "" {
			subtype = raw
		}
		return models.TypeSatellite, optionalSubtype(subtype)
	case models.TypeEthernetWAN:
		return models.TypeEthernetWAN, nil
	default:
		return raw, optionalSubtype(subtype)
	}
}

func classificationState(primary string, confidence float64, direct bool, conflicts []models.DataConflict) string {
	if primary == "Unknown" {
		if len(conflicts) > 0 {
			return "conflicting_evidence"
		}
		return "insufficient_evidence"
	}
	if hasWANClassificationConflict(conflicts) {
		return "conflicting_evidence"
	}
	if direct && confidence >= 0.85 {
		return "confirmed"
	}
	if confidence >= 0.60 {
		return "probable"
	}
	if confidence >= 0.40 {
		return "possible"
	}
	return "insufficient_evidence"
}

func buildUIOutput(result models.ScanResult, bag evidenceBag, primary string, subtype *string, state string) models.UIOutput {
	label := displayLabel(primary, subtype)
	ui := models.UIOutput{Badges: evidenceBadges(result.EvidenceTiers)}

	switch {
	case state == "confirmed":
		ui.Headline = "Detected: " + label
	case state == "probable":
		ui.Headline = "Probable: " + label
	case state == "possible":
		ui.Headline = "Possible: " + label
	default:
		ui.Headline = "Access type unknown"
	}

	if primary == "Unknown" {
		if isFiberVDSLTie(result.Scores) {
			ui.Summary = "Fiber and VDSL are both possible, but no physical WAN evidence was found."
		} else if candidates := likelyCandidateNames(result.Scores, 3); len(candidates) > 0 {
			ui.Summary = "The strongest candidates are " + strings.Join(candidates, ", ") + ", but the evidence is not strong enough to classify."
		} else {
			ui.Summary = "No direct WAN physical evidence was collected."
		}
	} else if result.Classification.SafeToDisplayAsFinal {
		ui.Summary = "Direct WAN physical evidence supports this access type."
	} else {
		ui.Summary = "This is not final; stronger WAN physical evidence would improve certainty."
	}
	ui.Warnings = append(ui.Warnings, result.UncertaintyReasons...)
	if primary == "Unknown" {
		ui.Warnings = append(ui.Warnings, "Do not treat latency, PTR, ASN, IPv6, public IP, CGNAT, or Ethernet adapter as proof of access type.")
	}
	for _, c := range result.Conflicts {
		ui.Warnings = append(ui.Warnings, "Conflict: "+c.Field+" ("+c.Effect+")")
	}
	return ui
}

func evidenceBadges(t models.EvidenceTiers) []string {
	var badges []string
	if t.DirectPhysical.Present {
		badges = append(badges, "Direct WAN evidence")
	}
	if t.DeviceModel.Present {
		badges = append(badges, "CPE model found")
	}
	if t.Topology.Present {
		badges = append(badges, "Topology context")
	}
	if t.Performance.Present {
		badges = append(badges, "Performance only")
	}
	if t.Regional.Present {
		badges = append(badges, "Operator hint only")
	}
	if !t.DeviceModel.Present {
		badges = append(badges, "No CPE model")
	}
	if !t.DirectPhysical.Present {
		badges = append(badges, "No direct WAN evidence")
	}
	return badges
}

func displayLabel(primary string, subtype *string) string {
	if primary == "Unknown" {
		return primary
	}
	if subtype == nil || *subtype == "" {
		if primary == models.TypeEthernetWAN {
			return "Ethernet WAN"
		}
		return primary
	}
	switch primary {
	case models.TypeFiber:
		return "Fiber/" + *subtype
	case models.TypeCable:
		return "Cable/" + *subtype
	default:
		return *subtype
	}
}

func stringPtr(s string) *string { return &s }

func optionalSubtype(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return stringPtr(s)
}
