package unit

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/thekiran/iad/internal/discovery"
	"github.com/thekiran/iad/internal/model"
	"github.com/thekiran/iad/internal/output"
	"github.com/thekiran/iad/internal/probe"
)

func TestDeviceClassificationRules(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name     string
		device   model.Device
		evidence []model.Evidence
		wantType model.DeviceType
		minConf  float64
		inferred bool
		maxConf  float64
		wantRole model.DeviceRole
	}{
		{
			name:     "default gateway classified as router",
			device:   model.Device{ID: "gw", IPAddresses: []string{"192.168.1.1"}, Roles: []model.DeviceRole{model.RoleDefaultGateway}},
			wantType: model.DeviceTypeRouter,
			minConf:  0.65,
		},
		{
			name: "UPnP IGD classified as router",
			device: model.Device{ID: "igd", IPAddresses: []string{"192.168.1.1"}, Evidence: []model.Evidence{
				ev("ssdp_upnp_probe", "igd", model.EvidenceStrong, 0.85, "UPnP InternetGatewayDevice", map[string]any{"device_type": "InternetGatewayDevice"}, now),
			}},
			wantType: model.DeviceTypeRouter,
			minConf:  0.85,
		},
		{
			name: "SNMP sysDescr switch classified as managed switch",
			device: model.Device{ID: "sw", IPAddresses: []string{"192.168.1.2"}, Evidence: []model.Evidence{
				ev("snmp_optin_probe", "sw", model.EvidenceStrong, 0.90, "SNMP sysDescr switch bridge-MIB", map[string]any{"sysDescr": "managed switch", "bridge_mib": true}, now),
			}},
			wantType: model.DeviceTypeManagedSwitch,
			minConf:  0.90,
			wantRole: model.RoleSwitchingDevice,
		},
		{
			name: "ARP-only host remains low confidence",
			device: model.Device{ID: "host", IPAddresses: []string{"192.168.1.50"}, Evidence: []model.Evidence{
				ev("arp_neighbor_probe", "host", model.EvidenceWeak, 0.35, "ARP table observed host", nil, now),
			}},
			wantType: model.DeviceTypeUnknown,
			minConf:  0.35,
			maxConf:  0.35,
		},
		{
			name: "mDNS printer classified as printer",
			device: model.Device{ID: "printer", Services: []model.ServiceInfo{{Name: "_ipp._tcp", Port: 631}}, Evidence: []model.Evidence{
				ev("mdns_probe", "printer", model.EvidenceMedium, 0.75, "_ipp._tcp printer service", nil, now),
			}},
			wantType: model.DeviceTypePrinter,
			minConf:  0.75,
		},
		{
			name: "NAS detected from mDNS HTTP SMB hints",
			device: model.Device{ID: "nas", OpenPorts: []model.PortInfo{{Port: 445, Protocol: "tcp", State: "open"}}, Services: []model.ServiceInfo{{Name: "_smb._tcp", Port: 445}}, Evidence: []model.Evidence{
				ev("mdns_probe", "nas", model.EvidenceMedium, 0.72, "NAS SMB service", nil, now),
			}},
			wantType: model.DeviceTypeNAS,
			minConf:  0.72,
		},
		{
			name:     "Wi-Fi BSSID creates AP node",
			device:   model.Device{ID: "ap", MACAddresses: []string{"aa:bb:cc:dd:ee:ff"}, Roles: []model.DeviceRole{model.RoleWiFiAP}},
			wantType: model.DeviceTypeAccessPoint,
			minConf:  0.85,
			wantRole: model.RoleWiFiAP,
		},
		{
			name:     "unmanaged switch is never marked confirmed",
			device:   model.Device{ID: "inferred_sw", DeviceType: model.DeviceTypeInferredSwitch, Inferred: true, Confidence: 0.95},
			wantType: model.DeviceTypeInferredSwitch,
			minConf:  0.45,
			maxConf:  0.60,
			inferred: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := discovery.ClassifyDevice(tt.device, tt.device.Evidence)
			if got.DeviceType != tt.wantType {
				t.Fatalf("type = %s, want %s", got.DeviceType, tt.wantType)
			}
			if got.Confidence < tt.minConf {
				t.Fatalf("confidence = %.2f, want >= %.2f", got.Confidence, tt.minConf)
			}
			if tt.maxConf > 0 && got.Confidence > tt.maxConf {
				t.Fatalf("confidence = %.2f, want <= %.2f", got.Confidence, tt.maxConf)
			}
			if got.Inferred != tt.inferred {
				t.Fatalf("inferred = %v, want %v", got.Inferred, tt.inferred)
			}
			if tt.wantRole != "" && !got.HasRole(tt.wantRole) {
				t.Fatalf("roles = %#v, want %s", got.Roles, tt.wantRole)
			}
		})
	}
}

