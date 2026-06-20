import { Handle, Position, type NodeProps } from '@xyflow/react';
import type { TopologyNode } from '../../lib/models';
import { discoverySourceBadges, nodeDisplayTitle, nodeIp, nodeSecondaryHostname } from '../../lib/topology-display';
import { useUIStore } from '../../store/useUIStore';

function nodeTone(node: TopologyNode) {
  if (node.isGateway) return 'border-blue-400/70 bg-blue-500/12 text-blue-50 shadow-blue-950/20';
  if (node.isAgent) return 'border-emerald-400/70 bg-emerald-500/14 text-emerald-50 shadow-emerald-950/20 ring-1 ring-emerald-400/25';
  if (node.isUnknown) return 'border-zinc-600/80 bg-zinc-900/94 text-zinc-100 shadow-black/20';
  return 'border-zinc-700/80 bg-zinc-900/92 text-zinc-100 shadow-black/20';
}

function mobileBadge(classification?: string, confidence = 0): { label: string; className: string } | null {
  const base = 'mt-1 inline-flex max-w-[170px] items-center gap-1 rounded border px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wide';
  const strong = confidence >= 0.65;
  const tone = classification === 'conflicting_mobile_os_evidence'
    ? 'border-amber-400/60 bg-amber-500/10 text-amber-300'
    : strong
      ? 'border-sky-400/60 bg-sky-500/12 text-sky-200'
      : 'border-zinc-700 bg-zinc-950/40 text-zinc-400';
  switch (classification) {
    case 'confirmed_ios': return { label: 'Confirmed iPhone', className: `${base} ${tone}` };
    case 'probable_ios': return { label: 'Probable iPhone', className: `${base} ${tone}` };
    case 'possible_ios': return { label: 'Possible iPhone', className: `${base} ${tone}` };
    case 'confirmed_ipados': return { label: 'Confirmed iPad', className: `${base} ${tone}` };
    case 'probable_ipados': return { label: 'Probable iPad', className: `${base} ${tone}` };
    case 'possible_ipados': return { label: 'Possible iPad', className: `${base} ${tone}` };
    case 'confirmed_android': return { label: 'Confirmed Android', className: `${base} ${tone}` };
    case 'probable_android': return { label: 'Probable Android', className: `${base} ${tone}` };
    case 'possible_android': return { label: 'Possible Android', className: `${base} ${tone}` };
    case 'conflicting_mobile_os_evidence': return { label: 'Conflict', className: `${base} ${tone}` };
    case 'unknown_mobile': return { label: 'Unknown mobile', className: `${base} ${tone}` };
    default: return null;
  }
}

export function TopologyNodeView({ data, selected }: NodeProps) {
  const node = data as unknown as TopologyNode;
  const selectNode = useUIStore((state) => state.selectNode);
  const reach = String((node as { reachability?: string }).reachability ?? '');
  const muted = (reach === 'arp_only' || reach === 'unknown') && !node.isGateway && !node.isAgent;
  const discovery = discoverySourceBadges((node as { discoverySources?: string[] }).discoverySources ?? [], reach);
  const title = nodeDisplayTitle(node);
  const ip = nodeIp(node);
  const hostname = nodeSecondaryHostname(node);
  const online = reach === 'self' || reach === 'reachable';
  const wireless = node.wireless && Object.keys(node.wireless).length > 0;
  const vendor = typeof node.vendor === 'string' ? node.vendor : null;
  const mac = typeof node.mac === 'string' ? node.mac : null;
  const evidenceCount = typeof node.evidenceCount === 'number' ? node.evidenceCount : 0;
  const mobile = node.mobileFingerprint ?? null;
  const badge = mobileBadge(mobile?.classification, mobile?.confidence ?? node.osConfidence ?? 0);

  return (
    <div
      data-topology-node-id={node.id}
      data-topology-node-type={node.type}
      onClick={(event) => {
        event.stopPropagation();
        selectNode(node.id);
      }}
      className={[
        'min-w-[190px] max-w-[226px] rounded-lg border px-3 py-2 shadow-sm backdrop-blur transition-opacity',
        nodeTone(node),
        selected ? 'ring-2 ring-blue-400/80' : '',
        node.certainty !== 'confirmed' ? 'border-dashed' : '',
        muted ? 'opacity-85' : '',
      ].join(' ')}
    >
      <Handle type="target" position={Position.Top} className="!h-2 !w-2 !border-0 !bg-transparent !opacity-0" />
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="truncate text-sm font-semibold leading-5">{title}</div>
          {ip ? (
            <div className="mt-0.5 truncate font-mono text-[12px] text-zinc-300">{ip}</div>
          ) : null}
          {hostname ? (
            <div className="mt-0.5 truncate font-mono text-[10px] text-zinc-500">{hostname}</div>
          ) : null}
          {(mac || vendor) ? (
            <div className="mt-0.5 truncate font-mono text-[9px] text-zinc-500">
              {[mac, vendor].filter(Boolean).join(' · ')}
            </div>
          ) : null}
          {badge ? (
            <div className={badge.className}>
              <span className="truncate">{badge.label}</span>
              <span className="shrink-0 text-zinc-500">{Math.round((mobile?.confidence ?? 0) * 100)}%</span>
            </div>
          ) : null}
        </div>
        {node.badge ? (
          <span className="rounded border border-zinc-600/70 bg-zinc-950/50 px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wide text-zinc-300">
            {node.badge}
          </span>
        ) : null}
      </div>
      <div className="mt-2 flex items-center justify-between gap-2">
        <span className="flex min-w-0 flex-wrap items-center gap-1">
          <span className={['rounded px-1 py-0.5 font-mono text-[9px] uppercase tracking-wide', online ? 'bg-emerald-500/15 text-emerald-300' : 'bg-zinc-800 text-zinc-400'].join(' ')}>
            {online ? 'online' : 'seen'}
          </span>
          {wireless ? (
            <span className="rounded border border-sky-500/40 bg-sky-500/10 px-1 py-0.5 font-mono text-[9px] uppercase tracking-wide text-sky-300">
              wifi
            </span>
          ) : null}
          {discovery.length ? (
            discovery.slice(0, 1).map((item) => (
              <span key={item} className="rounded border border-zinc-700 bg-zinc-950/35 px-1 py-0.5 font-mono text-[9px] uppercase tracking-wide text-zinc-400">
                {item}
              </span>
            ))
          ) : (
            <span className="truncate font-mono text-[10px] uppercase tracking-wide text-zinc-500">
              {node.roles.length ? node.roles.join(' ') : node.type}
            </span>
          )}
        </span>
        <span className="shrink-0 font-mono text-[10px] text-zinc-400">{Math.round(node.confidence * 100)}%</span>
      </div>
      <div className="mt-1 flex items-center justify-between gap-2 font-mono text-[9px] uppercase tracking-wide text-zinc-500">
        <span className="truncate">{node.role ?? node.roles[0] ?? node.type}</span>
        <span>{evidenceCount} ev</span>
      </div>
      <Handle type="source" position={Position.Bottom} className="!h-2 !w-2 !border-0 !bg-transparent !opacity-0" />
    </div>
  );
}
