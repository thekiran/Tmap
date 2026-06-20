package ippool

import (
	"fmt"
	"net"
)

// ScopeGuard enforces the safety scope: only private/local targets, and subnet
// sizes bounded unless the user explicitly confirms a larger scope. It is pure
// and has no side effects, so it is fully unit-testable.
type ScopeGuard struct {
	AllowLinkLocal bool // 169.254/16 + fe80::/10 (only when an interface is link-local)
	MaxAutoPrefix  int  // largest auto subnet, e.g. 24
	WarnPrefix     int  // warn at this prefix, e.g. 23
	BlockPrefix    int  // block this prefix and larger without confirmation, e.g. 16
}

// NewScopeGuard builds a guard from config.
func NewScopeGuard(cfg Config) ScopeGuard {
	return ScopeGuard{
		AllowLinkLocal: false,
		MaxAutoPrefix:  cfg.MaxAutoPrefix,
		WarnPrefix:     cfg.WarnPrefix,
		BlockPrefix:    cfg.BlockPrefix,
	}
}

// AllowIP reports whether a single IP may be probed. Only private (RFC1918 /
// RFC4193) and — when enabled — link-local addresses pass. Public, loopback,
// multicast, and unspecified addresses are always rejected.
func (g ScopeGuard) AllowIP(ipStr string) (bool, string) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, "not a valid IP address"
	}
	if ip.IsLoopback() {
		return false, "loopback addresses are not probed"
	}
	if ip.IsMulticast() || ip.IsUnspecified() {
		return false, "multicast/unspecified addresses are not probed"
	}
	if ip.IsPrivate() {
		return true, ""
	}
	if g.AllowLinkLocal && ip.IsLinkLocalUnicast() {
		return true, ""
	}
	if ip.IsLinkLocalUnicast() {
		return false, "link-local target rejected (enable only when the interface is link-local)"
	}
	return false, "public/non-private targets are never scanned by default"
}

// CIDRDecision is the outcome of evaluating a subnet for automatic scanning.
type CIDRDecision struct {
	Allowed bool
	Warning string
	Reason  string
	Prefix  int
}

// AllowCIDR evaluates whether a subnet may be enumerated automatically. The
// network must be private, and its size is bounded: at or below MaxAutoPrefix is
// fine, WarnPrefix emits a warning, and BlockPrefix (or larger) requires
// `confirmed` to proceed.
func (g ScopeGuard) AllowCIDR(cidr string, confirmed bool) CIDRDecision {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return CIDRDecision{Reason: fmt.Sprintf("invalid CIDR %q: %v", cidr, err)}
	}
	ones, bits := ipNet.Mask.Size()
	if bits != 32 {
		// Keep IPv6 enumeration out of automatic mode entirely.
		return CIDRDecision{Prefix: ones, Reason: "automatic enumeration is limited to IPv4 subnets"}
	}
	if ok, reason := g.AllowIP(ipNet.IP.String()); !ok {
		return CIDRDecision{Prefix: ones, Reason: reason}
	}

	d := CIDRDecision{Prefix: ones}
	switch {
	case ones >= g.MaxAutoPrefix:
		// /24 or smaller: always fine.
		d.Allowed = true
	case ones <= g.BlockPrefix:
		// /16 or larger: requires explicit confirmation.
		if confirmed {
			d.Allowed = true
			d.Warning = fmt.Sprintf("scanning /%d is a large scope (%d hosts) — proceeding because the user confirmed", ones, hostCount(ones))
		} else {
			d.Reason = fmt.Sprintf("/%d (%d hosts) exceeds the automatic limit; explicit user confirmation required", ones, hostCount(ones))
		}
	default:
		// Between block and max (e.g. /23..../17): allowed but warned.
		d.Allowed = true
		d.Warning = fmt.Sprintf("/%d (%d hosts) is larger than the default /%d — scanning will be slow and rate-limited", ones, hostCount(ones), g.MaxAutoPrefix)
	}
	return d
}

func hostCount(prefix int) int {
	if prefix >= 31 {
		return 1 << (32 - prefix)
	}
	return (1 << (32 - prefix)) - 2 // minus network + broadcast
}
