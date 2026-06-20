import { useEffect, useMemo, useState } from 'react';
import { wailsBridge, type NetworkInterface } from '../../lib/api/AgentBridge';
import { startScan } from '../../lib/scan-controller';
import { useScanStore } from '../../store/useScanStore';

function ipv4Of(iface: NetworkInterface): string | null {
  return iface.addresses?.find((a) => a.version === 4)?.ip ?? null;
}

/** Rank interfaces Wireshark-style: live, real, routable adapters first. */
function rank(iface: NetworkInterface): number {
  let score = 0;
  if (iface.up) score += 4;
  const ip = ipv4Of(iface);
  if (ip && !ip.startsWith('169.254.')) score += 4; // has a routable IPv4
  if (!iface.virtual) score += 2;
  if (!iface.loopback) score += 1;
  if (iface.selected) score += 8;
  return score;
}

export function LaunchScreen() {
  const isScanning = useScanStore((s) => s.isScanning);
  const scanError = useScanStore((s) => s.scanError);

  const [interfaces, setInterfaces] = useState<NetworkInterface[] | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [selected, setSelected] = useState<string | null>(null);

  const loadInterfaces = useMemo(
    () => async () => {
      setLoadError(null);
      setInterfaces(null);
      try {
        const list = await wailsBridge.listInterfaces();
        const sorted = [...list].sort((a, b) => rank(b) - rank(a));
        setInterfaces(sorted);
        setSelected(sorted.find((i) => i.selected)?.name ?? sorted[0]?.name ?? null);
      } catch (e) {
        setLoadError((e as Error).message);
        setInterfaces([]);
      }
    },
    [],
  );

  useEffect(() => {
    void loadInterfaces();
  }, [loadInterfaces]);

  // Always scan at maximum capability — full single-file report, nmap service
  // discovery, access classification, and the widest device discovery sweep.
  // startScan is non-blocking: it flips the lifecycle to "starting" and the app
  // immediately swaps to the live topology dashboard.
  const start = () => {
    if (!selected || isScanning) return;
    void startScan('full', selected);
  };

  return (
    <div className="flex h-full min-h-0 flex-col items-center overflow-hidden bg-white px-4 py-5 sm:px-6 sm:py-8 dark:bg-black/20">
      <div className="flex h-full min-h-0 w-full max-w-3xl flex-col">
        {/* Header — fixed */}
        <div className="shrink-0">
          <div className="mb-1 font-mono text-[11px] uppercase tracking-[0.28em] text-zinc-500">
            Internet Access Detector
          </div>
          <h1 className="mb-1 text-xl font-bold text-zinc-900 sm:text-2xl dark:text-zinc-100">
            Select a network interface
          </h1>
          <p className="mb-4 text-[13px] leading-relaxed text-zinc-500 sm:text-sm">
            Pick the adapter to map, then start the scan. Every scan runs at maximum depth to
            discover as many devices as possible. Only your own private network is probed —
            no logins, no brute force, no neighbor scanning.
          </p>
        </div>

        {/* Interface list — fills the remaining height and scrolls internally
            (no visible scrollbar), so the page itself never shows a side bar. */}
        <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-zinc-200 dark:border-zinc-800">
          <div className="grid shrink-0 grid-cols-[1.6fr_1fr_0.8fr_auto] gap-2 border-b border-zinc-200 bg-zinc-50 px-4 py-2 font-mono text-[10px] uppercase tracking-wider text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/60">
            <span>Interface</span>
            <span>IPv4 / CIDR</span>
            <span>Status</span>
            <span>Type</span>
          </div>

          <div className="no-scrollbar min-h-0 flex-1 overflow-y-auto">
            {interfaces === null ? (
              <div className="px-4 py-8 text-center text-sm text-zinc-500">Loading interfaces…</div>
            ) : interfaces.length === 0 ? (
              <div className="px-4 py-8 text-center text-sm text-zinc-500">
                No interfaces found.{' '}
                <button onClick={() => void loadInterfaces()} className="text-blue-500 hover:underline">
                  Retry
                </button>
              </div>
            ) : (
              interfaces.map((iface) => {
                const ip = ipv4Of(iface);
                const active = selected === iface.name;
                return (
                  <button
                    key={iface.name}
                    onClick={() => setSelected(iface.name)}
                    onDoubleClick={start}
                    className={[
                      'grid w-full grid-cols-[1.6fr_1fr_0.8fr_auto] items-center gap-2 border-b border-zinc-100 px-4 py-2.5 text-left text-sm transition-colors last:border-b-0 dark:border-zinc-800/60',
                      active
                        ? 'bg-blue-50 dark:bg-blue-500/10'
                        : 'hover:bg-zinc-50 dark:hover:bg-zinc-800/40',
                    ].join(' ')}
                  >
                    <span className="flex items-center gap-2 truncate">
                      <span
                        className={[
                          'h-2 w-2 shrink-0 rounded-full',
                          iface.up ? 'bg-emerald-500' : 'bg-zinc-400/60',
                        ].join(' ')}
                      />
                      <span className="truncate font-medium text-zinc-800 dark:text-zinc-200">{iface.name}</span>
                    </span>
                    <span className="truncate font-mono text-[12px] text-zinc-600 dark:text-zinc-400">
                      {ip ? `${ip}${iface.cidr ? ` · ${iface.cidr}` : ''}` : '—'}
                    </span>
                    <span className="font-mono text-[11px] text-zinc-500">{iface.up ? 'up' : 'down'}</span>
                    <span className="flex justify-end gap-1">
                      {iface.loopback && <Tag>loopback</Tag>}
                      {iface.virtual && <Tag>virtual</Tag>}
                      {iface.selected && <Tag tone="blue">default</Tag>}
                    </span>
                  </button>
                );
              })
            )}
          </div>
        </div>

        {loadError && (
          <div className="mt-3 shrink-0 rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs text-amber-700 dark:text-amber-300">
            Could not list interfaces: {loadError}
          </div>
        )}

        {/* Maximum-depth scan — no profile to choose, always the deepest sweep. */}
        <div className="mt-4 shrink-0 rounded-lg border border-blue-300/60 bg-blue-50 px-4 py-2.5 dark:border-blue-400/30 dark:bg-blue-500/10">
          <div className="flex items-center gap-2">
            <span className="h-2 w-2 shrink-0 rounded-full bg-blue-500" />
            <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-blue-600 dark:text-blue-300">
              Maximum scan
            </span>
            <span className="truncate text-[12px] text-zinc-600 dark:text-zinc-300">
              Full subnet sweep · active ARP · TCP probing · Nmap · access classification
            </span>
          </div>
        </div>

        {/* Start — fixed at the bottom */}
        <div className="mt-4 flex shrink-0 flex-wrap items-center gap-3">
          <button
            onClick={start}
            disabled={!selected || isScanning}
            className="inline-flex items-center gap-2 rounded-lg bg-blue-500 px-5 py-2.5 text-sm font-semibold text-white shadow-sm shadow-blue-500/20 transition-colors hover:bg-blue-600 disabled:opacity-50"
          >
            {isScanning ? (
              <>
                <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/60 border-t-transparent" />
                Scanning…
              </>
            ) : (
              <>Start full scan</>
            )}
          </button>
          <span className="text-xs text-zinc-500">
            {selected ? <>Target: <span className="font-mono text-zinc-600 dark:text-zinc-300">{selected}</span></> : 'Choose an interface above'}
          </span>
        </div>

        {scanError && (
          <div className="mt-3 shrink-0 rounded-md border border-red-500/40 bg-red-500/10 p-3">
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-red-500">Scan failed</div>
            <div className="font-mono text-[11px] leading-relaxed text-zinc-600 dark:text-zinc-300">{scanError}</div>
          </div>
        )}
      </div>
    </div>
  );
}

function Tag({ children, tone = 'zinc' }: { children: React.ReactNode; tone?: 'zinc' | 'blue' }) {
  const cls =
    tone === 'blue'
      ? 'border-blue-300 text-blue-600 dark:border-blue-400/50 dark:text-blue-300'
      : 'border-zinc-300 text-zinc-500 dark:border-zinc-700 dark:text-zinc-400';
  return (
    <span className={`rounded border px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wide ${cls}`}>
      {children}
    </span>
  );
}
