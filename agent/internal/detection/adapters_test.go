package detection

import "testing"

func TestIsAPIPA(t *testing.T) {
	if !isAPIPA("169.254.10.20") {
		t.Error("169.254.x.x must be APIPA")
	}
	if !isAPIPA("169.254.218.14/16") {
		t.Error("APIPA with CIDR suffix must be detected")
	}
	if isAPIPA("192.168.1.10") {
		t.Error("private address must not be APIPA")
	}
}

func TestIsVirtualAdapter(t *testing.T) {
	virtual := []string{"VMware Network Adapter VMnet8", "vEthernet (WSL)", "Bluetooth Ağ Bağlantısı", "Tailscale"}
	for _, n := range virtual {
		if !isVirtualAdapter(n) {
			t.Errorf("%q should be virtual", n)
		}
	}
	if isVirtualAdapter("Ethernet") {
		t.Error("Ethernet should not be virtual")
	}
}

func TestPickMainAdapter_SkipsVirtualAndAPIPA(t *testing.T) {
	adapters := []AdapterInfo{
		{Name: "Ethernet 2", Up: true, Addrs: []string{"169.254.218.14/16"}},                 // APIPA only
		{Name: "VMware Network Adapter VMnet8", Up: true, Addrs: []string{"192.168.50.1/24"}}, // virtual
		{Name: "Ethernet", Up: true, Addrs: []string{"192.168.31.147/24"}},                   // real uplink
		{Name: "Wi-Fi", Up: false, Addrs: []string{"192.168.31.5/24"}},                       // down
	}
	main := pickMainAdapter(adapters)
	if main.Name != "Ethernet" {
		t.Fatalf("main adapter = %q, want Ethernet", main.Name)
	}
	if main.IP != "192.168.31.147" {
		t.Errorf("main IP = %q, want 192.168.31.147", main.IP)
	}
	if main.Access != "Ethernet" {
		t.Errorf("local access = %q, want Ethernet", main.Access)
	}
}

func TestInferLocalAccess(t *testing.T) {
	cases := map[string]string{
		"Ethernet":         "Ethernet",
		"Wi-Fi":            "Wi-Fi",
		"Wireless LAN":     "Wi-Fi",
		"Cellular":         "Cellular",
		"Some TAP Adapter": "Unknown",
	}
	for name, want := range cases {
		if got := inferLocalAccess(name); got != want {
			t.Errorf("inferLocalAccess(%q) = %q, want %q", name, got, want)
		}
	}
}
