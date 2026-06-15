package discovery

import (
	"context"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/thekiran/iad/internal/topology"
	"github.com/thekiran/iad/pkg/models"
)

// ServiceScanner is an optional richer host/service scanner (e.g. an Nmap
// adapter). It is defined here so discovery does not depend on the nmap package;
// the caller wires an implementation in. nil disables it.
type ServiceScanner interface {
	Available() bool
	Scan(ctx context.Context, scope models.ScanScope) ([]ScannedHost, error)
}

// ScannedHost is a host returned by a ServiceScanner.
type ScannedHost struct {
	IP       string
	MAC      string
	Hostname string
	Services []models.Service
}

// Options configures a discovery run. The collaborator fields are injectable so
// the pipeline is testable without touching the real network.
type Options struct {
	RequestedCIDR  string // "auto" or a CIDR
	Profile        string // quick | standard | deep
	AllowPublic    bool   // dangerous: permit non-private scope
	InterfaceName  string // explicit interface (empty → auto-select)
	IncludeVirtual bool   // include virtual/loopback adapters

	Version  string
	Hostname string
	OS       string

	// Collaborators (defaults used when nil).
	Sweeper      Sweeper
	ARPReader    ARPReader
	Resolver     Resolver
	Service      ServiceScanner          // optional (Nmap)
	GatewayFn    func() (net.IP, error)  // default: network.Gateway
	DNSFn        func(context.Context) ([]string, error)
	GatewayChain []string                // optional private gateway chain (gw, upstream, ...) for route_hop edges
	Now          func() time.Time
}

// Run executes the discovery + topology pipeline and returns a complete,
// evidence-based ScanReport. A refused scope returns a *ScopeError.
func Run(ctx context.Context, opts Options) (models.ScanReport, error) {
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	start := now()
	store := topology.NewEvidenceStore(now)

	// 1. Interfaces + gateway.
	ifaces, err := Interfaces()
	if err != nil {
		return models.ScanReport{}, fmt.Errorf("listing interfaces: %w", err)
	}
	gatewayIP := ""
	if gw := opts.gateway(); gw != nil {
		gatewayIP = gw.String()
	}

	var primary models.InterfaceInfo
	var ok bool
	if opts.InterfaceName != "" {
		primary, ifaces, ok = FindByName(ifaces, opts.InterfaceName)
		if !ok {
			return models.ScanReport{}, &ScopeError{Reason: fmt.Sprintf("interface %q not found", opts.InterfaceName)}
		}
	} else {
		primary, ifaces, ok = SelectPrimary(ifaces, gatewayIP, opts.IncludeVirtual)
		if !ok {
			return models.ScanReport{}, &ScopeError{Reason: "no suitable active, non-virtual interface with a private IPv4 network was found; pass --interface or --cidr"}
		}
	}

	// 2. Scope validation (the safety boundary).
	scope, warnings, err := ParseScope(opts.RequestedCIDR, opts.Profile, opts.AllowPublic, primary)
	if err != nil {
		return models.ScanReport{}, err
	}

	// 3. DNS servers (context only).
	var dnsServers []string
	if opts.DNSFn != nil {
		dnsServers, _ = opts.DNSFn(ctx)
	}

	agentIP := primaryIPv4(primary)
	store.Add(topology.EvidenceInterface, "interface_probe",
		fmt.Sprintf("Selected interface %s (%s) on %s.", primary.Name, agentIP, primary.CIDR),
		map[string]string{"interface": primary.Name, "cidr": primary.CIDR, "ip": agentIP})

	gatewayEvID := ""
	if gatewayIP != "" {
		gatewayEvID = store.Add(topology.EvidenceGatewayRoute, "gateway_probe",
			fmt.Sprintf("Default route points to gateway %s.", gatewayIP),
			map[string]string{"gateway": gatewayIP})
	} else {
		warnings = append(warnings, models.Warning{Code: warnNoGateway, Severity: models.SeverityInfo,
			Message: "No default gateway was detected; gateway/default edges will be absent."})
	}

	// 4. Host discovery (only meaningful for IPv4 scopes).
	hits, arpEntries := opts.discover(ctx, scope, store)

	// 5. Normalize into devices.
	norm := &Normalizer{Store: store, Resolver: opts.resolver()}
	devices, agentDevID, gatewayDevID := buildDevices(ctx, norm, scope, primary, agentIP, gatewayIP, gatewayEvID, hits, arpEntries)

	// 6. Build topology.
	l2peers := sameSubnetPeers(devices, primary.CIDR, agentDevID, gatewayDevID)
	routeHops := buildRouteHops(store, opts.GatewayChain, devices, gatewayDevID)
	build := topology.Build(topology.BuildInput{
		AgentID:                 agentDevID,
		GatewayID:               gatewayDevID,
		Devices:                 devices,
		GatewayRouteEvidenceIDs: nonEmpty(gatewayEvID),
		RouteHops:               routeHops,
		L2Peers:                 l2peers,
		// Neighbors stays empty: no LLDP/CDP/SNMP evidence here, so no claimed
		// physical adjacency (honours the "never fake topology" rule).
	})

	if build.InferredOnly && len(build.Edges) > 0 {
		warnings = append(warnings, models.Warning{Code: warnInferredOnly, Severity: models.SeverityInfo,
			Message: "Topology is inferred from routing and subnet membership only. No LLDP/CDP/SNMP evidence was available, so physical links are not proven."})
	}

	evidence := store.Records()
	report := models.ScanReport{
		SchemaVersion: models.TopologyReportSchema,
		ScanID:        "topo_" + start.Format("20060102_150405"),
		CreatedAt:     start.UTC(),
		Agent: models.AgentInfo{
			Version: opts.Version, Hostname: opts.Hostname, OS: opts.OS,
			Gateway: gatewayIP, DNSServers: dnsServers, Interfaces: ifaces,
		},
		Scope:    scope,
		Devices:  build.Devices,
		Edges:    build.Edges,
		Evidence: evidence,
		Warnings: warnings,
		Summary: models.ScanSummary{
			DeviceCount: len(build.Devices), EdgeCount: len(build.Edges),
			EvidenceCount: len(evidence), HighConfidenceEdges: build.HighConfidenceEdges,
			InferredOnly: build.InferredOnly, Profile: scope.Profile,
			DurationMS: now().Sub(start).Milliseconds(),
		},
	}
	return report, nil
}

