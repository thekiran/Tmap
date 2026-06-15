package discovery

import (
	"fmt"
	"strings"

	"github.com/thekiran/iad/internal/model"
)

func ClassifyDevice(device model.Device, evidence []model.Evidence) model.Device {
	if device.DeviceType == "" {
		device.DeviceType = model.DeviceTypeUnknown
	}
	text := strings.ToLower(deviceText(device, evidence))

	switch {
	case device.DeviceType == model.DeviceTypeLocalHost:
		device.AddRole(model.RoleLocalHost)
		device.Confidence = max(device.Confidence, 0.85)
	case device.DeviceType == model.DeviceTypeVirtualAdapter:
		device.Inferred = false
		device.Confidence = min(max(device.Confidence, 0.70), 0.80)
	case device.DeviceType == model.DeviceTypeISPHop:
		device.Confidence = 0.50
	case device.DeviceType == model.DeviceTypeInferredSwitch:
		device.Inferred = true
		device.AddRole(model.RoleSwitchingDevice)
		device.Confidence = clampBand(max(device.Confidence, 0.45), 0.45, 0.60)
	case device.HasRole(model.RoleDefaultGateway) || containsAny(text, "internetgatewaydevice", "wanipconnection", "nat-pmp", "pcp", "default gateway"):
		device.DeviceType = model.DeviceTypeRouter
		device.Confidence = max(device.Confidence, routerConfidence(device, text))
	case containsAny(text, "sysdescr", "bridge-mib", "lldp", "cdp", "forwarding table", "switch"):
		device.DeviceType = model.DeviceTypeManagedSwitch
		device.AddRole(model.RoleSwitchingDevice)
		device.Confidence = max(device.Confidence, 0.90)
	case device.HasRole(model.RoleWiFiAP) || containsAny(text, "bssid", "wireless ap", "access point", "802.11"):
		device.DeviceType = model.DeviceTypeAccessPoint
		device.AddRole(model.RoleWiFiAP)
		device.Confidence = max(device.Confidence, 0.85)
	case containsAny(text, "ont", "gpon", "xgs-pon", "optical"):
		device.DeviceType = model.DeviceTypeONT
		device.Confidence = max(device.Confidence, 0.80)
	case containsAny(text, "docsis", "cable modem", "dsl", "vdsl", "adsl"):
		device.DeviceType = model.DeviceTypeModem
		device.Confidence = max(device.Confidence, 0.80)
	case containsAny(text, "_ipp._tcp", "printer", "jetdirect", "port 9100") || hasPort(device, 631) || hasPort(device, 9100):
		device.DeviceType = model.DeviceTypePrinter
		device.Confidence = max(device.Confidence, 0.75)
	case containsAny(text, "synology", "qnap", "_smb._tcp", "nas") || hasPort(device, 445) || hasPort(device, 5000) || hasPort(device, 5001):
		device.DeviceType = model.DeviceTypeNAS
		device.Confidence = max(device.Confidence, 0.72)
	case containsAny(text, "rtsp", "onvif", "camera") || hasPort(device, 554):
		device.DeviceType = model.DeviceTypeCamera
		device.Confidence = max(device.Confidence, 0.70)
	case containsAny(text, "iphone", "android", "phone"):
		device.DeviceType = model.DeviceTypePhone
		device.Confidence = max(device.Confidence, 0.55)
	case containsAny(text, "laptop", "macbook", "notebook"):
		device.DeviceType = model.DeviceTypeLaptop
		device.Confidence = max(device.Confidence, 0.55)
	case containsAny(text, "desktop", "workstation"):
		device.DeviceType = model.DeviceTypeDesktop
		device.Confidence = max(device.Confidence, 0.55)
	case containsAny(text, "iot", "chromecast", "airplay", "smart"):
		device.DeviceType = model.DeviceTypeIoT
		device.Confidence = max(device.Confidence, 0.55)
	}

	if device.DeviceType == model.DeviceTypeUnknown && hasOnlyWeakEvidence(evidence) {
		device.Confidence = max(device.Confidence, 0.35)
	}
	device.Confidence = topologyDeviceConfidence(device)
	return device
}

func routerConfidence(device model.Device, text string) float64 {
	switch {
	case containsAny(text, "internetgatewaydevice", "wanipconnection", "wanpppconnection"):
		return 0.85
	case containsAny(text, "router", "gateway", "openwrt"):
		return 0.75
	case device.HasRole(model.RoleDefaultGateway):
		return 0.65
	default:
		return 0.65
	}
}

func deviceText(device model.Device, evidence []model.Evidence) string {
	var parts []string
	parts = append(parts, string(device.DeviceType), device.Vendor, device.Manufacturer, device.Model, device.OSGuess)
	parts = append(parts, device.Hostnames...)
	for _, role := range device.Roles {
		parts = append(parts, string(role))
	}
	for _, service := range device.Services {
		parts = append(parts, service.Name, service.Product)
		parts = append(parts, fmt.Sprint(service.Raw))
	}
	for _, port := range device.OpenPorts {
		parts = append(parts, fmt.Sprintf("port %d", port.Port), port.Reason)
	}
	for _, ev := range evidence {
		parts = append(parts, ev.Source, ev.Target, ev.Reason, fmt.Sprint(ev.Raw))
	}
	return strings.Join(parts, " ")
}

func containsAny(s string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(s, strings.ToLower(token)) {
			return true
		}
	}
	return false
}

func hasPort(device model.Device, port int) bool {
	for _, p := range device.OpenPorts {
		if p.Port == port && p.State == "open" {
			return true
		}
	}
	for _, s := range device.Services {
		if s.Port == port {
			return true
		}
	}
	return false
}
