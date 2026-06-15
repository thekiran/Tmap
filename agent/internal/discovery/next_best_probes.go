package discovery

import "github.com/thekiran/iad/internal/model"

func NextBestProbes(devices []model.Device, topology model.Topology, conflicts []model.Conflict, mode model.ScanMode) []model.NextBestProbe {
	var out []model.NextBestProbe
	if mode == model.ScanModeSafe {
		out = append(out, model.NextBestProbe{
			ProbeName:        "ping_sweep_probe",
			Reason:           "Safe mode does not actively discover additional hosts.",
			ExpectedEvidence: []string{"responsive_private_hosts", "latency"},
			Safety:           "Private subnets only, rate limited, no public scanning.",
		})
	}
	if missingDeviceTypes(devices) {
		out = append(out,
			model.NextBestProbe{
				ProbeName:        "ssdp_upnp_probe",
				Reason:           "UPnP can identify routers, media devices, printers, and NAS devices without credentials.",
				ExpectedEvidence: []string{"device_type", "friendly_name", "manufacturer", "model", "service_list"},
				Safety:           "UDP SSDP M-SEARCH on the local network only.",
			},
			model.NextBestProbe{
				ProbeName:        "mdns_probe",
				Reason:           "mDNS service records can identify printers, NAS devices, phones, TVs, and workstations.",
				ExpectedEvidence: []string{"hostnames", "service_types", "printer_services", "nas_services", "workstation_services"},
				Safety:           "Local multicast only. No credentials.",
			},
			model.NextBestProbe{
				ProbeName:        "http_fingerprint_probe",
				Reason:           "HTTP headers and titles can identify device families.",
				ExpectedEvidence: []string{"status_code", "server_header", "auth_realm", "redirect", "vendor_model_strings"},
				Safety:           "Private IPs only. No login attempts.",
			},
		)
	}
	if topologyNeedsPhysicalEvidence(topology) {
		out = append(out,
			model.NextBestProbe{
				ProbeName:        "snmp_optin_probe",
				Reason:           "Read-only SNMP can expose switch, bridge, LLDP, and interface tables when the user provides credentials.",
				ExpectedEvidence: []string{"sysDescr", "ifTable", "bridge_mib", "lldp_mib", "forwarding_table"},
				Safety:           "Disabled by default. No community guessing or brute force.",
			},
			model.NextBestProbe{
				ProbeName:        "lldp_cdp_passive_probe",
				Reason:           "Passive LLDP/CDP frames can confirm physical neighbor relationships.",
				ExpectedEvidence: []string{"chassis_id", "port_id", "system_name", "management_ip", "neighbor_relationship"},
				Safety:           "Passive capture only when supported. No packet injection.",
			},
		)
	}
	if hasConflictType(conflicts, "same_ip_different_macs") || hasConflictType(conflicts, "gateway_chain_differs_between_probes") {
		out = append(out, model.NextBestProbe{
			ProbeName:        "arp_neighbor_probe",
			Reason:           "Neighbor tables can resolve identity churn and gateway-chain conflicts.",
			ExpectedEvidence: []string{"ip", "mac", "interface", "reachable_state"},
			Safety:           "Reads local OS neighbor state only.",
		})
	}
	if mode != model.ScanModeDeep {
		out = append(out, model.NextBestProbe{
			ProbeName:        "traceroute_isp_path_probe",
			Reason:           "Traceroute can identify private upstream gateways and observed ISP route hops.",
			ExpectedEvidence: []string{"private_hops", "first_public_hop", "asn", "reverse_dns", "latency_per_hop"},
			Safety:           "No public port scanning. Public hops are represented only as observed route hops.",
		})
	}
	return dedupeProbeSuggestions(out)
}

func missingDeviceTypes(devices []model.Device) bool {
	for _, d := range devices {
		if d.DeviceType == "" || d.DeviceType == model.DeviceTypeUnknown {
			return true
		}
	}
	return len(devices) == 0
}

func topologyNeedsPhysicalEvidence(topology model.Topology) bool {
	if len(topology.Edges) == 0 {
		return true
	}
	for _, edge := range topology.Edges {
		if edge.Inferred || edge.Confidence < 0.70 {
			return true
		}
	}
	return false
}

func hasConflictType(conflicts []model.Conflict, typ string) bool {
	for _, c := range conflicts {
		if c.Type == typ {
			return true
		}
	}
	return false
}

func dedupeProbeSuggestions(in []model.NextBestProbe) []model.NextBestProbe {
	seen := map[string]bool{}
	var out []model.NextBestProbe
	for _, p := range in {
		if p.ProbeName == "" || seen[p.ProbeName] {
			continue
		}
		seen[p.ProbeName] = true
		out = append(out, p)
	}
	return out
}
