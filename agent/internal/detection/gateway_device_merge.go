package detection

import (
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

var gatewayRolePriority = map[string]int{
	"possible_cpe":              5,
	"possible_modem":            5,
	"upstream_private_gateway":  4,
	"default_gateway":           3,
	"local_router":              2,
	"unknown":                   1,
	"":                          0,
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
			byIP[d.IP] = d
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
	if gatewayRolePriority[b.Role] > gatewayRolePriority[a.Role] {
		a.Role = b.Role
	}
	a.Reachable = a.Reachable || b.Reachable
	a.UPnPFound = a.UPnPFound || b.UPnPFound
	a.UPnPIGDFound = a.UPnPIGDFound || b.UPnPIGDFound
	a.WANCommonInterfaceFound = a.WANCommonInterfaceFound || b.WANCommonInterfaceFound
	a.TR064Found = a.TR064Found || b.TR064Found
	a.TR064AuthRequired = a.TR064AuthRequired || b.TR064AuthRequired

	a.HTTPTitle = firstNonEmpty(a.HTTPTitle, b.HTTPTitle)
	a.ServerHeader = firstNonEmpty(a.ServerHeader, b.ServerHeader)
	a.WWWAuthenticate = firstNonEmpty(a.WWWAuthenticate, b.WWWAuthenticate)
	a.WWWAuthRealm = firstNonEmpty(a.WWWAuthRealm, b.WWWAuthRealm)
	a.FaviconHash = firstNonEmpty(a.FaviconHash, b.FaviconHash)
	a.TLSCertCN = firstNonEmpty(a.TLSCertCN, b.TLSCertCN)
	a.TLSServerName = firstNonEmpty(a.TLSServerName, b.TLSServerName)
	a.RedirectPath = firstNonEmpty(a.RedirectPath, b.RedirectPath)
	a.RedirectLocation = firstNonEmpty(a.RedirectLocation, b.RedirectLocation)
	a.HTMLMetaGenerator = firstNonEmpty(a.HTMLMetaGenerator, b.HTMLMetaGenerator)
	a.WANAccessType = firstNonEmpty(a.WANAccessType, b.WANAccessType)
	a.PhysicalLinkStatus = firstNonEmpty(a.PhysicalLinkStatus, b.PhysicalLinkStatus)
	a.MACVendor = firstNonEmpty(a.MACVendor, b.MACVendor)
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
	a.AccessHints = appendUniqueStrings(a.AccessHints, b.AccessHints...)
	a.PhysicalHints = appendUniqueStrings(a.PhysicalHints, b.PhysicalHints...)
	a.Notes = appendUniqueStrings(a.Notes, b.Notes...)
	a.TR064Services = appendUniqueStrings(a.TR064Services, b.TR064Services...)

	a.DeviceConfidence = maxFloat(a.DeviceConfidence, b.DeviceConfidence)
	a.AccessConfidence = maxFloat(a.AccessConfidence, b.AccessConfidence)
	a.Confidence = maxFloat(a.Confidence, b.Confidence)
	return a
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
