import iadMark from '../../assets/mark-iad.svg';
import { Quit, WindowMinimise, WindowToggleMaximise } from '../../../wailsjs/runtime/runtime';

const dragRegion = {
  WebkitAppRegion: 'drag',
  '--wails-draggable': 'drag',
} as React.CSSProperties;
const noDragRegion = {
  WebkitAppRegion: 'no-drag',
  '--wails-draggable': 'no-drag',
} as React.CSSProperties;

/**
 * Bare application chrome for the interface-selection screen: brand + window
 * controls only, no navigation tabs, toolbar, sidebar, inspector or logs.
 * Once a scan/report is loaded the app swaps to the full DesktopLayout.
 */
export function LaunchChrome({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex h-screen w-screen flex-col overflow-hidden bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100">
      <header
        className="flex h-11 shrink-0 items-center bg-[#0e0e10] text-zinc-100"
        style={dragRegion}
      >
        <div className="flex items-center gap-2.5 px-3">
          <img src={iadMark} alt="IAD" className="h-7 w-7 shrink-0" />
          <span className="font-mono text-[14px] font-bold tracking-[0.14em] text-white">IAD</span>
          <span className="hidden whitespace-nowrap text-[12px] font-semibold text-zinc-400 sm:inline">
            Internet Access Detector
          </span>
        </div>
        <div className="ml-auto flex h-full items-stretch" style={noDragRegion}>
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
      </header>
      <main className="relative flex-1 overflow-hidden bg-white dark:bg-black/20">{children}</main>
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
