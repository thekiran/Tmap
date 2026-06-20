/**
 * upstream-intel.ts — frontend accessor for the agent's upstream-gateway
 * enrichment (see agent/internal/upstream). The data lives in the raw report's
 * passthrough `device_intel` section; this module reads it defensively and
 * shapes it for the Inspector. It NEVER fabricates values — missing fields stay
 * undefined and render as "Unknown / inferred".
 */
import type { RawScanReport } from './scan-schema';

export interface UpstreamReachability {
  icmp: boolean;
  tcpReachable: boolean;
  avgLatencyMs?: number;
  minLatencyMs?: number;
  maxLatencyMs?: number;
  packetLoss?: number;
  ttl?: number;
  hopDistance?: number;
  directlyReachable: boolean;
  method: string;
  note?: string;
}

export interface UpstreamRouting {
  kind: string;
  doubleNat: boolean;
  privateUpstream: boolean;
  sameSubnetAsAgent: boolean;
  hopDistance?: number;
  notes: string[];
}

export interface UpstreamEvidenceItem {
  type: string;
  value: string;
  source: string;
  confidenceImpact: number;
  timestamp: string;
}

export interface UpstreamService {
  port: number;
  name?: string;
  protocol?: string;
}

export interface UpstreamIntel {
  ip: string;
  confidence: number;
  vendor?: string;
  tags: string[];
  reachability?: UpstreamReachability;
  routing?: UpstreamRouting;
  evidence: UpstreamEvidenceItem[];
  warnings: string[];
  services: UpstreamService[];
  http: { server?: string; title?: string; statusCode?: number; wwwAuthenticate?: string }[];
  tls: { cn?: string; issuer?: string; sans?: string[] }[];
}

type AnyRecord = Record<string, unknown>;
const asRecord = (v: unknown): AnyRecord => (v && typeof v === 'object' ? (v as AnyRecord) : {});
const asArray = (v: unknown): unknown[] => (Array.isArray(v) ? v : []);
const num = (v: unknown): number | undefined => (typeof v === 'number' ? v : undefined);
const str = (v: unknown): string | undefined => (typeof v === 'string' && v ? v : undefined);

function mapReachability(v: unknown): UpstreamReachability | undefined {
  if (!v || typeof v !== 'object') return undefined;
  const r = v as AnyRecord;
  return {
    icmp: Boolean(r.icmp),
    tcpReachable: Boolean(r.tcp_reachable),
    avgLatencyMs: num(r.avg_latency_ms),
    minLatencyMs: num(r.min_latency_ms),
    maxLatencyMs: num(r.max_latency_ms),
    packetLoss: num(r.packet_loss),
    ttl: num(r.ttl),
    hopDistance: num(r.hop_distance),
    directlyReachable: Boolean(r.directly_reachable),
    method: str(r.method) ?? 'none',
    note: str(r.note),
  };
}

function mapRouting(v: unknown): UpstreamRouting | undefined {
  if (!v || typeof v !== 'object') return undefined;
  const r = v as AnyRecord;
  return {
    kind: str(r.kind) ?? 'unknown',
    doubleNat: Boolean(r.double_nat),
    privateUpstream: Boolean(r.private_upstream),
    sameSubnetAsAgent: Boolean(r.same_subnet_as_agent),
    hopDistance: num(r.hop_distance),
    notes: asArray(r.notes).map(String),
  };
}

/** Find the enrichment record for a device by matching any of its IPs. */
export function findUpstreamIntel(raw: RawScanReport | undefined, ips: string[]): UpstreamIntel | null {
  const intel = asRecord(asRecord(raw as unknown).device_intel);
  const devices = asArray(intel.devices);
  if (devices.length === 0 || ips.length === 0) return null;
  const wanted = new Set(ips.filter(Boolean));

  for (const d of devices) {
    const dev = asRecord(d);
    const deviceIPs = asArray(dev.ip_addresses).map(String);
    if (!deviceIPs.some((ip) => wanted.has(ip))) continue;

    const tags = asArray(dev.classification_tags).map(String);
    const reachability = mapReachability(dev.reachability);
    // Only treat it as "enriched" if the upstream phase actually attached data.
    if (tags.length === 0 && !reachability && !dev.routing_evidence) return null;

    const vendorRec = asRecord(dev.vendor);
    return {
      ip: deviceIPs[0] ?? ips[0],
      confidence: num(dev.confidence) ?? 0,
      vendor: str(vendorRec.fingerprint_vendor) ?? str(vendorRec.oui_vendor),
      tags,
      reachability,
      routing: mapRouting(dev.routing_evidence),
      evidence: asArray(dev.intel_evidence).map((e) => {
        const ev = asRecord(e);
        return {
          type: str(ev.type) ?? '',
          value: str(ev.value) ?? '',
          source: str(ev.source) ?? '',
          confidenceImpact: num(ev.confidence_impact) ?? 0,
          timestamp: str(ev.timestamp) ?? '',
        };
      }),
      warnings: asArray(dev.enrichment_warnings).map(String),
      services: asArray(dev.services).map((s) => {
        const svc = asRecord(s);
        return { port: num(svc.port) ?? 0, name: str(svc.name), protocol: str(svc.protocol) };
      }),
      http: asArray(dev.http_fingerprints).map((h) => {
        const o = asRecord(h);
        return { server: str(o.server_header) ?? str(o.server), title: str(o.title), statusCode: num(o.status_code), wwwAuthenticate: str(o.www_authenticate) };
      }),
      tls: asArray(dev.tls_fingerprints).map((tl) => {
        const o = asRecord(tl);
        return { cn: str(o.cn), issuer: str(o.issuer), sans: asArray(o.sans).map(String) };
      }),
    };
  }
  return null;
}

/**
 * The human headline for an upstream device, using the exact wording the spec
 * asks for. Returns null when the device is an ordinary same-subnet gateway.
 */
export function routingHeadline(intel: UpstreamIntel): string | null {
  const kind = intel.routing?.kind ?? '';
  const reachable = Boolean(intel.reachability?.icmp || intel.reachability?.tcpReachable);
  switch (kind) {
    case 'unreachable_inferred':
      return 'Inferred upstream gateway — not directly reachable';
    case 'double_nat_upstream':
      return 'Possible double NAT upstream router';
    case 'isp_cpe':
      return 'Possible ISP CPE / modem / ONT';
    case 'virtual_or_docker':
      return 'Virtual / Docker network artifact';
    case 'upstream_private_gateway':
      return reachable ? 'Upstream private gateway' : 'Inferred upstream gateway — not directly reachable';
    case 'unknown':
      return intel.tags.includes('UNKNOWN_INFRASTRUCTURE') ? 'Unknown but reachable infrastructure device' : null;
    default:
      return null;
  }
}

/** Friendly label for a classification tag (UPSTREAM_GATEWAY → "Upstream gateway"). */
export function tagLabel(tag: string): string {
  return tag.replace(/_/g, ' ').toLowerCase().replace(/^\w/, (c) => c.toUpperCase());
}
