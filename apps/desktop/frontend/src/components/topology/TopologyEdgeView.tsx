import { useMemo, useState } from 'react';
import {
  BaseEdge,
  EdgeLabelRenderer,
  getBezierPath,
  useInternalNode,
  type EdgeProps,
} from '@xyflow/react';
import type { TopologyEdge, TopologyNode } from '../../lib/models';
import { formatTopologyEdgeLabel } from '../../lib/topology-display';
import { getFloatingEdgeParams, isMeasured } from '../../lib/floating-edge';
import { derivePacketFlow, packetVisual } from '../../lib/packet-flow';
import { usePacketAnimation } from './packet-animation-context';
import { PacketFlowLayer } from './PacketFlowLayer';

const tierColor: Record<string, string> = {
  l2: '#94a3b8',
  l3: '#a1a1aa',
  nat: '#f59e0b',
  isp: '#38bdf8',
};

function nodeLabel(node: ReturnType<typeof useInternalNode>): string | null {
  const data = node?.data as TopologyNode | undefined;
  return data?.label ?? null;
}

export function TopologyEdgeView({
  id,
  source,
  target,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  data,
  selected,
}: EdgeProps) {
  const edge = data as unknown as TopologyEdge;
  const showLabel = (edge as { showLabel?: boolean }).showLabel !== false;
  const sourceNode = useInternalNode(source);
  const targetNode = useInternalNode(target);
  const anim = usePacketAnimation();
  const [hovered, setHovered] = useState(false);

  // Float the endpoints onto each node's border so links route along the line
  // between centers; fall back to the handle anchors until nodes are measured.
  let geom = { sx: sourceX, sy: sourceY, tx: targetX, ty: targetY, sourcePos: sourcePosition, targetPos: targetPosition };
  if (isMeasured(sourceNode) && isMeasured(targetNode)) {
    geom = getFloatingEdgeParams(sourceNode, targetNode);
  }
  const [path, labelX, labelY] = getBezierPath({
    sourceX: geom.sx,
    sourceY: geom.sy,
    targetX: geom.tx,
    targetY: geom.ty,
    sourcePosition: geom.sourcePos,
    targetPosition: geom.targetPos,
  });
  const stroke = selected ? '#60a5fa' : tierColor[edge.tier] ?? '#a1a1aa';
  const dash = edge.lineStyle === 'dotted' ? '1 8' : edge.lineStyle === 'dashed' ? '7 7' : undefined;
  const width = edge.lineStyle === 'solid' ? 2.2 : edge.thin ? 1.1 : 1.7;

  // Derive the packet-flow descriptor and the concrete render params. Memoized
  // so this only recomputes when the edge, selection, or global settings change
  // — never per animation frame (the browser drives the motion).
  const flow = useMemo(() => derivePacketFlow(edge), [edge]);
  const visual = useMemo(
    () =>
      anim.enabled
        ? packetVisual(flow, { intensity: anim.intensity, selected: Boolean(selected), maxParticles: anim.maxParticles })
        : null,
    [flow, anim.enabled, anim.intensity, anim.maxParticles, selected],
  );

  const showDetails = hovered || Boolean(selected);

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
          opacity: edge.lineStyle === 'dotted' || edge.confidence < 0.4 ? 0.52 : 0.95,
        }}
      />

      {/* Invisible, wide hit area so edges are easy to hover for details. */}
      <path
        d={path}
        fill="none"
        stroke="transparent"
        strokeWidth={16}
        style={{ pointerEvents: 'stroke', cursor: 'pointer' }}
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
      />

      {visual && <PacketFlowLayer path={path} visual={visual} edgeId={id} />}

      {showDetails ? (
        <EdgeLabelRenderer>
          <EdgeDetails
            edge={edge}
            flow={flow}
            sourceLabel={nodeLabel(sourceNode)}
            targetLabel={nodeLabel(targetNode)}
            x={labelX}
            y={labelY}
          />
        </EdgeLabelRenderer>
      ) : (
        showLabel &&
        (edge.boundary || edge.label) && (
          <EdgeLabelRenderer>
            <div
              className="pointer-events-none absolute max-w-[220px] rounded border border-zinc-700/80 bg-zinc-950/92 px-2 py-0.5 text-center font-mono text-[10px] uppercase tracking-wide text-zinc-300 shadow-sm shadow-black/30"
              style={{ transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)` }}
            >
              {formatTopologyEdgeLabel(edge)}
            </div>
          </EdgeLabelRenderer>
        )
      )}
    </>
  );
}

function EdgeDetails({
  edge,
  flow,
  sourceLabel,
  targetLabel,
  x,
  y,
}: {
  edge: TopologyEdge;
  flow: ReturnType<typeof derivePacketFlow>;
  sourceLabel: string | null;
  targetLabel: string | null;
  x: number;
  y: number;
}) {
  return (
    <div
      className="pointer-events-none absolute w-56 -translate-x-1/2 -translate-y-1/2 rounded-md border border-zinc-700/80 bg-zinc-950/95 px-2.5 py-2 text-[11px] text-zinc-300 shadow-lg shadow-black/40 backdrop-blur"
      style={{ transform: `translate(-50%, -50%) translate(${x}px,${y}px)` }}
    >
      <div className="mb-1 flex items-center gap-1.5 font-medium text-zinc-200">
        <span className="truncate">{sourceLabel ?? edge.source}</span>
        <span className="text-zinc-500">{flow.direction === 'reverse' ? '←' : flow.direction === 'bidirectional' || flow.direction === 'unknown' ? '↔' : '→'}</span>
        <span className="truncate">{targetLabel ?? edge.target}</span>
      </div>
      <dl className="grid grid-cols-[auto_1fr] gap-x-2 gap-y-0.5 font-mono text-[10px] text-zinc-400">
        <Row label="Traffic" value={flow.trafficState} />
        <Row label="Direction" value={flow.direction} />
        <Row label="Protocol" value={flow.protocol.toUpperCase()} />
        <Row label="Confidence" value={flow.confidence} />
        <Row label="Layer" value={formatTopologyEdgeLabel(edge)} />
      </dl>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <>
      <dt className="uppercase tracking-wide text-zinc-600">{label}</dt>
      <dd className="truncate text-zinc-300">{value}</dd>
    </>
  );
}
