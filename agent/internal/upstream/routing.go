package upstream

import "github.com/thekiran/iad/pkg/models"

// AnalyzeRouting interprets where an upstream device sits in the routing
// topology from the gathered Facts. It is pure and conservative: when the
// device is unreachable it is explicitly marked inferred, and double-NAT / CPE
// kinds require an actual signal (a second private gateway, CWMP, or an ONT/
// modem fingerprint) rather than an IP guess.
func AnalyzeRouting(f Facts) models.RoutingEvidence {
	r := models.RoutingEvidence{
		SameSubnetAsAgent: f.SameSubnetAsAgent,
		PrivateUpstream:   f.IsPrivate && !f.IsDefaultGateway,
	}
	if f.HopDistance > 0 {
		hd := f.HopDistance
		r.HopDistance = &hd
	}

	switch {
	case f.VirtualHint:
		r.Kind = "virtual_or_docker"
		r.Notes = append(r.Notes, "Address/vendor looks like a virtual or container network artifact, not a physical gateway.")
	case f.IsDefaultGateway:
		r.Kind = "default_gateway"
	case !f.ReachableICMP && !f.ReachableTCP:
		r.Kind = "unreachable_inferred"
		r.Notes = append(r.Notes, "Not directly reachable; existence is inferred from route-table / traceroute evidence only.")
	case f.DoubleNATHint && f.IsPrivate && !f.IsDefaultGateway:
		r.Kind = "double_nat_upstream"
		r.DoubleNAT = true
		r.Notes = append(r.Notes, "A second private gateway appears upstream of the default gateway — possible double NAT.")
	case f.HasCWMP || f.ONTHint || f.ModemHint:
		r.Kind = "isp_cpe"
		r.Notes = append(r.Notes, "Management/CWMP or ONT/modem fingerprint suggests ISP-provided CPE.")
	case f.IsPrivate && f.InGatewayChain && !f.IsDefaultGateway:
		r.Kind = "upstream_private_gateway"
	case f.IsPrivate:
		r.Kind = "unknown"
	default:
		r.Kind = "unknown"
	}
	return r
}
