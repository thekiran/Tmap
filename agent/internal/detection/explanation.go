package detection

import (
	"fmt"
	"strings"

	"github.com/thekiran/iad/internal/linestats"
	"github.com/thekiran/iad/pkg/models"
)

// buildExplanation produces the "why" behind a committed verdict. Each line is
// a concrete observation the user can verify. English-first by design.
func buildExplanation(primary string, bag evidenceBag, fp *Fingerprint, matched bool, fired []string) []string {
	var lines []string

	if matched && fp != nil {
		name := strings.TrimSpace(fp.Vendor + " " + fp.Model)
		supports := strings.Join(fp.Supports, ", ")
		lines = append(lines, fmt.Sprintf(
			"A modem model was identified: %s (%s) - supported technologies: %s.",
			name, fp.Category, supports))
	} else if bag.RouterModel != "" {
		lines = append(lines, fmt.Sprintf("Gateway device model: %s.", bag.RouterModel))
	}
	lines = append(lines, linestats.Summary(bag.LineProfile)...)
	lines = append(lines, wanSignalExplanation(bag)...)

	if bag.CGNAT {
		lines = append(lines, fmt.Sprintf(
			"The public IP is in the CGNAT range (%s) - carrier-grade NAT increases the likelihood of mobile/FWA/satellite.", bag.PublicIP))
	}
	if bag.PTR != "" {
		lines = append(lines, fmt.Sprintf("Reverse DNS (PTR) record: %s", bag.PTR))
	}
	if bag.Org != "" {
		lines = append(lines, fmt.Sprintf("Operator/ASN information: %s", bag.Org))
	}
	if bag.DoubleNATPossible {
		lines = append(lines, doubleNATLines(bag)...)
		lines = append(lines, gatewayDeviceExplanation(bag)...)
	}
	if bag.HasLatency {
		lines = append(lines, latencyExplain(bag.AvgMS))
	}

	if len(lines) == 0 {
		lines = append(lines, "No strong evidence was found; the estimate was made with low confidence.")
	}

	cat := models.CategoryFor(primary)
	lines = append(lines, fmt.Sprintf(
		"Result: the most likely access type is %s (category: %s); %d rule(s) matched.", primary, cat, len(fired)))
	return lines
}

// buildUncertainExplanation explains why a definite verdict was not made, while
// still naming the leading candidates.
func buildUncertainExplanation(primary string, scores map[string]float64, bag evidenceBag, matched bool, reasons []string) []string {
	var lines []string

	lines = append(lines, linestats.Summary(bag.LineProfile)...)
	if bag.Org != "" {
		lines = append(lines, fmt.Sprintf("The ISP address pool was identified (%s), but this does not prove the physical access type.", bag.Org))
	}
	if bag.PTR != "" {
		lines = append(lines, "The PTR record identifies an operator address pool, but it is not conclusive evidence of the physical access type.")
	}
	if bag.IPv6Context != nil && bag.IPv6Context.GlobalIPv6 && bag.PublicIP != "" {
		lines = append(lines, "The connection has global IPv6 and a public IPv4 address, which improves network context but does not identify the access medium.")
	}
	if !bag.UPnPFound && !matched {
		lines = append(lines, "No UPnP modem model was discovered.")
		lines = append(lines, "No TR-064 CPE data was available.")
		lines = append(lines, "Without a modem model or WAN physical-layer interface data, DSL/VDSL/Fiber cannot be distinguished conclusively.")
	}
	if bag.DoubleNATPossible {
		lines = append(lines, doubleNATLines(bag)...)
		lines = append(lines, gatewayDeviceExplanation(bag)...)
	}
	if bag.HasLatency {
		lines = append(lines, latencyExplain(bag.AvgMS))
	}
	if topTwo(categoryScores(scores)).Margin < minCategoryGap {
		lines = append(lines, "The DSL, VDSL and Fiber scores are too close to each other to classify confidently.")
	}
	if !hasStrongPhysicalEvidence(bag, matched) {
		candidates := likelyCandidateNames(scores, 2)
		if len(candidates) >= 2 {
			lines = append(lines, fmt.Sprintf("%s are among the candidates, but no strong physical-layer evidence was found, so no confident classification was made.", strings.Join(candidates, " and ")))
		} else {
			lines = append(lines, "No strong physical-layer evidence was found, so the access type remains Unknown.")
		}
	}

	if len(lines) == 0 {
		lines = append(lines, "The evidence is too weak to classify the access type.")
	}
	if tied := tiedLeadingCandidates(scores); len(tied) > 1 {
		lines = append(lines, fmt.Sprintf("Leading candidates: %s (not conclusive).", strings.Join(tied, ", ")))
	} else if primary != "" && hasStrongPhysicalEvidence(bag, matched) {
		lines = append(lines, fmt.Sprintf("Leading candidate: %s (not conclusive).", primary))
	}
	return lines
}

const doubleNATLine = "Multiple private gateways were seen on the local network; the device may be attached to an intermediate router rather than the actual modem."

