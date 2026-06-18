/* IAD UI kit — Devices screen. Inventory with search / filters / sort,
   table + list views, click to open details. window.IAD_SCREENS.devices */
(function () {
  const { useState, useMemo } = React;
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const { Input, SegmentedControl, Badge, ConfidenceBar, StatusDot } = NS;
  const I = window.Icons;
  const fmt = window.fmt;

  const TYPES = ['all', 'default_gateway', 'access_point', 'server', 'printer', 'mobile', 'iot', 'unknown'];
  const REACH = ['all', 'reachable', 'partial', 'unreachable'];

  function ConfPill({ v }) {
    return <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, fontWeight: 600, color: fmt.bandColor(v) }}>{fmt.pct(v)}</span>;
  }

  function Devices({ scan, openDrawer, drawer }) {
    const [q, setQ] = useState('');
    const [type, setType] = useState('all');
    const [reach, setReach] = useState('all');
    const [view, setView] = useState('table');
    const [sort, setSort] = useState({ k: 'ip', dir: 1 });
    const selId = drawer && drawer.kind === 'device' ? drawer.payload.id : null;

    const rows = useMemo(() => {
      let r = scan.devices.filter((d) => {
        if (type !== 'all' && d.type !== type) return false;
        if (reach !== 'all' && d.reachability !== reach && !(reach === 'reachable' && d.reachability === 'self')) return false;
        if (q) { const s = (d.ip + d.hostname + d.vendor + d.mac + d.role).toLowerCase(); if (!s.includes(q.toLowerCase())) return false; }
        return true;
      });
      const ipNum = (ip) => ip.split('.').reduce((a, o) => a * 256 + (+o || 0), 0);
      r = [...r].sort((a, b) => {
        let av, bv;
        if (sort.k === 'ip') { av = ipNum(a.ip); bv = ipNum(b.ip); }
        else if (sort.k === 'confidence') { av = a.confidence; bv = b.confidence; }
        else { av = String(a[sort.k]); bv = String(b[sort.k]); return av.localeCompare(bv) * sort.dir; }
        return (av - bv) * sort.dir;
      });
      return r;
    }, [scan, q, type, reach, sort]);

    const toggleSort = (k) => setSort((s) => (s.k === k ? { k, dir: -s.dir } : { k, dir: 1 }));
    const Th = ({ k, children, w }) => (
      <th onClick={() => toggleSort(k)} style={{ textAlign: 'left', padding: '0 12px', height: 34, font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)', cursor: 'pointer', whiteSpace: 'nowrap', width: w, userSelect: 'none' }}>
        {children}{sort.k === k && <span style={{ marginLeft: 4, color: 'var(--accent-bright)' }}>{sort.dir > 0 ? '↑' : '↓'}</span>}
      </th>
    );

    const Select = ({ value, onChange, options, label }) => (
      <label style={{ display: 'flex', alignItems: 'center', gap: 7 }}>
        <span style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)' }}>{label}</span>
        <select value={value} onChange={(e) => onChange(e.target.value)} style={{ height: 30, background: 'var(--bg-sunken)', color: 'var(--fg-1)', border: '1px solid var(--hairline-strong)', borderRadius: 'var(--radius-sm)', padding: '0 8px', fontFamily: 'var(--font-sans)', fontSize: 'var(--text-xs)' }}>
          {options.map((o) => <option key={o} value={o}>{o === 'all' ? 'All' : o.replace(/_/g, ' ')}</option>)}
        </select>
      </label>
    );

    return (
      <div style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 14, maxWidth: 1180, margin: '0 auto' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
          <h1 style={{ font: 'var(--type-h1)' }}>Devices</h1>
          <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>{rows.length} of {scan.devices.length} on 192.168.1.0/24</span>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <div style={{ width: 260 }}><Input value={q} onChange={setQ} placeholder="Search IP, host, vendor, MAC…" iconLeft={<I.search size={15} />} size="sm" /></div>
          <Select label="Type" value={type} onChange={setType} options={TYPES} />
          <Select label="Reach" value={reach} onChange={setReach} options={REACH} />
          <div style={{ marginLeft: 'auto' }}>
            <SegmentedControl size="sm" value={view} onChange={setView} options={[{ value: 'table', label: 'Table' }, { value: 'list', label: 'List' }]} />
          </div>
        </div>

        {view === 'table' ? (
          <div style={{ border: '1px solid var(--hairline)', borderRadius: 'var(--radius-lg)', overflow: 'hidden', background: 'var(--surface-card)' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead><tr style={{ borderBottom: '1px solid var(--hairline)', background: 'var(--bg-sunken)' }}>
                <Th k="ip" w="150">IP</Th><Th k="hostname">Hostname</Th><Th k="type" w="150">Type</Th>
                <Th k="vendor">Vendor</Th><Th k="reachability" w="120">Reach</Th><Th k="confidence" w="120">Confidence</Th>
              </tr></thead>
              <tbody>
                {rows.map((d) => {
                  const r = fmt.reach(d.reachability);
                  const Icon = I[fmt.deviceIcon(d.type)];
                  const on = d.id === selId;
                  return (
                    <tr key={d.id} onClick={() => openDrawer('device', d)} style={{ borderBottom: '1px solid var(--hairline)', cursor: 'pointer', background: on ? 'var(--accent-ghost)' : 'transparent' }}>
                      <td style={{ padding: '11px 12px', fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)', color: 'var(--fg-1)' }}>{d.ip}</td>
                      <td style={{ padding: '11px 12px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
                          <span style={{ color: 'var(--fg-3)', flex: '0 0 auto' }}><Icon size={16} /></span>
                          <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)' }}>{d.hostname !== '—' ? d.hostname : <span style={{ color: 'var(--fg-4)' }}>—</span>}</span>
                        </div>
                      </td>
                      <td style={{ padding: '11px 12px', fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>{d.type}</td>
                      <td style={{ padding: '11px 12px', fontSize: 'var(--text-sm)', color: 'var(--fg-2)' }}>{d.vendor}</td>
                      <td style={{ padding: '11px 12px' }}><Badge tone={r.tone} size="sm">{r.word}</Badge></td>
                      <td style={{ padding: '11px 12px' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <div style={{ flex: 1, height: 4, background: 'var(--surface-3)', borderRadius: 'var(--radius-pill)', overflow: 'hidden' }}>
                            <div style={{ width: fmt.pct(d.confidence), height: '100%', background: fmt.bandColor(d.confidence), borderRadius: 'var(--radius-pill)' }} />
                          </div>
                          <ConfPill v={d.confidence} />
                        </div>
                      </td>
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
              const r = fmt.reach(d.reachability);
              const Icon = I[fmt.deviceIcon(d.type)];
              return (
                <div key={d.id} onClick={() => openDrawer('device', d)} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: 14, background: d.id === selId ? 'var(--accent-ghost)' : 'var(--surface-card)', border: '1px solid ' + (d.id === selId ? 'var(--accent-ring)' : 'var(--hairline)'), borderRadius: 'var(--radius-lg)', cursor: 'pointer' }}>
                  <span style={{ width: 38, height: 38, borderRadius: 'var(--radius-md)', background: 'var(--surface-3)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--fg-2)', flex: '0 0 auto' }}><Icon size={19} /></span>
                  <div style={{ minWidth: 0, flex: 1 }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 600 }}>{d.hostname !== '—' ? d.hostname : d.ip}</span>
                      <Badge tone={r.tone} size="sm">{r.word}</Badge>
                    </div>
                    <div style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-3)', marginTop: 2 }}>{d.ip} · {d.role}</div>
                  </div>
                  <ConfPill v={d.confidence} />
                </div>
              );
            })}
          </div>
        )}
      </div>
    );
  }

  window.IAD_SCREENS = window.IAD_SCREENS || {};
  window.IAD_SCREENS.devices = Devices;
})();
