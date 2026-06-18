package detection

import (
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

func ResolveGatewayChainState(results []models.ProbeResult, defaultGateway, agentIP string) *models.GatewayChainState {
	var state models.GatewayChainState
	state.DefaultGateway = defaultGateway

	addSource := func(source string, chain []string, confidence float64, evidenceID string) {
		chain = privateChain(chain, agentIP)
		if len(chain) == 0 {
			return
		}
		src := models.GatewayChainSource{
			Source:                    source,
			Chain:                     chain,
			PrivateHops:               gatewayHops(chain, source),
			InternalDoubleNATPossible: len(chain) >= 2,
			Confidence:                confidence,
			EvidenceIDs:               nonEmptyStrings(evidenceID),
		}
		state.Sources = append(state.Sources, src)
		state.EvidenceIDs = appendUniqueStrings(state.EvidenceIDs, src.EvidenceIDs...)
	}

	if defaultGateway != "" && isPrivateIPv4(defaultGateway) && defaultGateway != agentIP {
		addSource("route_table", []string{defaultGateway}, 0.40, "route_table:default_gateway")
	}

	for _, r := range results {
		if r.Status != models.StatusSuccess {
			continue
		}
		evidenceID := r.ProbeName
		switch r.ProbeName {
		case "gateway_probe":
			gw := firstNonEmpty(getString(r.Evidence, "gateway"), defaultGateway)
			addSource(r.ProbeName, []string{gw}, maxFloat(r.Confidence, 0.35), evidenceID)
		case "traceroute_probe":
			chain := privateChain(getStrings(r.Evidence, "hops"), agentIP)
			addSource(r.ProbeName, chain, maxFloat(r.Confidence, 0.65), evidenceID)
		case "gateway_chain_probe", "upstream_private_cpe_probe":
			addSource(r.ProbeName, getStrings(r.Evidence, "gateway_chain"), maxFloat(r.Confidence, 0.45), evidenceID)
		case "stun_pcp_nat_probe":
			if nat := getNATTopology(r.Evidence, "nat_topology"); nat != nil {
				src := models.GatewayChainSource{
					Source:                    r.ProbeName,
					InternalDoubleNATPossible: nat.InternalDoubleNATPossible || nat.DoubleNAT,
					Confidence:                maxFloat(r.Confidence, getFloat(r.Evidence, "network_confidence")),
					EvidenceIDs:               []string{evidenceID},
				}
				state.Sources = append(state.Sources, src)
				state.EvidenceIDs = appendUniqueStrings(state.EvidenceIDs, evidenceID)
			}
		}
	}

	bestIdx := -1
	for i, src := range state.Sources {
		if len(src.Chain) == 0 {
			continue
		}
		if bestIdx < 0 || gatewaySourceBetter(src, state.Sources[bestIdx]) {
			bestIdx = i
		}
	}
	if bestIdx >= 0 {
		best := state.Sources[bestIdx]
		state.Chain = append([]string{}, best.Chain...)
		state.PrivateHops = append([]models.GatewayHop{}, best.PrivateHops...)
		state.Confidence = best.Confidence
	} else if defaultGateway != "" && isPrivateIPv4(defaultGateway) && defaultGateway != agentIP {
		state.Chain = []string{defaultGateway}
		state.PrivateHops = gatewayHops(state.Chain, "route_table")
		state.Confidence = 0.35
	}

	for _, src := range state.Sources {
		if src.InternalDoubleNATPossible {
			state.InternalDoubleNATPossible = true
			state.Confidence = maxFloat(state.Confidence, src.Confidence)
		}
	}
	if len(state.Chain) >= 2 {
		state.InternalDoubleNATPossible = true
	}
	state.Conflicts = gatewayChainConflicts(state)
	if len(state.Chain) == 0 && len(state.Sources) == 0 {
		return nil
	}
	return &state
}

func gatewaySourceBetter(a, b models.GatewayChainSource) bool {
	if a.Source == "traceroute_probe" && len(a.Chain) >= len(b.Chain) {
		return true
	}
	if b.Source == "traceroute_probe" && len(b.Chain) >= len(a.Chain) {
		return false
	}
	if len(a.Chain) != len(b.Chain) {
		return len(a.Chain) > len(b.Chain)
	}
	return a.Confidence > b.Confidence
}

func privateChain(hops []string, agentIP string) []string {
	var out []string
	seen := map[string]bool{}
	for _, h := range hops {
		h = strings.TrimSpace(h)
		if h == "" || h == "*" || h == agentIP {
			continue
		}
		if !isPrivateIPv4(h) {
			break
		}
		if !seen[h] {
			seen[h] = true
			out = append(out, h)
		}
	}
	return out
}

func gatewayHops(chain []string, source string) []models.GatewayHop {
	hops := make([]models.GatewayHop, 0, len(chain))
	for i, ip := range chain {
		role := "default_gateway"
		if i > 0 {
			role = "upstream_private_gateway"
		}
		hops = append(hops, models.GatewayHop{IP: ip, Role: role, Source: source, Order: i + 1, EvidenceID: source})
	}
	return hops
}

func gatewayChainConflicts(state models.GatewayChainState) []models.DataConflict {
	var conflicts []models.DataConflict
	finalChain := strings.Join(state.Chain, ",")
	for _, src := range state.Sources {
		if src.Source == "route_table" || src.Source == "gateway_probe" {
			continue
		}
		if len(src.Chain) > 0 && finalChain != "" && strings.Join(src.Chain, ",") != finalChain {
			conflicts = append(conflicts, models.DataConflict{
				Field:      "gateway_chain_state.chain",
				Severity:   "medium",
				Effect:     "preserve_source_conflict",
				Resolution: "Canonical gateway chain prefers traceroute-backed or longer private-hop evidence.",
				Values: []models.ConflictValue{
					{Source: src.Source, Value: src.Chain},
					{Source: "canonical_gateway_chain_state", Value: state.Chain},
				},
			})
		}
		if src.Source == "stun_pcp_nat_probe" {
			continue
		}
		if src.InternalDoubleNATPossible != state.InternalDoubleNATPossible {
			conflicts = append(conflicts, models.DataConflict{
				Field:      "gateway_chain_state.internal_double_nat_possible",
				Severity:   "medium",
				Effect:     "preserve_source_conflict",
				Resolution: "Positive private-hop evidence is monotonic and is not erased by weaker negative probes.",
				Values: []models.ConflictValue{
					{Source: src.Source, Value: src.InternalDoubleNATPossible},
					{Source: "canonical_gateway_chain_state", Value: state.InternalDoubleNATPossible},
				},
			})
		}
	}
	return conflicts
}

func nonEmptyStrings(values ...string) []string {
	var out []string
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, v)
		}
	}
	return out
}
