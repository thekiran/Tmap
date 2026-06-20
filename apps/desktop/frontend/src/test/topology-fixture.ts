import type { RawScanReport } from '../lib/scan-schema';

type NodeSpec = {
  id: string;
  label: string;
  type: string;
  role: string;
  ip: string;
  mac: string;
  vendor: string;
  confidence?: number;
  reachability?: string;
  wireless?: Record<string, unknown>;
  services?: { port: number; protocol: string; state: string; name: string }[];
};

type EdgeSpec = {
  source: string;
  target: string;
  type: string;
  relation: string;
  medium: string;
  confidence: number;
};

const now = '2026-06-19T10:00:00Z';

function mac(index: number) {
  return `02:00:00:00:${String(Math.floor(index / 256)).padStart(2, '0')}:${String(index % 256).padStart(2, '0')}`;
}

function node(spec: NodeSpec, index: number) {
  return {
    id: spec.id,
    label: spec.label,
    type: spec.type,
    category: ['gateway', 'router', 'switch', 'access_point', 'mesh_node', 'repeater'].includes(spec.role) ? 'network' : 'device',
    device_role: spec.role,
    ip_addresses: [spec.ip],
    mac_addresses: [spec.mac],
    vendor: spec.vendor,
    hostname: spec.label,
    services: spec.services ?? [],
    wireless: spec.wireless,
    first_seen: now,
    last_seen: now,
    confidence: spec.confidence ?? 0.72,
    evidence: [{ source: 'arp', value: `${spec.ip} ${spec.mac}`, confidence: 0.7, timestamp: now }],
    raw_sources: ['arp_table', spec.wireless ? 'passive_wifi' : 'tcp'],
    ui: { reachability: spec.reachability ?? 'reachable', fixture_index: index },
  };
}

function device(spec: NodeSpec) {
  return {
    id: spec.id,
    hostname: spec.label,
    hostnames: [spec.label],
    vendor: spec.vendor,
    mac: spec.mac,
    roles: [spec.role],
    is_gateway: spec.role === 'gateway',
    addresses: [{ ip: spec.ip, version: 4 }],
    interfaces: [{ mac: spec.mac, vendor: spec.vendor, ips: [spec.ip] }],
    services: spec.services ?? [],
    reachability: spec.reachability ?? 'reachable',
    discovery_sources: ['arp_table', spec.wireless ? 'passive_wifi' : 'tcp'],
    first_seen: now,
    last_seen: now,
    confidence: spec.confidence ?? 0.72,
    evidence_ids: [`ev-${spec.id}`],
  };
}

function edge(spec: EdgeSpec, index: number) {
  return {
    id: `edge-${index}-${spec.source}-${spec.target}-${spec.type}`,
    source: spec.source,
    target: spec.target,
    type: spec.type,
    relation: spec.relation,
    medium: spec.medium,
    confidence: spec.confidence,
    evidence: [{ source: spec.type.startsWith('reported') ? 'router_api' : spec.medium === 'wireless' ? 'passive_wifi' : 'arp', value: spec.relation, confidence: spec.confidence, timestamp: now }],
    explanation: `${spec.type} fixture relationship`,
    warnings: spec.type.includes('inferred') || spec.type === 'weak-inferred' ? ['not a proven physical link'] : [],
    first_seen: now,
    last_seen: now,
  };
}

