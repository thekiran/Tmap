package deviceintel

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/thekiran/iad/internal/safety"
	"github.com/thekiran/iad/pkg/models"
)

var safeTCPPorts = []int{
	22, 23, 53, 80, 139, 443, 445, 515, 631, 1900, 1883, 5000, 5001,
	5353, 5357, 7547, 8008, 8009, 8080, 8123, 8443, 9100, 32400, 554,
}

var serviceNames = map[int]string{
	22:    "ssh",
	23:    "telnet",
	53:    "domain",
	80:    "http",
	139:   "netbios-ssn",
	443:   "https",
	445:   "microsoft-ds",
	515:   "printer-lpd",
	631:   "ipp",
	554:   "rtsp",
	1883:  "mqtt",
	1900:  "ssdp",
	5000:  "nas-or-upnp",
	5001:  "nas-https",
	5353:  "mdns",
	5357:  "wsdapi",
	7547:  "cwmp",
	8008:  "chromecast",
	8009:  "chromecast",
	8080:  "http-alt",
	8123:  "home-assistant",
	8443:  "https-alt",
	9100:  "jetdirect",
	32400: "plex",
}

func SafeTCPPorts() []int {
	out := append([]int(nil), safeTCPPorts...)
	sort.Ints(out)
	return out
}

func ShouldProbeTarget(scope ScanScope, ip string) bool {
	if strings.TrimSpace(ip) == "" {
		return false
	}
	if safety.IsPrivateIPString(ip) {
		return true
	}
	return scope.PublicAllowed && !scope.PrivateOnly
}

func SNMPAllowed(scope ScanScope) bool {
	return scope.SNMPCredentials && containsString(scope.OptInProbes, "snmp")
}

func SSHAllowed(scope ScanScope) bool {
	return scope.SSHCredentials && containsString(scope.OptInProbes, "ssh")
}

func RouterAPIAllowed(scope ScanScope) bool {
	return scope.RouterAPICredentials && containsString(scope.OptInProbes, "router_api")
}

func TR064Allowed(scope ScanScope) bool {
	return scope.TR064Credentials && containsString(scope.OptInProbes, "tr064")
}

func TR181Allowed(scope ScanScope) bool {
	return scope.TR181Credentials && containsString(scope.OptInProbes, "tr181")
}

func deviceID(ip string) string {
	return "dev-" + ip
}

func firstIP(d models.DeviceIntelDevice) string {
	if len(d.IPAddresses) == 0 {
		return ""
	}
	return d.IPAddresses[0]
}

func serviceLabel(s models.DeviceIntelService) string {
	name := s.Name
	if name == "" {
		name = serviceName(s.Port)
	}
	if name == "" {
		name = s.Protocol
	}
	return fmt.Sprintf("%s/%d", name, s.Port)
}

func serviceName(port int) string {
	if name := serviceNames[port]; name != "" {
		return name
	}
	return ""
}

func serviceKey(protocol string, port int) string {
	if protocol == "" {
		protocol = "tcp"
	}
	return strings.ToLower(protocol) + ":" + strconv.Itoa(port)
}

func hasOpenPort(d *models.DeviceIntelDevice, port int) bool {
	for _, s := range d.Services {
		if s.Port == port && strings.EqualFold(s.State, "open") {
			return true
		}
	}
	return false
}

func hasAnyPort(d *models.DeviceIntelDevice, ports ...int) bool {
	for _, p := range ports {
		if hasOpenPort(d, p) {
			return true
		}
	}
	return false
}

func evidenceIDsForPorts(d *models.DeviceIntelDevice, ports ...int) []string {
	var ids []string
	for _, s := range d.Services {
		for _, p := range ports {
			if s.Port == p {
				ids = append(ids, s.EvidenceIDs...)
				break
			}
		}
	}
	return sortedUnique(ids)
}

func containsString(values []string, needle string) bool {
	needle = strings.ToLower(strings.TrimSpace(needle))
	for _, v := range values {
		if strings.ToLower(strings.TrimSpace(v)) == needle {
			return true
		}
	}
	return false
}

func appendUnique(values []string, next ...string) []string {
	seen := make(map[string]bool, len(values)+len(next))
	out := make([]string, 0, len(values)+len(next))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	for _, v := range next {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func sortedUnique(values []string) []string {
	out := appendUnique(nil, values...)
	sort.Strings(out)
	return out
}

func sortedUniqueInts(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	seen := map[int]bool{}
	out := make([]int, 0, len(values))
	for _, v := range values {
		if seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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

func confidenceLabel(v float64) string {
	switch {
	case v >= 0.90:
		return models.ConfVeryHigh
	case v >= 0.75:
		return models.ConfHigh
	case v >= 0.50:
		return models.ConfMedium
	case v >= 0.25:
		return models.ConfLow
	default:
		return models.ConfVeryLow
	}
}

func ipLess(a, b string) bool {
	ai, aok := ipv4ToUint(a)
	bi, bok := ipv4ToUint(b)
	if aok && bok {
		return ai < bi
	}
	return a < b
}

func ipv4ToUint(s string) (uint32, bool) {
	ip := net.ParseIP(s)
	if ip == nil {
		return 0, false
	}
	v4 := ip.To4()
	if v4 == nil {
		return 0, false
	}
	return uint32(v4[0])<<24 | uint32(v4[1])<<16 | uint32(v4[2])<<8 | uint32(v4[3]), true
}
