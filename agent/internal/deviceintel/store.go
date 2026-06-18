package deviceintel

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

type EvidenceStore struct {
	Devices      map[string]*models.DeviceIntelDevice
	Observations []models.DeviceIntelEvidence
	Conflicts    []models.DataConflict

	now          func() time.Time
	seq          int
	evidenceSeen map[string]bool
}

func NewEvidenceStore(now func() time.Time) *EvidenceStore {
	if now == nil {
		now = time.Now
	}
	return &EvidenceStore{
		Devices:      map[string]*models.DeviceIntelDevice{},
		now:          now,
		evidenceSeen: map[string]bool{},
	}
}

func (s *EvidenceStore) UpsertDevice(ip string) *models.DeviceIntelDevice {
	id := deviceID(ip)
	if d := s.Devices[id]; d != nil {
		if ip != "" {
			d.IPAddresses = appendUnique(d.IPAddresses, ip)
		}
		return d
	}
	d := &models.DeviceIntelDevice{
		ID:          id,
		IPAddresses: appendUnique(nil, ip),
		Vendor: models.DeviceVendor{
			OUIVendor:         nil,
			FingerprintVendor: nil,
			Confidence:        0,
		},
		DeviceType: models.DeviceTypeGuess{
			Primary:         models.DeviceTypeUnknown,
			Confidence:      0,
			MissingEvidence: []string{"No device-specific protocol, model, OS, or role evidence has been observed yet."},
		},
		OSGuess: models.OSGuess{Family: "unknown", Confidence: 0},
		SNMPInfo: &models.SNMPInfo{
			Enabled: false,
			Status:  "skipped",
			Reason:  "SNMP requires explicit opt-in credentials.",
		},
		SecurityPosture: models.SecurityPosture{RiskLevel: "unknown"},
		Confidence:      0,
		LastSeen:        s.now().UTC().Format(time.RFC3339),
	}
	s.Devices[id] = d
	return d
}

func (s *EvidenceStore) RegisterEvidence(ev models.Evidence, deviceID string, confidence float64) string {
	if ev.ID == "" {
		return ""
	}
	if s.evidenceSeen[ev.ID] {
		return ev.ID
	}
	raw := make(map[string]any, len(ev.Data))
	for k, v := range ev.Data {
		raw[k] = v
	}
	s.Observations = append(s.Observations, models.DeviceIntelEvidence{
		ID:            ev.ID,
		DeviceID:      deviceID,
		SourceProbe:   ev.Source,
		Kind:          ev.Kind,
		Raw:           raw,
		Confidence:    clamp01(confidence),
		Timestamp:     ev.Timestamp.UTC(),
		SafeToDisplay: true,
	})
	s.evidenceSeen[ev.ID] = true
	return ev.ID
}

func (s *EvidenceStore) AddObservation(ip, source, kind string, raw, normalized map[string]any, confidence float64, errText string) string {
	s.seq++
	id := fmt.Sprintf("di-ev-%d", s.seq)
	deviceIDValue := ""
	if ip != "" {
		deviceIDValue = deviceID(ip)
	}
	ev := models.DeviceIntelEvidence{
		ID:            id,
		DeviceID:      deviceIDValue,
		SourceProbe:   source,
		Kind:          kind,
		Raw:           raw,
		Normalized:    normalized,
		Confidence:    clamp01(confidence),
		Timestamp:     s.now().UTC(),
		Error:         errText,
		SafeToDisplay: true,
	}
	s.Observations = append(s.Observations, ev)
	s.evidenceSeen[id] = true
	if ip != "" {
		d := s.UpsertDevice(ip)
		d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, id))
		d.LastSeen = ev.Timestamp.Format(time.RFC3339)
	}
	return id
}

func (s *EvidenceStore) AddFailedAttempt(ip string, attempt models.ProbeAttempt) {
	d := s.UpsertDevice(ip)
	if attempt.Target == "" {
		attempt.Target = ip
	}
	if attempt.EvidenceID == "" {
		attempt.EvidenceID = s.AddObservation(ip, attempt.Source, "probe_failure",
			map[string]any{"target": attempt.Target, "port": attempt.Port, "url": attempt.URL, "method": attempt.Method},
			nil, 0.05, attempt.Error)
	}
	for _, existing := range d.FailedAttempts {
		if existing.Source == attempt.Source && existing.Target == attempt.Target &&
			existing.Port == attempt.Port && existing.URL == attempt.URL && existing.Error == attempt.Error {
			return
		}
	}
	d.FailedAttempts = append(d.FailedAttempts, attempt)
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, attempt.EvidenceID))
}

func (s *EvidenceStore) AddIP(ip, evidenceID string, confidence float64) {
	d := s.UpsertDevice(ip)
	if evidenceID != "" {
		d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, evidenceID))
	}
	d.Confidence = maxFloat(d.Confidence, clamp01(confidence))
}