func doubleNATLines(bag evidenceBag) []string {
	if len(bag.GatewayChain) < 2 {
		return []string{doubleNATLine}
	}
	upstream := bag.GatewayChain[1]
	if bag.LikelyModemIP != "" {
		upstream = bag.LikelyModemIP
	}
	return []string{
		"Two private gateways were observed on the local network.",
		fmt.Sprintf("The gateway chain indicates a possible upstream CPE: default gateway %s, upstream gateway %s.", bag.GatewayChain[0], bag.GatewayChain[1]),
		fmt.Sprintf("The actual modem is likely at %s, but it could not be verified.", upstream),
	}
}

func gatewayDeviceExplanation(bag evidenceBag) []string {
	if len(bag.GatewayDevices) == 0 {
		return nil
	}
	var lines []string
	for _, d := range bag.GatewayDevices {
		if d.Role == "default_gateway" && d.Reachable {
			lines = append(lines, fmt.Sprintf("The default gateway %s is reachable.", d.IP))
			if isGenericGatewayWebUI(d) {
				lines = append(lines, fmt.Sprintf("A generic %s web interface was seen on %s; this is not evidence of the physical access type.", genericGatewayServerName(d), d.IP))
			}
		}
	}
	for _, d := range bag.GatewayDevices {
		if d.Role == "upstream_private_gateway" && !d.Reachable {
			lines = append(lines, fmt.Sprintf("An upstream gateway %s was detected but could not be reached, so the modem model could not be verified.", d.IP))
			return lines
		}
	}
	for _, d := range bag.GatewayDevices {
		if d.IP != bag.LikelyModemIP {
			continue
		}
		if d.Manufacturer != "" || d.Model != "" || d.HTTPTitle != "" {
			name := strings.TrimSpace(strings.Join([]string{d.Manufacturer, d.Model}, " "))
			if name == "" {
				name = d.HTTPTitle
			}
			lines = append(lines, fmt.Sprintf("A %s modem interface was detected on %s.", name, d.IP))
			return lines
		}
	}
	if bag.DoubleNATPossible {
		lines = append(lines, "The upstream gateway may be reachable, but the modem model could not be verified.")
	}
	return lines
}

func wanSignalExplanation(bag evidenceBag) []string {
	if len(bag.WANSignals) == 0 {
		return nil
	}
	var lines []string
	for _, s := range bag.WANSignals {
		if s.Strength != string(models.EvidencePhysical) && !strings.EqualFold(s.Strength, "strong") {
			continue
		}
		where := s.IP
		if where == "" {
			where = "the CPE"
		}
		lines = append(lines, fmt.Sprintf("A WAN signal was received from %s via %s: %s.", where, s.Source, strings.TrimSpace(s.Value)))
	}
	return lines
}

func tiedLeadingCandidates(scores map[string]float64) []string {
	ranked := collapseParentSubtypeCandidates(rankAll(scores))
	if len(ranked) == 0 {
		return nil
	}
	top := ranked[0].Score
	var out []string
	for _, ts := range ranked {
		if ts.Score != top {
			break
		}
		out = append(out, ts.Type)
	}
	return out
}

func likelyCandidateNames(scores map[string]float64, n int) []string {
	ranked := collapseParentSubtypeCandidates(rankAll(scores))
	if len(ranked) > n {
		ranked = ranked[:n]
	}
	out := make([]string, 0, len(ranked))
	for _, ts := range ranked {
		out = append(out, ts.Type)
	}
	return out
}

func isGenericGatewayWebUI(d models.GatewayDevice) bool {
	if d.AccessConfidence > 0 || len(d.AccessHints) > 0 {
		return false
	}
	return genericGatewayServerName(d) != "" || strings.Contains(strings.ToLower(d.HTTPTitle), "login")
}

func genericGatewayServerName(d models.GatewayDevice) string {
	s := strings.ToLower(strings.TrimSpace(d.ServerHeader))
	for _, token := range genericServerTokens {
		if strings.Contains(s, token) {
			return token
		}
	}
	return "generic"
}

func latencyExplain(ms float64) string {
	switch {
	case ms <= 8:
		return fmt.Sprintf("Latency is very low (%.1f ms); this is compatible with Fiber or VDSL, but it is not conclusive by itself.", ms)
	case ms <= 25:
		return fmt.Sprintf("Latency is low (%.1f ms); this is compatible with fixed broadband but does not identify the type by itself.", ms)
	case ms <= 80:
		return fmt.Sprintf("Latency is moderate (%.1f ms); DSL or FWA is possible but not conclusive.", ms)
	case ms <= 200:
		return fmt.Sprintf("Latency is high (%.1f ms); mobile/FWA becomes more likely.", ms)
	case ms >= 500:
		return fmt.Sprintf("Latency is very high (%.1f ms); satellite becomes plausible.", ms)
	default:
		return fmt.Sprintf("Latency is %.1f ms (the 200-500 ms range is ambiguous).", ms)
	}
}
