// Package discovery performs authorized local network discovery: it detects the
// agent's interfaces, validates a scan scope, finds live hosts inside that scope,
// normalizes them into devices, and attaches evidence to every fact. It never
// scans outside the validated scope and refuses public address space unless the
// caller explicitly opts in to the dangerous mode.
package discovery

import (
	"fmt"
	"net"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

// Scope limits to keep scans bounded and safe. A normal LAN is a /24 (254 hosts);
// anything past maxScanHosts is refused outright, and past softWarnHosts warns.
const (
	maxScanHosts  = 65536 // ~/16; larger is almost certainly a mistake or abuse
	softWarnHosts = 1024  // larger than ~/22 is worth a heads-up
)

// Warning codes.
const (
	warnPublicScope   = "public_scope_enabled"
	warnLargeScope    = "large_scope"
	warnIPv6Scope     = "ipv6_scope_unsupported"
	warnInferredOnly  = "inferred_topology_only"
	warnNoGateway     = "no_default_gateway"
)

// ScopeError is returned when a requested scope is refused.
type ScopeError struct{ Reason string }

func (e *ScopeError) Error() string { return e.Reason }

// ParseScope resolves and validates the scan scope. requested is "auto" (use the
// selected interface's network) or an explicit CIDR. allowPublic must be true to
// permit anything outside private address space — this is the dangerous flag and
// always produces a danger-level warning. profile is recorded for the report.
func ParseScope(requested, profile string, allowPublic bool, primary models.InterfaceInfo) (models.ScanScope, []models.Warning, error) {
	scope := models.ScanScope{
		Requested:     strings.TrimSpace(requested),
		PublicAllowed: allowPublic,
		Profile:       profile,
		Interface:     primary.Name,
	}
	var warnings []models.Warning

	cidr := scope.Requested
	if cidr == "" || strings.EqualFold(cidr, "auto") {
		if primary.CIDR == "" {
			return scope, warnings, &ScopeError{Reason: "could not auto-detect a local IPv4 network on the selected interface; pass --cidr <network/prefix>"}
		}
		cidr = primary.CIDR
		scope.Requested = "auto"
	}

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return scope, warnings, &ScopeError{Reason: fmt.Sprintf("invalid CIDR %q: %v", cidr, err)}
	}
	scope.CIDR = ipnet.String()

	// IPv6 networks are accepted for reporting but a full sweep is infeasible.
	if ip.To4() == nil {
		scope.HostCount = 0
		scope.Private = isPrivateIP(ip)
		warnings = append(warnings, models.Warning{
			Code: warnIPv6Scope, Severity: models.SeverityWarning,
			Message: "IPv6 scopes are too large to sweep; only interface/gateway facts are reported for IPv6.",
		})
		if !scope.Private && !allowPublic {
			return scope, warnings, &ScopeError{Reason: "refusing a non-private (public) IPv6 scope; re-run with the explicit public-scope flag if you are authorized."}
		}
		if allowPublic {
			warnings = append(warnings, dangerPublicWarning())
		}
		return scope, warnings, nil
	}

	ones, bits := ipnet.Mask.Size()
	scope.HostCount = hostCount(ones, bits)
	scope.Private = isPrivateNetwork(ipnet)

	if !scope.Private && !allowPublic {
		return scope, warnings, &ScopeError{Reason: fmt.Sprintf(
			"refusing to scan non-private scope %s. Only RFC1918/link-local networks are scanned by default. Re-run with the explicit public-scope flag ONLY for networks you are authorized to scan.", scope.CIDR)}
	}
	if scope.HostCount > maxScanHosts {
		return scope, warnings, &ScopeError{Reason: fmt.Sprintf(
			"scope %s has %d addresses, above the %d limit; choose a smaller subnet.", scope.CIDR, scope.HostCount, maxScanHosts)}
	}
	if allowPublic && !scope.Private {
		warnings = append(warnings, dangerPublicWarning())
	}
	if scope.HostCount > softWarnHosts {
		warnings = append(warnings, models.Warning{
			Code: warnLargeScope, Severity: models.SeverityWarning,
			Message: fmt.Sprintf("scope %s has %d addresses; the scan may take a while.", scope.CIDR, scope.HostCount),
		})
	}
	return scope, warnings, nil
}

func dangerPublicWarning() models.Warning {
	return models.Warning{
		Code: warnPublicScope, Severity: models.SeverityDanger,
		Message: "Public/non-private scope scanning is ENABLED. Only do this on networks you own or are explicitly authorized to scan. This is your responsibility.",
	}
}

// hostCount returns the number of addressable hosts in an IPv4 network. For
// prefixes /0../30 it excludes the network and broadcast address; /31 and /32 are
// returned as-is (point-to-point / single host).
func hostCount(ones, bits int) int {
	hostBits := bits - ones
	if hostBits <= 1 {
		// /31 (2) and /32 (1): every address is usable.
		return 1 << hostBits
	}
	if hostBits >= 31 { // guard against overflow for absurd masks
		return maxScanHosts + 1
	}
	return (1 << hostBits) - 2
}

// HostsInScope enumerates the addressable IPv4 host addresses in the scope, in
// order, excluding network and broadcast (for /24-style nets). It returns nil for
// IPv6 or empty scopes. The result is bounded by maxScanHosts.
func HostsInScope(scope models.ScanScope) []string {
	if scope.CIDR == "" {
		return nil
	}
	_, ipnet, err := net.ParseCIDR(scope.CIDR)
	if err != nil || ipnet.IP.To4() == nil {
		return nil
	}
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones
	base := ipnet.IP.Mask(ipnet.Mask).To4()
	if base == nil {
		return nil
	}
	total := 1 << hostBits
	if total > maxScanHosts {
		return nil
	}
	out := make([]string, 0, total)
	start, end := 0, total
	if hostBits >= 2 { // skip network (0) and broadcast (total-1)
		start, end = 1, total-1
	}
	baseInt := uint32(base[0])<<24 | uint32(base[1])<<16 | uint32(base[2])<<8 | uint32(base[3])
	for i := start; i < end; i++ {
		v := baseInt + uint32(i)
		out = append(out, fmt.Sprintf("%d.%d.%d.%d", byte(v>>24), byte(v>>16), byte(v>>8), byte(v)))
	}
	return out
}

// InScope reports whether an IP string lies inside the validated scope.
func InScope(scope models.ScanScope, ipStr string) bool {
	if scope.CIDR == "" {
		return false
	}
	_, ipnet, err := net.ParseCIDR(scope.CIDR)
	if err != nil {
		return false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return ipnet.Contains(ip)
}

// isPrivateNetwork reports whether a network's address is private/link-local.
func isPrivateNetwork(ipnet *net.IPNet) bool { return isPrivateIP(ipnet.IP) }

// isPrivateIP reports whether an IP is in private or link-local space. Carrier
// CGNAT space (100.64/10) is deliberately NOT treated as private: it is the
// carrier's, not the user's LAN, so it is refused by default like public space.
func isPrivateIP(ip net.IP) bool {
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
	// IPv6 ULA fc00::/7.
	return len(ip) == net.IPv6len && (ip[0]&0xfe) == 0xfc
}
