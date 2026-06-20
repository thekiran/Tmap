package ippool

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	mobileClassConfirmedIOS     = "confirmed_ios"
	mobileClassProbableIOS      = "probable_ios"
	mobileClassPossibleIOS      = "possible_ios"
	mobileClassConfirmedIPadOS  = "confirmed_ipados"
	mobileClassProbableIPadOS   = "probable_ipados"
	mobileClassPossibleIPadOS   = "possible_ipados"
	mobileClassConfirmedAndroid = "confirmed_android"
	mobileClassProbableAndroid  = "probable_android"
	mobileClassPossibleAndroid  = "possible_android"
	mobileClassConflict         = "conflicting_mobile_os_evidence"
	mobileClassUnknownMobile    = "unknown_mobile"
	mobileClassUnknownDevice    = "unknown_device"

	mobileOSIOS     = "ios"
	mobileOSIPadOS  = "ipados"
	mobileOSAndroid = "android"
	mobileOSUnknown = "unknown"

	mobileTypePhone   = "phone"
	mobileTypeTablet  = "tablet"
	mobileTypeUnknown = "unknown"

	mobileStrengthStrong = "strong"
	mobileStrengthMedium = "medium"
	mobileStrengthWeak   = "weak"
)

var liveMobileOUIPrefixes = map[string]string{
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

var liveAppleDNSHints = []string{
	"apple.com",
	"icloud.com",
	"mzstatic.com",
	"itunes.apple.com",
	"push.apple.com",
	"aaplimg.com",
	"captive.apple.com",
}

var liveAndroidDNSHints = []string{
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

type liveMobileScorer struct {
	now          time.Time
	seq          int
	iosScore     int
	ipadScore    int
	androidScore int
	evidence     []MobileEvidenceItem
	conflicts    []MobileConflictItem
	warnings     []string
	idsByOS      map[string][]string
	strongByOS   map[string]int
	catsByOS     map[string]map[string]bool
	randomMAC    bool
}

func newLiveMobileScorer(now time.Time) *liveMobileScorer {
	if now.IsZero() {
		now = time.Now()
	}
	return &liveMobileScorer{
		now:        now.UTC(),
		idsByOS:    map[string][]string{},
		strongByOS: map[string]int{},
		catsByOS:   map[string]map[string]bool{},
	}
}

func fingerprintLiveMobileDevice(entry DevicePoolEntry, now time.Time) MobileFingerprint {
	s := newLiveMobileScorer(now)
	if entry.Hostname != "" {
		s.scoreHostname(entry.Hostname, "hostname", "device_registry", entry.LastSeen)
	}
	if entry.MAC != "" || entry.Vendor != "" {
		s.scoreMACOrVendor(entry.MAC, entry.Vendor, "device_registry", entry.LastSeen)
	}
	for _, item := range entry.Evidence {
		if isIdentityEvidence(item) {
			continue
		}
		s.scoreEvidence(item)
	}
	return s.finish()
}

func isIdentityEvidence(item EvidenceItem) bool {
	t := strings.ToLower(item.Type + " " + item.Source)
	return strings.Contains(t, "hostname") ||
		strings.Contains(t, "dhcp") ||
		strings.Contains(t, "netbios") ||
		strings.Contains(t, "llmnr") ||
		strings.Contains(t, "mac") ||
		strings.Contains(t, "oui") ||
		strings.Contains(t, "vendor")
}

func sanitizeEvidenceForRegistry(item EvidenceItem) (EvidenceItem, bool) {
	item.Type = strings.TrimSpace(item.Type)
	item.Source = strings.TrimSpace(item.Source)
	item.Value = strings.TrimSpace(item.Value)
	if item.Type == "" {
		item.Type = "metadata"
	}
	if item.Source == "" {
		item.Source = "discovery"
	}
	if item.Strength == "" {
		item.Strength = StrengthWeak
	}
	lowerType := strings.ToLower(item.Type + " " + item.Source)
	if isPassiveDNSMetadata(lowerType) {
		category := privacySafeDNSCategory(item.Value)
		if category == "" {
			return EvidenceItem{}, false
		}
		item.Type = "dns_query"
		item.Value = category
		item.Source = "passive_dns_metadata"
		if item.ConfidenceImpact <= 0 {
			item.ConfidenceImpact = 0.15
		}
		if item.Strength == "" {
			item.Strength = StrengthWeak
		}
	}
	return item, item.Value != "" || strings.Contains(lowerType, "service") || strings.Contains(lowerType, "port")
}

func isPassiveDNSMetadata(value string) bool {
	if strings.Contains(value, "mdns") || strings.Contains(value, "dns-sd") || strings.Contains(value, "bonjour") {
		return false
	}
	return strings.Contains(value, "dns")
}

func privacySafeDNSCategory(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "apple_service_domain_seen", "android_google_service_domain_seen":
		return strings.ToLower(strings.TrimSpace(raw))
	}
	domain := normalizeLiveMobileDomain(raw)
	if domain == "" {
		return ""
	}
	if matchedLiveMobileDomain(domain, liveAppleDNSHints) != "" {
		return "apple_service_domain_seen"
	}
	if matchedLiveMobileDomain(domain, liveAndroidDNSHints) != "" {
		return "android_google_service_domain_seen"
	}
	return ""
}

func meaningfulMobileEvidence(item EvidenceItem) bool {
	t := strings.ToLower(item.Type + " " + item.Source)
	switch {
	case strings.Contains(t, "hostname"),
		strings.Contains(t, "dhcp"),
		strings.Contains(t, "netbios"),
		strings.Contains(t, "llmnr"),
		strings.Contains(t, "arp"),
		strings.Contains(t, "mac"),
		strings.Contains(t, "oui"),
		strings.Contains(t, "vendor"),
		strings.Contains(t, "mdns"),
		strings.Contains(t, "bonjour"),
		strings.Contains(t, "dns-sd"),
		strings.Contains(t, "dns"),
		strings.Contains(t, "user_agent"),
		strings.Contains(t, "http"),
		strings.Contains(t, "service"),
		strings.Contains(t, "port"),
		strings.Contains(t, "passive"):
		return true
	default:
		return false
	}
}

func mergeEvidence(existing []EvidenceItem, item EvidenceItem) ([]EvidenceItem, bool) {
	key := evidenceMergeKey(item)
	for i, old := range existing {
		if evidenceMergeKey(old) != key {
			continue
		}
		if evidenceRank(item.Strength) > evidenceRank(old.Strength) ||
			(evidenceRank(item.Strength) == evidenceRank(old.Strength) && item.ConfidenceImpact > old.ConfidenceImpact) {
			existing[i] = item
			return existing, true
		}
		return existing, false
	}
	return append(existing, item), true
}

func evidenceMergeKey(item EvidenceItem) string {
	return strings.ToLower(strings.TrimSpace(item.Type) + "|" + strings.TrimSpace(item.Source) + "|" + strings.TrimSpace(item.Value))
}

func evidenceRank(strength EvidenceStrength) int {
	switch strength {
	case StrengthConfirmed:
		return 3
	case StrengthInferred:
		return 2
	default:
		return 1
	}
}

func (s *liveMobileScorer) scoreEvidence(item EvidenceItem) {
	t := strings.ToLower(item.Type + " " + item.Source)
	ts := parseEvidenceTimestamp(item.Timestamp, s.now)
	switch {
	case strings.Contains(t, "hostname"), strings.Contains(t, "dhcp"), strings.Contains(t, "netbios"), strings.Contains(t, "llmnr"):
		typ := "hostname"
		if strings.Contains(t, "dhcp") {
			typ = "dhcp_hostname"
		}
		s.scoreHostname(item.Value, typ, item.Source, item.Timestamp)
	case strings.Contains(t, "mac"), strings.Contains(t, "oui"), strings.Contains(t, "vendor"):
		mac := ""
		vendor := item.Value
		if looksLikeMAC(item.Value) {
			mac = item.Value
			vendor = ""
		}
		s.scoreMACOrVendor(mac, vendor, item.Source, item.Timestamp)
	case strings.Contains(t, "mdns"), strings.Contains(t, "bonjour"), strings.Contains(t, "dns-sd"):
		s.scoreMDNS(item.Value, item.Source, ts)
	case strings.Contains(t, "dns"):
		s.scoreDNSCategory(item.Value, item.Source, ts)
	case strings.Contains(t, "user_agent"), strings.Contains(t, "http"):
		s.scoreUserAgent(item.Value, item.Source, ts)
	case strings.Contains(t, "service"), strings.Contains(t, "port"), strings.Contains(t, "tcp"), strings.Contains(t, "udp"):
		s.scoreService(item.Value, item.Source, ts)
	case strings.Contains(t, "passive"), strings.Contains(t, "packet"):
		s.scorePassive(item.Value, item.Source, ts)
	}
}

func (s *liveMobileScorer) scoreHostname(value, evidenceType, source, timestamp string) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return
	}
	lower := strings.ToLower(raw)
	ts := parseEvidenceTimestamp(timestamp, s.now)
	switch {
	case strings.Contains(lower, "iphone"):
		s.addIOS(evidenceType, raw, 45, mobileStrengthStrong, source, ts, "Hostname explicitly contains iPhone.")
	case strings.Contains(lower, "ipad"):
		s.addIPad(evidenceType, raw, 45, mobileStrengthStrong, source, ts, "Hostname explicitly contains iPad.")
	case standaloneToken(lower, "ios"):
		s.addIOS(evidenceType, raw, 30, mobileStrengthMedium, source, ts, "Hostname contains an iOS token.")
	case containsAny(lower, "android", "galaxy", "samsung", "redmi", "xiaomi", "pixel", "oneplus", "oppo", "vivo", "realme", "huawei", "honor", "moto", "motorola", "nothing"):
		s.addAndroid(evidenceType, raw, 45, mobileStrengthStrong, source, ts, "Hostname contains an explicit Android or Android-vendor/model-family token.")
	}
}

