import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Background,
  MiniMap,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow,
  type Edge,
  type EdgeMouseHandler,
  type EdgeTypes,
  type Node,
  type NodeChange,
  type NodeMouseHandler,
  type NodeTypes,
  type OnNodeDrag,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useScanStore, isActiveScanStatus } from '../../store/useScanStore';
import { useUIStore } from '../../store/useUIStore';
import { useReducedMotion } from '../../lib/useReducedMotion';
import { layoutWithElk } from '../../lib/layout-elk';
import type { LayoutPosition, NormalizedScanReport, TopologyEdge, TopologyNode, TopologyViewModel } from '../../lib/models';
import { applyTopologyFilters, defaultTopologyFilters, topologyFilterOptions, type TopologyFilterState } from '../../lib/topology-filters';
import { formatTopologyEdgeLabel, nodeDisplayTitle } from '../../lib/topology-display';
import { Icons } from '../icons/Icon';
import { TopologyNodeView } from './TopologyNodeView';
import { TopologyEdgeView } from './TopologyEdgeView';
import { TopologyStatusBar } from './TopologyStatusBar';
import { PacketAnimationContext, type PacketAnimationConfig } from './packet-animation-context';
import type { PacketIntensity } from '../../lib/packet-flow';

const topologyNodeTypeKeys = [
  'iadNode',
  'gateway',
  'router',
  'managed_switch',
  'switch',
  'access_point',
  'mesh_node',
  'repeater',
  'wireless_client',
  'wired_client',
  'server',
  'printer',
  'phone',
  'mobile',
  'iot',
  'unknown',
  'host',
  'workstation',
  'local_host',
];
const nodeTypes: NodeTypes = Object.fromEntries(topologyNodeTypeKeys.map((key) => [key, TopologyNodeView]));
const edgeTypes: EdgeTypes = { iadEdge: TopologyEdgeView };
type InteractionMode = 'move' | 'pan';

function topologyStructureKey(topology: TopologyViewModel | null): string {
  if (!topology) return '';
  const nodes = topology.nodes.map((node) => node.id).sort().join('|');
  const edges = topology.edges.map((edge) => `${edge.id}:${edge.source}>${edge.target}`).sort().join('|');
  return `${nodes}::${edges}`;
}