func TestTopologyBuilderRules(t *testing.T) {
	now := time.Now().UTC()
	local := model.Device{ID: "local", DeviceType: model.DeviceTypeLocalHost, Roles: []model.DeviceRole{model.RoleLocalHost}, IPAddresses: []string{"192.168.1.10"}, Confidence: 0.90}
	gw := model.Device{ID: "gw", DeviceType: model.DeviceTypeRouter, Roles: []model.DeviceRole{model.RoleDefaultGateway}, IPAddresses: []string{"192.168.1.1"}, Confidence: 0.70}
	host1 := model.Device{ID: "host1", IPAddresses: []string{"192.168.1.20"}, DeviceType: model.DeviceTypeUnknown, Confidence: 0.35}
	host2 := model.Device{ID: "host2", IPAddresses: []string{"192.168.1.21"}, DeviceType: model.DeviceTypeUnknown, Confidence: 0.35}

	t.Run("LLDP neighbor creates high confidence edge", func(t *testing.T) {
		sw := model.Device{ID: "sw", DeviceType: model.DeviceTypeManagedSwitch, Confidence: 0.90}
		topology := discovery.BuildTopology([]model.Device{local, sw}, []model.Evidence{
			ev("lldp_cdp_passive_probe", "sw", model.EvidenceStrong, 0.90, "LLDP neighbor", map[string]any{"source": "local", "target": "sw", "edge_type": string(model.EdgeTypeEthernetLink)}, now),
		})
		edge := findEdge(topology, model.EdgeTypeEthernetLink)
		if edge == nil || edge.Confidence < 0.90 || edge.Inferred {
			t.Fatalf("expected high-confidence non-inferred LLDP edge, got %#v", topology.Edges)
		}
	})

	t.Run("unmanaged switch inferred from topology", func(t *testing.T) {
		topology := discovery.BuildTopology([]model.Device{local, gw, host1, host2}, []model.Evidence{
			ev("os_interface_probe", "local", model.EvidenceMedium, 0.70, "active ethernet", map[string]any{"active_interface": "ethernet"}, now),
		})
		node := findNode(topology, model.DeviceTypeInferredSwitch)
		if node == nil || !node.Inferred || node.Confidence > 0.60 {
			t.Fatalf("expected inferred switch <=0.60 confidence, got %#v", topology.Nodes)
		}
	})

	t.Run("Ethernet local host creates unknown or ethernet L2 link", func(t *testing.T) {
		topology := discovery.BuildTopology([]model.Device{local, gw}, []model.Evidence{
			ev("os_interface_probe", "local", model.EvidenceMedium, 0.70, "active ethernet", map[string]any{"active_interface": "ethernet"}, now),
		})
		if findEdge(topology, model.EdgeTypeEthernetLink) == nil && findEdge(topology, model.EdgeTypeUnknownLink) == nil {
			t.Fatalf("expected ethernet or unknown edge, got %#v", topology.Edges)
		}
	})

	t.Run("traceroute private hop creates upstream private gateway", func(t *testing.T) {
		upstream := model.Device{ID: "upstream", IPAddresses: []string{"10.0.0.1"}, DeviceType: model.DeviceTypeRouter, Roles: []model.DeviceRole{model.RoleUpstreamGateway}, Confidence: 0.60}
		topology := discovery.BuildTopology([]model.Device{local, gw, upstream}, nil)
		if findEdge(topology, model.EdgeTypeUpstreamNAT) == nil && findEdge(topology, model.EdgeTypeGatewayChain) == nil {
			t.Fatalf("expected upstream gateway edge, got %#v", topology.Edges)
		}
	})

	t.Run("public traceroute hop creates ISPRouteHop not LAN device", func(t *testing.T) {
		isp := model.Device{ID: "isp1", IPAddresses: []string{"8.8.8.8"}, DeviceType: model.DeviceTypeISPHop, Confidence: 0.50}
		topology := discovery.BuildTopology([]model.Device{local, gw, isp}, nil)
		if findNode(topology, model.DeviceTypeISPHop) == nil {
			t.Fatalf("expected ISP hop node, got %#v", topology.Nodes)
		}
		if findEdge(topology, model.EdgeTypeISPRouteHop) == nil {
			t.Fatalf("expected ISP route edge, got %#v", topology.Edges)
		}
		graph := output.GraphAdapter(topology)
		if len(graph.Nodes) == 0 || !hasISPSection(graph) {
			t.Fatalf("expected ISP hops separated in graph metadata, got %#v", graph.Nodes)
		}
	})
}

