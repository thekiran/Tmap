package probes

import (
	"context"

	"github.com/thekiran/iad/pkg/models"
)

type TR064ProbeV2 struct{ TR064Probe }
type UPnPIGDDeepProbeV2 struct{ UPnPIGDDeepProbe }
type HTTPFingerprintV3Probe struct{ HTTPFingerprintV2Probe }
type OSInterfaceProbeV3 struct{ OSInterfaceProbeV2 }
type LLDPCDPPassiveProbeV2 struct{ LLDPCDPPassiveProbe }
type IPv6TransitionProbeV2 struct{ IPv6TransitionProbe }
type PerformanceProfileProbeV2 struct{ PerformanceProfileProbe }

func (TR064ProbeV2) Name() string { return "tr064_probe_v2" }
func (p TR064ProbeV2) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res, err := p.TR064Probe.Run(ctx, in)
	if res != nil {
		res.ProbeName = p.Name()
	}
	return res, err
}

func (UPnPIGDDeepProbeV2) Name() string { return "upnp_igd_deep_probe_v2" }
func (p UPnPIGDDeepProbeV2) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res, err := p.UPnPIGDDeepProbe.Run(ctx, in)
	if res != nil {
		res.ProbeName = p.Name()
	}
	return res, err
}

func (HTTPFingerprintV3Probe) Name() string { return "http_fingerprint_v3" }
func (p HTTPFingerprintV3Probe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res, err := p.HTTPFingerprintV2Probe.Run(ctx, in)
	if res != nil {
		res.ProbeName = p.Name()
	}
	return res, err
}

func (OSInterfaceProbeV3) Name() string { return "os_interface_probe_v3" }
func (p OSInterfaceProbeV3) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res, err := p.OSInterfaceProbeV2.Run(ctx, in)
	if res != nil {
		res.ProbeName = p.Name()
	}
	return res, err
}

func (LLDPCDPPassiveProbeV2) Name() string { return "lldp_cdp_passive_probe_v2" }
func (p LLDPCDPPassiveProbeV2) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res, err := p.LLDPCDPPassiveProbe.Run(ctx, in)
	if res != nil {
		res.ProbeName = p.Name()
	}
	return res, err
}

func (IPv6TransitionProbeV2) Name() string { return "ipv6_transition_probe_v2" }
func (p IPv6TransitionProbeV2) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res, err := p.IPv6TransitionProbe.Run(ctx, in)
	if res != nil {
		res.ProbeName = p.Name()
	}
	return res, err
}

func (PerformanceProfileProbeV2) Name() string { return "performance_profile_probe_v2" }
func (p PerformanceProfileProbeV2) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res, err := p.PerformanceProfileProbe.Run(ctx, in)
	if res != nil {
		res.ProbeName = p.Name()
	}
	return res, err
}