func (s *liveMobileScorer) scoreMACOrVendor(mac, vendor, source, timestamp string) {
	ts := parseEvidenceTimestamp(timestamp, s.now)
	local := false
	if mac != "" {
		local = isLocalMAC(mac)
		if local {
			s.randomMAC = true
			s.warn("MAC randomization/private Wi-Fi addresses can reduce mobile OS accuracy.")
		}
		if vendor == "" {
			vendor = liveMobileOUIVendor(mac)
		}
	}
	family := liveMobileVendorFamily(vendor)
	if family == "" {
		return
	}
	impact := 35
	strength := mobileStrengthMedium
	explanation := fmt.Sprintf("MAC/OUI vendor %q is a mobile vendor signal, not standalone proof.", vendor)
	value := vendor
	if mac != "" {
		value = fmt.Sprintf("%s (%s)", redactLiveMAC(mac), vendor)
	}
	if local {
		impact = 8
		strength = mobileStrengthWeak
		explanation = fmt.Sprintf("MAC is locally administered/randomized; vendor hint %q is downgraded to weak support.", vendor)
	}
	if family == mobileOSAndroid {
		s.addAndroid("mac_oui", value, impact, strength, source, ts, explanation)
		return
	}
	s.addApple("mac_oui", value, impact, strength, source, ts, explanation)
}

