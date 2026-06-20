package upstream

import (
	"slices"
	"testing"
	"time"
)

func hasTag(tags []string, tag string) bool { return slices.Contains(tags, tag) }

func TestClassify_ReachableUpstreamRouter(t *testing.T) {
	f := Facts{
		IP: "192.168.1.1", IsPrivate: true, InGatewayChain: true, HopIndex: 2, HopDistance: 2,
		ReachableICMP: true, ReachableTCP: true, OpenPorts: []int{80, 53},
		HTTPVendor: "Huawei", HTTPAdminPanel: true, RouterLikeHTTP: true, HasDNSService: true,
		Now: time.Unix(1700000000, 0),
	}
	c := Classify(f)
	if !hasTag(c.Tags, TagUpstreamGateway) || !hasTag(c.Tags, TagPossibleCPE) {
		t.Fatalf("expected upstream+possible-cpe tags, got %v", c.Tags)
	}
	if c.Confidence < 0.6 {
		t.Fatalf("expected strong confidence from reachable+router-like+upnp evidence, got %.2f", c.Confidence)
	}
	if len(c.Evidence) == 0 {
		t.Fatal("expected weighted evidence items")
	}
}

func TestClassify_IPPatternOnlyStaysInferred(t *testing.T) {
	// Only "it's a private IP" — no reachability, no services, not in a chain.
	c := Classify(Facts{IP: "192.168.1.1", IsPrivate: true, Now: time.Unix(1700000000, 0)})
	if c.Confidence >= 0.30 {
		t.Fatalf("a bare private IP must not be confident, got %.2f", c.Confidence)
	}
	if !hasTag(c.Tags, TagUnknownInfra) {
		t.Fatalf("expected UNKNOWN_INFRASTRUCTURE fallback, got %v", c.Tags)
	}
	if len(c.Warnings) == 0 {
		t.Fatal("expected an inferred/low-evidence warning")
	}
}

func TestClassify_DoubleNATAndISPCPE(t *testing.T) {
	dn := Classify(Facts{IP: "192.168.1.1", IsPrivate: true, InGatewayChain: true, DoubleNATHint: true, ReachableTCP: true, Now: time.Now()})
	if !hasTag(dn.Tags, TagDoubleNATUpstream) {
		t.Fatalf("expected DOUBLE_NAT_UPSTREAM, got %v", dn.Tags)
	}
	cpe := Classify(Facts{IP: "192.168.1.1", IsPrivate: true, InGatewayChain: true, ReachableTCP: true, HasCWMP: true, OpenPorts: []int{7547}, Now: time.Now()})
	if !hasTag(cpe.Tags, TagISPCPE) {
		t.Fatalf("expected ISP_CPE from CWMP, got %v", cpe.Tags)
	}
}

func TestClassify_DefaultGateway(t *testing.T) {
	c := Classify(Facts{IP: "192.168.31.1", IsPrivate: true, SameSubnetAsAgent: true, IsDefaultGateway: true, InGatewayChain: true, ReachableICMP: true, Now: time.Now()})
	if !hasTag(c.Tags, TagGateway) || !hasTag(c.Tags, TagRouter) {
		t.Fatalf("expected GATEWAY+ROUTER, got %v", c.Tags)
	}
}

func TestAnalyzeHTTP(t *testing.T) {
	v := AnalyzeHTTP("Huawei HomeGateway", "Login - HG8245", "", "")
	if v.Vendor != "Huawei" || !v.AdminPanel || !v.RouterLike {
		t.Fatalf("huawei admin detection failed: %#v", v)
	}
	if ont := AnalyzeHTTP("", "GPON ONT Web Management", "", ""); !ont.ONT {
		t.Fatalf("expected ONT detection, got %#v", ont)
	}
	if realm := AnalyzeHTTP("", "", "RouterOS", ""); !realm.AdminPanel {
		t.Fatal("a WWW-Authenticate realm must flag an admin panel")
	}
	if benign := AnalyzeHTTP("nginx", "Welcome", "", "hello world"); benign.Vendor != "" || benign.RouterLike {
		t.Fatalf("benign page must not be flagged router-like: %#v", benign)
	}
}

func TestAnalyzeRouting(t *testing.T) {
	if r := AnalyzeRouting(Facts{IsPrivate: true, InGatewayChain: true, DoubleNATHint: true, ReachableTCP: true}); r.Kind != "double_nat_upstream" || !r.DoubleNAT {
		t.Fatalf("double nat routing = %#v", r)
	}
	if r := AnalyzeRouting(Facts{IsPrivate: true, InGatewayChain: true}); r.Kind != "unreachable_inferred" {
		t.Fatalf("unreachable routing = %#v", r)
	}
	if r := AnalyzeRouting(Facts{IsPrivate: true, InGatewayChain: true, ReachableICMP: true}); r.Kind != "upstream_private_gateway" {
		t.Fatalf("upstream routing = %#v", r)
	}
}

func TestParsePing_Windows(t *testing.T) {
	out := `Pinging 192.168.1.1 with 32 bytes of data:
Reply from 192.168.1.1: bytes=32 time=2ms TTL=64
Reply from 192.168.1.1: bytes=32 time=1ms TTL=64

Ping statistics for 192.168.1.1:
    Packets: Sent = 2, Received = 2, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 1ms, Maximum = 2ms, Average = 1ms`
	r := ParsePing(out)
	if !r.Reachable || r.TTL == nil || *r.TTL != 64 || r.AvgMs == nil || *r.AvgMs != 1 {
		t.Fatalf("windows parse = %#v ttl=%v avg=%v", r, r.TTL, r.AvgMs)
	}
	if r.LossPct == nil || *r.LossPct != 0 {
		t.Fatalf("expected 0%% loss, got %v", r.LossPct)
	}
}

func TestParsePing_Unix(t *testing.T) {
	out := `PING 192.168.1.1 (192.168.1.1): 56 data bytes
64 bytes from 192.168.1.1: icmp_seq=0 ttl=64 time=1.234 ms
64 bytes from 192.168.1.1: icmp_seq=1 ttl=64 time=2.345 ms

--- 192.168.1.1 ping statistics ---
2 packets transmitted, 2 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 1.234/1.789/2.345/0.555 ms`
	r := ParsePing(out)
	if !r.Reachable || r.AvgMs == nil || *r.AvgMs != 1.789 || r.MinMs == nil || *r.MinMs != 1.234 {
		t.Fatalf("unix parse = %#v", r)
	}
}

func TestParsePing_Unreachable(t *testing.T) {
	out := `Pinging 192.168.1.1 with 32 bytes of data:
Request timed out.
Request timed out.

Ping statistics for 192.168.1.1:
    Packets: Sent = 2, Received = 0, Lost = 2 (100% loss),`
	r := ParsePing(out)
	if r.Reachable {
		t.Fatalf("100%% loss must be unreachable: %#v", r)
	}
}
