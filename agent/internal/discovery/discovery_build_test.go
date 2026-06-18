package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/thekiran/iad/internal/topology"
	"github.com/thekiran/iad/pkg/models"
)

// An ARP-responding host with no open ports must still appear on the map.
func TestBuildDevicesIncludesARPOnlyHosts(t *testing.T) {
	now := func() time.Time { return time.Unix(0, 0).UTC() }
	norm := &Normalizer{Store: topology.NewEvidenceStore(now)} // nil Resolver: skip reverse DNS

	res := sweepResult{
		hits: []HostHit{{IP: "192.168.1.50", Alive: true, OpenPorts: []int{80}}},
		arp: []ARPEntry{
			{IP: "192.168.1.20", MAC: "aa:bb:cc:dd:ee:ff"}, // ARP-only host (no ports)
			{IP: "192.168.1.50", MAC: "11:22:33:44:55:66"},
		},
		sources: map[string][]string{
			"192.168.1.20": {srcARPSweep},
			"192.168.1.50": {srcTCP, srcARPSweep},
		},
	}

	devices, agentID, gwID := buildDevices(
		context.Background(), norm, models.InterfaceInfo{},
		"192.168.1.10", "192.168.1.1", "", res, now())

	byID := map[string]models.Device{}
	for _, d := range devices {
		byID[d.ID] = d
	}

	if len(devices) != 4 {
		t.Fatalf("device count = %d, want 4 (agent, gateway, arp-only, tcp): %+v", len(devices), devices)
	}

	arpOnly, ok := byID["dev-192.168.1.20"]
	if !ok {
		t.Fatalf("ARP-only host missing from devices")
	}
	if arpOnly.Reachability != "arp_only" {
		t.Errorf("reachability = %q, want arp_only", arpOnly.Reachability)
	}
	if len(arpOnly.Services) != 0 {
		t.Errorf("ARP-only host should have no services, got %d", len(arpOnly.Services))
	}
	if arpOnly.MAC != "aa:bb:cc:dd:ee:ff" {
		t.Errorf("mac = %q, want aa:bb:cc:dd:ee:ff", arpOnly.MAC)
	}
	if !hasString(arpOnly.DiscoverySources, srcARPSweep) {
		t.Errorf("discovery_sources = %v, want to contain %q", arpOnly.DiscoverySources, srcARPSweep)
	}

	if tcp := byID["dev-192.168.1.50"]; tcp.Reachability != "reachable" || len(tcp.Services) == 0 {
		t.Errorf("tcp host: reachability=%q services=%d, want reachable with services", tcp.Reachability, len(tcp.Services))
	}

	if agentID != "dev-192.168.1.10" || gwID != "dev-192.168.1.1" {
		t.Errorf("agent=%q gateway=%q", agentID, gwID)
	}
	if byID["dev-192.168.1.10"].Reachability != "self" {
		t.Errorf("agent reachability = %q, want self", byID["dev-192.168.1.10"].Reachability)
	}

	sum := buildDiscoverySummary(models.ScanScope{CIDR: "192.168.1.0/24", HostCount: 254}, res, len(devices), 1234)
	if sum.DevicesFound != 4 || sum.ARPFound != 2 || sum.TCPFound != 1 {
		t.Errorf("summary = %+v, want devices=4 arp=2 tcp=1", sum)
	}
	if sum.AddressesScanned != 254 || sum.CIDR != "192.168.1.0/24" || sum.ScanDurationMS != 1234 {
		t.Errorf("summary scope fields wrong: %+v", sum)
	}
}

func TestReachabilityState(t *testing.T) {
	cases := []struct {
		name                                       string
		isAgent, isGateway, tcpOpen, nmap, hasMAC bool
		want                                       string
	}{
		{"agent", true, false, false, false, false, "self"},
		{"gateway", false, true, false, false, false, "reachable"},
		{"tcp", false, false, true, false, true, "reachable"},
		{"nmap", false, false, false, true, false, "reachable"},
		{"arp-only", false, false, false, false, true, "arp_only"},
		{"nothing", false, false, false, false, false, "unknown"},
	}
	for _, c := range cases {
		if got := reachabilityState(c.isAgent, c.isGateway, c.tcpOpen, c.nmap, c.hasMAC); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func hasString(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