func (s *liveMobileScorer) scoreMDNS(value, source string, ts time.Time) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return
	}
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "iphone"):
		s.addIOS("mdns", raw, 45, mobileStrengthStrong, source, ts, "mDNS metadata explicitly contains iPhone.")
	case strings.Contains(lower, "ipad"):
		s.addIPad("mdns", raw, 45, mobileStrengthStrong, source, ts, "mDNS metadata explicitly contains iPad.")
	case containsAny(lower, "_apple-mobdev2._tcp", "_companion-link._tcp"):
		s.addApple("mdns", raw, 30, mobileStrengthStrong, source, ts, "Apple-specific Bonjour service metadata was observed.")
	case containsAny(lower, "_airplay._tcp", "_raop._tcp"):
		s.addApple("mdns", raw, 25, mobileStrengthMedium, source, ts, "AirPlay/RAOP Bonjour metadata supports Apple identity but is not Apple-only by itself.")
	case strings.Contains(lower, "android"):
		s.addAndroid("mdns", raw, 30, mobileStrengthStrong, source, ts, "mDNS/DNS-SD metadata explicitly contains Android.")
	case containsAny(lower, "_googlecast._tcp", "chromecast"):
		if s.androidScore > 0 {
			s.addAndroid("mdns", raw, 15, mobileStrengthMedium, source, ts, "Google Cast metadata is treated as Android support only because other Android evidence exists.")
		}
	case strings.Contains(lower, "_services._dns-sd._udp") || strings.Contains(lower, "_mdns"):
		s.warn("mDNS/DNS-SD presence alone is not used as Apple or Android proof.")
	}
}

