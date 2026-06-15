// Package probes collects evidence about the host's connection. Every probe
// emits the same models.ProbeResult shape (doc §10), which is what lets the
// detection engine and UI treat them uniformly and lets new probes be added
// without touching anything downstream.
package probes

import (
	"context"
	"sync"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

// Probe collects one category of evidence.
type Probe interface {
	Name() string
	Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error)
}

// newResult returns a success-status result with an initialized evidence map.
func newResult(name string) *models.ProbeResult {
	return &models.ProbeResult{
		ProbeName: name,
		Status:    models.StatusSuccess,
		Evidence:  map[string]any{},
	}
}

// Runner executes a set of probes concurrently, each bounded by Timeout. A
// failing or panicking probe never aborts the scan: its result is recorded with
// status "failed" and the others continue (graceful degradation, doc §cross-platform).
type Runner struct {
	Probes  []Probe
	Timeout time.Duration
}

// Run executes every probe and returns results in the same order as r.Probes.
func (r *Runner) Run(ctx context.Context, in models.ScanInput) []models.ProbeResult {
	results := make([]models.ProbeResult, len(r.Probes))
	var wg sync.WaitGroup
	for i, p := range r.Probes {
		wg.Add(1)
		go func(i int, p Probe) {
			defer wg.Done()
			defer func() {
				if rec := recover(); rec != nil {
					results[i] = models.ProbeResult{
						ProbeName: p.Name(),
						Status:    models.StatusFailed,
						Errors:    []string{"panic during probe"},
					}
				}
			}()
			pctx, cancel := context.WithTimeout(ctx, r.Timeout)
			defer cancel()
			res, err := p.Run(pctx, in)
			if res == nil {
				res = &models.ProbeResult{ProbeName: p.Name(), Status: models.StatusFailed}
			}
			if err != nil {
				res.Status = models.StatusFailed
				res.Errors = append(res.Errors, err.Error())
			}
			results[i] = *res
		}(i, p)
	}
	wg.Wait()
	return results
}

// Default returns the probe set for the given scan input. Quick scans run the
// fast LAN-side probes; deep scans add traceroute and ASN. Online-only probes
// are omitted entirely when the user runs offline.
func Default(in models.ScanInput) []Probe {
	ps := []Probe{
		AdapterProbe{},
		GatewayProbe{},
		GatewayChainProbe{},
		DNSProbe{},
		LatencyProbe{},
		UPnPProbe{},
		UPnPIGDProbe{},
		UPnPIGDDeepProbe{},
		UPnPIGDDeepProbeV2{},
		TR064Probe{},
		TR064ProbeV2{},
		HTTPFingerprintV2Probe{},
		HTTPFingerprintV3Probe{},
		IPv6TransitionProbe{},
		IPv6TransitionProbeV2{},
		STUNPCPNATProbe{},
		OSInterfaceProbeV2{},
		OSInterfaceProbeV3{},
		LLDPCDPPassiveProbe{},
		LLDPCDPPassiveProbeV2{},
		PerformanceProfileProbe{},
		PerformanceProfileProbeV2{},
		SNMPProbeOptIn{},
		GatewayReachabilityDiagnosticsProbe{},
	}
	if in.Online {
		ps = append(ps, PublicIPProbe{})
	}
	if in.Mode == models.ModeDeep {
		ps = append(ps, TracerouteProbe{})
		if in.Online {
			ps = append(ps, ASNProbe{})
		}
	}
	return ps
}