func TestMergeAndConflictRules(t *testing.T) {
	now := time.Now().UTC()
	t.Run("public IP is never scanned as local device", func(t *testing.T) {
		devices := discovery.MergeDevices([]model.ProbeResult{{ProbeName: "bad", Devices: []model.Device{{ID: "public", IPAddresses: []string{"8.8.8.8"}, DeviceType: model.DeviceTypeUnknown}}}})
		if len(devices) != 0 {
			t.Fatalf("public local devices should be excluded, got %#v", devices)
		}
	})

	t.Run("same MAC with multiple IPs merges into one device", func(t *testing.T) {
		results := []model.ProbeResult{
			{ProbeName: "arp", Devices: []model.Device{{ID: "a", IPAddresses: []string{"192.168.1.20"}, MACAddresses: []string{"AA:BB:CC:DD:EE:FF"}, Evidence: []model.Evidence{ev("arp", "a", model.EvidenceWeak, 0.35, "arp", nil, now)}}}},
			{ProbeName: "mdns", Devices: []model.Device{{ID: "b", IPAddresses: []string{"192.168.1.21"}, MACAddresses: []string{"aa:bb:cc:dd:ee:ff"}, Evidence: []model.Evidence{ev("mdns", "b", model.EvidenceMedium, 0.50, "mdns", nil, now)}}}},
		}
		devices := discovery.MergeDevices(results)
		if len(devices) != 1 || len(devices[0].IPAddresses) != 2 {
			t.Fatalf("expected one merged device with two IPs, got %#v", devices)
		}
	})

	t.Run("same IP with changed MAC creates conflict", func(t *testing.T) {
		devices := []model.Device{
			{ID: "a", IPAddresses: []string{"192.168.1.20"}, MACAddresses: []string{"aa:aa:aa:aa:aa:aa"}},
			{ID: "b", IPAddresses: []string{"192.168.1.20"}, MACAddresses: []string{"bb:bb:bb:bb:bb:bb"}},
		}
		conflicts := discovery.DetectConflicts(devices, nil)
		if !hasConflict(conflicts, "same_ip_different_macs") {
			t.Fatalf("expected same_ip_different_macs conflict, got %#v", conflicts)
		}
	})

	t.Run("virtual adapters are not mistaken as physical upstream devices", func(t *testing.T) {
		devices := []model.Device{{ID: "vpn", DeviceType: model.DeviceTypeVirtualAdapter, Roles: []model.DeviceRole{model.RoleUpstreamGateway}}}
		conflicts := discovery.DetectConflicts(devices, nil)
		if !hasConflict(conflicts, "virtual_adapter_mistaken_as_physical_upstream") {
			t.Fatalf("expected virtual adapter conflict, got %#v", conflicts)
		}
	})
}

func TestProbeSafetyRules(t *testing.T) {
	t.Run("SNMP disabled by default", func(t *testing.T) {
		res, err := (probe.SNMPOptInProbe{}).Run(context.Background(), model.ProbeInput{Mode: model.ScanModeDeep, CandidateIPs: []string{"192.168.1.1"}})
		if err != nil {
			t.Fatal(err)
		}
		if res.Status != model.ProbeStatusSkipped {
			t.Fatalf("SNMP status = %s, want skipped", res.Status)
		}
	})

	t.Run("no brute force behavior exists", func(t *testing.T) {
		res, err := (probe.SNMPOptInProbe{}).Run(context.Background(), model.ProbeInput{Mode: model.ScanModeDeep, CandidateIPs: []string{"192.168.1.1"}})
		if err != nil {
			t.Fatal(err)
		}
		text := strings.Join(res.Errors, " ")
		if !strings.Contains(text, "no community guessing") || !strings.Contains(text, "brute force") {
			t.Fatalf("expected explicit no guessing/bruteforce message, got %#v", res.Errors)
		}
	})
}

func TestISPTopologyIsObservedOnly(t *testing.T) {
	results := []model.ProbeResult{{
		ProbeName: "traceroute_isp_path_probe",
		Raw: map[string]any{"route_hops": []model.RouteHop{
			{Index: 1, IP: "192.168.1.1", Private: true},
			{Index: 2, IP: "100.64.1.1", Private: false},
			{Index: 3, IP: "8.8.8.8", Private: false, ASN: "AS15169"},
		}},
	}}
	path := discovery.BuildISPPath(results, model.NetworkContext{PublicIP: "100.64.9.9", ASN: "AS64500", ISP: "Example ISP"})
	if path.Warning != "Traceroute shows observed route hops only. It does not reveal full ISP infrastructure." {
		t.Fatalf("unexpected warning: %q", path.Warning)
	}
	if path.FirstPublicHop != "100.64.1.1" || len(path.PublicHops) != 2 {
		t.Fatalf("unexpected public hops: %#v", path)
	}
}

func ev(source, target string, strength model.EvidenceStrength, confidence float64, reason string, raw map[string]any, ts time.Time) model.Evidence {
	return model.NewEvidence(source, target, strength, confidence, reason, raw, ts)
}

func findEdge(topology model.Topology, rel model.EdgeType) *model.TopologyEdge {
	for i := range topology.Edges {
		if topology.Edges[i].Relationship == rel {
			return &topology.Edges[i]
		}
	}
	return nil
}

func findNode(topology model.Topology, typ model.DeviceType) *model.TopologyNode {
	for i := range topology.Nodes {
		if topology.Nodes[i].DeviceType == typ {
			return &topology.Nodes[i]
		}
	}
	return nil
}

func hasConflict(conflicts []model.Conflict, typ string) bool {
	for _, c := range conflicts {
		if c.Type == typ {
			return true
		}
	}
	return false
}

func hasISPSection(graph output.UIGraph) bool {
	for _, node := range graph.Nodes {
		if node.Metadata["section"] == "isp_path" {
			return true
		}
	}
	return false
}
