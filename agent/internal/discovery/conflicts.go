package discovery

import (
	"fmt"
	"strings"

	"github.com/thekiran/iad/internal/model"
	"github.com/thekiran/iad/internal/safety"
)

func DetectConflicts(devices []model.Device, results []model.ProbeResult) []model.Conflict {
	var conflicts []model.Conflict
	conflicts = append(conflicts, detectIPMACConflicts(devices)...)
	conflicts = append(conflicts, detectHostnameConflicts(devices)...)
	conflicts = append(conflicts, detectDeviceTypeConflicts(devices)...)
	conflicts = append(conflicts, detectWiFiGatewayVendorConflict(devices)...)
	conflicts = append(conflicts, detectProbeConflicts(results)...)
	conflicts = append(conflicts, detectPublicAndVirtualConflicts(devices)...)
	return conflicts
}

func detectWiFiGatewayVendorConflict(devices []model.Device) []model.Conflict {
	var gateway, ap *model.Device
	for i := range devices {
		if devices[i].HasRole(model.RoleDefaultGateway) {
			gateway = &devices[i]
		}
		if devices[i].DeviceType == model.DeviceTypeAccessPoint || devices[i].HasRole(model.RoleWiFiAP) {
			ap = &devices[i]
		}
	}
	if gateway == nil || ap == nil || gateway.Vendor == "" || ap.Vendor == "" || strings.EqualFold(gateway.Vendor, ap.Vendor) {
		return nil
	}
	return []model.Conflict{{
		Type:        "wifi_bssid_vendor_differs_from_gateway_vendor",
		Severity:    model.ConflictLow,
		Devices:     []string{gateway.ID, ap.ID},
		Description: "Wi-Fi BSSID vendor differs from the default gateway vendor.",
		Effect:      "The access point may be separate from the router, so direct router/AP equivalence is not assumed.",
		Resolution:  "Represent the AP separately unless stronger evidence proves it is integrated with the gateway.",
		Evidence:    mergeEvidence(nil, gateway.Evidence, ap.Evidence),
	}}
}

func detectIPMACConflicts(devices []model.Device) []model.Conflict {
	ipToMACs := map[string]map[string][]string{}
	macToIPs := map[string]map[string][]string{}
	for _, d := range devices {
		for _, ip := range d.IPAddresses {
			if ipToMACs[ip] == nil {
				ipToMACs[ip] = map[string][]string{}
			}
			for _, mac := range normalizeMACs(d.MACAddresses) {
				ipToMACs[ip][mac] = append(ipToMACs[ip][mac], d.ID)
			}
		}
		for _, mac := range normalizeMACs(d.MACAddresses) {
			if macToIPs[mac] == nil {
				macToIPs[mac] = map[string][]string{}
			}
			for _, ip := range d.IPAddresses {
				macToIPs[mac][ip] = append(macToIPs[mac][ip], d.ID)
			}
		}
	}
	var conflicts []model.Conflict
	for ip, macs := range ipToMACs {
		if len(macs) > 1 {
			conflicts = append(conflicts, model.Conflict{
				Type:        "same_ip_different_macs",
				Severity:    model.ConflictHigh,
				Devices:     idsFromNested(macs),
				Description: fmt.Sprintf("IP address %s was observed with multiple MAC addresses.", ip),
				Effect:      "Identity may have changed during scan, or stale neighbor data was present.",
				Resolution:  "Prefer the most recent ARP/NDP observation and rerun neighbor discovery.",
			})
		}
	}
	for mac, ips := range macToIPs {
		if len(ips) > 1 {
			conflicts = append(conflicts, model.Conflict{
				Type:        "same_mac_multiple_ips",
				Severity:    model.ConflictLow,
				Devices:     idsFromNested(ips),
				Description: fmt.Sprintf("MAC address %s has multiple IP addresses.", mac),
				Effect:      "This is often normal for dual-stack, DHCP churn, or bridged devices.",
				Resolution:  "Merge by MAC but keep all IP addresses with source attribution.",
			})
		}
	}
	return conflicts
}

func detectHostnameConflicts(devices []model.Device) []model.Conflict {
	var conflicts []model.Conflict
	for _, d := range devices {
		if len(d.Hostnames) > 1 {
			conflicts = append(conflicts, model.Conflict{
				Type:        "hostname_mismatch",
				Severity:    model.ConflictLow,
				Devices:     []string{d.ID},
				Description: "Device has multiple hostnames from different probes.",
				Effect:      "Hostnames may be aliases, stale cache entries, or service-specific names.",
				Resolution:  "Display all hostnames and prefer DNS/mDNS recency for the primary label.",
				Evidence:    d.Evidence,
			})
		}
	}
	return conflicts
}