// discover runs the sweep + ARP read (+ optional service scan), recording the
// liveness evidence implicitly through the normalizer later.
func (opts Options) discover(ctx context.Context, scope models.ScanScope, store *topology.EvidenceStore) ([]HostHit, []ARPEntry) {
	var hits []HostHit
	if scope.HostCount > 0 { // IPv4, in-range
		hits, _ = opts.sweeper().Sweep(ctx, scope)
	}
	// Optional richer scanner (Nmap) merges/extends liveness.
	if opts.Service != nil && opts.Service.Available() {
		if scanned, err := opts.Service.Scan(ctx, scope); err == nil {
			hits = mergeScanned(hits, scanned, store)
		}
	}
	arpEntries, _ := opts.arpReader().Read(ctx)
	// Keep only ARP entries inside scope (never leak neighbours from other nets).
	var inScope []ARPEntry
	for _, e := range arpEntries {
		if InScope(scope, e.IP) {
			inScope = append(inScope, e)
		}
	}
	return hits, inScope
}

// buildDevices assembles the device list: the agent, the gateway (if known), and
// every discovered host, deduped by IP and ordered by IP for determinism.
func buildDevices(ctx context.Context, norm *Normalizer, scope models.ScanScope, primary models.InterfaceInfo, agentIP, gatewayIP, gatewayEvID string, hits []HostHit, arp []ARPEntry) ([]models.Device, string, string) {
	macByIP := map[string]string{}
	for _, e := range arp {
		macByIP[e.IP] = e.MAC
	}
	hitByIP := map[string]HostHit{}
	for _, h := range hits {
		hitByIP[h.IP] = h
	}

	ips := map[string]bool{}
	if agentIP != "" {
		ips[agentIP] = true
	}
	if gatewayIP != "" {
		ips[gatewayIP] = true
	}
	for _, h := range hits {
		ips[h.IP] = true
	}
	for _, e := range arp {
		ips[e.IP] = true
	}

	ordered := make([]string, 0, len(ips))
	for ip := range ips {
		ordered = append(ordered, ip)
	}
	sort.Slice(ordered, func(i, j int) bool { return ipLess(ordered[i], ordered[j]) })

	agentDevID, gatewayDevID := "", ""
	if agentIP != "" {
		agentDevID = devID(agentIP)
	}
	if gatewayIP != "" {
		gatewayDevID = devID(gatewayIP)
	}

	devices := make([]models.Device, 0, len(ordered))
	for _, ip := range ordered {
		h := hitByIP[ip]
		isAgent := ip == agentIP
		isGateway := ip == gatewayIP
		var extra []string
		if isGateway && gatewayEvID != "" {
			extra = []string{gatewayEvID}
		}
		mac := macByIP[ip]
		if isAgent && mac == "" {
			mac = primary.MAC
		}
		alive := h.Alive || isAgent || isGateway
		devices = append(devices, norm.Device(ctx, ip, alive, h.OpenPorts, mac, isAgent, isGateway, extra))
	}
	return devices, agentDevID, gatewayDevID
}

