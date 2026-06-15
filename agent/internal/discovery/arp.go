package discovery

import (
	"fmt"
	"regexp"
	"strings"
)

// ARPEntry is one IP→MAC association read from the OS neighbour/ARP table.
type ARPEntry struct {
	IP  string
	MAC string // normalized lower-case colon form, e.g. "00:11:22:33:44:55"
}

var (
	arpIPv4Re = regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)
	arpMACRe  = regexp.MustCompile(`\b([0-9a-fA-F]{1,2}([:-])[0-9a-fA-F]{1,2}(?:[:-][0-9a-fA-F]{1,2}){4})\b`)
)

// ParseARPTable extracts IP→MAC pairs from the textual output of the OS ARP /
// neighbour table. It is format-agnostic: it handles Windows `arp -a`
// ("192.168.1.1  00-11-22-33-44-55  dynamic"), BSD/macOS/Linux `arp -a`
// ("host (192.168.1.1) at 00:11:22:.. [ether] on eth0") and Linux `ip neigh`
// ("192.168.1.1 dev eth0 lladdr 00:11:22:.. REACHABLE"). Broadcast and
// incomplete entries are skipped. The result preserves first-seen order and is
// deduplicated by IP.
//
// Parsing OS table text is the standard, non-intrusive way to read already-known
// neighbours; it is not packet crafting, scanning, or human-terminal parsing of a
// scanner tool (which the project forbids for Nmap).
func ParseARPTable(output string) []ARPEntry {
	var out []ARPEntry
	seen := map[string]bool{}
	for _, line := range strings.Split(output, "\n") {
		ip := arpIPv4Re.FindString(line)
		if ip == "" {
			continue
		}
		macMatch := arpMACRe.FindString(line)
		if macMatch == "" {
			continue
		}
		mac, ok := normalizeMAC(macMatch)
		if !ok {
			continue
		}
		if seen[ip] {
			continue
		}
		seen[ip] = true
		out = append(out, ARPEntry{IP: ip, MAC: mac})
	}
	return out
}

// normalizeMAC converts a MAC in colon or dash form to lower-case, zero-padded
// colon form. It returns ok=false for broadcast (ff:..:ff) and all-zero
// (incomplete) addresses, which are not real neighbours.
func normalizeMAC(raw string) (string, bool) {
	sep := ":"
	if strings.Contains(raw, "-") {
		sep = "-"
	}
	parts := strings.Split(raw, sep)
	if len(parts) != 6 {
		return "", false
	}
	octets := make([]string, 6)
	allFF, allZero := true, true
	for i, p := range parts {
		if len(p) == 0 || len(p) > 2 {
			return "", false
		}
		v := 0
		if _, err := fmt.Sscanf(strings.ToLower(p), "%x", &v); err != nil {
			return "", false
		}
		if v != 0xff {
			allFF = false
		}
		if v != 0 {
			allZero = false
		}
		octets[i] = fmt.Sprintf("%02x", v)
	}
	if allFF || allZero {
		return "", false
	}
	return strings.Join(octets, ":"), true
}
