package discovery

import (
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func ifaceWith(name, mac, ip, cidr string, up, virtual bool) models.InterfaceInfo {
	return models.InterfaceInfo{
		Name:      name,
		MAC:       mac,
		Up:        up,
		Virtual:   virtual,
		CIDR:      cidr,
		Addresses: []models.IPAddress{{IP: ip, Version: 4, CIDR: cidr}},
	}
}

// Mirrors the real bug report: an up VirtualBox host-only adapter holding an
// APIPA (169.254/16) address appeared before the real LAN. It must never be
// chosen over a routable RFC1918 interface, even when no gateway is known.
func TestSelectPrimaryPrefersRoutableLANOverLinkLocal(t *testing.T) {
	ifaces := []models.InterfaceInfo{
		ifaceWith("Ethernet 2", "0a:00:27:00:00:17", "169.254.218.14", "169.254.0.0/16", true, false),
		ifaceWith("Ethernet", "74:d4:dd:39:04:66", "192.168.31.147", "192.168.31.0/24", true, false),
	}

	primary, _, ok := SelectPrimary(ifaces, "", false)
	if !ok {
		t.Fatal("expected an interface to be selected")
	}
	if primary.Name != "Ethernet" {
		t.Fatalf("selected %q, want the routable LAN %q", primary.Name, "Ethernet")
	}
}

// A known default gateway inside an interface's network outranks everything else.
func TestSelectPrimaryPrefersGatewayNetwork(t *testing.T) {
	ifaces := []models.InterfaceInfo{
		ifaceWith("eth0", "aa:bb:cc:00:00:01", "192.168.1.10", "192.168.1.0/24", true, false),
		ifaceWith("eth1", "aa:bb:cc:00:00:02", "10.0.0.5", "10.0.0.0/24", true, false),
	}

	primary, _, ok := SelectPrimary(ifaces, "10.0.0.1", false)
	if !ok {
		t.Fatal("expected an interface to be selected")
	}
	if primary.Name != "eth1" {
		t.Fatalf("selected %q, want the gateway-bearing %q", primary.Name, "eth1")
	}
}

// With nothing but a link-local interface, it is still chosen as a last resort
// rather than failing outright.
func TestSelectPrimaryFallsBackToLinkLocal(t *testing.T) {
	ifaces := []models.InterfaceInfo{
		ifaceWith("Ethernet 2", "0a:00:27:00:00:17", "169.254.218.14", "169.254.0.0/16", true, false),
	}

	primary, _, ok := SelectPrimary(ifaces, "", false)
	if !ok {
		t.Fatal("expected the link-local interface to be selected as a last resort")
	}
	if primary.Name != "Ethernet 2" {
		t.Fatalf("selected %q, want %q", primary.Name, "Ethernet 2")
	}
}

// Virtual adapters are excluded unless explicitly included.
func TestSelectPrimarySkipsVirtualUnlessIncluded(t *testing.T) {
	ifaces := []models.InterfaceInfo{
		ifaceWith("vmnet8", "00:50:56:c0:00:08", "192.168.200.1", "192.168.200.0/24", true, true),
		ifaceWith("Ethernet", "74:d4:dd:39:04:66", "192.168.31.147", "192.168.31.0/24", true, false),
	}

	primary, _, ok := SelectPrimary(ifaces, "", false)
	if !ok || primary.Name != "Ethernet" {
		t.Fatalf("selected %q (ok=%v), want the physical %q", primary.Name, ok, "Ethernet")
	}

	primary, _, ok = SelectPrimary(ifaces, "", true)
	if !ok {
		t.Fatal("expected a selection when virtual adapters are included")
	}
}

// Down interfaces are never selected.
func TestSelectPrimarySkipsDownInterfaces(t *testing.T) {
	ifaces := []models.InterfaceInfo{
		ifaceWith("Ethernet", "74:d4:dd:39:04:66", "192.168.31.147", "192.168.31.0/24", false, false),
	}
	if _, _, ok := SelectPrimary(ifaces, "", false); ok {
		t.Fatal("a down interface must not be selected")
	}
}
