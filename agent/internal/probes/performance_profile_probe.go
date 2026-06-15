package probes

import (
	"context"
	"time"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

type PerformanceProfileProbe struct {
	funcs performanceProfileFuncs
}

type performanceProfileFuncs struct {
	gateway func() (string, error)
	measure func(context.Context, string) (time.Duration, time.Duration, string, error)
}

func (PerformanceProfileProbe) Name() string { return "performance_profile_probe" }

func (p PerformanceProfileProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	f := p.withDefaults()
	target := "1.1.1.1"
	if !in.Online {
		gw, err := f.gateway()
		if err != nil {
			res.Status = models.StatusSkipped
			return res, nil
		}
		target = gw
	}
	avg, jitter, method, err := f.measure(ctx, target)
	if err != nil {
		return res, err
	}
	profile := models.PerformanceProfile{
		Target:        target,
		Method:        method,
		IdleLatencyMS: round1(float64(avg) / float64(time.Millisecond)),
		JitterMS:      round1(float64(jitter) / float64(time.Millisecond)),
		PacketLossPct: 0,
	}
	res.Evidence["performance_profile"] = profile
	res.Evidence["performance_confidence"] = 0.35
	res.Confidence = 0.35
	return res, nil
}

func (p PerformanceProfileProbe) withDefaults() performanceProfileFuncs {
	f := p.funcs
	if f.gateway == nil {
		f.gateway = func() (string, error) {
			gw, err := network.Gateway()
			if err != nil {
				return "", err
			}
			return gw.String(), nil
		}
	}
	if f.measure == nil {
		f.measure = measure
	}
	return f
}