function TopologyCanvas() {
  const { fitView, zoomIn, zoomOut } = useReactFlow();
  const normalized = useScanStore((state) => state.normalized);
  const topology = normalized?.topology;
  const scanStatus = useScanStore((state) => state.scanStatus);
  const scanError = useScanStore((state) => state.scanError);
  const setScanError = useScanStore((state) => state.setScanError);
  const scanActive = isActiveScanStatus(scanStatus);
  const settings = useUIStore((state) => state.settings);
  const layoutPositions = useUIStore((state) => state.layoutPositions);
  const setNodePosition = useUIStore((state) => state.setNodePosition);
  const resetLayoutPositions = useUIStore((state) => state.resetLayoutPositions);
  const selectNode = useUIStore((state) => state.selectNode);
  const selectEdge = useUIStore((state) => state.selectEdge);
  const selectedNodeId = useUIStore((state) => state.selectedNodeId);
  const selectedEdgeId = useUIStore((state) => state.selectedEdgeId);
  const setPacketAnimation = useUIStore((state) => state.setPacketAnimation);
  const setPacketIntensity = useUIStore((state) => state.setPacketIntensity);
  const reducedMotion = useReducedMotion();
  const [autoPositions, setAutoPositions] = useState<Record<string, LayoutPosition>>({});
  const [interactionMode, setInteractionMode] = useState<InteractionMode>('move');
  const [showLineLabels, setShowLineLabels] = useState(false);
  const [showMiniMap, setShowMiniMap] = useState(true);
  const [filters, setFilters] = useState<TopologyFilterState>(defaultTopologyFilters);

  const filteredTopology = useMemo(() => {
    return applyTopologyFilters(topology, filters);
  }, [filters, topology]);
  const filteredTopologyRef = useRef<TopologyViewModel | null>(null);
  const structureKey = useMemo(() => topologyStructureKey(filteredTopology), [filteredTopology]);

  useEffect(() => {
    filteredTopologyRef.current = filteredTopology;
  }, [filteredTopology]);

  const filterOptions = useMemo(() => {
    return topologyFilterOptions(topology);
  }, [topology]);

  // Auto-layout runs ELK but only ASSIGNS positions to nodes that don't already
  // have one. Existing nodes keep their computed/manual position, so live
  // updates add new devices without shuffling the whole map (requirements 19–21).
  useEffect(() => {
    const currentTopology = filteredTopologyRef.current;
    if (!currentTopology) return;
    let cancelled = false;
    layoutWithElk(currentTopology, settings.layoutEngine).then((positions) => {
      if (cancelled) return;
      setAutoPositions((prev) => {
        let changed = false;
        const next = { ...prev };
        for (const [id, pos] of Object.entries(positions)) {
          if (next[id] == null) {
            next[id] = pos;
            changed = true;
          }
        }
        return changed ? next : prev;
      });
    });
    return () => {
      cancelled = true;
    };
  }, [settings.layoutEngine, structureKey]);

  // Explicit "Auto Layout": recompute the full ELK layout, clear manual drags,
  // and fit once. This is the only path that intentionally moves existing nodes.
  const runAutoLayout = useCallback(() => {
    if (!filteredTopology) return;
    layoutWithElk(filteredTopology, settings.layoutEngine).then((positions) => {
      setAutoPositions(positions);
      resetLayoutPositions();
      window.setTimeout(() => void fitView({ padding: 0.2, duration: 240 }), 40);
    });
  }, [filteredTopology, settings.layoutEngine, fitView, resetLayoutPositions]);

  // Fit the view only on the first populated render and whenever NEW nodes
  // appear (count grows) — never on data-only updates, so zoom/pan stay put
  // during live scanning (requirements 23–25).
  const fittedRef = useRef(false);
  const prevNodeCountRef = useRef(0);
  const visibleNodeCount = filteredTopology?.nodes.length ?? 0;
  useEffect(() => {
    const count = visibleNodeCount;
    if (count === 0 || Object.keys(autoPositions).length === 0) return;
    const grew = count > prevNodeCountRef.current;
    prevNodeCountRef.current = count;
    if (fittedRef.current && !grew) return;
    fittedRef.current = true;
    const timer = window.setTimeout(() => void fitView({ padding: 0.24, duration: 200 }), 50);
    return () => window.clearTimeout(timer);
  }, [autoPositions, fitView, visibleNodeCount]);

  // Nodes derived from the report + persisted/auto layout. React Flow needs to
  // own the live node array during interaction (so drags apply frame-by-frame
  // instead of teleporting on release), so we feed this into useNodesState and
  // re-sync whenever the underlying data changes.
  const desiredNodes: Node<TopologyNode>[] = useMemo(() => {
    if (!filteredTopology) return [];
    return filteredTopology.nodes.map((node: TopologyNode) => ({
      id: node.id,
      type: nodeTypes[String(node.type)] ? String(node.type) : 'unknown',
      data: node,
      position: layoutPositions[node.id] ?? autoPositions[node.id] ?? node.position,
      selected: selectedNodeId === node.id,
      draggable: interactionMode === 'move',
    }));
  }, [autoPositions, interactionMode, layoutPositions, selectedNodeId, filteredTopology]);

  const [nodes, setNodes, onNodesChange] = useNodesState<Node<TopologyNode>>([]);

  const handleNodesChange = useCallback((changes: NodeChange<Node<TopologyNode>>[]) => {
    onNodesChange(changes);
    changes.forEach((change) => {
      if (change.type === 'position' && change.position) {
        setNodePosition(change.id, change.position);
      }
    });
  }, [onNodesChange, setNodePosition]);

  useEffect(() => {
    setNodes(desiredNodes);
  }, [desiredNodes, setNodes]);

  const desiredEdges: Edge<TopologyEdge>[] = useMemo(() => {
    if (!filteredTopology) return [];
    return filteredTopology.edges.map((edge: TopologyEdge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      type: 'iadEdge',
      data: { ...edge, showLabel: showLineLabels },
      selected: selectedEdgeId === edge.id,
    }));
  }, [selectedEdgeId, showLineLabels, filteredTopology]);

  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge<TopologyEdge>>([]);

  useEffect(() => {
    setEdges(desiredEdges);
  }, [desiredEdges, setEdges]);

  const nodeLabels = useMemo(() => {
    const labels = new Map<string, string>();
    filteredTopology?.nodes.forEach((node) => labels.set(node.id, nodeDisplayTitle(node)));
    return labels;
  }, [filteredTopology]);

  const mapDeviceCount = filteredTopology?.nodes.filter((node) => Boolean(node.deviceId)).length ?? 0;
  const linkCount = filteredTopology?.edges.length ?? 0;
  const isEmpty = !filteredTopology || filteredTopology.nodes.length === 0;

  // Packet animation config, shared with every edge via context. Particle count
  // is capped automatically on dense graphs (performance mode) so the SVG cost
  // stays bounded regardless of how many edges are active.
  const packetConfig = useMemo<PacketAnimationConfig>(() => {
    const maxParticles = linkCount > 120 ? 1 : linkCount > 60 ? 2 : 3;
    // Honor the OS reduced-motion preference (SMIL ignores the CSS media query).
    return { enabled: settings.packetAnimation && !reducedMotion, intensity: settings.packetIntensity, maxParticles };
  }, [settings.packetAnimation, settings.packetIntensity, linkCount, reducedMotion]);

  const onNodeClick: NodeMouseHandler = (_, node) => selectNode(node.id);
  const onEdgeClick: EdgeMouseHandler = (_, edge) => selectEdge(edge.id);
  const onNodeDrag: OnNodeDrag = (_, node) => setNodePosition(node.id, node.position);
  const onNodeDragStop: OnNodeDrag = (_, node) => setNodePosition(node.id, node.position);

  // NOTE: the canvas is ALWAYS rendered now — even with zero nodes — so the UI
  // never blocks on a "Scanning…" screen. Empty/scan/error states are drawn as
  // non-blocking overlays on top of the live map (requirements 1–7, 14–15).
  return (
    <PacketAnimationContext.Provider value={packetConfig}>
    <div className="flex h-full min-h-0 flex-col">
      <TopologyStatusBar
        deviceCount={mapDeviceCount}
        linkCount={linkCount}
        onFit={() => void fitView({ padding: 0.2, duration: 240 })}
        onAutoLayout={runAutoLayout}
      />
      {normalized && <TopologyBanner normalized={normalized} />}
      <div className="flex h-11 items-center gap-4 border-b border-zinc-800 bg-zinc-950/80 px-4 text-xs text-zinc-400">
        <span className="font-mono uppercase tracking-[0.2em] text-zinc-500">Topology map</span>
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-2"><i className="h-px w-8 bg-zinc-500" /> physical only if proven</span>
          <span className="inline-flex items-center gap-2"><i className="h-px w-8 border-t border-dotted border-zinc-500" /> inferred</span>
        </div>
        <span className="ml-auto font-mono text-zinc-500">
          {normalized?.discoverySummary
            ? `${mapDeviceCount} devices on map / ${normalized.discoverySummary.devicesFound} LAN discovered / ${normalized.discoverySummary.addressesScanned} addresses scanned`
            : `${filteredTopology?.nodes.length ?? 0} nodes / ${filteredTopology?.edges.length ?? 0} edges`}
        </span>
      </div>
      <TopologyFilters filters={filters} onChange={setFilters} options={filterOptions} />
      <div className="relative min-h-0 flex-1 bg-[radial-gradient(circle_at_1px_1px,rgba(148,163,184,.16)_1px,transparent_0)] [background-size:22px_22px]">
        {/* Empty-state overlay — never blocks the canvas; pointer-events off so
            zoom/pan still work underneath. */}
        {isEmpty && (
          <div className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center p-8 text-center">
            <div className="max-w-sm">
              {scanActive ? (
                <>
                  <span className="mx-auto mb-3 block h-6 w-6 animate-spin rounded-full border-2 border-zinc-500 border-t-transparent" />
                  <div className="text-sm font-semibold text-zinc-300">Waiting for devices…</div>
                  <div className="mt-1 text-[12px] text-zinc-500">
                    The map will fill in automatically as the scan discovers hosts.
                  </div>
                </>
              ) : scanError ? (
                <div className="text-sm text-zinc-400">
                  No devices on the map yet. See the error above, then try a rescan.
                </div>
              ) : (
                <div className="text-sm text-zinc-400">
                  No scan loaded. Click <b className="text-zinc-200">Rescan</b> to map your network, or import a report.
                </div>
              )}
            </div>
          </div>
        )}

        {/* Non-blocking failure banner — the last good topology stays on screen
            (requirement 14). Dismissible. */}
        {scanError && (
          <div className="absolute right-4 top-4 z-30 max-w-md rounded-md border border-red-500/40 bg-red-950/90 px-3 py-2 shadow-lg shadow-black/40 backdrop-blur">
            <div className="flex items-start gap-2">
              <div className="min-w-0">
                <div className="text-[10px] font-semibold uppercase tracking-wide text-red-400">Scan error</div>
                <div className="mt-0.5 break-words font-mono text-[11px] leading-relaxed text-red-200/90">{scanError}</div>
              </div>
              <button
                type="button"
                onClick={() => setScanError(null)}
                aria-label="Dismiss error"
                className="ml-auto shrink-0 rounded px-1 text-red-300 hover:bg-red-500/20"
              >
                ✕
              </button>
            </div>
          </div>
        )}
        <MapToolbar
          interactionMode={interactionMode}
          onInteractionModeChange={setInteractionMode}
          showLineLabels={showLineLabels}
          onShowLineLabelsChange={setShowLineLabels}
          showMiniMap={showMiniMap}
          onShowMiniMapChange={setShowMiniMap}
          onFit={() => void fitView({ padding: 0.18, duration: 260 })}
          onZoomIn={() => void zoomIn({ duration: 160 })}
          onZoomOut={() => void zoomOut({ duration: 160 })}
          onResetLayout={() => {
            resetLayoutPositions();
            window.setTimeout(() => void fitView({ padding: 0.18, duration: 260 }), 0);
          }}
          packetAnimation={settings.packetAnimation}
          packetIntensity={settings.packetIntensity}
          onTogglePackets={() => setPacketAnimation(!settings.packetAnimation)}
          onCyclePackets={() => setPacketIntensity(nextIntensity(settings.packetIntensity))}
        />
        <div className="pointer-events-none absolute bottom-4 left-4 z-10 rounded-md border border-zinc-800 bg-zinc-950/90 px-3 py-2 text-[11px] text-zinc-400 shadow-sm shadow-black/30 backdrop-blur">
          <span className="font-mono uppercase tracking-[0.12em] text-zinc-500">
            {interactionMode === 'move' ? 'Move devices' : 'Pan map'}
          </span>
          <span className="ml-2">
            {interactionMode === 'move'
              ? 'drag devices or empty map; lines stay attached'
              : 'drag empty map space; device positions stay locked'}
          </span>
        </div>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          edgeTypes={edgeTypes}
          onNodesChange={handleNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeClick={onNodeClick}
          onEdgeClick={onEdgeClick}
          onNodeDrag={onNodeDrag}
          onNodeDragStop={onNodeDragStop}
          fitView
          fitViewOptions={{ padding: 0.24 }}
          minZoom={0.22}
          maxZoom={1.8}
          nodesDraggable={interactionMode === 'move'}
          panOnDrag
          panOnScroll
          zoomOnScroll
          nodesConnectable={false}
          elementsSelectable
          proOptions={{ hideAttribution: true }}
        >
          <Background color="rgba(148,163,184,.18)" gap={28} size={1} />
          {showMiniMap ? (
            <MiniMap
              pannable
              zoomable
              nodeColor={(node) => {
                const data = node.data as TopologyNode;
                if (data.isGateway) return '#3b82f6';
                if (data.isAgent) return '#10b981';
                if (data.isUnknown) return '#71717a';
                return '#a1a1aa';
              }}
              maskColor="rgba(0,0,0,.62)"
              style={{ background: '#09090b', border: '1px solid rgba(148,163,184,.2)', borderRadius: 10 }}
            />
          ) : null}
        </ReactFlow>
      </div>
      {filteredTopology && filteredTopology.edges.length > 0 ? (
        <div className="shrink-0 overflow-x-auto border-t border-zinc-800 bg-zinc-950 px-4 py-2 font-mono text-[11px] text-zinc-500">
          <div className="whitespace-nowrap">
            {filteredTopology.edges.map((edge: TopologyEdge) => `${nodeLabels.get(edge.source) ?? edge.source} -> ${nodeLabels.get(edge.target) ?? edge.target}: ${formatTopologyEdgeLabel(edge)}`).join('  |  ')}
          </div>
        </div>
      ) : null}
    </div>
    </PacketAnimationContext.Provider>
  );
}

