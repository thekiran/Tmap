import { useEffect, useRef, useState } from 'react';
import iadMark from '../../assets/mark-iad.svg';
import { useUIStore, type ScreenId } from '../../store/useUIStore';
import { useScanStore } from '../../store/useScanStore';
import { useImport } from '../../lib/useImport';
import { wailsBridge } from '../../lib/api/AgentBridge';
import { Icons, type IconKey } from '../icons/Icon';
import {
  Quit,
  WindowMinimise,
  WindowToggleMaximise,
} from '../../../wailsjs/runtime/runtime';

type PrimaryScreen = Extract<ScreenId, 'topology' | 'devices' | 'evidence'>;

interface CommandItem {
  label?: string;
  onClick?: () => void;
  disabled?: boolean;
  separator?: boolean;
}

const dragRegion = {
  WebkitAppRegion: 'drag',
  '--wails-draggable': 'drag',
} as React.CSSProperties;
const noDragRegion = {
  WebkitAppRegion: 'no-drag',
  '--wails-draggable': 'no-drag',
} as React.CSSProperties;

const primaryScreens: Array<{ id: PrimaryScreen; label: string; icon: IconKey }> = [
  { id: 'topology', label: 'Topology', icon: 'topology' },
  { id: 'devices', label: 'Devices', icon: 'devices' },
  { id: 'evidence', label: 'Evidence', icon: 'evidence' },
];

