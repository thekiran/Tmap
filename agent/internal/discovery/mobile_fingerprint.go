package discovery

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

type MobileDeviceFingerprintEngine struct {
	now func() time.Time
}

func NewMobileDeviceFingerprintEngine(now func() time.Time) MobileDeviceFingerprintEngine {
	if now == nil {
		now = time.Now
	}
	return MobileDeviceFingerprintEngine{now: now}
}

func ApplyMobileFingerprints(devices []models.Device, evidence []models.Evidence, now time.Time) []models.Device {
	engine := NewMobileDeviceFingerprintEngine(func() time.Time {
		if !now.IsZero() {
			return now
		}
		return time.Now()
	})
	out := make([]models.Device, len(devices))
	copy(out, devices)
	for i := range out {
		fp := engine.FingerprintDevice(out[i], evidence)
		out[i].MobileFingerprint = &fp
		out[i].OSHint = MobileOSHint(fp.Classification)
		out[i].OSConfidence = fp.Confidence
		out[i].DeviceTypeHint = MobileDeviceTypeHint(fp, append([]string{out[i].Hostname}, out[i].Hostnames...))
	}
	return out
}

func (e MobileDeviceFingerprintEngine) FingerprintDevice(device models.Device, evidence []models.Evidence) models.MobileFingerprint {
	return e.Fingerprint(mobileInputFromDevice(device, evidence, e.timestamp()))
}

func (e MobileDeviceFingerprintEngine) FingerprintDeviceIntel(device models.DeviceIntelDevice, evidence []models.Evidence) models.MobileFingerprint {
	return e.Fingerprint(mobileInputFromIntelDevice(device, evidence, e.timestamp()))
}

func (e MobileDeviceFingerprintEngine) Fingerprint(in MobileFingerprintInput) models.MobileFingerprint {
	ts := in.Timestamp
	if ts.IsZero() {
		ts = e.timestamp()
	}
	scorer := newMobileScorer(ts)
	scorer.evaluateHostnames(in.Hostnames, models.MobileEvidenceTypeHostname)
	scorer.evaluateHostnames(in.DHCPHostnames, models.MobileEvidenceTypeDHCPHostname)
	scorer.evaluateMACs(in.MACAddresses, in.OUIVendors)
	scorer.evaluateVendorHints(in.VendorHints)
	scorer.evaluateMDNS(in.MDNSRecords)
	scorer.evaluateDNS(in.DNSQueries)
	scorer.evaluateServices(in.Services)
	scorer.evaluateUserAgents(in.UserAgents)
	scorer.evaluatePassivePackets(in.PassivePackets)
	return scorer.finish()
}

func (e MobileDeviceFingerprintEngine) timestamp() time.Time {
	if e.now != nil {
		return e.now().UTC()
	}
	return time.Now().UTC()
}

func MobileOSHint(classification string) string {
	switch classification {
	case models.MobileClassificationConfirmedIOS, models.MobileClassificationProbableIOS, models.MobileClassificationPossibleIOS:
		return models.MobileOSHintIOS
	case models.MobileClassificationConfirmedIPadOS, models.MobileClassificationProbableIPadOS, models.MobileClassificationPossibleIPadOS:
		return models.MobileOSHintIPadOS
	case models.MobileClassificationConfirmedAndroid, models.MobileClassificationProbableAndroid, models.MobileClassificationPossibleAndroid:
		return models.MobileOSHintAndroid
	default:
		return models.MobileOSHintUnknown
	}
}

func MobileDeviceTypeHint(fp models.MobileFingerprint, names []string) string {
	switch MobileOSHint(fp.Classification) {
	case models.MobileOSHintIOS:
		return models.MobileDeviceTypeHintPhone
	case models.MobileOSHintIPadOS:
		return models.MobileDeviceTypeHintTablet
	case models.MobileOSHintAndroid:
		if mobileNameContains(names, "tablet", " tab", "-tab", "pad", "lenovo tab") {
			return models.MobileDeviceTypeHintTablet
		}
		if fp.Classification == models.MobileClassificationPossibleAndroid {
			return models.MobileDeviceTypeHintUnknown
		}
		return models.MobileDeviceTypeHintPhone
	default:
		if fp.Classification == models.MobileClassificationUnknownMobile {
			return models.MobileDeviceTypeHintUnknown
		}
		return models.MobileDeviceTypeHintUnknown
	}
}

type mobileScorer struct {
	now               time.Time
	seq               int
	iosScore          int
	ipadScore         int
	androidScore      int
	evidence          []models.MobileEvidenceItem
	conflicts         []models.MobileConflictItem
	warnings          []string
	categories        map[string]map[string]bool
	strongCounts      map[string]int
	idsByOS           map[string][]string
	vendorIDsByOS     map[string][]string
	identityIDsByOS   map[string][]string
	randomizedMAC     bool
	onlyWeakRandomMAC bool
	sawWeakWebOnly    bool
	sawDNS            bool
}

func newMobileScorer(now time.Time) *mobileScorer {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return &mobileScorer{
		now:             now.UTC(),
		categories:      map[string]map[string]bool{},
		strongCounts:    map[string]int{},
		idsByOS:         map[string][]string{},
		vendorIDsByOS:   map[string][]string{},
		identityIDsByOS: map[string][]string{},
	}
}

func (s *mobileScorer) evaluateHostnames(values []MobileObservedValue, evidenceType string) {
	for _, value := range values {
		raw := strings.TrimSpace(value.Value)
		if raw == "" {
			continue
		}
		lower := strings.ToLower(raw)
		switch {
		case strings.Contains(lower, "iphone"):
			s.addIOS(evidenceType, raw, 45, models.MobileEvidenceStrengthStrong, firstMobileSource(value.Source, evidenceType), value.Timestamp, "Hostname explicitly contains iPhone.")
		case strings.Contains(lower, "ipad"):
			s.addIPad(evidenceType, raw, 45, models.MobileEvidenceStrengthStrong, firstMobileSource(value.Source, evidenceType), value.Timestamp, "Hostname explicitly contains iPad.")
		case mobileContainsStandalone(lower, "ios"):
			s.addIOS(evidenceType, raw, 30, models.MobileEvidenceStrengthMedium, firstMobileSource(value.Source, evidenceType), value.Timestamp, "Hostname contains an iOS token.")
		case containsAnyMobile(lower, "android", "galaxy", "samsung", "redmi", "xiaomi", "pixel", "oneplus", "oppo", "vivo", "realme", "huawei", "honor", "moto", "motorola", "nothing"):
			s.addAndroid(evidenceType, raw, 45, models.MobileEvidenceStrengthStrong, firstMobileSource(value.Source, evidenceType), value.Timestamp, "Hostname contains an explicit Android or Android-vendor/model-family token.")
		}
	}
}

