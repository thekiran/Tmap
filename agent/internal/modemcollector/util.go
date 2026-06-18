package modemcollector

import (
	"net"
	"strconv"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

func isRFC1918(ip string) bool {
	parsed := net.ParseIP(stripCIDR(ip))
	if parsed == nil {
		return false
	}
	v4 := parsed.To4()
	if v4 == nil {
		return false
	}
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

func isExcludedIP(ip, agentIP string) bool {
	parsed := net.ParseIP(stripCIDR(ip))
	if parsed == nil {
		return true
	}
	if agentIP != "" && stripCIDR(ip) == stripCIDR(agentIP) {
		return true
	}
	return parsed.IsLoopback() || parsed.IsMulticast()
}

func stripCIDR(s string) string {
	if i := strings.IndexByte(s, '/'); i >= 0 {
		return s[:i]
	}
	return s
}

func triFromBool(ok bool) models.TriState {
	if ok {
		return models.TriTrue
	}
	return models.TriUnknown
}

func triFromString(s string) models.TriState {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "found", "reachable":
		return models.TriTrue
	case "false", "not_found", "absent":
		return models.TriFalse
	default:
		return models.TriUnknown
	}
}

func ptrString(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func ptrInt64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func appendUniqueStrings(s []string, values ...string) []string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		found := false
		for _, e := range s {
			if e == v {
				found = true
				break
			}
		}
		if !found {
			s = append(s, v)
		}
	}
	return s
}

func appendUniqueInts(s []int, values ...int) []int {
	for _, v := range values {
		if v == 0 {
			continue
		}
		found := false
		for _, e := range s {
			if e == v {
				found = true
				break
			}
		}
		if !found {
			s = append(s, v)
		}
	}
	return s
}

func appendUniqueHTTP(s []models.HTTPObservation, values ...models.HTTPObservation) []models.HTTPObservation {
	for _, v := range values {
		key := strings.Join([]string{v.Source, v.Method, v.URL, v.EvidenceID}, "\x00")
		found := false
		for _, e := range s {
			if strings.Join([]string{e.Source, e.Method, e.URL, e.EvidenceID}, "\x00") == key {
				found = true
				break
			}
		}
		if !found {
			s = append(s, v)
		}
	}
	return s
}

func appendUniqueTLS(s []models.TLSObservation, values ...models.TLSObservation) []models.TLSObservation {
	for _, v := range values {
		key := strings.Join([]string{v.Source, v.IP, strconv.Itoa(v.Port), v.CN, v.Issuer, v.EvidenceID}, "\x00")
		found := false
		for _, e := range s {
			if strings.Join([]string{e.Source, e.IP, strconv.Itoa(e.Port), e.CN, e.Issuer, e.EvidenceID}, "\x00") == key {
				found = true
				break
			}
		}
		if !found {
			s = append(s, v)
		}
	}
	return s
}

func appendUniqueAttempts(s []models.ProbeAttempt, values ...models.ProbeAttempt) []models.ProbeAttempt {
	for _, v := range values {
		key := strings.Join([]string{v.Source, v.Target, v.Protocol, strconv.Itoa(v.Port), v.URL, v.Method, v.Error, v.EvidenceID}, "\x00")
		found := false
		for _, e := range s {
			if strings.Join([]string{e.Source, e.Target, e.Protocol, strconv.Itoa(e.Port), e.URL, e.Method, e.Error, e.EvidenceID}, "\x00") == key {
				found = true
				break
			}
		}
		if !found {
			s = append(s, v)
		}
	}
	return s
}

func appendUniqueAccess(s []models.GatewayAccessEvidence, values ...models.GatewayAccessEvidence) []models.GatewayAccessEvidence {
	for _, v := range values {
		key := strings.Join([]string{v.Source, v.Type, v.Value, v.EvidenceID}, "\x00")
		found := false
		for _, e := range s {
			if strings.Join([]string{e.Source, e.Type, e.Value, e.EvidenceID}, "\x00") == key {
				found = true
				break
			}
		}
		if !found {
			s = append(s, v)
		}
	}
	return s
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if b > a {
		return b
	}
	return a
}

func rolePriority(role string) int {
	switch role {
	case "possible_cpe", "possible_modem", "possible_modem_or_ont":
		return 5
	case "upstream_private_gateway":
		return 4
	case "internet_gateway_device", "cpe_management_endpoint":
		return 4
	case "default_gateway":
		return 3
	default:
		return 1
	}
}

func sourceList(values ...string) []string {
	var out []string
	for _, v := range values {
		out = appendUniqueStrings(out, v)
	}
	return out
}