function MapToolbar({
  interactionMode,
  onInteractionModeChange,
  showLineLabels,
  onShowLineLabelsChange,
  showMiniMap,
  onShowMiniMapChange,
  onFit,
  onZoomIn,
  onZoomOut,
  onResetLayout,
  packetAnimation,
  packetIntensity,
  onTogglePackets,
  onCyclePackets,
}: {
  interactionMode: InteractionMode;
  onInteractionModeChange: (mode: InteractionMode) => void;
  showLineLabels: boolean;
  onShowLineLabelsChange: (show: boolean) => void;
  showMiniMap: boolean;
  onShowMiniMapChange: (show: boolean) => void;
  onFit: () => void;
  onZoomIn: () => void;
  onZoomOut: () => void;
  onResetLayout: () => void;
  packetAnimation: boolean;
  packetIntensity: PacketIntensity;
  onTogglePackets: () => void;
  onCyclePackets: () => void;
}) {
  return (
    <div className="absolute left-4 top-4 z-20 flex max-w-[calc(100%-2rem)] flex-wrap items-center gap-1.5 rounded-md border border-zinc-800 bg-zinc-950/94 p-1.5 shadow-lg shadow-black/35 backdrop-blur">
      <div className="flex overflow-hidden rounded border border-zinc-800 bg-zinc-900/80">
        <ToolbarButton active={interactionMode === 'move'} label="Move devices" onClick={() => onInteractionModeChange('move')}>
          <Icons.layers size={14} />
        </ToolbarButton>
        <ToolbarButton active={interactionMode === 'pan'} label="Pan map" onClick={() => onInteractionModeChange('pan')}>
          <Icons.fit size={14} />
        </ToolbarButton>
      </div>

      <div className="h-6 w-px bg-zinc-800" />

      <ToolbarButton label="Fit map" onClick={onFit}>
        <Icons.fit size={14} />
      </ToolbarButton>
      <ToolbarButton label="Zoom in" onClick={onZoomIn}>
        <Icons.zoomIn size={14} />
      </ToolbarButton>
      <ToolbarButton label="Zoom out" onClick={onZoomOut}>
        <Icons.zoomOut size={14} />
      </ToolbarButton>
      <ToolbarButton label="Reset layout" onClick={onResetLayout}>
        <Icons.reset size={14} />
      </ToolbarButton>

      <div className="h-6 w-px bg-zinc-800" />

      <ToolbarButton active={showLineLabels} label="Line labels" onClick={() => onShowLineLabelsChange(!showLineLabels)}>
        <Icons.info size={14} />
      </ToolbarButton>
      <ToolbarButton active={showMiniMap} label="Mini map" onClick={() => onShowMiniMapChange(!showMiniMap)}>
        <Icons.topology size={14} />
      </ToolbarButton>

      <div className="h-6 w-px bg-zinc-800" />

      {/* Packet flow animation: on/off + Low/Normal/High intensity. */}
      <ToolbarButton
        active={packetAnimation}
        label={packetAnimation ? 'Packet animation: on' : 'Packet animation: off'}
        onClick={onTogglePackets}
      >
        <span className={`inline-block h-2 w-2 rounded-full ${packetAnimation ? 'bg-cyan-400 shadow-[0_0_6px_1px] shadow-cyan-400/70' : 'bg-zinc-600'}`} />
        <span className="ml-1 hidden sm:inline">Packets</span>
      </ToolbarButton>
      <ToolbarButton
        disabled={!packetAnimation}
        label={`Animation intensity: ${packetIntensity}`}
        onClick={onCyclePackets}
      >
        <span className="font-mono uppercase">{packetIntensity === 'low' ? 'L' : packetIntensity === 'high' ? 'H' : 'N'}</span>
      </ToolbarButton>
    </div>
  );
}

