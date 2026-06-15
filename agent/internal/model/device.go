package model

import "time"

type DeviceType string

const (
	DeviceTypeLocalHost      DeviceType = "local_host"
	DeviceTypeRouter         DeviceType = "router"
	DeviceTypeModem          DeviceType = "modem"
	DeviceTypeONT            DeviceType = "ont"
	DeviceTypeAccessPoint    DeviceType = "access_point"
	DeviceTypeSwitch         DeviceType = "switch"
	DeviceTypeManagedSwitch  DeviceType = "managed_switch"
	DeviceTypeInferredSwitch DeviceType = "inferred_switch"
	DeviceTypePrinter        DeviceType = "printer"
	DeviceTypeNAS            DeviceType = "nas"
	DeviceTypeCamera         DeviceType = "camera"
	DeviceTypePhone          DeviceType = "phone"
	DeviceTypeLaptop         DeviceType = "laptop"
	DeviceTypeDesktop        DeviceType = "desktop"
	DeviceTypeIoT            DeviceType = "iot"
	DeviceTypeVMHost         DeviceType = "vm_host"
	DeviceTypeVirtualAdapter DeviceType = "virtual_adapter"
	DeviceTypeISPHop         DeviceType = "isp_hop"
	DeviceTypeUnknown        DeviceType = "unknown"
)

type DeviceRole string

const (
	RoleDefaultGateway  DeviceRole = "default_gateway"
	RoleDHCPServer      DeviceRole = "dhcp_server"
	RoleDNSServer       DeviceRole = "dns_server"
	RoleWiFiAP          DeviceRole = "wifi_ap"
	RoleUpstreamGateway DeviceRole = "upstream_gateway"
	RoleLocalHost       DeviceRole = "local_host"
	RoleInternetEdge    DeviceRole = "internet_edge"
	RoleBridge          DeviceRole = "bridge"
	RoleSwitchingDevice DeviceRole = "switching_device"
)

type PortInfo struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	State    string `json:"state"`
	Reason   string `json:"reason,omitempty"`
}

type ServiceInfo struct {
	Name     string         `json:"name"`
	Port     int            `json:"port,omitempty"`
	Protocol string         `json:"protocol,omitempty"`
	Product  string         `json:"product,omitempty"`
	Raw      map[string]any `json:"raw,omitempty"`
}

type Device struct {
	ID           string        `json:"id"`
	IPAddresses  []string      `json:"ip_addresses,omitempty"`
	MACAddresses []string      `json:"mac_addresses,omitempty"`
	Hostnames    []string      `json:"hostnames,omitempty"`
	Vendor       string        `json:"vendor,omitempty"`
	Manufacturer string        `json:"manufacturer,omitempty"`
	Model        string        `json:"model,omitempty"`
	SerialNumber string        `json:"serial_number,omitempty"`
	OSGuess      string        `json:"os_guess,omitempty"`
	DeviceType   DeviceType    `json:"device_type"`
	Roles        []DeviceRole  `json:"roles,omitempty"`
	OpenPorts    []PortInfo    `json:"open_ports,omitempty"`
	Services     []ServiceInfo `json:"services,omitempty"`
	Confidence   float64       `json:"confidence"`
	Evidence     []Evidence    `json:"evidence,omitempty"`
	LastSeen     time.Time     `json:"last_seen,omitempty"`
	Inferred     bool          `json:"inferred"`
}

func (d Device) HasRole(role DeviceRole) bool {
	for _, r := range d.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (d *Device) AddRole(role DeviceRole) {
	if d.HasRole(role) {
		return
	}
	d.Roles = append(d.Roles, role)
}
