package discovery

import (
	"github.com/thekiran/iad/internal/model"
	"github.com/thekiran/iad/internal/safety"
)

const ispPathWarning = "Traceroute shows observed route hops only. It does not reveal full ISP infrastructure."

func BuildISPPath(results []model.ProbeResult, ctx model.NetworkContext) model.ISPPath {
	path := model.ISPPath{
		PublicIP:     ctx.PublicIP,
		ASN:          ctx.ASN,
		Organization: ctx.ISP,
		Warning:      ispPathWarning,
	}
	if ctx.CGNAT != nil {
		path.CGNAT = *ctx.CGNAT
	}

	for _, result := range results {
		if hops, ok := result.Raw["route_hops"].([]model.RouteHop); ok {
			for _, hop := range hops {
				hop.Confidence = model.Clamp01(max(hop.Confidence, 0.50))
				if hop.Private || safety.IsPrivateIPString(hop.IP) {
					hop.Private = true
					path.PrivateHops = append(path.PrivateHops, hop)
					continue
				}
				if path.FirstPublicHop == "" {
					path.FirstPublicHop = hop.IP
				}
				path.PublicHops = append(path.PublicHops, hop)
			}
		}
		for _, ev := range result.Evidence {
			if ev.Raw == nil {
				continue
			}
			if publicIP, _ := ev.Raw["public_ip"].(string); publicIP != "" && path.PublicIP == "" {
				path.PublicIP = publicIP
			}
			if asn, _ := ev.Raw["asn"].(string); asn != "" && path.ASN == "" {
				path.ASN = asn
			}
			if org, _ := ev.Raw["organization"].(string); org != "" && path.Organization == "" {
				path.Organization = org
			}
			if cgnat, ok := ev.Raw["cgnat"].(bool); ok {
				path.CGNAT = cgnat
			}
		}
	}
	if len(path.PublicHops) > 0 || path.PublicIP != "" || path.ASN != "" {
		path.Confidence = 0.50
	}
	return path
}
