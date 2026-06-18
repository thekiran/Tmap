package deviceintel

import (
	"context"
	"time"
)

type Pipeline struct {
	Probes []Probe
}

func DefaultReadOnlyProbes() []Probe {
	return []Probe{
		SafeTCPProbe{},
		HTTPFingerprintProbe{},
		TLSFingerprintProbe{},
	}
}

func CredentialedOptInProbes() []Probe {
	return []Probe{
		SNMPOptInProbe(),
		SSHBannerOptInProbe(),
		RouterAPIOptInProbe(),
		TR064OptInProbe(),
		TR181OptInProbe(),
	}
}

func NewPipeline(probes ...Probe) Pipeline {
	if len(probes) == 0 {
		probes = DefaultReadOnlyProbes()
	}
	return Pipeline{Probes: probes}
}

func (p Pipeline) Run(ctx context.Context, scope ScanScope) (*EvidenceStore, []ProbeResult) {
	now := scope.Now
	if now == nil {
		now = time.Now
	}
	store := NewEvidenceStore(now)
	results := make([]ProbeResult, 0, len(p.Probes))
	for _, probe := range p.Probes {
		if probe == nil {
			continue
		}
		results = append(results, probe.Run(ctx, scope, store))
	}
	classifyAll(store)
	return store, results
}
