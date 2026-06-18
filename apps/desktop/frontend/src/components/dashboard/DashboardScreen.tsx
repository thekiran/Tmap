import type { ReactNode } from 'react';
import { useScanStore } from '../../store/useScanStore';
import { Card, Badge, StatusDot, Overline } from '../ui/primitives';
import { ConfidenceBar, MetricStat, TierBadge } from '../ui/data';
import { Icons } from '../icons/Icon';
import { pct, ago } from '../../lib/format';
import { bandColorVar } from '../../lib/confidence';
import type { NormalizedScanReport } from '../../lib/models';

const Mini = ({ eyebrow, children }: { eyebrow: string; children: ReactNode }) => (
  <Card padding="md"><Overline style={{ marginBottom: 8 }}>{eyebrow}</Overline>{children}</Card>
);

function VerdictCard({ scan }: { scan: NormalizedScanReport }) {
  const unknown = scan.isUnknown;
  return (
    <div style={{ gridColumn: 'span 2', background: 'var(--surface-card)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-lg)', padding: 24, display: 'flex', flexDirection: 'column', gap: 18, position: 'relative', overflow: 'hidden' }}>
      <div style={{ position: 'absolute', inset: 0, backgroundImage: 'radial-gradient(var(--grid-line) 1px, transparent 1px)', backgroundSize: '22px 22px', opacity: 0.6, pointerEvents: 'none' }} />
      <div style={{ position: 'relative', display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 20 }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <Overline>Access type decision</Overline>
          <div style={{ font: 'var(--type-verdict)', color: unknown ? 'var(--fg-2)' : 'var(--fg-1)', letterSpacing: 'var(--ls-tight)', marginTop: 6 }}>
            {unknown ? 'Unknown' : scan.primaryType}
          </div>
          <div style={{ display: 'flex', gap: 8, marginTop: 10, alignItems: 'center', flexWrap: 'wrap' }}>
            {scan.category && <Badge tone="neutral" appearance="outline" mono size="sm">{scan.category}</Badge>}
            <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>
              {unknown ? 'Not enough evidence to decide — candidates below' : 'estimated · physical medium not directly confirmed'}
            </span>
          </div>
        </div>
        <div style={{ width: 200, flex: '0 0 auto' }}>
          <ConfidenceBar value={scan.classificationConfidence} label="Classification" />
          <div style={{ marginTop: 12 }}><ConfidenceBar value={scan.contextConfidence} label="Network context" /></div>
        </div>
      </div>
      {scan.candidates.length > 0 && (
        <div style={{ position: 'relative', display: 'flex', gap: 8, flexWrap: 'wrap', borderTop: '1px solid var(--hairline)', paddingTop: 16 }}>
          {scan.candidates.map((c, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '6px 10px', background: i === 0 && !unknown ? 'var(--accent-ghost)' : 'var(--bg-sunken)', border: '1px solid ' + (i === 0 && !unknown ? 'var(--accent-ring)' : 'var(--hairline)'), borderRadius: 'var(--radius-md)' }}>
              <span style={{ fontSize: 'var(--text-xs)', fontWeight: 600, color: i === 0 && !unknown ? 'var(--accent-bright)' : 'var(--fg-2)' }}>{c.type}</span>
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: bandColorVar(c.score) }}>{pct(c.score)}</span>
            </div>
          ))}
        </div>
      )}
      {unknown && (
        <div style={{ position: 'relative', display: 'flex', gap: 9, padding: 11, background: 'var(--warn-bg)', borderRadius: 'var(--radius-md)' }}>
          <span style={{ color: 'var(--warn)', flex: '0 0 auto' }}><Icons.info size={15} /></span>
          <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-2)', lineHeight: 1.45 }}>Access type is unknown. The console will not claim Fiber/VDSL/DSL certainty without supporting evidence. See missing evidence under Uncertainty.</span>
        </div>
      )}
    </div>
  );
}

