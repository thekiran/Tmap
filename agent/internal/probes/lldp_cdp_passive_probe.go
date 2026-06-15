package probes

import (
	"context"
	"strings"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

type LLDPCDPPassiveProbe struct {
	funcs lldpCDPFuncs
}

type lldpCDPFuncs struct {
	read func(context.Context, time.Duration) ([]LLDPCDPFrame, error)
}

type LLDPCDPFrame struct {
	Protocol          string
	SystemName        string
	SystemDescription string
	Capabilities      []string
	PortID            string
}

func (LLDPCDPPassiveProbe) Name() string { return "lldp_cdp_passive_probe" }

func (p LLDPCDPPassiveProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	f := p.withDefaults()
	frames, err := f.read(ctx, 1200*time.Millisecond)
	if err != nil {
		return res, err
	}
	res.Evidence["frames"] = frames
	res.Evidence["network_confidence"] = 0.20
	var hints []string
	for _, frame := range frames {
		text := strings.Join([]string{
			frame.Protocol,
			frame.SystemName,
			frame.SystemDescription,
			strings.Join(frame.Capabilities, " "),
			frame.PortID,
		}, " ")
		for _, h := range inferAccessHints(text) {
			hints = appendUnique(hints, h)
		}
	}
	if len(hints) > 0 {
		res.Hints = hints
		res.Evidence["strong_access_evidence"] = true
		res.Evidence["access_confidence"] = 0.70
	} else {
		res.Evidence["access_confidence"] = 0.0
	}
	res.Confidence = 0.20
	return res, nil
}

func (p LLDPCDPPassiveProbe) withDefaults() lldpCDPFuncs {
	f := p.funcs
	if f.read == nil {
		f.read = func(ctx context.Context, wait time.Duration) ([]LLDPCDPFrame, error) {
			return nil, nil
		}
	}
	return f
}