func (s *mobileScorer) evaluateMACs(macs, vendors []string) {
	vendorByMAC := map[string]string{}
	for _, vendor := range vendors {
		vendor = strings.TrimSpace(vendor)
		if vendor == "" {
			continue
		}
		family := mobileVendorFamily(vendor)
		switch family {
		case models.MobileOSHintIOS:
			s.addApple(models.MobileEvidenceTypeVendorHint, vendor, 20, models.MobileEvidenceStrengthMedium, "vendor_hint", s.now, "Vendor string is Apple; this supports iOS/iPadOS only with other evidence.")
		case models.MobileOSHintAndroid:
			s.addAndroid(models.MobileEvidenceTypeVendorHint, vendor, 20, models.MobileEvidenceStrengthMedium, "vendor_hint", s.now, "Vendor string matches a common Android device manufacturer.")
		}
		vendorByMAC[strings.ToLower(vendor)] = vendor
	}

	for _, rawMAC := range macs {
		mac := strings.TrimSpace(rawMAC)
		if mac == "" {
			continue
		}
		local := isLocallyAdministeredMAC(mac)
		if local {
			s.randomizedMAC = true
			s.warn("MAC randomization/private Wi-Fi addresses can reduce mobile OS accuracy.")
		}
		vendor := mobileOUIVendor(mac)
		if vendor == "" {
			for key, value := range vendorByMAC {
				if key != "" && strings.Contains(strings.ToLower(mac), key) {
					vendor = value
					break
				}
			}
		}
		family := mobileVendorFamily(vendor)
		if family == "" {
			if local {
				s.onlyWeakRandomMAC = true
			}
			continue
		}
		impact := 35
		strength := models.MobileEvidenceStrengthMedium
		explanation := fmt.Sprintf("Global MAC/OUI vendor %q is a mobile vendor signal, not standalone proof.", vendor)
		if local {
			impact = 8
			strength = models.MobileEvidenceStrengthWeak
			explanation = fmt.Sprintf("MAC is locally administered/randomized; vendor hint %q is downgraded to weak support.", vendor)
			s.onlyWeakRandomMAC = true
		}
		value := fmt.Sprintf("%s (%s)", redactMACForEvidence(mac), vendor)
		switch family {
		case models.MobileOSHintIOS:
			s.addApple(models.MobileEvidenceTypeMACOUI, value, impact, strength, "mac_oui", s.now, explanation)
		case models.MobileOSHintAndroid:
			s.addAndroid(models.MobileEvidenceTypeMACOUI, value, impact, strength, "mac_oui", s.now, explanation)
		}
	}
}

func (s *mobileScorer) evaluateVendorHints(values []MobileObservedValue) {
	for _, value := range values {
		vendor := strings.TrimSpace(value.Value)
		if vendor == "" {
			continue
		}
		switch mobileVendorFamily(vendor) {
		case models.MobileOSHintIOS:
			s.addApple(models.MobileEvidenceTypeVendorHint, vendor, 20, models.MobileEvidenceStrengthMedium, firstMobileSource(value.Source, "vendor_hint"), value.Timestamp, "Vendor hint is Apple; it supports iOS/iPadOS only with other evidence.")
		case models.MobileOSHintAndroid:
			s.addAndroid(models.MobileEvidenceTypeVendorHint, vendor, 20, models.MobileEvidenceStrengthMedium, firstMobileSource(value.Source, "vendor_hint"), value.Timestamp, "Vendor hint matches a common Android device manufacturer.")
		}
	}
}

func (s *mobileScorer) evaluateMDNS(records []MobileMDNSRecord) {
	for _, rec := range records {
		combined := strings.TrimSpace(strings.Join([]string{rec.Name, rec.Service, rec.Target, strings.Join(rec.Text, " ")}, " "))
		lower := strings.ToLower(combined)
		if combined == "" {
			continue
		}
		source := firstMobileSource(rec.Source, "mdns")
		switch {
		case strings.Contains(lower, "iphone"):
			s.addIOS(models.MobileEvidenceTypeMDNS, combined, 45, models.MobileEvidenceStrengthStrong, source, rec.Timestamp, "mDNS metadata explicitly contains iPhone.")
		case strings.Contains(lower, "ipad"):
			s.addIPad(models.MobileEvidenceTypeMDNS, combined, 45, models.MobileEvidenceStrengthStrong, source, rec.Timestamp, "mDNS metadata explicitly contains iPad.")
		case containsAnyMobile(lower, "_apple-mobdev2._tcp", "_companion-link._tcp"):
			s.addApple(models.MobileEvidenceTypeMDNS, combined, 30, models.MobileEvidenceStrengthStrong, source, rec.Timestamp, "Apple-specific Bonjour service metadata was observed.")
		case containsAnyMobile(lower, "_airplay._tcp", "_raop._tcp"):
			s.addApple(models.MobileEvidenceTypeMDNS, combined, 30, models.MobileEvidenceStrengthMedium, source, rec.Timestamp, "AirPlay/RAOP Bonjour service metadata supports Apple identity but is not Apple-only by itself.")
		case strings.Contains(lower, "android"):
			s.addAndroid(models.MobileEvidenceTypeMDNS, combined, 30, models.MobileEvidenceStrengthStrong, source, rec.Timestamp, "mDNS/DNS-SD metadata explicitly contains Android.")
		case containsAnyMobile(lower, "_googlecast._tcp", "chromecast"):
			if s.androidScore > 0 {
				s.addAndroid(models.MobileEvidenceTypeMDNS, combined, 15, models.MobileEvidenceStrengthMedium, source, rec.Timestamp, "Google Cast metadata is treated as Android support only because other Android evidence exists.")
			}
		case strings.Contains(lower, "_services._dns-sd._udp") || strings.Contains(lower, "_mdns"):
			s.warn("mDNS/DNS-SD presence alone is not used as Apple or Android proof.")
		}
	}
}

