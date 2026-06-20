import type {
  DecisionQuality,
  EdgeCertainty,
  EdgeType,
  EvidenceClass,
  NodeType,
  ProbeStatus,
  RawScanReport,
  Reachability,
} from './scan-schema';

export interface DeviceService {
  name: string;
  port?: number;
  proto?: string;
  protocol?: string;
  state?: string;
  confidence?: number;
  evidenceIds: string[];
}

export interface RiskFinding {
  id: string;
  deviceId: string;
  deviceLabel: string;
  severity: string;
  title: string;
  description: string;
  recommendation: string | null;
  evidenceIds: string[];
}

export interface MobileEvidenceItem {
  id: string;
  type: string;
  value: string;
  osHint: 'ios' | 'ipados' | 'android' | 'unknown';
  confidenceImpact: number;
  strength: 'strong' | 'medium' | 'weak';
  source: string;
  timestamp: string;
  explanation: string;
}

export interface MobileConflictItem {
  reason: string;
  iosEvidenceIds: string[];
  androidEvidenceIds: string[];
  severity: 'info' | 'warning';
  resolutionHint: string;
}

export interface MobileFingerprint {
  classification: string;
  iosScore: number;
  androidScore: number;
  ipadScore: number;
  confidence: number;
  evidence: MobileEvidenceItem[];
  conflicts: MobileConflictItem[];
  warnings: string[];
  lastUpdatedAt: string | null;
  whyThisClassification: string | null;
  whyNotCertain: string | null;
}

export interface NetworkDevice {
  id: string;
  ips: string[];
  ip: string;
  mac: string | null;
  vendor: string | null;
  hostname: string | null;
  type: NodeType;
  role: string | null;
  roles: string[];
  isGateway: boolean;
  isAgent: boolean;
  isUnknown: boolean;
  reachability: Reachability;
  discoverySources: string[];
  confidence: number;
  source: string | null;
  services: DeviceService[];
  explanation: string | null;
  limitations: string | null;
  rawProbeRefs: string[];
  evidenceCount: number;
  rawSources: string[];
  wireless: Record<string, unknown> | null;
  riskLevel: string | null;
  riskFindings: RiskFinding[];
  mobileFingerprint: MobileFingerprint | null;
  deviceTypeHint: 'phone' | 'tablet' | 'computer' | 'iot' | 'router' | 'unknown' | null;
  osHint: 'ios' | 'ipados' | 'android' | 'unknown' | null;
  osConfidence: number | null;
  osEvidenceSummary: string[];
  raw: Record<string, unknown>;
}

export interface EvidenceRecord {
  id: string;
  probeName: string;
  status: ProbeStatus;
  confidence: number;
  timestamp: string;
  evidenceClass: EvidenceClass;
  reason: string | null;
  limitations: string | null;
  data: Record<string, unknown> | null;
  warnings: string[];
  errors: string[];
  emptyEvidenceWarning: boolean;
  source: string;
  kind: string;
  summary: string;
  safeToDisplay: boolean;
}

export type ProbeResult = EvidenceRecord;

export interface GatewayHop {
  hop: number;
  ip: string;
  kind: string;
  rttMs: number | null;
  label: string;
  private: boolean;
  note: string | null;
}

export interface AccessCandidate {
  type: string;
  score: number;
  note: string | null;
}

export interface ConfidenceFactor {
  factor: string;
  weight: number;
  contribution: number;
  direction: 'up' | 'down';
  detail: string | null;
}

export interface NextProbe {
  name: string;
  gain: number;
  requires: string | null;
  tier: EvidenceClass;
  detail: string | null;
  reason?: string;
  expectedEvidence?: string[];
  safety?: string;
}

export interface Advisory {
  level: 'warn' | 'info' | 'danger';
  text: string;
  code?: string;
}

export interface ProbeInventoryItem {
  name: string;
  category: string;
  status: string;
  durationMs: number | null;
  producedEvidenceCount: number | null;
  safetyMode: string | null;
  outputPath: string | null;
  reason: string | null;
  timeout: boolean;
}

export interface OpenService {
  id: string;
  deviceId: string;
  deviceLabel: string;
  port: number | null;
  protocol: string;
  name: string;
  state: string;
  confidence: number | null;
  evidenceIds: string[];
}

export interface AccessClassification {
  primaryType: string | null;
  category: string | null;
  subtype: string | null;
  confidence: number;
  contextConfidence: number;
  decisionQuality: DecisionQuality;
  state: string | null;
  safeToDisplayAsFinal: boolean;
  uncertaintyReasons: string[];
  candidates: AccessCandidate[];
}

