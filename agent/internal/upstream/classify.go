// Package upstream runs a dedicated, read-only enrichment phase for upstream
// gateway / CPE candidates (e.g. an off-subnet 192.168.1.1 discovered via the
// gateway chain) and classifies them from weighted evidence.
//
// SAFETY CONTRACT
//   - Only private/local/authorized targets are probed.
//   - No brute force, no credential guessing, no logins, no exploit checks, no
//     configuration changes, and no intrusive scripts.
//   - SNMP / SSH / router APIs / TR-064 / TR-181 stay opt-in elsewhere; this
//     package never authenticates.
//   - Classification NEVER asserts a strong physical role from an IP pattern
//     alone — it requires reachability/service/UPnP/OUI evidence and otherwise
//     marks the result inferred/unknown.
package upstream

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

// Classification tags. These are the UI-facing labels requested by the spec.
const (
	TagRouter            = "ROUTER"
	TagGateway           = "GATEWAY"
	TagUpstreamGateway   = "UPSTREAM_GATEWAY"
	TagPossibleCPE       = "POSSIBLE_CPE"
	TagISPCPE            = "ISP_CPE"
	TagDoubleNATUpstream = "DOUBLE_NAT_UPSTREAM"
	TagModem             = "MODEM"
	TagONT               = "ONT"
	TagAccessPoint       = "ACCESS_POINT"
	TagUnknownInfra      = "UNKNOWN_INFRASTRUCTURE"
	TagVirtualOrDocker   = "VIRTUAL_OR_DOCKER_ARTIFACT"
)

// Facts is the evidence gathered about one upstream candidate. Pure inputs —
// the orchestrator (enrich.go) fills these from safe probes; Classify never
// touches the network.
type Facts struct {
	IP                string
	IsPrivate         bool
	SameSubnetAsAgent bool
	IsDefaultGateway  bool
	InGatewayChain    bool
	HopIndex          int // position among private hops (0 = default gateway)
	HopDistance       int // hops from agent (0 = unknown)
	DoubleNATHint     bool

	ReachableICMP bool
	ReachableTCP  bool

	OpenPorts      []int
	HTTPVendor     string
	HTTPAdminPanel bool
	RouterLikeHTTP bool
	HasUPnPRoot    bool
	HasDNSService  bool
	HasCWMP        bool // TR-069 / CWMP management port 7547 open
	MACVendor      string
	VirtualHint    bool
	ONTHint        bool
	ModemHint      bool

	Now time.Time
}

// Classification is the weighted result for one device.
type Classification struct {
	Role       string
	Confidence float64
	Tags       []string
	Evidence   []models.IntelEvidenceItem
	Warnings   []string
}

// Evidence weights. Each present signal contributes its weight to the device
// confidence (summed, then clamped to [0,1]). Reachability + router-like HTTP +
// UPnP are the strongest because they are direct, observed facts.
const (
	wReachICMP    = 0.18
	wReachTCP     = 0.12
	wInTraceroute = 0.15
	wPrivateHop   = 0.15
	wHTTPAdmin    = 0.12
	wRouterHTTP   = 0.18
	wUPnPRoot     = 0.15
	wGatewayOUI   = 0.12
	wDNSService   = 0.10
	wCWMP         = 0.12
	wOpenServices = 0.06
)

// Classify turns gathered Facts into a weighted Classification. This is the
// single place confidence is computed; it is deterministic and side-effect free.
func Classify(f Facts) Classification {
	now := f.Now
	if now.IsZero() {
		now = time.Now()
	}
	ts := now.UTC().Format(time.RFC3339)

	var evidence []models.IntelEvidenceItem
	score := 0.0
	add := func(typ, val, src string, impact float64) {
		evidence = append(evidence, models.IntelEvidenceItem{Type: typ, Value: val, Source: src, ConfidenceImpact: impact, Timestamp: ts})
		score += impact
	}

	if f.ReachableICMP {
		add("reachable_by_icmp", "true", "ping", wReachICMP)
	}
	if f.ReachableTCP {
		add("reachable_by_tcp", "true", "tcp_connect", wReachTCP)
	}
	if f.InGatewayChain {
		add("visible_in_traceroute", "true", "gateway_chain", wInTraceroute)
	}
	if f.IsPrivate && f.InGatewayChain && !f.IsDefaultGateway {
		add("private_upstream_hop", "true", "gateway_chain", wPrivateHop)
	}
	if f.HTTPAdminPanel {
		add("has_http_admin_page", "true", "http_fingerprint", wHTTPAdmin)
	}
	if f.RouterLikeHTTP {
		val := "router-like HTTP title/header"
		if f.HTTPVendor != "" {
			val = f.HTTPVendor
		}
		add("router_like_http", val, "http_fingerprint", wRouterHTTP)
	}
	if f.HasUPnPRoot {
		add("upnp_root_device", "true", "ssdp", wUPnPRoot)
	}
	if f.MACVendor != "" && isNetworkVendor(f.MACVendor) {
		add("gateway_router_mac_oui", f.MACVendor, "arp_oui", wGatewayOUI)
	}
	if f.HasDNSService {
		add("dns_service", "53/tcp", "tcp_connect", wDNSService)
	}
	if f.HasCWMP {
		add("tr069_cwmp_port", "7547", "tcp_connect", wCWMP)
	}
	if len(f.OpenPorts) > 0 {
		add("open_services", joinInts(f.OpenPorts), "tcp_connect", wOpenServices)
	}

	confidence := clamp01(score)
	tags, role, warnings := deriveTags(f, confidence)

	return Classification{Role: role, Confidence: confidence, Tags: tags, Evidence: evidence, Warnings: warnings}
}

