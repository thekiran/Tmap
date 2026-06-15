package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type WiFiProbe struct{}

func (WiFiProbe) Name() string            { return "wifi_probe" }
func (WiFiProbe) SafeModeAllowed() bool   { return true }
func (WiFiProbe) NormalModeAllowed() bool { return true }
func (WiFiProbe) DeepModeAllowed() bool   { return true }

func (p WiFiProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	_ = ctx
	bssid, _ := input.Metadata["wifi_bssid"].(string)
	ssid, _ := input.Metadata["wifi_ssid"].(string)
	if bssid == "" {
		return skippedResult(p.Name(), "no platform Wi-Fi adapter supplied current BSSID"), nil
	}
	ev := baseEvidence(p.Name(), bssid, model.EvidenceStrong, 0.85, "Current Wi-Fi BSSID identifies the access point radio.", map[string]any{"ssid": ssid, "bssid": bssid})
	dev := model.Device{
		ID:           "mac_" + bssid,
		MACAddresses: []string{bssid},
		Hostnames:    uniqueStrings([]string{ssid}),
		DeviceType:   model.DeviceTypeAccessPoint,
		Roles:        []model.DeviceRole{model.RoleWiFiAP},
		Confidence:   0.85,
		Evidence:     []model.Evidence{ev},
	}
	return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Evidence: []model.Evidence{ev}, Devices: []model.Device{dev}}, nil
}
