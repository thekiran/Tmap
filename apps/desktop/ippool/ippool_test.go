package ippool

import (
	"context"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- ScopeGuard ---

func TestScopeGuard_RefusesPublicAllowsPrivate(t *testing.T) {
	g := NewScopeGuard(DefaultConfig())
	private := []string{"192.168.31.2", "10.0.0.1", "172.16.5.5", "172.31.255.254"}
	for _, ip := range private {
		if ok, reason := g.AllowIP(ip); !ok {
			t.Fatalf("private %s rejected: %s", ip, reason)
		}
	}
	public := []string{"8.8.8.8", "1.1.1.1", "203.0.113.5", "172.32.0.1"}
	for _, ip := range public {
		if ok, _ := g.AllowIP(ip); ok {
			t.Fatalf("public %s was allowed", ip)
		}
	}
	if ok, _ := g.AllowIP("127.0.0.1"); ok {
		t.Fatal("loopback allowed")
	}
	if ok, _ := g.AllowIP("169.254.1.1"); ok {
		t.Fatal("link-local allowed while disabled")
	}
}

func TestScopeGuard_RefusesHugeScopeWithoutConfirmation(t *testing.T) {
	g := NewScopeGuard(DefaultConfig())

	if d := g.AllowCIDR("192.168.31.0/24", false); !d.Allowed {
		t.Fatalf("/24 should be allowed: %+v", d)
	}
	if d := g.AllowCIDR("10.0.0.0/16", false); d.Allowed {
		t.Fatalf("/16 must require confirmation: %+v", d)
	}
	if d := g.AllowCIDR("10.0.0.0/16", true); !d.Allowed || d.Warning == "" {
		t.Fatalf("/16 with confirmation should be allowed with a warning: %+v", d)
	}
	if d := g.AllowCIDR("192.168.0.0/23", false); !d.Allowed || d.Warning == "" {
		t.Fatalf("/23 should be allowed but warned: %+v", d)
	}
	if d := g.AllowCIDR("8.8.0.0/24", false); d.Allowed {
		t.Fatalf("public /24 must be rejected: %+v", d)
	}
}

// --- CandidateGenerator ---

func TestCandidateGenerator_FromSeedSafeAndPrioritized(t *testing.T) {
	g := NewScopeGuard(DefaultConfig())
	gen := NewCandidateGenerator(g)
	cands := gen.GenerateFromSeed("192.168.31.2", "192.168.31.0/24", 0)

	if len(cands) != 254 {
		t.Fatalf("a /24 has 254 hosts, got %d", len(cands))
	}
	// All inside the authorized subnet, never network/broadcast/public.
	for _, c := range cands {
		if c == "192.168.31.0" || c == "192.168.31.255" {
			t.Fatalf("network/broadcast leaked: %s", c)
		}
		if len(c) < 11 || c[:11] != "192.168.31." {
			t.Fatalf("candidate outside subnet: %s", c)
		}
	}
	idx := func(ip string) int { return slices.Index(cands, ip) }
	// Infrastructure-likely IPs come before a random mid-range host.
	for _, infra := range []string{"192.168.31.1", "192.168.31.254", "192.168.31.2"} {
		if i := idx(infra); i < 0 || i >= 10 {
			t.Fatalf("%s should be prioritized near the front, idx=%d", infra, i)
		}
	}
	// DHCP-typical range is present and ahead of the tail sweep.
	if idx("192.168.31.100") < 0 {
		t.Fatal(".100 (DHCP) missing")
	}
	if idx("192.168.31.100") >= idx("192.168.31.50") {
		t.Fatal("DHCP range should be prioritized before the ascending tail")
	}
}

// --- DevicePool state machine ---

func TestDevicePool_Transitions_ActiveStaleActive(t *testing.T) {
	p := NewDevicePool(DefaultConfig())

	e, created := p.EnsureCandidate("10.0.0.5", SourceSeed)
	if !created || e.Status != StatusCandidate {
		t.Fatalf("expected new candidate, got created=%v status=%s", created, e.Status)
	}
	// Duplicate suppression.
	if _, created2 := p.EnsureCandidate("10.0.0.5", SourceSeed); created2 {
		t.Fatal("duplicate IP was not suppressed")
	}

	e, first := p.RecordSuccess("10.0.0.5", 4.0, 64, SourceCandidate)
	if !first || e.Status != StatusActive || e.ResponseCount != 1 {
		t.Fatalf("candidate->active failed: %+v first=%v", e, first)
	}

	// active -> recently_seen (RecentlyAfter=1)
	e, _ = p.RecordFailure("10.0.0.5")
	if e.Status != StatusRecentlySeen {
		t.Fatalf("active->recently_seen failed: %s", e.Status)
	}
	// recently_seen holds until StaleAfter consecutive failures (=3)
	p.RecordFailure("10.0.0.5")
	e, becameStale := p.RecordFailure("10.0.0.5")
	if e.Status != StatusStale || !becameStale {
		t.Fatalf("recently_seen->stale failed: %s stale=%v", e.Status, becameStale)
	}
	// stale -> active on a fresh response
	e, _ = p.RecordSuccess("10.0.0.5", 5.0, 64, SourceCandidate)
	if e.Status != StatusActive {
		t.Fatalf("stale->active failed: %s", e.Status)
	}

	// Stale is never auto-deleted; ClearStale is manual.
	p2 := NewDevicePool(DefaultConfig())
	p2.RecordSuccess("10.0.0.9", 1, 64, SourceCandidate)
	p2.RecordFailure("10.0.0.9")
	p2.RecordFailure("10.0.0.9")
	p2.RecordFailure("10.0.0.9")
	if _, ok := p2.Get("10.0.0.9"); !ok {
		t.Fatal("stale device was auto-deleted")
	}
	if n := p2.ClearStale(); n != 1 {
		t.Fatalf("ClearStale removed %d, want 1", n)
	}
	if _, ok := p2.Get("10.0.0.9"); ok {
		t.Fatal("ClearStale did not remove the stale device")
	}
}

func TestDevicePool_CandidateBecomesUnreachable(t *testing.T) {
	p := NewDevicePool(DefaultConfig())
	p.EnsureCandidate("10.0.0.7", SourceCandidate)
	p.RecordFailure("10.0.0.7")
	e, _ := p.RecordFailure("10.0.0.7") // CandidateFailMax=2
	if e.Status != StatusUnreachable {
		t.Fatalf("candidate->unreachable failed: %s", e.Status)
	}
}

func TestDevicePool_MobileFingerprintHostnameEvidence(t *testing.T) {
	p := NewDevicePool(DefaultConfig())
	p.EnsureCandidate("10.0.0.20", SourcePassive)

	e, changed, mobileChanged := p.AddEvidence("10.0.0.20", EvidenceItem{
		Type: "dhcp_hostname", Source: "dhcp", Value: "KIRAN-iPhone", Strength: StrengthConfirmed,
	})
	if !changed || !mobileChanged || e.MobileFingerprint == nil {
		t.Fatalf("hostname evidence did not update mobile fingerprint: changed=%v mobile=%v entry=%+v", changed, mobileChanged, e)
	}
	if e.MobileFingerprint.Classification != mobileClassPossibleIOS {
		t.Fatalf("classification = %s, want %s", e.MobileFingerprint.Classification, mobileClassPossibleIOS)
	}
	if e.DeviceTypeHint != mobileTypePhone || e.OSHint != mobileOSIOS {
		t.Fatalf("hints = type %q os %q, want phone/ios", e.DeviceTypeHint, e.OSHint)
	}
	if e.Hostname != "KIRAN-iPhone" {
		t.Fatalf("hostname not merged into entry: %q", e.Hostname)
	}
}

func TestDevicePool_MobileFingerprintMACOUIEvidence(t *testing.T) {
	p := NewDevicePool(DefaultConfig())
	p.EnsureCandidate("10.0.0.21", SourcePassive)

	e, _, mobileChanged := p.AddEvidence("10.0.0.21", EvidenceItem{
		Type: "mac_oui", Source: "arp_table", Value: "a4:83:e7:12:34:56", Strength: StrengthInferred,
	})
	if !mobileChanged || e.MobileFingerprint == nil {
		t.Fatalf("MAC/OUI evidence did not update fingerprint: %+v", e)
	}
	if e.MobileFingerprint.IOSScore != 35 || e.MobileFingerprint.IPadScore != 35 {
		t.Fatalf("Apple OUI scores = ios %d ipad %d, want 35/35", e.MobileFingerprint.IOSScore, e.MobileFingerprint.IPadScore)
	}
	if strings.Contains(e.MobileFingerprint.Classification, "probable") || strings.Contains(e.MobileFingerprint.Classification, "confirmed") {
		t.Fatalf("OUI alone classified too strongly: %s", e.MobileFingerprint.Classification)
	}
}

func TestDevicePool_MDNSAppleImprovesIOSScore(t *testing.T) {
	p := NewDevicePool(DefaultConfig())
	p.EnsureCandidate("10.0.0.22", SourcePassive)
	e, _, _ := p.AddEvidence("10.0.0.22", EvidenceItem{Type: "hostname", Source: "llmnr", Value: "KIRAN-iPhone", Strength: StrengthConfirmed})
	before := e.MobileFingerprint.IOSScore
	e, _, _ = p.AddEvidence("10.0.0.22", EvidenceItem{Type: "mdns", Source: "bonjour", Value: "_apple-mobdev2._tcp.local", Strength: StrengthInferred})

	if e.MobileFingerprint.IOSScore <= before {
		t.Fatalf("mDNS did not improve iOS score: before=%d after=%d", before, e.MobileFingerprint.IOSScore)
	}
	if e.MobileFingerprint.Classification != mobileClassProbableIOS {
		t.Fatalf("classification = %s, want %s", e.MobileFingerprint.Classification, mobileClassProbableIOS)
	}
}

func TestDevicePool_AndroidHostnameEvidence(t *testing.T) {
	p := NewDevicePool(DefaultConfig())
	p.EnsureCandidate("10.0.0.23", SourcePassive)
	e, _, _ := p.AddEvidence("10.0.0.23", EvidenceItem{Type: "hostname", Source: "netbios", Value: "Galaxy-S23", Strength: StrengthConfirmed})

	if e.MobileFingerprint == nil || e.MobileFingerprint.Classification != mobileClassPossibleAndroid {
		t.Fatalf("classification = %#v, want possible Android", e.MobileFingerprint)
	}
	if e.OSHint != mobileOSAndroid {
		t.Fatalf("os hint = %q, want android", e.OSHint)
	}
}

func TestDevicePool_GenericPortsDoNotClassifyMobile(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value string
	}{
		{name: "UDP 5353 alone", value: "udp/5353 mdns"},
		{name: "TCP 443 alone", value: "tcp/443 https"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := NewDevicePool(DefaultConfig())
			p.EnsureCandidate("10.0.0.24", SourcePassive)
			e, _, _ := p.AddEvidence("10.0.0.24", EvidenceItem{Type: "service_port", Source: "passive_packet", Value: tt.value, Strength: StrengthWeak})

			if e.MobileFingerprint == nil {
				t.Fatal("expected fingerprint with warning")
			}
			if e.MobileFingerprint.Classification != mobileClassUnknownDevice {
				t.Fatalf("classification = %s, want unknown_device", e.MobileFingerprint.Classification)
			}
			if e.MobileFingerprint.IOSScore != 0 || e.MobileFingerprint.AndroidScore != 0 || e.MobileFingerprint.IPadScore != 0 {
				t.Fatalf("generic port scored OS evidence: %+v", e.MobileFingerprint)
			}
		})
	}
}

