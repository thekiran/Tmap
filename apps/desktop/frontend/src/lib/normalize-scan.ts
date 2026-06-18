import type { RawDevice, RawEdge, RawEvidence, RawProbe, RawScanReport } from './scan-schema';
import type {
  AccessCandidate,
  AccessClassification,
  Advisory,
  DeviceService,
  EvidenceRecord,
  GatewayHop,
  NetworkDevice,
  NormalizedScanReport,
  OpenService,
  ProbeInventoryItem,
  RiskFinding,
  TopologyEdge,
  TopologyNode,
  TopologyViewModel,
} from './models';
import { deviceDisplayTitle, deviceSecondaryHostname } from './topology-display';

const clamp01 = (value: unknown, fallback = 0) => {
  const n = typeof value === 'number' && Number.isFinite(value) ? value : fallback;
  return Math.max(0, Math.min(1, n));
};

const str = (value: unknown): string | null =>
  typeof value === 'string' && value.trim().length > 0 ? value : null;

const bool = (value: unknown, fallback = false): boolean =>
  typeof value === 'boolean' ? value : fallback;

const num = (value: unknown): number | null =>
  typeof value === 'number' && Number.isFinite(value) ? value : null;

const rec = (value: unknown): Record<string, unknown> =>
  value && typeof value === 'object' && !Array.isArray(value) ? (value as Record<string, unknown>) : {};

const arr = <T = unknown>(value: unknown): T[] => (Array.isArray(value) ? (value as T[]) : []);

const uniq = (values: (string | null | undefined)[]) =>
  Array.from(new Set(values.filter((value): value is string => Boolean(value))));

function addressIps(device: RawDevice, intel: Record<string, unknown>): string[] {
  const addresses = arr<Record<string, unknown>>(device.addresses).map((address) => str(address.ip));
  const intelIps = arr<string>(intel.ip_addresses);
  return uniq([device.ip, ...(device.ips ?? []), ...(device.ip_addresses ?? []), ...addresses, ...intelIps]);
}

function macFor(device: RawDevice, intel: Record<string, unknown>): string | null {
  const interfaceMac = arr<Record<string, unknown>>(device.interfaces).map((item) => str(item.mac)).find(Boolean);
  const intelInterfaces = arr<Record<string, unknown>>(intel.interfaces).map((item) => str(item.mac)).find(Boolean);
  return str(device.mac) ?? interfaceMac ?? intelInterfaces ?? null;
}

function vendorFor(device: RawDevice, intel: Record<string, unknown>): string | null {
  if (typeof device.vendor === 'string') return device.vendor;
  if (device.oui_vendor) return device.oui_vendor;
  const vendor = rec(device.vendor);
  const intelVendor = rec(intel.vendor);
  return (
    str(vendor.oui_vendor) ??
    str(vendor.fingerprint_vendor) ??
    str(intelVendor.oui_vendor) ??
    str(intelVendor.fingerprint_vendor)
  );
}

function hostnameFor(device: RawDevice, intel: Record<string, unknown>): string | null {
  return str(device.hostname) ?? str(device.name) ?? arr<string>(device.hostnames)[0] ?? arr<string>(intel.hostnames)[0] ?? null;
}

function mapNodeType(value: string | null, roles: string[]): NetworkDevice['type'] {
  const lower = (value ?? '').toLowerCase();
  const roleSet = new Set(roles.map((role) => role.toLowerCase()));
  if (roleSet.has('gateway') || lower.includes('gateway') || lower.includes('router')) return 'router';
  if (roleSet.has('agent')) return 'local_host';
  if (lower.includes('windows') || lower.includes('pc') || lower.includes('workstation')) return 'workstation';
  if (lower.includes('server')) return 'server';
  if (lower.includes('printer')) return 'printer';
  if (lower.includes('mobile') || lower.includes('phone')) return 'mobile';
  if (lower.includes('switch')) return 'managed_switch';
  if (lower.includes('access_point') || lower.includes('wifi')) return 'access_point';
  if (lower.includes('unknown')) return 'unknown';
  return 'host';
}

