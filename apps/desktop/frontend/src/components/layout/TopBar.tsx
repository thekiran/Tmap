import { useImport } from '../../lib/useImport';
import { useScanStore } from '../../store/useScanStore';

export function TopBar() {
  const { importViaDialog, runAgent, busy } = useImport();
  const isScanning = useScanStore((state) => state.isScanning);
  const dragRegion = {
    WebkitAppRegion: 'drag',
    '--wails-draggable': 'drag',
  } as React.CSSProperties;
  const noDragRegion = {
    WebkitAppRegion: 'no-drag',
    '--wails-draggable': 'no-drag',
  } as React.CSSProperties;
  
  return (
    <div className="flex-1 flex items-center justify-between px-4" style={dragRegion}>
      <div className="flex items-center gap-3">
        <div className="w-6 h-6 bg-blue-500 rounded flex items-center justify-center text-white font-bold text-xs shadow-sm shadow-blue-500/20">
          I
        </div>
        <h1 className="font-semibold text-sm tracking-tight text-zinc-800 dark:text-zinc-200">
          Internet Access Detector
        </h1>
        <div className="h-4 w-px bg-zinc-300 dark:bg-zinc-700 mx-2" />
        <span className="text-xs text-zinc-500 font-medium">
          {isScanning || busy ? 'Busy...' : 'Idle'}
        </span>
      </div>

      <div className="flex items-center gap-2" style={noDragRegion}>
        <button 
          onClick={importViaDialog}
          disabled={busy || isScanning}
          className="px-3 py-1.5 text-xs font-medium rounded-md hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-600 dark:text-zinc-300 transition-colors disabled:opacity-50"
        >
          Import Report
        </button>
        <button 
          disabled={busy || isScanning}
          className="px-3 py-1.5 text-xs font-medium rounded-md hover:bg-zinc-100 dark:hover:bg-zinc-800 text-zinc-600 dark:text-zinc-300 transition-colors disabled:opacity-50"
        >
          Export
        </button>
        <button 
          onClick={() => runAgent()}
          disabled={busy || isScanning}
          className="px-4 py-1.5 text-xs font-semibold rounded-md bg-blue-500 hover:bg-blue-600 text-white shadow-sm shadow-blue-500/20 transition-colors ml-2 disabled:opacity-50"
        >
          Run Scan
        </button>
      </div>
    </div>
  );
}
