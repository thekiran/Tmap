/**
 * packet-flow.ts — derives packet-flow animation parameters for a topology edge.
 *
 * The agent does not (yet) capture live per-edge traffic, so we DERIVE a
 * conservative, evidence-led flow descriptor from the edge's existing certainty,
 * confidence and relationship. Any explicit fields a future backend sets on the
 * edge (trafficState / direction / confidence / packetRate / protocol / animated)
 * take precedence — derivation only fills the gaps. This keeps the animation
 * honest: confirmed, evidence-backed links flow; inferred/weak links stay idle.
 */
import type { TopologyEdge } from './models';

export type TrafficState = 'idle' | 'low' | 'medium' | 'high';
export type FlowDirection = 'forward' | 'reverse' | 'bidirectional' | 'unknown';
export type FlowConfidence = 'confirmed' | 'inferred' | 'weak';
export type PacketProtocol = 'tcp' | 'udp' | 'arp' | 'icmp' | 'dns' | 'other';
export type PacketIntensity = 'low' | 'normal' | 'high';

export interface PacketFlow {
  trafficState: TrafficState;
  direction: FlowDirection;
  confidence: FlowConfidence;
  protocol: PacketProtocol;
  packetRate: number | null;
  suspicious: boolean;
  animated: boolean;
}

/** Optional explicit fields a backend may attach to an edge in the future. */
interface PacketFlowFields {
  trafficState?: TrafficState;
  direction?: FlowDirection;
  flowConfidence?: FlowConfidence;
  packetRate?: number;
  protocol?: string;
  animated?: boolean;
  suspicious?: boolean;
}

function mapConfidence(edge: TopologyEdge, explicit?: FlowConfidence): FlowConfidence {
  if (explicit) return explicit;
  switch (edge.certainty) {
    case 'confirmed':
      return 'confirmed';
    case 'inferred':
      return 'inferred';
    default:
      return 'weak';
  }
}

function detectProtocol(edge: TopologyEdge, explicit?: string): PacketProtocol {
  const hint = (explicit ?? '').toLowerCase();
  const haystack = `${hint} ${edge.type} ${edge.label ?? ''} ${String(edge.relationship ?? '')} ${String(edge.medium ?? '')}`.toLowerCase();
  // Match at a word start (\b before the token). A trailing \b is avoided on
  // purpose: identifiers like "arp_confirmed" / "tcp_sweep" use "_", which is a
  // word char, so \barp\b would never match them.
  if (/\barp|neighbor/.test(haystack)) return 'arp';
  if (/\bdns|mdns|llmnr/.test(haystack)) return 'dns';
  if (/\bicmp|ping/.test(haystack)) return 'icmp';
  if (/\budp|ssdp/.test(haystack)) return 'udp';
  if (/\btcp|http|tls/.test(haystack)) return 'tcp';
  return 'other';
}

function deriveDirection(edge: TopologyEdge, explicit?: FlowDirection): FlowDirection {
  if (explicit) return explicit;
  const t = edge.type;
  // Hierarchical / routed links flow up toward the gateway/ISP (source→target).
  if (t === 'gateway_default' || t === 'local_interface' || t === 'route_hop' || t === 'isp_boundary' || t === 'upstream_private_gateway') {
    return 'forward';
  }
  // Peer LAN links are genuinely two-way.
  if (t === 'arp_confirmed' || t === 'mac_table_confirmed' || t === 'same_subnet' || t === 'wifi_association_inferred') {
    return 'bidirectional';
  }
  // Anything we can't attribute gets a neutral, subtle two-way shimmer.
  return 'unknown';
}

function deriveTrafficState(edge: TopologyEdge, confidence: FlowConfidence, explicit?: TrafficState, rate?: number): TrafficState {
  if (explicit) return explicit;
  if (typeof rate === 'number') {
    if (rate <= 0) return 'idle';
    if (rate < 20) return 'low';
    if (rate < 200) return 'medium';
    return 'high';
  }
  // Evidence rule: only confirmed links are treated as carrying observed
  // traffic. Inferred/weak links stay idle so the map differentiates clearly.
  if (confidence !== 'confirmed') return 'idle';
  const c = edge.confidence ?? 0;
  if (c >= 0.85) return 'medium';
  if (c >= 0.5) return 'low';
  return 'low';
}

