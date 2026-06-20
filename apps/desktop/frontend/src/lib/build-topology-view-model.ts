/**
 * build-topology-view-model.ts â€” NormalizedScanReport â†’ TopologyViewModel.
 *
 * SAFETY RULES (enforced here, not optional):
 *  - If the agent supplied topology.nodes/edges, we honor their declared
 *    certainty verbatim.
 *  - If it did NOT, we synthesize a CONSERVATIVE view and mark every synthesized
 *    node/edge as inferred or unknown â€” never confirmed.
 *  - Traceroute / gateway-chain hops are L3 routers (isp_route_hop / route_hop),
 *    NEVER physical switches.
 *  - Unmanaged switches only ever appear as `unmanaged_switch_inferred` with
 *    `unknown` certainty.
 *  - We never depict ISP-internal topology beyond the first public hop.
 *  - Only the local interface link is "confirmed" without external probe data,
 *    because the host's own default route is directly observable.
 */
import type { NormalizedScanReport, NetworkDevice } from './models';
import type { TopologyViewModel, TopologyNode, TopologyEdge } from './models';
import type { EdgeCertainty, EdgeType, NodeType } from './scan-schema';
import { deviceDisplayTitle, deviceSecondaryHostname } from './topology-display';

const COL = { lan: 150, spine: 430, branch: 700 };
const ROW = 118;

function deviceBadge(d: NetworkDevice): string | null {
  if (d.isGateway) return 'Gateway';
  if (d.isAgent) return 'Local';
  if (d.roles.some((role) => role.toLowerCase() === 'upstream_private_gateway' || role.toLowerCase() === 'possible_cpe')) {
    return 'Upstream';
  }
  if (d.isUnknown) return 'Unknown';
  return null;
}

function deviceNode(d: NetworkDevice, index: number): TopologyNode {
  return {
    id: d.id,
    type: d.type,
    label: deviceDisplayTitle(d),
    sublabel: d.ip,
    certainty: 'confirmed',
    layers: ['l2', 'l3'],
    badge: deviceBadge(d),
    deviceId: d.id,
    accent: d.isGateway || d.isAgent,
    position: {
      x: index % 2 === 0 ? COL.lan : COL.branch,
      y: 168 + Math.floor(index / 2) * 84,
    },
    confidence: d.confidence,
    roles: d.roles,
    isGateway: d.isGateway,
    isAgent: d.isAgent,
    isUnknown: d.isUnknown,
    ip: d.ip,
    hostname: deviceSecondaryHostname(d),
    reachability: d.reachability,
    discoverySources: d.discoverySources,
    mobileFingerprint: d.mobileFingerprint,
    deviceTypeHint: d.deviceTypeHint,
    osHint: d.osHint,
    osConfidence: d.osConfidence,
    osEvidenceSummary: d.osEvidenceSummary,
  } as TopologyNode;
}

function deviceEdgeType(d: NetworkDevice): { type: EdgeType; certainty: EdgeCertainty; label: string; basis: string } {
  const src = (d.source ?? '').toLowerCase();
  if (d.type === 'access_point' || d.type === 'mesh_node' || d.type === 'repeater') {
    return {
      type: 'wifi_association_inferred', certainty: 'inferred', label: 'Wi-Fi assoc (inferred)',
      basis: 'Service records suggest a wireless mesh node; the bridge topology is not confirmed.',
    };
  }
  if (d.confidence < 0.45) {
    return {
      type: 'unknown_l2_connection', certainty: 'unknown', label: 'Unknown L2',
      basis: 'Host answered on the local segment but its attachment point is not observable.',
    };
  }
  if (src.includes('arp')) {
    return { type: 'arp_confirmed', certainty: 'confirmed', label: 'ARP confirmed', basis: 'Host answered ARP on the local broadcast domain.' };
  }
  if (src.includes('mdns') || src.includes('mac')) {
    return { type: 'mac_table_confirmed', certainty: 'confirmed', label: 'Discovered', basis: 'Host advertised services / appeared in the neighbor table on the local link.' };
  }
  return { type: 'same_subnet', certainty: 'inferred', label: 'Same subnet (inferred)', basis: 'Shares the local subnet; exact attachment point unverified.' };
}

