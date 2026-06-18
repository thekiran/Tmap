/* IAD UI kit — Dashboard screen. Registers window.IAD_SCREENS.dashboard */
(function () {
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const { Card, MetricStat, ConfidenceBar, Badge, TierBadge, ProbeStatusBadge, StatusDot } = NS;
  const I = window.Icons;
  const fmt = window.fmt;
  window.IAD_SCREENS = window.IAD_SCREENS || {};

  const Eyebrow = ({ children }) => (
    <div style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-3)', marginBottom: 8 }}>{children}</div>
  );

  function VerdictCard({ scan }) {
    const c = scan.classification_confidence;
    return (
      <div style={{ gridColumn: 'span 2', background: 'var(--surface-card)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-lg)', padding: 24, display: 'flex', flexDirection: 'column', gap: 18, position: 'relative', overflow: 'hidden' }}>
        <div style={{ position: 'absolute', inset: 0, backgroundImage: 'radial-gradient(var(--grid-line) 1px, transparent 1px)', backgroundSize: '22px 22px', opacity: 0.6, pointerEvents: 'none' }} />
        <div style={{ position: 'relative', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 20 }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <Eyebrow>Access type decision</Eyebrow>
            <div style={{ font: 'var(--type-verdict)', color: 'var(--fg-1)', letterSpacing: 'var(--ls-tight)' }}>{scan.primary_type}</div>
            <div style={{ display: 'flex', gap: 8, marginTop: 10, alignItems: 'center', flexWrap: 'wrap' }}>
              <Badge tone="neutral" appearance="outline" mono size="sm">{scan.category}</Badge>
              <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>estimated · physical medium not directly confirmed</span>
            </div>
          </div>
          <div style={{ width: 200, flex: '0 0 auto' }}>
            <ConfidenceBar value={c} label="Classification" />
            <div style={{ marginTop: 12 }}><ConfidenceBar value={scan.context_confidence} label="Network context" /></div>
          </div>
        </div>
        <div style={{ position: 'relative', display: 'flex', gap: 8, flexWrap: 'wrap', borderTop: '1px solid var(--hairline)', paddingTop: 16 }}>
          {scan.candidates.map((c, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '6px 10px', background: i === 0 ? 'var(--accent-ghost)' : 'var(--bg-sunken)', border: '1px solid ' + (i === 0 ? 'var(--accent-ring)' : 'var(--hairline)'), borderRadius: 'var(--radius-md)' }}>
              <span style={{ fontSize: 'var(--text-xs)', fontWeight: 600, color: i === 0 ? 'var(--accent-bright)' : 'var(--fg-2)' }}>{c.type}</span>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: fmt.bandColor(c.score) }}>{fmt.pct(c.score)}</span>
            </div>
          ))}
        </div>
      </div>
    );
  }

  function UncertaintyCard({ scan }) {
    return (
      <Card eyebrow="Why not certain" title="Uncertainty" style={{ gridColumn: 'span 2' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 11 }}>
          {scan.uncertainty_reasons.map((r, i) => (
            <div key={i} style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
              <span style={{ color: 'var(--warn)', flex: '0 0 auto', marginTop: 1 }}><I.info size={15} /></span>
              <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)', lineHeight: 1.45 }}>{r}</span>
            </div>
          ))}
        </div>
      </Card>
    );
  }

  function GatewayChainCard({ scan }) {
    return (
      <Card eyebrow="Path to ISP" title="Gateway chain" style={{ gridColumn: 'span 2' }}>
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          {scan.gateway_chain.map((g, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '9px 0', borderBottom: i < scan.gateway_chain.length - 1 ? '1px solid var(--hairline)' : 'none' }}>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-4)', width: 18 }}>{g.hop}</span>
              <span style={{ width: 8, height: 8, borderRadius: '50%', background: g.private ? 'var(--fg-4)' : 'var(--tier-isp)', flex: '0 0 auto' }} />
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)', color: 'var(--fg-1)', minWidth: 116 }}>{g.ip}</span>
              <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)' }}>{g.label}</span>
              {g.note && <Badge tone="warn" size="sm">{g.note}</Badge>}
              <span style={{ marginLeft: 'auto', fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-3)' }}>{g.rtt_ms} ms</span>
            </div>
          ))}
        </div>
      </Card>
    );
  }

  function ConfidenceBreakdownCard({ scan }) {
    return (
      <Card eyebrow="What moved the needle" title="Confidence breakdown" style={{ gridColumn: 'span 2' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 13 }}>
          {scan.confidence_breakdown.map((f, i) => {
            const up = f.direction === 'up';
            const mag = Math.min(1, Math.abs(f.contribution) / 0.3);
            return (
              <div key={i} style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', gap: 8 }}>
                  <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 500 }}>{f.factor}</span>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, color: up ? 'var(--ok)' : 'var(--danger)', fontWeight: 600 }}>{up ? '+' : '−'}{Math.abs(f.contribution).toFixed(2)}</span>
                </div>
                <div style={{ height: 4, borderRadius: 'var(--radius-pill)', background: 'var(--surface-3)', overflow: 'hidden', display: 'flex', justifyContent: up ? 'flex-start' : 'flex-end' }}>
                  <div style={{ width: (mag * 100) + '%', height: '100%', background: up ? 'var(--ok)' : 'var(--danger)', borderRadius: 'var(--radius-pill)' }} />
                </div>
                <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.4 }}>{f.detail}</span>
              </div>
            );
          })}
        </div>
      </Card>
    );
  }

  function NextProbesCard({ scan }) {
    return (
      <Card eyebrow="Raise confidence" title="Next best probes" style={{ gridColumn: 'span 2' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {scan.next_best_probes.map((p, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: 12, background: 'var(--bg-sunken)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-md)' }}>
              <TierBadge tier={p.tier} appearance="dot" />
              <div style={{ minWidth: 0, flex: 1 }}>
                <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 500 }}>{p.name}</div>
                <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>requires: {p.requires}</div>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 1 }}>
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, color: 'var(--ok)', fontWeight: 600 }}>+{fmt.pct(p.gain)}</span>
                <span style={{ fontSize: 9, color: 'var(--fg-4)', textTransform: 'uppercase', letterSpacing: '.08em' }}>est. gain</span>
              </div>
            </div>
          ))}
        </div>
      </Card>
    );
  }

  function MiniCard({ eyebrow, children }) {
    return (
      <Card padding="md">
        <Eyebrow>{eyebrow}</Eyebrow>
        {children}
      </Card>
    );
  }

  function Dashboard({ scan }) {
    const ni = scan.detected_network_context.selected_interface;
    const p = scan.performance;
    return (
      <div style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 16, maxWidth: 1180, margin: '0 auto' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
          <h1 style={{ font: 'var(--type-h1)' }}>Overview</h1>
          <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>Scanned {fmt.ago(scan.created_at)} · {scan.duration_ms / 1000}s · {scan.evidence.length} probes</span>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16 }}>
          <VerdictCard scan={scan} />
          <UncertaintyCard scan={scan} />

          <MiniCard eyebrow="Local interface">
            <MetricStat label={ni.name + ' · IPv4'} value={ni.ipv4} secondary={'/' + ni.prefix + ' · ' + ni.link_speed_mbps + ' Mbps link · DHCP'} />
            <div style={{ marginTop: 10, fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-3)' }}>{ni.mac}</div>
          </MiniCard>
          <MiniCard eyebrow="Public IP">
            <MetricStat label="Address" value={scan.public_ip.address} secondary={'PTR ✓ ' + scan.public_ip.ptr} />
          </MiniCard>
          <MiniCard eyebrow="ISP / ASN">
            <MetricStat label={scan.public_ip.asn} value={scan.public_ip.org} secondary={scan.public_ip.city + ', ' + scan.public_ip.country} />
          </MiniCard>
          <MiniCard eyebrow="Throughput">
            <div style={{ display: 'flex', gap: 18 }}>
              <MetricStat label="Down" value={p.downstream_mbps} unit="Mbps" />
              <MetricStat label="Up" value={p.upstream_mbps} unit="Mbps" />
            </div>
            <div style={{ marginTop: 8, fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-3)' }}>{p.latency_ms} ms · jitter {p.jitter_ms} ms · loss {p.loss_pct}%</div>
          </MiniCard>

          <MiniCard eyebrow="IPv4 NAT status">
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Badge tone="warn" appearance="solid">CGNAT</Badge>
              <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)' }}>{scan.nat_topology.layers} layers</span>
            </div>
            <div style={{ marginTop: 8, fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.4 }}>{scan.nat_topology.note}</div>
          </MiniCard>
          <MiniCard eyebrow="IPv6 status">
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <StatusDot tone="success" />
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)', color: 'var(--fg-1)' }}>{scan.ipv6_context.global_address}</span>
            </div>
            <div style={{ marginTop: 8, fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.4 }}>{scan.ipv6_context.note}</div>
          </MiniCard>
          <MiniCard eyebrow="Network status">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 7 }}>
              <StatusDot tone="success" label="Internet reachable (IPv4 + IPv6)" />
              <StatusDot tone="warn" label="Inbound blocked by CGNAT" />
              <StatusDot tone="success" label={scan.devices.length + ' LAN devices discovered'} />
            </div>
          </MiniCard>

          <GatewayChainCard scan={scan} />
          <ConfidenceBreakdownCard scan={scan} />
          <NextProbesCard scan={scan} />

          <Card eyebrow="Advisories" title="Warnings" style={{ gridColumn: 'span 2' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 9 }}>
              {scan.warnings.map((w, i) => (
                <div key={i} style={{ display: 'flex', gap: 10, alignItems: 'flex-start', padding: 11, background: w.level === 'warn' ? 'var(--warn-bg)' : 'var(--info-bg)', borderRadius: 'var(--radius-md)' }}>
                  <span style={{ color: w.level === 'warn' ? 'var(--warn)' : 'var(--info)', flex: '0 0 auto' }}>{w.level === 'warn' ? <I.alert size={15} /> : <I.info size={15} />}</span>
                  <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)', lineHeight: 1.45 }}>{w.text}</span>
                </div>
              ))}
            </div>
          </Card>
        </div>
      </div>
    );
  }

  window.IAD_SCREENS.dashboard = Dashboard;
})();