func detectDeviceTypeConflicts(devices []model.Device) []model.Conflict {
	var conflicts []model.Conflict
	for _, d := range devices {
		text := strings.ToLower(deviceText(d, d.Evidence))
		if d.DeviceType == model.DeviceTypeRouter && containsAny(text, "printer", "_ipp._tcp") {
			conflicts = append(conflicts, typeConflict(d, "router_printer_type_conflict", "Router and printer evidence both point at this device."))
		}
		if containsAny(text, "sysdescr", "switch", "bridge-mib") && containsAny(text, "router", "internetgatewaydevice") {
			conflicts = append(conflicts, typeConflict(d, "snmp_switch_http_router_conflict", "SNMP suggests switch while HTTP/UPnP suggests router."))
		}
	}
	return conflicts
}

func detectProbeConflicts(results []model.ProbeResult) []model.Conflict {
	var conflicts []model.Conflict
	reachability := map[string]string{}
	var gatewayChains []string
	for _, result := range results {
		for _, ev := range result.Evidence {
			if ev.Raw == nil {
				continue
			}
			if ip, _ := ev.Raw["gateway_ip"].(string); ip != "" {
				if reachable, ok := ev.Raw["reachable"].(bool); ok {
					val := fmt.Sprintf("%t", reachable)
					if old := reachability[ip]; old != "" && old != val {
						conflicts = append(conflicts, model.Conflict{
							Type:        "default_gateway_unreachable_but_http_reachable",
							Severity:    model.ConflictMedium,
							Devices:     []string{"ip_" + ip},
							Description: "Gateway reachability observations disagree.",
							Effect:      "Gateway availability confidence is reduced.",
							Resolution:  "Prefer active reachability checks over stale summary fields.",
							Evidence:    []model.Evidence{ev},
						})
					}
					reachability[ip] = val
				}
			}
			if chain, _ := ev.Raw["gateway_chain"].(string); chain != "" {
				gatewayChains = append(gatewayChains, chain)
			}
		}
	}
	if len(uniqueSorted(gatewayChains)) > 1 {
		conflicts = append(conflicts, model.Conflict{
			Type:        "gateway_chain_differs_between_probes",
			Severity:    model.ConflictMedium,
			Description: "Gateway chain differs between probes.",
			Effect:      "Upstream gateway topology is uncertain.",
			Resolution:  "Rerun route table and traceroute probes close together in time.",
		})
	}
	return conflicts
}

func detectPublicAndVirtualConflicts(devices []model.Device) []model.Conflict {
	var conflicts []model.Conflict
	for _, d := range devices {
		if d.DeviceType != model.DeviceTypeISPHop {
			for _, ip := range d.IPAddresses {
				if ip != "" && !safety.IsPrivateIPString(ip) {
					conflicts = append(conflicts, model.Conflict{
						Type:        "public_ip_classified_as_local_device",
						Severity:    model.ConflictHigh,
						Devices:     []string{d.ID},
						Description: fmt.Sprintf("Public IP %s was attached to a LAN device.", ip),
						Effect:      "The device is excluded from LAN topology confidence.",
						Resolution:  "Represent public addresses only as public IP context or ISP route hops.",
						Evidence:    d.Evidence,
					})
				}
			}
		}
		if d.DeviceType == model.DeviceTypeVirtualAdapter && d.HasRole(model.RoleUpstreamGateway) {
			conflicts = append(conflicts, model.Conflict{
				Type:        "virtual_adapter_mistaken_as_physical_upstream",
				Severity:    model.ConflictHigh,
				Devices:     []string{d.ID},
				Description: "A virtual adapter was marked as an upstream gateway.",
				Effect:      "Physical topology confidence is reduced.",
				Resolution:  "Use route/interface metrics and ignore virtual adapters for upstream physical links.",
				Evidence:    d.Evidence,
			})
		}
	}
	return conflicts
}

func typeConflict(d model.Device, typ, description string) model.Conflict {
	return model.Conflict{
		Type:        typ,
		Severity:    model.ConflictMedium,
		Devices:     []string{d.ID},
		Description: description,
		Effect:      "Device type confidence is downgraded.",
		Resolution:  "Prefer protocol-specific evidence with stronger source attribution, such as SNMP/LLDP for switches and UPnP IGD/default route for routers.",
		Evidence:    d.Evidence,
	}
}

func idsFromNested(values map[string][]string) []string {
	seen := map[string]bool{}
	var out []string
	for _, ids := range values {
		for _, id := range ids {
			if id != "" && !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return uniqueSorted(out)
}

func containsString(values []string, needle string) bool {
	for _, v := range values {
		if v == needle {
			return true
		}
	}
	return false
}
