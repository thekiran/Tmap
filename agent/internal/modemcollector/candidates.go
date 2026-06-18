package modemcollector

import "github.com/thekiran/iad/pkg/models"

type CandidateBuilder struct {
	PublicAllowed bool
}

type CandidateInput struct {
	DefaultGateway string
	AgentIP        string
	ChainState     *models.GatewayChainState
	Devices        []models.GatewayDevice
	ManualTargets  []string
}

func (b CandidateBuilder) Build(in CandidateInput) []models.GatewayDevice {
	store := NewEvidenceStore()
	add := func(ip, role, source string, confidence float64) {
		if ip == "" || isExcludedIP(ip, in.AgentIP) {
			return
		}
		if !b.PublicAllowed && !isRFC1918(ip) {
			return
		}
		store.MergeDevice(models.GatewayDevice{
			IP: ip, Role: role, ReachableState: string(models.TriUnknown),
			Confidence: confidence, DeviceConfidence: confidence,
			EvidenceIDs: sourceList(source),
		})
	}

	add(in.DefaultGateway, "default_gateway", "route_table", 0.40)
	if in.ChainState != nil {
		for _, h := range in.ChainState.PrivateHops {
			role := h.Role
			if role == "" {
				role = "upstream_private_gateway"
			}
			add(h.IP, role, h.Source, in.ChainState.Confidence)
		}
	}
	for _, ip := range in.ManualTargets {
		add(ip, "manual_cpe_candidate", "manual", 0.35)
	}
	for _, d := range in.Devices {
		if d.IP == "" || isExcludedIP(d.IP, in.AgentIP) {
			continue
		}
		if !b.PublicAllowed && !isRFC1918(d.IP) {
			continue
		}
		if d.Role == "" {
			d.Role = roleFromEvidence(d)
		}
		store.MergeDevice(d)
	}
	return prioritizeCandidates(store.Devices())
}

func roleFromEvidence(d models.GatewayDevice) string {
	switch {
	case d.TR064Found || d.TR064AuthRequired:
		return "cpe_management_endpoint"
	case d.UPnPIGDFound || d.WANCommonInterfaceFound:
		return "internet_gateway_device"
	case d.CPEModelGuess != "" || d.Model != "" || len(d.AccessHints) > 0:
		return "possible_modem_or_ont"
	default:
		return "possible_cpe"
	}
}

func prioritizeCandidates(devices []models.GatewayDevice) []models.GatewayDevice {
	for i := range devices {
		if devices[i].Role == "upstream_private_gateway" && devices[i].Confidence < 0.55 {
			devices[i].Confidence = 0.55
		}
	}
	return devices
}
