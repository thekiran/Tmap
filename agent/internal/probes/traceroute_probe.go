package probes

import (
	"context"
	"net"

	"github.com/thekiran/iad/internal/system"
	"github.com/thekiran/iad/pkg/models"
)

// TracerouteProbe traces the path to a public host. The shape of the first few
// hops is a weak hint: an early carrier-grade-NAT hop (10.x or 100.64/10)
// suggests mobile/FWA/WISP/satellite access. It needs the internet, so it is a
// deep+online probe and is skipped offline.
type TracerouteProbe struct{}

func (TracerouteProbe) Name() string { return "traceroute_probe" }

const traceTarget = "8.8.8.8"

func (p TracerouteProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	if !in.Online {
		res.Status = models.StatusSkipped
		return res, nil
	}
	hops, err := system.Traceroute(ctx, traceTarget, 15)
	if err != nil {
		return res, err
	}
	res.Evidence["hops"] = hops
	res.Evidence["hop_count"] = len(hops)
	res.Confidence = 0.25

	for _, h := range hops {
		if isCarrierNAT(h) {
			res.Hints = appendUnique(res.Hints, models.TypeMobile)
			res.Hints = appendUnique(res.Hints, models.TypeFWA)
			res.Evidence["carrier_nat_hop"] = h
			break
		}
	}
	return res, nil
}

// isCarrierNAT reports whether ip is in 100.64.0.0/10 (CGNAT). The first private
// hop (192.168/10.x home LAN) is excluded by only flagging the CGNAT block,
// which providers use upstream of the subscriber.
func isCarrierNAT(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	_, cgnat, _ := net.ParseCIDR("100.64.0.0/10")
	return cgnat.Contains(parsed)
}
