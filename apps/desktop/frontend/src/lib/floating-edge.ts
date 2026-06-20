/**
 * floating-edge.ts — geometry for "floating" topology links.
 *
 * Instead of pinning every edge to a fixed Top (target) / Bottom (source)
 * handle, we anchor each end at the point where the straight line between the
 * two node centers crosses that node's border. This keeps links clean no matter
 * where the nodes sit relative to each other — above, below, or beside — so
 * uplinks to a gateway no longer leave the bottom of a node and loop all the way
 * back up. Adapted from the official React Flow "floating edges" example.
 */
import { Position, type InternalNode, type Node } from '@xyflow/react';

function center(node: InternalNode<Node>) {
  const pos = node.internals.positionAbsolute;
  const w = node.measured?.width ?? 0;
  const h = node.measured?.height ?? 0;
  return { x: pos.x + w / 2, y: pos.y + h / 2, w, h };
}

/** Point where the line from `node`'s center toward `other`'s center exits `node`. */
function intersection(node: InternalNode<Node>, other: InternalNode<Node>) {
  const { x: x2, y: y2, w, h } = center(node);
  const { x: x1, y: y1 } = center(other);
  const hw = w / 2;
  const hh = h / 2;
  if (hw === 0 || hh === 0) return { x: x2, y: y2 };

  const a = (x1 - x2) / (2 * hw) - (y1 - y2) / (2 * hh);
  const b = (x1 - x2) / (2 * hw) + (y1 - y2) / (2 * hh);
  const s = 1 / (Math.abs(a) + Math.abs(b) || 1);
  return { x: hw * (s * a + s * b) + x2, y: hh * (-s * a + s * b) + y2 };
}

/** Which side of `node` the intersection point lands on. */
function side(node: InternalNode<Node>, point: { x: number; y: number }): Position {
  const pos = node.internals.positionAbsolute;
  const w = node.measured?.width ?? 0;
  const h = node.measured?.height ?? 0;
  const nx = Math.round(pos.x);
  const ny = Math.round(pos.y);
  const px = Math.round(point.x);
  const py = Math.round(point.y);
  if (px <= nx + 1) return Position.Left;
  if (px >= nx + w - 1) return Position.Right;
  if (py <= ny + 1) return Position.Top;
  if (py >= ny + h - 1) return Position.Bottom;
  return Position.Top;
}

export function getFloatingEdgeParams(source: InternalNode<Node>, target: InternalNode<Node>) {
  const sourcePoint = intersection(source, target);
  const targetPoint = intersection(target, source);
  return {
    sx: sourcePoint.x,
    sy: sourcePoint.y,
    tx: targetPoint.x,
    ty: targetPoint.y,
    sourcePos: side(source, sourcePoint),
    targetPos: side(target, targetPoint),
  };
}

/** True once React Flow has measured the node, so its size/position is usable. */
export function isMeasured(node: InternalNode<Node> | null | undefined): node is InternalNode<Node> {
  return !!node && !!node.measured?.width && !!node.measured?.height;
}