/** Map a gateway-chain hop kind to a node type. Hops are routers, never switches. */
function hopNodeType(kind: string, isLast: boolean): NodeType {
  if (kind === 'default_gateway') return 'default_gateway';
  if (kind === 'isp_gateway') return 'isp_gateway';
  if (kind === 'upstream_private_gateway') return 'router';
  return isLast ? 'isp_gateway' : 'isp_route_hop';
}

function fromAgentTopology(scan: NormalizedScanReport): TopologyViewModel | null {
  if (!scan.rawTopologyNodes.length) return null;
  const nodes: TopologyNode[] = scan.rawTopologyNodes.map((n, i) => {
    const r = n as Record<string, unknown>;
    const deviceId = r.deviceId ?? r.device_id;
    const rawPosition = (r.position && typeof r.position === 'object' ? r.position : {}) as Record<string, unknown>;
    return {
      // Preserve already-normalized extras (reachability, discoverySources, …)
      // so node badges/muting survive this re-map.
      ...(r as unknown as TopologyNode),
      id: String(r.id ?? `n-${i}`),
      type: (String(r.type ?? 'unknown') as NodeType),
      label: String(r.label ?? r.id ?? 'Node'),
      sublabel: r.sublabel ? String(r.sublabel) : null,
      certainty: (String(r.certainty ?? 'inferred') as EdgeCertainty),
      layers: (Array.isArray(r.layers) ? (r.layers as string[]) : ['l3']) as TopologyNode['layers'],
      badge: r.badge ? String(r.badge) : null,
      deviceId: deviceId ? String(deviceId) : null,
      accent: Boolean(r.accent),
      position: {
        x: Number((rawPosition.x as number) ?? (r.x as number) ?? 0),
        y: Number((rawPosition.y as number) ?? (r.y as number) ?? 0),
      },
      confidence: Number((r.confidence as number) ?? 0),
      roles: Array.isArray(r.roles) ? (r.roles as string[]) : [],
      isGateway: Boolean(r.isGateway ?? r.is_gateway),
      isAgent: Boolean(r.isAgent ?? r.is_agent),
      isUnknown: Boolean(r.isUnknown ?? r.is_unknown),
    } as TopologyNode;
  });
  const edges: TopologyEdge[] = scan.rawTopologyEdges.map((e, i) => {
    const r = e as Record<string, unknown>;
    const physical = Boolean(r.physical);
    const lineStyle = (r.lineStyle as TopologyEdge['lineStyle']) ?? (physical ? 'solid' : 'dotted');
    return {
      // Preserve all already-normalized fields (physical / inferred / lineStyle
      // are set by normalize-scan and MUST survive into the rendered edge so
      // inferred links draw dotted/dashed, not solid).
      ...(r as unknown as TopologyEdge),
      id: String(r.id ?? `e-${i}`),
      source: String(r.source ?? r.from ?? ''),
      target: String(r.target ?? r.to ?? ''),
      type: (String(r.type ?? 'unknown_l2_connection') as EdgeType),
      certainty: (String(r.certainty ?? 'inferred') as EdgeCertainty),
      tier: (String(r.tier ?? 'l3') as TopologyEdge['tier']),
      confidence: Number((r.confidence as number) ?? 0.5),
      label: String(r.label ?? r.type ?? 'Edge'),
      basis: String(r.basis ?? 'Provided by agent topology.'),
      boundary: (r.boundary as 'NAT' | 'ISP' | null) ?? null,
      thin: Boolean(r.thin),
      physical,
      inferred: r.inferred !== false && !physical,
      lineStyle,
      layers: (Array.isArray(r.layers) ? (r.layers as string[]) : ['l3']) as TopologyEdge['layers'],
    } as TopologyEdge;
  });
  return { generated: scan.topologyGenerated, nodes, edges };
}