export function derivePacketFlow(edge: TopologyEdge): PacketFlow {
  const f = edge as unknown as PacketFlowFields;
  const confidence = mapConfidence(edge, f.flowConfidence);
  const protocol = detectProtocol(edge, f.protocol);
  const direction = deriveDirection(edge, f.direction);
  const trafficState = deriveTrafficState(edge, confidence, f.trafficState, f.packetRate);
  const suspicious = f.suspicious ?? Boolean(edge.boundary === 'ISP' && Array.isArray(edge.warnings) && edge.warnings.length > 0);
  const animated = f.animated ?? trafficState !== 'idle';
  return {
    trafficState,
    direction,
    confidence,
    protocol,
    packetRate: typeof f.packetRate === 'number' ? f.packetRate : null,
    suspicious,
    animated,
  };
}

const PROTOCOL_COLORS: Record<PacketProtocol, { color: string; glow: string }> = {
  tcp: { color: '#38bdf8', glow: '#0ea5e9' }, // cyan
  udp: { color: '#60a5fa', glow: '#3b82f6' }, // blue
  dns: { color: '#2dd4bf', glow: '#14b8a6' }, // teal
  icmp: { color: '#a78bfa', glow: '#8b5cf6' }, // violet
  arp: { color: '#cbd5e1', glow: '#94a3b8' }, // slate/white
  other: { color: '#7dd3fc', glow: '#38bdf8' }, // soft cyan-white
};

const SUSPICIOUS_COLOR = { color: '#fbbf24', glow: '#f59e0b' }; // amber

export interface PacketVisual {
  count: number;
  durationMs: number;
  color: string;
  glow: string;
  reverse: boolean;
  bidirectional: boolean;
  radius: number;
  opacity: number;
}

interface PacketVisualOptions {
  intensity: PacketIntensity;
  selected: boolean;
  /** Hard cap on particles per edge, lowered automatically for dense graphs. */
  maxParticles: number;
}

const BASE_COUNT: Record<Exclude<TrafficState, 'idle'>, number> = { low: 1, medium: 2, high: 3 };
const BASE_DURATION_MS: Record<Exclude<TrafficState, 'idle'>, number> = { low: 3200, medium: 2000, high: 1150 };
const INTENSITY_COUNT_BONUS: Record<PacketIntensity, number> = { low: 0, normal: 1, high: 2 };
const INTENSITY_SPEED_MUL: Record<PacketIntensity, number> = { low: 1.3, normal: 1, high: 0.75 };

/**
 * Translate a derived flow into concrete render params. Returns null when the
 * edge should show no packets (idle / not animated) so the caller renders
 * nothing at all — idle edges cost zero animation work.
 */
export function packetVisual(flow: PacketFlow, opts: PacketVisualOptions): PacketVisual | null {
  if (!flow.animated || flow.trafficState === 'idle') return null;

  const traffic = flow.trafficState; // narrowed to non-idle above
  let count = BASE_COUNT[traffic] + INTENSITY_COUNT_BONUS[opts.intensity] + (opts.selected ? 1 : 0);
  // Unknown-direction shimmer stays deliberately sparse.
  if (flow.direction === 'unknown') count = Math.min(count, 2);
  count = Math.max(1, Math.min(count, opts.maxParticles + (opts.selected ? 1 : 0)));

  const durationMs = Math.round(
    BASE_DURATION_MS[traffic] * INTENSITY_SPEED_MUL[opts.intensity] * (opts.selected ? 0.85 : 1),
  );

  const palette = flow.suspicious
    ? SUSPICIOUS_COLOR
    : flow.direction === 'unknown'
      ? PROTOCOL_COLORS.arp // neutral for unattributed links
      : PROTOCOL_COLORS[flow.protocol];

  const weak = flow.confidence === 'weak';
  return {
    count,
    durationMs,
    color: palette.color,
    glow: palette.glow,
    reverse: flow.direction === 'reverse',
    bidirectional: flow.direction === 'bidirectional' || flow.direction === 'unknown',
    radius: opts.selected ? 2.6 : 2.1,
    opacity: weak ? 0.55 : opts.selected ? 1 : 0.9,
  };
}