func (s *EvidenceStore) AddMAC(ip, mac, vendor, evidenceID string, confidence float64) {
	d := s.UpsertDevice(ip)
	d.MACAddresses = appendUnique(d.MACAddresses, mac)
	if vendor != "" {
		d.Vendor.OUIVendor = ptrString(vendor)
		d.Vendor.Confidence = maxFloat(d.Vendor.Confidence, clamp01(confidence))
	}
	if evidenceID != "" {
		d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, evidenceID))
	}
	if typ, ok := virtualTypeFromMAC(mac, vendor); ok {
		d.Roles = appendUnique(d.Roles, models.DeviceRoleVirtual)
		d.DeviceType.Candidates = mergeCandidate(d.DeviceType.Candidates, models.DeviceTypeCandidate{
			Type:            models.DeviceTypeVirtualMachine,
			Confidence:      0.70,
			SupportingFacts: []string{"Virtualization OUI observed in MAC address."},
			EvidenceIDs:     nonEmpty(evidenceID),
		})
		d.ClassificationExplanation = appendUnique(d.ClassificationExplanation, fmt.Sprintf("%s MAC/OUI suggests a virtual adapter.", typ))
	}
	d.Confidence = maxFloat(d.Confidence, clamp01(confidence))
}

func (s *EvidenceStore) AddHostname(ip, hostname, source, evidenceID string, confidence float64) {
	if hostname == "" {
		return
	}
	d := s.UpsertDevice(ip)
	d.Hostnames = appendUnique(d.Hostnames, strings.TrimSuffix(hostname, "."))
	if evidenceID != "" {
		d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, evidenceID))
	}
	lower := strings.ToLower(hostname)
	if strings.Contains(lower, "host.docker.internal") || strings.Contains(lower, "docker") || strings.Contains(lower, "wsl") {
		d.Roles = appendUnique(d.Roles, models.DeviceRoleVirtual)
		d.DeviceType.Candidates = mergeCandidate(d.DeviceType.Candidates, models.DeviceTypeCandidate{
			Type:            models.DeviceTypeVirtualMachine,
			Confidence:      0.45,
			SupportingFacts: []string{"Reverse DNS name is commonly generated by Docker/WSL virtualization."},
			EvidenceIDs:     nonEmpty(evidenceID),
		})
		d.SecurityPosture.Notes = appendUnique(d.SecurityPosture.Notes, "host.docker.internal is treated as low-confidence virtual/container context, not a physical device identity.")
		confidence = minFloat(confidence, 0.35)
	}
	d.Confidence = maxFloat(d.Confidence, clamp01(confidence))
	_ = source
}

func (s *EvidenceStore) AddService(ip string, svc models.DeviceIntelService) {
	if svc.Protocol == "" {
		svc.Protocol = "tcp"
	}
	if svc.State == "" {
		svc.State = "open"
	}
	if svc.Name == "" {
		svc.Name = serviceName(svc.Port)
	}
	if svc.Confidence == 0 {
		svc.Confidence = 0.75
	}
	d := s.UpsertDevice(ip)
	key := serviceKey(svc.Protocol, svc.Port)
	for i := range d.Services {
		if serviceKey(d.Services[i].Protocol, d.Services[i].Port) != key {
			continue
		}
		d.Services[i].State = strongerState(d.Services[i].State, svc.State)
		d.Services[i].Name = firstNonEmpty(d.Services[i].Name, svc.Name)
		d.Services[i].Product = firstNonEmpty(d.Services[i].Product, svc.Product)
		d.Services[i].Version = firstNonEmpty(d.Services[i].Version, svc.Version)
		d.Services[i].Banner = firstNonEmpty(d.Services[i].Banner, svc.Banner)
		d.Services[i].Confidence = maxFloat(d.Services[i].Confidence, svc.Confidence)
		d.Services[i].EvidenceIDs = sortedUnique(append(d.Services[i].EvidenceIDs, svc.EvidenceIDs...))
		d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, svc.EvidenceIDs...))
		return
	}
	svc.EvidenceIDs = sortedUnique(svc.EvidenceIDs)
	d.Services = append(d.Services, svc)
	sort.Slice(d.Services, func(i, j int) bool {
		if d.Services[i].Port != d.Services[j].Port {
			return d.Services[i].Port < d.Services[j].Port
		}
		return d.Services[i].Protocol < d.Services[j].Protocol
	})
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, svc.EvidenceIDs...))
	d.Confidence = maxFloat(d.Confidence, 0.70)
}

