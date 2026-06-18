package detection

import (
	"fmt"
	"strings"
	"time"

	evnorm "github.com/thekiran/iad/internal/detection/evidence"
	"github.com/thekiran/iad/internal/linestats"
	"github.com/thekiran/iad/internal/modemcollector"
	"github.com/thekiran/iad/internal/scoring"
	"github.com/thekiran/iad/pkg/models"
)

// Engine holds the loaded rules and fingerprint database and runs the detection
// pipeline over a set of probe results.
type Engine struct {
	Rules        []scoring.Rule
	Fingerprints []Fingerprint
}

// Rule file names loaded from the rules directory.
const (
	fileAccessRules   = "access_rules.yaml"
	fileInterfaceRule = "interface_patterns.yaml"
	fileISPPatterns   = "isp_patterns.yaml"
	fileFingerprints  = "modem_fingerprints.yaml"
)

// NewEngine loads the rules and fingerprints from rulesDir.
func NewEngine(rulesDir string) (*Engine, error) {
	rules, err := scoring.LoadRules(rulesDir, fileAccessRules, fileInterfaceRule, fileISPPatterns)
	if err != nil {
		return nil, err
	}
	fps, err := LoadFingerprints(rulesDir, fileFingerprints)
	if err != nil {
		return nil, err
	}
	return &Engine{Rules: rules, Fingerprints: fps}, nil
}

// evidenceBag is the normalized view of all probe evidence the engine reasons
// over. It feeds scoring/matching, the decision layer, and the network context.
type evidenceBag struct {
	RouterModel string
	Text        string
	Interfaces  []string
	Hints       []string
	Sources     int

	PublicIP   string
	CGNAT      bool
	PTR        string
	Org        string
	AvgMS      float64
	JitterMS   float64
	HasLatency bool
	Gateway    string
	AgentIP    string
	Active     []string

	UPnPFound         bool
	MainAdapter       string
	LocalAccess       string
	Hops              []string
	GatewayChain      []string
	GatewayChainState *models.GatewayChainState
	DoubleNATPossible bool
	GatewayDevices    []models.GatewayDevice
	LikelyModemIP     string
	GatewayDeviceText string
	WANSignals        []models.WANSignal
	WANSignalText     string
	StrongAccessHints []string
	TR064Found        bool

	// Raw CPE physical-layer key/values + text gathered from authorized telemetry,
	// and the LineProfile parsed from them (VDSL2 profile, DOCSIS, PON optical).
	CPEKV               map[string]string
	CPEText             string
	CPESource           string
	LineProfile         *models.LineProfile
	AccessArchitecture  models.AccessArchitecture
	IPv6Context         *models.IPv6Context
	NATTopology         *models.NATTopology
	PerformanceProfile  *models.PerformanceProfile
	PhysicalEvidence    float64
	DeviceEvidence      float64
	NetworkEvidence     float64
	PerformanceEvidence float64
	NormalizedEvidence  []evnorm.NormalizedEvidence
}

