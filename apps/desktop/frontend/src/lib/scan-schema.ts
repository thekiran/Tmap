import { z } from 'zod';

const record = z.record(z.unknown());
const number01 = z.number().min(0).max(1);

export const ProbeStatus = z.enum(['success', 'partial', 'no_data', 'skipped', 'failed', 'blocked', 'completed']);
export type ProbeStatus = z.infer<typeof ProbeStatus>;

export const Reachability = z.enum(['self', 'reachable', 'arp_only', 'partial', 'unreachable', 'unknown']);
export type Reachability = z.infer<typeof Reachability>;

export const DecisionQuality = z.enum(['low', 'medium', 'high']);
export type DecisionQuality = z.infer<typeof DecisionQuality>;

export type ConfidenceBand = 'low' | 'medium' | 'high';

export const EvidenceClass = z.enum(['physical', 'l2', 'l3', 'nat', 'isp', 'performance']);
export type EvidenceClass = z.infer<typeof EvidenceClass>;

export const NodeType = z.enum([
  'local_host',
  'interface',
  'subnet',
  'default_gateway',
  'router',
  'gateway',
  'modem_cpe',
  'access_point',
  'mesh_node',
  'repeater',
  'managed_switch',
  'unmanaged_switch_inferred',
  'unknown_l2_segment',
  'workstation',
  'mobile',
  'server',
  'printer',
  'iot',
  'dns_server',
  'isp_gateway',
  'isp_route_hop',
  'public_internet',
  'agent',
  'host',
  'unknown',
]);
export type NodeType = z.infer<typeof NodeType>;

export const EdgeType = z.enum([
  'local_interface',
  'same_subnet',
  'arp_confirmed',
  'mac_table_confirmed',
  'lldp_confirmed',
  'cdp_confirmed',
  'snmp_bridge_confirmed',
  'wifi_association_inferred',
  'ap_bridge_inferred',
  'gateway_default',
  'upstream_private_gateway',
  'route_hop',
  'nat_boundary',
  'isp_boundary',
  'unknown_l2_connection',
]);
export type EdgeType = z.infer<typeof EdgeType>;

export const EdgeCertainty = z.enum(['confirmed', 'inferred', 'unknown']);
export type EdgeCertainty = z.infer<typeof EdgeCertainty>;

const Service = z
  .object({
    port: z.number().optional(),
    protocol: z.string().optional(),
    state: z.string().optional(),
    name: z.string().optional(),
    confidence: z.number().optional(),
    evidence_ids: z.array(z.string()).optional(),
  })
  .passthrough();
export type RawService = z.infer<typeof Service>;

export const RawDevice = z
  .object({
    id: z.string().optional(),
    ip: z.string().optional(),
    ips: z.array(z.string()).optional(),
    ip_addresses: z.array(z.string()).optional(),
    addresses: z.array(record).optional(),
    interfaces: z.array(record).optional(),
    mac: z.string().nullish(),
    vendor: z.union([z.string(), record]).nullish(),
    hostname: z.string().nullish(),
    name: z.string().nullish(),
    type: z.string().optional(),
    roles: z.array(z.string()).optional(),
    role: z.string().nullish(),
    is_gateway: z.boolean().optional(),
    is_agent: z.boolean().optional(),
    reachability: z.string().optional(),
    discovery_sources: z.array(z.string()).optional(),
    oui_vendor: z.string().nullish(),
    hostnames: z.array(z.string()).optional(),
    first_seen: z.string().optional(),
    last_seen: z.string().optional(),
    confidence: z.number().optional(),
    source: z.string().nullish(),
    services: z.array(Service).optional(),
    evidence_ids: z.array(z.string()).optional(),
    security_posture: record.optional(),
    topology: record.optional(),
    device_type: record.optional(),
    os_guess: record.optional(),
  })
  .passthrough();
export type RawDevice = z.infer<typeof RawDevice>;

export const RawEdge = z
  .object({
    id: z.string().optional(),
    source: z.string(),
    target: z.string(),
    type: z.string().optional(),
    layer: z.string().optional(),
    relationship: z.string().optional(),
    physical: z.boolean().optional(),
    inferred: z.boolean().optional(),
    confidence: z.number().optional(),
    confidence_label: z.string().optional(),
    proof_source: z.string().optional(),
    ui_line_style: z.string().optional(),
    line_style: z.string().optional(),
    evidence_ids: z.array(z.string()).optional(),
    reason: z.string().optional(),
  })
  .passthrough();
export type RawEdge = z.infer<typeof RawEdge>;

export const RawEvidence = z
  .object({
    id: z.string().optional(),
    evidence_id: z.string().optional(),
    kind: z.string().optional(),
    source: z.string().optional(),
    summary: z.string().optional(),
    data: record.optional(),
    timestamp: z.string().optional(),
    confidence: z.number().optional(),
    safe_to_display: z.boolean().optional(),
  })
  .passthrough();
export type RawEvidence = z.infer<typeof RawEvidence>;

export const RawProbe = z
  .object({
    name: z.string().optional(),
    probe_name: z.string().optional(),
    category: z.string().optional(),
    status: z.string().optional(),
    duration_ms: z.number().optional(),
    produced_evidence_count: z.number().optional(),
    skipped_reason: z.string().optional(),
    reason: z.string().optional(),
    safety_mode: z.string().optional(),
    output_path: z.string().optional(),
    timeout: z.boolean().optional(),
    error_class: z.string().optional(),
  })
  .passthrough();
export type RawProbe = z.infer<typeof RawProbe>;

export const RawScanReport = z
  .object({
    schema_version: z.string().optional(),
    scan_id: z.string().optional(),
    created_at: z.string().optional(),
    agent: record.optional(),
    scope: record.optional(),
    summary: record.optional(),
    discovery_summary: record.optional(),
    devices: z.array(RawDevice).optional(),
    edges: z.array(RawEdge).optional(),
    evidence: z.array(RawEvidence).optional(),
    evidence_registry: z.array(RawEvidence).optional(),
    probe_inventory: z.array(RawProbe).optional(),
    warnings: z.array(z.union([z.string(), record])).optional(),
    capabilities: z.array(record).optional(),
    interface_selection: record.optional(),
    redaction_mode: z.string().optional(),
    privacy: record.optional(),
    safe_to_share: record.optional(),
    ui: record.optional(),
    device_intel: record.optional(),
    access_classification: record.optional(),
    primary_type: z.string().nullish(),
    category: z.string().nullish(),
    confidence: number01.optional(),
    decision_quality: z.string().optional(),
  })
  .passthrough();
export type RawScanReport = z.infer<typeof RawScanReport>;

export type ValidationResult =
  | { ok: true; data: RawScanReport }
  | { ok: false; errors: { path: string; message: string }[]; raw?: unknown };

export function validateScanJson(text: string): ValidationResult {
  let parsed: unknown;
  try {
    parsed = JSON.parse(text);
  } catch (error) {
    return {
      ok: false,
      errors: [{ path: '(root)', message: `Not valid JSON: ${(error as Error).message}` }],
    };
  }

  const result = RawScanReport.safeParse(parsed);
  if (result.success) return { ok: true, data: result.data };

  return {
    ok: false,
    raw: parsed,
    errors: result.error.issues.map((issue) => ({
      path: issue.path.join('.') || '(root)',
      message: issue.message,
    })),
  };
}
