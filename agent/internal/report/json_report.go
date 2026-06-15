// Package report serializes a ScanResult. JSON is the only format in the MVP;
// HTML/PDF are later phases that will consume the same ScanResult.
package report

import (
	"encoding/json"
	"math"
	"os"

	"github.com/thekiran/iad/pkg/models"
)

// ToJSON returns the indented JSON encoding of a scan result.
func ToJSON(r models.ScanResult) ([]byte, error) {
	rounded := roundScanResult(r)
	return json.MarshalIndent(rounded, "", "  ")
}

// WriteJSON writes the scan result as indented JSON to path.
func WriteJSON(path string, r models.ScanResult) error {
	data, err := ToJSON(r)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func roundScanResult(r models.ScanResult) models.ScanResult {
	r.Confidence = round2(r.Confidence)
	r.ClassificationConfidence = round2(r.ClassificationConfidence)
	r.ContextConfidence = round2(r.ContextConfidence)
	r.Classification.Confidence = round2(r.Classification.Confidence)
	r.ConfidenceBreakdown.Classification = round2(r.ConfidenceBreakdown.Classification)
	r.ConfidenceBreakdown.Context = round2(r.ConfidenceBreakdown.Context)
	r.ConfidenceBreakdown.Physical = round2(r.ConfidenceBreakdown.Physical)
	r.ConfidenceBreakdown.Device = round2(r.ConfidenceBreakdown.Device)
	r.ConfidenceBreakdown.Network = round2(r.ConfidenceBreakdown.Network)
	r.ConfidenceBreakdown.Performance = round2(r.ConfidenceBreakdown.Performance)
	r.ConfidenceBreakdown.Regional = round2(r.ConfidenceBreakdown.Regional)
	r.ConfidenceBreakdown.Penalty = round2(r.ConfidenceBreakdown.Penalty)
	roundEvidenceTiers(&r.EvidenceTiers)
	if r.Scores != nil {
		scores := make(map[string]float64, len(r.Scores))
		for k, v := range r.Scores {
			scores[k] = round2(v)
		}
		r.Scores = scores
	}
	r.Alternatives = append([]models.TypeScore(nil), r.Alternatives...)
	for i := range r.Alternatives {
		r.Alternatives[i].Score = round2(r.Alternatives[i].Score)
	}
	r.Candidates = append([]models.AccessCandidate(nil), r.Candidates...)
	for i := range r.Candidates {
		r.Candidates[i].Score = round2(r.Candidates[i].Score)
		r.Candidates[i].Confidence = round2(r.Candidates[i].Confidence)
		roundEvidenceItems(r.Candidates[i].SupportingEvidence)
	}
	r.ScoreContributions = append([]models.ScoreContribution(nil), r.ScoreContributions...)
	for i := range r.ScoreContributions {
		r.ScoreContributions[i].Amount = round2(r.ScoreContributions[i].Amount)
	}
	r.Evidence = append([]models.ProbeResult(nil), r.Evidence...)
	for i := range r.Evidence {
		r.Evidence[i].Confidence = round2(r.Evidence[i].Confidence)
		if r.Evidence[i].Evidence != nil {
			ev := make(map[string]any, len(r.Evidence[i].Evidence))
			for k, v := range r.Evidence[i].Evidence {
				if f, ok := v.(float64); ok && isRoundedEvidenceFloat(k) {
					ev[k] = round2(f)
					continue
				}
				ev[k] = v
			}
			if devices, ok := ev["gateway_devices"].([]models.GatewayDevice); ok {
				devices = append([]models.GatewayDevice(nil), devices...)
				for j := range devices {
					devices[j].Confidence = round2(devices[j].Confidence)
					devices[j].DeviceConfidence = round2(devices[j].DeviceConfidence)
					devices[j].AccessConfidence = round2(devices[j].AccessConfidence)
				}
				ev["gateway_devices"] = devices
			}
			if signals, ok := ev["wan_signals"].([]models.WANSignal); ok {
				signals = append([]models.WANSignal(nil), signals...)
				for j := range signals {
					signals[j].Confidence = round2(signals[j].Confidence)
				}
				ev["wan_signals"] = signals
			}
			r.Evidence[i].Evidence = ev
		}
	}
	if r.DetectedNetworkContext != nil {
		nc := *r.DetectedNetworkContext
		nc.GatewayChain = append([]string(nil), r.DetectedNetworkContext.GatewayChain...)
		nc.GatewayDevices = append([]models.GatewayDevice(nil), r.DetectedNetworkContext.GatewayDevices...)
		for i := range nc.GatewayDevices {
			nc.GatewayDevices[i].Confidence = round2(nc.GatewayDevices[i].Confidence)
			nc.GatewayDevices[i].DeviceConfidence = round2(nc.GatewayDevices[i].DeviceConfidence)
			nc.GatewayDevices[i].AccessConfidence = round2(nc.GatewayDevices[i].AccessConfidence)
		}
		nc.WANSignals = append([]models.WANSignal(nil), r.DetectedNetworkContext.WANSignals...)
		for i := range nc.WANSignals {
			nc.WANSignals[i].Confidence = round2(nc.WANSignals[i].Confidence)
		}
		if nc.PerformanceProfile != nil {
			pp := *nc.PerformanceProfile
			pp.IdleLatencyMS = round2(pp.IdleLatencyMS)
			pp.JitterMS = round2(pp.JitterMS)
			pp.PacketLossPct = round2(pp.PacketLossPct)
			pp.LoadedLatencyMS = round2(pp.LoadedLatencyMS)
			nc.PerformanceProfile = &pp
		}
		if nc.NATTopology != nil {
			nat := *nc.NATTopology
			nat.Notes = append([]string(nil), nc.NATTopology.Notes...)
			nc.NATTopology = &nat
		}
		r.DetectedNetworkContext = &nc
	}
	return r
}

func roundEvidenceTiers(t *models.EvidenceTiers) {
	t.DirectPhysical.Confidence = round2(t.DirectPhysical.Confidence)
	t.DeviceModel.Confidence = round2(t.DeviceModel.Confidence)
	t.Topology.Confidence = round2(t.Topology.Confidence)
	t.Performance.Confidence = round2(t.Performance.Confidence)
	t.Regional.Confidence = round2(t.Regional.Confidence)
	roundEvidenceItems(t.DirectPhysical.Items)
	roundEvidenceItems(t.DeviceModel.Items)
	roundEvidenceItems(t.Topology.Items)
	roundEvidenceItems(t.Performance.Items)
	roundEvidenceItems(t.Regional.Items)
}

func roundEvidenceItems(items []models.EvidenceItem) {
	for i := range items {
		items[i].Confidence = round2(items[i].Confidence)
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func isRoundedEvidenceFloat(key string) bool {
	switch key {
	case "confidence", "device_confidence", "access_confidence", "network_confidence", "performance_confidence":
		return true
	default:
		return false
	}
}