/** Cycle packet animation intensity: low → normal → high → low. */
function nextIntensity(current: PacketIntensity): PacketIntensity {
  return current === 'low' ? 'normal' : current === 'normal' ? 'high' : 'low';
}

function ToolbarButton({
  active,
  disabled,
  label,
  onClick,
  children,
}: {
  active?: boolean;
  disabled?: boolean;
  label: string;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      title={label}
      aria-label={label}
      aria-pressed={active || undefined}
      disabled={disabled}
      onClick={onClick}
      className={[
        'inline-flex h-7 items-center gap-1.5 rounded px-2 text-[11px] font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-40',
        active
          ? 'bg-blue-500 text-white shadow-sm shadow-blue-950/30'
          : 'text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100',
      ].join(' ')}
    >
      {children}
      <span className="sr-only">{label}</span>
    </button>
  );
}

function TopologyFilters({
  filters,
  onChange,
  options,
}: {
  filters: TopologyFilterState;
  onChange: (filters: TopologyFilterState) => void;
  options: { types: string[]; evidence: string[] };
}) {
  const update = <K extends keyof TopologyFilterState>(key: K, value: TopologyFilterState[K]) => onChange({ ...filters, [key]: value });
  const mobileOptions = [
    ['all', 'All OS'],
    ['ios', 'iPhone / iOS'],
    ['ipados', 'iPad / iPadOS'],
    ['android', 'Android'],
    ['unknown_mobile', 'Unknown mobile'],
    ['unknown_device', 'Unknown device'],
    ['conflict', 'Conflict'],
  ] as const;
  return (
    <div className="flex min-h-10 flex-wrap items-center gap-2 border-b border-zinc-800 bg-zinc-950/88 px-4 py-2 text-[11px] text-zinc-400">
      <span className="font-mono uppercase tracking-[0.16em] text-zinc-500">Filters</span>
      <select className="h-7 rounded border border-zinc-800 bg-zinc-900 px-2 text-zinc-300" value={filters.deviceType} onChange={(event) => update('deviceType', event.target.value)}>
        <option value="all">All types</option>
        {options.types.map((type) => <option key={type} value={type}>{type.replace(/_/g, ' ')}</option>)}
      </select>
      <select className="h-7 rounded border border-zinc-800 bg-zinc-900 px-2 text-zinc-300" value={filters.online} onChange={(event) => update('online', event.target.value)}>
        <option value="all">Any status</option>
        <option value="online">Online</option>
        <option value="offline">Offline/seen</option>
      </select>
      <select className="h-7 rounded border border-zinc-800 bg-zinc-900 px-2 text-zinc-300" value={filters.evidenceSource} onChange={(event) => update('evidenceSource', event.target.value)}>
        <option value="all">Any evidence</option>
        {options.evidence.map((source) => <option key={source} value={source}>{source}</option>)}
      </select>
      <div className="flex max-w-full flex-wrap items-center gap-1">
        {mobileOptions.map(([value, label]) => {
          const active = filters.mobileOS === value;
          const conflict = value === 'conflict';
          return (
            <button
              key={value}
              type="button"
              onClick={() => update('mobileOS', value)}
              className={[
                'h-7 rounded border px-2 font-mono text-[10px] uppercase tracking-wide transition-colors',
                active
                  ? conflict
                    ? 'border-amber-400 bg-amber-500/15 text-amber-200'
                    : 'border-sky-400 bg-sky-500/15 text-sky-200'
                  : 'border-zinc-800 bg-zinc-900 text-zinc-500 hover:border-zinc-600 hover:text-zinc-200',
              ].join(' ')}
            >
              {label}
            </button>
          );
        })}
      </div>
      <label className="inline-flex items-center gap-2">
        <span className="font-mono text-[10px] uppercase tracking-wide text-zinc-500">Confidence</span>
        <input
          type="range"
          min={0}
          max={100}
          step={5}
          value={Math.round(filters.minConfidence * 100)}
          onChange={(event) => update('minConfidence', Number(event.target.value) / 100)}
          className="w-24"
        />
        <span className="w-8 text-right font-mono">{Math.round(filters.minConfidence * 100)}%</span>
      </label>
      <input
        className="h-7 w-32 rounded border border-zinc-800 bg-zinc-900 px-2 font-mono text-zinc-300 placeholder:text-zinc-600"
        placeholder="SSID/BSSID"
        value={filters.wireless}
        onChange={(event) => update('wireless', event.target.value)}
      />
      <input
        className="h-7 w-28 rounded border border-zinc-800 bg-zinc-900 px-2 font-mono text-zinc-300 placeholder:text-zinc-600"
        placeholder="Subnet prefix"
        value={filters.subnet}
        onChange={(event) => update('subnet', event.target.value)}
      />
      <button
        type="button"
        className="ml-auto h-7 rounded border border-zinc-800 px-2 font-mono text-[10px] uppercase tracking-wide text-zinc-400 hover:border-blue-400 hover:text-blue-300"
        onClick={() => onChange(defaultTopologyFilters)}
      >
        Reset filters
      </button>
    </div>
  );
}

