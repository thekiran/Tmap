import type { ReactNode } from 'react';
import { useScanStore } from '../../store/useScanStore';
import { useUIStore } from '../../store/useUIStore';
import { Badge, IconButton, Overline } from '../ui/primitives';
import { ConfidenceBar, ProbeStatusBadge, TierBadge } from '../ui/data';
import { Icons, Icon } from '../icons/Icon';
import { REACH_META, deviceIconKey } from '../../lib/format';
import type { NetworkDevice, EvidenceRecord, TopologyEdge } from '../../lib/models';

const Row = ({ k, v, mono = true }: { k: string; v: ReactNode; mono?: boolean }) => (
  <div style={{ display: 'flex', justifyContent: 'space-between', gap: 14, padding: '8px 0', borderBottom: '1px solid var(--hairline)' }}>
    <Overline>{k}</Overline>
    <span style={{ fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)', fontSize: 'var(--text-sm)', color: 'var(--fg-1)', textAlign: 'right', wordBreak: 'break-word' }}>{v}</span>
  </div>
);
const Section = ({ title, children }: { title: string; children: ReactNode }) => (
  <div style={{ marginTop: 18 }}><Overline style={{ marginBottom: 8 }}>{title}</Overline>{children}</div>
);

function DeviceBody({ d, evidence }: { d: NetworkDevice; evidence: EvidenceRecord[] }) {
  const r = REACH_META[d.reachability];
  const refs = evidence.filter((e) => d.rawProbeRefs.includes(e.id));
  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 4 }}>
        <span style={{ width: 40, height: 40, borderRadius: 'var(--radius-md)', background: 'var(--surface-3)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--fg-2)', flex: '0 0 auto' }}>
          <Icon name={deviceIconKey(d.type) as never} size={20} />
        </span>
        <div style={{ minWidth: 0 }}>
          <div style={{ font: 'var(--type-h3)', color: 'var(--fg-1)' }}>{d.hostname ?? d.ip}</div>
          <div style={{ display: 'flex', gap: 6, marginTop: 4 }}>
            <Badge tone={r.tone as never} size="sm">{r.word}</Badge>
            <Badge tone="neutral" appearance="outline" mono size="sm">{d.type}</Badge>
          </div>
        </div>
      </div>
      <Section title="Confidence"><ConfidenceBar value={d.confidence} label="Identification" /></Section>
      <Section title="Addresses">
        <Row k="IPv4" v={d.ips.join(', ') || '—'} />
        <Row k="MAC" v={d.mac ?? '—'} />
        <Row k="Vendor" v={d.vendor ?? '—'} mono={false} />
        <Row k="Hostname" v={d.hostname ?? '—'} />
      </Section>
      <Section title="Role & services">
        <Row k="Role" v={d.role ?? '—'} mono={false} />
        <Row k="Source probe" v={d.source ?? '—'} />
        <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', marginTop: 10 }}>
          {d.services.length ? d.services.map((s, i) => <Badge key={i} tone="neutral" mono size="sm">{s.name}</Badge>) : <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-4)' }}>No services fingerprinted</span>}
        </div>
      </Section>
      {refs.length > 0 && (
        <Section title="Raw probe references">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 7 }}>
            {refs.map((e) => (
              <div key={e.id} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '7px 9px', background: 'var(--bg-sunken)', borderRadius: 'var(--radius-sm)' }}>
                <ProbeStatusBadge status={e.status} size="sm" />
                <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-2)' }}>{e.probeName}</span>
              </div>
            ))}
          </div>
        </Section>
      )}
      {d.explanation && <Section title="Explanation"><p style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.5 }}>{d.explanation}</p></Section>}
      {d.limitations && <Section title="Limitations"><p style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.5 }}>{d.limitations}</p></Section>}
    </div>
  );
}

function ProbeBody({ e }: { e: EvidenceRecord }) {
  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 6 }}>
        <div style={{ font: 'var(--type-h3)', color: 'var(--fg-1)', fontFamily: 'var(--font-mono)' }}>{e.probeName}</div>
        <ProbeStatusBadge status={e.status} />
      </div>
      <div style={{ display: 'flex', gap: 6 }}>
        <TierBadge tier={e.evidenceClass} />
        <Badge tone="neutral" appearance="outline" mono size="sm">{e.timestamp}</Badge>
      </div>
      {e.confidence > 0 && <Section title="Confidence"><ConfidenceBar value={e.confidence} label="Probe confidence" size="sm" /></Section>}
      {e.reason && <Section title="Reason"><p style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)', lineHeight: 1.5 }}>{e.reason}</p></Section>}
      {e.limitations && <Section title="Limitations"><p style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.5 }}>{e.limitations}</p></Section>}
      <Section title="Raw evidence">
        <pre style={{ margin: 0, padding: 12, background: 'var(--ink-900)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-md)', fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-2)', lineHeight: 1.5, overflow: 'auto', whiteSpace: 'pre-wrap' }}>{JSON.stringify(e.data, null, 2)}</pre>
      </Section>
    </div>
  );
}

