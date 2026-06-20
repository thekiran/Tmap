import { useEffect, useState } from 'react';
import { useScanStore, isActiveScanStatus, type ScanStatus } from '../../store/useScanStore';
import { cancelScan, rescan } from '../../lib/scan-controller';
import { Icons } from '../icons/Icon';

const STATUS_META: Record<ScanStatus, { label: string; dot: string; text: string }> = {
  idle: { label: 'Idle', dot: 'bg-zinc-500', text: 'text-zinc-400' },
  starting: { label: 'Starting', dot: 'bg-amber-400 animate-pulse', text: 'text-amber-300' },
  scanning: { label: 'Scanning', dot: 'bg-amber-400 animate-pulse', text: 'text-amber-300' },
  updating: { label: 'Updating topology', dot: 'bg-sky-400 animate-pulse', text: 'text-sky-300' },
  completed: { label: 'Completed', dot: 'bg-emerald-400', text: 'text-emerald-300' },
  failed: { label: 'Failed', dot: 'bg-red-500', text: 'text-red-300' },
  cancelled: { label: 'Cancelled', dot: 'bg-zinc-400', text: 'text-zinc-300' },
};

function prettyPhase(phase: string | null): string {
  if (!phase) return '—';
  return phase.replace(/_/g, ' ');
}

function useRelativeTime(ts: number | null): string {
  const [, force] = useState(0);
  useEffect(() => {
    if (ts == null) return;
    const id = window.setInterval(() => force((n) => n + 1), 1000);
    return () => window.clearInterval(id);
  }, [ts]);
  if (ts == null) return 'never';
  const secs = Math.max(0, Math.round((Date.now() - ts) / 1000));
  if (secs < 2) return 'just now';
  if (secs < 60) return `${secs}s ago`;
  const mins = Math.floor(secs / 60);
  return `${mins}m ${secs % 60}s ago`;
}

interface Props {
  deviceCount: number;
  linkCount: number;
  onFit: () => void;
  onAutoLayout: () => void;
}

/**
 * The always-visible scan status panel above the topology map (requirement 11).
 * Shows the lifecycle status, current phase, last update time, and device/link
 * counts, and hosts the cancel, rescan, fit, auto-layout and live-update
 * controls. It never blocks the map — it sits in a thin bar above it.
 */
export function TopologyStatusBar({ deviceCount, linkCount, onFit, onAutoLayout }: Props) {
  const scanStatus = useScanStore((s) => s.scanStatus);
  const scanPhase = useScanStore((s) => s.scanPhase);
  const lastUpdatedAt = useScanStore((s) => s.lastUpdatedAt);
  const liveUpdateEnabled = useScanStore((s) => s.liveUpdateEnabled);
  const toggleLiveUpdate = useScanStore((s) => s.toggleLiveUpdate);
  const hasPending = useScanStore((s) => s.pendingReport !== null);
  const applyPending = useScanStore((s) => s.applyPending);

  const meta = STATUS_META[scanStatus];
  const relative = useRelativeTime(lastUpdatedAt);
  const active = isActiveScanStatus(scanStatus);

  return (
    <div className="flex min-h-11 flex-wrap items-center gap-x-4 gap-y-1.5 border-b border-zinc-800 bg-zinc-950/85 px-4 py-1.5 text-[11px] text-zinc-400">
      {/* Status */}
      <span className="inline-flex items-center gap-2">
        <span className={`h-2 w-2 shrink-0 rounded-full ${meta.dot}`} />
        <span className={`font-semibold uppercase tracking-[0.14em] ${meta.text}`}>{meta.label}</span>
      </span>

      <Stat label="Phase" value={prettyPhase(scanPhase)} />
      <Stat label="Devices" value={String(deviceCount)} />
      <Stat label="Links" value={String(linkCount)} />
      <Stat label="Updated" value={relative} />

      {/* Controls — right aligned */}
      <div className="ml-auto flex flex-wrap items-center gap-1.5">
        {hasPending && !liveUpdateEnabled && (
          <button
            type="button"
            onClick={applyPending}
            className="inline-flex h-7 items-center gap-1.5 rounded border border-sky-500/50 bg-sky-500/10 px-2.5 font-semibold text-sky-300 hover:bg-sky-500/20"
          >
            <Icons.refresh size={13} /> Apply update
          </button>
        )}

        <button
          type="button"
          onClick={toggleLiveUpdate}
          aria-pressed={liveUpdateEnabled}
          title="Auto-update the map as new results arrive"
          className={[
            'inline-flex h-7 items-center gap-1.5 rounded border px-2.5 font-semibold transition-colors',
            liveUpdateEnabled
              ? 'border-emerald-500/50 bg-emerald-500/10 text-emerald-300'
              : 'border-zinc-700 text-zinc-400 hover:bg-zinc-800',
          ].join(' ')}
        >
          <span className={`h-1.5 w-1.5 rounded-full ${liveUpdateEnabled ? 'bg-emerald-400' : 'bg-zinc-500'}`} />
          Live {liveUpdateEnabled ? 'on' : 'off'}
        </button>

        <button
          type="button"
          onClick={onFit}
          title="Fit the whole map in view"
          className="inline-flex h-7 items-center gap-1.5 rounded border border-zinc-700 px-2.5 font-semibold text-zinc-300 hover:bg-zinc-800"
        >
          <Icons.fit size={13} /> Fit view
        </button>

        <button
          type="button"
          onClick={onAutoLayout}
          title="Recompute the automatic layout"
          className="inline-flex h-7 items-center gap-1.5 rounded border border-zinc-700 px-2.5 font-semibold text-zinc-300 hover:bg-zinc-800"
        >
          <Icons.layers size={13} /> Auto layout
        </button>

        {active ? (
          <button
            type="button"
            onClick={() => void cancelScan()}
            className="inline-flex h-7 items-center gap-1.5 rounded border border-red-500/50 bg-red-500/10 px-2.5 font-semibold text-red-300 hover:bg-red-500/20"
          >
            Stop
          </button>
        ) : (
          <button
            type="button"
            onClick={() => void rescan()}
            className="inline-flex h-7 items-center gap-1.5 rounded bg-blue-500 px-3 font-semibold text-white shadow-sm shadow-blue-500/20 hover:bg-blue-600"
          >
            <Icons.refresh size={13} /> Rescan
          </button>
        )}
      </div>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span className="font-mono uppercase tracking-[0.12em] text-zinc-600">{label}</span>
      <span className="font-mono text-zinc-300">{value}</span>
    </span>
  );
}