func (s *mobileScorer) evaluateDNS(values []MobileObservedValue) {
	appleDomains := map[string]bool{}
	androidDomains := map[string]bool{}
	for _, value := range values {
		domain := normalizeMobileDomain(value.Value)
		if domain == "" {
			continue
		}
		if matched := matchedMobileDomain(domain, appleDomainHints); matched != "" {
			appleDomains[matched] = true
		}
		if matched := matchedMobileDomain(domain, androidDomainHints); matched != "" {
			androidDomains[matched] = true
		}
	}
	if len(appleDomains) > 0 {
		s.sawDNS = true
		domains := sortedMapKeys(appleDomains)
		impact, strength := dnsImpact(domains, true)
		s.addApple(models.MobileEvidenceTypeDNSQuery, strings.Join(domains, ", "), impact, strength, "passive_dns_metadata", s.now, "Only matched Apple OS-related domain hints are stored; DNS evidence alone cannot confirm iOS/iPadOS.")
	}
	if len(androidDomains) > 0 {
		s.sawDNS = true
		domains := sortedMapKeys(androidDomains)
		impact, strength := dnsImpact(domains, false)
		s.addAndroid(models.MobileEvidenceTypeDNSQuery, strings.Join(domains, ", "), impact, strength, "passive_dns_metadata", s.now, "Only matched Android/Google service domain hints are stored; DNS evidence alone cannot confirm Android.")
	}
	if s.sawDNS {
		s.warn("Passive DNS metadata is reduced to matched OS-related domain hints; raw browsing history and packet payloads are not stored.")
	}
}

func (s *mobileScorer) evaluateServices(services []MobileServiceObservation) {
	for _, svc := range services {
		protocol := strings.ToLower(firstMobileSource(svc.Protocol, "tcp"))
		nameProduct := strings.ToLower(strings.TrimSpace(svc.Name + " " + svc.Product))
		value := fmt.Sprintf("%s/%d", protocol, svc.Port)
		if svc.Name != "" {
			value += " " + svc.Name
		}
		switch {
		case protocol == "tcp" && (svc.Port == 5223 || svc.Port == 2197):
			s.addApple(models.MobileEvidenceTypeServicePort, value, 25, models.MobileEvidenceStrengthMedium, firstMobileSource(svc.Source, "service_port"), svc.Timestamp, "APNS-related TCP port observed; this is supporting Apple evidence, not standalone proof.")
		case protocol == "udp" && svc.Port == 5353:
			s.warn("UDP 5353/mDNS alone is not used as iOS or Android proof.")
		case (svc.Port == 80 || svc.Port == 443) && (protocol == "tcp" || protocol == "udp"):
			s.sawWeakWebOnly = true
		case protocol == "udp" && ((svc.Port >= 3478 && svc.Port <= 3497) || (svc.Port >= 16384 && svc.Port <= 16403)):
			if s.iosScore > 0 || s.ipadScore > 0 {
				s.addApple(models.MobileEvidenceTypeServicePort, value, 10, models.MobileEvidenceStrengthWeak, firstMobileSource(svc.Source, "service_port"), svc.Timestamp, "FaceTime/Game Center related UDP range supports Apple only because other Apple evidence exists.")
			}
		case protocol == "tcp" && (svc.Port == 3689 || svc.Port == 5000 || svc.Port == 6000 || svc.Port == 7000):
			if containsAnyMobile(nameProduct, "airplay", "raop", "daap") || s.iosScore > 0 || s.ipadScore > 0 {
				s.addApple(models.MobileEvidenceTypeServicePort, value, 15, models.MobileEvidenceStrengthMedium, firstMobileSource(svc.Source, "service_port"), svc.Timestamp, "AirPlay/DAAP service metadata supports Apple identity when paired with service naming or other Apple evidence.")
			}
		case (svc.Port == 8008 || svc.Port == 8009) && containsAnyMobile(nameProduct, "chromecast", "googlecast"):
			if s.androidScore > 0 {
				s.addAndroid(models.MobileEvidenceTypeServicePort, value, 15, models.MobileEvidenceStrengthMedium, firstMobileSource(svc.Source, "service_port"), svc.Timestamp, "Google Cast service metadata supports Android only because other Android evidence exists.")
			}
		}

		if containsAnyMobile(nameProduct, "airplay", "raop", "apple-mobdev2") {
			s.addApple(models.MobileEvidenceTypeServiceName, strings.TrimSpace(svc.Name+" "+svc.Product), 15, models.MobileEvidenceStrengthMedium, firstMobileSource(svc.Source, "service_name"), svc.Timestamp, "Service name/product contains Apple-specific local service metadata.")
		}
		if containsAnyMobile(nameProduct, "android", "googlecast") && s.androidScore > 0 {
			s.addAndroid(models.MobileEvidenceTypeServiceName, strings.TrimSpace(svc.Name+" "+svc.Product), 15, models.MobileEvidenceStrengthMedium, firstMobileSource(svc.Source, "service_name"), svc.Timestamp, "Service name/product contains Android/Google local service metadata and is paired with other Android evidence.")
		}
	}
	if s.sawWeakWebOnly {
		s.warn("TCP/UDP 80 and 443 are common on many devices and are not used as mobile OS proof.")
	}
}

