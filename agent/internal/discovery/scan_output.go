package discovery

import (
	"fmt"
	"time"

	"github.com/thekiran/iad/internal/model"
)

func BuildScanOutput(scanID string, mode model.ScanMode, scope model.ScanScope, results []model.ProbeResult, ctx model.NetworkContext, createdAt time.Time) model.ScanOutput {
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	devices := MergeDevices(results)
	conflicts := DetectConflicts(devices, results)
	for i := range devices {
		devices[i].Confidence = downgradeForConflicts(devices[i].Confidence, conflicts, devices[i].ID)
	}
	evidence := collectEvidence(results, devices)
	topology := BuildTopology(devices, evidence)
	ispPath := BuildISPPath(results, ctx)
	localHost := findLocalHost(devices)
	summary := summarize(devices)
	dataQuality := model.DataQuality{HasConflicts: len(conflicts) > 0, Conflicts: conflicts}
	next := NextBestProbes(devices, topology, conflicts, mode)
	ui := buildUI(summary, dataQuality, ispPath)
	return model.ScanOutput{
		ScanID:         scanID,
		CreatedAt:      createdAt,
		Mode:           mode,
		Scope:          scope,
		LocalHost:      localHost,
		NetworkContext: ctx,
		Devices:        devices,
		Topology:       topology,
		ISPPath:        ispPath,
		Evidence:       evidence,
		DataQuality:    dataQuality,
		Summary:        summary,
		NextBestProbes: next,
		UI:             ui,
	}
}

func collectEvidence(results []model.ProbeResult, devices []model.Device) []model.Evidence {
	var evidence []model.Evidence
	for _, result := range results {
		evidence = mergeEvidence(evidence, result.Evidence)
	}
	for _, device := range devices {
		evidence = mergeEvidence(evidence, device.Evidence)
	}
	return evidence
}

func findLocalHost(devices []model.Device) model.Device {
	for _, d := range devices {
		if d.DeviceType == model.DeviceTypeLocalHost || d.HasRole(model.RoleLocalHost) {
			return d
		}
	}
	return model.Device{ID: "local_host", DeviceType: model.DeviceTypeLocalHost, Confidence: 0}
}

func summarize(devices []model.Device) model.Summary {
	var s model.Summary
	s.DeviceCount = len(devices)
	for _, d := range devices {
		if d.Inferred {
			s.InferredDevices++
		} else {
			s.ConfirmedDevices++
		}
		switch d.DeviceType {
		case model.DeviceTypeRouter:
			s.Routers++
		case model.DeviceTypeSwitch, model.DeviceTypeManagedSwitch, model.DeviceTypeInferredSwitch:
			s.Switches++
		case model.DeviceTypeAccessPoint:
			s.AccessPoints++
		case model.DeviceTypeUnknown, "":
			s.UnknownDevices++
		}
	}
	return s
}

func buildUI(summary model.Summary, dataQuality model.DataQuality, isp model.ISPPath) model.UIOutput {
	headline := fmt.Sprintf("%d local devices mapped", summary.DeviceCount)
	if summary.DeviceCount == 0 {
		headline = "No local devices mapped"
	}
	warnings := []string{"Only authorized private/local ranges are scanned. Public hops are route observations, not confirmed ISP devices."}
	if isp.Warning != "" {
		warnings = append(warnings, isp.Warning)
	}
	if dataQuality.HasConflicts {
		warnings = append(warnings, "Some observations conflict; affected confidence scores were downgraded.")
	}
	return model.UIOutput{
		Headline: headline,
		Summary:  fmt.Sprintf("%d confirmed, %d inferred, %d unknown.", summary.ConfirmedDevices, summary.InferredDevices, summary.UnknownDevices),
		Warnings: warnings,
		Badges:   []string{"read-only", "private-scope", "no-bruteforce"},
	}
}
