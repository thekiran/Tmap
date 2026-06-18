/* IAD UI kit — right-side details drawer. window.DetailsDrawer
   Renders device / node / edge / probe detail. Read-only: no delete actions. */
(function () {
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const { Badge, IconButton, ConfidenceBar, ProbeStatusBadge, TierBadge, StatusDot } = NS;
  const I = window.Icons;
  const fmt = window.fmt;

  const Row = ({ k, v, mono = true }) => (
    <div style={{ display: 'flex', justifyContent: 'space-between', gap: 14, padding: '8px 0', borderBottom: '1px solid var(--hairline)' }}>
      <span style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)', whiteSpace: 'nowrap' }}>{k}</span>
      <span style={{ fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)', fontSize: 'var(--text-sm)', color: 'var(--fg-1)', textAlign: 'right', wordBreak: 'break-word' }}>{v}</span>
    </div>
  );

  const Section = ({ title, children }) => (
    <div style={{ marginTop: 18 }}>
      <div style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-3)', marginBottom: 8 }}>{title}</div>
      {children}
    </div>
  );

  function DeviceBody({ d, scan }) {
    const r = fmt.reach(d.reachability);
    const Icon = I[fmt.deviceIcon(d.type)];
    const ev = scan.evidence.filter((e) => e.evidence_class === 'l2' || e.evidence_class === 'l3').slice(0, 3);
    return (
      <div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 4 }}>
          <span style={{ width: 40, height: 40, borderRadius: 'var(--radius-md)', background: 'var(--surface-3)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--fg-2)', flex: '0 0 auto' }}><Icon size={20} /></span>
          <div style={{ minWidth: 0 }}>
            <div style={{ font: 'var(--type-h3)', color: 'var(--fg-1)' }}>{d.hostname !== '—' ? d.hostname : d.ip}</div>
            <div style={{ display: 'flex', gap: 6, marginTop: 4 }}>
              <Badge tone={r.tone} size="sm">{r.word}</Badge>
              <Badge tone="neutral" appearance="outline" mono size="sm">{d.type}</Badge>
            </div>
          </div>
        </div>
        <Section title="Confidence">
          <ConfidenceBar value={d.confidence} label="Identification" />
        </Section>
        <Section title="Addresses">
          <Row k="IPv4" v={d.ip} />
          <Row k="MAC" v={d.mac} />
          <Row k="Vendor" v={d.vendor} mono={false} />
          <Row k="Hostname" v={d.hostname} />
        </Section>
        <Section title="Role & services">
          <Row k="Role" v={d.role} mono={false} />
          <Row k="Source probe" v={d.source} />
          <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', marginTop: 10 }}>
            {d.services.length && d.services[0] !== '—' ? d.services.map((s, i) => <Badge key={i} tone="neutral" mono size="sm">{s}</Badge>) : <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-4)' }}>No services fingerprinted</span>}
          </div>
        </Section>
        <Section title="Evidence">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 7 }}>
            {ev.map((e) => (
              <div key={e.id} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '7px 9px', background: 'var(--bg-sunken)', borderRadius: 'var(--radius-sm)' }}>
                <ProbeStatusBadge status={e.status} size="sm" />
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-2)' }}>{e.probe_name}</span>
              </div>
            ))}
          </div>
        </Section>
        <Section title="Limitations">
          <p style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.5 }}>
            Identity inferred from {d.source.toUpperCase()} responses on the local broadcast domain. Device role is a best-effort classification and may be wrong for multi-purpose or virtualized hosts.
          </p>
        </Section>
      </div>
    );
  }

  function ProbeBody({ e }) {
    return (
      <div>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 6 }}>
          <div style={{ font: 'var(--type-h3)', color: 'var(--fg-1)', fontFamily: 'var(--font-mono)' }}>{e.probe_name}</div>
          <ProbeStatusBadge status={e.status} />
        </div>
        <div style={{ display: 'flex', gap: 6 }}>
          <TierBadge tier={e.evidence_class} />
          <Badge tone="neutral" appearance="outline" mono size="sm">{e.ts}</Badge>
        </div>
        {e.status !== 'no_data' && e.status !== 'skipped' && e.status !== 'blocked' && e.status !== 'failed' && (
          <Section title="Confidence"><ConfidenceBar value={e.confidence} label="Probe confidence" /></Section>
        )}
        <Section title="Reason"><p style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)', lineHeight: 1.5 }}>{e.reason}</p></Section>
        {e.limitations && <Section title="Limitations"><p style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.5 }}>{e.limitations}</p></Section>}
        <Section title="Raw evidence">
          <pre style={{ margin: 0, padding: 12, background: 'var(--ink-900)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-md)', fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-2)', lineHeight: 1.5, overflow: 'auto', whiteSpace: 'pre-wrap' }}>{JSON.stringify(e.data, null, 2)}</pre>
        </Section>
      </div>
    );
  }

  function EdgeBody({ edge }) {
    return (
      <div>
        <div style={{ font: 'var(--type-h3)', color: 'var(--fg-1)', marginBottom: 6 }}>{edge.label}</div>
        <div style={{ display: 'flex', gap: 6 }}>
          <Badge tone={edge.kind === 'confirmed' ? 'success' : edge.kind === 'inferred' ? 'warn' : 'neutral'} size="sm" uppercase mono>{edge.kind}</Badge>
          {edge.tier && <TierBadge tier={edge.tier} />}
        </div>
        <Section title="Relationship">
          <Row k="From" v={edge.fromLabel} mono={false} />
          <Row k="To" v={edge.toLabel} mono={false} />
          <Row k="Type" v={edge.type} />
        </Section>
        <Section title="Basis"><p style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)', lineHeight: 1.5 }}>{edge.basis}</p></Section>
        {edge.kind !== 'confirmed' && (
          <Section title="Caveat">
            <div style={{ display: 'flex', gap: 9, padding: 11, background: 'var(--warn-bg)', borderRadius: 'var(--radius-md)' }}>
              <span style={{ color: 'var(--warn)', flex: '0 0 auto' }}><I.alert size={15} /></span>
              <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-2)', lineHeight: 1.45 }}>This link is inferred, not confirmed. The physical path may differ — traceroute hops are L3 routers, not switches.</span>
            </div>
          </Section>
        )}
      </div>
    );
  }

  function DetailsDrawer({ drawer, onClose, scan }) {
    let title = 'Details';
    let body = null;
    if (drawer.kind === 'device') { title = 'Device'; body = <DeviceBody d={drawer.payload} scan={scan} />; }
    else if (drawer.kind === 'probe') { title = 'Probe'; body = <ProbeBody e={drawer.payload} />; }
    else if (drawer.kind === 'node') { title = drawer.payload.device ? 'Device' : 'Node'; body = drawer.payload.device ? <DeviceBody d={drawer.payload.device} scan={scan} /> : <ProbeBody e={drawer.payload} />; }
    else if (drawer.kind === 'edge') { title = 'Edge evidence'; body = <EdgeBody edge={drawer.payload} />; }

    return (
      <aside style={{ width: 360, flex: '0 0 auto', background: 'var(--surface-card)', borderLeft: '1px solid var(--hairline)', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
        <div style={{ height: 'var(--topbar-h)', flex: '0 0 auto', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 16px', borderBottom: '1px solid var(--hairline)' }}>
          <span style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-3)' }}>{title}</span>
          <IconButton label="Close" onClick={onClose}><I.close size={16} /></IconButton>
        </div>
        <div style={{ flex: 1, overflow: 'auto', padding: 18 }}>{body}</div>
      </aside>
    );
  }

  window.DetailsDrawer = DetailsDrawer;
})();