func (s *mobileScorer) evaluateUserAgents(values []MobileObservedValue) {
	for _, value := range values {
		ua := strings.TrimSpace(value.Value)
		if ua == "" {
			continue
		}
		lower := strings.ToLower(ua)
		switch {
		case strings.Contains(lower, "iphone"):
			s.addIOS(models.MobileEvidenceTypeUserAgent, safeUserAgentValue(ua), 45, models.MobileEvidenceStrengthStrong, firstMobileSource(value.Source, "clear_local_http_user_agent"), value.Timestamp, "Clear local HTTP User-Agent explicitly contains iPhone.")
		case strings.Contains(lower, "ipad"):
			s.addIPad(models.MobileEvidenceTypeUserAgent, safeUserAgentValue(ua), 45, models.MobileEvidenceStrengthStrong, firstMobileSource(value.Source, "clear_local_http_user_agent"), value.Timestamp, "Clear local HTTP User-Agent explicitly contains iPad.")
		case strings.Contains(lower, "android"):
			s.addAndroid(models.MobileEvidenceTypeUserAgent, safeUserAgentValue(ua), 45, models.MobileEvidenceStrengthStrong, firstMobileSource(value.Source, "clear_local_http_user_agent"), value.Timestamp, "Clear local HTTP User-Agent explicitly contains Android.")
		}
	}
	if len(values) > 0 {
		s.warn("User-Agent evidence is only used when visible in clear local HTTP metadata; HTTPS is not decrypted or intercepted.")
	}
}

func (s *mobileScorer) evaluatePassivePackets(packets []MobilePassivePacket) {
	for _, packet := range packets {
		protocol := strings.ToLower(packet.Protocol)
		value := strings.ToLower(packet.Value)
		switch {
		case strings.Contains(value, "iphone"):
			s.addIOS(models.MobileEvidenceTypePassivePacket, packet.Value, 30, models.MobileEvidenceStrengthMedium, firstMobileSource(packet.Source, "passive_packet_metadata"), packet.Timestamp, "Passive metadata contains iPhone without storing packet payload.")
		case strings.Contains(value, "ipad"):
			s.addIPad(models.MobileEvidenceTypePassivePacket, packet.Value, 30, models.MobileEvidenceStrengthMedium, firstMobileSource(packet.Source, "passive_packet_metadata"), packet.Timestamp, "Passive metadata contains iPad without storing packet payload.")
		case strings.Contains(value, "android"):
			s.addAndroid(models.MobileEvidenceTypePassivePacket, packet.Value, 30, models.MobileEvidenceStrengthMedium, firstMobileSource(packet.Source, "passive_packet_metadata"), packet.Timestamp, "Passive metadata contains Android without storing packet payload.")
		case protocol == "udp" && (packet.SourcePort == 5353 || packet.DestinationPort == 5353):
			s.warn("Passive UDP 5353/mDNS packet metadata alone is not used as mobile OS proof.")
		case packet.SourcePort == 443 || packet.DestinationPort == 443 || packet.SourcePort == 80 || packet.DestinationPort == 80:
			s.sawWeakWebOnly = true
		}
	}
}

func (s *mobileScorer) finish() models.MobileFingerprint {
	s.detectConflicts()

	classification := s.chooseClassification()
	confidence := s.confidenceFor(classification)
	fp := models.MobileFingerprint{
		Classification:        classification,
		IOSScore:              s.iosScore,
		AndroidScore:          s.androidScore,
		IPadScore:             s.ipadScore,
		Confidence:            confidence,
		Evidence:              sortedMobileEvidence(s.evidence),
		Conflicts:             s.conflicts,
		Warnings:              sortedUnique(s.warnings),
		LastUpdatedAt:         s.now.Format(time.RFC3339),
		WhyThisClassification: s.whyThisClassification(classification),
		WhyNotCertain:         s.whyNotCertain(classification),
	}
	return fp
}

func (s *mobileScorer) chooseClassification() string {
	if len(s.conflicts) > 0 && hasWarningMobileConflict(s.conflicts) {
		return models.MobileClassificationConflict
	}
	appleScore := maxInt(s.iosScore, s.ipadScore)
	if appleScore >= 60 && s.androidScore >= 60 {
		return models.MobileClassificationConflict
	}

	type candidate struct {
		os    string
		score int
	}
	candidates := []candidate{
		{os: models.MobileOSHintIOS, score: s.iosScore},
		{os: models.MobileOSHintIPadOS, score: s.ipadScore},
		{os: models.MobileOSHintAndroid, score: s.androidScore},
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return osRank(candidates[i].os) < osRank(candidates[j].os)
	})
	top := candidates[0]
	if top.score == 0 {
		return models.MobileClassificationUnknownDevice
	}
	if top.score < 45 && !s.dnsOnly(top.os) {
		return models.MobileClassificationUnknownMobile
	}
	return s.classificationForOS(top.os, top.score)
}

func (s *mobileScorer) classificationForOS(os string, score int) string {
	confirmed := score >= 85 && s.strongCounts[os] >= 2 && len(s.categories[os]) >= 2 && !s.dnsOnly(os) && !s.macOnlyRandomized(os)
	probable := score >= 70 && len(s.categories[os]) >= 2 && !s.dnsOnly(os) && !s.macOnlyRandomized(os)
	possible := score >= 45 || s.dnsOnly(os)
	switch os {
	case models.MobileOSHintIPadOS:
		if confirmed {
			return models.MobileClassificationConfirmedIPadOS
		}
		if probable {
			return models.MobileClassificationProbableIPadOS
		}
		if possible {
			return models.MobileClassificationPossibleIPadOS
		}
	case models.MobileOSHintAndroid:
		if confirmed {
			return models.MobileClassificationConfirmedAndroid
		}
		if probable {
			return models.MobileClassificationProbableAndroid
		}
		if possible {
			return models.MobileClassificationPossibleAndroid
		}
	default:
		if confirmed {
			return models.MobileClassificationConfirmedIOS
		}
		if probable {
			return models.MobileClassificationProbableIOS
		}
		if possible {
			return models.MobileClassificationPossibleIOS
		}
	}
	return models.MobileClassificationUnknownMobile
}

