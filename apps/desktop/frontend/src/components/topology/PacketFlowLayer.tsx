import { memo } from 'react';
import type { PacketVisual } from '../../lib/packet-flow';

interface Props {
  /** The edge's SVG path `d` — the exact curve packets travel along. */
  path: string;
  visual: PacketVisual;
  edgeId: string;
}

/**
 * Renders the moving packet particles for a single edge.
 *
 * Motion is done with native SVG <animateMotion> (SMIL) pointed at the edge's
 * own path string. This runs entirely in the browser's animation engine — there
 * are NO per-frame React renders and no requestAnimationFrame loop, so the cost
 * is independent of React reconciliation. The component is memoized, so it only
 * re-renders when the path (node moved) or the visual params (settings/selection)
 * actually change. Idle edges render this component not at all.
 *
 * Each packet is a bright core plus a larger, faint halo (a cheap glow that
 * avoids SVG filters). Particles are phase-staggered via negative `begin` so the
 * stream looks continuous from the first frame.
 */
function PacketFlowLayerImpl({ path, visual, edgeId }: Props) {
  const { count, durationMs, color, glow, reverse, bidirectional, radius, opacity } = visual;
  const dur = `${durationMs}ms`;

  const forwardCount = bidirectional ? Math.ceil(count / 2) : count;
  const reverseCount = bidirectional ? Math.floor(count / 2) : 0;

  const packet = (i: number, total: number, rev: boolean, tag: string) => {
    // Negative begin offsets distribute the packets evenly along the path
    // immediately, instead of all bunching at the start.
    const begin = `${-(durationMs * i) / Math.max(total, 1)}ms`;
    return (
      <g key={`${edgeId}-${tag}-${i}`} opacity={opacity}>
        <circle r={radius * 2.4} fill={glow} opacity={0.28} />
        <circle r={radius} fill={color} />
        <circle r={radius * 0.45} fill="#ffffff" opacity={0.85} />
        <animateMotion
          dur={dur}
          begin={begin}
          repeatCount="indefinite"
          path={path}
          // Reverse travels target→source by walking the path backwards.
          keyPoints={rev ? '1;0' : undefined}
          keyTimes={rev ? '0;1' : undefined}
          calcMode={rev ? 'linear' : undefined}
        />
      </g>
    );
  };

  const packets = [];
  for (let i = 0; i < forwardCount; i++) packets.push(packet(i, forwardCount, reverse, 'f'));
  for (let i = 0; i < reverseCount; i++) packets.push(packet(i, reverseCount, true, 'r'));

  return <g style={{ pointerEvents: 'none' }}>{packets}</g>;
}

function areEqual(prev: Props, next: Props): boolean {
  return (
    prev.path === next.path &&
    prev.edgeId === next.edgeId &&
    prev.visual.count === next.visual.count &&
    prev.visual.durationMs === next.visual.durationMs &&
    prev.visual.color === next.visual.color &&
    prev.visual.reverse === next.visual.reverse &&
    prev.visual.bidirectional === next.visual.bidirectional &&
    prev.visual.radius === next.visual.radius &&
    prev.visual.opacity === next.visual.opacity
  );
}

export const PacketFlowLayer = memo(PacketFlowLayerImpl, areEqual);
