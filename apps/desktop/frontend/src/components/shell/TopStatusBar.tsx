import type { ReactNode } from 'react';
import { useScanStore } from '../../store/useScanStore';
import { useUIStore } from '../../store/useUIStore';
import { Badge, Button } from '../ui/primitives';
import { Icons } from '../icons/Icon';
import { timeShort } from '../../lib/format';
import { bandColorVar, bandWord } from '../../lib/confidence';
import { useImport } from '../../lib/useImport';

function Metric({ label, value, mono = true, minW = 88, maxW }: { label: string; value: ReactNode; mono?: boolean; minW?: number; maxW?: number }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 1, minWidth: minW, maxWidth: maxW, flex: '0 0 auto' }}>
      <span style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)', fontSize: 9 }}>{label}</span>
      <span style={{ fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)', fontSize: 12, fontWeight: 600, color: 'var(--fg-1)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{value}</span>
    </div>
  );
}

const Divider = () => <div style={{ width: 1, height: 26, background: 'var(--hairline)', flex: '0 0 auto' }} />;

export function TopStatusBar() {
  const scan = useScanStore((s) => s.normalized);
  const setScreen = useUIStore((s) => s.setScreen);
  const { importViaDialog } = useImport();

  if (!scan) {
    return (
      <header style={{ height: 'var(--topbar-h)', flex: '0 0 auto', background: 'var(--ink-850)', borderBottom: '1px solid var(--hairline)', display: 'flex', alignItems: 'center', gap: 16, padding: '0 16px' }}>
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, color: 'var(--fg-3)' }}>No scan loaded</span>
        <div style={{ marginLeft: 'auto' }}>
          <Button size="sm" variant="primary" iconLeft={<Icons.upload size={14} />} onClick={importViaDialog}>Import scan</Button>
        </div>
      </header>
    );
  }

  const c = scan.confidence;
  const iface = scan.selectedInterface?.name ?? '—';
  const pubip = scan.publicIp?.address ?? '—';
  const isp = scan.publicIp ? `${scan.publicIp.asn ?? ''} · ${scan.publicIp.org ?? ''}` : '—';

  return (
    <header style={{ height: 'var(--topbar-h)', flex: '0 0 auto', background: 'var(--ink-850)', borderBottom: '1px solid var(--hairline)', display: 'flex', alignItems: 'center', gap: 16, padding: '0 16px', overflow: 'hidden' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: 9, color: 'var(--fg-4)', textTransform: 'uppercase', letterSpacing: '.1em' }}>Scan</span>
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, color: 'var(--fg-1)', fontWeight: 600 }}>{scan.scanId}</span>
        </div>
        <Badge tone="neutral" appearance="outline" size="sm" mono>{scan.mode}</Badge>
      </div>
      <Divider />
      <Metric label="Time" value={timeShort(scan.createdAt)} mono={false} minW={96} />
      <Metric label="Interface" value={iface} minW={80} />
      <Metric label="Public IP" value={pubip} minW={110} />
      <Metric label="ISP" value={isp} mono={false} minW={120} maxW={190} />
      <div style={{ display: 'flex', alignItems: 'center', gap: 7, marginLeft: 'auto' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 1, alignItems: 'flex-end' }}>
          <span style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)', fontSize: 9 }}>Confidence</span>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 700, color: bandColorVar(c) }}>{Math.round(c * 100)}%</span>
            <span style={{ fontSize: 9, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '.08em', color: bandColorVar(c) }}>{bandWord(c)}</span>
          </div>
        </div>
        <Badge tone={scan.decisionQuality === 'high' ? 'success' : scan.decisionQuality === 'medium' ? 'warn' : 'neutral'} size="sm" uppercase mono>{scan.decisionQuality} quality</Badge>
      </div>
      <Divider />
      <div style={{ display: 'flex', gap: 8, flex: '0 0 auto' }}>
        <Button size="sm" variant="secondary" iconLeft={<Icons.upload size={14} />} onClick={importViaDialog}>Import</Button>
        <Button size="sm" variant="primary" iconLeft={<Icons.download size={14} />} onClick={() => setScreen('reports')}>Export</Button>
      </div>
    </header>
  );
}