func (s *mobileScorer) confidenceFor(classification string) float64 {
	top := maxInt(s.androidScore, maxInt(s.iosScore, s.ipadScore))
	switch classification {
	case models.MobileClassificationConflict:
		return 0.50
	case models.MobileClassificationConfirmedIOS, models.MobileClassificationConfirmedIPadOS, models.MobileClassificationConfirmedAndroid:
		return clampMobile(float64(top)/100, 0.85, 0.98)
	case models.MobileClassificationProbableIOS, models.MobileClassificationProbableIPadOS, models.MobileClassificationProbableAndroid:
		return clampMobile(float64(top)/100, 0.70, 0.84)
	case models.MobileClassificationPossibleIOS, models.MobileClassificationPossibleIPadOS, models.MobileClassificationPossibleAndroid:
		return clampMobile(float64(top)/100, 0.45, 0.69)
	case models.MobileClassificationUnknownMobile:
		return 0.25
	default:
		return 0
	}
}

func (s *mobileScorer) whyThisClassification(classification string) string {
	switch classification {
	case models.MobileClassificationConflict:
		return "Apple/iOS and Android evidence both exist, so the classifier surfaces the conflict instead of forcing one OS."
	case models.MobileClassificationUnknownDevice:
		return "No explicit mobile OS evidence was observed from hostname, vendor/OUI, service metadata, DNS hints, or clear local HTTP metadata."
	case models.MobileClassificationUnknownMobile:
		return "Some weak mobile-related metadata was observed, but it is below the threshold for a possible OS classification."
	default:
		os := MobileOSHint(classification)
		score := s.scoreForOS(os)
		return fmt.Sprintf("%s evidence reached score %d using %d evidence categor%s.", displayMobileOS(os), score, len(s.categories[os]), pluralY(len(s.categories[os])))
	}
}

func (s *mobileScorer) whyNotCertain(classification string) string {
	switch classification {
	case models.MobileClassificationConfirmedIOS, models.MobileClassificationConfirmedIPadOS, models.MobileClassificationConfirmedAndroid:
		return "Confirmed means strong local metadata agrees; it is still not a hardware attestation and can be affected by private MAC addresses or shared services."
	case models.MobileClassificationProbableIOS, models.MobileClassificationProbableIPadOS, models.MobileClassificationProbableAndroid:
		return "The evidence spans multiple categories, but it lacks enough strong independent direct evidence for a confirmed result."
	case models.MobileClassificationPossibleIOS, models.MobileClassificationPossibleIPadOS, models.MobileClassificationPossibleAndroid:
		return "The evidence points toward this OS, but it is weak, single-category, DNS-only, or otherwise insufficient for probable classification."
	case models.MobileClassificationConflict:
		return "Conflicting hostname, vendor, DNS, or service evidence needs user review or additional passive observations."
	default:
		return "Open web ports, mDNS presence, random high ports, latency, TTL, PTR, public IP, ASN, and generic traffic metadata are not enough to identify a mobile OS."
	}
}

func (s *mobileScorer) detectConflicts() {
	appleIDs := sortedUnique(append(append([]string{}, s.idsByOS[models.MobileOSHintIOS]...), s.idsByOS[models.MobileOSHintIPadOS]...))
	androidIDs := sortedUnique(s.idsByOS[models.MobileOSHintAndroid])
	if maxInt(s.iosScore, s.ipadScore) >= 60 && s.androidScore >= 60 {
		s.addConflict("iOS/iPadOS and Android scores are both high.", appleIDs, androidIDs, models.MobileConflictSeverityWarning, "Review hostname, vendor/OUI, mDNS, DNS, and user-agent evidence before assigning a mobile OS.")
	}

	appleIdentity := sortedUnique(append(append([]string{}, s.identityIDsByOS[models.MobileOSHintIOS]...), s.identityIDsByOS[models.MobileOSHintIPadOS]...))
	androidIdentity := sortedUnique(s.identityIDsByOS[models.MobileOSHintAndroid])
	appleVendor := sortedUnique(append(append([]string{}, s.vendorIDsByOS[models.MobileOSHintIOS]...), s.vendorIDsByOS[models.MobileOSHintIPadOS]...))
	androidVendor := sortedUnique(s.vendorIDsByOS[models.MobileOSHintAndroid])
	if len(appleIdentity) > 0 && len(androidVendor) > 0 {
		s.addConflict("Hostname/User-Agent says Apple mobile, but vendor/OUI suggests Android.", appleIdentity, androidVendor, models.MobileConflictSeverityWarning, "Treat as conflicting until a second independent source agrees.")
	}
	if len(androidIdentity) > 0 && len(appleVendor) > 0 {
		s.addConflict("Hostname/User-Agent says Android, but vendor/OUI suggests Apple.", appleVendor, androidIdentity, models.MobileConflictSeverityWarning, "Treat as conflicting until a second independent source agrees.")
	}
}

func (s *mobileScorer) addConflict(reason string, iosIDs, androidIDs []string, severity, hint string) {
	if len(iosIDs) == 0 && len(androidIDs) == 0 {
		return
	}
	key := reason + "|" + strings.Join(iosIDs, ",") + "|" + strings.Join(androidIDs, ",")
	for _, conflict := range s.conflicts {
		existing := conflict.Reason + "|" + strings.Join(conflict.IOSEvidenceIDs, ",") + "|" + strings.Join(conflict.AndroidEvidenceIDs, ",")
		if existing == key {
			return
		}
	}
	s.conflicts = append(s.conflicts, models.MobileConflictItem{
		Reason:             reason,
		IOSEvidenceIDs:     sortedUnique(iosIDs),
		AndroidEvidenceIDs: sortedUnique(androidIDs),
		Severity:           severity,
		ResolutionHint:     hint,
	})
}

func (s *mobileScorer) addIOS(evidenceType, value string, impact int, strength, source string, ts time.Time, explanation string) {
	s.iosScore += impact
	s.addEvidence(models.MobileOSHintIOS, evidenceType, value, impact, strength, source, ts, explanation)
}

func (s *mobileScorer) addIPad(evidenceType, value string, impact int, strength, source string, ts time.Time, explanation string) {
	s.ipadScore += impact
	s.addEvidence(models.MobileOSHintIPadOS, evidenceType, value, impact, strength, source, ts, explanation)
}

func (s *mobileScorer) addApple(evidenceType, value string, impact int, strength, source string, ts time.Time, explanation string) {
	s.iosScore += impact
	s.ipadScore += impact
	id := s.addEvidence(models.MobileOSHintIOS, evidenceType, value, impact, strength, source, ts, explanation)
	s.trackCategory(models.MobileOSHintIPadOS, evidenceType, strength, id)
}