func TestDevicePool_RandomizedMACDowngradesConfidence(t *testing.T) {
	p := NewDevicePool(DefaultConfig())
	p.EnsureCandidate("10.0.0.25", SourcePassive)
	p.AddEvidence("10.0.0.25", EvidenceItem{Type: "mac", Source: "arp_table", Value: "02:00:00:12:34:56", Strength: StrengthConfirmed})
	e, _, _ := p.AddEvidence("10.0.0.25", EvidenceItem{Type: "oui_vendor", Source: "lookup", Value: "Apple", Strength: StrengthInferred})

	if e.MobileFingerprint == nil {
		t.Fatal("missing fingerprint")
	}
	if e.MobileFingerprint.IOSScore >= 35 {
		t.Fatalf("randomized Apple MAC score = %d, want downgraded below clear OUI score", e.MobileFingerprint.IOSScore)
	}
	if !containsText(e.MobileFingerprint.Warnings, "randomization") {
		t.Fatalf("warnings = %#v, want MAC randomization warning", e.MobileFingerprint.Warnings)
	}
}

func TestDevicePool_ConflictingAppleAndroidEvidence(t *testing.T) {
	p := NewDevicePool(DefaultConfig())
	p.EnsureCandidate("10.0.0.26", SourcePassive)
	p.AddEvidence("10.0.0.26", EvidenceItem{Type: "hostname", Source: "dhcp", Value: "KIRAN-iPhone", Strength: StrengthConfirmed})
	e, _, _ := p.AddEvidence("10.0.0.26", EvidenceItem{Type: "oui_vendor", Source: "lookup", Value: "Samsung", Strength: StrengthInferred})

	if e.MobileFingerprint == nil || e.MobileFingerprint.Classification != mobileClassConflict {
		t.Fatalf("classification = %#v, want conflict", e.MobileFingerprint)
	}
	if len(e.MobileFingerprint.Conflicts) == 0 {
		t.Fatal("conflict details missing")
	}
}