// sameSubnetPeers returns device IDs (excluding agent and gateway) whose IP is in
// the agent interface's network — candidates for inferred_l2 edges.
func sameSubnetPeers(devices []models.Device, cidr, agentID, gatewayID string) []string {
	if cidr == "" {
		return nil
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}
	var peers []string
	for _, d := range devices {
		if d.ID == agentID || d.ID == gatewayID {
			continue
		}
		for _, a := range d.Addresses {
			if ip := net.ParseIP(a.IP); ip != nil && ipnet.Contains(ip) {
				peers = append(peers, d.ID)
				break
			}
		}
	}
	return peers
}

// buildRouteHops turns a private gateway chain into route_hop links between
// consecutive hops, minting an evidence record for the chain.
func buildRouteHops(store *topology.EvidenceStore, chain []string, devices []models.Device, gatewayDevID string) []topology.RouteHop {
	if len(chain) < 2 {
		return nil
	}
	known := map[string]bool{}
	for _, d := range devices {
		known[d.ID] = true
	}
	evID := store.Add(topology.EvidenceGatewayRoute, "gateway_chain",
		"A multi-hop private gateway chain was observed (possible router-behind-router).",
		map[string]string{"chain": joinStrings(chain, " -> ")})
	var hops []topology.RouteHop
	for i := 0; i+1 < len(chain); i++ {
		from, to := devID(chain[i]), devID(chain[i+1])
		if !known[from] || !known[to] {
			continue
		}
		hops = append(hops, topology.RouteHop{FromID: from, ToID: to, EvidenceIDs: []string{evID}})
	}
	return hops
}

func mergeScanned(hits []HostHit, scanned []discoveryScannedHostAlias, store *topology.EvidenceStore) []HostHit {
	byIP := map[string]*HostHit{}
	out := make([]HostHit, 0, len(hits)+len(scanned))
	for i := range hits {
		out = append(out, hits[i])
		byIP[hits[i].IP] = &out[len(out)-1]
	}
	for _, s := range scanned {
		if _, ok := byIP[s.IP]; ok {
			continue
		}
		out = append(out, HostHit{IP: s.IP, Alive: true})
	}
	return out
}

// discoveryScannedHostAlias keeps mergeScanned readable; it mirrors ScannedHost.
type discoveryScannedHostAlias = ScannedHost

// ---- option defaults --------------------------------------------------------

func (opts Options) sweeper() Sweeper {
	if opts.Sweeper != nil {
		return opts.Sweeper
	}
	return TCPSweeper{Profile: opts.Profile}
}

func (opts Options) arpReader() ARPReader {
	if opts.ARPReader != nil {
		return opts.ARPReader
	}
	return OSARPReader{}
}

func (opts Options) resolver() Resolver {
	if opts.Resolver != nil {
		return opts.Resolver
	}
	return NetResolver{}
}

func (opts Options) gateway() net.IP {
	if opts.GatewayFn == nil {
		return nil
	}
	ip, err := opts.GatewayFn()
	if err != nil {
		return nil
	}
	return ip
}

func nonEmpty(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}

// ipLess orders IPv4 numerically and falls back to string order otherwise.
func ipLess(a, b string) bool {
	ai, aok := ipv4ToUint(a)
	bi, bok := ipv4ToUint(b)
	if aok && bok {
		return ai < bi
	}
	return a < b
}

func ipv4ToUint(s string) (uint32, bool) {
	ip := net.ParseIP(s)
	if ip == nil {
		return 0, false
	}
	v4 := ip.To4()
	if v4 == nil {
		return 0, false
	}
	return uint32(v4[0])<<24 | uint32(v4[1])<<16 | uint32(v4[2])<<8 | uint32(v4[3]), true
}