func (s *mobileScorer) addAndroid(evidenceType, value string, impact int, strength, source string, ts time.Time, explanation string) {
	s.androidScore += impact
	s.addEvidence(models.MobileOSHintAndroid, evidenceType, value, impact, strength, source, ts, explanation)
}

func (s *mobileScorer) addEvidence(osHint, evidenceType, value string, impact int, strength, source string, ts time.Time, explanation string) string {
	if ts.IsZero() {
		ts = s.now
	}
	s.seq++
	id := fmt.Sprintf("mobile-%s-%s-%02d", osHint, strings.ReplaceAll(evidenceType, "_", "-"), s.seq)
	item := models.MobileEvidenceItem{
		ID:               id,
		Type:             evidenceType,
		Value:            value,
		OSHint:           osHint,
		ConfidenceImpact: impact,
		Strength:         strength,
		Source:           firstMobileSource(source, "mobile_fingerprint"),
		Timestamp:        ts.UTC().Format(time.RFC3339),
		Explanation:      explanation,
	}
	s.evidence = append(s.evidence, item)
	s.trackCategory(osHint, evidenceType, strength, id)
	return id
}

func (s *mobileScorer) trackCategory(osHint, evidenceType, strength, id string) {
	if osHint == models.MobileOSHintUnknown {
		return
	}
	if s.categories[osHint] == nil {
		s.categories[osHint] = map[string]bool{}
	}
	s.categories[osHint][evidenceType] = true
	if strength == models.MobileEvidenceStrengthStrong {
		s.strongCounts[osHint]++
	}
	s.idsByOS[osHint] = appendUniqueString(s.idsByOS[osHint], id)
	switch evidenceType {
	case models.MobileEvidenceTypeMACOUI, models.MobileEvidenceTypeVendorHint:
		s.vendorIDsByOS[osHint] = appendUniqueString(s.vendorIDsByOS[osHint], id)
	case models.MobileEvidenceTypeHostname, models.MobileEvidenceTypeDHCPHostname, models.MobileEvidenceTypeUserAgent:
		s.identityIDsByOS[osHint] = appendUniqueString(s.identityIDsByOS[osHint], id)
	}
}

func (s *mobileScorer) warn(message string) {
	s.warnings = appendUniqueString(s.warnings, message)
}

func (s *mobileScorer) dnsOnly(os string) bool {
	cats := s.categories[os]
	return len(cats) == 1 && cats[models.MobileEvidenceTypeDNSQuery]
}

func (s *mobileScorer) macOnlyRandomized(os string) bool {
	cats := s.categories[os]
	return s.randomizedMAC && len(cats) == 1 && cats[models.MobileEvidenceTypeMACOUI]
}

func (s *mobileScorer) scoreForOS(os string) int {
	switch os {
	case models.MobileOSHintIPadOS:
		return s.ipadScore
	case models.MobileOSHintAndroid:
		return s.androidScore
	default:
		return s.iosScore
	}
}

var appleDomainHints = []string{
	"apple.com",
	"icloud.com",
	"mzstatic.com",
	"itunes.apple.com",
	"push.apple.com",
	"aaplimg.com",
	"captive.apple.com",
}

var androidDomainHints = []string{
	"android.clients.google.com",
	"clients3.google.com",
	"gstatic.com",
	"googleapis.com",
	"mtalk.google.com",
	"gvt1.com",
	"play.googleapis.com",
	"connectivitycheck.gstatic.com",
	"android.googleapis.com",
}

var mobileOUIPrefixes = map[string]string{
	"A483E7": "Apple",
	"E89C25": "Apple",
	"F4F5D8": "Apple",
	"74D4DD": "Apple",
	"D8BBC1": "Apple",
	"3C5AB4": "Google",
	"F88FCA": "Google",
	"001A11": "Google",
	"EC1F72": "Samsung",
	"70F1A1": "Samsung",
	"283926": "Samsung",
	"9852B1": "Samsung",
	"ACAFB9": "Samsung",
	"D88039": "Xiaomi",
	"64CC2E": "Xiaomi",
	"7802F8": "Xiaomi",
	"344DF7": "OnePlus",
	"94652D": "OnePlus",
	"A4C0E1": "OPPO",
	"487412": "OPPO",
	"703A0E": "Vivo",
	"9078B2": "Huawei",
	"009ACD": "Huawei",
	"38378B": "Honor",
	"4C7766": "Motorola",
	"582059": "Nothing",
	"304596": "Sony",
}

func mobileInputFromDevice(device models.Device, evidence []models.Evidence, now time.Time) MobileFingerprintInput {
	input := MobileFingerprintInput{DeviceID: device.ID, Timestamp: now}
	input.Hostnames = appendObserved(input.Hostnames, device.Hostname, "device_hostname", now)
	for _, hostname := range device.Hostnames {
		input.Hostnames = appendObserved(input.Hostnames, hostname, "device_hostname", now)
	}
	input.MACAddresses = appendUniqueString(input.MACAddresses, device.MAC)
	input.OUIVendors = appendUniqueString(input.OUIVendors, device.OUIVendor)
	input.VendorHints = appendObserved(input.VendorHints, device.Vendor, "device_vendor", now)
	for _, iface := range device.Interfaces {
		input.MACAddresses = appendUniqueString(input.MACAddresses, iface.MAC)
		input.OUIVendors = appendUniqueString(input.OUIVendors, iface.Vendor)
	}
	for _, svc := range device.Services {
		input.Services = append(input.Services, MobileServiceObservation{
			Port: svc.Port, Protocol: svc.Protocol, Name: svc.Name, Product: svc.Product, Source: "device_service", Timestamp: now,
		})
	}
	ips := deviceIPStrings(device)
	idSet := stringSet(device.EvidenceIDs)
	for _, ev := range evidence {
		if !evidenceMatchesDevice(ev, device.ID, ips, idSet) {
			continue
		}
		applyEvidenceToMobileInput(&input, ev)
	}
	return input
}

