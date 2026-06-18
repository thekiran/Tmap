import { Handle, Position, type NodeProps } from '@xyflow/react';
import type { TopologyNode } from '../../lib/models';
import { discoverySourceBadges, nodeDisplayTitle, nodeIp, nodeSecondaryHostname } from '../../lib/topology-display';

function nodeTone(node: TopologyNode) {
  if (node.isGateway) return 'border-blue-400/70 bg-blue-500/12 text-blue-50 shadow-blue-950/20';
  if (node.isAgent) return 'border-emerald-400/70 bg-emerald-500/14 text-emerald-50 shadow-emerald-950/20 ring-1 ring-emerald-400/25';
  if (node.isUnknown) return 'border-zinc-600/80 bg-zinc-900/94 text-zinc-100 shadow-black/20';
  return 'border-zinc-700/80 bg-zinc-900/92 text-zinc-100 shadow-black/20';
}

export function TopologyNodeView({ data, selected }: NodeProps) {
  const node = data as unknown as TopologyNode;
  const reach = String((node as { reachability?: string }).reachability ?? '');
  const muted = (reach === 'arp_only' || reach === 'unknown') && !node.isGateway && !node.isAgent;
  const discovery = discoverySourceBadges((node as { discoverySources?: string[] }).discoverySources ?? [], reach);
  const title = nodeDisplayTitle(node);
  const ip = nodeIp(node);
  const hostname = nodeSecondaryHostname(node);

  return (
    <div
      className={[
        'min-w-[190px] max-w-[226px] rounded-lg border px-3 py-2 shadow-sm backdrop-blur transition-opacity',
        nodeTone(node),
        selected ? 'ring-2 ring-blue-400/80' : '',
        node.certainty !== 'confirmed' ? 'border-dashed' : '',
        muted ? 'opacity-85' : '',
      ].join(' ')}
    >
      <Handle type="target" position={Position.Top} className="!h-2 !w-2 !border-zinc-500 !bg-zinc-900" />
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="truncate text-sm font-semibold leading-5">{title}</div>
          {ip ? (
            <div className="mt-0.5 truncate font-mono text-[12px] text-zinc-300">{ip}</div>
          ) : null}
          {hostname ? (
            <div className="mt-0.5 truncate font-mono text-[10px] text-zinc-500">{hostname}</div>
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
          {discovery.length ? (
            discovery.slice(0, 2).map((item) => (
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
      <Handle type="source" position={Position.Bottom} className="!h-2 !w-2 !border-zinc-500 !bg-zinc-900" />
    </div>
  );
}