func TestDevicePool_NoFakeModelNameInvented(t *testing.T) {
	p := NewDevicePool(DefaultConfig())
	p.EnsureCandidate("10.0.0.27", SourcePassive)
	e, _, _ := p.AddEvidence("10.0.0.27", EvidenceItem{Type: "hostname", Source: "dhcp", Value: "Pixel-8", Strength: StrengthConfirmed})

	blob := e.MobileFingerprint.WhyThisClassification + " " + e.MobileFingerprint.WhyNotCertain
	for _, fake := range []string{"Pixel 9", "iPhone 15", "Galaxy S24"} {
		if strings.Contains(blob, fake) {
			t.Fatalf("invented model name %q in explanation: %s", fake, blob)
		}
	}
}

func TestDevicePool_DNSPrivacyStoresOnlyOSCategories(t *testing.T) {
	p := NewDevicePool(DefaultConfig())
	p.EnsureCandidate("10.0.0.28", SourcePassive)
	if _, changed, _ := p.AddEvidence("10.0.0.28", EvidenceItem{Type: "dns_query", Source: "passive_dns", Value: "news.example.com/path", Strength: StrengthWeak}); changed {
		t.Fatal("unrelated DNS query should not be stored")
	}
	e, changed, _ := p.AddEvidence("10.0.0.28", EvidenceItem{Type: "dns_query", Source: "passive_dns", Value: "p42-escrowproxy.icloud.com", Strength: StrengthWeak})
	if !changed || len(e.Evidence) != 1 {
		t.Fatalf("expected one privacy-safe DNS evidence item, got changed=%v evidence=%#v", changed, e.Evidence)
	}
	if e.Evidence[0].Value != "apple_service_domain_seen" {
		t.Fatalf("DNS value = %q, want privacy-safe category", e.Evidence[0].Value)
	}
	if strings.Contains(e.Evidence[0].Value, "icloud.com") {
		t.Fatalf("raw DNS domain leaked into registry evidence: %#v", e.Evidence[0])
	}
}

