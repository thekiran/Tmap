import { expect, test, type Page } from '@playwright/test';
import { buildLargeTopologyFixture } from '../src/test/topology-fixture';

const fixture = buildLargeTopologyFixture();
const backend = fixture.report.topology as unknown as {
  nodes: Array<{ id: string; label: string; type: string; device_role: string; ip_addresses?: string[] }>;
  edges: Array<{ id: string; source: string; target: string; type: string }>;
};
const backendNodeIds = new Set(backend.nodes.map((node) => node.id));
const validBackendEdges = backend.edges.filter((edge) => backendNodeIds.has(edge.source) && backendNodeIds.has(edge.target));
const reportText = JSON.stringify(fixture.report);

async function installMockBackend(page: Page) {
  await page.addInitScript(({ report }) => {
    const win = window as typeof window & {
      __iadE2EMockReady?: boolean;
      go?: {
        main?: {
          App?: {
            ListInterfaces?: () => Promise<string>;
            RunScan?: () => Promise<string>;
            ImportReport?: () => Promise<string>;
            ExportReport?: () => Promise<string>;
            SaveExport?: () => Promise<void>;
          };
        };
      };
      runtime?: {
        EventsOn?: () => void;
        EventsOff?: () => void;
      };
      __iadE2EReport?: string;
      __iadE2ELayout?: string;
    };
    win.__iadE2EReport = report;
    win.__iadE2ELayout =
      localStorage.getItem('iad.topology.layoutPositions.v2') ??
      sessionStorage.getItem('iad.topology.layoutPositions.session.v2') ??
      undefined;

    win.go = {
      main: {
        App: {
          ListInterfaces: async () => JSON.stringify([
            {
              name: 'Ethernet Test',
              up: true,
              loopback: false,
              virtual: false,
              selected: true,
              cidr: '10.10.0.0/16',
              addresses: [{ ip: '10.10.0.10', version: 4, cidr: '10.10.0.0/16' }],
            },
          ]),
          RunScan: async () => report,
          ImportReport: async () => report,
          ExportReport: async () => report,
          SaveExport: async () => undefined,
        },
      },
    };
    win.runtime = {
      EventsOn: () => undefined,
      EventsOff: () => undefined,
    };
    win.__iadE2EMockReady = true;
  }, { report: reportText });
}

async function loadTopology(page: Page) {
  await installMockBackend(page);
  await page.goto('/');
  if (!(await page.locator('.react-flow').isVisible({ timeout: 5_000 }).catch(() => false))) {
    await page.evaluate((report) => {
      (window as typeof window & { __iadE2EReport?: string }).__iadE2EReport = report;
      localStorage.setItem('iad.scan.report.v2', report);
    }, reportText);
    await page.reload();
  }
  await expect(page.locator('.react-flow')).toBeVisible();
  await expect(page.locator('[data-topology-node-id]')).toHaveCount(backend.nodes.length);
}

async function renderedNodeIds(page: Page) {
  return page.locator('[data-topology-node-id]').evaluateAll((nodes) =>
    nodes.map((node) => node.getAttribute('data-topology-node-id')).filter(Boolean) as string[],
  );
}

async function renderedEdgeIds(page: Page) {
  return page.locator('.react-flow__edge').evaluateAll((edges) =>
    edges.map((edge) => edge.getAttribute('data-id') || edge.id).filter(Boolean),
  );
}