function ensureAllDevices(topology: TopologyViewModel, scan: NormalizedScanReport): TopologyViewModel {
  const nodes = [...topology.nodes];
  const edges = [...topology.edges];
  const nodeIds = new Set(nodes.map((node) => node.id));
  const deviceIds = new Set(nodes.map((node) => node.deviceId).filter((id): id is string => Boolean(id)));

  scan.devices.forEach((device, index) => {
    if (nodeIds.has(device.id) || deviceIds.has(device.id)) return;
    const node = deviceNode(device, index);
    nodes.push(node);
    nodeIds.add(node.id);
    deviceIds.add(device.id);
  });

  const connected = new Set(edges.flatMap((edge) => [edge.source, edge.target]));
  const anchor =
    nodes.find((node) => node.isGateway)?.id ??
    nodes.find((node) => node.isAgent)?.id ??
    nodes[0]?.id ??
    null;

  if (anchor) {
    for (const node of nodes) {
      if (!node.deviceId || node.id === anchor || connected.has(node.id)) continue;
      const device = scan.devices.find((item) => item.id === node.deviceId);
      const edgeInfo = device ? deviceEdgeType(device) : {
        type: 'same_subnet' as EdgeType,
        certainty: 'inferred' as EdgeCertainty,
        label: 'Discovered',
        basis: 'Agent reported the device but no explicit topology edge was supplied.',
      };
      edges.push({
        id: `e-discovered-${anchor}-${node.id}`,
        source: anchor,
        target: node.id,
        type: edgeInfo.type,
        certainty: edgeInfo.certainty,
        tier: 'l2',
        confidence: device?.confidence ?? node.confidence ?? 0.3,
        label: edgeInfo.label,
        basis: edgeInfo.basis,
        boundary: null,
        thin: true,
        physical: false,
        inferred: edgeInfo.certainty !== 'confirmed',
        lineStyle: edgeInfo.certainty === 'unknown' ? 'dotted' : 'dashed',
        relationship: edgeInfo.type === 'gateway_default' ? 'default_gateway' : edgeInfo.type === 'same_subnet' ? 'same_subnet' : edgeInfo.label,
        layers: edgeInfo.certainty === 'unknown' ? ['unknown'] : ['l2'],
      } as TopologyEdge);
      connected.add(node.id);
      connected.add(anchor);
    }
  }

  return { ...topology, nodes, edges };
}

