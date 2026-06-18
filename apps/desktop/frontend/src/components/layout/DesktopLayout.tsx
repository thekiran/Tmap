import { type ReactNode } from 'react';

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
  return (
    <div className="flex flex-col h-screen w-screen overflow-hidden bg-zinc-50 dark:bg-zinc-950 text-zinc-900 dark:text-zinc-100 font-sans selection:bg-blue-500/30">
      {/* Custom application chrome */}
      <header className="h-14 shrink-0 border-b border-zinc-200 dark:border-white/10 bg-white dark:bg-[#0e0e10] flex items-center z-20">
        {topBar}
      </header>

      {/* Main Workspace Area */}
      <div className="flex flex-1 overflow-hidden relative">
        
        {/* Left Sidebar - Fixed width, collapsible */}
        {isSidebarOpen && (
          <aside className="w-64 shrink-0 border-r border-zinc-200 dark:border-zinc-800/50 bg-zinc-50/50 dark:bg-zinc-900/20 flex flex-col z-10">
            {sidebar}
          </aside>
        )}

        {/* Center Canvas & Right Inspector */}
        <div className="flex flex-1 overflow-hidden relative">
          
          {/* Main Canvas (React Flow) */}
          <main className="flex-1 relative bg-white dark:bg-black/20">
             {mainCanvas}
          </main>

          {/* Right Inspector - Fixed width when open */}
          {isInspectorOpen && inspector && (
            <aside className="w-80 shrink-0 border-l border-zinc-200 dark:border-zinc-800/50 bg-zinc-50/50 dark:bg-zinc-900/20 flex flex-col z-10 shadow-xl shadow-black/5">
              {inspector}
            </aside>
          )}
        </div>
      </div>

      {/* Bottom Logs - Fixed height when open */}
      {isLogsOpen && bottomLogs && (
        <footer className="h-48 shrink-0 border-t border-zinc-200 dark:border-zinc-800/50 bg-white dark:bg-zinc-900/80 z-20">
          {bottomLogs}
        </footer>
      )}
    </div>
  );
}
