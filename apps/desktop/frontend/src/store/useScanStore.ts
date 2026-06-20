import { create } from 'zustand';
import { normalizeScan } from '../lib/normalize-scan';
import { buildTopologyViewModel } from '../lib/build-topology-view-model';
import type { RawScanReport } from '../lib/scan-schema';
import type { NormalizedScanReport, TopologyViewModel } from '../lib/models';

const reportStorageKey = 'iad.scan.report.v2';

/**
 * The visible scan lifecycle. The live dashboard renders off this instead of a
 * single isScanning boolean so the status panel can show exactly where a scan
 * is, and so a failure/cancel never blanks the last good topology.
 */
export type ScanStatus =
  | 'idle'
  | 'starting'
  | 'scanning'
  | 'updating'
  | 'completed'
  | 'failed'
  | 'cancelled';

const ACTIVE_STATUSES: ReadonlySet<ScanStatus> = new Set<ScanStatus>(['starting', 'scanning', 'updating']);

/** A single terminal/log line shown in the bottom panel. */
export type LogLevel = 'debug' | 'info' | 'success' | 'warn' | 'error';
export interface LogEntry {
  id: number;
  ts: number;
  level: LogLevel;
  source: string;
  message: string;
}

const MAX_LOG_LINES = 500;
let logSeq = 0;

interface ScanState {
  report: RawScanReport | null;
  normalized: NormalizedScanReport | null;
  // Legacy boolean kept in sync with scanStatus for existing components.
  isScanning: boolean;
  scanProgress: number;
  scanError: string | null;
  logs: LogEntry[];

  // Live lifecycle state.
  scanId: string | null;
  scanStatus: ScanStatus;
  scanPhase: string | null;
  lastUpdatedAt: number | null;
  liveUpdateEnabled: boolean;
  /** A report received while live updates are paused, awaiting manual refresh. */
  pendingReport: RawScanReport | null;
  /** Remembered for the rescan button. */
  lastMode: string;
  lastInterface: string;

  setReport: (report: RawScanReport, source?: string) => void;
  /** Merge a report into the existing topology (dedupe by id, keep positions). */
  mergeReport: (report: RawScanReport) => void;
  /** Merge a synthetic live-discovery payload without replacing stored reports. */
  mergeLiveReport: (report: RawScanReport) => void;
  setScanning: (isScanning: boolean) => void;
  setScanError: (error: string | null) => void;
  pushLog: (level: LogLevel, message: string, source?: string) => void;
  clearLogs: () => void;
  clearReport: () => void;

  // Lifecycle transitions used by the scan controller / event subscription.
  beginScan: (mode: string, iface: string) => void;
  setScanId: (id: string | null) => void;
  onScanStarted: (id: string, phase?: string) => void;
  onScanProgress: (phase: string | null, ts?: number) => void;
  onScanUpdating: () => void;
  completeScan: (ts?: number) => void;
  failScan: (message: string) => void;
  markCancelled: () => void;

  setLiveUpdate: (enabled: boolean) => void;
  toggleLiveUpdate: () => void;
  bufferReport: (report: RawScanReport) => void;
  applyPending: () => void;
}

function normalizeReport(report: RawScanReport): NormalizedScanReport {
  const normalizedBase = normalizeScan(report);
  const topology = buildTopologyViewModel(normalizedBase);
  return { ...normalizedBase, topology };
}

/**
 * Merge an incoming topology into an existing one. Nodes and edges are
 * de-duplicated by their stable id (derived from MAC/IP/hostname upstream), the
 * incoming record wins field-by-field, and previously discovered entities that
 * are absent from the new scan are preserved — we never wipe the map.
 */
function mergeTopology(prev: TopologyViewModel, next: TopologyViewModel): TopologyViewModel {
  const nodes = new Map(prev.nodes.map((n) => [n.id, n]));
  for (const n of next.nodes) nodes.set(n.id, { ...nodes.get(n.id), ...n });
  const edges = new Map(prev.edges.map((e) => [e.id, e]));
  for (const e of next.edges) edges.set(e.id, { ...edges.get(e.id), ...e });
  return { generated: next.generated, nodes: [...nodes.values()], edges: [...edges.values()] };
}

function mergeDevices(prev: NormalizedScanReport['devices'], next: NormalizedScanReport['devices']) {
  const devices = new Map(prev.map((device) => [device.id, device]));
  for (const device of next) devices.set(device.id, { ...devices.get(device.id), ...device });
  return [...devices.values()];
}

function mergeNormalized(prev: NormalizedScanReport, incoming: NormalizedScanReport, keepPreviousMetadata: boolean): NormalizedScanReport {
  const topology = mergeTopology(prev.topology, incoming.topology);
  const devices = mergeDevices(prev.devices, incoming.devices);
  const base = keepPreviousMetadata ? prev : incoming;
  return {
    ...base,
    devices,
    unknownDevices: devices.filter((device) => device.isUnknown),
    topology,
    rawTopologyNodes: topology.nodes,
    rawTopologyEdges: topology.edges,
    summary: {
      ...base.summary,
      deviceCount: devices.length,
      edgeCount: topology.edges.length,
    },
  };
}

