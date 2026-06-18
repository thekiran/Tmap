package discovery

import (
	"net"
	"sort"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

// virtualIfaceTokens are substrings (case-insensitive) that mark an interface as
// virtual/overlay — Docker bridges, VM host adapters, VPN tunnels, etc. These are
// ignored by default so the topology reflects the real LAN, not virtual plumbing.
var virtualIfaceTokens = []string{
	"veth", "docker", "br-", "virbr", "vmnet", "vmware", "vbox", "virtualbox",
	"vethernet", "hyper-v", "tap", "tun", "tailscale", "zerotier", "wg", "utun",
	"awdl", "llw", "ppp", "isatap", "teredo", "bluetooth", "loopback", "pseudo",
}

// Interfaces lists the host's network interfaces, classified. Uses only the
// standard library so it behaves identically on Windows, Linux and macOS.
func Interfaces() ([]models.InterfaceInfo, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	out := make([]models.InterfaceInfo, 0, len(ifaces))
	for _, ifc := range ifaces {
		info := models.InterfaceInfo{
			Name:     ifc.Name,
			MAC:      ifc.HardwareAddr.String(),
			Up:       ifc.Flags&net.FlagUp != 0,
			Loopback: ifc.Flags&net.FlagLoopback != 0,
			Virtual:  isVirtualName(ifc.Name),
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			addr := models.IPAddress{IP: ipnet.IP.String(), Version: ipVersion(ipnet.IP)}
			if info.CIDR == "" && ipnet.IP.To4() != nil && !ipnet.IP.IsLoopback() {
				// First usable IPv4 network defines the interface's CIDR.
				network := &net.IPNet{IP: ipnet.IP.Mask(ipnet.Mask), Mask: ipnet.Mask}
				addr.CIDR = network.String()
				info.CIDR = network.String()
			}
			info.Addresses = append(info.Addresses, addr)
		}
		out = append(out, info)
	}
	return out, nil
}

// SelectPrimary chooses the interface the scan should use by default: an up,
// non-loopback, non-virtual interface that has a private IPv4 network. If a
// gateway IP is known, an interface whose network contains it is preferred. The
// chosen interface is marked Selected=true in the returned slice (a copy).
func SelectPrimary(ifaces []models.InterfaceInfo, gatewayIP string, includeVirtual bool) (models.InterfaceInfo, []models.InterfaceInfo, bool) {
	result := make([]models.InterfaceInfo, len(ifaces))
	copy(result, ifaces)

	bestIdx := -1
	bestScore := -1
	for i, ifc := range result {
		if !ifc.Up || ifc.Loopback {
			continue
		}
		if ifc.Virtual && !includeVirtual {
			continue
		}
		if ifc.CIDR == "" {
			continue
		}
		score := 0
		switch {
		case hasRoutablePrivateIPv4(ifc):
			// A real RFC1918 LAN (10/8, 172.16/12, 192.168/16): the normal case.
			score += 4
		case hasLinkLocalIPv4(ifc):
			// 169.254/16 (APIPA) means DHCP failed / the link is unconfigured.
			// Such an interface is a dead end, so it is only ever chosen as a
			// last resort — never over a routable LAN.
			score += 1
		}
		if gatewayIP != "" && cidrContains(ifc.CIDR, gatewayIP) {
			score += 10
		}
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return models.InterfaceInfo{}, result, false
	}
	result[bestIdx].Selected = true
	return result[bestIdx], result, true
}

// FindByName returns the named interface and marks it selected.
func FindByName(ifaces []models.InterfaceInfo, name string) (models.InterfaceInfo, []models.InterfaceInfo, bool) {
	result := make([]models.InterfaceInfo, len(ifaces))
	copy(result, ifaces)
	for i := range result {
		if result[i].Name == name {
			result[i].Selected = true
			return result[i], result, true
		}
	}
	return models.InterfaceInfo{}, result, false
}

func isVirtualName(name string) bool {
	l := strings.ToLower(name)
	for _, tok := range virtualIfaceTokens {
		if strings.Contains(l, tok) {
			return true
		}
	}
	return false
}

// hasRoutablePrivateIPv4 reports whether the interface has an RFC1918 IPv4
// address (10/8, 172.16/12, 192.168/16). Link-local (169.254/16) is deliberately
// excluded: it is not a routable LAN and must not look like one when selecting
// the primary interface.
func hasRoutablePrivateIPv4(ifc models.InterfaceInfo) bool {
	for _, a := range ifc.Addresses {
		ip := net.ParseIP(a.IP)
		if ip == nil || ip.To4() == nil {
			continue
		}
		if isPrivateIP(ip) && !ip.IsLinkLocalUnicast() {
			return true
		}
	}
	return false
}

// hasLinkLocalIPv4 reports whether the interface has only an APIPA (169.254/16)
// IPv4 address — the signature of a failed DHCP lease or an unconfigured link.
func hasLinkLocalIPv4(ifc models.InterfaceInfo) bool {
	for _, a := range ifc.Addresses {
		if ip := net.ParseIP(a.IP); ip != nil && ip.To4() != nil && ip.IsLinkLocalUnicast() {
			return true
		}
	}
	return false
}

func cidrContains(cidr, ipStr string) bool {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(ipStr)
	return ip != nil && ipnet.Contains(ip)
}

func ipVersion(ip net.IP) int {
	if ip.To4() != nil {
		return 4
	}
	return 6
}

// primaryIPv4 returns the interface's first private IPv4 address (the agent's own
// address on the scanned network), or "".
func primaryIPv4(ifc models.InterfaceInfo) string {
	var candidates []string
	for _, a := range ifc.Addresses {
		if ip := net.ParseIP(a.IP); ip != nil && ip.To4() != nil && isPrivateIP(ip) {
			candidates = append(candidates, a.IP)
		}
	}
	sort.Strings(candidates)
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}
