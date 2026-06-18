import { BaseEdge, EdgeLabelRenderer, getBezierPath, type EdgeProps } from '@xyflow/react';
import type { TopologyEdge } from '../../lib/models';
import { formatTopologyEdgeLabel } from '../../lib/topology-display';

const tierColor: Record<string, string> = {
  l2: '#94a3b8',
  l3: '#a1a1aa',
  nat: '#f59e0b',
  isp: '#38bdf8',
};

export function TopologyEdgeView({ id, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, data, selected }: EdgeProps) {
  const edge = data as unknown as TopologyEdge;
  const showLabel = (edge as { showLabel?: boolean }).showLabel !== false;
  const [path, labelX, labelY] = getBezierPath({ sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition });
  const stroke = selected ? '#60a5fa' : tierColor[edge.tier] ?? '#a1a1aa';
  const dash = edge.physical ? undefined : '2 7';
  const width = edge.physical ? 2.2 : edge.thin ? 1.1 : 1.7;

  return (
    <>
      <BaseEdge
        id={id}
        path={path}
        style={{
          stroke,
          strokeWidth: width,
          strokeDasharray: dash,
          strokeLinecap: 'round',
          opacity: edge.confidence < 0.4 ? 0.62 : 0.95,
        }}
      />
      {showLabel && (edge.boundary || edge.label) && (
        <EdgeLabelRenderer>
          <div
            className="pointer-events-none absolute max-w-[220px] rounded border border-zinc-700/80 bg-zinc-950/92 px-2 py-0.5 text-center font-mono text-[10px] uppercase tracking-wide text-zinc-300 shadow-sm shadow-black/30"
            style={{ transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)` }}
          >
            {formatTopologyEdgeLabel(edge)}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
}
