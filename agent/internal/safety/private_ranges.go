package safety

import (
	"net"
)

func IsPrivateIPString(s string) bool {
	ip := net.ParseIP(s)
	return IsPrivateIP(ip)
}

func IsPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		switch {
		case v4[0] == 10:
			return true
		case v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31:
			return true
		case v4[0] == 192 && v4[1] == 168:
			return true
		default:
			return false
		}
	}
	return len(ip) == net.IPv6len && (ip[0]&0xfe) == 0xfc
}

func IsPrivateCIDR(cidr string) bool {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return IsPrivateIP(ip) && IsPrivateIP(network.IP)
}

func FilterPrivateIPs(ips []string) []string {
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		if IsPrivateIPString(ip) {
			out = append(out, ip)
		}
	}
	return out
}
