package detection

import "github.com/thekiran/iad/pkg/models"

func ResolveNATTopology(bag evidenceBag) *models.NATTopology {
	nat := models.NATTopology{}
	if bag.NATTopology != nil {
		nat = *bag.NATTopology
	}
	if nat.PublicIP == "" {
		nat.PublicIP = bag.PublicIP
	}
	nat.CGNAT = nat.CGNAT || bag.CGNAT
	nat.InternalDoubleNATPossible = bag.DoubleNATPossible || nat.InternalDoubleNATPossible || nat.DoubleNAT
	nat.DoubleNAT = nat.InternalDoubleNATPossible
	nat.GatewayNATControlReachable = nat.PCPReachable || nat.NATPMPReachable || nat.GatewayNATControlReachable
	if nat.PublicIP != "" && nat.STUNPublicIP != "" {
		nat.PublicIPMatches = nat.PublicIP == nat.STUNPublicIP
		nat.ExternalPublicIPConsistent = nat.PublicIPMatches
		if nat.PublicIPMatches {
			nat.Notes = appendUniqueStrings(nat.Notes, "STUN public IP matches the public IP probe; this does not disprove internal double NAT.")
		}
	} else if nat.PublicIP != "" {
		nat.ExternalPublicIPConsistent = true
	}
	if nat.GatewayNATControlReachable {
		nat.Notes = appendUniqueStrings(nat.Notes, "PCP/NAT-PMP reachability is NAT-control context only; it does not prove the access type.")
	}

	switch {
	case nat.CGNAT && nat.InternalDoubleNATPossible:
		nat.Topology = "cgnat_possible"
		nat.Notes = appendUniqueStrings(nat.Notes, "CGNAT and internal double NAT can coexist.")
	case nat.CGNAT:
		nat.Topology = "cgnat_possible"
	case bag.IPv6Context != nil && bag.IPv6Context.GlobalIPv6 && nat.InternalDoubleNATPossible:
		nat.Topology = "dual_stack_with_internal_double_nat"
	case bag.IPv6Context != nil && bag.IPv6Context.GlobalIPv6 && nat.PublicIP != "":
		nat.Topology = "dual_stack_public_ipv4"
	case bag.IPv6Context != nil && bag.IPv6Context.GlobalIPv6 && nat.PublicIP == "":
		nat.Topology = "ipv6_only_transition"
	case nat.InternalDoubleNATPossible && nat.PublicIP != "":
		nat.Topology = "internal_double_nat_public_ipv4"
	case nat.PublicIP != "":
		nat.Topology = "single_private_gateway_public_ipv4"
	default:
		nat.Topology = "unknown"
	}
	return &nat
}