export function buildLargeTopologyFixture() {
  const specs: NodeSpec[] = [];
  let i = 1;
  const add = (role: string, type: string, label: string, extra: Partial<NodeSpec> = {}) => {
    const spec: NodeSpec = {
      id: `dev-${role}-${i}`,
      label,
      type,
      role,
      ip: `10.10.${Math.floor(i / 250)}.${(i % 250) + 1}`,
      mac: mac(i),
      vendor: `${role}-vendor`,
      ...extra,
    };
    specs.push(spec);
    i += 1;
    return spec;
  };

  const gateway = add('gateway', 'gateway', 'Gateway', { confidence: 0.96 });
  const router = add('router', 'router', 'Router', { confidence: 0.9 });
  const switches = [add('switch', 'managed_switch', 'Switch A'), add('switch', 'managed_switch', 'Switch B')];
  const aps = [add('access_point', 'access_point', 'AP 1'), add('access_point', 'access_point', 'AP 2'), add('access_point', 'access_point', 'AP 3')];
  const mesh = add('mesh_node', 'mesh_node', 'Mesh Node', { wireless: { ssid: 'IAD-Test', bssid: '02:00:00:aa:00:01', channel: 6 } });
  const repeater = add('repeater', 'repeater', 'Repeater', { wireless: { ssid: 'IAD-Test', bssid: '02:00:00:aa:00:02', channel: 6 } });
  const server = add('server', 'server', 'Server 1', { services: [{ port: 443, protocol: 'tcp', state: 'open', name: 'https' }] });
  const phone = add('phone', 'phone', 'Phone 1', { wireless: { ssid: 'IAD-Test', bssid: aps[0].mac, channel: 6 } });
  const wired = Array.from({ length: 20 }, (_, n) => add('wired_client', 'wired_client', `Wired ${n + 1}`));
  const wireless = Array.from({ length: 20 }, (_, n) => add('wireless_client', 'wireless_client', `Wireless ${n + 1}`, { wireless: { ssid: 'IAD-Test', bssid: aps[n % aps.length].mac, channel: 6 } }));
  const iot = Array.from({ length: 5 }, (_, n) => add('iot', 'iot', `IoT ${n + 1}`));
  const printers = Array.from({ length: 3 }, (_, n) => add('printer', 'printer', `Printer ${n + 1}`, { services: [{ port: 9100, protocol: 'tcp', state: 'open', name: 'jetdirect' }] }));
  const unknowns = Array.from({ length: 5 }, (_, n) => add('unknown', n === 0 ? 'mystery_sensor' : 'unknown', `Unknown ${n + 1}`, { confidence: 0.35, reachability: 'unknown' }));

  const edges: EdgeSpec[] = [
    { source: gateway.id, target: router.id, type: 'reported-by-router', relation: 'router_table', medium: 'l3', confidence: 0.9 },
    ...switches.map((sw) => ({ source: router.id, target: sw.id, type: 'switch-uplink', relation: 'uplink', medium: 'l2', confidence: 0.86 })),
    ...aps.map((ap, n) => ({ source: switches[n % switches.length].id, target: ap.id, type: 'reported-by-ap', relation: 'ap_uplink', medium: 'l2', confidence: 0.84 })),
    { source: aps[0].id, target: mesh.id, type: 'mesh-backhaul', relation: 'mesh_backhaul', medium: 'wireless', confidence: 0.78 },
    { source: aps[1].id, target: repeater.id, type: 'repeater-uplink', relation: 'repeater_uplink', medium: 'wireless', confidence: 0.74 },
    { source: switches[0].id, target: server.id, type: 'reported-by-router', relation: 'server_seen', medium: 'l3', confidence: 0.82 },
    { source: aps[0].id, target: phone.id, type: 'wireless-associated', relation: 'station', medium: 'wireless', confidence: 0.86 },
    ...wired.map((client, n) => ({ source: switches[n % switches.length].id, target: client.id, type: 'subnet-inferred', relation: 'same_subnet', medium: 'l2', confidence: 0.46 })),
    ...wireless.map((client, n) => ({ source: aps[n % aps.length].id, target: client.id, type: n % 2 === 0 ? 'wireless-associated' : 'wireless-observed', relation: 'station', medium: 'wireless', confidence: n % 2 === 0 ? 0.88 : 0.58 })),
    ...iot.map((client, n) => ({ source: aps[n % aps.length].id, target: client.id, type: 'wireless-observed', relation: 'iot_seen', medium: 'wireless', confidence: 0.55 })),
    ...printers.map((client, n) => ({ source: switches[n % switches.length].id, target: client.id, type: 'subnet-inferred', relation: 'same_subnet', medium: 'l2', confidence: 0.5 })),
    ...unknowns.map((client) => ({ source: gateway.id, target: client.id, type: 'weak-inferred', relation: 'weak_correlation', medium: 'l3', confidence: 0.25 })),
  ];

  const topologyNodes = specs.map(node);
  const topologyEdges = edges.map(edge);
  const report: RawScanReport = {
    schema_version: 'iad.topology/v2',
    scan_id: 'fixture-large-topology',
    created_at: now,
    agent: { version: 'test', capabilities: [] },
    scope: { profile: 'full', cidr: '10.10.0.0/16' },
    summary: { device_count: specs.length, edge_count: topologyEdges.length, evidence_count: specs.length },
    devices: specs.map(device),
    evidence: specs.map((spec) => ({ id: `ev-${spec.id}`, kind: 'fixture', source: 'fixture', summary: spec.label, timestamp: now, confidence: spec.confidence ?? 0.72 })),
    topology: {
      schema_version: 'iad.topology/v2',
      generated_at: now,
      root_id: gateway.id,
      nodes: topologyNodes,
      edges: topologyEdges,
      warnings: [],
    },
    warnings: [],
    wireless: specs.filter((spec) => spec.wireless).map((spec) => ({ station_mac: spec.mac, ...(spec.wireless ?? {}) })),
  };
  return { report, specs, topologyNodes, topologyEdges };
}

export function buildInvalidEdgeFixture(): RawScanReport {
  const { report } = buildLargeTopologyFixture();
  return {
    ...report,
    topology: {
      ...(report.topology as Record<string, unknown>),
      edges: [
        ...((report.topology as { edges: unknown[] }).edges),
        {
          id: 'edge-invalid-missing-target',
          source: 'dev-gateway-1',
          target: 'dev-does-not-exist',
          type: 'subnet-inferred',
          relation: 'bad_fixture',
          medium: 'l2',
          confidence: 0.4,
          explanation: 'invalid endpoint fixture',
        },
      ],
    },
  };
}
