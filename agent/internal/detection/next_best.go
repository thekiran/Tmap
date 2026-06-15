package detection

import (
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

// nextBestProbes recommends, in English, which probes would most improve the
// verdict given what is currently missing. It is dynamic: the suggestions depend
// on which evidence gaps exist. Returns nil when strong physical evidence already
// exists (nothing more is needed to commit).
func nextBestProbes(bag evidenceBag, matched bool) []models.NextBestProbe {
	if hasStrongPhysicalEvidence(bag, matched) {
		return nil
	}
	var out []models.NextBestProbe

	noCPEModel := bag.RouterModel == "" && len(bag.GatewayDevices) == 0
	upstreamUnreachable := false
	for _, d := range bag.GatewayDevices {
		if d.Role == "upstream_private_gateway" && !d.Reachable {
			upstreamUnreachable = true
		}
	}

	// Missing CPE model / WAN evidence → try to read the CPE directly.
	if noCPEModel || !bag.UPnPFound {
		out = append(out,
			models.NextBestProbe{
				ProbeName:        "tr064_probe_v2",
				Reason:           "No CPE WAN evidence is available.",
				ExpectedEvidence: "CPE model, WAN services, DSL/PTM/GPON/DOCSIS/LTE indicators.",
				Safety:           "Only observed private gateway IPs are queried; no authentication bypass or brute force.",
			},
			models.NextBestProbe{
				ProbeName:        "upnp_igd_deep_probe_v2",
				Reason:           "WANCommonInterfaceConfig data would identify the WAN access medium.",
				ExpectedEvidence: "WANAccessType, PhysicalLinkStatus, Layer1 up/downstream bitrates.",
				Safety:           "SSDP and IGD control URLs on private gateways only; no public scanning.",
			},
			models.NextBestProbe{
				ProbeName:        "http_fingerprint_v3",
				Reason:           "No CPE model or fingerprint was identified.",
				ExpectedEvidence: "HTTP title, realm, favicon hash, TLS CN/SAN, HTML meta generator.",
				Safety:           "Only private gateway HTTP/HTTPS endpoints are read; no credential attempts.",
			},
		)
	}

	// Upstream CPE seen but unreachable → diagnose reachability.
	if upstreamUnreachable {
		out = append(out, models.NextBestProbe{
			ProbeName:        "gateway_reachability_diagnostics_probe",
			Reason:           "An upstream CPE was detected but could not be reached.",
			ExpectedEvidence: "Route table, local firewall state, reachability of the upstream gateway.",
			Safety:           "Local route/firewall inspection only; no remote scanning.",
		})
	}

	// No strong physical evidence at all → opt-in SNMP can read line/DOCSIS MIBs.
	out = append(out, models.NextBestProbe{
		ProbeName:        "snmp_probe_opt_in",
		Reason:           "No physical-layer evidence of the access type was found.",
		ExpectedEvidence: "ifType/ifDescr (dsl/ptm/atm/gpon/docsis/lte), DOCSIS or DSL line MIBs.",
		Safety:           "Disabled by default; runs only with user-provided credentials against private gateway IPs; no community guessing.",
	})

	// No NAT clarity → STUN/PCP/traceroute.
	if !hasUsefulNATContext(bag.NATTopology) {
		out = append(out, models.NextBestProbe{
			ProbeName:        "stun_pcp_nat_probe",
			Reason:           "The NAT topology is unclear.",
			ExpectedEvidence: "STUN public IP/port, PCP/NAT-PMP reachability, CGNAT indication.",
			Safety:           "Standard STUN servers and the local gateway only; no scanning.",
		})
	}

	// No local medium identified → enumerate interfaces.
	if bag.LocalAccess == "" {
		out = append(out, models.NextBestProbe{
			ProbeName:        "os_interface_probe_v3",
			Reason:           "The local access medium was not identified.",
			ExpectedEvidence: "Active adapter type (Ethernet/Wi-Fi/Cellular), link speed.",
			Safety:           "Reads local OS interface state only.",
		})
	}

	return out
}

// conflictReasons surfaces, in English, contradictions between strong signals so
// the user understands why confidence was reduced.
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
			reasons = append(reasons, "The Fiber/FTTH score is high while WANAccessType reports Ethernet; an ONT behind a router is possible.")
		}
	}
	return reasons
}
