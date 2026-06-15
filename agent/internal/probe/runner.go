package probe

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/thekiran/iad/internal/config"
	"github.com/thekiran/iad/internal/model"
	"github.com/thekiran/iad/internal/safety"
)

type Probe interface {
	Name() string
	Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error)
	SafeModeAllowed() bool
	NormalModeAllowed() bool
	DeepModeAllowed() bool
}

type Runner struct {
	Probes []Probe
	Config config.Config
}

func DefaultProbes() []Probe {
	return []Probe{
		OSInterfacesProbe{},
		RouteTableProbe{},
		ARPNeighborProbe{},
		NDPNeighborProbe{},
		DHCPProbe{},
		WiFiProbe{},
		PingSweepProbe{},
		SSDPUPnPProbe{},
		MDNSProbe{},
		NetBIOSLLMNRProbe{},
		HTTPFingerprintProbe{},
		TCPConnectProbe{},
		UDPProbe{},
		SNMPOptInProbe{},
		LLDPCDPPassiveProbe{},
		TracerouteISPPathProbe{},
	}
}

func NewRunner(probes []Probe, cfg config.Config) Runner {
	if cfg.Mode == "" {
		cfg = config.Default(model.ScanModeSafe)
	}
	if len(probes) == 0 {
		probes = DefaultProbes()
	}
	return Runner{Probes: probes, Config: cfg}
}

func (r Runner) Run(ctx context.Context, input model.ProbeInput) []model.ProbeResult {
	cfg := r.Config
	if cfg.Mode == "" {
		cfg = config.Default(input.Mode)
	}
	if input.Mode == "" {
		input.Mode = cfg.Mode
	}
	input.SNMP = cfg.SNMP

	if err := safety.ValidateScope(input.Scope); err != nil {
		return []model.ProbeResult{failedResult("scope_validation", err)}
	}
	if err := safety.ValidatePrivateTargets(input.CandidateIPs); err != nil {
		return []model.ProbeResult{failedResult("target_validation", err)}
	}

	runCtx := ctx
	cancel := func() {}
	if cfg.GlobalTimeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, cfg.GlobalTimeout)
	}
	defer cancel()

	allowed := make([]Probe, 0, len(r.Probes))
	for _, p := range r.Probes {
		if modeAllowed(p, input.Mode) {
			allowed = append(allowed, p)
		}
	}

	results := make([]model.ProbeResult, len(allowed))
	limiter := safety.NewRateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)
	var wg sync.WaitGroup
	for i, p := range allowed {
		wg.Add(1)
		go func(i int, p Probe) {
			defer wg.Done()
			if err := limiter.Wait(runCtx); err != nil {
				results[i] = failedResult(p.Name(), err)
				return
			}
			results[i] = r.runOne(runCtx, p, input)
		}(i, p)
	}
	wg.Wait()
	return results
}

func (r Runner) runOne(ctx context.Context, p Probe, input model.ProbeInput) model.ProbeResult {
	started := time.Now().UTC()
	pctx := ctx
	cancel := func() {}
	if r.Config.PerProbeTimeout > 0 {
		pctx, cancel = context.WithTimeout(ctx, r.Config.PerProbeTimeout)
	}
	defer cancel()

	result, err := p.Run(pctx, input)
	if result.ProbeName == "" {
		result.ProbeName = p.Name()
	}
	result.StartedAt = started
	if result.FinishedAt.IsZero() {
		result.FinishedAt = time.Now().UTC()
	}
	if err != nil {
		result.Status = model.ProbeStatusFailed
		result.Errors = append(result.Errors, err.Error())
	}
	if result.Status == "" {
		result.Status = model.ProbeStatusSuccess
	}
	for i := range result.Evidence {
		if result.Evidence[i].Source == "" {
			result.Evidence[i].Source = p.Name()
		}
		if result.Evidence[i].Timestamp.IsZero() {
			result.Evidence[i].Timestamp = result.FinishedAt
		}
		result.Evidence[i].Confidence = model.Clamp01(result.Evidence[i].Confidence)
	}
	return result
}

func modeAllowed(p Probe, mode model.ScanMode) bool {
	switch mode {
	case model.ScanModeSafe:
		return p.SafeModeAllowed()
	case model.ScanModeNormal:
		return p.NormalModeAllowed()
	case model.ScanModeDeep:
		return p.DeepModeAllowed()
	default:
		return p.SafeModeAllowed()
	}
}

func failedResult(name string, err error) model.ProbeResult {
	now := time.Now().UTC()
	return model.ProbeResult{
		ProbeName:  name,
		Status:     model.ProbeStatusFailed,
		StartedAt:  now,
		FinishedAt: now,
		Errors:     []string{err.Error()},
	}
}

func skippedResult(name, reason string) model.ProbeResult {
	now := time.Now().UTC()
	return model.ProbeResult{
		ProbeName:  name,
		Status:     model.ProbeStatusSkipped,
		StartedAt:  now,
		FinishedAt: now,
		Errors:     []string{reason},
	}
}

func baseEvidence(source, target string, strength model.EvidenceStrength, confidence float64, reason string, raw map[string]any) model.Evidence {
	return model.NewEvidence(source, target, strength, confidence, reason, raw, time.Now().UTC())
}

func serviceName(port int) string {
	switch port {
	case 22:
		return "ssh"
	case 23:
		return "telnet"
	case 53:
		return "dns"
	case 80, 8080:
		return "http"
	case 443, 8443:
		return "https"
	case 445:
		return "smb"
	case 554:
		return "rtsp"
	case 631:
		return "ipp"
	case 1900:
		return "ssdp"
	case 5000, 5001:
		return "nas/http"
	case 5353:
		return "mdns"
	case 9100:
		return "printer"
	default:
		return fmt.Sprintf("tcp/%d", port)
	}
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range in {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
