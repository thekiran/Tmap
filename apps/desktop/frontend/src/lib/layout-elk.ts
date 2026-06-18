/**
 * layout-elk.ts — optional automatic layout via ELK.js.
 *
 * The topology view model already ships heuristic positions so the canvas renders
 * instantly without async work. When the user picks the "elk_layered" engine we
 * refine positions with ELK's layered algorithm; "force" uses ELK's force model.
 * "manual" skips ELK entirely and uses stored/heuristic positions.
 *
 * ELK is loaded lazily so it never blocks first paint.
 */
import type { TopologyViewModel, LayoutPosition } from './models';
import type { LayoutEngine } from './models';

const NODE_W = 210;
const NODE_H = 78;

let elkInstance: unknown = null;
async function getElk(): Promise<{ layout: (g: unknown) => Promise<unknown> }> {
  if (!elkInstance) {
    const mod = await import('elkjs/lib/elk.bundled.js');
    const ELK = (mod.default ?? mod) as new () => { layout: (g: unknown) => Promise<unknown> };
    elkInstance = new ELK();
  }
  return elkInstance as { layout: (g: unknown) => Promise<unknown> };
}

function layoutLanHub(vm: TopologyViewModel): Record<string, LayoutPosition> | null {
  const gateway = vm.nodes.find((node) => node.isGateway);
  if (!gateway || vm.nodes.length > 14) return null;

  const agent = vm.nodes.find((node) => node.isAgent);
  const others = vm.nodes
    .filter((node) => node.id !== gateway.id && node.id !== agent?.id)
    .sort((a, b) => {
      if (a.isUnknown !== b.isUnknown) return a.isUnknown ? -1 : 1;
      return a.id.localeCompare(b.id);
    });
  const positions: Record<string, LayoutPosition> = {
    [gateway.id]: { x: 390, y: 42 },
  };

  if (agent) positions[agent.id] = { x: 150, y: 225 };

  const slots = [
    { x: 390, y: 225 },
    { x: 630, y: 225 },
    { x: 270, y: 380 },
    { x: 510, y: 380 },
    { x: 750, y: 380 },
    { x: 30, y: 380 },
  ];

  others.forEach((node, index) => {
    const slot = slots[index];
    positions[node.id] = slot ?? {
      x: 90 + (index % 4) * 220,
      y: 540 + Math.floor(index / 4) * 140,
    };
  });

  return positions;
}

export async function layoutWithElk(
  vm: TopologyViewModel,
  engine: LayoutEngine,
): Promise<Record<string, LayoutPosition>> {
  if (engine === 'manual') {
    return Object.fromEntries(vm.nodes.map((n) => [n.id, n.position]));
  }

  const lanHubLayout = layoutLanHub(vm);
  if (lanHubLayout) return lanHubLayout;

  const elk = await getElk();
  const algorithm = engine === 'force' ? 'force' : 'layered';
  const graph = {
    id: 'root',
    layoutOptions: {
      'elk.algorithm': algorithm,
      'elk.direction': 'DOWN',
      'elk.layered.spacing.nodeNodeBetweenLayers': '70',
      'elk.spacing.nodeNode': '40',
    },
    children: vm.nodes.map((n) => ({ id: n.id, width: NODE_W, height: NODE_H })),
    edges: vm.edges.map((e) => ({ id: e.id, sources: [e.source], targets: [e.target] })),
  };

  try {
    const res = (await elk.layout(graph)) as { children?: { id: string; x: number; y: number }[] };
    const out: Record<string, LayoutPosition> = {};
    for (const c of res.children ?? []) out[c.id] = { x: c.x, y: c.y };
    return out;
  } catch {
    // ELK failure must never break the canvas — fall back to heuristic positions.
    return Object.fromEntries(vm.nodes.map((n) => [n.id, n.position]));
  }
}
