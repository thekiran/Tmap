import { describe, expect, it } from 'vitest';
import { buildTopologyViewModel } from './build-topology-view-model';
import { deviceIconKey } from './format';
import { normalizeScan } from './normalize-scan';
import { applyTopologyFilters, defaultTopologyFilters } from './topology-filters';
import type { NormalizedScanReport, TopologyViewModel } from './models';
import type { RawScanReport } from './scan-schema';
import { buildInvalidEdgeFixture, buildLargeTopologyFixture } from '../test/topology-fixture';

function normalizeFixture(report = buildLargeTopologyFixture().report): NormalizedScanReport {
  const normalizedBase = normalizeScan(report);
  return { ...normalizedBase, topology: buildTopologyViewModel(normalizedBase) };
}

function backendTopology(report: RawScanReport) {
  return report.topology as unknown as {
    nodes: Array<{ id: string; type?: string; device_role?: string }>;
    edges: Array<{ id: string; source: string; target: string; type?: string }>;
  };
}

describe('topology JSON mapping', () => {
  it('maps every backend topology node into the frontend view model', () => {
    const fixture = buildLargeTopologyFixture();
    const normalized = normalizeFixture(fixture.report);
    const backend = backendTopology(fixture.report);

    expect(normalized.topology.nodes).toHaveLength(backend.nodes.length);
    expect(new Set(normalized.topology.nodes.map((node) => node.id))).toEqual(new Set(backend.nodes.map((node) => node.id)));
  });

  it('keeps frontend node IDs unique and stable from backend IDs', () => {
    const fixture = buildLargeTopologyFixture();
    const normalizedA = normalizeFixture(fixture.report);
    const normalizedB = normalizeFixture(JSON.parse(JSON.stringify(fixture.report)));
    const idsA = normalizedA.topology.nodes.map((node) => node.id);
    const idsB = normalizedB.topology.nodes.map((node) => node.id);

    expect(idsA).toEqual(idsB);
    expect(new Set(idsA).size).toBe(idsA.length);
    expect(idsA.every((id) => !/^node-\d+$|^n-\d+$/.test(id))).toBe(true);
  });

  it('falls back unknown backend types to unknown nodes and icons', () => {
    const normalized = normalizeFixture();
    const unknowns = normalized.topology.nodes.filter((node) => node.type === 'unknown');

    expect(unknowns.length).toBeGreaterThanOrEqual(5);
    expect(deviceIconKey('mystery_sensor')).toBe('unknown');
    expect(unknowns.every((node) => deviceIconKey(String(node.type)) === 'unknown')).toBe(true);
  });

  it('validates edge source/target and drops edges with missing endpoint nodes', () => {
    const report = buildInvalidEdgeFixture();
    const normalized = normalizeFixture(report);
    const backend = backendTopology(report);
    const backendNodeIds = new Set(backend.nodes.map((node) => node.id));
    const validBackendEdges = backend.edges.filter((edge) => backendNodeIds.has(edge.source) && backendNodeIds.has(edge.target));

    expect(normalized.topology.edges).toHaveLength(validBackendEdges.length);
    expect(normalized.topology.edges.some((edge) => edge.id === 'edge-invalid-missing-target')).toBe(false);
    expect(normalized.topology.nodes.some((node) => node.id === 'dev-does-not-exist')).toBe(false);
  });

  it('default filter state does not hide devices or edges', () => {
    const normalized = normalizeFixture();
    const filtered = applyTopologyFilters(normalized.topology, defaultTopologyFilters) as TopologyViewModel;

    expect(filtered.nodes).toHaveLength(normalized.topology.nodes.length);
    expect(filtered.edges).toHaveLength(normalized.topology.edges.length);
  });

  it('frontend node count equals backend topology node count', () => {
    const fixture = buildLargeTopologyFixture();
    const normalized = normalizeFixture(fixture.report);

    expect(normalized.topology.nodes.length).toBe(backendTopology(fixture.report).nodes.length);
  });

  it('frontend edge count equals valid backend topology edge count', () => {
    const fixture = buildLargeTopologyFixture();
    const normalized = normalizeFixture(fixture.report);
    const backend = backendTopology(fixture.report);
    const backendNodeIds = new Set(backend.nodes.map((node) => node.id));
    const validBackendEdges = backend.edges.filter((edge) => backendNodeIds.has(edge.source) && backendNodeIds.has(edge.target));

    expect(normalized.topology.edges.length).toBe(validBackendEdges.length);
  });

  it('supports every required node type with expected fallback component type', () => {
    const normalized = normalizeFixture();
    const types = normalized.topology.nodes.map((node) => node.type);

    expect(types).toEqual(expect.arrayContaining([
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
    ]));
  });

  it('maps edge evidence strength to solid, dashed, and dotted line styles', () => {
    const normalized = normalizeFixture();
    const byRawType = new Map<string, string>();
    for (const edge of normalized.topology.edges) {
      const raw = edge.rawEdge as { type?: string } | undefined;
      if (raw?.type && !byRawType.has(raw.type)) byRawType.set(raw.type, edge.lineStyle ?? '');
    }

    expect(byRawType.get('reported-by-router')).toBe('solid');
    expect(byRawType.get('reported-by-ap')).toBe('solid');
    expect(byRawType.get('wireless-associated')).toBe('solid');
    expect(byRawType.get('subnet-inferred')).toBe('dashed');
    expect(byRawType.get('wireless-observed')).toBe('dashed');
    expect(byRawType.get('weak-inferred')).toBe('dotted');
  });
});
