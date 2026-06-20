package ippool

import (
	"encoding/binary"
	"net"
)

// CandidateGenerator produces safe, prioritized host candidates inside an
// authorized private subnet. It never emits public, network, or broadcast
// addresses, and it orders likely-infrastructure IPs first so discovery finds
// gateways/servers before sweeping the rest of the subnet.
type CandidateGenerator struct {
	guard ScopeGuard
}

func NewCandidateGenerator(guard ScopeGuard) CandidateGenerator {
	return CandidateGenerator{guard: guard}
}

// GenerateFromSeed builds candidates for the subnet that contains seedIP. If
// cidr is empty, a /24 around the seed is assumed (the default auto scope).
// `limit` caps the number returned (0 = subnet default, bounded by host count).
//
// Priority order:
//  1. Common gateway/infra hosts: .1, .254, .253, .252, .2
//  2. Near neighbors of the seed (seed±1, ±2, …)
//  3. DHCP-typical range .100–.200
//  4. The remaining hosts in ascending order
//
// Network and broadcast addresses are skipped; everything is re-checked through
// the ScopeGuard so nothing outside private scope can leak in.
func (c CandidateGenerator) GenerateFromSeed(seedIP, cidr string, limit int) []string {
	ip := net.ParseIP(seedIP).To4()
	if ip == nil {
		return nil
	}
	var ipNet *net.IPNet
	if cidr != "" {
		if _, n, err := net.ParseCIDR(cidr); err == nil {
			ipNet = n
		}
	}
	if ipNet == nil {
		// Default to a /24 around the seed.
		mask := net.CIDRMask(24, 32)
		ipNet = &net.IPNet{IP: ip.Mask(mask), Mask: mask}
	}

	ones, _ := ipNet.Mask.Size()
	// Refuse to enumerate anything bigger than the guard allows without it being
	// pre-approved; the manager handles confirmation, the generator stays safe.
	if ones < c.guard.MaxAutoPrefix {
		// For larger subnets we still only emit the prioritized infra + DHCP +
		// near-neighbor set (a bounded, useful slice), never the full sweep.
		return c.boundedLargeSubnet(seedIP, ipNet, limit)
	}

	base := binary.BigEndian.Uint32(ipNet.IP.To4())
	hostBits := 32 - ones
	size := uint32(1) << hostBits
	networkAddr := base
	broadcastAddr := base + size - 1

	seen := map[uint32]bool{networkAddr: true, broadcastAddr: true}
	var ordered []uint32
	add := func(host uint32) {
		if host <= networkAddr || host >= broadcastAddr || seen[host] {
			return
		}
		seen[host] = true
		ordered = append(ordered, host)
	}

	// 1) Likely infrastructure.
	for _, last := range []uint32{1, 254, 253, 252, 2} {
		add(networkAddr + last)
	}
	// 2) Near neighbors of the seed.
	seed := binary.BigEndian.Uint32(ip)
	for d := uint32(1); d <= 4; d++ {
		add(seed + d)
		add(seed - d)
	}
	add(seed)
	// 3) DHCP-typical range .100–.200 (clamped to the subnet).
	for last := uint32(100); last <= 200; last++ {
		add(networkAddr + last)
	}
	// 4) The rest, ascending.
	for host := networkAddr + 1; host < broadcastAddr; host++ {
		add(host)
	}

	out := make([]string, 0, len(ordered))
	for _, host := range ordered {
		s := uint32ToIP(host).String()
		if ok, _ := c.guard.AllowIP(s); !ok {
			continue
		}
		out = append(out, s)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

// boundedLargeSubnet returns only the high-value, bounded candidate set for a
// subnet larger than /24 (infra + DHCP + near neighbors), never a full sweep.
func (c CandidateGenerator) boundedLargeSubnet(seedIP string, ipNet *net.IPNet, limit int) []string {
	// Derive the /24 that contains the seed and prioritize within it.
	ip := net.ParseIP(seedIP).To4()
	mask := net.CIDRMask(24, 32)
	local := &net.IPNet{IP: ip.Mask(mask), Mask: mask}
	if !ipNet.Contains(local.IP) {
		local = ipNet
	}
	if limit <= 0 {
		limit = 64
	}
	sub := NewCandidateGenerator(c.guard)
	return sub.GenerateFromSeed(seedIP, local.String(), limit)
}

func uint32ToIP(v uint32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return net.IP(b)
}

// ipLess orders two IPv4 strings numerically (falls back to string compare).
func ipLess(a, b string) bool {
	ai, bi := net.ParseIP(a).To4(), net.ParseIP(b).To4()
	if ai == nil || bi == nil {
		return a < b
	}
	return binary.BigEndian.Uint32(ai) < binary.BigEndian.Uint32(bi)
}

// SubnetOf returns the /24 CIDR string that contains ip (the default auto scope).
func SubnetOf(ip string, prefix int) string {
	parsed := net.ParseIP(ip).To4()
	if parsed == nil {
		return ""
	}
	if prefix <= 0 || prefix > 32 {
		prefix = 24
	}
	mask := net.CIDRMask(prefix, 32)
	n := &net.IPNet{IP: parsed.Mask(mask), Mask: mask}
	return n.String()
}