export interface TopologyNode {
  [key: string]: unknown;
  id: string;
  type: NodeType;
  label: string;
  sublabel: string | null;
  certainty: EdgeCertainty;
  layers: ('l2' | 'l3' | 'nat' | 'isp' | 'unknown')[];
  badge: string | null;
  deviceId: string | null;
  accent: boolean;
  position: { x: number; y: number };
  confidence: number;
  roles: string[];
  isGateway: boolean;
  isAgent: boolean;
  isUnknown: boolean;
  ip?: string | null;
  mac?: string | null;
  vendor?: string | null;
  role?: string | null;
  evidenceCount?: number;
  wireless?: Record<string, unknown> | null;
  rawSources?: string[];
  mobileFingerprint?: MobileFingerprint | null;
  deviceTypeHint?: NetworkDevice['deviceTypeHint'];
  osHint?: NetworkDevice['osHint'];
  osConfidence?: number | null;
  osEvidenceSummary?: string[];
}

export interface TopologyEdge {
  [key: string]: unknown;
  id: string;
  source: string;
  target: string;
  type: EdgeType;
  certainty: EdgeCertainty;
  tier: 'l2' | 'l3' | 'nat' | 'isp';
  confidence: number;
  label: string;
  basis: string;
  boundary: 'NAT' | 'ISP' | null;
  thin: boolean;
  physical: boolean;
  inferred: boolean;
  lineStyle: 'solid' | 'dashed' | 'dotted';
  layers: ('l2' | 'l3' | 'nat' | 'isp' | 'unknown')[];
  relation?: string | null;
  medium?: string | null;
  explanation?: string | null;
  warnings?: string[];
  evidence?: Record<string, unknown>[];
  evidenceIds?: string[];
  rawEdge?: Record<string, unknown>;
}

export interface TopologyViewModel {
  generated: boolean;
  nodes: TopologyNode[];
  edges: TopologyEdge[];
}

export interface DiscoverySummary {
  cidr: string;
  addressesScanned: number;
  devicesFound: number;
  arpFound: number;
  icmpFound: number;
  tcpFound: number;
  mdnsFound: number;
  ssdpFound: number;
  llmnrFound: number;
  netbiosFound: number;
  nmapFound: number;
  scanDurationMs: number;
}

export interface NormalizedScanReport {
  scanId: string;
  schemaVersion: string | null;
  createdAt: string;
  status: string;
  mode: string;
  durationMs: number | null;
  safeMode: boolean;
  sourceProfile: string | null;
  raw: RawScanReport;

  primaryType: string | null;
  isUnknown: boolean;
  category: string | null;
  confidence: number;
  classificationConfidence: number;
  contextConfidence: number;
  decisionQuality: DecisionQuality;
  uncertaintyReasons: string[];
  candidates: AccessCandidate[];
  access: AccessClassification;

  selectedInterface: {
    name: string;
    type: string;
    ipv4: string | null;
    prefix: number | null;
    mac: string | null;
    mtu: number | null;
    gateway: string | null;
    dns: string[];
    linkSpeedMbps: number | null;
    dhcp: boolean | null;
  } | null;
  gatewayChain: GatewayHop[];
  gatewayDevice: NetworkDevice | null;
  nat: { type: string | null; layers: number | null; publicReachable: boolean | null; note: string | null } | null;
  ipv6: { available: boolean; globalAddress: string | null; delegatedPrefix: string | null; note: string | null } | null;
  publicIp: {
    address: string | null;
    ptr: string | null;
    asn: string | null;
    org: string | null;
    city: string | null;
    country: string | null;
    geoConfidence: number | null;
  } | null;
  performance: {
    downstreamMbps: number | null;
    upstreamMbps: number | null;
    latencyMs: number | null;
    jitterMs: number | null;
    lossPct: number | null;
  } | null;

  confidenceBreakdown: ConfidenceFactor[];
  nextBestProbes: NextProbe[];
  warnings: Advisory[];

  devices: NetworkDevice[];
  unknownDevices: NetworkDevice[];
  openServices: OpenService[];
  riskFindings: RiskFinding[];
  evidence: EvidenceRecord[];
  probes: ProbeInventoryItem[];

  topologyGenerated: boolean;
  rawTopologyNodes: Record<string, unknown>[];
  rawTopologyEdges: Record<string, unknown>[];
  topology: TopologyViewModel;

  summary: {
    deviceCount: number;
    edgeCount: number;
    evidenceCount: number;
    probeCount: number;
    warningCount: number;
    serviceCount: number;
    riskFindingCount: number;
    inferredOnly: boolean;
    physicalEdgeCount: number;
  };

  discoverySummary: DiscoverySummary | null;
}

export interface LayoutPosition {
  x: number;
  y: number;
}

export type ThemeMode = 'dark' | 'light';
export type LayoutEngine = 'elk_layered' | 'force' | 'manual';

export interface UISettings {
  theme: ThemeMode;
  colorMode: 'black_white';
  layoutEngine: LayoutEngine;
  showLowConfidenceEdges: boolean;
  showUnknownSegments: boolean;
  showIspRouteContext: boolean;
  showL2: boolean;
  showL3: boolean;
  showNat: boolean;
  persistNodePositions: boolean;
}
