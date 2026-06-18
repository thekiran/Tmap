package discovery

import (
	"context"
	"net"
	"os/exec"
	"sort"
	"sync"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

// HostHit is the liveness result for one address.
type HostHit struct {
	IP        string
	Alive     bool
	OpenPorts []int // TCP ports that accepted a full connection
}

// Sweeper finds live hosts within a validated scope. It is an interface so the
// real network sweep can be swapped for a deterministic fake in tests.
type Sweeper interface {
	Sweep(ctx context.Context, scope models.ScanScope) ([]HostHit, error)
}

// ARPReader reads the OS neighbour/ARP table.
type ARPReader interface {
	Read(ctx context.Context) ([]ARPEntry, error)
}

// ARPSweeper actively resolves each in-scope address to a MAC (e.g. via the
// Windows SendARP API). This discovers L2 devices that answer ARP even when they
// have no open ports and drop ICMP — printers, IoT, phones. It returns one
// ARPEntry per responder. Implementations are platform-specific; the non-Windows
// build is a no-op and relies on the TCP-connect sweep populating the kernel
// neighbour table (read afterwards via ARPReader).
type ARPSweeper interface {
	SweepARP(ctx context.Context, scope models.ScanScope) []ARPEntry
}

// Resolver does reverse DNS. Returns "" (no error) when there is no PTR.
type Resolver interface {
	LookupAddr(ctx context.Context, ip string) string
}

// profilePorts returns the TCP ports the built-in sweeper probes for a profile.
// These are common service ports only — this is liveness/identification, not a
// full port scan (use Nmap for that). No stealth, SYN-only, or evasion behaviour.
func profilePorts(profile string) []int {
	switch profile {
	case "deep", "full":
		return []int{22, 23, 25, 53, 80, 110, 135, 139, 143, 443, 445, 587, 993, 995,
			1883, 3389, 5000, 5060, 5357, 7547, 8080, 8443, 8843, 9000, 49152}
	case "normal", "standard":
		return []int{22, 23, 53, 80, 135, 139, 443, 445, 3389, 5357, 7547, 8080, 8443}
	default: // quick
		return []int{22, 53, 80, 135, 443, 445, 3389}
	}
}

// TCPSweeper is the default Sweeper: a bounded-concurrency, full-TCP-connect
// liveness probe across the addresses in scope. A host is "alive" if any probed
// port accepts a connection. It is intentionally simple and non-intrusive.
type TCPSweeper struct {
	Profile     string
	Concurrency int           // default 64
	DialTimeout time.Duration // default 600ms
}

// Sweep probes every in-scope address. It honours ctx cancellation and never
// touches an address outside the scope (the address list comes from HostsInScope).
func (s TCPSweeper) Sweep(ctx context.Context, scope models.ScanScope) ([]HostHit, error) {
	hosts := HostsInScope(scope)
	if len(hosts) == 0 {
		return nil, nil
	}
	conc := s.Concurrency
	if conc <= 0 {
		conc = 64
	}
	timeout := s.DialTimeout
	if timeout <= 0 {
		timeout = 600 * time.Millisecond
	}
	ports := profilePorts(scope.Profile)
	if scope.Profile == "" {
		ports = profilePorts(s.Profile)
	}

	results := make([]HostHit, len(hosts))
	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup
	for i, ip := range hosts {
		select {
		case <-ctx.Done():
			wg.Wait()
			return collectAlive(results), ctx.Err()
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func(i int, ip string) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = probeHost(ctx, ip, ports, timeout)
		}(i, ip)
	}
	wg.Wait()
	return collectAlive(results), nil
}

func probeHost(ctx context.Context, ip string, ports []int, timeout time.Duration) HostHit {
	hit := HostHit{IP: ip}
	dialer := net.Dialer{Timeout: timeout}
	for _, p := range ports {
		if ctx.Err() != nil {
			break
		}
		conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(ip, itoa(p)))
		if err != nil {
			continue
		}
		_ = conn.Close()
		hit.Alive = true
		hit.OpenPorts = append(hit.OpenPorts, p)
	}
	return hit
}

func collectAlive(all []HostHit) []HostHit {
	var out []HostHit
	for _, h := range all {
		if h.Alive {
			sort.Ints(h.OpenPorts)
			out = append(out, h)
		}
	}
	return out
}

// OSARPReader reads the ARP/neighbour table via the OS tool, context-bounded. It
// tries `arp -a` then `ip neigh`, parsing whatever succeeds.
type OSARPReader struct{}

func (OSARPReader) Read(ctx context.Context) ([]ARPEntry, error) {
	for _, c := range [][]string{{"arp", "-a"}, {"ip", "neigh"}} {
		out, err := exec.CommandContext(ctx, c[0], c[1:]...).Output()
		if err != nil {
			continue
		}
		if entries := ParseARPTable(string(out)); len(entries) > 0 {
			return entries, nil
		}
	}
	return nil, nil
}

// NetResolver is the default reverse-DNS resolver.
type NetResolver struct {
	Timeout time.Duration // default 1s
}

func (r NetResolver) LookupAddr(ctx context.Context, ip string) string {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	names, err := net.DefaultResolver.LookupAddr(cctx, ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return trimDot(names[0])
}

func trimDot(s string) string {
	if len(s) > 0 && s[len(s)-1] == '.' {
		return s[:len(s)-1]
	}
	return s
}

// itoa avoids importing strconv in hot paths for a single small int.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [6]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
