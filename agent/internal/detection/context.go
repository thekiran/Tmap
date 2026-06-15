package detection

import (
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

// GatewayChainResult describes the leading run of private gateways seen in a
// traceroute. More than one distinct private gateway before the first public hop
// indicates the host is likely behind a router that is itself behind the ISP
// modem (double NAT / router chain).
type GatewayChainResult struct {
	Chain             []string
	DoubleNATPossible bool
	PrivateHopCount   int
}

// detectGatewayChain extracts the leading private-IP hops from a traceroute.
// Timed-out hops ("*") are tolerated within the run; the run ends at the first
// public address.
func detectGatewayChain(hops []string) GatewayChainResult {
	var chain []string
	seen := map[string]bool{}
	for _, h := range hops {
		if h == "*" {
			continue
		}
		if !isPrivateIPv4(h) {
			break // first public hop: the local chain is over
		}
		if !seen[h] {
			seen[h] = true
			chain = append(chain, h)
		}
	}
	return GatewayChainResult{
		Chain:             chain,
		DoubleNATPossible: len(chain) >= 2,
		PrivateHopCount:   len(chain),
	}
}

// buildNetworkContext assembles the factual network situation from the evidence
// bag. It is always reported, even for an Unknown verdict.
func buildNetworkContext(bag evidenceBag, matched bool) *models.NetworkContext {
	return &models.NetworkContext{
		ISP:                bag.Org,
		PublicIP:           bag.PublicIP,
		PTR:                bag.PTR,
		CGNAT:              bag.CGNAT,
		Gateway:            bag.Gateway,
		GatewayChain:       bag.GatewayChain,
		DoubleNATPossible:  bag.DoubleNATPossible,
		LocalAccess:        bag.LocalAccess,
		MainAdapter:        bag.MainAdapter,
		RouterModel:        bag.RouterModel,
		FingerprintMatched: matched,
		UPnPFound:          bag.UPnPFound,
		TR064Found:         bag.TR064Found,
		GatewayDevices:     bag.GatewayDevices,
		LikelyModemIP:      bag.LikelyModemIP,
		LikelyCPEIP:        bag.LikelyModemIP,
		WANSignals:         bag.WANSignals,
		LineProfile:        bag.LineProfile,
		AccessArchitecture: bag.AccessArchitecture,
		IPv6Context:        bag.IPv6Context,
		NATTopology:        bag.NATTopology,
		PerformanceProfile: bag.PerformanceProfile,
	}
}

func gatewayDeviceText(devices []models.GatewayDevice) string {
	var parts []string
	for _, d := range devices {
		if d.AccessConfidence <= 0 && len(d.AccessHints) == 0 {
			continue
		}
		parts = append(parts,
			d.IP,
			d.Role,
			d.Manufacturer,
			d.Model,
			d.MACVendor,
			strings.Join(d.AccessHints, " "),
		)
	}
	return strings.Join(parts, " ")
}

func wanSignalText(signals []models.WANSignal) string {
	var parts []string
	for _, s := range signals {
		if s.Strength == string(models.EvidencePhysical) || strings.EqualFold(s.Strength, "strong") {
			parts = append(parts, s.Source, s.IP, s.Type, s.Value, s.Detail)
		}
	}
	return strings.Join(parts, " ")
}
