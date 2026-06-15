package probes

import (
	"context"
	"net"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

type IPv6TransitionProbe struct {
	funcs ipv6TransitionFuncs
}

type ipv6TransitionFuncs struct {
	interfaces func() ([]net.Interface, error)
	addrs      func(net.Interface) ([]net.Addr, error)
}

func (IPv6TransitionProbe) Name() string { return "ipv6_transition_probe" }

func (p IPv6TransitionProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	f := p.withDefaults()
	ifaces, err := f.interfaces()
	if err != nil {
		return res, err
	}
	var hints []string
	ipv6Available := false
	globalIPv6 := false
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := f.addrs(ifc)
		name := strings.ToLower(ifc.Name)
		if strings.Contains(name, "clat") || strings.Contains(name, "464") {
			hints = appendUnique(hints, "464XLAT")
		}
		if strings.Contains(name, "dslite") || strings.Contains(name, "ds-lite") {
			hints = appendUnique(hints, "DS-Lite")
		}
		if strings.Contains(name, "map-e") || strings.Contains(name, "mape") {
			hints = appendUnique(hints, "MAP-E")
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil || ip == nil || ip.To4() != nil {
				continue
			}
			ipv6Available = true
			if ip.IsGlobalUnicast() && !ip.IsPrivate() {
				globalIPv6 = true
			}
		}
	}
	ctxInfo := models.IPv6Context{
		IPv6Available: ipv6Available,
		GlobalIPv6:    globalIPv6,
		DefaultRoute:  ipv6RouteLabel(ipv6Available, globalIPv6),
		DNS64NAT64:    false,
		TransitionHints: hints,
	}
	res.Evidence["ipv6_context"] = ctxInfo
	res.Evidence["ip_architecture"] = ipArchitecture(ctxInfo)
	res.Evidence["network_confidence"] = 0.30
	res.Confidence = 0.30
	return res, nil
}

func (p IPv6TransitionProbe) withDefaults() ipv6TransitionFuncs {
	f := p.funcs
	if f.interfaces == nil {
		f.interfaces = net.Interfaces
	}
	if f.addrs == nil {
		f.addrs = func(ifc net.Interface) ([]net.Addr, error) {
			return ifc.Addrs()
		}
	}
	return f
}

func ipv6RouteLabel(available, global bool) string {
	switch {
	case global:
		return "dual_stack_or_ipv6_default"
	case available:
		return "local_or_ula_ipv6"
	default:
		return "ipv4_only_observed"
	}
}

func ipArchitecture(ctx models.IPv6Context) string {
	if len(ctx.TransitionHints) > 0 {
		return strings.Join(ctx.TransitionHints, ",")
	}
	if ctx.GlobalIPv6 {
		return "dual_stack"
	}
	if ctx.IPv6Available {
		return "ipv6_local_only"
	}
	return "ipv4_only"
}

