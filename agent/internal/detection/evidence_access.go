package detection

import "github.com/thekiran/iad/pkg/models"

// Helpers for reading the loosely-typed ProbeResult.Evidence maps. Values are
// native Go types in-memory but may be JSON-decoded ([]any, json.Number) when a
// result is loaded from a fixture, so each getter tolerates both.

func getString(ev map[string]any, key string) string {
	if v, ok := ev[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getBool(ev map[string]any, key string) bool {
	if v, ok := ev[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getFloat(ev map[string]any, key string) float64 {
	v, ok := ev[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return 0
}

func getStrings(ev map[string]any, key string) []string {
	v, ok := ev[key]
	if !ok {
		return nil
	}
	switch arr := v.(type) {
	case []string:
		return arr
	case []any:
		out := make([]string, 0, len(arr))
		for _, e := range arr {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// getStringMap reads a map[string]string, tolerating the map[string]any form a
// JSON-decoded fixture produces.
func getStringMap(ev map[string]any, key string) map[string]string {
	v, ok := ev[key]
	if !ok {
		return nil
	}
	switch m := v.(type) {
	case map[string]string:
		return m
	case map[string]any:
		out := make(map[string]string, len(m))
		for k, e := range m {
			if s, ok := e.(string); ok {
				out[k] = s
			}
		}
		return out
	}
	return nil
}

func getGatewayDevices(ev map[string]any, key string) []models.GatewayDevice {
	v, ok := ev[key]
	if !ok {
		return nil
	}
	switch arr := v.(type) {
	case []models.GatewayDevice:
		return arr
	case []any:
		out := make([]models.GatewayDevice, 0, len(arr))
		for _, e := range arr {
			if m, ok := e.(map[string]any); ok {
				out = append(out, gatewayDeviceFromMap(m))
			}
		}
		return out
	}
	return nil
}

func getWANSignals(ev map[string]any, key string) []models.WANSignal {
	v, ok := ev[key]
	if !ok {
		return nil
	}
	switch arr := v.(type) {
	case []models.WANSignal:
		return arr
	case []any:
		out := make([]models.WANSignal, 0, len(arr))
		for _, e := range arr {
			if m, ok := e.(map[string]any); ok {
				out = append(out, wanSignalFromMap(m))
			}
		}
		return out
	}
	return nil
}

func getAccessArchitecture(ev map[string]any, key string) *models.AccessArchitecture {
	v, ok := ev[key]
	if !ok {
		return nil
	}
	switch a := v.(type) {
	case models.AccessArchitecture:
		return &a
	case map[string]any:
		return &models.AccessArchitecture{
			LocalMedium:    anyString(a["local_medium"]),
			WANMedium:      anyString(a["wan_medium"]),
			IPArchitecture: anyString(a["ip_architecture"]),
			NATTopology:    anyString(a["nat_topology"]),
			LikelyCPERole:  anyString(a["likely_cpe_role"]),
		}
	default:
		return nil
	}
}

func getIPv6Context(ev map[string]any, key string) *models.IPv6Context {
	v, ok := ev[key]
	if !ok {
		return nil
	}
	switch c := v.(type) {
	case models.IPv6Context:
		return &c
	case map[string]any:
		return &models.IPv6Context{
			IPv6Available:   anyBool(c["ipv6_available"]),
			GlobalIPv6:      anyBool(c["global_ipv6"]),
			DefaultRoute:    anyString(c["default_route"]),
			DNS64NAT64:      anyBool(c["dns64_nat64"]),
			TransitionHints: anyStrings(c["transition_hints"]),
		}
	default:
		return nil
	}
}

func getNATTopology(ev map[string]any, key string) *models.NATTopology {
	v, ok := ev[key]
	if !ok {
		return nil
	}
	switch n := v.(type) {
	case models.NATTopology:
		return &n
	case map[string]any:
		return &models.NATTopology{
			PublicIP:                   anyString(n["public_ip"]),
			STUNPublicIP:               anyString(n["stun_public_ip"]),
			STUNPublicPort:             int(anyFloat(n["stun_public_port"])),
			PublicIPMatches:            anyBool(n["public_ip_matches"]),
			CGNAT:                      anyBool(n["cgnat"]),
			DoubleNAT:                  anyBool(n["double_nat"]),
			InternalDoubleNATPossible:  anyBool(n["internal_double_nat_possible"]),
			ExternalPublicIPConsistent: anyBool(n["external_public_ip_consistent"]),
			GatewayNATControlReachable: anyBool(n["gateway_nat_control_reachable"]),
			PCPReachable:               anyBool(n["pcp_reachable"]),
			NATPMPReachable:            anyBool(n["nat_pmp_reachable"]),
			Topology:                   anyString(n["topology"]),
			Notes:                      anyStrings(n["notes"]),
		}
	default:
		return nil
	}
}

func getPerformanceProfile(ev map[string]any, key string) *models.PerformanceProfile {
	v, ok := ev[key]
	if !ok {
		return nil
	}
	switch p := v.(type) {
	case models.PerformanceProfile:
		return &p
	case map[string]any:
		return &models.PerformanceProfile{
			Target:          anyString(p["target"]),
			Method:          anyString(p["method"]),
			IdleLatencyMS:   anyFloat(p["idle_latency_ms"]),
			JitterMS:        anyFloat(p["jitter_ms"]),
			PacketLossPct:   anyFloat(p["packet_loss_pct"]),
			LoadedLatencyMS: anyFloat(p["loaded_latency_ms"]),
		}
	default:
		return nil
	}
}

func gatewayDeviceFromMap(m map[string]any) models.GatewayDevice {
	return models.GatewayDevice{
		IP:                         anyString(m["ip"]),
		Role:                       anyString(m["role"]),
		ReachableState:             anyString(m["reachable_state"]),
		Reachable:                  anyBool(m["reachable"]),
		OpenPorts:                  anyInts(m["open_ports"]),
		HTTPObservations:           anyHTTPObservations(m["http_observations"]),
		HTTPTitle:                  anyString(m["http_title"]),
		ServerHeader:               anyString(m["server_header"]),
		WWWAuthenticate:            anyString(m["www_authenticate"]),
		FaviconHash:                anyString(m["favicon_hash"]),
		WWWAuthRealm:               anyString(m["www_authenticate_realm"]),
		RedirectLocation:           anyString(m["redirect_location"]),
		RedirectPath:               anyString(m["redirect_path"]),
		TLSObservations:            anyTLSObservations(m["tls_observations"]),
		TLSCertCN:                  anyString(m["tls_cert_cn"]),
		TLSCertSANs:                anyStrings(m["tls_cert_sans"]),
		TLSCertIssuer:              anyString(m["tls_cert_issuer"]),
		TLSServerName:              anyString(m["tls_server_name"]),
		HTMLMetaGenerator:          anyString(m["html_meta_generator"]),
		LoginLabels:                anyStrings(m["login_labels"]),
		UPnPState:                  anyString(m["upnp_state"]),
		UPnPFound:                  anyBool(m["upnp_found"]),
		UPnPIGDFound:               anyBool(m["upnp_igd_found"]),
		WANCommonInterfaceFound:    anyBool(m["wan_common_interface_found"]),
		WANAccessType:              anyString(m["wan_access_type"]),
		PhysicalLinkStatus:         anyString(m["physical_link_status"]),
		Layer1UpstreamMaxBitRate:   int64(anyFloat(m["layer1_upstream_max_bitrate"])),
		Layer1DownstreamMaxBitRate: int64(anyFloat(m["layer1_downstream_max_bitrate"])),
		TR064State:                 anyString(m["tr064_state"]),
		TR064Found:                 anyBool(m["tr064_found"]),
		TR064AuthRequired:          anyBool(m["tr064_auth_required"]),
		TR064Services:              anyStrings(m["tr064_services"]),
		SNMPState:                  anyString(m["snmp_state"]),
		MACVendor:                  anyString(m["mac_vendor"]),
		CPEModelGuess:              anyString(m["cpe_model_guess"]),
		Model:                      anyString(m["model"]),
		Manufacturer:               anyString(m["manufacturer"]),
		FingerprintID:              anyString(m["fingerprint_id"]),
		AccessEvidence:             anyGatewayAccessEvidence(m["access_evidence"]),
		AccessHints:                anyStrings(m["access_hints"]),
		PhysicalHints:              anyStrings(m["physical_hints"]),
		Notes:                      anyStrings(m["notes"]),
		FailedAttempts:             anyProbeAttempts(m["failed_attempts"]),
		DeviceConfidence:           anyFloat(m["device_confidence"]),
		AccessConfidence:           anyFloat(m["access_confidence"]),
		Confidence:                 anyFloat(m["confidence"]),
		EvidenceIDs:                anyStrings(m["evidence_ids"]),
	}
}

func wanSignalFromMap(m map[string]any) models.WANSignal {
	return models.WANSignal{
		Source:     anyString(m["source"]),
		IP:         anyString(m["ip"]),
		Type:       anyString(m["type"]),
		Value:      anyString(m["value"]),
		Strength:   anyString(m["strength"]),
		Detail:     anyString(m["detail"]),
		Confidence: anyFloat(m["confidence"]),
	}
}

func anyString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func anyBool(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func anyFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

func anyStrings(v any) []string {
	switch arr := v.(type) {
	case []string:
		return arr
	case []any:
		out := make([]string, 0, len(arr))
		for _, e := range arr {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func anyInts(v any) []int {
	switch arr := v.(type) {
	case []int:
		return arr
	case []any:
		out := make([]int, 0, len(arr))
		for _, e := range arr {
			if n := int(anyFloat(e)); n != 0 {
				out = append(out, n)
			}
		}
		return out
	default:
		return nil
	}
}

func anyHTTPObservations(v any) []models.HTTPObservation {
	switch arr := v.(type) {
	case []models.HTTPObservation:
		return arr
	case []any:
		out := make([]models.HTTPObservation, 0, len(arr))
		for _, e := range arr {
			if m, ok := e.(map[string]any); ok {
				out = append(out, models.HTTPObservation{
					Source: anyString(m["source"]), URL: anyString(m["url"]), Method: anyString(m["method"]),
					StatusCode: int(anyFloat(m["status_code"])), Title: anyString(m["title"]),
					ServerHeader: anyString(m["server_header"]), WWWAuthenticate: anyString(m["www_authenticate"]),
					WWWAuthRealm: anyString(m["www_authenticate_realm"]), RedirectLocation: anyString(m["redirect_location"]),
					RedirectPath: anyString(m["redirect_path"]), FaviconHash: anyString(m["favicon_hash"]),
					HTMLMetaGenerator: anyString(m["html_meta_generator"]), LoginLabels: anyStrings(m["login_labels"]),
					EvidenceID: anyString(m["evidence_id"]),
				})
			}
		}
		return out
	default:
		return nil
	}
}

func anyTLSObservations(v any) []models.TLSObservation {
	switch arr := v.(type) {
	case []models.TLSObservation:
		return arr
	case []any:
		out := make([]models.TLSObservation, 0, len(arr))
		for _, e := range arr {
			if m, ok := e.(map[string]any); ok {
				out = append(out, models.TLSObservation{
					Source: anyString(m["source"]), IP: anyString(m["ip"]), Port: int(anyFloat(m["port"])),
					CN: anyString(m["cn"]), SANs: anyStrings(m["sans"]), Issuer: anyString(m["issuer"]),
					ServerName: anyString(m["server_name"]), EvidenceID: anyString(m["evidence_id"]),
				})
			}
		}
		return out
	default:
		return nil
	}
}

func anyGatewayAccessEvidence(v any) []models.GatewayAccessEvidence {
	switch arr := v.(type) {
	case []models.GatewayAccessEvidence:
		return arr
	case []any:
		out := make([]models.GatewayAccessEvidence, 0, len(arr))
		for _, e := range arr {
			if m, ok := e.(map[string]any); ok {
				out = append(out, models.GatewayAccessEvidence{
					Source: anyString(m["source"]), Type: anyString(m["type"]), Value: anyString(m["value"]),
					Strength: anyString(m["strength"]), Confidence: anyFloat(m["confidence"]),
					Hints: anyStrings(m["hints"]), EvidenceID: anyString(m["evidence_id"]),
				})
			}
		}
		return out
	default:
		return nil
	}
}

func anyProbeAttempts(v any) []models.ProbeAttempt {
	switch arr := v.(type) {
	case []models.ProbeAttempt:
		return arr
	case []any:
		out := make([]models.ProbeAttempt, 0, len(arr))
		for _, e := range arr {
			if m, ok := e.(map[string]any); ok {
				out = append(out, models.ProbeAttempt{
					Source: anyString(m["source"]), Target: anyString(m["target"]), Protocol: anyString(m["protocol"]),
					Port: int(anyFloat(m["port"])), URL: anyString(m["url"]), Method: anyString(m["method"]),
					Error: anyString(m["error"]), Timeout: anyBool(m["timeout"]), EvidenceID: anyString(m["evidence_id"]),
				})
			}
		}
		return out
	default:
		return nil
	}
}