function EdgeBody({ edge, nodeLabel }: { edge: TopologyEdge; nodeLabel: (id: string) => string }) {
  const tone = edge.certainty === 'confirmed' ? 'success' : edge.certainty === 'inferred' ? 'warn' : 'neutral';
  return (
    <div>
      <div style={{ font: 'var(--type-h3)', color: 'var(--fg-1)', marginBottom: 6 }}>{edge.label}</div>
      <div style={{ display: 'flex', gap: 6 }}>
        <Badge tone={tone as never} size="sm" uppercase mono>{edge.certainty}</Badge>
        <TierBadge tier={edge.tier} />
      </div>
      <Section title="Relationship">
        <Row k="From" v={nodeLabel(edge.source)} mono={false} />
        <Row k="To" v={nodeLabel(edge.target)} mono={false} />
        <Row k="Type" v={edge.type} />
      </Section>
      <Section title="Basis"><p style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)', lineHeight: 1.5 }}>{edge.basis}</p></Section>
      {edge.certainty !== 'confirmed' && (
        <Section title="Caveat">
          <div style={{ display: 'flex', gap: 9, padding: 11, background: 'var(--warn-bg)', borderRadius: 'var(--radius-md)' }}>
            <span style={{ color: 'var(--warn)', flex: '0 0 auto' }}><Icons.alert size={15} /></span>
            <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-2)', lineHeight: 1.45 }}>This link is {edge.certainty}, not confirmed. The physical path may differ — traceroute hops are L3 routers, not switches.</span>
          </div>
        </Section>
      )}
    </div>
  );
}

export function DetailsDrawer() {
  const drawer = useUIStore((s) => s.drawer);
  const closeDrawer = useUIStore((s) => s.closeDrawer);
  const scan = useScanStore((s) => s.normalized);
  const topology = useScanStore((s) => s.topology);
  if (!drawer || !scan) return null;

  let title = 'Details';
  let body: ReactNode = null;
  if (drawer.kind === 'device' || drawer.kind === 'node') {
    const node = topology?.nodes.find((n) => n.id === drawer.id);
    const devId = drawer.kind === 'device' ? drawer.id : node?.deviceId;
    const device = scan.devices.find((d) => d.id === devId);
    if (device) { title = 'Device'; body = <DeviceBody d={device} evidence={scan.evidence} />; }
    else if (node) { title = 'Node'; body = <div><div style={{ font: 'var(--type-h3)', color: 'var(--fg-1)' }}>{node.label}</div><p style={{ marginTop: 8, fontSize: 'var(--text-sm)', color: 'var(--fg-3)' }}>{node.sublabel}</p><Section title="Provenance"><Badge tone={node.certainty === 'confirmed' ? 'success' : 'warn'} size="sm" uppercase mono>{node.certainty}</Badge></Section></div>; }
  } else if (drawer.kind === 'probe') {
    const e = scan.evidence.find((x) => x.id === drawer.id);
    if (e) { title = 'Probe'; body = <ProbeBody e={e} />; }
  } else if (drawer.kind === 'edge') {
    const edge = topology?.edges.find((x) => x.id === drawer.id);
    if (edge) { title = 'Edge evidence'; body = <EdgeBody edge={edge} nodeLabel={(id) => topology?.nodes.find((n) => n.id === id)?.label ?? id} />; }
  }

  return (
    <aside style={{ width: 360, flex: '0 0 auto', background: 'var(--surface-card)', borderLeft: '1px solid var(--hairline)', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
      <div style={{ height: 'var(--topbar-h)', flex: '0 0 auto', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 16px', borderBottom: '1px solid var(--hairline)' }}>
        <Overline>{title}</Overline>
        <IconButton label="Close" onClick={closeDrawer}><Icons.close size={16} /></IconButton>
      </div>
      <div className="app-scroll" style={{ flex: 1, overflow: 'auto', padding: 18 }}>{body}</div>
    </aside>
  );
}
