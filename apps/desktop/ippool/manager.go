package ippool

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"
)

// taskQueueSize bounds the in-flight probe channel so memory stays bounded even
// if generation outruns probing.
const taskQueueSize = 512

// TopologyPayload is the batched live-topology update sent to the frontend. It
// contains only devices that have actually responded (active / recently_seen /
// stale) — never raw candidates or unreachable IPs, so the map stays clean.
type TopologyPayload struct {
	Devices   []DevicePoolEntry `json:"devices"`
	Timestamp int64             `json:"timestamp"`
}

// Manager is the IPDiscoveryPoolManager: it owns the pool, scheduler, candidate
// backlog and emits batched Wails events. One instance per app; Start/Stop are
// safe to call repeatedly.
type Manager struct {
	cfg     Config
	guard   ScopeGuard
	gen     CandidateGenerator
	probe   ReachabilityProbe
	sched   *Scheduler
	emitter Emitter

	mu       sync.Mutex
	running  bool
	paused   bool
	cancel   context.CancelFunc
	seeds    []string
	subnets  map[string]bool
	expanded map[string]bool // /24s already turned into candidates
	warnings []string

	pool *DevicePool

	taskCh  chan string
	results chan ProbeOutcome

	inflightMu sync.Mutex
	inflight   map[string]bool

	backlogMu sync.Mutex
	backlog   []string

	// batching / metrics
	batchMu      sync.Mutex
	pendingProbe []ProbeResultPayload
	lastProbeAt  int64
	probeWindow  int
	probesPerSec float64

	lastMobileRefresh time.Time
}

// New constructs a Manager. `hide` hides the ping console window on Windows
// (pass the app's hideConsole); nil is fine in tests.
func New(cfg Config, emitter Emitter, hide HideConsoleFunc) *Manager {
	cfg = cfg.normalized()
	guard := NewScopeGuard(cfg)
	return &Manager{
		cfg:      cfg,
		guard:    guard,
		gen:      NewCandidateGenerator(guard),
		probe:    NewReachabilityProbe(cfg, hide),
		sched:    NewScheduler(cfg),
		emitter:  emitter,
		pool:     NewDevicePool(cfg),
		subnets:  map[string]bool{},
		expanded: map[string]bool{},
		inflight: map[string]bool{},
	}
}

// Running reports whether discovery is active.
func (m *Manager) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// SetPaused pauses/resumes candidate expansion (the active/stale retests keep
// running; only new candidate probing stops).
func (m *Manager) SetPaused(p bool) {
	m.mu.Lock()
	m.paused = p
	m.mu.Unlock()
}

// Start begins continuous discovery from the given seeds. confirmLargeScope must
// be true to expand subnets larger than the guard's automatic limit.
func (m *Manager) Start(seeds []string, confirmLargeScope bool) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return errors.New("discovery already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.running = true
	m.paused = false
	m.warnings = nil
	m.seeds = nil
	m.taskCh = make(chan string, taskQueueSize)
	m.results = make(chan ProbeOutcome, m.cfg.MaxConcurrency*2)
	m.lastMobileRefresh = time.Now()
	m.mu.Unlock()

	m.emit(EvtStarted, m.Snapshot())

	// Seed the pool (validated through the ScopeGuard).
	for _, s := range seeds {
		m.addSeedLocked(s, SourceSeed, confirmLargeScope)
	}

	go m.sched.Run(ctx, m.taskCh, m.probe.Probe, m.results)
	go func() {
		m.consume(ctx)
		close(m.results) // safe: scheduler has returned by the time ctx is done
	}()
	go m.feedCandidates(ctx)
	go m.retestLoop(ctx, m.cfg.ActiveInterval, StatusActive, StatusRecentlySeen)
	go m.retestLoop(ctx, m.cfg.StaleInterval, StatusStale)
	go m.batchLoop(ctx)
	return nil
}

// Stop ends discovery and releases all goroutines.
func (m *Manager) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	cancel := m.cancel
	m.cancel = nil
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	m.emit(EvtStopped, m.Snapshot())
}