func mobileInputFromIntelDevice(device models.DeviceIntelDevice, evidence []models.Evidence, now time.Time) MobileFingerprintInput {
	input := MobileFingerprintInput{DeviceID: device.ID, Timestamp: now}
	for _, hostname := range device.Hostnames {
		input.Hostnames = appendObserved(input.Hostnames, hostname, "device_intel_hostname", now)
	}
	input.MACAddresses = append(input.MACAddresses, device.MACAddresses...)
	if device.Vendor.OUIVendor != nil {
		input.OUIVendors = appendUniqueString(input.OUIVendors, *device.Vendor.OUIVendor)
		input.VendorHints = appendObserved(input.VendorHints, *device.Vendor.OUIVendor, "device_intel_oui_vendor", now)
	}
	if device.Vendor.FingerprintVendor != nil {
		input.VendorHints = appendObserved(input.VendorHints, *device.Vendor.FingerprintVendor, "device_intel_fingerprint_vendor", now)
	}
	for _, svc := range device.Services {
		input.Services = append(input.Services, MobileServiceObservation{
			Port: svc.Port, Protocol: svc.Protocol, Name: svc.Name, Product: firstMobileSource(svc.Product, svc.Version), Source: "device_intel_service", Timestamp: now,
		})
	}
	for _, rec := range device.MDNSRecords {
		input.MDNSRecords = append(input.MDNSRecords, MobileMDNSRecord{Name: rec.Name, Service: rec.Service, Target: rec.Target, Text: rec.Text, Source: "device_intel_mdns", Timestamp: now})
	}
	for _, rec := range device.NBNSRecords {
		input.Hostnames = appendObserved(input.Hostnames, rec.Name, "netbios", now)
	}
	for _, rec := range device.LLMNRRecords {
		input.Hostnames = appendObserved(input.Hostnames, rec.Name, "llmnr", now)
	}
	ips := stringSet(device.IPAddresses)
	idSet := stringSet(device.EvidenceIDs)
	for _, ev := range evidence {
		if !evidenceMatchesDevice(ev, device.ID, ips, idSet) {
			continue
		}
		applyEvidenceToMobileInput(&input, ev)
	}
	return input
}

func applyEvidenceToMobileInput(input *MobileFingerprintInput, ev models.Evidence) {
	kind := strings.ToLower(ev.Kind + " " + ev.Source)
	ts := ev.Timestamp
	if ts.IsZero() {
		ts = input.Timestamp
	}
	source := firstMobileSource(ev.Source, ev.Kind)
	if strings.Contains(kind, "dhcp") {
		input.DHCPHostnames = appendObserved(input.DHCPHostnames, firstMapValue(ev.Data, "hostname", "host_name", "name"), source, ts)
	}
	if strings.Contains(kind, "reverse_dns") || strings.Contains(kind, "llmnr") || strings.Contains(kind, "nbns") || strings.Contains(kind, "netbios") {
		input.Hostnames = appendObserved(input.Hostnames, firstMapValue(ev.Data, "hostname", "name", "host"), source, ts)
	}
	if strings.Contains(kind, "mdns") {
		input.MDNSRecords = append(input.MDNSRecords, MobileMDNSRecord{
			Name:      firstMapValue(ev.Data, "name", "instance", "hostname"),
			Service:   firstMapValue(ev.Data, "service", "service_type", "type"),
			Target:    firstMapValue(ev.Data, "target", "host"),
			Text:      splitMobileList(firstMapValue(ev.Data, "txt", "text")),
			Source:    source,
			Timestamp: ts,
		})
	}
	if strings.Contains(kind, "dns") {
		for _, key := range []string{"domain", "query", "qname", "hostname", "name"} {
			input.DNSQueries = appendObserved(input.DNSQueries, ev.Data[key], source, ts)
		}
	}
	if strings.Contains(kind, "http") || strings.Contains(kind, "user_agent") {
		input.UserAgents = appendObserved(input.UserAgents, firstMapValue(ev.Data, "user_agent", "ua"), source, ts)
	}
	if strings.Contains(kind, "tcp") || strings.Contains(kind, "udp") || strings.Contains(kind, "nmap") {
		for _, port := range parseMobilePorts(firstMapValue(ev.Data, "ports", "open_ports", "port")) {
			input.Services = append(input.Services, MobileServiceObservation{
				Port: port, Protocol: firstMobileSource(ev.Data["protocol"], "tcp"), Name: ev.Data["service"], Product: ev.Data["product"], Source: source, Timestamp: ts,
			})
		}
	}
	if strings.Contains(kind, "passive") || strings.Contains(kind, "packet") {
		input.PassivePackets = append(input.PassivePackets, MobilePassivePacket{
			Protocol:        strings.ToLower(ev.Data["protocol"]),
			SourcePort:      parseMobilePort(ev.Data["source_port"]),
			DestinationPort: parseMobilePort(firstMapValue(ev.Data, "destination_port", "dest_port", "port")),
			Value:           firstMapValue(ev.Data, "service", "summary", "metadata"),
			Source:          source,
			Timestamp:       ts,
		})
	}
}

func evidenceMatchesDevice(ev models.Evidence, deviceID string, ips map[string]bool, evidenceIDs map[string]bool) bool {
	if ev.ID != "" && evidenceIDs[ev.ID] {
		return true
	}
	if ev.Data == nil {
		return false
	}
	for _, key := range []string{"ip", "target", "host", "gateway", "source_ip", "destination_ip"} {
		if ips[ev.Data[key]] {
			return true
		}
	}
	return deviceID != "" && (ev.Data["device_id"] == deviceID || ev.Data["target_id"] == deviceID)
}

func deviceIPStrings(device models.Device) map[string]bool {
	out := map[string]bool{}
	for _, address := range device.Addresses {
		if address.IP != "" {
			out[address.IP] = true
		}
	}
	for _, iface := range device.Interfaces {
		for _, ip := range iface.IPs {
			if ip != "" {
				out[ip] = true
			}
		}
	}
	return out
}

