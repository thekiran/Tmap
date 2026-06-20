package passive

import (
	"testing"
	"time"
)

func TestCollectorAggregatesMetadataOnlyObservations(t *testing.T) {
	now := time.Unix(10, 0).UTC()
	c := NewCollector()

	c.Record(LANObservation{
		Interface:      "Ethernet",
		SourceMAC:      "AA-BB-CC-00-00-01",
		DestinationMAC: "ff:ff:ff:ff:ff:ff",
		EtherType:      0x0806,
		SourceIP:       "192.168.1.20",
		DestinationIP:  "192.168.1.1",
		Protocol:       "arp",
		Timestamp:      now,
	})
	c.Record(LANObservation{
		Interface:     "Ethernet",
		SourceMAC:     "aa:bb:cc:00:00:01",
		SourceIP:      "192.168.1.20",
		DestinationIP: "224.0.0.251",
		Protocol:      "udp/5353",
		Timestamp:     now.Add(time.Second),
	})

	hosts := c.Hosts()
	if len(hosts) != 2 {
		t.Fatalf("hosts = %d, want source host plus gateway IP hint: %#v", len(hosts), hosts)
	}

	var source HostObservation
	for _, h := range hosts {
		if h.MAC == "aa:bb:cc:00:00:01" {
			source = h
		}
	}
	if source.PacketCount != 2 {
		t.Fatalf("source packet count = %d, want 2: %#v", source.PacketCount, source)
	}
	if source.DirectionHint != "local-broadcast" {
		t.Fatalf("direction hint = %q, want local-broadcast", source.DirectionHint)
	}
	if !contains(source.DiscoverySources, "arp") || !contains(source.DiscoverySources, "mdns") {
		t.Fatalf("discovery sources = %#v, want arp and mdns", source.DiscoverySources)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