// AddSeed adds a new seed IP at runtime.
func (m *Manager) AddSeed(ip string) error {
	if ok, reason := m.guard.AllowIP(ip); !ok {
		m.warn("rejected seed " + ip + ": " + reason)
		return errors.New(reason)
	}
	m.addSeedLocked(ip, SourceSeed, false)
	return nil
}

func (m *Manager) addSeedLocked(ip string, source Source, confirm bool) {
	if ok, reason := m.guard.AllowIP(ip); !ok {
		m.warn("rejected target " + ip + ": " + reason)
		return
	}
	m.mu.Lock()
	if !slices.Contains(m.seeds, ip) {
		m.seeds = append(m.seeds, ip)
	}
	m.mu.Unlock()

	m.pool.EnsureCandidate(ip, source)
	m.pool.AddEvidence(ip, EvidenceItem{Type: "gateway_seed", Source: string(source), Value: ip, Timestamp: ts(time.Now()), ConfidenceImpact: 0.2, Strength: StrengthInferred})
	m.emit(EvtSeedAdded, map[string]any{"ip": ip, "source": source})
	m.enqueue(ip)
	m.expandSubnet(ip, confirm)
}

// expandSubnet turns the /24 around ip into prioritized candidates (once).
func (m *Manager) expandSubnet(ip string, confirm bool) {
	if m.isPaused() {
		return
	}
	cidr := SubnetOf(ip, m.cfg.MaxAutoPrefix)
	if cidr == "" {
		return
	}
	m.mu.Lock()
	if m.expanded[cidr] {
		m.mu.Unlock()
		return
	}
	dec := m.guard.AllowCIDR(cidr, confirm)
	if dec.Warning != "" {
		m.warnings = append(m.warnings, dec.Warning)
	}
	if !dec.Allowed {
		m.warnings = append(m.warnings, dec.Reason)
		m.mu.Unlock()
		m.emit(EvtWarning, map[string]any{"message": dec.Reason})
		return
	}
	m.expanded[cidr] = true
	m.subnets[cidr] = true
	m.mu.Unlock()
	if dec.Warning != "" {
		m.emit(EvtWarning, map[string]any{"message": dec.Warning})
	}

	candidates := m.gen.GenerateFromSeed(ip, cidr, 0)
	m.backlogMu.Lock()
	m.backlog = append(m.backlog, candidates...)
	m.backlogMu.Unlock()
}

// feedCandidates drains the backlog into the probe queue, respecting pause and
// the scheduler's own rate limit (we just don't outrun the bounded taskCh).
func (m *Manager) feedCandidates(ctx context.Context) {
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if m.isPaused() {
				continue
			}
			m.backlogMu.Lock()
			if len(m.backlog) == 0 {
				m.backlogMu.Unlock()
				continue
			}
			ip := m.backlog[0]
			m.backlog = m.backlog[1:]
			m.backlogMu.Unlock()

			if _, created := m.pool.EnsureCandidate(ip, SourceCandidate); created {
				m.emit(EvtCandidateAdded, map[string]any{"ip": ip})
			}
			m.enqueue(ip)
		}
	}
}

// retestLoop periodically re-enqueues every IP in the given statuses.
func (m *Manager) retestLoop(ctx context.Context, interval time.Duration, statuses ...DeviceStatus) {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			for _, ip := range m.pool.IPsByStatus(statuses...) {
				m.enqueue(ip)
			}
		}
	}
}

// enqueue submits an IP for probing unless it is already in flight or the queue
// is full (bounded — a dropped probe is simply retried on the next tick).
func (m *Manager) enqueue(ip string) {
	m.inflightMu.Lock()
	if m.inflight[ip] {
		m.inflightMu.Unlock()
		return
	}
	m.inflight[ip] = true
	m.inflightMu.Unlock()

	select {
	case m.taskCh <- ip:
	default:
		// Queue full: release the in-flight mark so it can be retried later.
		m.inflightMu.Lock()
		delete(m.inflight, ip)
		m.inflightMu.Unlock()
	}
}