func (s *liveMobileScorer) scoreDNSCategory(value, source string, ts time.Time) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "apple_service_domain_seen":
		s.addApple("dns_query", "apple_service_domain_seen", 20, mobileStrengthWeak, source, ts, "Only matched Apple OS-related domain hints are stored; DNS evidence alone cannot confirm iOS/iPadOS.")
		s.warn("Passive DNS metadata is reduced to matched OS-related domain hints; raw browsing history and packet payloads are not stored.")
	case "android_google_service_domain_seen":
		s.addAndroid("dns_query", "android_google_service_domain_seen", 20, mobileStrengthWeak, source, ts, "Only matched Android/Google service domain hints are stored; DNS evidence alone cannot confirm Android.")
		s.warn("Passive DNS metadata is reduced to matched OS-related domain hints; raw browsing history and packet payloads are not stored.")
	}
}

func (s *liveMobileScorer) scoreUserAgent(value, source string, ts time.Time) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return
	}
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "iphone"):
		s.addIOS("user_agent", safeLiveUserAgent(raw), 45, mobileStrengthStrong, source, ts, "Clear local HTTP User-Agent explicitly contains iPhone.")
	case strings.Contains(lower, "ipad"):
		s.addIPad("user_agent", safeLiveUserAgent(raw), 45, mobileStrengthStrong, source, ts, "Clear local HTTP User-Agent explicitly contains iPad.")
	case strings.Contains(lower, "android"):
		s.addAndroid("user_agent", safeLiveUserAgent(raw), 45, mobileStrengthStrong, source, ts, "Clear local HTTP User-Agent explicitly contains Android.")
	}
	if raw != "" {
		s.warn("User-Agent evidence is only used when visible in clear local HTTP metadata; HTTPS is not decrypted or intercepted.")
	}
}

func (s *liveMobileScorer) scoreService(value, source string, ts time.Time) {
	proto, port, name := parseServiceValue(value)
	if port == 0 {
		return
	}
	lowerName := strings.ToLower(name)
	display := fmt.Sprintf("%s/%d", proto, port)
	if name != "" {
		display += " " + name
	}
	switch {
	case proto == "tcp" && (port == 5223 || port == 2197):
		s.addApple("service_port", display, 25, mobileStrengthMedium, source, ts, "APNS-related TCP port observed; this is supporting Apple evidence, not standalone proof.")
	case proto == "udp" && port == 5353:
		s.warn("UDP 5353/mDNS alone is not used as iOS or Android proof.")
	case (proto == "tcp" || proto == "udp") && (port == 80 || port == 443):
		s.warn("TCP/UDP 80 and 443 are common on many devices and are not used as mobile OS proof.")
	case proto == "udp" && ((port >= 3478 && port <= 3497) || (port >= 16384 && port <= 16403)):
		if s.iosScore > 0 || s.ipadScore > 0 {
			s.addApple("service_port", display, 10, mobileStrengthWeak, source, ts, "FaceTime/Game Center related UDP range supports Apple only because other Apple evidence exists.")
		}
	case proto == "tcp" && (port == 3689 || port == 5000 || port == 6000 || port == 7000):
		if containsAny(lowerName, "airplay", "raop", "daap") || s.iosScore > 0 || s.ipadScore > 0 {
			s.addApple("service_port", display, 15, mobileStrengthMedium, source, ts, "AirPlay/DAAP service metadata supports Apple identity when paired with service naming or other Apple evidence.")
		}
	case (port == 8008 || port == 8009) && containsAny(lowerName, "chromecast", "googlecast"):
		if s.androidScore > 0 {
			s.addAndroid("service_port", display, 15, mobileStrengthMedium, source, ts, "Google Cast service metadata supports Android only because other Android evidence exists.")
		}
	}
	if containsAny(lowerName, "airplay", "raop", "apple-mobdev2") {
		s.addApple("service_name", name, 15, mobileStrengthMedium, source, ts, "Service name/product contains Apple-specific local service metadata.")
	}
	if containsAny(lowerName, "android", "googlecast") && s.androidScore > 0 {
		s.addAndroid("service_name", name, 15, mobileStrengthMedium, source, ts, "Service name/product contains Android/Google local service metadata and is paired with other Android evidence.")
	}
}

