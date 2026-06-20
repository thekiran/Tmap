package ippool

import (
	"strings"
	"sync"
	"time"
)

// DevicePool is the thread-safe registry of tracked devices and their lifecycle
// state. Entries are keyed by IP (de-duplicated) and never auto-deleted — stale
// devices are kept until the user clears them.
type DevicePool struct {
	mu      sync.RWMutex
	entries map[string]*DevicePoolEntry
	// consecutiveFailures tracks failures since the last success, per IP, to
	// drive the active -> recently_seen -> stale transitions.
	consecutiveFail map[string]int
	cfg             Config
	now             func() time.Time
}

func NewDevicePool(cfg Config) *DevicePool {
	return &DevicePool{
		entries:         make(map[string]*DevicePoolEntry),
		consecutiveFail: make(map[string]int),
		cfg:             cfg.normalized(),
		now:             time.Now,
	}
}

func ts(t time.Time) string { return t.UTC().Format(time.RFC3339) }

// EnsureCandidate adds an IP as a candidate if unseen. Returns the entry and
// whether it was newly created. Respects the pool size bound.
func (p *DevicePool) EnsureCandidate(ip string, source Source) (*DevicePoolEntry, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if e, ok := p.entries[ip]; ok {
		return e, false
	}
	if len(p.entries) >= p.cfg.MaxPoolSize {
		return nil, false
	}
	now := ts(p.now())
	e := &DevicePoolEntry{
		ID:        "ip-" + ip,
		IP:        ip,
		FirstSeen: now,
		Status:    StatusCandidate,
		Source:    source,
	}
	p.entries[ip] = e
	return e, true
}

// RecordSuccess marks a probe success: promotes the entry to active, updates the
// running latency/TTL, and resets the failure streak. Returns the entry and
// whether this was the first time the device responded (a "device_found").
func (p *DevicePool) RecordSuccess(ip string, latencyMs float64, ttl int, source Source) (*DevicePoolEntry, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	e := p.entries[ip]
	firstResponse := false
	if e == nil {
		now := ts(p.now())
		e = &DevicePoolEntry{ID: "ip-" + ip, IP: ip, FirstSeen: now, Source: source, Status: StatusCandidate}
		p.entries[ip] = e
	}
	now := p.now()
	if e.ResponseCount == 0 {
		firstResponse = true
	}
	e.ResponseCount++
	e.LastSeen = ts(now)
	e.LastProbeAt = ts(now)
	e.Status = StatusActive
	p.consecutiveFail[ip] = 0
	if latencyMs > 0 {
		if e.AvgLatencyMs == nil {
			v := latencyMs
			e.AvgLatencyMs = &v
		} else {
			// Running average across all responses.
			v := (*e.AvgLatencyMs*float64(e.ResponseCount-1) + latencyMs) / float64(e.ResponseCount)
			e.AvgLatencyMs = &v
		}
	}
	if ttl > 0 {
		t := ttl
		e.TTL = &t
	}
	return e, firstResponse
}

// RecordFailure marks a probe failure and advances the state machine:
//
//	active        -> recently_seen   (after RecentlyAfter consecutive failures)
//	recently_seen -> stale           (after StaleAfter consecutive failures)
//	candidate     -> unreachable     (after CandidateFailMax failures; never on the map)
//	stale stays stale; unreachable stays unreachable.
//
// Returns the entry and whether it just transitioned INTO the stale state.
func (p *DevicePool) RecordFailure(ip string) (*DevicePoolEntry, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	e := p.entries[ip]
	if e == nil {
		return nil, false
	}
	now := p.now()
	e.FailureCount++
	e.LastProbeAt = ts(now)
	p.consecutiveFail[ip]++
	fails := p.consecutiveFail[ip]
	becameStale := false

	switch e.Status {
	case StatusActive:
		if fails >= p.cfg.RecentlyAfter {
			e.Status = StatusRecentlySeen
		}
	case StatusRecentlySeen:
		if fails >= p.cfg.StaleAfter {
			e.Status = StatusStale
			becameStale = true
		}
	case StatusCandidate:
		if fails >= p.cfg.CandidateFailMax {
			e.Status = StatusUnreachable
		}
	}
	return e, becameStale
}

// Get returns a copy of one entry.
func (p *DevicePool) Get(ip string) (DevicePoolEntry, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	e, ok := p.entries[ip]
	if !ok {
		return DevicePoolEntry{}, false
	}
	return *e, true
}