// consume processes probe outcomes: updates the pool state machine, emits
// found/stale events, generates candidates from new devices, and accumulates
// batched probe metrics.
func (m *Manager) consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case out, ok := <-m.results:
			if !ok {
				return
			}
			m.inflightMu.Lock()
			delete(m.inflight, out.IP)
			m.inflightMu.Unlock()

			now := time.Now().UnixMilli()
			m.batchMu.Lock()
			m.lastProbeAt = now
			m.probeWindow++
			if len(m.pendingProbe) < 200 {
				m.pendingProbe = append(m.pendingProbe, ProbeResultPayload{
					IP: out.IP, Reachable: out.Reachable, LatencyMs: out.LatencyMs,
					TTL: out.TTL, Method: out.Method, Timestamp: now,
				})
			}
			m.batchMu.Unlock()

			if out.Reachable {
				entry, first := m.pool.RecordSuccess(out.IP, out.LatencyMs, out.TTL, SourceCandidate)
				updated, _, mobileChanged := m.pool.AddEvidence(out.IP, EvidenceItem{Type: "ping_response", Source: out.Method, Value: out.IP, Timestamp: ts(time.Now()), ConfidenceImpact: 0.4, Strength: StrengthConfirmed})
				if updated.ID != "" {
					entry = &updated
				}
				if first {
					m.emit(EvtDeviceFound, *entry)
					// A newly found device expands its own /24 once.
					m.expandSubnet(out.IP, false)
				}
				if mobileChanged {
					m.emitMobileFingerprintUpdated(*entry)
				}
				m.emitDeviceUpdated(*entry)
			} else {
				entry, becameStale := m.pool.RecordFailure(out.IP)
				if entry != nil {
					m.emitDeviceUpdated(*entry)
				}
				if becameStale && entry != nil {
					m.emit(EvtDeviceStale, *entry)
				}
			}
		}
	}
}

// batchLoop flushes coalesced probe results, the topology payload and the status
// snapshot on a fixed cadence so the frontend never gets one event per ping.
func (m *Manager) batchLoop(ctx context.Context) {
	interval := m.cfg.BatchInterval
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			m.flush(interval)
			return
		case <-tick.C:
			m.flush(interval)
		}
	}
}

func (m *Manager) flush(interval time.Duration) {
	m.batchMu.Lock()
	probes := m.pendingProbe
	m.pendingProbe = nil
	window := m.probeWindow
	m.probeWindow = 0
	m.probesPerSec = float64(window) / interval.Seconds()
	m.batchMu.Unlock()

	if len(probes) > 0 {
		m.emit(EvtProbeResult, map[string]any{"results": probes})
	}
	m.refreshMobileFingerprintsIfDue()
	m.emitTopologyUpdated(TopologyPayload{Devices: m.topologyDevices(), Timestamp: time.Now().UnixMilli()})
	m.emit(EvtStatus, m.Snapshot())
}

func (m *Manager) refreshMobileFingerprintsIfDue() {
	m.batchMu.Lock()
	due := m.lastMobileRefresh.IsZero() || time.Since(m.lastMobileRefresh) >= m.cfg.MobileRefreshInterval
	if due {
		m.lastMobileRefresh = time.Now()
	}
	m.batchMu.Unlock()
	if !due {
		return
	}
	for _, entry := range m.pool.RefreshMobileFingerprints() {
		m.emitMobileFingerprintUpdated(entry)
		m.emitDeviceUpdated(entry)
	}
}

// topologyDevices returns only responded devices (never raw candidates).
func (m *Manager) topologyDevices() []DevicePoolEntry {
	all := m.pool.Snapshot()
	out := make([]DevicePoolEntry, 0, len(all))
	for _, e := range all {
		switch e.Status {
		case StatusActive, StatusRecentlySeen, StatusStale:
			out = append(out, e)
		}
	}
	return out
}

