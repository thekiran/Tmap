import { useUIStore } from '../../store/useUIStore';
import { useScanStore } from '../../store/useScanStore';
import type { NetworkDevice, TopologyEdge } from '../../lib/models';
import {
  UNKNOWN_DEVICE_REASON,
  deviceDisplayTitle,
  deviceSecondaryHostname,
  discoverySourceBadges,
  formatTopologyEdgeLabel,
} from '../../lib/topology-display';

export function Inspector() {
  const { selectedDeviceId, selectedEdgeId, toggleInspector } = useUIStore();
  const normalized = useScanStore((state) => state.normalized);

  const device = normalized?.devices.find((d) => d.id === selectedDeviceId) ?? null;
  const edge = normalized?.topology.edges.find((e) => e.id === selectedEdgeId) ?? null;

  const title = edge && !device ? 'Link Details' : 'Device Details';

  return (
    <div className="flex h-full flex-col bg-white dark:bg-zinc-900/50">
      <div className="flex items-center justify-between border-b border-zinc-200 bg-zinc-50 px-4 py-3 dark:border-zinc-800/50 dark:bg-zinc-900">
        <h3 className="text-sm font-semibold text-zinc-800 dark:text-zinc-200">{title}</h3>
        <button onClick={toggleInspector} className="text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300">
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-4">
        {edge && !device ? (
          <EdgeDetails edge={edge} normalized={normalized!} />
        ) : device ? (
          <DeviceDetails device={device} normalized={normalized!} />
        ) : (
          <div className="mt-10 text-center text-sm text-zinc-500">
            Select a node or link in the topology map to view details.
          </div>
        )}
      </div>
    </div>
  );
}

function Row({ label, value, mono }: { label: string; value: React.ReactNode; mono?: boolean }) {
  return (
    <div className="flex justify-between gap-4 border-b border-zinc-100 py-1 dark:border-zinc-800">
      <span className="shrink-0 text-zinc-500">{label}</span>
      <span className={['text-right', mono ? 'font-mono text-[12px]' : ''].join(' ')}>{value}</span>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="mt-5">
      <div className="mb-2 font-mono text-[10px] uppercase tracking-wider text-zinc-500">{title}</div>
      {children}
    </div>
  );
}

function ConfidenceBar({ value }: { value: number }) {
  const pct = Math.round((value ?? 0) * 100);
  const color = pct >= 75 ? 'bg-emerald-500' : pct >= 45 ? 'bg-amber-500' : 'bg-zinc-400';
  return (
    <div className="flex items-center gap-2">
      <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-zinc-200 dark:bg-zinc-800">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${pct}%` }} />
      </div>
      <span className="font-mono text-[11px] text-zinc-600 dark:text-zinc-400">{pct}%</span>
    </div>
  );
}