func (s *liveMobileScorer) scorePassive(value, source string, ts time.Time) {
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(lower, "iphone"):
		s.addIOS("passive_packet", value, 30, mobileStrengthMedium, source, ts, "Passive metadata contains iPhone without storing packet payload.")
	case strings.Contains(lower, "ipad"):
		s.addIPad("passive_packet", value, 30, mobileStrengthMedium, source, ts, "Passive metadata contains iPad without storing packet payload.")
	case strings.Contains(lower, "android"):
		s.addAndroid("passive_packet", value, 30, mobileStrengthMedium, source, ts, "Passive metadata contains Android without storing packet payload.")
	case strings.Contains(lower, "5353"):
		s.warn("Passive UDP 5353/mDNS packet metadata alone is not used as mobile OS proof.")
	case strings.Contains(lower, "443") || strings.Contains(lower, "80"):
		s.warn("TCP/UDP 80 and 443 are common on many devices and are not used as mobile OS proof.")
	}
}

func (s *liveMobileScorer) finish() MobileFingerprint {
	s.detectConflicts()
	classification := s.classification()
	confidence := s.confidence(classification)
	evidence := append([]MobileEvidenceItem(nil), s.evidence...)
	sort.SliceStable(evidence, func(i, j int) bool {
		if evidence[i].Strength != evidence[j].Strength {
			return mobileStrengthRank(evidence[i].Strength) < mobileStrengthRank(evidence[j].Strength)
		}
		return evidence[i].ID < evidence[j].ID
	})
	return MobileFingerprint{
		Classification:        classification,
		IOSScore:              s.iosScore,
		AndroidScore:          s.androidScore,
		IPadScore:             s.ipadScore,
		Confidence:            confidence,
		Evidence:              evidence,
		Conflicts:             s.conflicts,
		Warnings:              sortedStrings(s.warnings),
		LastUpdatedAt:         s.now.Format(time.RFC3339),
		WhyThisClassification: s.why(classification),
		WhyNotCertain:         s.whyNotCertain(classification),
	}
}

func (s *liveMobileScorer) detectConflicts() {
	appleScore := maxInt(s.iosScore, s.ipadScore)
	if appleScore >= 35 && s.androidScore >= 35 {
		s.conflicts = append(s.conflicts, MobileConflictItem{
			Reason:             "Apple mobile evidence and Android evidence were both observed.",
			IOSEvidenceIDs:     sortedStrings(append(append([]string{}, s.idsByOS[mobileOSIOS]...), s.idsByOS[mobileOSIPadOS]...)),
			AndroidEvidenceIDs: sortedStrings(s.idsByOS[mobileOSAndroid]),
			Severity:           "warning",
			ResolutionHint:     "Treat as conflicting until a second independent source agrees.",
		})
	}
}

func (s *liveMobileScorer) classification() string {
	if len(s.conflicts) > 0 {
		return mobileClassConflict
	}
	candidates := []struct {
		os    string
		score int
	}{
		{mobileOSIOS, s.iosScore},
		{mobileOSIPadOS, s.ipadScore},
		{mobileOSAndroid, s.androidScore},
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return mobileOSRank(candidates[i].os) < mobileOSRank(candidates[j].os)
	})
	best := candidates[0]
	if best.score == 0 {
		if len(s.warnings) > 0 {
			return mobileClassUnknownDevice
		}
		return mobileClassUnknownDevice
	}
	if s.onlyDNS(best.os) {
		return possibleClass(best.os)
	}
	if s.randomMAC && s.onlyMAC(best.os) && best.score < 20 {
		return mobileClassUnknownMobile
	}
	if best.score >= 80 && s.strongByOS[best.os] >= 2 {
		return confirmedClass(best.os)
	}
	if best.score >= 60 {
		return probableClass(best.os)
	}
	if best.score >= 20 {
		return possibleClass(best.os)
	}
	return mobileClassUnknownMobile
}