export function MenuBar() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [aboutOpen, setAboutOpen] = useState(false);
  const barRef = useRef<HTMLDivElement>(null);

  const ui = useUIStore();
  const report = useScanStore((s) => s.report);
  const scan = useScanStore((s) => s.normalized);
  const clearReport = useScanStore((s) => s.clearReport);
  const isScanning = useScanStore((s) => s.isScanning);
  const { importViaDialog, runAgent, busy } = useImport();

  useEffect(() => {
    const onDown = (e: MouseEvent) => {
      if (barRef.current && !barRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setMenuOpen(false);
        setAboutOpen(false);
      }
    };
    document.addEventListener('mousedown', onDown);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDown);
      document.removeEventListener('keydown', onKey);
    };
  }, []);

  const exportReport = async () => {
    if (!report) return;
    const name = `iad-scan-${report.scan_id ?? 'report'}.json`;
    await wailsBridge.saveExport(name, JSON.stringify(report, null, 2));
  };

  const copyJson = async () => {
    if (!report) return;
    try {
      await navigator.clipboard.writeText(JSON.stringify(report, null, 2));
    } catch {
      /* Clipboard can be unavailable in browser preview. */
    }
  };

  const reportMenu: CommandItem[] = [
    { label: 'New scan', onClick: () => clearReport() },
    { label: 'Import report', onClick: () => void importViaDialog() },
    { label: 'Export report JSON', onClick: () => void exportReport(), disabled: !report },
    { label: 'Copy report JSON', onClick: () => void copyJson(), disabled: !report },
    { separator: true },
    { label: 'Reset map layout', onClick: () => ui.resetLayoutPositions(), disabled: !report },
    { label: 'About IAD', onClick: () => setAboutOpen(true) },
  ];

  const runMenuItem = (item: CommandItem) => {
    if (item.disabled) return;
    item.onClick?.();
    setMenuOpen(false);
  };

  const toggleSearch = () => {
    if (!ui.isSearchOpen) {
      ui.setActiveScreen('devices');
    }
    ui.toggleSearch();
  };

  const statusText = isScanning ? 'Scanning' : busy ? 'Working' : scan ? 'Report loaded' : 'Ready';
  const statusTone = isScanning || busy ? 'bg-amber-400' : scan ? 'bg-emerald-400' : 'bg-zinc-500';
  const scanId = scan?.scanId ?? 'No scan loaded';
  const interfaceName = scan?.selectedInterface?.name ?? 'Interface not selected';
  const publicIp = scan?.publicIp?.address ?? 'Public IP unknown';
  const confidence = scan ? `${Math.round(scan.confidence * 100)}%` : '--';
  const deviceCount = scan?.summary.deviceCount ?? 0;
  const evidenceCount = scan?.summary.evidenceCount ?? 0;

  return (
    <div
      ref={barRef}
      className="relative flex h-full w-full items-stretch overflow-visible bg-[#0e0e10] text-zinc-100"
      style={dragRegion}
    >
      <div className="flex min-w-0 flex-1 items-center gap-3 px-3">
        <div className="flex min-w-[214px] items-center gap-2.5">
          <img src={iadMark} alt="IAD" className="h-8 w-8 shrink-0" />
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <span className="font-mono text-[15px] font-bold tracking-[0.14em] text-white">IAD</span>
              <span className="hidden whitespace-nowrap text-[13px] font-semibold text-zinc-300 lg:inline">
                Internet Access Detector
              </span>
            </div>
            <div className="truncate font-mono text-[10px] uppercase tracking-[0.12em] text-zinc-500">
              {scanId}
            </div>
          </div>
        </div>

        <div className="hidden h-8 w-px shrink-0 bg-white/10 lg:block" />

        <nav
          className="hidden shrink-0 items-center rounded-md border border-white/10 bg-white/[0.035] p-0.5 md:flex"
          style={noDragRegion}
          aria-label="Primary"
        >
          {primaryScreens.map((screen) => {
            const Icon = Icons[screen.icon];
            const active = ui.activeScreen === screen.id;
            return (
              <button
                key={screen.id}
                onClick={() => ui.setActiveScreen(screen.id)}
                aria-current={active ? 'page' : undefined}
                className={[
                  'flex h-8 min-w-[102px] items-center justify-center gap-2 rounded px-3 text-xs font-semibold transition-colors',
                  active
                    ? 'bg-blue-500 text-white shadow-sm shadow-blue-950/40'
                    : 'text-zinc-400 hover:bg-white/10 hover:text-zinc-100',
                ].join(' ')}
              >
                <Icon size={15} />
                {screen.label}
              </button>
            );
          })}
        </nav>

        <div className="hidden min-w-0 items-center gap-2 2xl:flex">
          <Metric label="Interface" value={interfaceName} />
          <Metric label="Public" value={publicIp} />
          <Metric label="Confidence" value={confidence} />
          <Metric label="Devices" value={String(deviceCount)} />
          <Metric label="Evidence" value={String(evidenceCount)} />
        </div>

        <div className="ml-auto flex items-center gap-2" style={noDragRegion}>
          <StatusPill tone={statusTone} label={statusText} />

          <IconAction title="Toggle sidebar" active={ui.isSidebarOpen} onClick={ui.toggleSidebar}>
            <path d="M4 6h16M4 12h16M4 18h16" />
          </IconAction>
          <div className="relative">
            <IconAction title="Search devices" active={ui.isSearchOpen} onClick={toggleSearch}>
              <circle cx="11" cy="11" r="7" />
              <path d="m21 21-4.3-4.3" />
            </IconAction>
            {ui.isSearchOpen && (
              <label className="absolute right-0 top-10 z-50 flex h-9 w-64 items-center gap-2 rounded-lg border border-white/10 bg-zinc-950 px-3 text-zinc-400 shadow-2xl shadow-black/40">
                <Icons.search size={14} />
                <input
                  autoFocus
                  value={ui.searchQuery}
                  onChange={(e) => ui.setSearchQuery(e.target.value)}
                  placeholder="Search devices"
                  className="min-w-0 flex-1 bg-transparent text-xs text-zinc-100 outline-none placeholder:text-zinc-600"
                />
              </label>
            )}
          </div>
          <IconAction title="Toggle inspector" active={ui.isInspectorOpen} onClick={ui.toggleInspector}>
            <rect x="3" y="4" width="18" height="16" rx="2" />
            <path d="M14 4v16" />
          </IconAction>
          <IconAction title="Toggle logs" active={ui.isLogsOpen} onClick={ui.toggleLogs}>
            <path d="M5 6h14M5 12h14M5 18h8" />
          </IconAction>

          <button
            onClick={() => void importViaDialog()}
            disabled={busy}
            className="hidden h-8 items-center gap-2 whitespace-nowrap rounded-md border border-white/10 bg-white/[0.045] px-3 text-xs font-semibold text-zinc-200 transition-colors hover:bg-white/10 disabled:opacity-50 lg:flex"
          >
            <Icons.upload size={14} />
            Import
          </button>
          <button
            onClick={() => void runAgent('standard')}
            disabled={busy || isScanning}
            className="flex h-8 items-center gap-2 whitespace-nowrap rounded-md bg-blue-500 px-3.5 text-xs font-bold text-white shadow-sm shadow-blue-950/40 transition-colors hover:bg-blue-400 disabled:opacity-50"
          >
            <Icons.refresh size={14} />
            {isScanning ? 'Scanning...' : 'Run Scan'}
          </button>

          <div className="relative">
            <button
              onClick={() => setMenuOpen((open) => !open)}
              className="flex h-8 items-center gap-2 whitespace-nowrap rounded-md border border-white/10 bg-white/[0.045] px-3 text-xs font-semibold text-zinc-200 transition-colors hover:bg-white/10"
              aria-expanded={menuOpen}
            >
              <Icons.reports size={14} />
              Report
            </button>
            {menuOpen && (
              <div className="absolute right-0 top-full z-50 mt-2 w-56 overflow-hidden rounded-lg border border-white/10 bg-zinc-950 py-1 shadow-2xl shadow-black/45">
                {reportMenu.map((item, index) =>
                  item.separator ? (
                    <div key={index} className="my-1 h-px bg-white/10" />
                  ) : (
                    <button
                      key={item.label}
                      onClick={() => runMenuItem(item)}
                      disabled={item.disabled}
                      className="flex w-full items-center px-3 py-2 text-left text-xs font-medium text-zinc-300 transition-colors enabled:hover:bg-blue-500/10 enabled:hover:text-white disabled:text-zinc-600"
                    >
                      {item.label}
                    </button>
                  ),
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      <WindowControls />
      {aboutOpen && <AboutDialog onClose={() => setAboutOpen(false)} />}
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-[74px] max-w-[150px]">
      <div className="font-mono text-[9px] uppercase tracking-[0.12em] text-zinc-600">{label}</div>
      <div className="truncate font-mono text-[11px] font-semibold text-zinc-300">{value}</div>
    </div>
  );
}

function StatusPill({ tone, label }: { tone: string; label: string }) {
  return (
    <div className="hidden h-8 items-center gap-2 rounded-md border border-white/10 bg-white/[0.035] px-2.5 md:flex">
      <span className={`h-2 w-2 rounded-full ${tone}`} />
      <span className="font-mono text-[10px] uppercase tracking-[0.12em] text-zinc-300">{label}</span>
    </div>
  );
}

function IconAction({
  title,
  active,
  onClick,
  children,
}: {
  title: string;
  active?: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      title={title}
      aria-label={title}
      aria-pressed={active || undefined}
      onClick={onClick}
      className={[
        'flex h-8 w-8 items-center justify-center rounded-md border border-white/10 transition-colors',
        active
          ? 'bg-blue-500/18 text-blue-200'
          : 'bg-white/[0.035] text-zinc-400 hover:bg-white/10 hover:text-zinc-100',
      ].join(' ')}
    >
      <svg
        className="h-4 w-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.8}
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        {children}
      </svg>
    </button>
  );
}

function WindowControls() {
  return (
    <div className="flex h-full shrink-0 items-stretch border-l border-white/10" style={noDragRegion}>
      <WindowButton title="Minimize" onClick={() => callWindow(WindowMinimise)}>
        <path d="M6 12h12" />
      </WindowButton>
      <WindowButton title="Maximize" onClick={() => callWindow(WindowToggleMaximise)}>
        <rect x="7" y="7" width="10" height="10" rx="1.5" />
      </WindowButton>
      <WindowButton title="Close" danger onClick={() => callWindow(Quit)}>
        <path d="M8 8l8 8M16 8l-8 8" />
      </WindowButton>
    </div>
  );
}

function WindowButton({
  title,
  onClick,
  danger,
  children,
}: {
  title: string;
  onClick: () => void;
  danger?: boolean;
  children: React.ReactNode;
}) {
  return (
    <button
      title={title}
      aria-label={title}
      onClick={onClick}
      className={[
        'flex w-11 items-center justify-center text-zinc-400 transition-colors',
        danger ? 'hover:bg-red-500 hover:text-white' : 'hover:bg-white/10 hover:text-white',
      ].join(' ')}
    >
      <svg
        className="h-4 w-4"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        strokeWidth={1.8}
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        {children}
      </svg>
    </button>
  );
}

function callWindow(action: () => void) {
  try {
    action();
  } catch {
    /* Wails runtime is unavailable during browser preview. */
  }
}

function AboutDialog({ onClose }: { onClose: () => void }) {
  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60"
      style={noDragRegion}
      onClick={onClose}
    >
      <div
        className="w-[390px] rounded-xl border border-white/10 bg-zinc-950 p-6 shadow-2xl shadow-black/60"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center gap-3">
          <img src={iadMark} alt="IAD" className="h-9 w-9" />
          <div>
            <h3 className="m-0 text-base font-semibold text-zinc-100">IAD Console</h3>
            <p className="m-0 font-mono text-[10px] uppercase tracking-[0.14em] text-zinc-500">
              Internet Access Detector
            </p>
          </div>
        </div>
        <p className="text-sm leading-relaxed text-zinc-400">
          Read-only desktop console for authorized network scans. It renders iad-agent topology
          reports and keeps scan evidence immutable.
        </p>
        <div className="mt-5 flex justify-end">
          <button
            onClick={onClose}
            className="rounded-md bg-blue-500 px-4 py-1.5 text-sm font-semibold text-white hover:bg-blue-400"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
