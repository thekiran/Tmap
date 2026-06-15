package detection

import (
	"net"
	"strings"

	"github.com/thekiran/iad/internal/network"
)

// AdapterInfo is the engine-side view of a network adapter. It is decoupled from
// network.Adapter so it can also be built from JSON fixtures.
type AdapterInfo struct {
	Name  string
	Up    bool
	Addrs []string
}

// MainAdapterInfo is the chosen "real" adapter the host uses to reach the LAN.
type MainAdapterInfo struct {
	Name   string
	Access string // Ethernet | Wi-Fi | Cellular | Unknown
	IP     string
}

// virtualAdapterKeywords are substrings that mark an adapter as virtual/helper
// rather than the host's real uplink. Such adapters must never be treated as
// evidence of an access type.
var virtualAdapterKeywords = []string{
	"vmware", "virtualbox", "hyper-v", "vethernet", "bluetooth", "loopback",
	"docker", "wsl", "tailscale", "zerotier", "warp", "hamachi", "tap",
	"npcap", "wireguard", "openvpn", "vpn", "virtual",
}

// isVirtualAdapter reports whether name looks like a virtual/helper adapter.
func isVirtualAdapter(name string) bool {
	n := strings.ToLower(name)
	for _, kw := range virtualAdapterKeywords {
		if strings.Contains(n, kw) {
			return true
		}
	}
	return false
}

// isAPIPA reports whether ip is an automatic private (link-local) address
// 169.254.0.0/16 — i.e. "no real connectivity".
func isAPIPA(ip string) bool {
	parsed := net.ParseIP(stripCIDR(ip))
	if parsed == nil {
		return false
	}
	return parsed.IsLinkLocalUnicast() && parsed.To4() != nil
}

// isPrivateIPv4 reports whether ip is an RFC1918 private IPv4 address.
func isPrivateIPv4(ip string) bool {
	parsed := net.ParseIP(stripCIDR(ip))
	if parsed == nil || parsed.To4() == nil {
		return false
	}
	return parsed.IsPrivate()
}

// inferLocalAccess maps an adapter name to how the host is locally attached. This
// is NOT the internet access type: Ethernet does not imply fiber, Wi-Fi does not
// imply mobile. It only describes the first local hop.
func inferLocalAccess(name string) string {
	n := strings.ToLower(name)
	switch {
	case containsAny(n, "wi-fi", "wifi", "wlan", "wireless", "kablosuz"):
		return "Wi-Fi"
	case containsAny(n, "cellular", "wwan", "mobile", "lte", "5g"):
		return "Cellular"
	case containsAny(n, "ethernet", "eth", "local area", "yerel ağ", "gigabit"):
		return "Ethernet"
	default:
		return "Unknown"
	}
}

// pickMainAdapter selects the host's real uplink adapter: up, not virtual, and
// holding a usable (non-APIPA) IPv4. The first qualifying adapter wins.
func pickMainAdapter(adapters []AdapterInfo) MainAdapterInfo {
	for _, a := range adapters {
		if !a.Up || isVirtualAdapter(a.Name) {
			continue
		}
		if ip := usableIPv4(a.Addrs); ip != "" {
			return MainAdapterInfo{Name: a.Name, Access: inferLocalAccess(a.Name), IP: ip}
		}
	}
	return MainAdapterInfo{}
}

// usableIPv4 returns the first non-APIPA IPv4 address in addrs, or "".
func usableIPv4(addrs []string) string {
	for _, ad := range addrs {
		ip := stripCIDR(ad)
		parsed := net.ParseIP(ip)
		if parsed == nil || parsed.To4() == nil || isAPIPA(ip) {
			continue
		}
		return ip
	}
	return ""
}

func stripCIDR(s string) string {
	if i := strings.IndexByte(s, '/'); i >= 0 {
		return s[:i]
	}
	return s
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// toAdapterInfos normalizes the adapter_probe evidence (native []network.Adapter
// in the live path, []any of maps when decoded from a JSON fixture) into a
// uniform []AdapterInfo.
func toAdapterInfos(v any) []AdapterInfo {
	switch arr := v.(type) {
	case []AdapterInfo:
		return arr
	case []network.Adapter:
		out := make([]AdapterInfo, 0, len(arr))
		for _, a := range arr {
			out = append(out, AdapterInfo{Name: a.Name, Up: a.Up, Addrs: a.Addrs})
		}
		return out
	case []any:
		out := make([]AdapterInfo, 0, len(arr))
		for _, e := range arr {
			m, ok := e.(map[string]any)
			if !ok {
				continue
			}
			ai := AdapterInfo{Name: getString(m, "name"), Up: getBool(m, "up")}
			if addrs, ok := m["addrs"].([]any); ok {
				for _, x := range addrs {
					if s, ok := x.(string); ok {
						ai.Addrs = append(ai.Addrs, s)
					}
				}
			}
			out = append(out, ai)
		}
		return out
	}
	return nil
}