function reachabilityFor(device: RawDevice, roles: string[]): NetworkDevice['reachability'] {
  // Honor the agent's own reachability verdict (incl. self / arp_only) verbatim.
  const r = (device.reachability ?? '').toLowerCase();
  if (r === 'self' || r === 'reachable' || r === 'arp_only' || r === 'partial' || r === 'unreachable') {
    return r as NetworkDevice['reachability'];
  }
  if (roles.includes('agent') || device.is_agent) return 'self';
  if ((device.services ?? []).some((service) => (service.state ?? '').toLowerCase() === 'open')) return 'reachable';
  return 'unknown';
}

function normalizeServices(deviceId: string, deviceLabel: string, services: unknown): DeviceService[] {
  return arr<Record<string, unknown>>(services).map((service) => ({
    name: str(service.name) ?? 'unknown',
    port: num(service.port) ?? undefined,
    proto: str(service.protocol) ?? str(service.proto) ?? 'tcp',
    protocol: str(service.protocol) ?? str(service.proto) ?? 'tcp',
    state: str(service.state) ?? 'unknown',
    confidence: num(service.confidence) ?? undefined,
    evidenceIds: arr<string>(service.evidence_ids),
  })).filter((service, index, all) => {
    const key = `${deviceId}:${service.protocol}:${service.port ?? index}:${service.name}:${deviceLabel}`;
    return all.findIndex((item, i) => `${deviceId}:${item.protocol}:${item.port ?? i}:${item.name}:${deviceLabel}` === key) === index;
  });
}

function normalizeRiskFindings(deviceId: string, deviceLabel: string, posture: Record<string, unknown>): RiskFinding[] {
  return arr<Record<string, unknown>>(posture.findings).map((finding, index) => ({
    id: str(finding.id) ?? `${deviceId}-risk-${index}`,
    deviceId,
    deviceLabel,
    severity: str(finding.severity) ?? 'info',
    title: str(finding.title) ?? 'Finding',
    description: str(finding.description) ?? '',
    recommendation: str(finding.safe_recommendation) ?? str(finding.recommendation),
    evidenceIds: Array.isArray(finding.evidence_ids)
      ? arr<string>(finding.evidence_ids)
      : uniq(String(finding.evidence_ids ?? '').split(/\s+/)),
  }));
}

