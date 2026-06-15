package discovery

import (
	"context"
	"fmt"
	"sort"

	"github.com/thekiran/iad/internal/topology"
	"github.com/thekiran/iad/pkg/models"
)

// devID returns the stable device id for an IP.
func devID(ip string) string { return "dev-" + ip }

// commonServiceName maps well-known TCP ports to a service label. Empty when
// unknown; the agent never invents a product/version it did not observe.
var commonServiceName = map[int]string{
	22: "ssh", 23: "telnet", 25: "smtp", 53: "domain", 80: "http", 110: "pop3",
	135: "msrpc", 139: "netbios-ssn", 143: "imap", 443: "https", 445: "microsoft-ds",
	587: "submission", 993: "imaps", 995: "pop3s", 1883: "mqtt", 3389: "ms-wbt-server",
	5000: "upnp", 5060: "sip", 5357: "wsdapi", 7547: "cwmp", 8080: "http-proxy",
	8443: "https-alt", 8843: "https-alt", 9000: "http-alt", 49152: "upnp",
}

// Normalizer turns raw discovery observations into evidence-backed devices. It
// mints evidence through the shared store so every device fact is auditable.
type Normalizer struct {
	Store    *topology.EvidenceStore
	Resolver Resolver
}

// Device builds a normalized device for an address. alive/openPorts come from the
// sweep (may be empty for a device known only from ARP or as the gateway); mac
// comes from the ARP table when available; isAgent/isGateway are set by the
// caller from proven facts. extraEvidence (e.g. a gateway_route id) is appended.
func (n *Normalizer) Device(ctx context.Context, ip string, alive bool, openPorts []int, mac string, isAgent, isGateway bool, extraEvidence []string) models.Device {
	d := models.Device{
		ID:        devID(ip),
		Addresses: []models.IPAddress{ipAddress(ip)},
		IsAgent:   isAgent,
		IsGateway: isGateway,
	}
	evIDs := append([]string(nil), extraEvidence...)

	// Liveness / services from the TCP sweep.
	if alive && len(openPorts) > 0 {
		ports := make([]string, 0, len(openPorts))
		for _, p := range openPorts {
			ports = append(ports, itoa(p))
		}
		evID := n.Store.Add(topology.EvidenceTCPConnect, "tcp_sweep",
			fmt.Sprintf("%s accepted TCP connections on %d port(s).", ip, len(openPorts)),
			map[string]string{"ip": ip, "ports": joinStrings(ports, ",")})
		evIDs = append(evIDs, evID)
		for _, p := range openPorts {
			d.Services = append(d.Services, models.Service{
				Port: p, Protocol: "tcp", State: "open",
				Name: commonServiceName[p], EvidenceIDs: []string{evID},
			})
		}
	}

	// MAC from the ARP table.
	if mac != "" {
		evID := n.Store.Add(topology.EvidenceARPTable, "arp_table",
			fmt.Sprintf("%s resolved to MAC %s in the neighbour table.", ip, mac),
			map[string]string{"ip": ip, "mac": mac})
		evIDs = append(evIDs, evID)
		d.Interfaces = append(d.Interfaces, models.DeviceInterface{MAC: mac, IPs: []string{ip}})
	}

	// Reverse DNS hostname.
	if n.Resolver != nil {
		if host := n.Resolver.LookupAddr(ctx, ip); host != "" {
			evID := n.Store.Add(topology.EvidenceReverseDNS, "reverse_dns",
				fmt.Sprintf("%s reverse-resolves to %s.", ip, host),
				map[string]string{"ip": ip, "hostname": host})
			evIDs = append(evIDs, evID)
			d.Hostname = host
		}
	}

	d.EvidenceIDs = sortedUnique(evIDs)
	d.Confidence = deviceConfidence(d, alive)
	return d
}

// attachNmap enriches a device with Nmap host/service data: it records an nmap
// evidence entry, merges open services (deduped by port+protocol), and fills in a
// hostname when reverse DNS did not provide one.
func (n *Normalizer) attachNmap(d *models.Device, s ScannedHost) {
	evID := n.Store.Add(topology.EvidenceNmap, "nmap",
		fmt.Sprintf("Nmap reported %d open service(s) on %s.", len(s.Services), s.IP),
		map[string]string{"ip": s.IP, "services": itoa(len(s.Services))})

	existing := map[string]bool{}
	for _, sv := range d.Services {
		existing[sv.Protocol+":"+itoa(sv.Port)] = true
	}
	for _, sv := range s.Services {
		key := sv.Protocol + ":" + itoa(sv.Port)
		if existing[key] {
			continue
		}
		existing[key] = true
		sv.EvidenceIDs = append(sv.EvidenceIDs, evID)
		d.Services = append(d.Services, sv)
	}
	sort.Slice(d.Services, func(i, j int) bool {
		if d.Services[i].Port != d.Services[j].Port {
			return d.Services[i].Port < d.Services[j].Port
		}
		return d.Services[i].Protocol < d.Services[j].Protocol
	})
	if d.Hostname == "" && s.Hostname != "" {
		d.Hostname = s.Hostname
	}
	d.EvidenceIDs = sortedUnique(append(d.EvidenceIDs, evID))
}

func ipAddress(ip string) models.IPAddress {
	v := 4
	if isIPv6(ip) {
		v = 6
	}
	return models.IPAddress{IP: ip, Version: v}
}

func deviceConfidence(d models.Device, alive bool) float64 {
	c := 0.0
	switch {
	case d.IsAgent:
		c = 0.95 // the agent knows itself
	case d.IsGateway:
		c = 0.90 // proven via the routing table
	case alive:
		c = 0.70 // responded to a TCP connect
	default:
		c = 0.50 // known only from the ARP table
	}
	if len(d.Interfaces) > 0 {
		c += 0.05
	}
	if d.Hostname != "" {
		c += 0.05
	}
	if c > 1 {
		c = 1
	}
	return c
}

func isIPv6(ip string) bool {
	for i := 0; i < len(ip); i++ {
		if ip[i] == ':' {
			return true
		}
	}
	return false
}

func sortedUnique(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func joinStrings(parts []string, sep string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}