export function DashboardScreen() {
  const scan = useScanStore((s) => s.normalized)!;
  const ni = scan.selectedInterface;
  const p = scan.performance;
  return (
    <div className="app-scroll" style={{ flex: 1, overflow: 'auto' }}>
      <div style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 16, maxWidth: 1180, margin: '0 auto' }}>
        <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
          <h1 style={{ font: 'var(--type-h1)' }}>Overview</h1>
          <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>Scanned {ago(scan.createdAt)}{scan.durationMs ? ` · ${(scan.durationMs / 1000).toFixed(2)}s` : ''} · {scan.evidence.length} probes</span>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16 }}>
          <VerdictCard scan={scan} />

          <Card eyebrow="Why not certain" title="Uncertainty" style={{ gridColumn: 'span 2' }}>
            {scan.uncertaintyReasons.length ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 11 }}>
                {scan.uncertaintyReasons.map((r, i) => (
                  <div key={i} style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
                    <span style={{ color: 'var(--warn)', flex: '0 0 auto', marginTop: 1 }}><Icons.info size={15} /></span>
                    <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)', lineHeight: 1.45 }}>{r}</span>
                  </div>
                ))}
              </div>
            ) : <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-3)' }}>No uncertainty notes recorded.</span>}
          </Card>

          {ni && (
            <Mini eyebrow="Local interface">
              <MetricStat label={`${ni.name} · IPv4`} value={ni.ipv4 ?? '—'} secondary={`${ni.prefix ? '/' + ni.prefix : ''}${ni.linkSpeedMbps ? ' · ' + ni.linkSpeedMbps + ' Mbps' : ''}${ni.dhcp ? ' · DHCP' : ''}`} />
              {ni.mac && <div style={{ marginTop: 10, fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-3)' }}>{ni.mac}</div>}
            </Mini>
          )}
          {scan.publicIp && (
            <Mini eyebrow="Public IP">
              <MetricStat label="Address" value={scan.publicIp.address ?? '—'} secondary={scan.publicIp.ptr ? 'PTR ✓ ' + scan.publicIp.ptr : 'No PTR'} />
            </Mini>
          )}
          {scan.publicIp && (
            <Mini eyebrow="ISP / ASN">
              <MetricStat label={scan.publicIp.asn ?? 'ASN'} value={scan.publicIp.org ?? '—'} secondary={[scan.publicIp.city, scan.publicIp.country].filter(Boolean).join(', ')} />
            </Mini>
          )}
          {p && (
            <Mini eyebrow="Throughput">
              <div style={{ display: 'flex', gap: 18 }}>
                <MetricStat label="Down" value={p.downstreamMbps ?? '—'} unit="Mbps" />
                <MetricStat label="Up" value={p.upstreamMbps ?? '—'} unit="Mbps" />
              </div>
              <div style={{ marginTop: 8, fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-3)' }}>
                {p.latencyMs ?? '—'} ms · jitter {p.jitterMs ?? '—'} ms · loss {p.lossPct ?? '—'}%
              </div>
            </Mini>
          )}

          {scan.nat && (
            <Mini eyebrow="IPv4 NAT status">
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Badge tone="warn" appearance="solid">{(scan.nat.type ?? 'nat').toUpperCase()}</Badge>
                {scan.nat.layers != null && <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)' }}>{scan.nat.layers} layers</span>}
              </div>
              {scan.nat.note && <div style={{ marginTop: 8, fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.4 }}>{scan.nat.note}</div>}
            </Mini>
          )}
          {scan.ipv6 && (
            <Mini eyebrow="IPv6 status">
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <StatusDot tone={scan.ipv6.available ? 'success' : 'neutral'} />
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)', color: 'var(--fg-1)' }}>{scan.ipv6.globalAddress ?? (scan.ipv6.available ? 'available' : 'none')}</span>
              </div>
              {scan.ipv6.note && <div style={{ marginTop: 8, fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.4 }}>{scan.ipv6.note}</div>}
            </Mini>
          )}
          <Mini eyebrow="Network status">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 7 }}>
              <StatusDot tone="success" label={`Internet reachable${scan.ipv6?.available ? ' (IPv4 + IPv6)' : ''}`} />
              {scan.nat && scan.nat.publicReachable === false && <StatusDot tone="warn" label="Inbound blocked by NAT" />}
              <StatusDot tone="success" label={`${scan.devices.length} LAN devices discovered`} />
            </div>
          </Mini>

          <Card eyebrow="Path to ISP" title="Gateway chain" style={{ gridColumn: 'span 2' }}>
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {scan.gatewayChain.map((g, i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '9px 0', borderBottom: i < scan.gatewayChain.length - 1 ? '1px solid var(--hairline)' : 'none' }}>
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-4)', width: 18 }}>{g.hop}</span>
                  <span style={{ width: 8, height: 8, borderRadius: '50%', background: g.private ? 'var(--fg-4)' : 'var(--tier-isp)', flex: '0 0 auto' }} />
                  <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)', color: 'var(--fg-1)', minWidth: 116 }}>{g.ip}</span>
                  <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)' }}>{g.label}</span>
                  {g.note && <Badge tone="warn" size="sm">{g.note}</Badge>}
                  {g.rttMs != null && <span style={{ marginLeft: 'auto', fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-3)' }}>{g.rttMs} ms</span>}
                </div>
              ))}
            </div>
          </Card>

          <Card eyebrow="What moved the needle" title="Confidence breakdown" style={{ gridColumn: 'span 2' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 13 }}>
              {scan.confidenceBreakdown.map((f, i) => {
                const up = f.direction === 'up';
                const mag = Math.min(1, Math.abs(f.contribution) / 0.3);
                return (
                  <div key={i} style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', gap: 8 }}>
                      <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 500 }}>{f.factor}</span>
                      <span style={{ fontFamily: 'var(--font-mono)', fontSize: 12, color: up ? 'var(--ok)' : 'var(--danger)', fontWeight: 600 }}>{up ? '+' : '−'}{Math.abs(f.contribution).toFixed(2)}</span>
                    </div>
                    <div style={{ height: 4, borderRadius: 'var(--radius-pill)', background: 'var(--surface-3)', overflow: 'hidden', display: 'flex', justifyContent: up ? 'flex-start' : 'flex-end' }}>
                      <div style={{ width: mag * 100 + '%', height: '100%', background: up ? 'var(--ok)' : 'var(--danger)', borderRadius: 'var(--radius-pill)' }} />
                    </div>
                    {f.detail && <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.4 }}>{f.detail}</span>}
                  </div>
                );
              })}
            </div>
          </Card>

          {scan.nextBestProbes.length > 0 && (
            <Card eyebrow="Raise confidence" title="Next best probes" style={{ gridColumn: 'span 2' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                {scan.nextBestProbes.map((p2, i) => (
                  <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: 12, background: 'var(--bg-sunken)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-md)' }}>
                    <TierBadge tier={p2.tier} appearance="dot" />
                    <div style={{ minWidth: 0, flex: 1 }}>
                      <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 500 }}>{p2.name}</div>
                      {p2.requires && <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>requires: {p2.requires}</div>}
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 1 }}>
                      <span style={{ fontFamily: 'var(--font-mono)', fontSize: 13, color: 'var(--ok)', fontWeight: 600 }}>+{pct(p2.gain)}</span>
                      <span style={{ fontSize: 9, color: 'var(--fg-4)', textTransform: 'uppercase', letterSpacing: '.08em' }}>est. gain</span>
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          )}

          {scan.warnings.length > 0 && (
            <Card eyebrow="Advisories" title="Warnings" style={{ gridColumn: 'span 4' }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 9 }}>
                {scan.warnings.map((w, i) => (
                  <div key={i} style={{ display: 'flex', gap: 10, alignItems: 'flex-start', padding: 11, background: w.level === 'warn' ? 'var(--warn-bg)' : w.level === 'danger' ? 'var(--danger-bg)' : 'var(--info-bg)', borderRadius: 'var(--radius-md)' }}>
                    <span style={{ color: w.level === 'warn' ? 'var(--warn)' : w.level === 'danger' ? 'var(--danger)' : 'var(--info)', flex: '0 0 auto' }}>{w.level === 'info' ? <Icons.info size={15} /> : <Icons.alert size={15} />}</span>
                    <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)', lineHeight: 1.45 }}>{w.text}</span>
                  </div>
                ))}
              </div>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