/** Conservative synthesis when the agent gives no topology. */
function synthesize(scan: NormalizedScanReport): TopologyViewModel {
  const nodes: TopologyNode[] = [];
  const edges: TopologyEdge[] = [];

  const host = scan.devices.find((d) => d.reachability === 'self') ?? null;
  const hostId = host ? `dev-${host.id}` : 'dev-self';
  nodes.push({
    id: hostId, type: 'local_host', label: host ? deviceDisplayTitle(host) : 'This PC',
    sublabel: host?.ips?.[0] ?? 'Unknown IP', certainty: 'confirmed', layers: ['l3'],
    badge: 'Local', deviceId: host?.id ?? null, accent: true, position: { x: COL.spine, y: 50 },
    confidence: host?.confidence ?? 1,
    roles: host?.roles ?? ['agent'],
    isGateway: false,
    isAgent: true,
    isUnknown: false,
    ip: host?.ip ?? null,
    hostname: host ? deviceSecondaryHostname(host) : null,
    mobileFingerprint: host?.mobileFingerprint ?? null,
    deviceTypeHint: host?.deviceTypeHint ?? null,
    osHint: host?.osHint ?? null,
    osConfidence: host?.osConfidence ?? null,
    osEvidenceSummary: host?.osEvidenceSummary ?? [],
  } as TopologyNode);

  // Gateway chain spine (down the middle). First hop is the LAN gateway.
  let prevId = hostId;
  const chain = scan.gatewayChain;
  chain.forEach((g, i) => {
    const isLast = i === chain.length - 1;
    const id = `gw-${i}`;
    const isNat = g.kind === 'upstream_private_gateway' || /100\.6[4-9]|cgnat/i.test(g.note ?? '');
    nodes.push({
      id, type: hopNodeType(g.kind, isLast), label: g.label,
      sublabel: g.ip, certainty: 'confirmed', layers: ['l3'],
      badge: g.kind.replace(/_/g, ' '), deviceId: null, accent: false,
      position: { x: COL.spine, y: 50 + ROW * (i + 1) },
    } as TopologyNode);
    // edge from previous spine node
    if (i === 0) {
      edges.push({
        id: `e-host-gw0`, source: hostId, target: id, type: 'local_interface',
        certainty: 'confirmed', tier: 'l3', confidence: 1.0,
        label: 'Default route', basis: 'Local routing table',
        boundary: null, thin: false, physical: false, inferred: true, lineStyle: 'dotted', layers: ['l3'],
        relationship: 'default_gateway',
      } as TopologyEdge);
    } else {
      const boundary = isNat ? 'NAT' : null;
      edges.push({
        id: `e-gw${i - 1}-gw${i}`, source: prevId, target: id,
        type: g.kind === 'upstream_private_gateway' ? 'upstream_private_gateway' : 'route_hop',
        certainty: 'inferred', tier: isNat ? 'nat' : 'isp', confidence: 0.8,
        label: `Hop ${i + 1}`, basis: 'Traceroute inference',
        boundary, thin: !isNat, physical: false, inferred: true, lineStyle: 'dotted', layers: isNat ? ['l3', 'nat'] : ['l3', 'isp'],
      } as TopologyEdge);
    }
    prevId = id;
  });

  // Public internet terminal
  if (scan.publicIp?.address) {
    const inetId = 'inet';
    nodes.push({
      id: inetId, type: 'public_internet', label: 'Public internet',
      sublabel: scan.publicIp.address, certainty: 'confirmed', layers: ['isp'],
      badge: null, deviceId: null, accent: false, position: { x: COL.spine, y: 50 + ROW * (chain.length + 1) },
    } as TopologyNode);
    if (prevId !== hostId) {
      edges.push({
        id: 'e-isp-inet', source: prevId, target: inetId, type: 'isp_boundary',
        certainty: 'confirmed', tier: 'isp', confidence: 1.0,
        label: 'Public IP', basis: 'STUN / API lookup',
        boundary: 'ISP', thin: false, layers: ['isp'],
      } as TopologyEdge);
    }
  }

  // LAN devices hang off the first gateway (the home router), fanned L/R.
  const routerId = chain.length ? 'gw-0' : hostId;
  const lan = scan.devices.filter((d) => d.reachability !== 'self');
  let li = 0;
  let lowConfCount = 0;
  for (const d of lan) {
    const id = `dev-${d.id}`;
    const col = li % 2 === 0 ? COL.lan : COL.branch;
    const row = 168 + Math.floor(li / 2) * 84;
    const e = deviceEdgeType(d);
    nodes.push({
      id, type: d.type, label: deviceDisplayTitle(d), sublabel: d.ip,
      certainty: 'confirmed', layers: ['l2', 'l3'],
      badge: deviceBadge(d), deviceId: d.id, accent: d.isAgent || d.isGateway, position: { x: col, y: row },
      isGateway: d.isGateway, isAgent: d.isAgent, isUnknown: d.isUnknown, ip: d.ip, hostname: deviceSecondaryHostname(d),
      reachability: d.reachability, discoverySources: d.discoverySources, confidence: d.confidence, roles: d.roles,
      mobileFingerprint: d.mobileFingerprint, deviceTypeHint: d.deviceTypeHint, osHint: d.osHint, osConfidence: d.osConfidence,
      osEvidenceSummary: d.osEvidenceSummary,
    } as TopologyNode);
    edges.push({
      id: `e-router-${id}`, source: routerId, target: id, type: e.type,
      certainty: e.certainty, tier: 'l2', confidence: d.confidence,
      label: e.label, basis: e.basis, boundary: null, thin: true, physical: false, inferred: true, lineStyle: 'dotted',
      relationship: e.type === 'same_subnet' ? 'same_subnet' : e.label,
      layers: e.certainty === 'unknown' ? ['unknown'] : ['l2'],
    } as TopologyEdge);
    if (e.certainty === 'unknown') lowConfCount++;
    li++;
  }

  // If several low-confidence hosts exist, posit ONE inferred unmanaged switch +
  // an unknown L2 segment â€” explicitly inferred, never confirmed.
  if (lowConfCount >= 1) {
    nodes.push({
      id: 'sw-inferred', type: 'unmanaged_switch_inferred', label: 'Switch (inferred)',
      sublabel: 'Unknown segment', certainty: 'unknown', layers: ['unknown'],
      badge: null, deviceId: null, accent: false, position: { x: COL.branch, y: 168 + Math.ceil(li / 2) * 84 },
    } as TopologyNode);
    edges.push({
      id: 'e-router-sw', source: routerId, target: 'sw-inferred', type: 'unknown_l2_connection',
      certainty: 'unknown', tier: 'l2', confidence: 0.1,
      label: 'Inferred segment', basis: 'Devices lack observable L2 attachment.',
      boundary: null, thin: false, layers: ['unknown'],
      physical: false, inferred: true, lineStyle: 'dotted', relationship: 'attachment_unknown',
    } as TopologyEdge);
  }

  return { generated: true, nodes, edges };
}

export function buildTopologyViewModel(scan: NormalizedScanReport): TopologyViewModel {
  return ensureAllDevices(fromAgentTopology(scan) ?? synthesize(scan), scan);
}