func (s *EvidenceStore) AddHTTPObservation(ip string, obs models.HTTPObservation, success bool, attempt models.ProbeAttempt) {
	d := s.UpsertDevice(ip)
	if !success {
		s.AddFailedAttempt(ip, attempt)
		return
	}
	for _, existing := range d.HTTPFingerprints {
		if existing.URL == obs.URL && existing.Method == obs.Method && existing.StatusCode == obs.StatusCode &&
			existing.ServerHeader == obs.ServerHeader && existing.Title == obs.Title {
			return
		}
	}
	d.HTTPFingerprints = append(d.HTTPFingerprints, obs)
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, obs.EvidenceID))
	d.Confidence = maxFloat(d.Confidence, 0.75)
}

func (s *EvidenceStore) AddTLSObservation(ip string, obs models.TLSObservation) {
	d := s.UpsertDevice(ip)
	for _, existing := range d.TLSFingerprints {
		if existing.Port == obs.Port && existing.CN == obs.CN && existing.Issuer == obs.Issuer {
			return
		}
	}
	d.TLSFingerprints = append(d.TLSFingerprints, obs)
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, obs.EvidenceID))
	d.Confidence = maxFloat(d.Confidence, 0.75)
}

func (s *EvidenceStore) AddMDNS(ip string, rec models.MDNSRecord) {
	d := s.UpsertDevice(ip)
	d.MDNSRecords = append(d.MDNSRecords, rec)
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, rec.EvidenceIDs...))
}

func (s *EvidenceStore) AddSSDP(ip string, rec models.SSDPRecord) {
	d := s.UpsertDevice(ip)
	d.SSDPRecords = append(d.SSDPRecords, rec)
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, rec.EvidenceIDs...))
}

func (s *EvidenceStore) AddNBNS(ip string, rec models.NBNSRecord) {
	d := s.UpsertDevice(ip)
	d.NBNSRecords = append(d.NBNSRecords, rec)
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, rec.EvidenceIDs...))
	if rec.Name != "" {
		d.Hostnames = appendUnique(d.Hostnames, rec.Name)
	}
}

func (s *EvidenceStore) AddLLMNR(ip string, rec models.LLMNRRecord) {
	d := s.UpsertDevice(ip)
	d.LLMNRRecords = append(d.LLMNRRecords, rec)
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, rec.EvidenceIDs...))
	if rec.Name != "" {
		d.Hostnames = appendUnique(d.Hostnames, rec.Name)
	}
}

func (s *EvidenceStore) AddConflict(conflict models.DataConflict) {
	s.Conflicts = append(s.Conflicts, conflict)
}

func (s *EvidenceStore) DeviceList() []models.DeviceIntelDevice {
	out := make([]models.DeviceIntelDevice, 0, len(s.Devices))
	for _, d := range s.Devices {
		d.EvidenceIDs = sortedUnique(d.EvidenceIDs)
		d.Roles = sortedUnique(d.Roles)
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool {
		return ipLess(firstIP(out[i]), firstIP(out[j]))
	})
	return out
}

func virtualTypeFromMAC(mac, vendor string) (string, bool) {
	key := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(mac, "-", ":"), ".", ""))
	compact := strings.ReplaceAll(key, ":", "")
	vendorLower := strings.ToLower(vendor)
	switch {
	case strings.HasPrefix(compact, "000569"), strings.HasPrefix(compact, "000C29"), strings.HasPrefix(compact, "001C14"), strings.HasPrefix(compact, "005056"), strings.Contains(vendorLower, "vmware"):
		return "VMware", true
	case strings.HasPrefix(compact, "080027"), strings.Contains(vendorLower, "virtualbox"):
		return "VirtualBox", true
	case strings.HasPrefix(compact, "00155D"), strings.Contains(vendorLower, "hyper-v"), strings.Contains(vendorLower, "microsoft"):
		return "Hyper-V", true
	default:
		return "", false
	}
}

func strongerState(existing, next string) string {
	if strings.EqualFold(existing, "open") || existing == "" {
		return firstNonEmpty(existing, next)
	}
	if strings.EqualFold(next, "open") {
		return next
	}
	return existing
}

func mergeCandidate(in []models.DeviceTypeCandidate, next models.DeviceTypeCandidate) []models.DeviceTypeCandidate {
	for i := range in {
		if in[i].Type != next.Type {
			continue
		}
		in[i].Confidence = maxFloat(in[i].Confidence, next.Confidence)
		in[i].SupportingFacts = appendUnique(in[i].SupportingFacts, next.SupportingFacts...)
		in[i].MissingEvidence = appendUnique(in[i].MissingEvidence, next.MissingEvidence...)
		in[i].EvidenceIDs = sortedUnique(append(in[i].EvidenceIDs, next.EvidenceIDs...))
		return in
	}
	next.EvidenceIDs = sortedUnique(next.EvidenceIDs)
	return append(in, next)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func nonEmpty(values ...string) []string {
	var out []string
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, v)
		}
	}
	return out
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
