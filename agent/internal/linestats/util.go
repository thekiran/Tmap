package linestats

import (
	"fmt"
	"strconv"
	"strings"
)

// normalizeKV lowercases keys and strips every non-alphanumeric character, so
// "NewDownstreamNoiseMargin", "downstream_noise_margin" and "Downstream Noise
// Margin" all collapse to the same lookup key.
func normalizeKV(kv map[string]string) map[string]string {
	if len(kv) == 0 {
		return nil
	}
	out := make(map[string]string, len(kv))
	for k, v := range kv {
		nk := normalizeKey(k)
		if nk == "" {
			continue
		}
		// First non-empty value wins; do not clobber a real value with "".
		if _, ok := out[nk]; ok && strings.TrimSpace(v) == "" {
			continue
		}
		out[nk] = v
	}
	return out
}

func normalizeKey(k string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(k) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// kvText joins all key/value pairs into one searchable string.
func kvText(kv map[string]string) string {
	if len(kv) == 0 {
		return ""
	}
	parts := make([]string, 0, len(kv))
	for k, v := range kv {
		parts = append(parts, k+" "+v)
	}
	return strings.Join(parts, " ")
}

func containsAny(hay string, subs ...string) bool {
	for _, s := range subs {
		if s != "" && strings.Contains(hay, s) {
			return true
		}
	}
	return false
}

func hasAnyKey(norm map[string]string, keys ...string) bool {
	for _, k := range keys {
		if _, ok := norm[k]; ok {
			return true
		}
	}
	return false
}

// lookupKV returns the first non-empty value among the given (already-normalized)
// alias keys.
func lookupKV(norm map[string]string, aliases ...string) string {
	for _, a := range aliases {
		if v, ok := norm[a]; ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// lookupNumber parses the first numeric value found among the aliases as a float
// (decimals kept as-is). Used for dBmV/dBm/MER and channel counts.
func lookupNumber(norm map[string]string, aliases ...string) float64 {
	for _, a := range aliases {
		if v, ok := norm[a]; ok {
			if f, ok := parseFloatLoose(v); ok {
				return f
			}
		}
	}
	return 0
}

// lookupTenthsDB parses a dB value that CPEs (TR-064, ADSL-LINE-MIB) encode in
// units of 0.1 dB when given as a bare integer. A value already containing a
// decimal point (web-UI style "6.1 dB") is taken literally.
func lookupTenthsDB(norm map[string]string, aliases ...string) float64 {
	for _, a := range aliases {
		v, ok := norm[a]
		if !ok {
			continue
		}
		v = strings.TrimSpace(strings.ReplaceAll(v, "−", "-")) // unicode minus
		if v == "" {
			continue
		}
		num := firstNumberToken(v)
		if num == "" {
			continue
		}
		if isIntToken(num) && !strings.Contains(strings.ToLower(v), ".") {
			if n, err := strconv.ParseInt(num, 10, 64); err == nil {
				return float64(n) / 10.0
			}
		}
		if f, err := strconv.ParseFloat(num, 64); err == nil {
			return f
		}
	}
	return 0
}

// lookupKbps parses a kbit/s rate as an integer.
func lookupKbps(norm map[string]string, aliases ...string) int64 {
	for _, a := range aliases {
		if v, ok := norm[a]; ok {
			if f, ok := parseFloatLoose(v); ok {
				return int64(f)
			}
		}
	}
	return 0
}

func parseFloatLoose(s string) (float64, bool) {
	s = strings.ReplaceAll(s, "−", "-")
	num := firstNumberToken(s)
	if num == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func firstNumberToken(s string) string {
	return reFloat.FindString(s)
}

func isIntToken(s string) bool {
	return reIntOnly.MatchString(s)
}

func mbps(kbps int64) string {
	if kbps <= 0 {
		return "?"
	}
	return fmt.Sprintf("%.1f", float64(kbps)/1000.0)
}

func oneDP(v float64) string { return fmt.Sprintf("%.1f", v) }

func itoa(n int) string { return strconv.Itoa(n) }