function normalizeDevices(raw: RawScanReport): NetworkDevice[] {
  const rootDevices = raw.devices ?? [];
  const intelDevices = arr<Record<string, unknown>>(rec(raw.device_intel).devices);
  const byId = new Map<string, { root?: RawDevice; intel?: Record<string, unknown> }>();

  for (const device of rootDevices) {
    const id = device.id ?? device.ip ?? `device-${byId.size + 1}`;
    byId.set(id, { ...(byId.get(id) ?? {}), root: device });
  }
  for (const intel of intelDevices) {
    const id = str(intel.id) ?? arr<string>(intel.ip_addresses)[0] ?? `intel-${byId.size + 1}`;
    byId.set(id, { ...(byId.get(id) ?? {}), intel });
  }

  return Array.from(byId.entries()).map(([id, parts]) => {
    const root = parts.root ?? ({ id } as RawDevice);
    const intel = parts.intel ?? {};
    const ips = addressIps(root, intel);
    const primaryIp = ips[0] ?? id;
    const roles = uniq([...(root.roles ?? []), ...arr<string>(intel.roles), root.role]);
    const deviceType = rec(intel.device_type);
    const primaryType = str(deviceType.primary) ?? str(root.type);
    const type = mapNodeType(primaryType, roles);
    const isGateway = bool(root.is_gateway) || bool(rec(intel.topology).is_gateway) || roles.some((role) => /gateway|router/.test(role));
    const isAgent = bool(root.is_agent) || bool(rec(intel.topology).is_agent) || roles.includes('agent');
    const hostname = hostnameFor(root, intel);
    const mac = macFor(root, intel);
    const vendor = vendorFor(root, intel);
    const label = hostname ?? primaryIp;
    const services = normalizeServices(id, label, intel.services ?? root.services);
    const riskPosture = rec(intel.security_posture);
    const riskFindings = normalizeRiskFindings(id, label, riskPosture);
    const confidence = clamp01(root.confidence ?? intel.confidence ?? deviceType.confidence, 0);
    const explicitlyUnknown =
      type === 'unknown' ||
      roles.some((role) => role.toLowerCase().includes('unknown')) ||
      String(primaryType ?? '').toLowerCase().includes('unknown');
    const weakIdentity = !isGateway && !isAgent && !hostname && !vendor && services.length === 0;
    const isUnknown = explicitlyUnknown || weakIdentity;

    return {
      id,
      ips,
      ip: primaryIp,
      mac,
      vendor,
      hostname,
      type,
      role: roles[0] ?? primaryType ?? null,
      roles,
      isGateway,
      isAgent,
      isUnknown,
      reachability: reachabilityFor(root, roles),
      discoverySources: uniq([...arr<string>(root.discovery_sources), ...arr<string>(intel.discovery_sources)]),
      confidence,
      source: str(root.source) ?? 'iad-agent',
      services,
      explanation: str(rec(deviceType.candidates)?.supporting_facts),
      limitations: arr<string>(deviceType.missing_evidence).join(' ') || null,
      rawProbeRefs: uniq([...(root.evidence_ids ?? []), ...arr<string>(intel.evidence_ids)]),
      riskLevel: str(riskPosture.risk_level),
      riskFindings,
      raw: { ...root, ...intel },
    };
  });
}

function evidenceStatus(item: RawEvidence): EvidenceRecord['status'] {
  const source = `${item.kind ?? ''} ${item.source ?? ''}`.toLowerCase();
  if (source.includes('failed')) return 'failed';
  return 'success';
}

function normalizeEvidence(raw: RawScanReport): EvidenceRecord[] {
  const byId = new Map<string, RawEvidence>();
  for (const item of [...(raw.evidence ?? []), ...(raw.evidence_registry ?? [])]) {
    const id = item.id ?? item.evidence_id;
    if (id && !byId.has(id)) byId.set(id, item);
  }

  return Array.from(byId.entries()).map(([id, item]) => ({
    id,
    probeName: str(item.source) ?? 'iad-agent',
    status: evidenceStatus(item),
    confidence: clamp01(item.confidence, 0),
    timestamp: str(item.timestamp) ?? '0001-01-01T00:00:00Z',
    evidenceClass: 'l3',
    reason: str(item.summary),
    limitations: null,
    data: item.data ?? null,
    warnings: [],
    errors: [],
    emptyEvidenceWarning: false,
    source: str(item.source) ?? 'iad-agent',
    kind: str(item.kind) ?? 'evidence',
    summary: str(item.summary) ?? str(item.kind) ?? id,
    safeToDisplay: item.safe_to_display !== false,
  }));
}

function normalizeProbes(raw: RawScanReport): ProbeInventoryItem[] {
  return (raw.probe_inventory ?? []).map((probe: RawProbe) => ({
    name: str(probe.name) ?? str(probe.probe_name) ?? 'probe',
    category: str(probe.category) ?? 'probe',
    status: str(probe.status) ?? 'unknown',
    durationMs: num(probe.duration_ms),
    producedEvidenceCount: num(probe.produced_evidence_count),
    safetyMode: str(probe.safety_mode),
    outputPath: str(probe.output_path),
    reason: str(probe.reason) ?? str(probe.skipped_reason),
    timeout: bool(probe.timeout),
  }));
}

