package discovery

import "github.com/thekiran/iad/internal/model"

func topologyDeviceConfidence(d model.Device) float64 {
	conf := d.Confidence
	for _, ev := range d.Evidence {
		if ev.Confidence > conf {
			conf = ev.Confidence
		}
	}
	switch d.DeviceType {
	case model.DeviceTypeManagedSwitch:
		conf = max(conf, 0.90)
	case model.DeviceTypeRouter:
		if d.HasRole(model.RoleDefaultGateway) {
			conf = max(conf, 0.65)
		}
	case model.DeviceTypeInferredSwitch:
		conf = clampBand(max(conf, 0.45), 0.45, 0.60)
		d.Inferred = true
	case model.DeviceTypeISPHop:
		conf = min(max(conf, 0.50), 0.50)
	case model.DeviceTypeUnknown:
		if conf == 0 && hasOnlyWeakEvidence(d.Evidence) {
			conf = 0.35
		}
	}
	if d.Inferred && conf > 0.60 {
		conf = 0.60
	}
	return model.Clamp01(conf)
}

func topologyEdgeConfidence(edgeType model.EdgeType, requested float64, inferred bool) float64 {
	conf := requested
	if conf == 0 {
		switch edgeType {
		case model.EdgeTypeWiFiLink:
			conf = 0.85
		case model.EdgeTypeEthernetLink:
			conf = 0.70
		case model.EdgeTypeRoutedHop, model.EdgeTypeGatewayChain:
			conf = 0.60
		case model.EdgeTypeISPRouteHop:
			conf = 0.50
		case model.EdgeTypeInferredL2:
			conf = 0.50
		case model.EdgeTypeUpstreamNAT:
			conf = 0.55
		default:
			conf = 0.35
		}
	}
	if inferred && conf > 0.60 {
		conf = 0.60
	}
	return model.Clamp01(conf)
}

func downgradeForConflicts(conf float64, conflicts []model.Conflict, deviceID string) float64 {
	for _, c := range conflicts {
		if len(c.Devices) > 0 && !containsString(c.Devices, deviceID) {
			continue
		}
		switch c.Severity {
		case model.ConflictHigh:
			if conf > 0.45 {
				conf = 0.45
			}
		case model.ConflictMedium:
			conf -= 0.10
		case model.ConflictLow:
			conf -= 0.05
		}
	}
	return model.Clamp01(conf)
}

func hasOnlyWeakEvidence(evidence []model.Evidence) bool {
	if len(evidence) == 0 {
		return true
	}
	for _, ev := range evidence {
		if ev.Strength != model.EvidenceWeak && ev.Strength != model.EvidenceNone {
			return false
		}
	}
	return true
}

func clampBand(v, low, high float64) float64 {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