func (s *liveMobileScorer) confidence(classification string) float64 {
	switch classification {
	case mobileClassConfirmedIOS, mobileClassConfirmedIPadOS, mobileClassConfirmedAndroid:
		return 0.9
	case mobileClassProbableIOS, mobileClassProbableIPadOS, mobileClassProbableAndroid:
		return 0.72
	case mobileClassPossibleIOS, mobileClassPossibleIPadOS, mobileClassPossibleAndroid:
		if s.randomMAC {
			return 0.34
		}
		return 0.42
	case mobileClassConflict:
		return 0.45
	case mobileClassUnknownMobile:
		return 0.22
	default:
		return 0
	}
}

func (s *liveMobileScorer) why(classification string) string {
	if classification == mobileClassUnknownDevice {
		return ""
	}
	if classification == mobileClassConflict {
		return "Conflicting Apple and Android evidence was observed; the registry keeps the conflict visible instead of choosing a fake winner."
	}
	label := mobileDisplayClassification(classification)
	if len(s.evidence) == 0 {
		return label
	}
	parts := make([]string, 0, minInt(3, len(s.evidence)))
	for _, ev := range s.evidence {
		if ev.Explanation != "" {
			parts = append(parts, strings.TrimSuffix(ev.Explanation, "."))
		} else {
			parts = append(parts, ev.Type+" "+ev.Value)
		}
		if len(parts) == 3 {
			break
		}
	}
	return fmt.Sprintf("%s because %s.", label, strings.Join(parts, ", "))
}

func (s *liveMobileScorer) whyNotCertain(classification string) string {
	switch classification {
	case mobileClassConfirmedIOS, mobileClassConfirmedIPadOS, mobileClassConfirmedAndroid, mobileClassUnknownDevice:
		return ""
	case mobileClassConflict:
		return "The evidence points to more than one mobile OS family, so the result is not certain."
	default:
		return "This classification is evidence-based, not guaranteed. Weak or single-source signals are shown as possible/probable rather than confirmed."
	}
}

func (s *liveMobileScorer) addIOS(evidenceType, value string, impact int, strength, source string, ts time.Time, explanation string) {
	s.iosScore += impact
	s.addEvidence(mobileOSIOS, evidenceType, value, impact, strength, source, ts, explanation)
}

func (s *liveMobileScorer) addIPad(evidenceType, value string, impact int, strength, source string, ts time.Time, explanation string) {
	s.ipadScore += impact
	s.addEvidence(mobileOSIPadOS, evidenceType, value, impact, strength, source, ts, explanation)
}

func (s *liveMobileScorer) addApple(evidenceType, value string, impact int, strength, source string, ts time.Time, explanation string) {
	s.iosScore += impact
	s.ipadScore += impact
	id := s.addEvidence(mobileOSIOS, evidenceType, value, impact, strength, source, ts, explanation)
	s.track(mobileOSIPadOS, evidenceType, strength, id)
}

func (s *liveMobileScorer) addAndroid(evidenceType, value string, impact int, strength, source string, ts time.Time, explanation string) {
	s.androidScore += impact
	s.addEvidence(mobileOSAndroid, evidenceType, value, impact, strength, source, ts, explanation)
}

func (s *liveMobileScorer) addEvidence(osHint, evidenceType, value string, impact int, strength, source string, ts time.Time, explanation string) string {
	if ts.IsZero() {
		ts = s.now
	}
	s.seq++
	id := fmt.Sprintf("mobile-%s-%s-%02d", osHint, strings.ReplaceAll(evidenceType, "_", "-"), s.seq)
	item := MobileEvidenceItem{
		ID:               id,
		Type:             evidenceType,
		Value:            strings.TrimSpace(value),
		OSHint:           osHint,
		ConfidenceImpact: impact,
		Strength:         strength,
		Source:           firstNonEmpty(source, "mobile_fingerprint"),
		Timestamp:        ts.UTC().Format(time.RFC3339),
		Explanation:      explanation,
	}
	s.evidence = append(s.evidence, item)
	s.track(osHint, evidenceType, strength, id)
	return id
}

func (s *liveMobileScorer) track(osHint, evidenceType, strength, id string) {
	if s.catsByOS[osHint] == nil {
		s.catsByOS[osHint] = map[string]bool{}
	}
	s.catsByOS[osHint][evidenceType] = true
	if strength == mobileStrengthStrong {
		s.strongByOS[osHint]++
	}
	s.idsByOS[osHint] = appendUnique(s.idsByOS[osHint], id)
}

func (s *liveMobileScorer) onlyDNS(os string) bool {
	cats := s.catsByOS[os]
	return len(cats) == 1 && cats["dns_query"]
}