function warningLevel(value: unknown): Advisory['level'] {
  const text = String(value ?? '').toLowerCase();
  if (text.includes('danger') || text.includes('error') || text.includes('high')) return 'danger';
  if (text.includes('info')) return 'info';
  return 'warn';
}

function normalizeWarnings(raw: RawScanReport): Advisory[] {
  const fromRoot = (raw.warnings ?? []).map((warning) => {
    if (typeof warning === 'string') return { level: 'warn' as const, text: warning };
    const entry = rec(warning);
    return {
      level: warningLevel(entry.severity),
      text: str(entry.message) ?? str(entry.text) ?? str(entry.code) ?? 'Warning',
      code: str(entry.code) ?? undefined,
    };
  });
  const fromUi = arr<string>(rec(raw.ui).warnings).map((text) => ({ level: 'warn' as const, text }));
  return [...fromRoot, ...fromUi].filter((warning, index, all) => all.findIndex((item) => item.text === warning.text) === index);
}

function selectedInterface(raw: RawScanReport): NormalizedScanReport['selectedInterface'] {
  const agent = rec(raw.agent);
  const scope = rec(raw.scope);
  const selected = arr<Record<string, unknown>>(agent.interfaces).find((iface) => iface.selected === true);
  const addresses = arr<Record<string, unknown>>(selected?.addresses);
  const ipv4 = addresses.map((address) => str(address.ip)).find((ip) => Boolean(ip && /^\d+\./.test(ip)));
  const cidr = str(scope.cidr) ?? str(selected?.cidr);
  const prefix = cidr?.includes('/') ? Number(cidr.split('/')[1]) : null;

  if (!selected && !scope.interface) return null;
  return {
    name: str(scope.interface) ?? str(selected?.name) ?? 'Interface',
    type: 'ethernet',
    ipv4: ipv4 ?? null,
    prefix: Number.isFinite(prefix) ? prefix : null,
    mac: str(selected?.mac),
    mtu: num(selected?.mtu),
    gateway: str(agent.gateway),
    dns: arr<string>(agent.dns_servers),
    linkSpeedMbps: null,
    dhcp: null,
  };
}

function gatewayChain(raw: RawScanReport): GatewayHop[] {
  const ac = rec(raw.access_classification);
  const context = rec(ac.detected_network_context);
  const chainState = rec(context.gateway_chain_state);
  const hops = arr<Record<string, unknown>>(chainState.private_hops);
  if (hops.length) {
    return hops.map((hop, index) => ({
      hop: num(hop.order) ?? index + 1,
      ip: str(hop.ip) ?? 'unknown',
      kind: str(hop.role) ?? 'gateway',
      rttMs: null,
      label: str(hop.role) ?? 'Gateway',
      private: true,
      note: str(hop.source),
    }));
  }
  return arr<string>(context.gateway_chain).map((ip, index) => ({
    hop: index + 1,
    ip,
    kind: index === 0 ? 'default_gateway' : 'route_hop',
    rttMs: null,
    label: index === 0 ? 'Default gateway' : 'Hop',
    private: true,
    note: null,
  }));
}

function normalizeCandidates(value: unknown): AccessCandidate[] {
  if (Array.isArray(value)) {
    return value.map((item) => {
      const entry = rec(item);
      return {
        type: str(entry.type) ?? str(entry.name) ?? 'unknown',
        score: clamp01(entry.score ?? entry.confidence, 0),
        note: str(entry.note) ?? str(entry.reason),
      };
    });
  }
  const scores = rec(value);
  return Object.entries(scores).map(([type, score]) => ({ type, score: clamp01(score, 0), note: null }));
}

