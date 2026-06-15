// Package detection is the engine that turns probe evidence into a ranked
// verdict: normalize → fingerprint match → rule scoring → classify → confidence
// → explanation (doc §5–6).
package detection

import "strings"

// NormalizeModel tidies a raw router/model string for display and matching:
// trims, turns underscores into spaces and collapses runs of whitespace. It
// deliberately does not try to canonicalize vendor names — the fingerprint
// matcher uses substring matching with per-device aliases instead, which is more
// robust than maintaining a rename table.
func NormalizeModel(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	fields := strings.Fields(s) // splits on any whitespace and drops empties
	return strings.Join(fields, " ")
}