// ClearStale drops stale/unreachable devices (user-initiated).
func (m *Manager) ClearStale() int {
	n := m.pool.ClearStale()
	if n > 0 {
		m.emitTopologyUpdated(TopologyPayload{Devices: m.topologyDevices(), Timestamp: time.Now().UnixMilli()})
		m.emit(EvtStatus, m.Snapshot())
	}
	return n
}

// Devices returns the full pool snapshot (for the details panel / list).
func (m *Manager) Devices() []DevicePoolEntry { return m.pool.Snapshot() }

// AddDeviceEvidence attaches externally observed metadata (ARP, DHCP hostname,
// mDNS/DNS-SD, passive packet metadata, DNS OS hints, visible HTTP metadata,
// OUI/vendor, service observations) to a DevicePool entry and emits live update
// events if the record or mobile fingerprint changed.
func (m *Manager) AddDeviceEvidence(ip string, item EvidenceItem) bool {
	m.pool.EnsureCandidate(ip, SourcePassive)
	entry, changed, mobileChanged := m.pool.AddEvidence(ip, item)
	if !changed {
		return false
	}
	if mobileChanged {
		m.emitMobileFingerprintUpdated(entry)
	}
	m.emitDeviceUpdated(entry)
	m.emitTopologyUpdated(TopologyPayload{Devices: m.topologyDevices(), Timestamp: time.Now().UnixMilli()})
	return true
}

// Snapshot returns the live status summary for the frontend panel.
func (m *Manager) Snapshot() StatusSnapshot {
	active, recently, stale, candidate := m.pool.Counts()
	m.mu.Lock()
	seeds := append([]string(nil), m.seeds...)
	subnets := make([]string, 0, len(m.subnets))
	for s := range m.subnets {
		subnets = append(subnets, s)
	}
	warnings := append([]string(nil), m.warnings...)
	running := m.running
	m.mu.Unlock()

	m.backlogMu.Lock()
	queue := len(m.backlog)
	m.backlogMu.Unlock()

	m.batchMu.Lock()
	lastProbe := m.lastProbeAt
	pps := m.probesPerSec
	m.batchMu.Unlock()

	return StatusSnapshot{
		Running: running, Seeds: seeds,
		ActiveCount: active, RecentlyCount: recently, StaleCount: stale, CandidateCount: candidate,
		CandidateQueue: queue, Subnets: subnets, LastProbeAt: lastProbe,
		ProbesPerSec: pps, Concurrency: m.cfg.MaxConcurrency, Warnings: warnings,
	}
}

func (m *Manager) isPaused() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.paused
}

func (m *Manager) warn(msg string) {
	m.mu.Lock()
	m.warnings = append(m.warnings, msg)
	m.mu.Unlock()
	m.emit(EvtWarning, map[string]any{"message": msg})
}

func (m *Manager) emit(event string, data any) {
	if m.emitter != nil {
		m.emitter.Emit(event, data)
	}
}

func (m *Manager) emitDeviceUpdated(entry DevicePoolEntry) {
	m.emit(EvtDeviceUpdated, entry)
	m.emit(EvtDiscoveryDeviceUpdated, entry)
}

func (m *Manager) emitTopologyUpdated(payload TopologyPayload) {
	m.emit(EvtTopologyUpdated, payload)
	m.emit(EvtDiscoveryTopologyUpdated, payload)
}

func (m *Manager) emitMobileFingerprintUpdated(entry DevicePoolEntry) {
	if entry.MobileFingerprint == nil {
		return
	}
	updatedAt := entry.MobileFingerprint.LastUpdatedAt
	if updatedAt == "" {
		updatedAt = ts(time.Now())
	}
	m.emit(EvtDeviceMobileFingerprintUpdated, MobileFingerprintUpdatedPayload{
		DeviceID:          entry.ID,
		IPAddresses:       []string{entry.IP},
		Hostname:          entry.Hostname,
		MobileFingerprint: *entry.MobileFingerprint,
		UpdatedAt:         updatedAt,
	})
}
