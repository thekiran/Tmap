import { type CSSProperties, type ReactNode } from 'react';
import { useBreakpoint } from '../../lib/useBreakpoint';

interface DesktopLayoutProps {
  topBar: ReactNode;
  sidebar: ReactNode;
  mainCanvas: ReactNode;
  inspector?: ReactNode;
  bottomLogs?: ReactNode;
  isSidebarOpen?: boolean;
  isInspectorOpen?: boolean;
  isLogsOpen?: boolean;
}

/**
 * Responsive desktop app shell.
 *
 * Layout is a CSS grid driven by design tokens (see globals.css): the toolbar
 * and status bar span the full width, the sidebar and details panel use clamped
 * widths that stay compact on 1366×768 and don't over-stretch on 4K/ultrawide,
 * and the canvas is a `minmax(0, 1fr)` column so it always fills the remaining
 * space and never overflows.
 *
 * The user's open/closed intent (from the UI store) is combined with the
 * current window width so panels auto-collapse on small windows:
 *   - sidebar hides below 1024px,
 *   - the bottom status/log row collapses below 1024px,
 *   - the details panel OVERLAYS the canvas below 1280px (instead of taking its
 *     own column) so the map stays full-width and the panel never permanently
 *     covers content — it's only shown when the user opened it and is closable.
 */
export function DesktopLayout({
  topBar,
  sidebar,
  mainCanvas,
  inspector,
  bottomLogs,
  isSidebarOpen = true,
  isInspectorOpen = true,
  isLogsOpen = true,
}: DesktopLayoutProps) {
  const { isCompact, isNarrow } = useBreakpoint();

  const showSidebar = isSidebarOpen && Boolean(sidebar) && !isCompact;
  const inspectorOpen = isInspectorOpen && Boolean(inspector);
  const inspectorColumn = inspectorOpen && !isNarrow;
  const inspectorOverlay = inspectorOpen && isNarrow;
  const showLogs = isLogsOpen && Boolean(bottomLogs) && !isCompact;

  const gridStyle: CSSProperties = {
    gridTemplateColumns: [
      showSidebar ? 'var(--sidebar-w)' : '0px',
      'minmax(0, 1fr)',
      inspectorColumn ? 'var(--details-w)' : '0px',
    ].join(' '),
    gridTemplateRows: showLogs
      ? 'var(--toolbar-h) minmax(0, 1fr) var(--statusbar-h)'
      : 'var(--toolbar-h) minmax(0, 1fr)',
  };

  return (
    <div
      className="grid h-screen w-screen overflow-hidden bg-zinc-50 font-sans text-zinc-900 selection:bg-blue-500/30 dark:bg-zinc-950 dark:text-zinc-100"
      style={gridStyle}
    >
      {/* Toolbar — spans all columns. */}
      <header
        className="z-20 flex items-center border-b border-zinc-200 bg-white dark:border-white/10 dark:bg-[#0e0e10]"
        style={{ gridColumn: '1 / -1', gridRow: 1 }}
      >
        {topBar}
      </header>

      {/* Sidebar. */}
      {showSidebar && (
        <aside
          className="z-10 flex min-w-0 flex-col overflow-hidden border-r border-zinc-200 bg-zinc-50/50 dark:border-zinc-800/50 dark:bg-zinc-900/20"
          style={{ gridColumn: 1, gridRow: 2 }}
        >
          {sidebar}
        </aside>
      )}

      {/* Main canvas — min-w-0/min-h-0 keep React Flow from overflowing the grid
          cell; the overlay details panel (narrow widths) lives inside here. */}
      <main
        className="relative min-h-0 min-w-0 overflow-hidden bg-white dark:bg-black/20"
        style={{ gridColumn: 2, gridRow: 2 }}
      >
        {mainCanvas}

        {inspectorOverlay && inspector && (
          <aside
            className="absolute right-0 top-0 bottom-0 z-30 flex min-w-0 flex-col overflow-hidden border-l border-zinc-200 bg-white shadow-2xl shadow-black/30 dark:border-zinc-800/50 dark:bg-zinc-900"
            style={{ width: 'var(--details-w)', maxWidth: '92%' }}
          >
            {inspector}
          </aside>
        )}
      </main>

      {/* Details panel — its own column on wide screens. */}
      {inspectorColumn && inspector && (
        <aside
          className="z-10 flex min-w-0 flex-col overflow-hidden border-l border-zinc-200 bg-zinc-50/50 shadow-xl shadow-black/5 dark:border-zinc-800/50 dark:bg-zinc-900/20"
          style={{ gridColumn: 3, gridRow: 2 }}
        >
          {inspector}
        </aside>
      )}

      {/* Status / logs row — spans all columns. */}
      {showLogs && (
        <footer
          className="z-20 min-h-0 overflow-hidden border-t border-zinc-200 bg-white dark:border-zinc-800/50 dark:bg-zinc-900/80"
          style={{ gridColumn: '1 / -1', gridRow: 3 }}
        >
          {bottomLogs}
        </footer>
      )}
    </div>
  );
}