test.describe('topology map JSON rendering', () => {
  test('topology page opens and all backend nodes/valid edges are rendered', async ({ page }) => {
    await loadTopology(page);

    const ids = await renderedNodeIds(page);
    expect(ids).toHaveLength(backend.nodes.length);
    expect(new Set(ids).size).toBe(ids.length);
    expect(new Set(ids)).toEqual(new Set(backend.nodes.map((node) => node.id)));

    await expect(page.locator('.react-flow__edge')).toHaveCount(validBackendEdges.length);
    const edgeIds = await renderedEdgeIds(page);
    expect(new Set(edgeIds).size).toBe(edgeIds.length);

    for (const node of backend.nodes) {
      await expect(page.locator(`[data-topology-node-id="${node.id}"]`)).toBeVisible();
      await expect(page.locator(`[data-topology-node-id="${node.id}"]`)).toContainText(node.label);
    }
  });

  test('required node types render through concrete or unknown fallback components', async ({ page }) => {
    await loadTopology(page);

    for (const type of [
      'gateway',
      'router',
      'managed_switch',
      'access_point',
      'mesh_node',
      'repeater',
      'wireless_client',
      'wired_client',
      'server',
      'printer',
      'phone',
      'iot',
      'unknown',
    ]) {
      await expect(page.locator(`[data-topology-node-type="${type}"]`).first()).toBeVisible();
    }
  });

  test('default filters and empty search do not hide devices', async ({ page }) => {
    await loadTopology(page);

    await expect(page.locator('[data-topology-node-id]')).toHaveCount(backend.nodes.length);
    await expect(page.locator('.react-flow__edge')).toHaveCount(validBackendEdges.length);

    await page.getByRole('complementary').getByRole('button', { name: /^Devices$/i }).click();
    await expect(page.locator('tbody tr')).toHaveCount(backend.nodes.length);
    await expect(page.getByText(`${backend.nodes.length} of ${backend.nodes.length}`)).toBeVisible();

    await page.getByRole('complementary').getByRole('button', { name: /^Topology Map$/i }).click();
    await expect(page.locator('[data-topology-node-id]')).toHaveCount(backend.nodes.length);
  });

  test('device type filter hides only expected devices and reset restores all', async ({ page }) => {
    await loadTopology(page);
    const wirelessCount = backend.nodes.filter((node) => node.type === 'wireless_client').length;

    await page.locator('select').first().selectOption('wireless_client');
    await expect(page.locator('[data-topology-node-id]')).toHaveCount(wirelessCount);
    const filteredTypes = await page.locator('[data-topology-node-id]').evaluateAll((nodes) =>
      nodes.map((node) => node.getAttribute('data-topology-node-type')),
    );
    expect(new Set(filteredTypes)).toEqual(new Set(['wireless_client']));

    await page.getByRole('button', { name: /reset filters/i }).click();
    await expect(page.locator('[data-topology-node-id]')).toHaveCount(backend.nodes.length);
  });

  test('fitView places all nodes within the visible React Flow canvas', async ({ page }) => {
    await loadTopology(page);
    await page.getByRole('button', { name: /fit map/i }).click();
    await page.waitForTimeout(400);

    const outside = await page.evaluate(() => {
      const canvas = document.querySelector('.react-flow')?.getBoundingClientRect();
      if (!canvas) return ['missing-canvas'];
      return Array.from(document.querySelectorAll('[data-topology-node-id]')).flatMap((node) => {
        const rect = node.closest('.react-flow__node')?.getBoundingClientRect();
        const id = node.getAttribute('data-topology-node-id') ?? 'unknown';
        if (!rect || rect.width <= 0 || rect.height <= 0) return [`${id}:invalid-position`];
        const visible =
          rect.right >= canvas.left - 2 &&
          rect.left <= canvas.right + 2 &&
          rect.bottom >= canvas.top - 2 &&
          rect.top <= canvas.bottom + 2;
        return visible ? [] : [`${id}:outside-viewport`];
      });
    });

    expect(outside).toEqual([]);
  });

  test('clicking every node opens the matching detail panel data', async ({ page }) => {
    await loadTopology(page);

    for (const node of backend.nodes) {
      await page.locator(`[data-topology-node-id="${node.id}"]`).dispatchEvent('click', { bubbles: true });
      await expect(page.getByText('Device Details')).toBeVisible();
      await expect(page.locator('body')).toContainText(node.label);
      await expect(page.locator('body')).toContainText(node.ip_addresses?.[0] ?? node.label);
    }
  });

  test('refresh does not delete nodes and dragged positions persist', async ({ page }) => {
    await loadTopology(page);
    const logsToggle = page.getByRole('button', { name: /toggle logs/i });
    if (await logsToggle.getAttribute('aria-pressed') === 'true') {
      await logsToggle.click();
    }
    const nodeId = backend.nodes.find((node) => node.type === 'server')?.id ?? backend.nodes[0].id;
    const nodeLocator = page.locator(`.react-flow__node:has([data-topology-node-id="${nodeId}"])`);
    const box = await nodeLocator.boundingBox();
    expect(box).not.toBeNull();

    const canvas = page.locator('.react-flow');
    const canvasBox = await canvas.boundingBox();
    expect(canvasBox).not.toBeNull();
    await nodeLocator.dragTo(canvas, {
      force: true,
      sourcePosition: { x: box!.width / 2, y: box!.height / 2 },
      targetPosition: { x: canvasBox!.width * 0.58, y: canvasBox!.height * 0.52 },
    });
    await page.waitForTimeout(250);
    await expect(page.locator('[data-topology-node-id]')).toHaveCount(backend.nodes.length);
    await expect(page.locator('.react-flow__edge')).toHaveCount(validBackendEdges.length);

    const savedPosition = { x: 1234, y: 567 };
    await page.evaluate(({ id, position }) => {
      const value = JSON.stringify({ [id]: position });
      localStorage.setItem('iad.topology.layoutPositions.v2', value);
      sessionStorage.setItem('iad.topology.layoutPositions.session.v2', value);
    }, { id: nodeId, position: savedPosition });
    const beforeReload = await page.evaluate((id) => {
      const raw = localStorage.getItem('iad.topology.layoutPositions.v2');
      const positions = raw ? JSON.parse(raw) as Record<string, { x: number; y: number }> : {};
      return positions[id] ?? null;
    }, nodeId);
    expect(beforeReload).toEqual(savedPosition);
    await page.reload();
    await expect(page.locator('[data-topology-node-id]')).toHaveCount(backend.nodes.length);
    await expect(page.locator(`.react-flow__node:has([data-topology-node-id="${nodeId}"])`)).toHaveAttribute(
      'style',
      /translate\(1234px, 567px\)/,
    );
  });
});
