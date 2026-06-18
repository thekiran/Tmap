import { useScanStore } from '../../store/useScanStore';

export function BottomLogs() {
  const { logs, clearLogs } = useScanStore();

  return (
    <div className="flex flex-col h-full">
      {/* Log Header */}
      <div className="flex items-center justify-between px-4 py-1.5 border-b border-zinc-200 dark:border-zinc-800/50 bg-zinc-50 dark:bg-zinc-900/50">
        <h3 className="text-xs font-semibold text-zinc-600 dark:text-zinc-400">Agent Logs</h3>
        <div className="flex gap-2">
          <button onClick={clearLogs} className="text-[10px] text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300">Clear</button>
        </div>
      </div>
      
      {/* Log Content Area */}
      <div className="flex-1 overflow-y-auto p-4 bg-white dark:bg-zinc-950 font-mono text-[11px] leading-relaxed">
        {logs.length === 0 ? (
          <div className="text-zinc-500">[System] Ready. Waiting for scan initiation...</div>
        ) : (
          logs.map((log, i) => (
            <div key={i} className="text-zinc-600 dark:text-zinc-400">
              {log}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