func deriveTags(f Facts, confidence float64) (tags []string, role string, warnings []string) {
	addTag := func(t string) { tags = appendUnique(tags, t) }

	switch {
	case f.VirtualHint:
		addTag(TagVirtualOrDocker)
		role = "virtual_or_docker_artifact"
	case f.IsDefaultGateway && f.SameSubnetAsAgent:
		addTag(TagGateway)
		addTag(TagRouter)
		role = "gateway_router"
	case f.IsPrivate && f.InGatewayChain && !f.IsDefaultGateway:
		addTag(TagUpstreamGateway)
		addTag(TagPossibleCPE)
		role = "upstream_private_gateway"
	case f.IsPrivate:
		addTag(TagUnknownInfra)
		role = "unknown_infrastructure"
	}

	// Double NAT: a second private gateway appears upstream of the default one.
	if f.DoubleNATHint && !f.IsDefaultGateway && f.IsPrivate {
		addTag(TagDoubleNATUpstream)
	}
	// ISP CPE: TR-069/CWMP management or an ISP-style router admin surface.
	if f.HasCWMP || (f.RouterLikeHTTP && isISPVendor(f.HTTPVendor)) {
		addTag(TagISPCPE)
	}
	// Modem/ONT only when an actual fingerprint says so — never invented.
	if f.ONTHint {
		addTag(TagONT)
	}
	if f.ModemHint {
		addTag(TagModem)
	}

	if !f.ReachableICMP && !f.ReachableTCP {
		warnings = append(warnings, "Not directly reachable — classification is inferred from routing evidence only.")
		if f.IsPrivate && f.InGatewayChain {
			addTag(TagUpstreamGateway)
			addTag(TagPossibleCPE)
		}
	}

	// Evidence requirement: never assert a confident physical role from the IP
	// pattern alone. Below the floor, flag the result as inferred/unknown.
	if confidence < 0.30 {
		warnings = append(warnings, "Low evidence — treat this as inferred, not confirmed.")
		if len(tags) == 0 {
			addTag(TagUnknownInfra)
		}
	}
	if len(tags) == 0 {
		addTag(TagUnknownInfra)
	}
	if role == "" {
		role = "unknown_infrastructure"
	}
	sort.Strings(tags)
	return tags, role, warnings
}

// networkVendors are OUI vendor substrings that indicate router/gateway/AP
// hardware makers. Used only to weight evidence, never to invent a model name.
var networkVendors = []string{
	"tp-link", "tplink", "keenetic", "xiaomi", "huawei", "zte", "zyxel", "asus",
	"asustek", "mikrotik", "tenda", "netis", "fiberhome", "nokia", "alcatel",
	"technicolor", "arris", "ubee", "sagemcom", "ubiquiti", "netgear",
	"d-link", "dlink", "cisco", "juniper", "mercusys", "totolink", "draytek",
}

func isNetworkVendor(vendor string) bool {
	v := strings.ToLower(vendor)
	for _, n := range networkVendors {
		if strings.Contains(v, n) {
			return true
		}
	}
	return false
}

// ispVendors are makers commonly shipped as ISP-provided CPE.
var ispVendors = []string{"huawei", "zte", "fiberhome", "nokia", "alcatel", "technicolor", "arris", "ubee", "sagemcom", "zyxel"}

func isISPVendor(vendor string) bool {
	v := strings.ToLower(vendor)
	for _, n := range ispVendors {
		if strings.Contains(v, n) {
			return true
		}
	}
	return false
}

func joinInts(values []int) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, strconv.Itoa(v))
	}
	return strings.Join(parts, ",")
}

func appendUnique(values []string, next ...string) []string {
	seen := make(map[string]bool, len(values)+len(next))
	out := values
	for _, v := range values {
		seen[v] = true
	}
	for _, v := range next {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
