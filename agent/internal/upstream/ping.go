package upstream

import (
	"context"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// PingResult is the parsed output of a short system ping.
type PingResult struct {
	Reachable bool
	AvgMs     *float64
	MinMs     *float64
	MaxMs     *float64
	TTL       *int
	LossPct   *float64
}

// pingArgs builds a safe, bounded ping invocation: no continuous ping, a low
// packet count, and a short timeout. Never requires admin / raw sockets.
func pingArgs(ip string) []string {
	if runtime.GOOS == "windows" {
		// -n 2 = 2 echoes, -w 1000 = 1s per-reply timeout (milliseconds).
		return []string{"-n", "2", "-w", "1000", ip}
	}
	// -c 2 = 2 echoes, -W 1 = 1s timeout (seconds; macOS treats -W as ms but
	// still bounds the run, which is all we need).
	return []string{"-c", "2", "-W", "1", ip}
}

// Ping runs one short system ping against ip and parses the result. It returns
// the parsed result even on a non-zero exit (e.g. 100% loss) so the caller can
// record "unreachable" rather than treating it as a hard error.
func Ping(ctx context.Context, ip string) (PingResult, error) {
	cmd := exec.CommandContext(ctx, "ping", pingArgs(ip)...)
	out, err := cmd.CombinedOutput()
	res := ParsePing(string(out))
	return res, err
}

var (
	reTTL      = regexp.MustCompile(`(?i)ttl[=\s:]+(\d+)`)
	reWinAvg   = regexp.MustCompile(`(?i)Average\s*=\s*(\d+)\s*ms`)
	reWinMin   = regexp.MustCompile(`(?i)Minimum\s*=\s*(\d+)\s*ms`)
	reWinMax   = regexp.MustCompile(`(?i)Maximum\s*=\s*(\d+)\s*ms`)
	reUnixStat = regexp.MustCompile(`=\s*([\d.]+)/([\d.]+)/([\d.]+)(?:/([\d.]+))?\s*ms`)
	reLossWin  = regexp.MustCompile(`\((\d+)%\s*loss\)`)
	reLossUnix = regexp.MustCompile(`([\d.]+)%\s*packet\s*loss`)
)

// ParsePing extracts latency / TTL / loss from Windows or Unix ping output.
// Exposed for unit testing the parser without spawning a process.
func ParsePing(output string) PingResult {
	var res PingResult

	if m := reTTL.FindStringSubmatch(output); m != nil {
		if v, err := strconv.Atoi(m[1]); err == nil {
			res.TTL = &v
		}
	}

	// Loss: Windows "(0% loss)" or Unix "0% packet loss".
	if m := reLossWin.FindStringSubmatch(output); m != nil {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			res.LossPct = &v
		}
	} else if m := reLossUnix.FindStringSubmatch(output); m != nil {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			res.LossPct = &v
		}
	}

	// Latency: Windows reports Min/Max/Average; Unix reports min/avg/max/mdev.
	if m := reWinAvg.FindStringSubmatch(output); m != nil {
		res.AvgMs = parseFloatPtr(m[1])
		if mn := reWinMin.FindStringSubmatch(output); mn != nil {
			res.MinMs = parseFloatPtr(mn[1])
		}
		if mx := reWinMax.FindStringSubmatch(output); mx != nil {
			res.MaxMs = parseFloatPtr(mx[1])
		}
	} else if m := reUnixStat.FindStringSubmatch(output); m != nil {
		res.MinMs = parseFloatPtr(m[1])
		res.AvgMs = parseFloatPtr(m[2])
		res.MaxMs = parseFloatPtr(m[3])
	}

	// Reachable if we got any reply latency or an explicit <100% loss, and we
	// actually saw a reply marker.
	gotReply := res.AvgMs != nil || res.TTL != nil
	lostAll := res.LossPct != nil && *res.LossPct >= 100
	lower := strings.ToLower(output)
	if strings.Contains(lower, "unreachable") || strings.Contains(lower, "100% packet loss") || strings.Contains(lower, "100% loss") {
		lostAll = true
	}
	res.Reachable = gotReply && !lostAll
	return res
}

func parseFloatPtr(s string) *float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}
