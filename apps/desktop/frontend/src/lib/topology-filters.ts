import type { TopologyEdge, TopologyNode, TopologyViewModel } from './models';

export type TopologyFilterState = {
  deviceType: string;
  minConfidence: number;
  wireless: string;
  subnet: string;
  online: string;
  evidenceSource: string;
  mobileOS: string;
};

export const defaultTopologyFilters: TopologyFilterState = {
  deviceType: 'all',
  minConfidence: 0,
  wireless: '',
  subnet: '',
  online: 'all',
  evidenceSource: 'all',
  mobileOS: 'all',
};

export function topologyFilterOptions(topology: TopologyViewModel | null | undefined) {
  const types = new Set<string>();
  const evidence = new Set<string>();
  topology?.nodes.forEach((node) => {
    if (node.type) types.add(node.type);
    if (node.role) types.add(node.role);
    node.roles.forEach((role) => types.add(role));
    (node.rawSources ?? []).forEach((source) => evidence.add(source));
    ((node as { discoverySources?: string[] }).discoverySources ?? []).forEach((source) => evidence.add(source));
  });
  topology?.edges.forEach((edge) => (edge.evidence ?? []).forEach((item) => {
    const source = String(item.source ?? '');
    if (source) evidence.add(source);
  }));
  return { types: Array.from(types).sort(), evidence: Array.from(evidence).sort() };
}

export function applyTopologyFilters(
  topology: TopologyViewModel | null | undefined,
  filters: TopologyFilterState,
): TopologyViewModel | null {
  if (!topology) return null;
  const query = filters.wireless.trim().toLowerCase();
  const subnet = filters.subnet.trim();
  const nodeAllowed = (node: TopologyNode) => {
    if (filters.deviceType !== 'all' && node.type !== filters.deviceType && node.role !== filters.deviceType && !node.roles.includes(filters.deviceType)) return false;
    if ((node.confidence ?? 0) < filters.minConfidence) return false;
    if (subnet && !(node.ip ?? '').startsWith(subnet)) return false;
    const reachability = String((node as { reachability?: string }).reachability ?? '');
    if (filters.online === 'online' && !(reachability === 'self' || reachability === 'reachable')) return false;
    if (filters.online === 'offline' && (reachability === 'self' || reachability === 'reachable')) return false;
    if (query) {
      const wireless = node.wireless ?? {};
      const haystack = JSON.stringify(wireless).toLowerCase();
      if (!haystack.includes(query)) return false;
    }
    if (filters.evidenceSource !== 'all') {
      const sources = new Set([...(node.rawSources ?? []), ...((node as { discoverySources?: string[] }).discoverySources ?? [])]);
      if (!sources.has(filters.evidenceSource)) return false;
    }
    if (!matchesMobileOSFilter(node, filters.mobileOS)) return false;
    return true;
  };
  const nodes = topology.nodes.filter(nodeAllowed);
  const ids = new Set(nodes.map((node) => node.id));
  const edges = topology.edges.filter((edge: TopologyEdge) => {
    if (!ids.has(edge.source) || !ids.has(edge.target)) return false;
    if (filters.evidenceSource === 'all') return true;
    const embeddedSources = (edge.evidence ?? []).map((item) => String(item.source ?? ''));
    return embeddedSources.includes(filters.evidenceSource) || String(edge.proofSource ?? edge.basis ?? '') === filters.evidenceSource;
  });
  return { ...topology, nodes, edges };
}

function matchesMobileOSFilter(node: TopologyNode, filter: string): boolean {
  if (filter === 'all') return true;
  const classification = String(node.mobileFingerprint?.classification ?? '').toLowerCase();
  const osHint = String(node.osHint ?? '').toLowerCase();
  switch (filter) {
    case 'ios':
      return osHint === 'ios' || classification.endsWith('_ios');
    case 'ipados':
      return osHint === 'ipados' || classification.endsWith('_ipados');
    case 'android':
      return osHint === 'android' || classification.endsWith('_android');
    case 'unknown_mobile':
      return classification === 'unknown_mobile';
    case 'unknown_device':
      return node.isUnknown || classification === 'unknown_device' || (!classification && osHint !== 'ios' && osHint !== 'ipados' && osHint !== 'android');
    case 'conflict':
      return classification === 'conflicting_mobile_os_evidence' || (node.mobileFingerprint?.conflicts?.length ?? 0) > 0;
    default:
      return true;
  }
}
