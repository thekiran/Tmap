/**
 * useScanEvents — subscribes the store to backend scan lifecycle updates and
 * mirrors them into the terminal/log panel as timestamped, leveled lines.
 *
 * Mounted once (in App). Prefers Wails runtime events; if they are unavailable
 * it falls back to polling the latest snapshot every 2s while a scan is active.
 * All listeners/timers are torn down on unmount so a hot-reload or remount never
 * leaves duplicate subscriptions running.
 */
import { useEffect } from 'react';
import { wailsBridge } from './api/AgentBridge';
import { applyRawReport } from './scan-controller';
import { useScanStore, type LogLevel } from '../store/useScanStore';
import type { RawScanReport } from './scan-schema';
import type { MobileFingerprint } from './models';

interface ProgressPayload {
  scanId?: string;
  phase?: string;
  timestamp?: number;
  elapsedMs?: number;
}
interface ResultPayload {
  scanId?: string;
  phase?: string;
  timestamp?: number;
  raw?: string;
}
interface FailurePayload {
  scanId?: string;
  error?: string;
}
interface LiveEvidenceItem {
  type?: string;
  source?: string;
  value?: string;
  timestamp?: string;
  confidenceImpact?: number;
  strength?: string;
}
interface LiveDevicePoolEntry {
  id?: string;
  ip?: string;
  mac?: string;
  hostname?: string;
  vendor?: string;
  status?: string;
  source?: string;
  evidence?: LiveEvidenceItem[];
  mobileFingerprint?: MobileFingerprint | null;
  deviceTypeHint?: string;
  osHint?: string;
  osConfidence?: number;
  osEvidenceSummary?: string[];
}
interface LiveTopologyPayload {
  devices?: LiveDevicePoolEntry[];
  timestamp?: number;
}
interface LiveMobileFingerprintPayload {
  deviceId?: string;
  ipAddresses?: string[];
  hostname?: string;
  mobileFingerprint?: MobileFingerprint;
  updatedAt?: string;
}

function prettyPhase(phase: string | null | undefined): string {
  return (phase ?? '').replace(/_/g, ' ') || 'unknown';
}

function topologyCounts(): { devices: number; links: number } {
  const topo = useScanStore.getState().normalized?.topology;
  if (!topo) return { devices: 0, links: 0 };
  return { devices: topo.nodes.filter((n) => Boolean(n.deviceId)).length, links: topo.edges.length };
}

function isoFromTimestamp(timestamp: number | undefined): string {
  return new Date(timestamp && timestamp > 0 ? timestamp : Date.now()).toISOString();
}

function reachabilityFromLiveStatus(status: string | undefined): string {
  switch ((status ?? '').toLowerCase()) {
    case 'active': return 'reachable';
    case 'recently_seen': return 'partial';
    case 'stale': return 'arp_only';
    case 'unreachable': return 'unreachable';
    default: return 'unknown';
  }
}

function osHintFromMobileClassification(classification: string | undefined): string {
  if (!classification) return 'unknown';
  if (classification.includes('ipados')) return 'ipados';
  if (classification.includes('ios')) return 'ios';
  if (classification.includes('android')) return 'android';
  return 'unknown';
}

function rawEvidenceForDevice(device: LiveDevicePoolEntry): RawScanReport['evidence'] {
  const id = device.id ?? (device.ip ? `ip-${device.ip}` : 'live-device');
  return (device.evidence ?? []).map((item, index) => ({
    id: `${id}-live-ev-${index}`,
    kind: item.type ?? 'metadata',
    source: item.source ?? 'live_discovery',
    summary: item.value ?? item.type ?? 'live evidence',
    timestamp: item.timestamp,
    confidence: Math.max(0, Math.min(1, item.confidenceImpact ?? 0.2)),
    safe_to_display: true,
    data: {
      value: item.value,
      strength: item.strength,
    },
  }));
}

function liveDeviceToRaw(device: LiveDevicePoolEntry): Record<string, unknown> {
  const ip = device.ip ?? device.id ?? 'unknown';
  const id = device.id ?? `ip-${ip}`;
  const evidence = rawEvidenceForDevice(device) ?? [];
  const osHint = device.osHint ?? osHintFromMobileClassification(device.mobileFingerprint?.classification);
  const type = device.deviceTypeHint && device.deviceTypeHint !== 'unknown'
    ? device.deviceTypeHint === 'tablet' ? 'mobile' : device.deviceTypeHint
    : osHint !== 'unknown' ? 'mobile' : 'unknown';
  return {
    id,
    ip,
    ips: [ip],
    ip_addresses: [ip],
    mac: device.mac ?? null,
    hostname: device.hostname ?? null,
    vendor: device.vendor ?? null,
    type,
    roles: [type, osHint].filter((value) => value && value !== 'unknown'),
    reachability: reachabilityFromLiveStatus(device.status),
    discovery_sources: [device.source, 'live_discovery'].filter(Boolean),
    source: device.source ?? 'live_discovery',
    confidence: device.osConfidence ?? device.mobileFingerprint?.confidence ?? 0.35,
    evidence_ids: evidence.map((item) => item.id).filter(Boolean),
    mobileFingerprint: device.mobileFingerprint ?? null,
    deviceTypeHint: device.deviceTypeHint ?? null,
    osHint,
    osConfidence: device.osConfidence ?? device.mobileFingerprint?.confidence ?? null,
    osEvidenceSummary: device.osEvidenceSummary ?? [],
  };
}