func (s *liveMobileScorer) onlyMAC(os string) bool {
	cats := s.catsByOS[os]
	return len(cats) == 1 && cats["mac_oui"]
}

func (s *liveMobileScorer) warn(message string) {
	s.warnings = appendUnique(s.warnings, message)
}

func applyLiveMobileHints(e *DevicePoolEntry, fp MobileFingerprint) {
	e.MobileFingerprint = &fp
	e.OSHint = liveMobileOSHint(fp.Classification)
	e.OSConfidence = fp.Confidence
	e.DeviceTypeHint = liveMobileDeviceTypeHint(fp, []string{e.Hostname})
	e.OSEvidenceSummary = liveMobileEvidenceSummary(fp)
}

func liveMobileOSHint(classification string) string {
	switch classification {
	case mobileClassConfirmedIOS, mobileClassProbableIOS, mobileClassPossibleIOS:
		return mobileOSIOS
	case mobileClassConfirmedIPadOS, mobileClassProbableIPadOS, mobileClassPossibleIPadOS:
		return mobileOSIPadOS
	case mobileClassConfirmedAndroid, mobileClassProbableAndroid, mobileClassPossibleAndroid:
		return mobileOSAndroid
	default:
		return mobileOSUnknown
	}
}

func liveMobileDeviceTypeHint(fp MobileFingerprint, names []string) string {
	switch liveMobileOSHint(fp.Classification) {
	case mobileOSIOS:
		return mobileTypePhone
	case mobileOSIPadOS:
		return mobileTypeTablet
	case mobileOSAndroid:
		joined := strings.ToLower(strings.Join(names, " "))
		if containsAny(joined, "tablet", " tab", "-tab", "pad", "lenovo tab") {
			return mobileTypeTablet
		}
		return mobileTypePhone
	default:
		return mobileTypeUnknown
	}
}

