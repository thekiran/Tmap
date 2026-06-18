import { useState } from 'react';
import { useScanStore } from '../../store/useScanStore';
import { Badge, Overline } from '../ui/primitives';
import { ConfidenceBar, ProbeStatusBadge, TierBadge, Input } from '../ui/data';
import { Icons } from '../icons/Icon';
import type { ProbeStatus } from '../../lib/scan-schema';

const STATUSES: (ProbeStatus | 'all')[] = ['all', 'success', 'partial', 'no_data', 'skipped', 'failed', 'blocked'];

export function EvidenceScreen() {
  const scan = useScanStore((s) => s.normalized)!;
  const [status, setStatus] = useState<ProbeStatus | 'all'>('all');
  const [q, setQ] = useState('');
  const [open, setOpen] = useState<string | null>(scan.evidence[0]?.id ?? null);

  const rows = scan.evidence.filter((e) => {
    if (status !== 'all' && e.status !== status) return false;
    if (q && !(e.probeName + (e.reason ?? '')).toLowerCase().includes(q.toLowerCase())) return false;
    return true;
  });

  return (
    <div className="app-scroll" style={{ flex: 1, overflow: 'auto' }}>
      <div style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 14, maxWidth: 1000, margin: '0 auto' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
          <h1 style={{ font: 'var(--type-h1)' }}>Evidence</h1>
          <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>{scan.evidence.length} probes · {scan.evidence.filter((e) => e.status === 'success').length} success</span>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
          <div style={{ width: 240 }}><Input value={q} onChange={setQ} placeholder="Search probes…" iconLeft={<Icons.search size={15} />} size="sm" /></div>
          <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
            {STATUSES.map((s) => {
              const on = status === s;
              const n = s === 'all' ? scan.evidence.length : scan.evidence.filter((e) => e.status === s).length;
              return (
                <button key={s} onClick={() => setStatus(s)} style={{ display: 'inline-flex', alignItems: 'center', gap: 6, height: 28, padding: '0 10px', borderRadius: 'var(--radius-sm)', cursor: 'pointer', border: '1px solid ' + (on ? 'var(--hairline-strong)' : 'transparent'), background: on ? 'var(--surface-2)' : 'transparent', color: on ? 'var(--fg-1)' : 'var(--fg-3)', fontFamily: 'var(--font-mono)', fontSize: 11, textTransform: 'uppercase', letterSpacing: '.06em' }}>
                  {s}<span style={{ color: 'var(--fg-4)' }}>{n}</span>
                </button>
              );
            })}
          </div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {rows.map((e) => {
            const isOpen = open === e.id;
            return (
              <div key={e.id} style={{ border: '1px solid var(--hairline)', borderRadius: 'var(--radius-lg)', background: 'var(--surface-card)', overflow: 'hidden' }}>
                <div onClick={() => setOpen(isOpen ? null : e.id)} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '13px 16px', cursor: 'pointer' }}>
                  <span style={{ color: 'var(--fg-4)', transform: isOpen ? 'rotate(90deg)' : 'none', transition: 'transform var(--dur-fast) var(--ease-out)', display: 'inline-flex' }}><Icons.chevronRight size={15} /></span>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 600, minWidth: 150 }}>{e.probeName}</span>
                  <ProbeStatusBadge status={e.status} size="sm" />
                  <TierBadge tier={e.evidenceClass} appearance="dot" />
                  <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-3)', flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{e.reason}</span>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-4)' }}>{e.timestamp}</span>
                </div>
                {isOpen && (
                  <div style={{ padding: '0 16px 16px 42px', display: 'flex', flexDirection: 'column', gap: 14 }}>
                    {e.emptyEvidenceWarning && (
                      <div style={{ display: 'flex', gap: 10, padding: 11, background: 'var(--warn-bg)', borderRadius: 'var(--radius-md)' }}>
                        <span style={{ color: 'var(--warn)', flex: '0 0 auto' }}><Icons.alert size={15} /></span>
                        <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-2)', lineHeight: 1.45 }}>Probe completed but returned no useful evidence. Consider normalizing this as <b style={{ fontFamily: 'var(--font-mono)' }}>no_data</b>.</span>
                      </div>
                    )}
                    {e.confidence > 0 && <div style={{ maxWidth: 300 }}><ConfidenceBar value={e.confidence} label="Probe confidence" size="sm" /></div>}
                    {e.limitations && (
                      <div><Overline style={{ marginBottom: 5 }}>Limitations</Overline><p style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.5 }}>{e.limitations}</p></div>
                    )}
                    {(e.warnings.length > 0 || e.errors.length > 0) && (
                      <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
                        {e.errors.map((m, i) => <div key={'e' + i} style={{ fontSize: 'var(--text-xs)', color: 'var(--danger)' }}>• {m}</div>)}
                        {e.warnings.map((m, i) => <div key={'w' + i} style={{ fontSize: 'var(--text-xs)', color: 'var(--warn)' }}>• {m}</div>)}
                      </div>
                    )}
                    <div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                        <Overline>Raw evidence</Overline>
                        <Badge tone="neutral" appearance="outline" mono size="sm">{e.evidenceClass}</Badge>
                      </div>
                      <pre style={{ margin: 0, padding: 12, background: 'var(--ink-900)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-md)', fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-2)', lineHeight: 1.55, overflow: 'auto', whiteSpace: 'pre-wrap' }}>{JSON.stringify(e.data, null, 2)}</pre>
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