function liveReportFromDevices(devices: LiveDevicePoolEntry[], timestamp?: number, scanId = 'live-discovery'): RawScanReport {
  const created = isoFromTimestamp(timestamp);
  const evidence = devices.flatMap((device) => rawEvidenceForDevice(device) ?? []);
  return {
    schema_version: 'live-discovery-v1',
    scan_id: scanId,
    created_at: created,
    devices: devices.map(liveDeviceToRaw) as RawScanReport['devices'],
    evidence,
    summary: {
      device_count: devices.length,
      edge_count: 0,
      evidence_count: evidence.length,
      inferred_only: true,
    },
    discovery_summary: {
      devices_found: devices.length,
      addresses_scanned: devices.length,
    },
    ui: { warnings: [] },
  };
}

function liveReportFromMobileUpdate(payload: LiveMobileFingerprintPayload): RawScanReport | null {
  if (!payload.mobileFingerprint) return null;
  const ip = payload.ipAddresses?.[0] ?? payload.deviceId;
  if (!ip && !payload.deviceId) return null;
  return liveReportFromDevices([{
    id: payload.deviceId,
    ip,
    hostname: payload.hostname,
    mobileFingerprint: payload.mobileFingerprint,
    osHint: osHintFromMobileClassification(payload.mobileFingerprint.classification),
    osConfidence: payload.mobileFingerprint.confidence,
  }], payload.updatedAt ? Date.parse(payload.updatedAt) : Date.now(), 'live-mobile-fingerprint');
}

export function useScanEvents(): void {
  useEffect(() => {
    const store = useScanStore.getState;
    const log = (level: LogLevel, message: string, source = 'agent') => store().pushLog(level, message, source);

    // Track the last emitted phase so progress ticks (~1/s) only log on change.
    let lastPhase = '';

    if (wailsBridge.hasEvents()) {
      const offs: Array<() => void> = [
        wailsBridge.onScanEvent('scan:started', (p: ProgressPayload) => {
          lastPhase = p?.phase ?? '';
          store().onScanStarted(p?.scanId ?? '', p?.phase);
          log('info', `scan started${p?.scanId ? ` · ${p.scanId}` : ''}`, 'agent');
        }),
        wailsBridge.onScanEvent('scan:progress', (p: ProgressPayload) => {
          store().onScanProgress(p?.phase ?? null, p?.timestamp);
          const phase = p?.phase ?? '';
          if (phase && phase !== lastPhase) {
            lastPhase = phase;
            const secs = p?.elapsedMs ? ` (${Math.round(p.elapsedMs / 1000)}s)` : '';
            log('debug', `phase → ${prettyPhase(phase)}${secs}`, 'scan');
          }
        }),
        wailsBridge.onScanEvent('topology:updated', (p: ResultPayload) => {
          if (!p?.raw) return;
          store().onScanUpdating();
          if (applyRawReport(p.raw)) {
            const { devices, links } = topologyCounts();
            log('success', `topology updated · ${devices} devices / ${links} links`, 'topology');
          }
        }),
        wailsBridge.onScanEvent('scan:completed', (p: ResultPayload) => {
          if (p?.raw) applyRawReport(p.raw);
          store().completeScan(p?.timestamp);
          const { devices, links } = topologyCounts();
          log('success', `scan completed · ${devices} devices / ${links} links`, 'agent');
        }),
        wailsBridge.onScanEvent('scan:failed', (p: FailurePayload) => {
          store().failScan(p?.error ?? 'Scan failed.');
          log('error', p?.error ?? 'scan failed', 'agent');
        }),
        wailsBridge.onScanEvent('scan:cancelled', () => {
          store().markCancelled();
          log('warn', 'scan cancelled', 'agent');
        }),
        wailsBridge.onScanEvent('discovery:topology_updated', (p: LiveTopologyPayload) => {
          const devices = Array.isArray(p?.devices) ? p.devices : [];
          if (devices.length === 0) return;
          store().mergeLiveReport(liveReportFromDevices(devices, p?.timestamp));
          const { devices: deviceCount, links } = topologyCounts();
          log('debug', `live discovery topology · ${deviceCount} devices / ${links} links`, 'discovery');
        }),
        wailsBridge.onScanEvent('discovery:device_updated', (p: LiveDevicePoolEntry) => {
          if (!p?.id && !p?.ip) return;
          store().mergeLiveReport(liveReportFromDevices([p], Date.now(), 'live-device-update'));
        }),
        wailsBridge.onScanEvent('discovery:device_mobile_fingerprint_updated', (p: LiveMobileFingerprintPayload) => {
          const report = liveReportFromMobileUpdate(p);
          if (!report) return;
          store().mergeLiveReport(report);
          const label = p.hostname ?? p.ipAddresses?.[0] ?? p.deviceId ?? 'device';
          log('debug', `mobile fingerprint updated · ${label}`, 'discovery');
        }),
      ];
      return () => offs.forEach((off) => off());
    }

    // Polling fallback: only meaningful when the blocking RunScan path is in use
    // and the snapshot getter exists. Pulls the latest completed report so a map
    // still appears without events.
    let lastSeen = '';
    log('info', 'live events unavailable — polling for results', 'system');
    const timer = window.setInterval(async () => {
      const s = store();
      if (s.scanStatus !== 'scanning' && s.scanStatus !== 'updating' && s.scanStatus !== 'starting') return;
      try {
        const raw = await wailsBridge.latestSnapshot();
        if (raw && raw !== lastSeen) {
          lastSeen = raw;
          s.onScanUpdating();
          if (applyRawReport(raw)) {
            const { devices, links } = topologyCounts();
            log('success', `snapshot · ${devices} devices / ${links} links`, 'topology');
          }
        }
      } catch {
        // Ignore transient polling errors; the next tick retries.
      }
    }, 2000);
    return () => window.clearInterval(timer);
  }, []);
}
