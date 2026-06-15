package probes

import (
	"context"
	"fmt"
	"math"
	"net"
	"time"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

// LatencyProbe measures round-trip latency and jitter. It first tries ICMP
// (via pro-bing); when ICMP is unavailable — typically because raw sockets need
// elevation on Windows — it falls back to timing TCP connects, so the probe
// still produces a result for an unprivileged user.
//
// In online mode it targets a public anycast resolver; in offline mode it
// targets the LAN gateway so no traffic leaves the local network.
type LatencyProbe struct{}

func (LatencyProbe) Name() string { return "latency_probe" }

const latencySamples = 5

func (p LatencyProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())

	target := "1.1.1.1"
	if !in.Online {
		gw, err := network.Gateway()
		if err != nil {
			res.Status = models.StatusSkipped
			return res, nil
		}
		target = gw.String()
	}
	res.Evidence["target"] = target

	avg, jitter, method, err := measure(ctx, target)
	if err != nil {
		return res, err
	}
	avgMS := float64(avg) / float64(time.Millisecond)
	jitterMS := float64(jitter) / float64(time.Millisecond)
	res.Evidence["avg_ms"] = round1(avgMS)
	res.Evidence["jitter_ms"] = round1(jitterMS)
	res.Evidence["method"] = method
	res.Confidence = 0.3

	// Latency is weak, corroborating evidence only. The detection engine applies
	// the banded interpretation (and never lets latency decide a verdict alone),
	// so the probe just reports the measurement.
	return res, nil
}

// measure tries ICMP first and falls back to TCP-connect timing.
func measure(ctx context.Context, target string) (avg, jitter time.Duration, method string, err error) {
	if a, j, e := icmpMeasure(ctx, target); e == nil {
		return a, j, "icmp", nil
	}
	a, j, e := tcpMeasure(ctx, target, []string{"443", "80"})
	if e != nil {
		return 0, 0, "", fmt.Errorf("latency measurement failed: %w", e)
	}
	return a, j, "tcp", nil
}

func icmpMeasure(ctx context.Context, target string) (avg, jitter time.Duration, err error) {
	pinger, err := probing.NewPinger(target)
	if err != nil {
		return 0, 0, err
	}
	pinger.Count = latencySamples
	pinger.Interval = 200 * time.Millisecond
	pinger.Timeout = 4 * time.Second
	pinger.SetPrivileged(true) // required for raw ICMP on Windows
	if err := pinger.RunWithContext(ctx); err != nil {
		return 0, 0, err
	}
	st := pinger.Statistics()
	if st.PacketsRecv == 0 {
		return 0, 0, fmt.Errorf("no ICMP replies")
	}
	return st.AvgRtt, st.StdDevRtt, nil
}

func tcpMeasure(ctx context.Context, host string, ports []string) (avg, jitter time.Duration, err error) {
	var samples []time.Duration
	var lastErr error
	d := net.Dialer{Timeout: 3 * time.Second}
	for _, port := range ports {
		samples = samples[:0]
		addr := net.JoinHostPort(host, port)
		for i := 0; i < latencySamples; i++ {
			start := time.Now()
			conn, e := d.DialContext(ctx, "tcp", addr)
			if e != nil {
				lastErr = e
				break
			}
			samples = append(samples, time.Since(start))
			conn.Close()
		}
		if len(samples) == latencySamples {
			a, j := stats(samples)
			return a, j, nil
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no reachable port")
	}
	return 0, 0, lastErr
}

// stats returns the mean and standard deviation (jitter) of the samples.
func stats(samples []time.Duration) (avg, jitter time.Duration) {
	if len(samples) == 0 {
		return 0, 0
	}
	var sum float64
	for _, s := range samples {
		sum += float64(s)
	}
	mean := sum / float64(len(samples))
	var variance float64
	for _, s := range samples {
		d := float64(s) - mean
		variance += d * d
	}
	variance /= float64(len(samples))
	return time.Duration(mean), time.Duration(math.Sqrt(variance))
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