function normalizeAccess(raw: RawScanReport): AccessClassification {
  const ac = rec(raw.access_classification);
  const cls = rec(ac.classification);
  const primaryType = str(cls.primary_type) ?? str(ac.primary_type) ?? str(raw.primary_type);
  const category = str(cls.category) ?? str(ac.category) ?? str(raw.category);
  const confidence = clamp01(cls.confidence ?? ac.confidence ?? raw.confidence, 0);
  const decision = (str(cls.decision_quality) ?? str(ac.decision_quality) ?? raw.decision_quality ?? 'low').toLowerCase();
  const candidates = normalizeCandidates(ac.candidates ?? ac.scores ?? raw['scores']);

  return {
    primaryType: primaryType && !/^unknown$/i.test(primaryType) ? primaryType : null,
    category,
    subtype: str(cls.subtype),
    confidence,
    contextConfidence: clamp01(ac.context_confidence, 0),
    decisionQuality: decision === 'high' || decision === 'medium' ? decision : 'low',
    state: str(cls.state) ?? str(ac.state),
    safeToDisplayAsFinal: bool(cls.safe_to_display_as_final, true),
    uncertaintyReasons: arr<string>(ac.uncertainty_reasons),
    candidates,
  };
}

function edgeTier(layer: string | null): TopologyEdge['tier'] {
  const value = (layer ?? '').toLowerCase();
  if (value.includes('nat')) return 'nat';
  if (value.includes('isp')) return 'isp';
  if (value.includes('l2')) return 'l2';
  return 'l3';
}

function edgeType(value: string | null): TopologyEdge['type'] {
  const normalized = (value ?? '').toLowerCase();
  const aliases: Record<string, TopologyEdge['type']> = {
    local_interface: 'local_interface',
    same_subnet: 'same_subnet',
    same_subnet_inferred: 'same_subnet',
    inferred_l2: 'same_subnet',
    inferred_l2_link: 'same_subnet',
    arp_neighbor: 'arp_confirmed',
    arp_confirmed: 'arp_confirmed',
    mac_table_confirmed: 'mac_table_confirmed',
    direct_lldp: 'lldp_confirmed',
    lldp_confirmed: 'lldp_confirmed',
    lldp_physical_neighbor: 'lldp_confirmed',
    direct_cdp: 'cdp_confirmed',
    cdp_confirmed: 'cdp_confirmed',
    cdp_physical_neighbor: 'cdp_confirmed',
    snmp_bridge: 'snmp_bridge_confirmed',
    snmp_bridge_confirmed: 'snmp_bridge_confirmed',
    snmp_bridge_fdb: 'snmp_bridge_confirmed',
    wifi_link: 'wifi_association_inferred',
    wifi_association_unknown: 'wifi_association_inferred',
    wifi_association_inferred: 'wifi_association_inferred',
    ap_bridge_inferred: 'ap_bridge_inferred',
    gateway_default: 'gateway_default',
    default_gateway: 'gateway_default',
    default_gateway_route: 'gateway_default',
    upstream_private_gateway: 'upstream_private_gateway',
    upstream_nat: 'upstream_private_gateway',
    possible_cpe_path: 'upstream_private_gateway',
    routed_hop: 'route_hop',
    route_hop: 'route_hop',
    traceroute_hop: 'route_hop',
    gateway_chain: 'route_hop',
    isp_route_hop: 'route_hop',
    nat_boundary: 'nat_boundary',
    isp_boundary: 'isp_boundary',
    unknown_link: 'unknown_l2_connection',
    unknown_l2_connection: 'unknown_l2_connection',
  };
  return aliases[normalized] ?? 'unknown_l2_connection';
}

function endpointLookup(devices: NetworkDevice[]): Map<string, string> {
  const out = new Map<string, string>();
  for (const device of devices) {
    for (const value of [
      device.id,
      device.ip,
      ...device.ips,
      device.mac,
      device.hostname,
      ...(device.roles.includes('agent') ? ['agent'] : []),
      ...(device.isGateway ? ['gateway', 'default_gateway'] : []),
    ]) {
      if (value) out.set(value, device.id);
    }
  }
  return out;
}

function normalizeEndpoint(value: unknown, lookup: Map<string, string>): string {
  const raw = str(value) ?? '';
  return lookup.get(raw) ?? raw;
}