func mobileVendorFamily(vendor string) string {
	lower := strings.ToLower(vendor)
	switch {
	case strings.Contains(lower, "apple"):
		return models.MobileOSHintIOS
	case containsAnyMobile(lower, "samsung", "xiaomi", "oppo", "vivo", "google", "oneplus", "huawei", "honor", "realme", "motorola", "moto", "nothing", "sony", "tcl", "tecno", "infinix", "lenovo"):
		return models.MobileOSHintAndroid
	default:
		return ""
	}
}

func mobileOUIVendor(mac string) string {
	compact := strings.ToUpper(strings.NewReplacer(":", "", "-", "", ".", "").Replace(strings.TrimSpace(mac)))
	if len(compact) < 6 {
		if hw, err := net.ParseMAC(mac); err == nil && len(hw) >= 3 {
			compact = strings.ToUpper(fmt.Sprintf("%02X%02X%02X", hw[0], hw[1], hw[2]))
		}
	}
	if len(compact) < 6 {
		return ""
	}
	return mobileOUIPrefixes[compact[:6]]
}

func isLocallyAdministeredMAC(mac string) bool {
	hw, err := net.ParseMAC(mac)
	if err != nil || len(hw) == 0 {
		return false
	}
	return hw[0]&0x02 != 0
}

func redactMACForEvidence(mac string) string {
	hw, err := net.ParseMAC(mac)
	if err != nil || len(hw) < 3 {
		return mac
	}
	return fmt.Sprintf("%02x:%02x:%02x:xx:xx:xx", hw[0], hw[1], hw[2])
}

func dnsImpact(domains []string, apple bool) (int, string) {
	if len(domains) >= 2 {
		return 45, models.MobileEvidenceStrengthMedium
	}
	if !apple && len(domains) == 1 && (domains[0] == "mtalk.google.com" || domains[0] == "android.clients.google.com") {
		return 45, models.MobileEvidenceStrengthMedium
	}
	return 20, models.MobileEvidenceStrengthWeak
}

func normalizeMobileDomain(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = strings.TrimPrefix(value, "http://")
	value = strings.TrimPrefix(value, "https://")
	if slash := strings.IndexByte(value, '/'); slash >= 0 {
		value = value[:slash]
	}
	if colon := strings.IndexByte(value, ':'); colon >= 0 {
		value = value[:colon]
	}
	return strings.Trim(value, ". ")
}

func matchedMobileDomain(domain string, hints []string) string {
	for _, hint := range hints {
		if domain == hint || strings.HasSuffix(domain, "."+hint) {
			return hint
		}
	}
	return ""
}

func parseMobilePorts(raw string) []int {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	replacer := strings.NewReplacer(";", ",", " ", ",", "/", ",")
	parts := strings.Split(replacer.Replace(raw), ",")
	var out []int
	for _, part := range parts {
		if port := parseMobilePort(part); port > 0 {
			out = append(out, port)
		}
	}
	sort.Ints(out)
	return out
}

func parseMobilePort(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	port, _ := strconv.Atoi(raw)
	return port
}

func safeUserAgentValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 160 {
		return value
	}
	return value[:160] + "..."
}

func appendObserved(values []MobileObservedValue, value, source string, ts time.Time) []MobileObservedValue {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing.Value == value && existing.Source == source {
			return values
		}
	}
	return append(values, MobileObservedValue{Value: value, Source: source, Timestamp: ts})
}

func appendUniqueString(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		if value != "" {
			out[value] = true
		}
	}
	return out
}

func sortedMapKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func sortedMobileEvidence(values []models.MobileEvidenceItem) []models.MobileEvidenceItem {
	out := append([]models.MobileEvidenceItem(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Strength != out[j].Strength {
			return mobileStrengthRank(out[i].Strength) < mobileStrengthRank(out[j].Strength)
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func mobileStrengthRank(strength string) int {
	switch strength {
	case models.MobileEvidenceStrengthStrong:
		return 0
	case models.MobileEvidenceStrengthMedium:
		return 1
	default:
		return 2
	}
}

func splitMobileList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	raw = strings.NewReplacer(";", ",", "|", ",").Replace(raw)
	parts := strings.Split(raw, ",")
	var out []string
	for _, part := range parts {
		out = appendUniqueString(out, part)
	}
	return out
}

func firstMapValue(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(values[key]); value != "" {
			return value
		}
	}
	return ""
}

func firstMobileSource(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func containsAnyMobile(value string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(value, strings.ToLower(token)) {
			return true
		}
	}
	return false
}

func mobileContainsStandalone(value, token string) bool {
	value = strings.ToLower(value)
	token = strings.ToLower(token)
	index := strings.Index(value, token)
	for index >= 0 {
		beforeOK := index == 0 || !isMobileAlphaNum(value[index-1])
		after := index + len(token)
		afterOK := after == len(value) || !isMobileAlphaNum(value[after])
		if beforeOK && afterOK {
			return true
		}
		next := index + 1
		if next >= len(value) {
			break
		}
		index = strings.Index(value[next:], token)
		if index >= 0 {
			index += next
		}
	}
	return false
}

func isMobileAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func mobileNameContains(values []string, tokens ...string) bool {
	joined := strings.ToLower(strings.Join(values, " "))
	for _, token := range tokens {
		if strings.Contains(joined, strings.ToLower(token)) {
			return true
		}
	}
	return false
}

func displayMobileOS(os string) string {
	switch os {
	case models.MobileOSHintAndroid:
		return "Android"
	case models.MobileOSHintIPadOS:
		return "iPadOS"
	case models.MobileOSHintIOS:
		return "iOS"
	default:
		return "Mobile OS"
	}
}

func pluralY(count int) string {
	if count == 1 {
		return "y"
	}
	return "ies"
}

func hasWarningMobileConflict(conflicts []models.MobileConflictItem) bool {
	for _, conflict := range conflicts {
		if conflict.Severity == models.MobileConflictSeverityWarning {
			return true
		}
	}
	return false
}

func osRank(os string) int {
	switch os {
	case models.MobileOSHintIOS:
		return 0
	case models.MobileOSHintIPadOS:
		return 1
	case models.MobileOSHintAndroid:
		return 2
	default:
		return 3
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampMobile(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
