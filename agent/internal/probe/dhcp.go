package probe

import (
	"context"

	"github.com/thekiran/iad/internal/model"
)

type DHCPProbe struct{}

func (DHCPProbe) Name() string            { return "dhcp_probe" }
func (DHCPProbe) SafeModeAllowed() bool   { return true }
func (DHCPProbe) NormalModeAllowed() bool { return true }
func (DHCPProbe) DeepModeAllowed() bool   { return true }

func (p DHCPProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	_ = ctx
	server, _ := input.Metadata["dhcp_server"].(string)
	if server == "" {
		return skippedResult(p.Name(), "no platform DHCP adapter supplied lease context"), nil
	}
	ev := baseEvidence(p.Name(), server, model.EvidenceStrong, 0.75, "Local DHCP lease identifies this DHCP server.", map[string]any{"dhcp_server": server})
	dev := model.Device{ID: "ip_" + server, IPAddresses: []string{server}, DeviceType: model.DeviceTypeRouter, Roles: []model.DeviceRole{model.RoleDHCPServer}, Confidence: 0.75, Evidence: []model.Evidence{ev}}
	return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Evidence: []model.Evidence{ev}, Devices: []model.Device{dev}, Raw: map[string]any{"dhcp_server": server}}, nil
}