function edgeSource(edge: Record<string, unknown>, lookup: Map<string, string>) {
  return normalizeEndpoint(edge.source ?? edge.from ?? edge.from_id ?? edge.source_id, lookup);
}

function edgeTarget(edge: Record<string, unknown>, lookup: Map<string, string>) {
  return normalizeEndpoint(edge.target ?? edge.to ?? edge.to_id ?? edge.target_id, lookup);
}

function edgeLayer(edge: Record<string, unknown>): string | null {
  const explicit = str(edge.layer) ?? str(edge.tier);
  if (explicit) return explicit;
  const kind = `${edge.type ?? ''} ${edge.relationship ?? ''}`.toLowerCase();
  if (kind.includes('l2') || kind.includes('subnet') || kind.includes('arp') || kind.includes('lldp') || kind.includes('cdp') || kind.includes('snmp')) return 'L2';
  if (kind.includes('nat')) return 'NAT';
  if (kind.includes('isp')) return 'ISP';
  return 'L3';
}

function normalizeEdges(rawEdges: Record<string, unknown>[], lookup: Map<string, string>): TopologyEdge[] {
  return rawEdges.map((edge, index) => {
    const physical = edge.physical === true;
    const rawLineStyle = (str(edge.ui_line_style) ?? str(edge.line_style) ?? '').toLowerCase();
    const lineStyle: TopologyEdge['lineStyle'] = physical ? 'solid' : rawLineStyle === 'dashed' ? 'dashed' : 'dotted';
    const tier = edgeTier(edgeLayer(edge));
    const type = edgeType(str(edge.type) ?? str(edge.relationship));
    const certainty: TopologyEdge['certainty'] = physical ? 'confirmed' : lineStyle === 'dotted' ? 'unknown' : 'inferred';
    const boundary: TopologyEdge['boundary'] = tier === 'nat' ? 'NAT' : tier === 'isp' ? 'ISP' : null;
    return {
      id: str(edge.id) ?? `edge-${index}`,
      source: edgeSource(edge, lookup),
      target: edgeTarget(edge, lookup),
      type,
      certainty,
      tier,
      confidence: clamp01(edge.confidence, physical ? 1 : 0.3),
      label: str(edge.relationship) ?? str(edge.type) ?? 'link',
      basis: str(edge.proof_source) ?? str(edge.reason) ?? 'iad-agent',
      boundary,
      thin: type === 'route_hop',
      physical,
      inferred: edge.inferred !== false && !physical,
      lineStyle,
      layers: [tier],
      // Extra detail surfaced in the edge inspector (TopologyEdge has an index
      // signature, so these ride along without widening the declared shape).
      relationship: str(edge.relationship),
      proofSource: str(edge.proof_source),
      reason: str(edge.reason),
      confidenceLabel: str(edge.confidence_label),
      evidenceIds: arr<string>(edge.evidence_ids),
    };
  }).filter((edge) => edge.source && edge.target && edge.source !== edge.target);
}

