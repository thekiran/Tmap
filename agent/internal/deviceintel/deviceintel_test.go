package deviceintel

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

var testTime = time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)

func TestDeviceIntelRequestedCases(t *testing.T) {
	tests := []struct {
		name  string
		build func() models.ScanReport
		check func(t *testing.T, out models.DeviceIntelReport)
	}{
		{
			name: "ARP-only host becomes discovered unknown host",
			build: func() models.ScanReport {
				return reportWith(models.Device{
					ID:          "dev-192.168.31.157",
					Addresses:   []models.IPAddress{{IP: "192.168.31.157", Version: 4}},
					Interfaces:  []models.DeviceInterface{{MAC: "aa:bb:cc:dd:ee:ff", IPs: []string{"192.168.31.157"}}},
					Confidence:  0.50,
					EvidenceIDs: []string{"ev-arp"},
				}, nil, []models.Evidence{ev("ev-arp", "arp_table", "arp_table", map[string]string{"ip": "192.168.31.157", "mac": "aa:bb:cc:dd:ee:ff"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.157")
				if d.DeviceType.Primary != models.DeviceTypeUnknown {
					t.Fatalf("type = %s, want unknown", d.DeviceType.Primary)
				}
				if !hasFinding(d, "unknown_device") {
					t.Fatalf("unknown_device finding missing: %#v", d.SecurityPosture.Findings)
				}
			},
		},
		{
			name: "TCP open port creates service entry",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.20", []models.Service{{Port: 80, Protocol: "tcp", State: "open", Name: "http", EvidenceIDs: []string{"ev-tcp"}}}), nil,
					[]models.Evidence{ev("ev-tcp", "tcp_connect", "tcp_sweep", map[string]string{"ip": "192.168.31.20", "ports": "80"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.20")
				if !hasService(d, 80) {
					t.Fatalf("port 80 service missing: %#v", d.Services)
				}
				if len(d.Services[0].EvidenceIDs) == 0 {
					t.Fatal("service evidence_ids missing")
				}
			},
		},
		{
			name: "ICMP timeout does not mark host offline",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.21", nil), nil,
					[]models.Evidence{ev("ev-icmp", "icmp_echo", "icmp_sweep", map[string]string{"ip": "192.168.31.21", "status": "timeout"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.21")
				if d.Confidence == 0 {
					t.Fatal("device disappeared after ICMP timeout")
				}
				if len(d.FailedAttempts) == 0 {
					t.Fatal("ICMP timeout should be recorded as failed attempt")
				}
			},
		},
		{
			name: "HTTP success followed by timeout keeps HTTP evidence",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.1", nil), nil, []models.Evidence{
					ev("ev-http-ok", "http_fingerprint", "http_fingerprint_v2", map[string]string{"ip": "192.168.31.1", "url": "http://192.168.31.1/", "method": "GET", "status_code": "200", "server": "nginx"}),
					ev("ev-http-fail", "http_fingerprint", "http_fingerprint_v3", map[string]string{"ip": "192.168.31.1", "url": "http://192.168.31.1/", "status": "timeout", "error": "timeout"}),
				}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.1")
				if len(d.HTTPFingerprints) != 1 {
					t.Fatalf("http fingerprints = %d, want 1", len(d.HTTPFingerprints))
				}
				if len(d.FailedAttempts) == 0 {
					t.Fatal("failed HTTP attempt missing")
				}
			},
		},
		{
			name: "Generic nginx does not classify model",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.1", []models.Service{{Port: 80, Protocol: "tcp", State: "open", EvidenceIDs: []string{"ev-tcp"}}}), nil,
					[]models.Evidence{ev("ev-http", "http_fingerprint", "http_fingerprint_v2", map[string]string{"ip": "192.168.31.1", "server": "nginx", "status_code": "200"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.1")
				if d.Vendor.FingerprintVendor != nil {
					t.Fatalf("generic nginx created vendor: %v", *d.Vendor.FingerprintVendor)
				}
				if !containsJoined(d.ClassificationExplanation, "Generic nginx") {
					t.Fatalf("generic nginx warning missing: %#v", d.ClassificationExplanation)
				}
			},
		},
		{
			name: "mDNS printer service classifies printer candidate",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.30", nil), nil,
					[]models.Evidence{ev("ev-mdns", "mdns", "mdns_probe", map[string]string{"ip": "192.168.31.30", "service": "_ipp._tcp", "name": "Office Printer"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.30")
				if d.DeviceType.Primary != models.DeviceTypePrinter {
					t.Fatalf("type = %s, want printer", d.DeviceType.Primary)
				}
			},
		},
		{
			name: "TCP 9100 classifies printer candidate",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.31", []models.Service{{Port: 9100, Protocol: "tcp", State: "open", EvidenceIDs: []string{"ev-p9100"}}}), nil,
					[]models.Evidence{ev("ev-p9100", "tcp_connect", "tcp_sweep", map[string]string{"ip": "192.168.31.31", "ports": "9100"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.31")
				if d.DeviceType.Primary != models.DeviceTypePrinter || d.PrinterInfo == nil {
					t.Fatalf("printer not classified: %#v", d.DeviceType)
				}
			},
		},
		{
			name: "SMB 445 plus NetBIOS classifies Windows candidate",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.40", []models.Service{{Port: 445, Protocol: "tcp", State: "open", EvidenceIDs: []string{"ev-smb"}}}), nil,
					[]models.Evidence{ev("ev-nbns", "nbns", "nbns_probe", map[string]string{"ip": "192.168.31.40", "name": "WINBOX", "workgroup": "WORKGROUP"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.40")
				if d.DeviceType.Primary != models.DeviceTypeWindowsHost || d.OSGuess.Family != "windows" {
					t.Fatalf("windows classification failed: type=%s os=%s", d.DeviceType.Primary, d.OSGuess.Family)
				}
			},
		},
		{
			name: "VMware OUI classifies virtual device",
			build: func() models.ScanReport {
				d := device("192.168.31.50", nil)
				d.Interfaces = []models.DeviceInterface{{MAC: "00:0c:29:aa:bb:cc", Vendor: "VMware", IPs: []string{"192.168.31.50"}}}
				d.EvidenceIDs = []string{"ev-arp"}
				return reportWith(d, nil, []models.Evidence{ev("ev-arp", "arp_table", "arp_table", map[string]string{"ip": "192.168.31.50", "mac": "00:0c:29:aa:bb:cc", "vendor": "VMware"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.50")
				if d.DeviceType.Primary != models.DeviceTypeVirtualMachine {
					t.Fatalf("type = %s, want virtual_machine", d.DeviceType.Primary)
				}
			},
		},
		{
			name: "host.docker.internal is low-confidence virtual hostname",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.51", nil), nil,
					[]models.Evidence{ev("ev-rdns", "reverse_dns", "reverse_dns", map[string]string{"ip": "192.168.31.51", "hostname": "host.docker.internal"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.51")
				if !containsCandidate(d.DeviceType.Candidates, models.DeviceTypeVirtualMachine) {
					t.Fatalf("virtual candidate missing: %#v", d.DeviceType.Candidates)
				}
				if d.DeviceType.Primary != models.DeviceTypeUnknown {
					t.Fatalf("host.docker.internal should stay low confidence unknown, got %s", d.DeviceType.Primary)
				}
			},
		},
		{
			name: "Same subnet edge is inferred low confidence",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.20", nil), []models.TopologyEdge{{ID: "edge-l2", Source: "dev-192.168.31.147", Target: "dev-192.168.31.20", Type: models.EdgeInferredL2, Confidence: 0.70, EvidenceIDs: []string{"ev-l2"}, Reason: "same subnet"}}, nil, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				e := mustEdge(t, out, models.DeviceEdgeSameSubnetInferred)
				if !e.Inferred || e.Physical || e.Confidence > 0.35 {
					t.Fatalf("bad inferred edge: %#v", e)
				}
			},
		},
		{
			name: "Default gateway edge is route edge not physical cable",
			build: func() models.ScanReport {
				return reportWith(gatewayDevice(), []models.TopologyEdge{{ID: "edge-gw", Source: "dev-192.168.31.147", Target: "dev-192.168.31.1", Type: models.EdgeGatewayDefault, Confidence: 0.80, EvidenceIDs: []string{"ev-gw"}, Reason: "default route"}}, nil, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				e := mustEdge(t, out, models.DeviceEdgeDefaultGatewayRoute)
				if !e.Inferred || e.Physical {
					t.Fatalf("default gateway edge should be inferred route: %#v", e)
				}
			},
		},
		{
			name: "LLDP neighbor creates high confidence physical edge",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.2", nil), []models.TopologyEdge{{ID: "edge-lldp", Source: "dev-192.168.31.147", Target: "dev-192.168.31.2", Type: models.EdgeDirectLLDP, Confidence: 0.95, EvidenceIDs: []string{"ev-lldp"}, Reason: "lldp"}}, nil, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				e := mustEdge(t, out, models.DeviceEdgeLLDPPhysicalNeighbor)
				if !e.Physical || e.Inferred || e.Confidence < 0.90 {
					t.Fatalf("bad lldp edge: %#v", e)
				}
			},
		},
		{
			name: "Public IP evidence skipped when public not allowed",
			build: func() models.ScanReport {
				return reportWith(models.Device{}, nil, []models.Evidence{ev("ev-pub", "tcp_connect", "tcp_sweep", map[string]string{"ip": "8.8.8.8", "ports": "80"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				if findDevice(out, "8.8.8.8") != nil {
					t.Fatal("public device should be skipped")
				}
			},
		},
		{
			name: "Traceroute upstream private hop becomes upstream gateway candidate",
			build: func() models.ScanReport {
				state := &models.GatewayChainState{
					DefaultGateway: "192.168.31.1",
					PrivateHops: []models.GatewayHop{
						{IP: "192.168.31.1", Role: "default_gateway", Source: "traceroute_probe", Order: 1, EvidenceID: "ev-trace"},
						{IP: "192.168.1.1", Role: "upstream_private_gateway", Source: "traceroute_probe", Order: 2, EvidenceID: "ev-trace"},
					},
					InternalDoubleNATPossible: true,
					Confidence:                0.65,
				}
				return reportWith(gatewayDevice(), nil, []models.Evidence{ev("ev-trace", "gateway_route", "traceroute_probe", map[string]string{"ip": "192.168.31.1"})},
					&models.ScanResult{DetectedNetworkContext: &models.NetworkContext{GatewayChainState: state}})
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.1.1")
				if !d.Topology.IsUpstreamGatewayCandidate || d.DeviceType.Primary != models.DeviceTypeUpstreamCPE {
					t.Fatalf("upstream candidate not marked: %#v", d)
				}
				_ = mustEdge(t, out, models.DeviceEdgeUpstreamPrivate)
			},
		},
		{
			name: "Unknown MAC/IP remains unknown not fake classified",
			build: func() models.ScanReport {
				d := device("192.168.31.160", nil)
				d.Interfaces = []models.DeviceInterface{{MAC: "aa:aa:aa:aa:aa:aa", IPs: []string{"192.168.31.160"}}}
				return reportWith(d, nil, nil, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.160")
				if d.DeviceType.Primary != models.DeviceTypeUnknown {
					t.Fatalf("type = %s, want unknown", d.DeviceType.Primary)
				}
			},
		},
		{
			name: "Device type includes alternatives and missing evidence",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.61", []models.Service{
					{Port: 445, Protocol: "tcp", State: "open", EvidenceIDs: []string{"ev-smb"}},
					{Port: 9100, Protocol: "tcp", State: "open", EvidenceIDs: []string{"ev-prn"}},
				}), nil, nil, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.61")
				if len(d.DeviceType.Alternatives) == 0 || len(d.DeviceType.MissingEvidence) == 0 {
					t.Fatalf("alternatives/missing evidence absent: %#v", d.DeviceType)
				}
			},
		},
		{
			name: "Telnet finding is detection only and no login",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.70", []models.Service{{Port: 23, Protocol: "tcp", State: "open", EvidenceIDs: []string{"ev-telnet"}}}), nil, nil, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.70")
				if !hasFinding(d, "open_telnet") || !containsFindingText(d, "did not attempt login") {
					t.Fatalf("telnet safe finding missing: %#v", d.SecurityPosture.Findings)
				}
			},
		},
		{
			name: "Printer detection does not send print jobs",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.71", []models.Service{{Port: 9100, Protocol: "tcp", State: "open", EvidenceIDs: []string{"ev-prn"}}}), nil, nil, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.71")
				if !containsFindingText(d, "did not send print jobs") {
					t.Fatalf("print-job safety text missing: %#v", d.SecurityPosture.Findings)
				}
			},
		},
		{
			name:  "Security notes forbid exploit checks",
			build: func() models.ScanReport { return reportWith(device("192.168.31.72", nil), nil, nil, nil) },
			check: func(t *testing.T, out models.DeviceIntelReport) {
				if !containsJoined(out.SecurityNotes, "exploit checks") {
					t.Fatalf("exploit safety note missing: %#v", out.SecurityNotes)
				}
			},
		},
		{
			name:  "Security notes forbid brute force",
			build: func() models.ScanReport { return reportWith(device("192.168.31.73", nil), nil, nil, nil) },
			check: func(t *testing.T, out models.DeviceIntelReport) {
				if !containsJoined(out.UI.Badges, "no-bruteforce") {
					t.Fatalf("no-bruteforce badge missing: %#v", out.UI.Badges)
				}
			},
		},
		{
			name: "All service facts include evidence id",
			build: func() models.ScanReport {
				return reportWith(device("192.168.31.74", []models.Service{{Port: 80, Protocol: "tcp", State: "open", EvidenceIDs: []string{"ev-http"}}}), nil,
					[]models.Evidence{ev("ev-http", "tcp_connect", "tcp_sweep", map[string]string{"ip": "192.168.31.74", "ports": "80"})}, nil)
			},
			check: func(t *testing.T, out models.DeviceIntelReport) {
				d := mustDevice(t, out, "192.168.31.74")
				if len(d.EvidenceIDs) == 0 || len(d.Services) == 0 || len(d.Services[0].EvidenceIDs) == 0 {
					t.Fatalf("missing evidence IDs: device=%#v services=%#v", d.EvidenceIDs, d.Services)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, Build(tt.build()))
		})
	}
}

func TestProbeSafetyAndOptIn(t *testing.T) {
	t.Run("SNMP skipped by default", func(t *testing.T) {
		store := NewEvidenceStore(func() time.Time { return testTime })
		res := SNMPOptInProbe().Run(context.Background(), ScanScope{PrivateOnly: true, Targets: []string{"192.168.31.1"}}, store)
		if res.Status != StatusSkipped || !containsJoined(res.Errors, "requires explicit opt-in credentials") {
			t.Fatalf("SNMP result = %#v", res)
		}
	})

	t.Run("SNMP credentials and opt-in required", func(t *testing.T) {
		if SNMPAllowed(ScanScope{SNMPCredentials: true}) {
			t.Fatal("SNMP must require both credentials and opt-in probe")
		}
		if !SNMPAllowed(ScanScope{SNMPCredentials: true, OptInProbes: []string{"snmp"}}) {
			t.Fatal("SNMP should be allowed only with credentials plus opt-in")
		}
	})

	t.Run("Safe TCP probe skips public targets", func(t *testing.T) {
		calls := 0
		probe := SafeTCPProbe{Ports: []int{80}, Dial: func(ctx context.Context, network, address string) error {
			calls++
			return nil
		}}
		store := NewEvidenceStore(func() time.Time { return testTime })
		res := probe.Run(context.Background(), ScanScope{PrivateOnly: true, PublicAllowed: false, Targets: []string{"8.8.8.8"}}, store)
		if calls != 0 || res.Status != StatusSkipped {
			t.Fatalf("public target was probed: calls=%d result=%#v", calls, res)
		}
	})

	t.Run("HTTP login page is not submitted", func(t *testing.T) {
		client := &recordingHTTPDoer{}
		_, err := httpProbeOnce(context.Background(), client, "http://192.168.31.1/")
		if err != nil {
			t.Fatalf("http probe failed: %v", err)
		}
		if containsJoined(client.methods, http.MethodPost) {
			t.Fatalf("login form submission attempted: %v", client.methods)
		}
	})

	t.Run("TCP probe records open service without login", func(t *testing.T) {
		probe := SafeTCPProbe{Ports: []int{23}, Dial: func(ctx context.Context, network, address string) error {
			if !strings.Contains(address, ":23") {
				return errors.New("unexpected port")
			}
			return nil
		}}
		store := NewEvidenceStore(func() time.Time { return testTime })
		res := probe.Run(context.Background(), ScanScope{PrivateOnly: true, Targets: []string{"192.168.31.80"}}, store)
		if res.Status != StatusSuccess || len(res.Observations) != 1 {
			t.Fatalf("tcp result = %#v", res)
		}
		classifyAll(store)
		d := store.Devices["dev-192.168.31.80"]
		if d == nil || !hasFinding(*d, "open_telnet") {
			t.Fatalf("telnet finding missing after safe TCP probe: %#v", d)
		}
	})
}

func TestCurrentScanFixture(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "fixtures", "device_intel_current_scan.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var out models.DeviceIntelReport
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("fixture JSON invalid: %v", err)
	}
	if out.SchemaVersion != models.DeviceIntelSchema {
		t.Fatalf("schema = %s", out.SchemaVersion)
	}
	gw := mustDevice(t, out, "192.168.31.1")
	if gw.DeviceType.Primary != models.DeviceTypeGatewayRouter || !containsJoined(gw.ClassificationExplanation, "Generic nginx") {
		t.Fatalf("gateway fixture classification wrong: %#v", gw.DeviceType)
	}
	upstream := mustDevice(t, out, "192.168.1.1")
	if !upstream.Topology.IsUpstreamGatewayCandidate {
		t.Fatalf("upstream CPE candidate missing: %#v", upstream.Topology)
	}
	if e := mustEdge(t, out, models.DeviceEdgeUpstreamPrivate); e.Physical {
		t.Fatalf("upstream path must not be physical: %#v", e)
	}
	if out.Summary.PhysicalEdges != 0 {
		t.Fatalf("physical edge count = %d, want 0", out.Summary.PhysicalEdges)
	}
}

type recordingHTTPDoer struct {
	methods []string
}

func (r *recordingHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	r.methods = append(r.methods, req.Method)
	body := ioNopCloser{strings.NewReader("<html><title>Login</title><form></form></html>")}
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       body,
		Request:    req,
	}, nil
}

type ioNopCloser struct {
	*strings.Reader
}

func (c ioNopCloser) Close() error { return nil }

func reportWith(d models.Device, edges []models.TopologyEdge, evidence []models.Evidence, access *models.ScanResult) models.ScanReport {
	devices := []models.Device{agentDevice()}
	if d.ID != "" {
		devices = append(devices, d)
	}
	return models.ScanReport{
		SchemaVersion: models.TopologyReportSchema,
		ScanID:        "test-scan",
		CreatedAt:     testTime,
		Agent: models.AgentInfo{
			Gateway: "192.168.31.1",
			Interfaces: []models.InterfaceInfo{{
				Name:      "Ethernet",
				Selected:  true,
				CIDR:      "192.168.31.0/24",
				Addresses: []models.IPAddress{{IP: "192.168.31.147", Version: 4}},
			}},
		},
		Scope: models.ScanScope{
			CIDR:          "192.168.31.0/24",
			Interface:     "Ethernet",
			Private:       true,
			PublicAllowed: false,
			Profile:       "deep",
		},
		Devices:              devices,
		Edges:                edges,
		Evidence:             evidence,
		AccessClassification: access,
	}
}

func agentDevice() models.Device {
	return models.Device{
		ID:          "dev-192.168.31.147",
		Addresses:   []models.IPAddress{{IP: "192.168.31.147", Version: 4}},
		IsAgent:     true,
		Roles:       []string{models.RoleAgent},
		Confidence:  0.95,
		EvidenceIDs: []string{"ev-agent"},
	}
}

func gatewayDevice() models.Device {
	d := device("192.168.31.1", []models.Service{
		{Port: 53, Protocol: "tcp", State: "open", EvidenceIDs: []string{"ev-gw-svc"}},
		{Port: 80, Protocol: "tcp", State: "open", EvidenceIDs: []string{"ev-gw-svc"}},
	})
	d.IsGateway = true
	d.Roles = []string{models.RoleGateway, models.RoleRouter}
	return d
}

func device(ip string, services []models.Service) models.Device {
	return models.Device{
		ID:          "dev-" + ip,
		Addresses:   []models.IPAddress{{IP: ip, Version: 4}},
		Services:    services,
		Confidence:  0.70,
		EvidenceIDs: []string{"ev-" + strings.ReplaceAll(ip, ".", "-")},
	}
}

func ev(id, kind, source string, data map[string]string) models.Evidence {
	return models.Evidence{ID: id, Kind: kind, Source: source, Summary: id, Data: data, Timestamp: testTime}
}

func mustDevice(t *testing.T, out models.DeviceIntelReport, ip string) models.DeviceIntelDevice {
	t.Helper()
	d := findDevice(out, ip)
	if d == nil {
		t.Fatalf("device %s not found in %#v", ip, out.Devices)
	}
	return *d
}

func findDevice(out models.DeviceIntelReport, ip string) *models.DeviceIntelDevice {
	for i := range out.Devices {
		for _, gotIP := range out.Devices[i].IPAddresses {
			if gotIP == ip {
				return &out.Devices[i]
			}
		}
	}
	return nil
}

func mustEdge(t *testing.T, out models.DeviceIntelReport, typ string) models.DeviceIntelEdge {
	t.Helper()
	for _, e := range out.Edges {
		if e.Type == typ {
			return e
		}
	}
	t.Fatalf("edge type %s not found: %#v", typ, out.Edges)
	return models.DeviceIntelEdge{}
}

func hasService(d models.DeviceIntelDevice, port int) bool {
	for _, s := range d.Services {
		if s.Port == port {
			return true
		}
	}
	return false
}

func hasFinding(d models.DeviceIntelDevice, id string) bool {
	for _, f := range d.SecurityPosture.Findings {
		if f.ID == id {
			return true
		}
	}
	return false
}

func containsFindingText(d models.DeviceIntelDevice, needle string) bool {
	for _, f := range d.SecurityPosture.Findings {
		if strings.Contains(f.Description, needle) || strings.Contains(f.SafeRecommendation, needle) {
			return true
		}
	}
	return false
}

func containsCandidate(candidates []models.DeviceTypeCandidate, typ string) bool {
	for _, c := range candidates {
		if c.Type == typ {
			return true
		}
	}
	return false
}

func containsJoined(values []string, needle string) bool {
	return strings.Contains(strings.Join(values, " | "), needle)
}