// --- Scheduler ---

func TestScheduler_ConcurrencyCap(t *testing.T) {
	s := &Scheduler{Concurrency: 4, RatePerSec: 1000, Jitter: 0, rng: nil}
	// (rng nil is fine: Jitter==0 path never touches it.)
	var cur, max int32
	probe := func(ctx context.Context, ip string) ProbeOutcome {
		c := atomic.AddInt32(&cur, 1)
		for {
			old := atomic.LoadInt32(&max)
			if c <= old || atomic.CompareAndSwapInt32(&max, old, c) {
				break
			}
		}
		time.Sleep(25 * time.Millisecond)
		atomic.AddInt32(&cur, -1)
		return ProbeOutcome{IP: ip, Reachable: true}
	}
	tasks := make(chan string)
	results := make(chan ProbeOutcome, 64)
	go func() {
		for i := 0; i < 20; i++ {
			tasks <- "10.0.0.1"
		}
		close(tasks)
	}()
	go func() {
		for range results {
		}
	}()
	s.Run(context.Background(), tasks, probe, results)
	if max > 4 {
		t.Fatalf("concurrency cap exceeded: max=%d", max)
	}
	if max < 2 {
		t.Fatalf("worker pool did not parallelize: max=%d", max)
	}
}

func TestScheduler_RateLimiting(t *testing.T) {
	s := &Scheduler{Concurrency: 8, RatePerSec: 20, Jitter: 0}
	probe := func(ctx context.Context, ip string) ProbeOutcome { return ProbeOutcome{IP: ip, Reachable: true} }
	tasks := make(chan string)
	results := make(chan ProbeOutcome, 64)
	go func() {
		for range results {
		}
	}()
	go func() {
		for i := 0; i < 10; i++ {
			tasks <- "10.0.0.1"
		}
		close(tasks)
	}()
	start := time.Now()
	s.Run(context.Background(), tasks, probe, results)
	elapsed := time.Since(start)
	// 10 probes at 20/s ≈ 500ms; allow slack but require the limiter to bite.
	if elapsed < 350*time.Millisecond {
		t.Fatalf("rate limiter not enforced: 10 probes took only %v", elapsed)
	}
}

