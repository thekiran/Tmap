package ippool

import (
	"context"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ProbeOutcome is the result of one read-only reachability probe.
type ProbeOutcome struct {
	IP        string
	Reachable bool
	LatencyMs float64
	TTL       int
	LossPct   float64
	Method    string // icmp_ping | tcp_connect | none
	Err       string
}

// HideConsole optionally hides the spawned ping's console window (Windows). The
// app wires this to its hideConsole helper; tests leave it nil.
type HideConsoleFunc func(*exec.Cmd)

// ReachabilityProbe performs a safe, bounded ping with a TCP-connect fallback.
// It never logs in or sends anything but ICMP echo / a TCP SYN.
type ReachabilityProbe struct {
	Timeout    time.Duration
	Count      int
	TCPPorts   []int // fallback connect ports when ping is blocked
	HideWindow HideConsoleFunc
}

func NewReachabilityProbe(cfg Config, hide HideConsoleFunc) ReachabilityProbe {
	count := cfg.PingCount
	if count <= 0 {
		count = 1
	}
	return ReachabilityProbe{
		Timeout:    cfg.PingTimeout,
		Count:      count,
		TCPPorts:   []int{80, 443, 22, 53, 7547},
		HideWindow: hide,
	}
}

// Probe runs ping; if ping is unreachable/blocked it tries a quick TCP connect.
// A failure of one IP never panics or blocks — it returns an outcome with Err.
func (p ReachabilityProbe) Probe(ctx context.Context, ip string) ProbeOutcome {
	out := ProbeOutcome{IP: ip, Method: "none"}
	pctx, cancel := context.WithTimeout(ctx, p.Timeout+time.Second)
	res, err := p.ping(pctx, ip)
	cancel()
	if res.Reachable {
		res.Method = "icmp_ping"
		return res
	}
	if err != nil {
		out.Err = err.Error()
	}

	// TCP fallback: a SYN-ACK on any common port also proves reachability.
	for _, port := range p.TCPPorts {
		dctx, dcancel := context.WithTimeout(ctx, p.Timeout)
		start := time.Now()
		var d net.Dialer
		conn, derr := d.DialContext(dctx, "tcp", net.JoinHostPort(ip, strconv.Itoa(port)))
		dcancel()
		if derr == nil {
			_ = conn.Close()
			out.Reachable = true
			out.Method = "tcp_connect"
			out.LatencyMs = float64(time.Since(start).Microseconds()) / 1000.0
			out.Err = ""
			return out
		}
	}
	return out
}

func (p ReachabilityProbe) ping(ctx context.Context, ip string) (ProbeOutcome, error) {
	cmd := exec.CommandContext(ctx, "ping", pingArgs(ip, p.Count, p.Timeout)...)
	if p.HideWindow != nil {
		p.HideWindow(cmd)
	}
	raw, err := cmd.CombinedOutput()
	res := parsePing(string(raw))
	res.IP = ip
	return res, err
}

func pingArgs(ip string, count int, timeout time.Duration) []string {
	ms := max(int(timeout/time.Millisecond), 500)
	if runtime.GOOS == "windows" {
		return []string{"-n", strconv.Itoa(count), "-w", strconv.Itoa(ms), ip}
	}
	secs := max(ms/1000, 1)
	return []string{"-c", strconv.Itoa(count), "-W", strconv.Itoa(secs), ip}
}

var (
	reTTL      = regexp.MustCompile(`(?i)ttl[=\s:]+(\d+)`)
	reWinAvg   = regexp.MustCompile(`(?i)Average\s*=\s*(\d+)\s*ms`)
	reUnixStat = regexp.MustCompile(`=\s*[\d.]+/([\d.]+)/[\d.]+`)
	reLossWin  = regexp.MustCompile(`\((\d+)%\s*loss\)`)
	reLossUnix = regexp.MustCompile(`([\d.]+)%\s*packet\s*loss`)
)

// parsePing extracts reachable/latency/TTL/loss from Windows or Unix output.
// Exposed for unit testing without spawning a process.
func parsePing(output string) ProbeOutcome {
	var out ProbeOutcome
	if m := reTTL.FindStringSubmatch(output); m != nil {
		out.TTL, _ = strconv.Atoi(m[1])
	}
	if m := reLossWin.FindStringSubmatch(output); m != nil {
		out.LossPct, _ = strconv.ParseFloat(m[1], 64)
	} else if m := reLossUnix.FindStringSubmatch(output); m != nil {
		out.LossPct, _ = strconv.ParseFloat(m[1], 64)
	}
	if m := reWinAvg.FindStringSubmatch(output); m != nil {
		out.LatencyMs, _ = strconv.ParseFloat(m[1], 64)
	} else if m := reUnixStat.FindStringSubmatch(output); m != nil {
		out.LatencyMs, _ = strconv.ParseFloat(m[1], 64)
	}
	lower := strings.ToLower(output)
	lostAll := out.LossPct >= 100 ||
		strings.Contains(lower, "100% packet loss") ||
		strings.Contains(lower, "100% loss") ||
		strings.Contains(lower, "unreachable") ||
		strings.Contains(lower, "request timed out")
	gotReply := out.TTL > 0 || out.LatencyMs > 0 || strings.Contains(lower, "bytes from") || strings.Contains(lower, "reply from")
	out.Reachable = gotReply && !lostAll
	return out
}
