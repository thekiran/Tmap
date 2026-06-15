package system

import (
	"regexp"
	"strings"
)

// ipv4Re matches a dotted-quad. RTT values like "12 ms" never contain dots, so
// they are naturally excluded.
var ipv4Re = regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)

// parseHops extracts the ordered hop IPs from traceroute/tracert output. The
// approach works for both layouts because each hop line carries exactly one IP
// (the responding router) and we take the last dotted-quad on the line:
//
//	Windows: "  3    18 ms    17 ms    18 ms  81.x.x.x"
//	Unix:    " 3  81.x.x.x  18.0 ms  17.1 ms  18.2 ms"
//
// Header lines ("Tracing route to ...", "traceroute to ...") are skipped
// because they do not begin with a hop number. Timed-out hops become "*".
func parseHops(out string) []string {
	var hops []string
	for _, line := range strings.Split(out, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || t[0] < '0' || t[0] > '9' {
			continue
		}
		ips := ipv4Re.FindAllString(t, -1)
		if len(ips) == 0 {
			hops = append(hops, "*")
			continue
		}
		hops = append(hops, ips[len(ips)-1])
	}
	return hops
}