function normalizeTopology(raw: RawScanReport, devices: NetworkDevice[]): TopologyViewModel {
  const topology = rec(raw.topology);
  const topologyGraph = rec(topology.graph);
  const uiGraph = rec(rec(raw.ui).graph);
  const deviceIntel = rec(raw.device_intel);
  const lookup = endpointLookup(devices);
  const sourceNodes = [
    ...arr<Record<string, unknown>>(topology.nodes),
    ...arr<Record<string, unknown>>(topologyGraph.nodes),
    ...arr<Record<string, unknown>>(uiGraph.nodes),
  ];
  const sourceEdges = [
    ...((raw.edges ?? []) as RawEdge[]),
    ...arr<Record<string, unknown>>(topology.edges),
    ...arr<Record<string, unknown>>(topologyGraph.edges),
    ...arr<Record<string, unknown>>(uiGraph.edges),
    ...arr<Record<string, unknown>>(deviceIntel.edges),
  ].filter((edge, index, all) => {
    const r = edge as Record<string, unknown>;
    const key = str(r.id) ?? `${edgeSource(r, lookup)}>${edgeTarget(r, lookup)}>${str(r.type) ?? str(r.relationship) ?? index}`;
    return all.findIndex((item, itemIndex) => {
      const other = item as Record<string, unknown>;
      const otherKey = str(other.id) ?? `${edgeSource(other, lookup)}>${edgeTarget(other, lookup)}>${str(other.type) ?? str(other.relationship) ?? itemIndex}`;
      return otherKey === key;
    }) === index;
  }) as Record<string, unknown>[];
  const edges = normalizeEdges(sourceEdges, lookup);

  const nodeIdFor = (node: Record<string, unknown>) =>
    normalizeEndpoint(node.id ?? node.device_id ?? node.deviceId ?? node.ip ?? rec(node.metadata).device_id, lookup);
  const nodeIds = new Set<string>([
    ...devices.map((device) => device.id),
    ...sourceNodes.map(nodeIdFor),
    ...edges.flatMap((edge) => [edge.source, edge.target]),
  ]);

  const nodes: TopologyNode[] = Array.from(nodeIds).filter(Boolean).map((id, index) => {
    const device = devices.find((item) => item.id === id);
    const uiNode = sourceNodes.find((node) => nodeIdFor(node) === id) ?? {};
    const label = device ? deviceDisplayTitle(device) : str(uiNode.label) ?? id;
    const type = device?.type ?? mapNodeType(str(uiNode.type) ?? str(uiNode.device_type), arr<string>(uiNode.roles));
    const inferred = bool(uiNode.inferred, false);
    const secondaryHostname = device ? deviceSecondaryHostname(device) : null;
    return {
      id,
      type,
      label,
      sublabel: device?.ip ?? (str(uiNode.label) === label ? null : str(uiNode.label)),
      certainty: inferred ? 'inferred' : 'confirmed',
      layers: ['l3'],
      badge: device?.isGateway ? 'Gateway' : device?.isAgent ? 'Local' : device?.isUnknown ? 'Unknown' : null,
      deviceId: device?.id ?? null,
      accent: Boolean(device?.isGateway || device?.isAgent),
      position: { x: 80 + (index % 4) * 230, y: 70 + Math.floor(index / 4) * 150 },
      confidence: device?.confidence ?? clamp01(uiNode.confidence, 0),
      roles: device?.roles ?? [],
      isGateway: device?.isGateway ?? false,
      isAgent: device?.isAgent ?? false,
      isUnknown: device?.isUnknown ?? false,
      ip: device?.ip ?? null,
      hostname: secondaryHostname,
      reachability: device?.reachability ?? 'unknown',
      discoverySources: device?.discoverySources ?? [],
    };
  });

  return { generated: sourceEdges.length === 0, nodes, edges };
}

function normalizeDiscoverySummary(raw: RawScanReport): NormalizedScanReport['discoverySummary'] {
  const d = rec(raw.discovery_summary);
  if (Object.keys(d).length === 0) return null;
  return {
    cidr: str(d.cidr) ?? '',
    addressesScanned: num(d.addresses_scanned) ?? 0,
    devicesFound: num(d.devices_found) ?? 0,
    arpFound: num(d.arp_found) ?? 0,
    icmpFound: num(d.icmp_found) ?? 0,
    tcpFound: num(d.tcp_found) ?? 0,
    mdnsFound: num(d.mdns_found) ?? 0,
    ssdpFound: num(d.ssdp_found) ?? 0,
    llmnrFound: num(d.llmnr_found) ?? 0,
    netbiosFound: num(d.netbios_found) ?? 0,
    nmapFound: num(d.nmap_found) ?? 0,
    scanDurationMs: num(d.scan_duration_ms) ?? 0,
  };
}