// Analyze runs the full pipeline and returns the verdict plus the raw evidence.
func (e *Engine) Analyze(in models.ScanInput, results []models.ProbeResult) models.ScanResult {
	now := time.Now()
	bag := mergeProbeResults(results)
	if len(bag.GatewayChain) == 0 && bag.Gateway != "" {
		bag.GatewayChain = []string{bag.Gateway}
	}

	// Merge duplicate gateway devices by IP and reconcile all NAT signals into a
	// single consistent topology (spec §A.2, §A.3).
	bag.GatewayDevices = MergeGatewayDevices(bag.GatewayDevices)
	bag.NATTopology = ResolveNATTopology(bag)
	if hasUsefulNATContext(bag.NATTopology) {
		bag.DoubleNATPossible = bag.NATTopology.InternalDoubleNATPossible
		bag.AccessArchitecture.NATTopology = bag.NATTopology.Topology
	} else {
		bag.NATTopology = nil
		bag.AccessArchitecture.NATTopology = ""
	}

	board := scoring.NewBoard()

	// Fingerprint match against the combined model/text — the strongest signal.
	fp, matched := MatchFingerprint(e.Fingerprints, bag.RouterModel+" "+bag.Text+" "+bag.GatewayDeviceText+" "+bag.WANSignalText)
	if matched {
		board.AddHints(fp.AccessHints, scoring.WeightFingerprintHint)
		board.AddHints(fp.Supports, scoring.WeightFingerprintSupp)
	}

	// Probe hints (CGNAT, PTR keywords, cellular adapter, ...).
	board.AddHints(bag.Hints, scoring.WeightProbeHint)
	board.AddHints(bag.StrongAccessHints, scoring.WeightStrongAccessHint)

	// Latency: small, supportive, banded contribution only.
	applyLatencySignal(board, bag)

	// YAML rules. Their hint conditions see both probe and fingerprint hints.
	matchHints := bag.Hints
	if matched {
		matchHints = append(append([]string{}, bag.Hints...), fp.AccessHints...)
	}
	matchHints = append(matchHints, bag.StrongAccessHints...)
	ctx := scoring.MatchContext{
		RouterModel: bag.RouterModel,
		Text:        strings.TrimSpace(bag.Text + " " + bag.GatewayDeviceText + " " + bag.WANSignalText),
		Interfaces:  bag.Interfaces,
		Hints:       matchHints,
	}
	for _, r := range e.Rules {
		if r.Matches(ctx) {
			board.Add(r.Then.AddScore)
			board.MarkFired(r.ID)
		}
	}

	scores := board.Normalize()
	ranked := collapseParentSubtypeCandidates(rankAll(scores))

	result := models.ScanResult{
		ScanID:                 fmt.Sprintf("scan_%s", now.Format("20060102_150405")),
		CreatedAt:              now,
		Status:                 "completed",
		Mode:                   in.Mode,
		Scores:                 scores,
		Evidence:               results,
		DetectedNetworkContext: buildNetworkContext(bag, matched),
	}
	result.ConfidenceBreakdown = computeConfidenceBreakdown(scores, bag, matched)
	result.EvidenceTiers = buildEvidenceTiers(bag, matched)
	result.Conflicts = detectConflicts(results, bag, scores)
	if result.Conflicts == nil {
		result.Conflicts = []models.DataConflict{}
	}
	result.DataQuality = models.DataQuality{HasConflicts: len(result.Conflicts) > 0, Conflicts: result.Conflicts}
	if len(result.Conflicts) > 0 {
		result.ConfidenceBreakdown.Penalty = maxFloat(result.ConfidenceBreakdown.Penalty, conflictPenalty(result.Conflicts))
	}

	// Context vs classification split, score audit, candidates, evidence summary
	// (spec §A.4, §A.5, §C). These are independent of the final verdict.
	summary := buildEvidenceStrengthSummary(bag, matched)
	if result.DetectedNetworkContext != nil {
		result.DetectedNetworkContext.EvidenceStrength = summary
	}
	result.ScoreContributions = buildContributions(bag, fp, matched, e.Rules, board.Fired())
	result.Candidates = buildCandidates(scores, bag, matched)
	ctxConf := contextConfidence(bag)
	result.ContextConfidence = ctxConf
	result.ConfidenceBreakdown.Context = ctxConf

	// No scoreable physical evidence -> honest Unknown with zero classification
	// confidence. Context evidence may still be present and is reported above.
	if len(ranked) == 0 {
		result.PrimaryType = "Unknown"
		result.Category = models.CatUnknown
		result.Confidence = 0
		result.ClassificationConfidence = 0
		result.ConfidenceBreakdown.Classification = 0
		result.DecisionQuality = "low"
		result.UncertaintyReasons = []string{"No scoreable physical access evidence was collected."}
		result.Explanation = buildUncertainExplanation("", scores, bag, matched, result.UncertaintyReasons)
		result.NextBestProbes = nextBestProbes(bag, matched, result.Conflicts, scores)
		populateOutputContract(&result, bag, matched, "")
		attachModemCollection(&result)
		return result
	}

	leading := ranked[0].Type
	confidence := computeClassificationConfidence(scores, bag, matched, result.Conflicts)

	direct := hasDirectPhysicalEvidence(bag)
	device := hasDeviceModelEvidence(bag, matched)
	result.Confidence = confidence
	result.ClassificationConfidence = confidence
	result.ConfidenceBreakdown.Classification = confidence
	catPair := topTwo(categoryScores(scores))
	result.DecisionQuality = decisionQuality(confidence, catPair.Margin, catPair.FirstScore, direct, device)

	unknown, reasons := shouldReturnUnknown(scores, confidence, bag, matched)
	reasons = append(reasons, conflictReasons(scores, bag)...)
	if hasHighSeverityConflict(result.Conflicts) {
		reasons = append(reasons, "Conflicting probe results reduced classification confidence.")
		if !direct || hasWANClassificationConflict(result.Conflicts) {
			unknown = true
		}
	}
	if unknown {
		// Keep the candidates visible (including the leader) but do not commit.
		result.PrimaryType = "Unknown"
		result.Category = models.CatUnknown
		result.Alternatives = capN(ranked, 4)
		result.UncertaintyReasons = reasons
		result.NextBestProbes = nextBestProbes(bag, matched, result.Conflicts, scores)
		result.Explanation = buildUncertainExplanation(leading, scores, bag, matched, reasons)
		populateOutputContract(&result, bag, matched, leading)
		attachModemCollection(&result)
		return result
	}

	result.PrimaryType = leading
	result.Category = models.CategoryFor(leading)
	result.Alternatives = capN(ranked[1:], 3)
	result.Explanation = buildExplanation(leading, bag, fp, matched, board.Fired())
	if !direct || result.Confidence < highConfidence {
		result.NextBestProbes = nextBestProbes(bag, matched, result.Conflicts, scores)
	}
	populateOutputContract(&result, bag, matched, leading)
	attachModemCollection(&result)
	return result
}

