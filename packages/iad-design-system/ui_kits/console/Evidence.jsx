/* IAD UI kit — Evidence screen. Probe explorer with status, tier, raw JSON.
   Flags the "success but empty evidence" anti-pattern. window.IAD_SCREENS.evidence */
(function () {
  const { useState } = React;
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const { ProbeStatusBadge, TierBadge, Badge, ConfidenceBar, Input } = NS;
  const I = window.Icons;

  const STATUSES = ['all', 'success', 'partial', 'no_data', 'skipped', 'failed', 'blocked'];

  function isEmptyEvidence(e) {
    if (!e.data) return true;
    const vals = Object.values(e.data);
    return vals.length === 0 || vals.every((v) => v == null || v === '' || v === 0 || v === false);
  }

  function Evidence({ scan, openDrawer, drawer }) {
    const [status, setStatus] = useState('all');
    const [q, setQ] = useState('');
    const [open, setOpen] = useState(scan.evidence[0].id);
    const selId = drawer && drawer.kind === 'probe' ? drawer.payload.id : null;

    const rows = scan.evidence.filter((e) => {
      if (status !== 'all' && e.status !== status) return false;
      if (q && !(e.probe_name + e.reason).toLowerCase().includes(q.toLowerCase())) return false;
      return true;
    });

    const counts = STATUSES.slice(1).map((s) => ({ s, n: scan.evidence.filter((e) => e.status === s).length }));

    return (
      <div style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 14, maxWidth: 1000, margin: '0 auto' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
          <h1 style={{ font: 'var(--type-h1)' }}>Evidence</h1>
          <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>{scan.evidence.length} probes · {scan.evidence.filter((e) => e.status === 'success').length} success</span>
        </div>

        {/* Status filter chips */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
          <div style={{ width: 240 }}><Input value={q} onChange={setQ} placeholder="Search probes…" iconLeft={<I.search size={15} />} size="sm" /></div>
          <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
            {STATUSES.map((s) => {
              const on = status === s;
              const n = s === 'all' ? scan.evidence.length : (counts.find((c) => c.s === s) || {}).n;
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
            const emptyWarn = e.status === 'success' && isEmptyEvidence(e);
            return (
              <div key={e.id} style={{ border: '1px solid ' + (e.id === selId ? 'var(--accent-ring)' : 'var(--hairline)'), borderRadius: 'var(--radius-lg)', background: 'var(--surface-card)', overflow: 'hidden' }}>
                <div onClick={() => setOpen(isOpen ? null : e.id)} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '13px 16px', cursor: 'pointer' }}>
                  <span style={{ color: 'var(--fg-4)', transform: isOpen ? 'rotate(90deg)' : 'none', transition: 'transform var(--dur-fast) var(--ease-out)', display: 'inline-flex' }}><I.chevronRight size={15} /></span>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 600, minWidth: 150 }}>{e.probe_name}</span>
                  <ProbeStatusBadge status={e.status} size="sm" />
                  <TierBadge tier={e.evidence_class} appearance="dot" />
                  <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-3)', flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{e.reason}</span>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-4)' }}>{e.ts}</span>
                </div>
                {isOpen && (
                  <div style={{ padding: '0 16px 16px 42px', display: 'flex', flexDirection: 'column', gap: 14 }}>
                    {emptyWarn && (
                      <div style={{ display: 'flex', gap: 10, padding: 11, background: 'var(--warn-bg)', borderRadius: 'var(--radius-md)' }}>
                        <span style={{ color: 'var(--warn)', flex: '0 0 auto' }}><I.alert size={15} /></span>
                        <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-2)', lineHeight: 1.45 }}>Probe completed but returned no useful evidence. Consider normalizing this as <b style={{ fontFamily: 'var(--font-mono)' }}>no_data</b>.</span>
                      </div>
                    )}
                    {e.confidence > 0 && <div style={{ maxWidth: 300 }}><ConfidenceBar value={e.confidence} label="Probe confidence" size="sm" /></div>}
                    {e.limitations && (
                      <div>
                        <div style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)', marginBottom: 5 }}>Limitations</div>
                        <p style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.5 }}>{e.limitations}</p>
                      </div>
                    )}
                    <div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                        <span style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)' }}>Raw evidence</span>
                        <Badge tone="neutral" appearance="outline" mono size="sm">{e.evidence_class}</Badge>
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
    );
  }

  window.IAD_SCREENS = window.IAD_SCREENS || {};
  window.IAD_SCREENS.evidence = Evidence;
})();
