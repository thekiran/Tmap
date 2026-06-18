package modemcollector

import (
	"sort"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

type EvidenceStore struct {
	devices      map[string]models.GatewayDevice
	order        []string
	observations []Observation
	conflicts    []models.DataConflict
}

func NewEvidenceStore() *EvidenceStore {
	return &EvidenceStore{devices: map[string]models.GatewayDevice{}}
}

func (s *EvidenceStore) AddObservation(obs Observation) {
	if strings.TrimSpace(obs.ID) == "" {
		obs.ID = obs.Source + ":" + obs.TargetIP + ":" + obs.Kind
	}
	s.observations = append(s.observations, obs)
}

func (s *EvidenceStore) MergeDevice(d models.GatewayDevice) {
	if d.IP == "" {
		return
	}
	d = normalizeDevice(d)
	existing, ok := s.devices[d.IP]
	if !ok {
		s.devices[d.IP] = d
		s.order = append(s.order, d.IP)
		return
	}
	s.devices[d.IP] = mergeDevice(existing, d)
}

func (s *EvidenceStore) AddConflict(c models.DataConflict) {
	s.conflicts = append(s.conflicts, c)
}

func (s *EvidenceStore) Devices() []models.GatewayDevice {
	out := make([]models.GatewayDevice, 0, len(s.order))
	for _, ip := range s.order {
		out = append(out, s.devices[ip])
	}
	sort.SliceStable(out, func(i, j int) bool {
		return rolePriority(out[i].Role) > rolePriority(out[j].Role)
	})
	return out
}

func (s *EvidenceStore) Conflicts() []models.DataConflict {
	return append([]models.DataConflict{}, s.conflicts...)
}

func (s *EvidenceStore) Observations() []Observation {
	return append([]Observation{}, s.observations...)
}

func mergeDevice(a, b models.GatewayDevice) models.GatewayDevice {
	a = normalizeDevice(a)
	b = normalizeDevice(b)
	if rolePriority(b.Role) > rolePriority(a.Role) {
		a.Role = b.Role
	}
	a.Reachable = a.Reachable || b.Reachable
	a.ReachableState = mergeTriString(a.ReachableState, b.ReachableState)
	a.OpenPorts = appendUniqueInts(a.OpenPorts, b.OpenPorts...)
	a.HTTPObservations = appendUniqueHTTP(a.HTTPObservations, b.HTTPObservations...)
	a.TLSObservations = appendUniqueTLS(a.TLSObservations, b.TLSObservations...)
	a.FailedAttempts = appendUniqueAttempts(a.FailedAttempts, b.FailedAttempts...)
	a.EvidenceIDs = appendUniqueStrings(a.EvidenceIDs, b.EvidenceIDs...)
	a.AccessEvidence = appendUniqueAccess(a.AccessEvidence, b.AccessEvidence...)
	a.AccessHints = appendUniqueStrings(a.AccessHints, b.AccessHints...)
	a.PhysicalHints = appendUniqueStrings(a.PhysicalHints, b.PhysicalHints...)
	a.TR064Services = appendUniqueStrings(a.TR064Services, b.TR064Services...)
	a.TLSCertSANs = appendUniqueStrings(a.TLSCertSANs, b.TLSCertSANs...)
	a.LoginLabels = appendUniqueStrings(a.LoginLabels, b.LoginLabels...)
	a.Notes = appendUniqueStrings(a.Notes, b.Notes...)

	a.HTTPTitle = firstNonEmpty(a.HTTPTitle, b.HTTPTitle)
	a.ServerHeader = firstNonEmpty(a.ServerHeader, b.ServerHeader)
	a.WWWAuthenticate = firstNonEmpty(a.WWWAuthenticate, b.WWWAuthenticate)
	a.WWWAuthRealm = firstNonEmpty(a.WWWAuthRealm, b.WWWAuthRealm)
	a.FaviconHash = firstNonEmpty(a.FaviconHash, b.FaviconHash)
	a.RedirectPath = firstNonEmpty(a.RedirectPath, b.RedirectPath)
	a.RedirectLocation = firstNonEmpty(a.RedirectLocation, b.RedirectLocation)
	a.HTMLMetaGenerator = firstNonEmpty(a.HTMLMetaGenerator, b.HTMLMetaGenerator)
	a.TLSCertCN = firstNonEmpty(a.TLSCertCN, b.TLSCertCN)
	a.TLSCertIssuer = firstNonEmpty(a.TLSCertIssuer, b.TLSCertIssuer)
	a.TLSServerName = firstNonEmpty(a.TLSServerName, b.TLSServerName)
	a.UPnPState = mergeFeatureState(a.UPnPState, b.UPnPState, a.UPnPFound || b.UPnPFound)
	a.TR064State = mergeFeatureState(a.TR064State, b.TR064State, a.TR064Found || b.TR064Found)
	a.SNMPState = mergeFeatureState(a.SNMPState, b.SNMPState, false)
	a.WANAccessType = firstNonEmpty(a.WANAccessType, b.WANAccessType)
	a.PhysicalLinkStatus = firstNonEmpty(a.PhysicalLinkStatus, b.PhysicalLinkStatus)
	a.MACVendor = firstNonEmpty(a.MACVendor, b.MACVendor)
	a.CPEModelGuess = firstNonEmpty(a.CPEModelGuess, b.CPEModelGuess)
	a.Model = firstNonEmpty(a.Model, b.Model)
	a.Manufacturer = firstNonEmpty(a.Manufacturer, b.Manufacturer)
	a.FingerprintID = firstNonEmpty(a.FingerprintID, b.FingerprintID)

	a.UPnPFound = a.UPnPFound || b.UPnPFound
	a.UPnPIGDFound = a.UPnPIGDFound || b.UPnPIGDFound
	a.WANCommonInterfaceFound = a.WANCommonInterfaceFound || b.WANCommonInterfaceFound
	a.TR064Found = a.TR064Found || b.TR064Found
	a.TR064AuthRequired = a.TR064AuthRequired || b.TR064AuthRequired
	if b.Layer1UpstreamMaxBitRate > a.Layer1UpstreamMaxBitRate {
		a.Layer1UpstreamMaxBitRate = b.Layer1UpstreamMaxBitRate
	}
	if b.Layer1DownstreamMaxBitRate > a.Layer1DownstreamMaxBitRate {
		a.Layer1DownstreamMaxBitRate = b.Layer1DownstreamMaxBitRate
	}
	a.DeviceConfidence = max(a.DeviceConfidence, b.DeviceConfidence)
	a.AccessConfidence = max(a.AccessConfidence, b.AccessConfidence)
	a.Confidence = max(a.Confidence, b.Confidence)
	return normalizeDevice(a)
}

func normalizeDevice(d models.GatewayDevice) models.GatewayDevice {
	if d.Reachable {
		d.ReachableState = string(models.TriTrue)
	} else if d.ReachableState == "" {
		d.ReachableState = string(models.TriUnknown)
	}
	if d.CPEModelGuess == "" {
		d.CPEModelGuess = strings.TrimSpace(strings.Join([]string{d.Manufacturer, d.Model}, " "))
	}
	if d.UPnPState == "" {
		d.UPnPState = featureState(d.UPnPFound)
	}
	if d.TR064State == "" {
		d.TR064State = featureState(d.TR064Found)
		if d.TR064AuthRequired && !d.TR064Found {
			d.TR064State = "auth_required"
		}
	}
	if d.SNMPState == "" {
		d.SNMPState = "skipped"
	}
	return d
}

func featureState(found bool) string {
	if found {
		return string(models.TriTrue)
	}
	return string(models.TriUnknown)
}

func mergeTriString(a, b string) string {
	if a == string(models.TriTrue) || b == string(models.TriTrue) {
		return string(models.TriTrue)
	}
	if a == string(models.TriFalse) || b == string(models.TriFalse) {
		return string(models.TriFalse)
	}
	return string(models.TriUnknown)
}

func mergeFeatureState(a, b string, found bool) string {
	if found || a == string(models.TriTrue) || b == string(models.TriTrue) || a == "found" || b == "found" {
		return string(models.TriTrue)
	}
	if a == string(models.TriFalse) || b == string(models.TriFalse) {
		return string(models.TriFalse)
	}
	return string(models.TriUnknown)
}
