package detection

import (
	"strconv"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

var gatewayRolePriority = map[string]int{
	"possible_cpe":             5,
	"possible_modem":           5,
	"upstream_private_gateway": 4,
	"default_gateway":          3,
	"local_router":             2,
	"unknown":                  1,
	"":                         0,
}

// MergeGatewayDevices deduplicates gateway devices by IP while preserving the
// strongest role and every positive piece of evidence. It never overwrites a
// populated model/identity/access signal with an empty value.
func MergeGatewayDevices(devices []models.GatewayDevice) []models.GatewayDevice {
	byIP := map[string]models.GatewayDevice{}
	var order []string
	var ipless []models.GatewayDevice
	for _, d := range devices {
		if d.IP == "" {
			ipless = append(ipless, d)
			continue
		}
		existing, ok := byIP[d.IP]
		if !ok {
			byIP[d.IP] = normalizeGatewayDevice(d)
			order = append(order, d.IP)
			continue
		}
		byIP[d.IP] = mergeGatewayDevice(existing, d)
	}
	out := make([]models.GatewayDevice, 0, len(order))
	for _, ip := range order {
		out = append(out, byIP[ip])
	}
	out = append(out, ipless...)
	return out
}

func mergeGatewayDevice(a, b models.GatewayDevice) models.GatewayDevice {
	a = normalizeGatewayDevice(a)
	b = normalizeGatewayDevice(b)
	if gatewayRolePriority[b.Role] > gatewayRolePriority[a.Role] {
		a.Role = b.Role
	}
	a.Reachable = a.Reachable || b.Reachable
	a.ReachableState = mergeReachableState(a.ReachableState, b.ReachableState)
	a.UPnPFound = a.UPnPFound || b.UPnPFound
	a.UPnPIGDFound = a.UPnPIGDFound || b.UPnPIGDFound
	a.WANCommonInterfaceFound = a.WANCommonInterfaceFound || b.WANCommonInterfaceFound
	a.TR064Found = a.TR064Found || b.TR064Found
	a.TR064AuthRequired = a.TR064AuthRequired || b.TR064AuthRequired

	a.OpenPorts = appendUniqueInts(a.OpenPorts, b.OpenPorts...)
	a.HTTPObservations = appendUniqueHTTPObservations(a.HTTPObservations, b.HTTPObservations...)
	a.HTTPTitle = firstNonEmpty(a.HTTPTitle, b.HTTPTitle)
	a.ServerHeader = firstNonEmpty(a.ServerHeader, b.ServerHeader)
	a.WWWAuthenticate = firstNonEmpty(a.WWWAuthenticate, b.WWWAuthenticate)
	a.WWWAuthRealm = firstNonEmpty(a.WWWAuthRealm, b.WWWAuthRealm)
	a.FaviconHash = firstNonEmpty(a.FaviconHash, b.FaviconHash)
	a.TLSObservations = appendUniqueTLSObservations(a.TLSObservations, b.TLSObservations...)
	a.TLSCertCN = firstNonEmpty(a.TLSCertCN, b.TLSCertCN)
	a.TLSCertIssuer = firstNonEmpty(a.TLSCertIssuer, b.TLSCertIssuer)
	a.TLSServerName = firstNonEmpty(a.TLSServerName, b.TLSServerName)
	a.RedirectPath = firstNonEmpty(a.RedirectPath, b.RedirectPath)
	a.RedirectLocation = firstNonEmpty(a.RedirectLocation, b.RedirectLocation)
	a.HTMLMetaGenerator = firstNonEmpty(a.HTMLMetaGenerator, b.HTMLMetaGenerator)
	a.UPnPState = mergeState(a.UPnPState, b.UPnPState, a.UPnPFound)
	a.WANAccessType = firstNonEmpty(a.WANAccessType, b.WANAccessType)
	a.PhysicalLinkStatus = firstNonEmpty(a.PhysicalLinkStatus, b.PhysicalLinkStatus)
	a.TR064State = mergeState(a.TR064State, b.TR064State, a.TR064Found)
	a.SNMPState = mergeState(a.SNMPState, b.SNMPState, false)
	a.MACVendor = firstNonEmpty(a.MACVendor, b.MACVendor)
	a.CPEModelGuess = firstNonEmpty(a.CPEModelGuess, b.CPEModelGuess)
	a.Model = firstNonEmpty(a.Model, b.Model)
	a.Manufacturer = firstNonEmpty(a.Manufacturer, b.Manufacturer)
	a.FingerprintID = firstNonEmpty(a.FingerprintID, b.FingerprintID)

	if b.Layer1UpstreamMaxBitRate > a.Layer1UpstreamMaxBitRate {
		a.Layer1UpstreamMaxBitRate = b.Layer1UpstreamMaxBitRate
	}
	if b.Layer1DownstreamMaxBitRate > a.Layer1DownstreamMaxBitRate {
		a.Layer1DownstreamMaxBitRate = b.Layer1DownstreamMaxBitRate
	}
	a.TLSCertSANs = appendUniqueStrings(a.TLSCertSANs, b.TLSCertSANs...)
	a.LoginLabels = appendUniqueStrings(a.LoginLabels, b.LoginLabels...)
	a.AccessEvidence = appendUniqueGatewayAccessEvidence(a.AccessEvidence, b.AccessEvidence...)
	a.AccessHints = appendUniqueStrings(a.AccessHints, b.AccessHints...)
	a.PhysicalHints = appendUniqueStrings(a.PhysicalHints, b.PhysicalHints...)
	a.Notes = appendUniqueStrings(a.Notes, b.Notes...)
	a.TR064Services = appendUniqueStrings(a.TR064Services, b.TR064Services...)
	a.FailedAttempts = appendUniqueProbeAttempts(a.FailedAttempts, b.FailedAttempts...)
	a.EvidenceIDs = appendUniqueStrings(a.EvidenceIDs, b.EvidenceIDs...)

	a.DeviceConfidence = maxFloat(a.DeviceConfidence, b.DeviceConfidence)
	a.AccessConfidence = maxFloat(a.AccessConfidence, b.AccessConfidence)
	a.Confidence = maxFloat(a.Confidence, b.Confidence)
	a = normalizeGatewayDevice(a)
	return a
}

func normalizeGatewayDevice(d models.GatewayDevice) models.GatewayDevice {
	if d.Reachable {
		d.ReachableState = models.ReachableTrue
	} else if d.ReachableState == "" {
		if len(d.FailedAttempts) > 0 {
			d.ReachableState = models.ReachableUnknown
		} else {
			d.ReachableState = models.ReachableUnknown
		}
	}
	if d.CPEModelGuess == "" {
		d.CPEModelGuess = strings.TrimSpace(strings.Join([]string{d.Manufacturer, d.Model}, " "))
	}
	if d.UPnPState == "" {
		d.UPnPState = boolState(d.UPnPFound)
	}
	if d.TR064State == "" {
		d.TR064State = boolState(d.TR064Found)
		if d.TR064AuthRequired && !d.TR064Found {
			d.TR064State = "auth_required"
		}
	}
	if d.SNMPState == "" {
		d.SNMPState = "not_probed"
	}
	return d
}

func boolState(ok bool) string {
	if ok {
		return "found"
	}
	return "not_found"
}

func mergeReachableState(a, b string) string {
	if a == models.ReachableTrue || b == models.ReachableTrue {
		return models.ReachableTrue
	}
	if a == models.ReachableFalse || b == models.ReachableFalse {
		return models.ReachableFalse
	}
	return models.ReachableUnknown
}

func mergeState(a, b string, found bool) string {
	if found {
		return "found"
	}
	if a == "" {
		return b
	}
	if b == "" || b == "not_found" || b == "not_probed" {
		return a
	}
	return b
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

func appendUniqueHTTPObservations(s []models.HTTPObservation, values ...models.HTTPObservation) []models.HTTPObservation {
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

func appendUniqueTLSObservations(s []models.TLSObservation, values ...models.TLSObservation) []models.TLSObservation {
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

func appendUniqueGatewayAccessEvidence(s []models.GatewayAccessEvidence, values ...models.GatewayAccessEvidence) []models.GatewayAccessEvidence {
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

func appendUniqueProbeAttempts(s []models.ProbeAttempt, values ...models.ProbeAttempt) []models.ProbeAttempt {
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

func appendUniqueStrings(s []string, values ...string) []string {
	for _, v := range values {
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

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