function loadStoredReport(): { report: RawScanReport; normalized: NormalizedScanReport } | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.localStorage.getItem(reportStorageKey) ?? (window as unknown as { __iadE2EReport?: string }).__iadE2EReport;
    if (!raw) return null;
    const report = JSON.parse(raw) as RawScanReport;
    return { report, normalized: normalizeReport(report) };
  } catch (error) {
    console.warn('Failed to restore saved scan report', error);
    try {
      window.localStorage.removeItem(reportStorageKey);
    } catch {
      // Ignore storage cleanup failures.
    }
    return null;
  }
}

function saveStoredReport(report: RawScanReport | null) {
  if (typeof window === 'undefined') return;
  try {
    if (report) {
      window.localStorage.setItem(reportStorageKey, JSON.stringify(report));
    } else {
      window.localStorage.removeItem(reportStorageKey);
    }
  } catch (error) {
    console.warn('Failed to persist scan report', error);
  }
}

const stored = loadStoredReport();

export const useScanStore = create<ScanState>((set, get) => ({
  report: stored?.report ?? null,
  normalized: stored?.normalized ?? null,
  isScanning: false,
  scanProgress: 0,
  scanError: null,
  logs: [],

  scanId: null,
  scanStatus: 'idle',
  scanPhase: null,
  lastUpdatedAt: stored ? Date.now() : null,
  liveUpdateEnabled: true,
  pendingReport: null,
  lastMode: 'full',
  lastInterface: '',

  setReport: (report: RawScanReport, _source?: string) => {
    try {
      const normalized = normalizeReport(report);
      saveStoredReport(report);
      set({ report, normalized, scanError: null, lastUpdatedAt: Date.now() });
    } catch (e) {
      console.error('Failed to normalize report', e);
      set({ report, normalized: null });
    }
  },

  mergeReport: (report: RawScanReport) => {
    try {
      const incoming = normalizeReport(report);
      saveStoredReport(report);
      set((state) => {
        const merged: NormalizedScanReport = state.normalized
          ? mergeNormalized(state.normalized, incoming, false)
          : incoming;
        return { report, normalized: merged, scanError: null, lastUpdatedAt: Date.now() };
      });
    } catch (e) {
      console.error('Failed to merge report', e);
      set({ report });
    }
  },

  mergeLiveReport: (report: RawScanReport) => {
    try {
      const incoming = normalizeReport(report);
      set((state) => {
        const merged = state.normalized ? mergeNormalized(state.normalized, incoming, true) : incoming;
        return { report: state.report ?? report, normalized: merged, scanError: null, lastUpdatedAt: Date.now() };
      });
    } catch (e) {
      console.error('Failed to merge live discovery report', e);
    }
  },

  clearReport: () => {
    saveStoredReport(null);
    set({
      report: null,
      normalized: null,
      scanStatus: 'idle',
      scanPhase: null,
      scanError: null,
      isScanning: false,
      pendingReport: null,
      scanId: null,
    });
  },
  setScanning: (isScanning: boolean) => set({ isScanning }),
  setScanError: (scanError: string | null) => set({ scanError }),
  pushLog: (level, message, source = 'agent') =>
    set((state) => {
      const entry: LogEntry = { id: ++logSeq, ts: Date.now(), level, source, message };
      const next = state.logs.length >= MAX_LOG_LINES ? state.logs.slice(state.logs.length - MAX_LOG_LINES + 1) : state.logs;
      return { logs: [...next, entry] };
    }),
  clearLogs: () => set({ logs: [] }),

  beginScan: (mode, iface) =>
    set({
      scanStatus: 'starting',
      scanPhase: 'starting',
      scanError: null,
      isScanning: true,
      scanId: null,
      lastMode: mode,
      lastInterface: iface,
      lastUpdatedAt: Date.now(),
    }),
  setScanId: (scanId) => set({ scanId }),
  onScanStarted: (id, phase = 'starting') =>
    set({ scanId: id, scanStatus: 'scanning', scanPhase: phase, isScanning: true, lastUpdatedAt: Date.now() }),
  onScanProgress: (phase, ts) =>
    set((state) => ({
      scanPhase: phase ?? state.scanPhase,
      // Keep "updating" if a merge is mid-flight; otherwise reflect "scanning".
      scanStatus: state.scanStatus === 'updating' ? 'updating' : 'scanning',
      isScanning: true,
      lastUpdatedAt: ts ?? Date.now(),
    })),
  onScanUpdating: () => set({ scanStatus: 'updating', isScanning: true, lastUpdatedAt: Date.now() }),
  completeScan: (ts) =>
    set({ scanStatus: 'completed', scanPhase: 'completed', isScanning: false, lastUpdatedAt: ts ?? Date.now() }),
  failScan: (message) =>
    // Keep the last good topology on screen; surface the error non-blockingly.
    set({ scanStatus: 'failed', scanError: message, isScanning: false, lastUpdatedAt: Date.now() }),
  markCancelled: () => set({ scanStatus: 'cancelled', isScanning: false, lastUpdatedAt: Date.now() }),

  setLiveUpdate: (enabled) => set({ liveUpdateEnabled: enabled }),
  toggleLiveUpdate: () => set((state) => ({ liveUpdateEnabled: !state.liveUpdateEnabled })),
  bufferReport: (report) => set({ pendingReport: report, lastUpdatedAt: Date.now() }),
  applyPending: () => {
    const pending = get().pendingReport;
    if (!pending) return;
    get().mergeReport(pending);
    set({ pendingReport: null });
  },
}));

export const isActiveScanStatus = (status: ScanStatus): boolean => ACTIVE_STATUSES.has(status);