// AddEvidence attaches an evidence item to an entry. It returns a copy of the
// updated entry, whether the stored evidence changed, and whether the mobile
// fingerprint materially changed. Stronger evidence replaces weaker duplicate
// evidence; weaker duplicates never overwrite stronger observations.
func (p *DevicePool) AddEvidence(ip string, item EvidenceItem) (DevicePoolEntry, bool, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	e := p.entries[ip]
	if e == nil {
		return DevicePoolEntry{}, false, false
	}
	if item.Timestamp == "" {
		item.Timestamp = ts(p.now())
	}
	var ok bool
	item, ok = sanitizeEvidenceForRegistry(item)
	if !ok {
		return *e, false, false
	}
	var changed bool
	e.Evidence, changed = mergeEvidence(e.Evidence, item)
	if changed {
		applyEvidenceIdentity(e, item)
	}
	mobileChanged := false
	if changed && meaningfulMobileEvidence(item) {
		mobileChanged = p.rescoreMobileLocked(e)
	}
	return *e, changed, mobileChanged
}

// RefreshMobileFingerprints runs a low-frequency re-score across the registry.
// It emits no changes when the same evidence produces the same classification.
func (p *DevicePool) RefreshMobileFingerprints() []DevicePoolEntry {
	p.mu.Lock()
	defer p.mu.Unlock()
	var changed []DevicePoolEntry
	for _, e := range p.entries {
		if p.rescoreMobileLocked(e) {
			changed = append(changed, *e)
		}
	}
	sortByIP(changed)
	return changed
}

func (p *DevicePool) rescoreMobileLocked(e *DevicePoolEntry) bool {
	if e == nil {
		return false
	}
	fp := fingerprintLiveMobileDevice(*e, p.now())
	if !mobileFingerprintChanged(e.MobileFingerprint, fp) {
		return false
	}
	applyLiveMobileHints(e, fp)
	return true
}

func applyEvidenceIdentity(e *DevicePoolEntry, item EvidenceItem) {
	lower := strings.ToLower(item.Type + " " + item.Source)
	switch {
	case strings.Contains(lower, "hostname"), strings.Contains(lower, "dhcp"), strings.Contains(lower, "netbios"), strings.Contains(lower, "llmnr"):
		if e.Hostname == "" {
			e.Hostname = item.Value
		}
	case strings.Contains(lower, "mac") || strings.Contains(lower, "arp"):
		if e.MAC == "" && looksLikeMAC(item.Value) {
			e.MAC = item.Value
		}
	case strings.Contains(lower, "oui"), strings.Contains(lower, "vendor"):
		if e.Vendor == "" && !looksLikeMAC(item.Value) {
			e.Vendor = item.Value
		}
	}
}

// Snapshot returns a copy of every entry, ordered by IP.
func (p *DevicePool) Snapshot() []DevicePoolEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]DevicePoolEntry, 0, len(p.entries))
	for _, e := range p.entries {
		out = append(out, *e)
	}
	sortByIP(out)
	return out
}

// IPsByStatus returns the IPs currently in any of the given statuses.
func (p *DevicePool) IPsByStatus(statuses ...DeviceStatus) []string {
	want := make(map[DeviceStatus]bool, len(statuses))
	for _, s := range statuses {
		want[s] = true
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	var out []string
	for ip, e := range p.entries {
		if want[e.Status] {
			out = append(out, ip)
		}
	}
	return out
}

// Counts returns the per-status tally for the status snapshot.
func (p *DevicePool) Counts() (active, recently, stale, candidate int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, e := range p.entries {
		switch e.Status {
		case StatusActive:
			active++
		case StatusRecentlySeen:
			recently++
		case StatusStale:
			stale++
		case StatusCandidate:
			candidate++
		}
	}
	return
}

// ClearStale removes stale and unreachable entries (manual, user-initiated).
// Active/recently_seen/candidate devices are kept. Returns the number removed.
func (p *DevicePool) ClearStale() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	removed := 0
	for ip, e := range p.entries {
		if e.Status == StatusStale || e.Status == StatusUnreachable {
			delete(p.entries, ip)
			delete(p.consecutiveFail, ip)
			removed++
		}
	}
	return removed
}

func sortByIP(entries []DevicePoolEntry) {
	// Simple insertion-free sort by dotted-quad numeric value; small N.
	for i := 1; i < len(entries); i++ {
		j := i
		for j > 0 && ipLess(entries[j].IP, entries[j-1].IP) {
			entries[j], entries[j-1] = entries[j-1], entries[j]
			j--
		}
	}
}
