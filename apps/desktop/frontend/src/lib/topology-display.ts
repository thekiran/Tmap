import type { NetworkDevice, TopologyEdge, TopologyNode } from './models';

type DeviceLike = Pick<
  NetworkDevice,
  'ip' | 'hostname' | 'isAgent' | 'isGateway' | 'isUnknown' | 'discoverySources' | 'reachability' | 'roles'
>;

export function discoverySourceLabel(source: string): string {
  const value = source.toLowerCase().replace(/[\s-]+/g, '_');
  if (value.includes('nmap')) return 'Nmap';
  if (value.includes('mdns')) return 'mDNS';
  if (value.includes('ssdp')) return 'SSDP';
  if (value.includes('icmp') || value.includes('ping')) return 'ICMP';
  if (value.includes('tcp') || value.includes('port')) return 'TCP';
  if (value.includes('arp') || value.includes('neighbor')) return 'ARP';
  if (value.includes('llmnr')) return 'LLMNR';
  if (value.includes('netbios')) return 'NetBIOS';
  return source.replace(/_/g, ' ');
}

export function discoverySourceBadges(sources: string[], reachability?: string): string[] {
  const labels = sources.map(discoverySourceLabel);
  if (reachability === 'arp_only') labels.push('ARP');
  return Array.from(new Set(labels.filter(Boolean)));
}

export function deviceDisplayTitle(device: DeviceLike): string {
  if (device.isAgent) return 'This PC';
  if (device.isGateway) return 'Gateway / Router';
  if (hasUpstreamGatewayRole(device.roles)) return 'Upstream gateway / CPE';
  if (device.isUnknown) return device.ip;
  return device.hostname ?? device.ip;
}

export function deviceSecondaryHostname(device: DeviceLike): string | null {
  const hostname = device.hostname?.trim();
  if (!hostname || hostname === device.ip) return null;
  return deviceDisplayTitle(device) === hostname ? null : hostname;
}

export function nodeDisplayTitle(node: TopologyNode): string {
  if (node.isAgent) return 'This PC';
  if (node.isGateway) return 'Gateway / Router';
  if (hasUpstreamGatewayRole(node.roles)) return 'Upstream gateway / CPE';
  if (node.isUnknown) return String((node as { ip?: string }).ip ?? node.sublabel ?? node.label);
  return node.label;
}

function hasUpstreamGatewayRole(roles: string[] = []): boolean {
  return roles.some((role) => {
    const lower = role.toLowerCase();
    return lower === 'upstream_private_gateway' || lower === 'possible_cpe';
  });
}

export function nodeIp(node: TopologyNode): string | null {
  const ip = (node as { ip?: string }).ip;
  return ip ?? node.sublabel ?? null;
}

export function nodeSecondaryHostname(node: TopologyNode): string | null {
  const hostname = (node as { hostname?: string | null }).hostname?.trim();
  const ip = nodeIp(node);
  if (!hostname || hostname === ip) return null;
  if (hostname === nodeDisplayTitle(node)) return null;
  return hostname;
}

function relationshipLabel(edge: TopologyEdge): string {
  const relationship = String((edge.relationship as string | undefined) ?? edge.label ?? edge.type).toLowerCase();
  if (edge.boundary) return `${edge.boundary.toLowerCase()} boundary`;
  if (relationship.includes('default_gateway') || edge.type === 'gateway_default') return 'default gateway';
  if (relationship.includes('same_subnet') || edge.type === 'same_subnet') return 'same subnet';
  if (relationship.includes('route') || edge.type === 'route_hop') return 'route hop';
  if (relationship.includes('upstream') || edge.type === 'upstream_private_gateway') return 'upstream gateway';
  if (relationship.includes('arp')) return 'ARP observed';
  if (relationship.includes('unknown')) return 'attachment unknown';
  return relationship.replace(/_/g, ' ');
}

function edgeLayerLabel(edge: TopologyEdge): string {
  if (edge.type === 'gateway_default' || edge.type === 'same_subnet') return 'L3';
  if (edge.tier === 'nat') return 'NAT';
  if (edge.tier === 'isp') return 'ISP';
  if (edge.tier === 'l2' && edge.physical) return 'L2';
  return 'L3';
}

export function formatTopologyEdgeLabel(edge: TopologyEdge): string {
  const status = edge.physical ? 'physical' : 'inferred';
  const confidence = Math.round((edge.confidence ?? 0) * 100);
  return `${edgeLayerLabel(edge)} ${status} · ${relationshipLabel(edge)} · ${confidence}%`;
}

export const UNKNOWN_DEVICE_REASON =
  'The device was detected on the LAN, but no strong service, hostname, vendor, or protocol fingerprint was available.';
