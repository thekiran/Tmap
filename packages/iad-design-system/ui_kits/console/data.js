/* IAD UI kit — a single realistic NormalizedScanReport sample.
   This is a recreation fixture for the kit's screens, NOT production data.
   Models the shape described in docs/design/07-api-design.md: an honest scan
   that knows a lot about the LAN edge but is calibrated about the physical
   medium. Exposed as window.IAD_SCAN. */
window.IAD_SCAN = {
  scan_id: 'scan_01HV6Q8M2K',
  created_at: '2026-06-16T09:41:22Z',
  status: 'complete',
  mode: 'standard',           // quick | standard | deep
  duration_ms: 6840,
  safe_mode: true,

  // ---- Headline decision -------------------------------------------------
  primary_type: 'Fiber (FTTH)',
  category: 'fixed_broadband',
  confidence: 0.74,                 // overall
  classification_confidence: 0.71,  // which access type
  context_confidence: 0.83,         // network context certainty
  decision_quality: 'medium',       // low | medium | high
  uncertainty_reasons: [
    'Physical medium cannot be confirmed from inside the LAN — no DSL/DOCSIS modem stats exposed.',
    'CPE management interface (192.168.100.1) did not respond to SNMP.',
    'Downstream/upstream symmetry is consistent with fiber but also with some cable plans.',
  ],

  candidates: [
    { type: 'Fiber (FTTH)', score: 0.74, note: 'Low latency, symmetric throughput, fiber-typical jitter floor.' },
    { type: 'Cable (DOCSIS)', score: 0.41, note: 'Cannot rule out — no modem telemetry to confirm or deny.' },
    { type: 'Fixed Wireless (FWA)', score: 0.12, note: 'Latency too low and too stable for typical FWA.' },
    { type: 'VDSL', score: 0.06, note: 'Throughput exceeds VDSL2 profile ceilings.' },
  ],

  // ---- Network context ---------------------------------------------------
  detected_network_context: {
    selected_interface: { name: 'Ethernet', type: 'ethernet', ipv4: '192.168.1.24', prefix: 24, mac: '9C:30:5B:A1:4F:02', mtu: 1500, gateway: '192.168.1.1', dns: ['192.168.1.1', '1.1.1.1'] },
    link_speed_mbps: 1000,
    dhcp: true,
  },
  gateway_chain: [
    { hop: 1, ip: '192.168.1.1', kind: 'default_gateway', rtt_ms: 1.1, label: 'Home router', private: true },
    { hop: 2, ip: '100.64.12.1', kind: 'upstream_private_gateway', rtt_ms: 8.7, label: 'CGNAT gateway', private: true, note: 'RFC 6598 shared address space' },
    { hop: 3, ip: '203.0.113.1', kind: 'isp_gateway', rtt_ms: 11.4, label: 'ISP edge', private: false },
  ],
  nat_topology: { type: 'cgnat', layers: 2, public_reachable: false, note: 'Double NAT detected: local NAT behind carrier-grade NAT (100.64/10).' },
  ipv6_context: { available: true, global_address: '2a01:598:8000::5b2a', delegated_prefix: '2a01:598:8000::/56', note: 'Native dual-stack; IPv6 path avoids CGNAT.' },
  public_ip: { address: '203.0.113.42', ptr: 'cpe-203-0-113-42.cust.example-isp.net', asn: 'AS3320', org: 'Example ISP GmbH', city: 'Frankfurt', country: 'DE', geo_confidence: 0.6 },
  performance: { downstream_mbps: 412.6, upstream_mbps: 198.3, latency_ms: 11.4, jitter_ms: 0.8, loss_pct: 0.0 },

  // ---- Confidence breakdown ---------------------------------------------
  confidence_breakdown: [
    { factor: 'Latency & jitter profile', weight: 0.28, contribution: 0.24, direction: 'up', detail: 'Sub-12ms, jitter <1ms — fiber-typical.' },
    { factor: 'Symmetric throughput', weight: 0.22, contribution: 0.16, direction: 'up', detail: '412↓/198↑ Mbps; high upstream favors fiber.' },
    { factor: 'No modem telemetry', weight: 0.20, contribution: -0.12, direction: 'down', detail: 'CPE SNMP blocked; physical layer unverified.' },
    { factor: 'ASN / ISP profile', weight: 0.18, contribution: 0.11, direction: 'up', detail: 'AS3320 deploys predominantly fiber in this region.' },
    { factor: 'CGNAT presence', weight: 0.12, contribution: -0.03, direction: 'down', detail: 'Common across access types; weak signal.' },
  ],

  next_best_probes: [
    { name: 'CPE SNMP walk', gain: 0.18, requires: 'CPE credentials or read community', tier: 'physical', detail: 'Would expose DOCSIS/GPON line stats and confirm medium.' },
    { name: 'TR-069 / management VLAN', gain: 0.12, requires: 'ISP management access', tier: 'physical' },
    { name: 'Sustained throughput test', gain: 0.06, requires: 'User consent (data usage)', tier: 'performance' },
  ],

  warnings: [
    { level: 'warn', text: 'Double NAT (CGNAT) — inbound connections will not reach this host.' },
    { level: 'info', text: 'IPv6 is native and bypasses CGNAT; prefer it for reachability.' },
  ],

  // ---- Devices (LAN inventory) ------------------------------------------
  devices: [
    { id: 'd-host', ip: '192.168.1.24', mac: '9C:30:5B:A1:4F:02', vendor: 'Dell Inc.', hostname: 'WS-OPS-14', type: 'local_host', role: 'This host', reachability: 'self', confidence: 1.0, source: 'interface', services: ['—'] },
    { id: 'd-gw', ip: '192.168.1.1', mac: 'F0:9F:C2:1A:88:E0', vendor: 'AVM GmbH', hostname: 'fritz.box', type: 'default_gateway', role: 'Router / NAT', reachability: 'reachable', confidence: 0.98, source: 'arp', services: ['DNS', 'HTTP', 'HTTPS'] },
    { id: 'd-ap', ip: '192.168.1.2', mac: 'F0:9F:C2:1A:88:E1', vendor: 'AVM GmbH', hostname: 'repeater-og', type: 'access_point', role: 'Mesh Wi-Fi', reachability: 'reachable', confidence: 0.86, source: 'mdns', services: ['HTTP'] },
    { id: 'd-nas', ip: '192.168.1.30', mac: '00:11:32:7C:A9:01', vendor: 'Synology', hostname: 'nas-archive', type: 'server', role: 'File / NAS', reachability: 'reachable', confidence: 0.93, source: 'mdns', services: ['SMB', 'HTTPS', 'NFS'] },
    { id: 'd-print', ip: '192.168.1.41', mac: '3C:2A:F4:11:0D:7B', vendor: 'Brother', hostname: 'HL-L2350DW', type: 'printer', role: 'Printer', reachability: 'reachable', confidence: 0.81, source: 'mdns', services: ['IPP', 'HTTP'] },
    { id: 'd-phone', ip: '192.168.1.57', mac: 'A4:83:E7:5F:22:9C', vendor: 'Apple, Inc.', hostname: 'iphone-k', type: 'mobile', role: 'Mobile', reachability: 'reachable', confidence: 0.64, source: 'arp', services: [] },
    { id: 'd-iot', ip: '192.168.1.88', mac: 'D8:A0:11:42:6E:33', vendor: 'Espressif', hostname: '—', type: 'iot', role: 'IoT (sensor?)', reachability: 'partial', confidence: 0.38, source: 'arp', services: ['?'] },
    { id: 'd-unknown', ip: '192.168.1.103', mac: '5E:B2:9A:00:1F:44', vendor: '— (locally administered)', hostname: '—', type: 'unknown', role: 'Unknown', reachability: 'partial', confidence: 0.22, source: 'arp', services: [] },
  ],

  // ---- Evidence / probes -------------------------------------------------
  evidence: [
    { id: 'p-iface', probe_name: 'interface_enum', status: 'success', confidence: 1.0, ts: '09:41:16', evidence_class: 'l3', reason: 'Enumerated active interfaces and addresses.', limitations: 'OS-reported; does not confirm physical medium.', data: { interfaces: 2, selected: 'Ethernet', ipv4: '192.168.1.24/24', mac: '9C:30:5B:A1:4F:02' } },
    { id: 'p-arp', probe_name: 'arp_sweep', status: 'success', confidence: 0.92, ts: '09:41:17', evidence_class: 'l2', reason: 'ARP-resolved 8 hosts on local subnet.', limitations: 'Only reaches the local broadcast domain.', data: { subnet: '192.168.1.0/24', responded: 8, mac_table_seen: false } },
    { id: 'p-gw', probe_name: 'gateway_trace', status: 'success', confidence: 0.88, ts: '09:41:18', evidence_class: 'l3', reason: 'Traced 3-hop gateway chain to ISP edge.', limitations: 'Hops are L3 routers, not physical switches.', data: { hops: 3, cgnat: true } },
    { id: 'p-dns', probe_name: 'public_ip_dns', status: 'success', confidence: 0.9, ts: '09:41:19', evidence_class: 'isp', reason: 'Resolved public IP, PTR, and ASN.', limitations: 'Geo from ASN registry; city-level only.', data: { ip: '203.0.113.42', asn: 'AS3320', ptr: 'cpe-203-0-113-42.cust.example-isp.net' } },
    { id: 'p-snmp', probe_name: 'cpe_snmp', status: 'blocked', confidence: 0.0, ts: '09:41:20', evidence_class: 'physical', reason: 'CPE management host did not respond to SNMP (timeout).', limitations: 'Cannot read DOCSIS/GPON line stats; medium stays inferred.', data: { target: '192.168.100.1', community: 'public', result: 'timeout' } },
    { id: 'p-lldp', probe_name: 'lldp_listen', status: 'no_data', confidence: 0.0, ts: '09:41:20', evidence_class: 'l2', reason: 'No LLDP/CDP frames observed in capture window.', limitations: 'Consumer gear rarely emits LLDP; absence is not evidence.', data: { window_s: 4, frames: 0 } },
    { id: 'p-perf', probe_name: 'perf_sample', status: 'partial', confidence: 0.6, ts: '09:41:21', evidence_class: 'performance', reason: 'Short latency/throughput sample taken.', limitations: 'Burst sample, not sustained; throughput is a lower bound.', data: { down_mbps: 412.6, up_mbps: 198.3, latency_ms: 11.4 } },
    { id: 'p-ipv6', probe_name: 'ipv6_probe', status: 'success', confidence: 0.85, ts: '09:41:21', evidence_class: 'l3', reason: 'Native IPv6 GUA + /56 delegation observed.', limitations: null, data: { gua: '2a01:598:8000::5b2a', prefix: '/56' } },
    { id: 'p-mdns', probe_name: 'mdns_discovery', status: 'success', confidence: 0.78, ts: '09:41:21', evidence_class: 'l2', reason: 'Discovered service records for 4 hosts.', limitations: 'mDNS scope is the local link only.', data: { hosts: 4, services: ['_ipp', '_smb', '_http', '_raop'] } },
    { id: 'p-portscan', probe_name: 'service_probe', status: 'skipped', confidence: 0.0, ts: '—', evidence_class: 'l3', reason: 'Skipped in standard mode (safe mode).', limitations: 'Enable deep mode to fingerprint services.', data: { reason: 'safe_mode' } },
  ],

  // ---- Topology (conservative, generated from context) -------------------
  topology: {
    generated: true,   // not natively provided; derived from context + evidence
    layers: { l2: true, l3: true, nat: true, isp_route_context: true, unknown: true },
  },
};
