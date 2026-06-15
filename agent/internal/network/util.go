package network

import "regexp"

var ipv4Re = regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\b`)

// extractIPv4 returns the unique dotted-quad addresses found in s, preserving
// first-seen order. Used to parse the loosely-structured output of OS tools.
func extractIPv4(s string) []string {
	matches := ipv4Re.FindAllString(s, -1)
	seen := make(map[string]bool, len(matches))
	var out []string
	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	return out
}
