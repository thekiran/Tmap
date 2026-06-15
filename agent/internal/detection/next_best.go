package detection

import (
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

// nextBestProbes recommends probes based on missing evidence and conflicts. It
// deliberately names read-only, local/private probes that can improve certainty
// without unsafe scanning or authentication bypass.
func nextBestProbes(bag evidenceBag, matched bool, conflicts []models.DataConflict, scores map[string]float64) []models.NextBestProbe {
	if hasStrongPhysicalEvidence(bag, matched) && len(conflicts) == 0 {
		return nil
	}

	var out []models.NextBestProbe
	noCPEModel := bag.RouterModel == "" && len(bag.GatewayDevices) == 0

	if !hasDirectPhysicalEvidence(bag) {
		out = append(out,
			models.NextBestProbe{
				ProbeName:        "upnp_wan_common_interface_probe",
				Reason:           "WANCommonInterfaceConfig may expose WANAccessType and link properties.",
				ExpectedEvidence: []string{"WANAccessType", "Layer1UpstreamMaxBitRate", "Layer1DownstreamMaxBitRate", "PhysicalLinkStatus"},
				Safety:           "Local private gateway only. Read-only. No authentication bypass.",
			},
			models.NextBestProbe{
				ProbeName:        "tr064_probe",
				Reason:           "TR-064 WAN/DSL descriptors can expose DSL/PTM/ATM and link status.",
				ExpectedEvidence: []string{"WANDSLInterfaceConfig", "WANCommonInterfaceConfig", "PTM/ATM link", "VDSL2/ADSL modulation"},
				Safety:           "Private gateway only. Authenticated calls require user-provided credentials.",
			},
			models.NextBestProbe{
				ProbeName:        "http_fingerprint_probe",
				Reason:           "A local gateway fingerprint may identify the CPE model.",
				ExpectedEvidence: []string{"HTTP title", "Server header", "WWW-Authenticate realm", "favicon hash", "TLS CN/SAN"},
				Safety:           "Private gateway HTTP/HTTPS only. No login attempts.",
			},
			models.NextBestProbe{
				ProbeName:        "snmp_opt_in_probe",
				Reason:           "SNMP can expose ifType/ifDescr and DSL/DOCSIS/optical MIB objects when explicitly enabled.",
				ExpectedEvidence: []string{"IF-MIB ifType", "ifDescr/ifName", "ADSL-LINE-MIB", "VDSL2-LINE-MIB", "DOCSIS MIB", "optical/cellular interfaces"},
				Safety:           "Disabled by default; requires user-provided read-only credentials. No community guessing.",
			},
			models.NextBestProbe{
				ProbeName:        "tr181_interface_stack_probe",
				Reason:           "TR-181 InterfaceStack can reveal the physical layer below the IP interface.",
				ExpectedEvidence: []string{"Device.DSL.Line", "Device.PTM.Link", "Device.ATM.Link", "Device.Optical.Interface", "Device.Cellular.Interface", "Device.Ethernet.Link"},
				Safety:           "Requires supported local management API or user-provided CPE access. Read-only.",
			},
		)
	}

	if hasConflictField(conflicts, "gateway_devices") || hasConflictField(conflicts, "gateway_chain") {
		out = append(out,
			models.NextBestProbe{
				ProbeName:        "gateway_reachability_probe",
				Reason:           "Gateway reachability facts disagree and should be rechecked directly.",
				ExpectedEvidence: []string{"management_reachable", "reachable_protocols", "tcp_ports_reachable"},
				Safety:           "Private gateway candidates only. No public scanning or credential attempts.",
			},
			models.NextBestProbe{
				ProbeName:        "route_table_probe",
				Reason:           "The route table can identify the current default gateway and active interface.",
				ExpectedEvidence: []string{"default_gateway", "active_interface", "route_metric"},
				Safety:           "Reads local OS route state only.",
			},
			models.NextBestProbe{
				ProbeName:        "traceroute_private_hop_probe",
				Reason:           "Private traceroute hops can resolve gateway-chain ambiguity.",
				ExpectedEvidence: []string{"private_hops", "gateway_chain", "double_nat_possible"},
				Safety:           "Bounded traceroute; local private-hop interpretation only.",
			},
		)
	}

	if noCPEModel {
		out = append(out,
			models.NextBestProbe{
				ProbeName:        "upnp_device_description_probe",
				Reason:           "UPnP device descriptions may expose manufacturer and model.",
				ExpectedEvidence: []string{"manufacturer", "modelName", "modelNumber", "friendlyName", "serviceList"},
				Safety:           "SSDP M-SEARCH on local network only. Read-only.",
			},
			models.NextBestProbe{
				ProbeName:        "favicon_hash_probe",
				Reason:           "Favicon hash can help identify CPE web UI model family.",
				ExpectedEvidence: []string{"favicon_hash"},
				Safety:           "Private gateway HTTP/HTTPS only. No login attempts.",
			},
		)
	}

	if isFiberVDSLTie(scores) {
		out = append(out,
			models.NextBestProbe{
				ProbeName:        "tr064_dsl_probe",
				Reason:           "A Fiber/VDSL tie needs explicit DSL/PTM/ATM or optical WAN evidence.",
				ExpectedEvidence: []string{"WANDSLInterfaceConfig", "PTM/ATM link", "VDSL2/ADSL modulation"},
				Safety:           "Read-only private CPE probe; authenticated calls require user-provided credentials.",
			},
			models.NextBestProbe{
				ProbeName:        "upnp_wan_common_interface_probe",
				Reason:           "WANAccessType may resolve DSL vs Ethernet WAN.",
				ExpectedEvidence: []string{"WANAccessType", "PhysicalLinkStatus"},
				Safety:           "Local private gateway only. Read-only. No authentication bypass.",
			},
		)
	}

	return dedupeNextBest(out)
}

func hasConflictField(conflicts []models.DataConflict, token string) bool {
	for _, c := range conflicts {
		if strings.Contains(c.Field, token) {
			return true
		}
	}
	return false
}

func isFiberVDSLTie(scores map[string]float64) bool {
	fiber := maxFloat(scores[models.TypeFiber], maxFloat(scores[models.TypeFTTH], scores[models.TypeGPON]))
	vdsl := maxFloat(scores[models.TypeVDSL], scores[models.TypeVDSL2])
	if fiber <= 0 || vdsl <= 0 {
		return false
	}
	diff := fiber - vdsl
	if diff < 0 {
		diff = -diff
	}
	return diff < minCategoryGap
}

func dedupeNextBest(in []models.NextBestProbe) []models.NextBestProbe {
	seen := map[string]bool{}
	out := make([]models.NextBestProbe, 0, len(in))
	for _, p := range in {
		if p.ProbeName == "" || seen[p.ProbeName] {
			continue
		}
		seen[p.ProbeName] = true
		out = append(out, p)
	}
	return out
}

// conflictReasons surfaces contradictions in plain language for legacy
// uncertainty_reasons consumers.
func conflictReasons(scores map[string]float64, bag evidenceBag) []string {
	text := strings.ToLower(bag.WANSignalText)
	if text == "" {
		return nil
	}
	var reasons []string
	if strings.Contains(text, "ethernet") {
		if scores[models.TypeDSL] > 0.35 || scores[models.TypeVDSL] > 0.35 || scores[models.TypeADSL] > 0.35 {
			reasons = append(reasons, "The CPE model suggests DSL/VDSL while the reported WANAccessType is Ethernet.")
		}
		if scores[models.TypeFiber] > 0.35 || scores[models.TypeFTTH] > 0.35 || scores[models.TypeGPON] > 0.35 {
			reasons = append(reasons, "The Fiber/FTTH score is high while WANAccessType reports Ethernet; Ethernet WAN is not automatically Fiber.")
		}
	}
	return reasons
}
