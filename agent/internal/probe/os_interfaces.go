package probe

import (
	"context"
	"net"
	"strings"

	"github.com/thekiran/iad/internal/model"
)

type OSInterfacesProbe struct{}

func (OSInterfacesProbe) Name() string            { return "os_interface_probe" }
func (OSInterfacesProbe) SafeModeAllowed() bool   { return true }
func (OSInterfacesProbe) NormalModeAllowed() bool { return true }
func (OSInterfacesProbe) DeepModeAllowed() bool   { return true }

func (p OSInterfacesProbe) Run(ctx context.Context, input model.ProbeInput) (model.ProbeResult, error) {
	select {
	case <-ctx.Done():
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusFailed}, ctx.Err()
	default:
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusFailed}, err
	}
	result := model.ProbeResult{ProbeName: p.Name(), Status: model.ProbeStatusSuccess, Raw: map[string]any{}}
	var devices []model.Device
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		var ips []string
		var rawAddrs []string
		for _, addr := range addrs {
			rawAddrs = append(rawAddrs, addr.String())
			ip, _, _ := net.ParseCIDR(addr.String())
			if ip != nil {
				ips = append(ips, ip.String())
			}
		}
		virtual := isVirtualInterface(iface.Name)
		raw := map[string]any{
			"name":      iface.Name,
			"mac":       iface.HardwareAddr.String(),
			"addresses": rawAddrs,
			"flags":     iface.Flags.String(),
			"virtual":   virtual,
		}
		result.Evidence = append(result.Evidence, baseEvidence(p.Name(), iface.Name, model.EvidenceMedium, 0.70, "OS network interface enumeration.", raw))
		if len(ips) > 0 {
			dtype := model.DeviceTypeLocalHost
			if virtual {
				dtype = model.DeviceTypeVirtualAdapter
			}
			devices = append(devices, model.Device{
				ID:           "local_" + iface.Name,
				IPAddresses:  uniqueStrings(ips),
				MACAddresses: uniqueStrings([]string{iface.HardwareAddr.String()}),
				DeviceType:   dtype,
				Roles:        []model.DeviceRole{model.RoleLocalHost},
				Confidence:   0.85,
				Evidence:     []model.Evidence{result.Evidence[len(result.Evidence)-1]},
				Inferred:     false,
			})
		}
	}
	result.Devices = devices
	result.Raw["interface_count"] = len(ifaces)
	return result, nil
}

func isVirtualInterface(name string) bool {
	n := strings.ToLower(name)
	for _, token := range []string{"virtual", "vmware", "vbox", "hyper-v", "loopback", "tun", "tap", "vpn", "docker", "wsl"} {
		if strings.Contains(n, token) {
			return true
		}
	}
	return false
}