func liveMobileEvidenceSummary(fp MobileFingerprint) []string {
	out := make([]string, 0, 3)
	for _, item := range fp.Evidence {
		text := item.Explanation
		if text == "" {
			text = strings.TrimSpace(item.Type + " " + item.Value)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		out = appendUnique(out, strings.TrimSuffix(text, "."))
		if len(out) == 3 {
			break
		}
	}
	return out
}

func mobileFingerprintChanged(prev *MobileFingerprint, next MobileFingerprint) bool {
	if prev == nil {
		return next.Classification != mobileClassUnknownDevice || len(next.Evidence) > 0 || len(next.Warnings) > 0
	}
	return prev.Classification != next.Classification ||
		prev.IOSScore != next.IOSScore ||
		prev.IPadScore != next.IPadScore ||
		prev.AndroidScore != next.AndroidScore ||
		prev.Confidence != next.Confidence ||
		len(prev.Evidence) != len(next.Evidence) ||
		len(prev.Conflicts) != len(next.Conflicts) ||
		len(prev.Warnings) != len(next.Warnings)
}

func mobileDisplayClassification(classification string) string {
	switch classification {
	case mobileClassConfirmedIOS:
		return "Confirmed iPhone"
	case mobileClassProbableIOS:
		return "Probable iPhone"
	case mobileClassPossibleIOS:
		return "Possible iPhone"
	case mobileClassConfirmedIPadOS:
		return "Confirmed iPad"
	case mobileClassProbableIPadOS:
		return "Probable iPad"
	case mobileClassPossibleIPadOS:
		return "Possible iPad"
	case mobileClassConfirmedAndroid:
		return "Confirmed Android"
	case mobileClassProbableAndroid:
		return "Probable Android"
	case mobileClassPossibleAndroid:
		return "Possible Android"
	case mobileClassConflict:
		return "Conflicting mobile OS evidence"
	case mobileClassUnknownMobile:
		return "Unknown mobile"
	default:
		return "Unknown device"
	}
}

func confirmedClass(os string) string {
	switch os {
	case mobileOSIPadOS:
		return mobileClassConfirmedIPadOS
	case mobileOSAndroid:
		return mobileClassConfirmedAndroid
	default:
		return mobileClassConfirmedIOS
	}
}

func probableClass(os string) string {
	switch os {
	case mobileOSIPadOS:
		return mobileClassProbableIPadOS
	case mobileOSAndroid:
		return mobileClassProbableAndroid
	default:
		return mobileClassProbableIOS
	}
}

func possibleClass(os string) string {
	switch os {
	case mobileOSIPadOS:
		return mobileClassPossibleIPadOS
	case mobileOSAndroid:
		return mobileClassPossibleAndroid
	default:
		return mobileClassPossibleIOS
	}
}

func liveMobileVendorFamily(vendor string) string {
	lower := strings.ToLower(vendor)
	switch {
	case strings.Contains(lower, "apple"):
		return mobileOSIOS
	case containsAny(lower, "samsung", "xiaomi", "oppo", "vivo", "google", "oneplus", "huawei", "honor", "realme", "motorola", "moto", "nothing", "sony", "tcl", "tecno", "infinix", "lenovo"):
		return mobileOSAndroid
	default:
		return ""
	}
}

func liveMobileOUIVendor(mac string) string {
	compact := strings.ToUpper(strings.NewReplacer(":", "", "-", "", ".", "").Replace(strings.TrimSpace(mac)))
	if len(compact) < 6 {
		if hw, err := net.ParseMAC(mac); err == nil && len(hw) >= 3 {
			compact = strings.ToUpper(fmt.Sprintf("%02X%02X%02X", hw[0], hw[1], hw[2]))
		}
	}
	if len(compact) < 6 {
		return ""
	}
	return liveMobileOUIPrefixes[compact[:6]]
}

func isLocalMAC(mac string) bool {
	hw, err := net.ParseMAC(mac)
	return err == nil && len(hw) > 0 && hw[0]&0x02 != 0
}

func looksLikeMAC(value string) bool {
	_, err := net.ParseMAC(strings.TrimSpace(value))
	return err == nil
}

func redactLiveMAC(mac string) string {
	hw, err := net.ParseMAC(mac)
	if err != nil || len(hw) < 3 {
		return mac
	}
	return fmt.Sprintf("%02x:%02x:%02x:xx:xx:xx", hw[0], hw[1], hw[2])
}

func normalizeLiveMobileDomain(raw string) string {
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

func matchedLiveMobileDomain(domain string, hints []string) string {
	for _, hint := range hints {
		if domain == hint || strings.HasSuffix(domain, "."+hint) {
			return hint
		}
	}
	return ""
}

func parseServiceValue(value string) (string, int, string) {
	lower := strings.ToLower(strings.TrimSpace(value))
	proto := "tcp"
	if strings.Contains(lower, "udp") {
		proto = "udp"
	}
	replacer := strings.NewReplacer("/", " ", ":", " ", ",", " ", ";", " ")
	for _, part := range strings.Fields(replacer.Replace(lower)) {
		if port, err := strconv.Atoi(part); err == nil && port > 0 && port <= 65535 {
			return proto, port, value
		}
	}
	return proto, 0, value
}

func parseEvidenceTimestamp(value string, fallback time.Time) time.Time {
	if t, err := time.Parse(time.RFC3339, strings.TrimSpace(value)); err == nil {
		return t.UTC()
	}
	if fallback.IsZero() {
		return time.Now().UTC()
	}
	return fallback.UTC()
}

func safeLiveUserAgent(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 160 {
		return value
	}
	return value[:160] + "..."
}

func containsAny(value string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(value, strings.ToLower(token)) {
			return true
		}
	}
	return false
}

func standaloneToken(value, token string) bool {
	index := strings.Index(value, token)
	for index >= 0 {
		beforeOK := index == 0 || !isAlphaNum(value[index-1])
		after := index + len(token)
		afterOK := after == len(value) || !isAlphaNum(value[after])
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

func isAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func appendUnique(values []string, value string) []string {
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

func sortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	uniq := out[:0]
	for _, value := range out {
		if value == "" {
			continue
		}
		if len(uniq) == 0 || uniq[len(uniq)-1] != value {
			uniq = append(uniq, value)
		}
	}
	return uniq
}

func mobileStrengthRank(strength string) int {
	switch strength {
	case mobileStrengthStrong:
		return 0
	case mobileStrengthMedium:
		return 1
	default:
		return 2
	}
}

func mobileOSRank(os string) int {
	switch os {
	case mobileOSIOS:
		return 0
	case mobileOSIPadOS:
		return 1
	case mobileOSAndroid:
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