function TopologyBanner({ normalized }: { normalized: NormalizedScanReport }) {
  const [open, setOpen] = useState(false);
  const wanKnown = Boolean(normalized.primaryType);
  const ds = normalized.discoverySummary;
  const noteCount = normalized.warnings.length;

  return (
    <div className="border-b border-zinc-800 bg-zinc-950/70">
      {/* One compact info banner above the map. */}
      <div className="flex items-center gap-3 px-4 py-1.5 text-[11px]">
        <span className="text-sky-300">
          Topology is inferred unless physical evidence such as LLDP/CDP/SNMP is available.
        </span>
        <button
          onClick={() => setOpen((v) => !v)}
          className="ml-auto rounded border border-zinc-700 px-2 py-0.5 font-mono text-[10px] uppercase tracking-wide text-zinc-400 hover:border-blue-400 hover:text-blue-300"
        >
          Evidence &amp; Safety Notes{noteCount ? ` (${noteCount})` : ''} {open ? 'Hide' : 'Show'}
        </button>
      </div>

      {open && (
        <div className="flex flex-col gap-2 border-t border-zinc-800 px-4 py-3 text-xs">
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1">
            <span className="font-mono uppercase tracking-wider text-zinc-500">WAN access type:</span>
            {wanKnown ? (
              <span className="font-semibold text-zinc-200">
                {normalized.primaryType}
                {normalized.category ? <span className="ml-1 text-zinc-500">({normalized.category})</span> : null}
              </span>
            ) : (
              <span className="font-semibold text-amber-400">
                Undetermined
                <span className="ml-2 font-normal text-zinc-500">
                  - no direct physical CPE/ONT/DSL/DOCSIS/cellular evidence.
                </span>
              </span>
            )}
            <span className="ml-auto font-mono text-[11px] text-zinc-400">
              confidence {Math.round((normalized.confidence ?? 0) * 100)}% - {normalized.decisionQuality}
            </span>
          </div>

          {ds && (
            <div className="font-mono text-[11px] text-zinc-500">
              Discovery: {ds.devicesFound} found across {ds.addressesScanned} addresses - ARP {ds.arpFound} - TCP{' '}
              {ds.tcpFound}
              {ds.nmapFound ? ` - Nmap ${ds.nmapFound}` : ''} - {(ds.scanDurationMs / 1000).toFixed(1)}s
            </div>
          )}

          {normalized.warnings.map((w, i) => (
            <div
              key={i}
              className={[
                'rounded-md px-3 py-1.5 text-[11px]',
                w.level === 'danger'
                  ? 'border border-red-500/30 bg-red-500/10 text-red-300'
                  : w.level === 'info'
                    ? 'border border-zinc-700 bg-zinc-500/5 text-zinc-300'
                    : 'border border-amber-500/30 bg-amber-500/10 text-amber-300',
              ].join(' ')}
            >
              {w.text}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export function TopologyScreen() {
  return (
    <ReactFlowProvider>
      <TopologyCanvas />
    </ReactFlowProvider>
  );
}