func TestScheduler_ContextCancellation(t *testing.T) {
	s := &Scheduler{Concurrency: 4, RatePerSec: 1000, Jitter: 0}
	probe := func(ctx context.Context, ip string) ProbeOutcome {
		<-ctx.Done() // block until cancelled
		return ProbeOutcome{IP: ip}
	}
	tasks := make(chan string, 8)
	for i := 0; i < 8; i++ {
		tasks <- "10.0.0.1"
	}
	results := make(chan ProbeOutcome, 8)
	go func() {
		for range results {
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { s.Run(ctx, tasks, probe, results); close(done) }()
	time.Sleep(40 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not stop promptly after cancellation")
	}
}

// --- Manager / duplicate suppression / topology payload ---

type fakeEmitter struct {
	mu       sync.Mutex
	events   map[string]int
	payloads map[string][]any
}

func newFakeEmitter() *fakeEmitter {
	return &fakeEmitter{events: map[string]int{}, payloads: map[string][]any{}}
}
func (f *fakeEmitter) Emit(event string, payload any) {
	f.mu.Lock()
	f.events[event]++
	f.payloads[event] = append(f.payloads[event], payload)
	f.mu.Unlock()
}

func TestManager_EnqueueSuppressesDuplicates(t *testing.T) {
	m := New(DefaultConfig(), newFakeEmitter(), nil)
	m.taskCh = make(chan string, 16)
	m.enqueue("10.0.0.2")
	m.enqueue("10.0.0.2") // in-flight: must not double-enqueue
	if len(m.taskCh) != 1 {
		t.Fatalf("duplicate enqueue not suppressed: queue=%d", len(m.taskCh))
	}
}

func TestManager_TopologyPayloadExcludesCandidates(t *testing.T) {
	m := New(DefaultConfig(), newFakeEmitter(), nil)
	// One responded (active), one stale, one raw candidate.
	m.pool.RecordSuccess("192.168.31.10", 2, 64, SourceCandidate)
	m.pool.RecordSuccess("192.168.31.20", 2, 64, SourceCandidate)
	m.pool.RecordFailure("192.168.31.20")
	m.pool.RecordFailure("192.168.31.20")
	m.pool.RecordFailure("192.168.31.20") // -> stale
	m.pool.EnsureCandidate("192.168.31.30", SourceCandidate)

	devs := m.topologyDevices()
	got := map[string]DeviceStatus{}
	for _, d := range devs {
		got[d.IP] = d.Status
	}
	if got["192.168.31.10"] != StatusActive {
		t.Fatalf("active device missing from topology: %+v", got)
	}
	if got["192.168.31.20"] != StatusStale {
		t.Fatalf("stale device missing from topology: %+v", got)
	}
	if _, present := got["192.168.31.30"]; present {
		t.Fatal("raw candidate must NOT appear on the topology")
	}
}

func TestManager_EmitsMobileFingerprintDiscoveryEvents(t *testing.T) {
	em := newFakeEmitter()
	m := New(DefaultConfig(), em, nil)

	if ok := m.AddDeviceEvidence("192.168.31.44", EvidenceItem{Type: "hostname", Source: "dhcp", Value: "Pixel-8", Strength: StrengthConfirmed}); !ok {
		t.Fatal("AddDeviceEvidence returned false")
	}

	em.mu.Lock()
	defer em.mu.Unlock()
	if em.events[EvtDeviceMobileFingerprintUpdated] != 1 {
		t.Fatalf("mobile event count = %d, want 1", em.events[EvtDeviceMobileFingerprintUpdated])
	}
	if em.events[EvtDiscoveryDeviceUpdated] == 0 {
		t.Fatal("discovery:device_updated was not emitted")
	}
	if em.events[EvtDiscoveryTopologyUpdated] == 0 {
		t.Fatal("discovery:topology_updated was not emitted")
	}
	payload, ok := em.payloads[EvtDeviceMobileFingerprintUpdated][0].(MobileFingerprintUpdatedPayload)
	if !ok {
		t.Fatalf("mobile payload type = %T", em.payloads[EvtDeviceMobileFingerprintUpdated][0])
	}
	if payload.MobileFingerprint.Classification != mobileClassPossibleAndroid {
		t.Fatalf("classification = %s, want possible Android", payload.MobileFingerprint.Classification)
	}
}

func TestManager_StartStopLifecycle(t *testing.T) {
	em := newFakeEmitter()
	cfg := DefaultConfig()
	cfg.ActiveInterval = 50 * time.Millisecond
	cfg.StaleInterval = 50 * time.Millisecond
	cfg.BatchInterval = 30 * time.Millisecond
	m := New(cfg, em, nil)

	if err := m.Start([]string{"192.168.31.2"}, false); err != nil {
		t.Fatalf("start: %v", err)
	}
	if !m.Running() {
		t.Fatal("manager should be running")
	}
	if err := m.Start(nil, false); err == nil {
		t.Fatal("double start should error")
	}
	time.Sleep(120 * time.Millisecond)
	m.Stop()
	if m.Running() {
		t.Fatal("manager should be stopped")
	}
	em.mu.Lock()
	started, stopped := em.events[EvtStarted], em.events[EvtStopped]
	em.mu.Unlock()
	if started == 0 || stopped == 0 {
		t.Fatalf("lifecycle events missing: started=%d stopped=%d", started, stopped)
	}
}

// --- ping parser ---

func TestParsePing_WindowsAndUnreachable(t *testing.T) {
	win := "Reply from 192.168.31.2: bytes=32 time=2ms TTL=64\nPackets: Sent = 1, Received = 1, Lost = 0 (0% loss),\nAverage = 2ms"
	r := parsePing(win)
	if !r.Reachable || r.TTL != 64 || r.LatencyMs != 2 {
		t.Fatalf("windows ping parse = %+v", r)
	}
	if u := parsePing("Request timed out.\nPackets: Sent = 1, Received = 0, Lost = 1 (100% loss),"); u.Reachable {
		t.Fatalf("100%% loss must be unreachable: %+v", u)
	}
}

func containsText(values []string, needle string) bool {
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
