import { useScanStore } from '../../store/useScanStore';
import { useUIStore } from '../../store/useUIStore';
import { StatusDot } from '../ui/primitives';

const LINES: Record<string, string> = {
  dashboard: 'scan complete · derived view ready',
  topology: 'topology generated from context · inferred links shown dashed',
  devices: 'device inventory · local broadcast domain only',
  evidence: 'probe evidence · raw JSON read-only',
  reports: 'ready · export NormalizedScanReport',
  settings: 'UI layout positions stored locally · scan data immutable',
};

export function LogStrip() {
  const screen = useUIStore((s) => s.screen);
  const scan = useScanStore((s) => s.normalized);
  const source = useScanStore((s) => s.source);
  const text = scan
    ? `${scan.scanId} · ${source} · ${LINES[screen] ?? ''}`
    : 'no scan loaded · import JSON or run the agent';
  return (
    <footer style={{ height: 28, flex: '0 0 auto', background: 'var(--ink-900)', borderTop: '1px solid var(--hairline)', display: 'flex', alignItems: 'center', gap: 12, padding: '0 16px', overflow: 'hidden' }}>
      <StatusDot tone={scan ? 'success' : 'neutral'} size={7} />
      <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-3)', whiteSpace: 'nowrap' }}>{text}</span>
    </footer>
  );
}
