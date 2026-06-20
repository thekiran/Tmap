/**
 * scan-controller.ts — imperative scan lifecycle actions.
 *
 * These are plain functions (not a hook) so any button can trigger a scan
 * without prop-drilling. They drive the Zustand scan store and talk to the Go
 * backend through the Wails bridge. The actual progress/results are applied by
 * the event subscription in useScanEvents (or the polling fallback there).
 */
import { useScanStore, type LogLevel } from '../store/useScanStore';
import { wailsBridge, type ScanMode } from './api/AgentBridge';
import { validateScanJson } from './scan-schema';

const log = (level: LogLevel, message: string, source = 'console') => useScanStore.getState().pushLog(level, message, source);

/**
 * Validate a raw report string and route it into the store. When live updates
 * are paused, the report is buffered for a manual refresh instead of merged.
 * `force` (used by the completed-by-blocking-call path) merges regardless.
 */
export function applyRawReport(raw: string, force = false): boolean {
  const store = useScanStore.getState();
  const result = validateScanJson(raw);
  if (!result.ok) {
    store.failScan('Scan output failed validation — the agent returned unexpected JSON.');
    log('error', 'scan output failed schema validation', 'console');
    return false;
  }
  if (force || store.liveUpdateEnabled) {
    store.mergeReport(result.data);
  } else {
    store.bufferReport(result.data);
  }
  return true;
}

/**
 * Start a scan. Prefers the non-blocking, event-driven StartScan binding; if it
 * is unavailable (older bindings or browser dev), falls back to the blocking
 * RunScan and applies its result when it resolves.
 */
export async function startScan(mode: ScanMode = 'full', iface = ''): Promise<void> {
  const store = useScanStore.getState();
  store.beginScan(mode, iface);
  log('info', `requesting scan · interface ${iface || 'auto'} · mode ${mode}`, 'console');
  try {
    const id = await wailsBridge.startScan(mode, iface);
    if (id) {
      store.setScanId(id);
      return; // events take over from here
    }
    await legacyRunScan(mode, iface);
  } catch (error) {
    store.failScan((error as Error).message);
    log('error', (error as Error).message, 'console');
  }
}

async function legacyRunScan(mode: ScanMode, iface: string): Promise<void> {
  const store = useScanStore.getState();
  store.onScanStarted(`local-${Date.now()}`);
  try {
    const text = await wailsBridge.runScan(mode, iface);
    if (!text) {
      store.failScan('Agent returned no output. Is iad-agent installed next to the app?');
      return;
    }
    if (applyRawReport(text, true)) {
      store.completeScan();
    }
  } catch (error) {
    store.failScan((error as Error).message);
  }
}

/** Cancel the in-flight scan. The backend confirms via the scan:cancelled event. */
export async function cancelScan(): Promise<void> {
  log('warn', 'cancel requested', 'console');
  try {
    await wailsBridge.cancelScan();
  } catch {
    // Best effort — the lifecycle event (or lack of one) is authoritative.
  }
}

/** Re-run the most recent scan with the same interface/mode. */
export async function rescan(): Promise<void> {
  const { lastMode, lastInterface } = useScanStore.getState();
  await startScan((lastMode as ScanMode) || 'full', lastInterface || '');
}
