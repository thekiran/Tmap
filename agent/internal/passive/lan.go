package passive

import (
	"context"
	"net"
	"sort"
	"strings"
	"time"
)

// LANObserver collects local-link metadata only. Implementations may be backed
// by pcap, ETW, AF_PACKET, eBPF, or OS-specific event streams, but must not
// retain payload bodies.
type LANObserver interface {
	Observe(ctx context.Context, iface string, sink ObservationSink) error
}

type ObservationSink interface {
	Record(obs LANObservation)
}

// LANObservation is one metadata-only packet/event observation.
type LANObservation struct {
	Interface      string
	SourceMAC      string
	DestinationMAC string
	EtherType      uint16
	SourceIP       string
	DestinationIP  string
	Protocol       string
	Discovery      string // arp, mdns, ssdp, dhcp, llmnr, nbns, other
	Direction      string // inbound, outbound, local-broadcast, local-multicast, unknown
	Timestamp      time.Time
}

type HostObservation struct {
	Interface        string    `json:"interface,omitempty"`
	MAC              string    `json:"mac,omitempty"`
	IPAddresses      []string  `json:"ip_addresses,omitempty"`
	Protocols        []string  `json:"protocols,omitempty"`
	DiscoverySources []string  `json:"discovery_sources,omitempty"`
	PacketCount      int       `json:"packet_count"`
	FirstSeen        time.Time `json:"first_seen,omitempty"`
	LastSeen         time.Time `json:"last_seen,omitempty"`
	DirectionHint    string    `json:"direction_hint,omitempty"`
}

// Collector aggregates metadata observations into host-level hints.
type Collector struct {
	byKey map[string]*HostObservation
}

func NewCollector() *Collector {
	return &Collector{byKey: map[string]*HostObservation{}}
}

func (c *Collector) Record(obs LANObservation) {
	if c.byKey == nil {
		c.byKey = map[string]*HostObservation{}
	}
	obs = normalizeObservation(obs)
	if obs.Timestamp.IsZero() {
		obs.Timestamp = time.Now().UTC()
	}
	for _, endpoint := range observationEndpoints(obs) {
		key := endpointKey(endpoint.mac, endpoint.ip)
		if key == "" {
			continue
		}
		h := c.byKey[key]
		if h == nil {
			h = &HostObservation{MAC: endpoint.mac, Interface: obs.Interface, FirstSeen: obs.Timestamp}
			c.byKey[key] = h
		}
		if h.Interface == "" {
			h.Interface = obs.Interface
		}
		h.IPAddresses = appendUnique(h.IPAddresses, endpoint.ip)
		h.Protocols = appendUnique(h.Protocols, obs.Protocol)
		h.DiscoverySources = appendUnique(h.DiscoverySources, obs.Discovery)
		h.PacketCount++
		if h.FirstSeen.IsZero() || obs.Timestamp.Before(h.FirstSeen) {
			h.FirstSeen = obs.Timestamp
		}
		if obs.Timestamp.After(h.LastSeen) {
			h.LastSeen = obs.Timestamp
		}
		if h.DirectionHint == "" || h.DirectionHint == "unknown" {
			h.DirectionHint = obs.Direction
		}
	}
}

func (c *Collector) Hosts() []HostObservation {
	out := make([]HostObservation, 0, len(c.byKey))
	for _, h := range c.byKey {
		cp := *h
		sort.Strings(cp.IPAddresses)
		sort.Strings(cp.Protocols)
		sort.Strings(cp.DiscoverySources)
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].MAC != out[j].MAC {
			return out[i].MAC < out[j].MAC
		}
		return strings.Join(out[i].IPAddresses, ",") < strings.Join(out[j].IPAddresses, ",")
	})
	return out
}

type endpoint struct {
	mac string
	ip  string
}

func normalizeObservation(obs LANObservation) LANObservation {
	if obs.Direction == "" {
		obs.Direction = directionHint(obs)
	}
	obs.SourceMAC = normalizeMAC(obs.SourceMAC)
	obs.DestinationMAC = normalizeMAC(obs.DestinationMAC)
	obs.SourceIP = strings.TrimSpace(obs.SourceIP)
	obs.DestinationIP = strings.TrimSpace(obs.DestinationIP)
	obs.Protocol = strings.ToLower(strings.TrimSpace(obs.Protocol))
	obs.Discovery = strings.ToLower(strings.TrimSpace(obs.Discovery))
	if obs.Discovery == "" {
		obs.Discovery = discoveryKind(obs)
	}
	return obs
}

func observationEndpoints(obs LANObservation) []endpoint {
	return []endpoint{
		{mac: obs.SourceMAC, ip: obs.SourceIP},
		{mac: obs.DestinationMAC, ip: obs.DestinationIP},
	}
}

func endpointKey(mac, ip string) string {
	if mac != "" {
		return "mac:" + mac
	}
	if ip == "" || isBroadcastIP(ip) || isMulticastIP(ip) {
		return ""
	}
	if ip != "" {
		return "ip:" + ip
	}
	return ""
}

func discoveryKind(obs LANObservation) string {
	isUDP := obs.Protocol == "udp" || strings.HasPrefix(obs.Protocol, "udp/")
	switch {
	case obs.EtherType == 0x0806:
		return "arp"
	case isUDP && (isPortDiscovery(obs, "5353") || multicastIP(obs.DestinationIP, "224.0.0.251", "ff02::fb")):
		return "mdns"
	case isUDP && (isPortDiscovery(obs, "1900") || multicastIP(obs.DestinationIP, "239.255.255.250", "ff02::c")):
		return "ssdp"
	case isUDP && isPortDiscovery(obs, "5355"):
		return "llmnr"
	case isUDP && isPortDiscovery(obs, "137"):
		return "nbns"
	case isUDP && (isPortDiscovery(obs, "67") || isPortDiscovery(obs, "68")):
		return "dhcp"
	default:
		return "other"
	}
}

func directionHint(obs LANObservation) string {
	if isBroadcastMAC(obs.DestinationMAC) || isBroadcastIP(obs.DestinationIP) {
		return "local-broadcast"
	}
	if isMulticastMAC(obs.DestinationMAC) || isMulticastIP(obs.DestinationIP) {
		return "local-multicast"
	}
	return "unknown"
}

func normalizeMAC(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(raw, "-", ":")))
	if raw == "" || raw == "ff:ff:ff:ff:ff:ff" || raw == "00:00:00:00:00:00" {
		return ""
	}
	return raw
}

func appendUnique(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func isPortDiscovery(obs LANObservation, port string) bool {
	return strings.Contains(obs.Protocol, ":"+port) || strings.Contains(obs.Protocol, "/"+port)
}

func multicastIP(ip string, values ...string) bool {
	for _, value := range values {
		if strings.EqualFold(ip, value) {
			return true
		}
	}
	return false
}

func isBroadcastMAC(mac string) bool { return mac == "ff:ff:ff:ff:ff:ff" }

func isMulticastMAC(mac string) bool {
	return strings.HasPrefix(mac, "01:00:5e:") || strings.HasPrefix(mac, "33:33:")
}

func isBroadcastIP(ip string) bool { return ip == "255.255.255.255" }

func isMulticastIP(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil && parsed.IsMulticast()
}