function DeviceDetails({ device, normalized }: { device: NetworkDevice; normalized: NonNullable<ReturnType<typeof useScanStore.getState>['normalized']> }) {
  const openServices = device.services.filter((s) => (s.state ?? '').toLowerCase() === 'open');
  const evidence = normalized.evidence.filter((e) => device.rawProbeRefs.includes(e.id));
  const discovery = discoverySourceBadges(device.discoverySources, device.reachability);
  const secondaryHostname = deviceSecondaryHostname(device);

  return (
    <div className="text-sm text-zinc-700 dark:text-zinc-300">
      <div className="flex items-start justify-between gap-2">
        <div>
          <h2 className="text-lg font-bold text-zinc-900 dark:text-zinc-100">{deviceDisplayTitle(device)}</h2>
          <div className="font-mono text-xs text-zinc-500">{device.ip}</div>
          {secondaryHostname && <div className="mt-0.5 font-mono text-[11px] text-zinc-500">{secondaryHostname}</div>}
        </div>
        <div className="flex flex-wrap justify-end gap-1">
          {device.isGateway && <Tag tone="blue">gateway</Tag>}
          {device.isAgent && <Tag tone="green">this pc</Tag>}
          {device.isUnknown && <Tag>unknown</Tag>}
        </div>
      </div>

      <Section title="Confidence">
        <ConfidenceBar value={device.confidence} />
      </Section>

      {discovery.length > 0 && (
        <Section title="Discovered via">
          <div className="flex flex-wrap gap-1">
            {discovery.map((s) => (
              <Tag key={s} tone={s === 'TCP' || s === 'Nmap' ? 'green' : 'zinc'}>
                {s}
              </Tag>
            ))}
          </div>
        </Section>
      )}

      <Section title="Identity">
        <Row label="IP addresses" value={device.ips.length ? device.ips.join(', ') : device.ip} mono />
        <Row label="Hostname" value={device.hostname ?? 'Unknown'} mono />
        <Row label="MAC address" value={device.mac ?? 'Unknown'} mono />
        <Row label="Vendor / OUI" value={device.vendor ?? 'Unknown'} />
        <Row label="Type" value={<span className="capitalize">{device.type.replace(/_/g, ' ')}</span>} />
        <Row label="Roles" value={device.roles.length ? device.roles.join(', ') : 'Unknown'} />
        <Row label="Reachability" value={<span className="capitalize">{device.reachability}</span>} />
      </Section>

      {device.isUnknown && (
        <Section title="Why unknown">
          <p className="rounded-md border border-zinc-700 bg-zinc-950/35 px-3 py-2 text-xs leading-relaxed text-zinc-400">
            {UNKNOWN_DEVICE_REASON}
          </p>
        </Section>
      )}

      <Section title={`Open services (${openServices.length})`}>
        {openServices.length === 0 ? (
          <div className="text-xs text-zinc-500">No open services detected</div>
        ) : (
          <table className="w-full text-left text-[12px]">
            <thead>
              <tr className="font-mono text-[10px] uppercase text-zinc-400">
                <th className="py-1 font-normal">Service</th>
                <th className="py-1 font-normal">Port</th>
                <th className="py-1 font-normal">Proto</th>
                <th className="py-1 font-normal">State</th>
              </tr>
            </thead>
            <tbody>
              {openServices.map((s, i) => (
                <tr key={i} className="border-t border-zinc-100 dark:border-zinc-800">
                  <td className="py-1">{s.name}</td>
                  <td className="py-1 font-mono">{s.port ?? '—'}</td>
                  <td className="py-1 font-mono">{s.protocol ?? s.proto ?? 'tcp'}</td>
                  <td className="py-1 text-emerald-600 dark:text-emerald-400">{s.state ?? 'open'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>

      {device.riskFindings.length > 0 && (
        <Section title="Risk notes">
          <ul className="space-y-1 text-xs">
            {device.riskFindings.map((r) => (
              <li key={r.id} className="rounded border border-amber-500/30 bg-amber-500/5 px-2 py-1">
                <span className="font-semibold text-amber-600 dark:text-amber-400">{r.severity}</span> · {r.title}
              </li>
            ))}
          </ul>
        </Section>
      )}

      <Section title={`Evidence (${evidence.length})`}>
        {evidence.length === 0 ? (
          <div className="text-xs text-zinc-500">No linked evidence.</div>
        ) : (
          <ul className="space-y-1 text-xs">
            {evidence.map((e) => (
              <li key={e.id} className="rounded bg-zinc-100 px-2 py-1 dark:bg-zinc-800/60">
                <span className="font-mono text-[10px] text-zinc-500">{e.kind}</span> · {e.summary}
              </li>
            ))}
          </ul>
        )}
      </Section>

      <RawAccordion data={device.raw} />
    </div>
  );
}

function EdgeDetails({ edge, normalized }: { edge: TopologyEdge; normalized: NonNullable<ReturnType<typeof useScanStore.getState>['normalized']> }) {
  const label = (id: string) => {
    const d = normalized.devices.find((x) => x.id === id);
    return d ? deviceDisplayTitle(d) : id.replace(/^dev-/, '');
  };
  const evidenceIds = (edge.evidenceIds as string[] | undefined) ?? [];
  const evidence = normalized.evidence.filter((e) => evidenceIds.includes(e.id));
  const reason = (edge.reason as string | null) ?? null;
  const proof = (edge.proofSource as string | null) ?? edge.basis;
  const confLabel = (edge.confidenceLabel as string | null) ?? null;

  return (
    <div className="text-sm text-zinc-700 dark:text-zinc-300">
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <Tag tone={edge.physical ? 'green' : edge.lineStyle === 'dotted' ? 'zinc' : 'amber'}>
          {edge.physical ? 'physical' : edge.inferred ? 'inferred' : 'link'}
        </Tag>
        <Tag>{edge.tier.toUpperCase()}</Tag>
        {confLabel && <Tag>{confLabel} confidence</Tag>}
      </div>

      <div className="rounded-lg border border-zinc-200 bg-zinc-50 p-3 dark:border-zinc-800 dark:bg-zinc-900/50">
        <div className="font-medium">{label(edge.source)}</div>
        <div className="my-1 text-center font-mono text-[11px] text-zinc-400">
          {edge.physical ? '-----' : '. . . . .'} {formatTopologyEdgeLabel(edge)}
        </div>
        <div className="font-medium">{label(edge.target)}</div>
      </div>

      <Section title="Link">
        <Row label="Relationship" value={formatTopologyEdgeLabel(edge)} />
        <Row label="Layer" value={edge.tier.toUpperCase()} />
        <Row label="Physical" value={edge.physical ? 'Yes (proven)' : 'No'} />
        <Row label="Inferred" value={edge.inferred ? 'Yes' : 'No'} />
        <Row label="Proof source" value={proof} mono />
      </Section>

      <Section title="Confidence">
        <ConfidenceBar value={edge.confidence} />
      </Section>

      {reason && (
        <Section title="Why this link">
          <p className="text-xs leading-relaxed text-zinc-600 dark:text-zinc-400">{reason}</p>
        </Section>
      )}

      {!edge.physical && (
        <div className="mt-4 rounded-md border border-sky-500/30 bg-sky-500/10 px-3 py-2 text-xs text-sky-700 dark:text-sky-300">
          This link is inferred from routing/subnet evidence — it is not a proven physical cable.
        </div>
      )}

      {evidence.length > 0 && (
        <Section title={`Evidence (${evidence.length})`}>
          <ul className="space-y-1 text-xs">
            {evidence.map((e) => (
              <li key={e.id} className="rounded bg-zinc-100 px-2 py-1 dark:bg-zinc-800/60">
                <span className="font-mono text-[10px] text-zinc-500">{e.kind}</span> · {e.summary}
              </li>
            ))}
          </ul>
        </Section>
      )}

      <RawAccordion data={edge.rawEdge as Record<string, unknown> ?? edge} />
    </div>
  );
}

function RawAccordion({ data }: { data: unknown }) {
  return (
    <details className="mt-5">
      <summary className="cursor-pointer font-mono text-[10px] uppercase tracking-wider text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300">
        Raw JSON
      </summary>
      <pre className="mt-2 max-h-64 overflow-auto rounded-md border border-zinc-200 bg-zinc-950 p-3 font-mono text-[10px] leading-relaxed text-zinc-300 dark:border-zinc-800">
        {JSON.stringify(data, null, 2)}
      </pre>
    </details>
  );
}

function Tag({ children, tone = 'zinc' }: { children: React.ReactNode; tone?: 'zinc' | 'blue' | 'green' | 'amber' }) {
  const cls = {
    zinc: 'border-zinc-300 text-zinc-500 dark:border-zinc-700 dark:text-zinc-400',
    blue: 'border-blue-300 text-blue-600 dark:border-blue-400/50 dark:text-blue-300',
    green: 'border-emerald-300 text-emerald-600 dark:border-emerald-400/50 dark:text-emerald-300',
    amber: 'border-amber-300 text-amber-600 dark:border-amber-400/50 dark:text-amber-300',
  }[tone];
  return (
    <span className={`rounded border px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wide ${cls}`}>
      {children}
    </span>
  );
}
