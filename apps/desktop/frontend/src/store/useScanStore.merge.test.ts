import { describe, expect, it, beforeEach } from 'vitest';
import { useScanStore } from './useScanStore';
import { buildLargeTopologyFixture } from '../test/topology-fixture';
import type { RawScanReport } from '../lib/scan-schema';

function nodeCount(): number {
  return useScanStore.getState().normalized?.topology.nodes.length ?? 0;
}
function nodeIds(): Set<string> {
  return new Set(useScanStore.getState().normalized?.topology.nodes.map((n) => n.id) ?? []);
}

describe('useScanStore live merge', () => {
  beforeEach(() => {
    useScanStore.getState().clearReport();
  });

  it('de-duplicates by stable id when the same report is merged twice', () => {
    const { report } = buildLargeTopologyFixture();
    useScanStore.getState().setReport(report);
    const baseline = nodeCount();
    expect(baseline).toBeGreaterThan(0);

    useScanStore.getState().mergeReport(report);

    // Merging identical data must not duplicate nodes — id is the stable key.
    expect(nodeCount()).toBe(baseline);
  });

  it('unions newly discovered devices into the existing map without wiping it', () => {
    const { report } = buildLargeTopologyFixture();
    useScanStore.getState().setReport(report);
    const baselineIds = nodeIds();
    const baseline = baselineIds.size;

    const clone = JSON.parse(JSON.stringify(report)) as RawScanReport;
    const topo = clone.topology as unknown as { nodes: Array<Record<string, unknown>> };
    topo.nodes.push({
      id: 'merge-test-extra',
      label: 'New Device',
      type: 'wired_client',
      device_role: 'client',
      ip_addresses: ['10.99.99.99'],
      mac_addresses: ['02:00:00:99:99:99'],
      confidence: 0.6,
    });

    useScanStore.getState().mergeReport(clone);

    const after = nodeIds();
    expect(after.size).toBe(baseline + 1);
    expect(after.has('merge-test-extra')).toBe(true);
    // Every previously discovered node is still on the map (never cleared).
    for (const id of baselineIds) expect(after.has(id)).toBe(true);
    expect(useScanStore.getState().lastUpdatedAt).not.toBeNull();
  });

  it('buffers updates when live mode is paused and applies them on demand', () => {
    const { report } = buildLargeTopologyFixture();
    useScanStore.getState().setLiveUpdate(false);
    expect(useScanStore.getState().normalized).toBeNull();

    // Simulate the controller buffering an arriving report while live is off.
    useScanStore.getState().bufferReport(report);
    expect(useScanStore.getState().normalized).toBeNull();
    expect(useScanStore.getState().pendingReport).not.toBeNull();

    useScanStore.getState().applyPending();
    expect(useScanStore.getState().pendingReport).toBeNull();
    expect(nodeCount()).toBeGreaterThan(0);

    useScanStore.getState().setLiveUpdate(true);
  });
});
