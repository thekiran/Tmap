package probes

import "strings"

// appendUnique appends v to s only if it is not already present. Hints are small
// slices, so the linear scan is cheap and keeps the result tidy.
func appendUnique(s []string, v string) []string {
	for _, e := range s {
		if e == v {
			return s
		}
	}
	return append(s, v)
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
