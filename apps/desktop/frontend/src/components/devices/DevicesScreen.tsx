import { useMemo, useState } from 'react';
import { useScanStore } from '../../store/useScanStore';
import { useUIStore } from '../../store/useUIStore';
import { Badge } from '../ui/primitives';
import { Input, SegmentedControl } from '../ui/data';
import { Icon, Icons } from '../icons/Icon';
import { REACH_META, deviceIconKey, ipKey } from '../../lib/format';
import { bandColorVar } from '../../lib/confidence';
import type { NetworkDevice } from '../../lib/models';

type SortKey = 'ip' | 'hostname' | 'confidence' | 'type';

export function DevicesScreen() {
  const scan = useScanStore((s) => s.normalized)!;
  const selectDevice = useUIStore((s) => s.setSelectedDeviceId);
  const selId = useUIStore((s) => s.selectedDeviceId);
  const globalQuery = useUIStore((s) => s.searchQuery);

  const [q, setQ] = useState('');
  const [type, setType] = useState('all');
  const [reach, setReach] = useState('all');
  const [source, setSource] = useState('all');
  const [view, setView] = useState<'table' | 'list'>('table');
  const [sort, setSort] = useState<{ k: SortKey; dir: 1 | -1 }>({ k: 'ip', dir: 1 });

  const types = useMemo(() => ['all', ...Array.from(new Set(scan.devices.map((d) => d.type)))], [scan]);
  const sources = useMemo(() => ['all', ...Array.from(new Set(scan.devices.map((d) => d.source).filter(Boolean) as string[]))], [scan]);

  const rows = useMemo(() => {
    const term = (q || globalQuery).toLowerCase();
    let r = scan.devices.filter((d) => {
      if (type !== 'all' && d.type !== type) return false;
      if (reach !== 'all' && d.reachability !== reach && !(reach === 'reachable' && d.reachability === 'self')) return false;
      if (source !== 'all' && d.source !== source) return false;
      if (term) {
        const s = `${d.ip}${d.hostname ?? ''}${d.vendor ?? ''}${d.mac ?? ''}${d.role ?? ''}`.toLowerCase();
        if (!s.includes(term)) return false;
      }
      return true;
    });
    r = [...r].sort((a, b) => {
      if (sort.k === 'ip') return (ipKey(a.ip) - ipKey(b.ip)) * sort.dir;
      if (sort.k === 'confidence') return (a.confidence - b.confidence) * sort.dir;
      return String(a[sort.k] ?? '').localeCompare(String(b[sort.k] ?? '')) * sort.dir;
    });
    return r;
  }, [scan, q, globalQuery, type, reach, source, sort]);

  const toggleSort = (k: SortKey) => setSort((s) => (s.k === k ? { k, dir: (s.dir * -1) as 1 | -1 } : { k, dir: 1 }));

  const Th = ({ k, children, w }: { k: SortKey; children: React.ReactNode; w?: number }) => (
    <th onClick={() => toggleSort(k)} style={{ textAlign: 'left', padding: '0 12px', height: 34, font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)', cursor: 'pointer', whiteSpace: 'nowrap', width: w, userSelect: 'none' }}>
      {children}{sort.k === k && <span style={{ marginLeft: 4, color: 'var(--accent-bright)' }}>{sort.dir > 0 ? '↑' : '↓'}</span>}
    </th>
  );

  const Filter = ({ label, value, onChange, options }: { label: string; value: string; onChange: (v: string) => void; options: string[] }) => (
    <label style={{ display: 'flex', alignItems: 'center', gap: 7 }}>
      <span style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)' }}>{label}</span>
      <select value={value} onChange={(e) => onChange(e.target.value)} style={{ height: 30, background: 'var(--bg-sunken)', color: 'var(--fg-1)', border: '1px solid var(--hairline-strong)', borderRadius: 'var(--radius-sm)', padding: '0 8px', fontFamily: 'var(--font-sans)', fontSize: 'var(--text-xs)' }}>
        {options.map((o) => <option key={o} value={o}>{o === 'all' ? 'All' : o.replace(/_/g, ' ')}</option>)}
      </select>
    </label>
  );

  const Conf = ({ d }: { d: NetworkDevice }) => (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <div style={{ flex: 1, height: 4, background: 'var(--surface-3)', borderRadius: 'var(--radius-pill)', overflow: 'hidden' }}>
        <div style={{ width: Math.round(d.confidence * 100) + '%', height: '100%', background: bandColorVar(d.confidence), borderRadius: 'var(--radius-pill)' }} />
      </div>
      <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, fontWeight: 600, color: bandColorVar(d.confidence) }}>{Math.round(d.confidence * 100)}%</span>
    </div>
  );

  return (
    <div className="app-scroll" style={{ flex: 1, overflow: 'auto' }}>
      <div style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 14, maxWidth: 1180, margin: '0 auto' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
          <h1 style={{ font: 'var(--type-h1)' }}>Devices</h1>
          <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>{rows.length} of {scan.devices.length}</span>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <div style={{ width: 240 }}><Input value={q} onChange={setQ} placeholder="Search IP, host, vendor, MAC…" iconLeft={<Icons.search size={15} />} size="sm" /></div>
          <Filter label="Type" value={type} onChange={setType} options={types} />
          <Filter label="Reach" value={reach} onChange={setReach} options={['all', 'reachable', 'partial', 'unreachable']} />
          <Filter label="Source" value={source} onChange={setSource} options={sources} />
          <div style={{ marginLeft: 'auto' }}>
            <SegmentedControl size="sm" value={view} onChange={setView} options={[{ value: 'table', label: 'Table' }, { value: 'list', label: 'List' }]} />
          </div>
        </div>

        {view === 'table' ? (
          <div style={{ border: '1px solid var(--hairline)', borderRadius: 'var(--radius-lg)', overflow: 'hidden', background: 'var(--surface-card)' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead><tr style={{ borderBottom: '1px solid var(--hairline)', background: 'var(--bg-sunken)' }}>
                <Th k="ip" w={150}>IP</Th><Th k="hostname">Hostname</Th><Th k="type" w={150}>Type</Th>
                <th style={{ textAlign: 'left', padding: '0 12px', font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)' }}>Vendor</th>
                <th style={{ textAlign: 'left', padding: '0 12px', width: 120, font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)' }}>Reach</th>
                <Th k="confidence" w={130}>Confidence</Th>
              </tr></thead>
              <tbody>
                {rows.map((d) => {
                  const r = REACH_META[d.reachability];
                  return (
                    <tr key={d.id} onClick={() => selectDevice(d.id)} style={{ borderBottom: '1px solid var(--hairline)', cursor: 'pointer', background: d.id === selId ? 'var(--accent-ghost)' : 'transparent' }}>
                      <td style={{ padding: '11px 12px', fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)', color: 'var(--fg-1)' }}>{d.ip}</td>
                      <td style={{ padding: '11px 12px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
                          <span style={{ color: 'var(--fg-3)', flex: '0 0 auto' }}><Icon name={deviceIconKey(d.type) as never} size={16} /></span>
                          <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)' }}>{d.hostname ?? <span style={{ color: 'var(--fg-4)' }}>—</span>}</span>
                        </div>
                      </td>
                      <td style={{ padding: '11px 12px', fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>{d.type}</td>
                      <td style={{ padding: '11px 12px', fontSize: 'var(--text-sm)', color: 'var(--fg-2)' }}>{d.vendor ?? '—'}</td>
                      <td style={{ padding: '11px 12px' }}><Badge tone={r.tone as never} size="sm">{r.word}</Badge></td>
                      <td style={{ padding: '11px 12px' }}><Conf d={d} /></td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {rows.length === 0 && <div style={{ padding: 40, textAlign: 'center', color: 'var(--fg-4)', fontSize: 'var(--text-sm)' }}>No devices match these filters.</div>}
          </div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 12 }}>
            {rows.map((d) => {
              const r = REACH_META[d.reachability];
              return (
                <div key={d.id} onClick={() => selectDevice(d.id)} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: 14, background: d.id === selId ? 'var(--accent-ghost)' : 'var(--surface-card)', border: '1px solid ' + (d.id === selId ? 'var(--accent-ring)' : 'var(--hairline)'), borderRadius: 'var(--radius-lg)', cursor: 'pointer' }}>
                  <span style={{ width: 38, height: 38, borderRadius: 'var(--radius-md)', background: 'var(--surface-3)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--fg-2)', flex: '0 0 auto' }}><Icon name={deviceIconKey(d.type) as never} size={19} /></span>
                  <div style={{ minWidth: 0, flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 600 }}>{d.hostname ?? d.ip}</span>
                      <Badge tone={r.tone as never} size="sm">{r.word}</Badge>
                    </div>
                    <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-3)', marginTop: 2 }}>{d.ip} · {d.role ?? d.type}</div>
                  </div>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, fontWeight: 600, color: bandColorVar(d.confidence) }}>{Math.round(d.confidence * 100)}%</span>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