func attachModemCollection(result *models.ScanResult) {
	collection := modemcollector.Build(modemcollector.BuildInput{Result: *result})
	result.ModemCollection = &collection
}

// applyLatencySignal adds the small, banded latency contribution. Latency never
// decides a verdict alone; these weights only nudge.
func applyLatencySignal(board *scoring.Board, bag evidenceBag) {
	if !bag.HasLatency {
		return
	}
	switch ms := bag.AvgMS; {
	case ms <= 8:
		board.Add(map[string]float64{models.TypeFiber: scoring.LatLowFiber, models.TypeVDSL: scoring.LatLowVDSL})
	case ms <= 25:
		board.Add(map[string]float64{
			models.TypeDSL: scoring.LatMidGeneric, models.TypeVDSL: scoring.LatMidGeneric,
			models.TypeFiber: scoring.LatMidGeneric, models.TypeCable: scoring.LatMidGeneric,
		})
	case ms <= 80:
		board.Add(map[string]float64{models.TypeDSL: scoring.LatDSLFWA, models.TypeFWA: scoring.LatDSLFWA})
	case ms <= 200:
		board.Add(map[string]float64{models.TypeMobile: scoring.LatMobile, models.TypeFWA: scoring.LatMobile})
	case ms >= 500:
		board.Add(map[string]float64{models.TypeSatellite: scoring.LatSatellite})
	}
}

// mergeProbeResults folds probe results into a single normalized evidence bag.
// It keeps direct WAN evidence separate from device, topology, performance, and
// regional context so the classifier can apply hard confidence caps later.
func mergeProbeResults(results []models.ProbeResult) evidenceBag {
	return aggregate(results)
}

