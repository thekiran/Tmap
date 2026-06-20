import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useScanStore, type LogEntry, type LogLevel } from '../../store/useScanStore';

const LEVEL_STYLE: Record<LogLevel, { text: string; tag: string; label: string }> = {
  debug: { text: 'text-zinc-400', tag: 'text-zinc-500', label: 'DBG' },
  info: { text: 'text-sky-300', tag: 'text-sky-400/80', label: 'INF' },
  success: { text: 'text-emerald-300', tag: 'text-emerald-400/80', label: 'OK ' },
  warn: { text: 'text-amber-300', tag: 'text-amber-400/80', label: 'WRN' },
  error: { text: 'text-red-300', tag: 'text-red-400/90', label: 'ERR' },
};

function ts(ms: number): string {
  const d = new Date(ms);
  const p = (n: number) => String(n).padStart(2, '0');
  return `${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`;
}

export function BottomLogs() {
  const logs = useScanStore((s) => s.logs);
  const clearLogs = useScanStore((s) => s.clearLogs);

  const scrollRef = useRef<HTMLDivElement>(null);
  // Auto-scroll to the newest line, but only while the user is already pinned to
  // the bottom — scrolling up to read history pauses the follow.
  const [stick, setStick] = useState(true);
  const [copied, setCopied] = useState(false);

  useLayoutEffect(() => {
    if (stick && scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs, stick]);

  useEffect(() => {
    if (!copied) return;
    const t = window.setTimeout(() => setCopied(false), 1200);
    return () => window.clearTimeout(t);
  }, [copied]);

  const onScroll = () => {
    const el = scrollRef.current;
    if (!el) return;
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 24;
    setStick(atBottom);
  };

  const copyAll = async () => {
    const text = logs.map((l) => `${ts(l.ts)} ${LEVEL_STYLE[l.level].label} ${l.source} · ${l.message}`).join('\n');
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
    } catch {
      // Clipboard may be unavailable; ignore.
    }
  };

  return (
    <div className="flex h-full min-h-0 flex-col bg-white dark:bg-zinc-950">
      {/* Header */}
      <div className="flex shrink-0 items-center gap-2 border-b border-zinc-200 bg-zinc-50 px-4 py-1.5 dark:border-zinc-800/60 dark:bg-zinc-900/60">
        <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-zinc-500">Terminal</span>
        <span className="rounded bg-zinc-200/60 px-1.5 font-mono text-[10px] text-zinc-500 dark:bg-zinc-800/80">{logs.length}</span>
        {!stick && (
          <button
            onClick={() => setStick(true)}
            className="rounded px-1.5 font-mono text-[10px] text-amber-500 hover:bg-amber-500/10"
            title="Jump to latest"
          >
            ▼ follow
          </button>
        )}
        <div className="ml-auto flex items-center gap-2">
          <button
            onClick={copyAll}
            disabled={logs.length === 0}
            className="font-mono text-[10px] text-zinc-500 hover:text-zinc-700 disabled:opacity-40 dark:hover:text-zinc-300"
          >
            {copied ? 'copied' : 'copy'}
          </button>
          <button
            onClick={clearLogs}
            disabled={logs.length === 0}
            className="font-mono text-[10px] text-zinc-500 hover:text-zinc-700 disabled:opacity-40 dark:hover:text-zinc-300"
          >
            clear
          </button>
        </div>
      </div>

      {/* Log body */}
      <div
        ref={scrollRef}
        onScroll={onScroll}
        className="app-scroll min-h-0 flex-1 overflow-y-auto px-3 py-2 font-mono text-[11px] leading-[1.55]"
      >
        {logs.length === 0 ? (
          <div className="select-none text-zinc-500">
            <span className="text-emerald-400">$</span> waiting for scan — press <span className="text-zinc-300">Rescan</span> to begin…
          </div>
        ) : (
          logs.map((l) => <LogRow key={l.id} entry={l} />)
        )}
      </div>
    </div>
  );
}

function LogRow({ entry }: { entry: LogEntry }) {
  const s = LEVEL_STYLE[entry.level];
  return (
    <div className="flex gap-2 whitespace-pre-wrap break-words">
      <span className="shrink-0 tabular-nums text-zinc-600">{ts(entry.ts)}</span>
      <span className={`shrink-0 ${s.tag}`}>{s.label}</span>
      <span className="shrink-0 text-zinc-500">{entry.source}</span>
      <span className="text-zinc-600">·</span>
      <span className={s.text}>{entry.message}</span>
    </div>
  );
}