export function normalizeScan(raw: RawScanReport): NormalizedScanReport {
  const devices = normalizeDevices(raw);
  const evidence = normalizeEvidence(raw);
  const probes = normalizeProbes(raw);
  const warnings = normalizeWarnings(raw);
  const access = normalizeAccess(raw);
  const topology = normalizeTopology(raw, devices);
  const services: OpenService[] = devices.flatMap((device) =>
    device.services
      .filter((service) => (service.state ?? '').toLowerCase() === 'open')
      .map((service, index) => ({
        id: `${device.id}-${service.protocol ?? 'tcp'}-${service.port ?? index}-${service.name}`,
        deviceId: device.id,
        deviceLabel: device.hostname ?? device.ip,
        port: service.port ?? null,
        protocol: service.protocol ?? service.proto ?? 'tcp',
        name: service.name,
        state: service.state ?? 'open',
        confidence: service.confidence ?? null,
        evidenceIds: service.evidenceIds,
      })),
  );
  const riskFindings = devices.flatMap((device) => device.riskFindings);
  const summary = rec(raw.summary);
  const intelSummary = rec(rec(raw.device_intel).summary);
  const selected = selectedInterface(raw);
  const gateways = gatewayChain(raw);
  const gatewayDevice = devices.find((device) => device.isGateway) ?? devices.find((device) => gateways.some((hop) => device.ips.includes(hop.ip))) ?? null;

  return {
    scanId: raw.scan_id ?? 'unknown',
    schemaVersion: raw.schema_version ?? null,
    createdAt: raw.created_at ?? new Date().toISOString(),
    status: 'complete',
    mode: str(rec(raw.scope).profile) ?? str(rec(raw.access_classification).mode) ?? 'standard',
    durationMs: num(summary.duration_ms),
    safeMode: true,
    sourceProfile: str(rec(raw.scope).profile),
    raw,

    primaryType: access.primaryType,
    isUnknown: !access.primaryType,
    category: access.category,
    confidence: access.confidence,
    classificationConfidence: access.confidence,
    contextConfidence: access.contextConfidence,
    decisionQuality: access.decisionQuality,
    uncertaintyReasons: access.uncertaintyReasons,
    candidates: access.candidates,
    access,

    selectedInterface: selected,
    gatewayChain: gateways,
    gatewayDevice,
    nat: null,
    ipv6: null,
    publicIp: null,
    performance: null,

    confidenceBreakdown: [],
    nextBestProbes: arr<Record<string, unknown>>(rec(raw.access_classification).next_best_probes).map((probe) => ({
      name: str(probe.probe_name) ?? 'probe',
      gain: 0,
      requires: null,
      tier: 'physical',
      detail: str(probe.reason),
      reason: str(probe.reason) ?? undefined,
      expectedEvidence: arr<string>(probe.expected_evidence),
      safety: str(probe.safety) ?? undefined,
    })),
    warnings,

    devices,
    unknownDevices: devices.filter((device) => device.isUnknown),
    openServices: services,
    riskFindings,
    evidence,
    probes,

    topologyGenerated: topology.generated,
    rawTopologyNodes: topology.nodes,
    rawTopologyEdges: topology.edges,
    topology,

    summary: {
      deviceCount: num(summary.device_count) ?? devices.length,
      edgeCount: num(summary.edge_count) ?? topology.edges.length,
      evidenceCount: num(summary.evidence_count) ?? evidence.length,
      probeCount: probes.length,
      warningCount: warnings.length,
      serviceCount: num(intelSummary.service_count) ?? services.length,
      riskFindingCount: num(intelSummary.security_finding_count) ?? riskFindings.length,
      inferredOnly: bool(summary.inferred_only, topology.edges.every((edge) => !edge.physical)),
      physicalEdgeCount: num(intelSummary.physical_edges) ?? topology.edges.filter((edge) => edge.physical).length,
    },

    discoverySummary: normalizeDiscoverySummary(raw),
  };
}