// aggregate folds the probe results into a single normalized evidence bag.
func aggregate(results []models.ProbeResult) evidenceBag {
	var bag evidenceBag
	var textParts []string

	for _, r := range results {
		if r.Status != models.StatusSuccess {
			continue
		}
		contributed := len(r.Hints) > 0
		for _, h := range r.Hints {
			bag.Hints = appendUnique(bag.Hints, h)
		}
		bag.NormalizedEvidence = append(bag.NormalizedEvidence, normalizeProbeResult(r)...)

		switch r.ProbeName {
		case "upnp_probe":
			bag.UPnPFound = getBool(r.Evidence, "found")
			model := getString(r.Evidence, "router_model")
			if model != "" {
				bag.RouterModel = NormalizeModel(model)
				contributed = true
			}
			if t := getString(r.Evidence, "router_text"); t != "" {
				textParts = append(textParts, t)
			}
		case "adapter_probe":
			bag.Active = getStrings(r.Evidence, "active")
			adapters := toAdapterInfos(r.Evidence["adapters"])
			main := pickMainAdapter(adapters)
			bag.MainAdapter = main.Name
			bag.LocalAccess = main.Access
			bag.AgentIP = main.IP
			for _, a := range adapters {
				bag.Interfaces = appendUnique(bag.Interfaces, a.Name)
			}
		case "public_ip_probe":
			bag.PublicIP = getString(r.Evidence, "public_ip")
			bag.CGNAT = getBool(r.Evidence, "cgnat")
		case "asn_probe":
			bag.PTR = getString(r.Evidence, "ptr")
			bag.Org = getString(r.Evidence, "org")
			if bag.PTR != "" || bag.Org != "" {
				textParts = append(textParts, bag.PTR, bag.Org)
				contributed = true
			}
		case "latency_probe":
			bag.AvgMS = getFloat(r.Evidence, "avg_ms")
			bag.JitterMS = getFloat(r.Evidence, "jitter_ms")
			bag.HasLatency = bag.AvgMS > 0
		case "gateway_probe":
			bag.Gateway = getString(r.Evidence, "gateway")
		case "traceroute_probe":
			bag.Hops = getStrings(r.Evidence, "hops")
			chain := detectGatewayChain(bag.Hops)
			bag.GatewayChain = chain.Chain
			bag.DoubleNATPossible = chain.DoubleNATPossible
			if chain.DoubleNATPossible {
				contributed = true
			}
		case "gateway_chain_probe":
			if chain := getStrings(r.Evidence, "gateway_chain"); len(chain) > 0 {
				bag.GatewayChain = chain
				bag.DoubleNATPossible = getBool(r.Evidence, "double_nat_possible")
				contributed = true
			}
			bag.LikelyModemIP = getString(r.Evidence, "likely_modem_ip")
			bag.GatewayDevices = getGatewayDevices(r.Evidence, "gateway_devices")
			if len(bag.GatewayDevices) > 0 {
				bag.GatewayDeviceText = gatewayDeviceText(bag.GatewayDevices)
				for _, d := range bag.GatewayDevices {
					bag.DeviceEvidence = maxFloat(bag.DeviceEvidence, d.DeviceConfidence)
					if gatewayDeviceHasDirectPhysicalEvidence(d) && d.AccessConfidence > 0 {
						bag.PhysicalEvidence = maxFloat(bag.PhysicalEvidence, d.AccessConfidence)
						for _, h := range gatewayDevicePhysicalHints(d) {
							bag.StrongAccessHints = appendUnique(bag.StrongAccessHints, h)
						}
					} else if d.AccessConfidence > 0 {
						bag.DeviceEvidence = maxFloat(bag.DeviceEvidence, d.AccessConfidence)
					}
				}
				contributed = true
			}
		case "upnp_igd_probe", "upnp_igd_deep_probe", "upnp_igd_deep_probe_v2", "tr064_probe", "tr064_probe_v2", "snmp_probe_opt_in", "tr181_interface_stack_probe":
			if getBool(r.Evidence, "tr064_found") {
				bag.TR064Found = true
			}
			if getBool(r.Evidence, "igd_wan_common_found") {
				bag.UPnPFound = true
			}
			model := strings.TrimSpace(getString(r.Evidence, "manufacturer") + " " + getString(r.Evidence, "model"))
			if model != "" {
				bag.RouterModel = NormalizeModel(model)
				textParts = append(textParts, model)
				contributed = true
			}
			bag.DeviceEvidence = maxFloat(bag.DeviceEvidence, getFloat(r.Evidence, "device_confidence"))
			directPhysical := directProbeHasPhysicalEvidence(r)
			if directPhysical {
				bag.PhysicalEvidence = maxFloat(bag.PhysicalEvidence, getFloat(r.Evidence, "access_confidence"))
			} else if getFloat(r.Evidence, "access_confidence") > 0 {
				bag.DeviceEvidence = maxFloat(bag.DeviceEvidence, getFloat(r.Evidence, "access_confidence"))
			}
			if signals := getWANSignals(r.Evidence, "wan_signals"); len(signals) > 0 {
				bag.WANSignals = append(bag.WANSignals, signals...)
				bag.WANSignalText = strings.TrimSpace(bag.WANSignalText + " " + wanSignalText(signals))
				contributed = true
			}
			if kv := getStringMap(r.Evidence, "cpe_kv"); len(kv) > 0 {
				if bag.CPEKV == nil {
					bag.CPEKV = map[string]string{}
				}
				for k, v := range kv {
					if _, ok := bag.CPEKV[k]; !ok || strings.TrimSpace(v) != "" {
						bag.CPEKV[k] = v
					}
				}
				bag.CPESource = r.ProbeName
				contributed = true
			}
			if t := getString(r.Evidence, "cpe_text"); t != "" {
				bag.CPEText = strings.TrimSpace(bag.CPEText + " " + t)
			}
			if directPhysical && (getBool(r.Evidence, "strong_access_evidence") || getFloat(r.Evidence, "access_confidence") > 0) {
				for _, h := range r.Hints {
					bag.StrongAccessHints = appendUnique(bag.StrongAccessHints, h)
				}
				contributed = true
			}
		case "http_fingerprint_v2", "http_fingerprint_v3", "upstream_private_cpe_probe":
			devices := getGatewayDevices(r.Evidence, "gateway_devices")
			if len(devices) > 0 {
				bag.GatewayDevices = append(bag.GatewayDevices, devices...)
				bag.GatewayDeviceText = strings.TrimSpace(bag.GatewayDeviceText + " " + gatewayDeviceText(devices))
				for _, d := range devices {
					bag.DeviceEvidence = maxFloat(bag.DeviceEvidence, d.DeviceConfidence)
					if gatewayDeviceHasDirectPhysicalEvidence(d) && d.AccessConfidence > 0 {
						bag.PhysicalEvidence = maxFloat(bag.PhysicalEvidence, d.AccessConfidence)
						for _, h := range gatewayDevicePhysicalHints(d) {
							bag.StrongAccessHints = appendUnique(bag.StrongAccessHints, h)
						}
					} else if d.AccessConfidence > 0 {
						bag.DeviceEvidence = maxFloat(bag.DeviceEvidence, d.AccessConfidence)
					}
				}
				contributed = true
			}
		case "ipv6_transition_probe", "ipv6_transition_probe_v2":
			if ctx := getIPv6Context(r.Evidence, "ipv6_context"); ctx != nil {
				bag.IPv6Context = ctx
				bag.AccessArchitecture.IPArchitecture = getString(r.Evidence, "ip_architecture")
				bag.NetworkEvidence = maxFloat(bag.NetworkEvidence, getFloat(r.Evidence, "network_confidence"))
				contributed = true
			}
		case "stun_pcp_nat_probe":
			if nat := getNATTopology(r.Evidence, "nat_topology"); nat != nil {
				bag.NATTopology = nat
				bag.AccessArchitecture.NATTopology = nat.Topology
				bag.NetworkEvidence = maxFloat(bag.NetworkEvidence, getFloat(r.Evidence, "network_confidence"))
				contributed = true
			}
		case "os_interface_probe_v2", "os_interface_probe_v3":
			if arch := getAccessArchitecture(r.Evidence, "access_architecture"); arch != nil {
				if arch.LocalMedium != "" {
					bag.AccessArchitecture.LocalMedium = arch.LocalMedium
					bag.LocalAccess = arch.LocalMedium
				}
				bag.DeviceEvidence = maxFloat(bag.DeviceEvidence, getFloat(r.Evidence, "device_confidence"))
				if getBool(r.Evidence, "local_cellular_evidence") {
					bag.PhysicalEvidence = maxFloat(bag.PhysicalEvidence, getFloat(r.Evidence, "access_confidence"))
				}
				contributed = true
			}
		case "lldp_cdp_passive_probe", "lldp_cdp_passive_probe_v2":
			bag.NetworkEvidence = maxFloat(bag.NetworkEvidence, getFloat(r.Evidence, "network_confidence"))
			if getBool(r.Evidence, "strong_access_evidence") || getFloat(r.Evidence, "access_confidence") > 0 {
				bag.PhysicalEvidence = maxFloat(bag.PhysicalEvidence, getFloat(r.Evidence, "access_confidence"))
				for _, h := range r.Hints {
					bag.StrongAccessHints = appendUnique(bag.StrongAccessHints, h)
				}
			}
			if len(r.Hints) > 0 || getFloat(r.Evidence, "network_confidence") > 0 {
				contributed = true
			}
		case "performance_profile_probe", "performance_profile_probe_v2":
			if pp := getPerformanceProfile(r.Evidence, "performance_profile"); pp != nil {
				bag.PerformanceProfile = pp
				bag.PerformanceEvidence = maxFloat(bag.PerformanceEvidence, getFloat(r.Evidence, "performance_confidence"))
				contributed = true
			}
		case "gateway_reachability_diagnostics_probe":
			bag.NetworkEvidence = maxFloat(bag.NetworkEvidence, getFloat(r.Evidence, "network_confidence"))
			if getBool(r.Evidence, "route_present") {
				contributed = true
			}
		}

		if contributed {
			bag.Sources++
		}
	}

	if bag.AccessArchitecture.LocalMedium == "" {
		bag.AccessArchitecture.LocalMedium = bag.LocalAccess
	}
	if chainState := ResolveGatewayChainState(results, bag.Gateway, bag.AgentIP); chainState != nil {
		bag.GatewayChainState = chainState
		if len(chainState.Chain) > 0 {
			bag.GatewayChain = append([]string{}, chainState.Chain...)
		}
		if chainState.InternalDoubleNATPossible {
			bag.DoubleNATPossible = true
		}
	}
	if bag.AccessArchitecture.WANMedium == "" {
		bag.AccessArchitecture.WANMedium = firstWANMedium(bag.WANSignals, bag.StrongAccessHints)
	}
	if bag.LikelyModemIP != "" {
		bag.AccessArchitecture.LikelyCPERole = "possible_modem"
	}
	if bag.NATTopology != nil {
		bag.NATTopology.InternalDoubleNATPossible = bag.NATTopology.InternalDoubleNATPossible || bag.NATTopology.DoubleNAT || bag.DoubleNATPossible
		bag.NATTopology.DoubleNAT = bag.NATTopology.InternalDoubleNATPossible
	} else if bag.DoubleNATPossible {
		bag.NATTopology = &models.NATTopology{DoubleNAT: true, InternalDoubleNATPossible: true, Topology: "double_nat_possible"}
		bag.AccessArchitecture.NATTopology = "double_nat_possible"
	}
	bag.Text = strings.TrimSpace(strings.Join(append([]string{bag.RouterModel}, textParts...), " "))

	// Parse the fine-grained physical-layer line profile (VDSL2 profile, DSL line
	// stats, DOCSIS, PON optical) from authorized CPE telemetry. When present it is
	// strong physical evidence: feed its access-type keys in so the engine can
	// commit to a subtype out of Unknown.
	lineText := strings.TrimSpace(strings.Join([]string{bag.WANSignalText, bag.CPEText, bag.RouterModel}, " "))
	if lp := linestats.Parse(bag.CPEKV, lineText, bag.CPESource); lp != nil {
		bag.LineProfile = lp
		for _, h := range linestats.AccessHints(lp) {
			bag.StrongAccessHints = appendUnique(bag.StrongAccessHints, h)
		}
		bag.PhysicalEvidence = maxFloat(bag.PhysicalEvidence, lp.Confidence)
	}
	return bag
}

// capN returns at most n entries from ts.
func capN(ts []models.TypeScore, n int) []models.TypeScore {
	if len(ts) > n {
		return ts[:n]
	}
	return ts
}

func appendUnique(s []string, v string) []string {
	if v == "" {
		return s
	}
	for _, e := range s {
		if e == v {
			return s
		}
	}
	return append(s, v)
}

func firstWANMedium(signals []models.WANSignal, hints []string) string {
	for _, s := range signals {
		if s.Strength == string(models.EvidencePhysical) || strings.EqualFold(s.Strength, "strong") {
			if s.Value != "" {
				return s.Value
			}
		}
	}
	if len(hints) > 0 {
		return strings.Join(hints, ",")
	}
	return ""
}
