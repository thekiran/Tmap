import { useEffect, useMemo, useState } from 'react';
import {
  Background,
  Controls,
  MiniMap,
  ReactFlow,
  ReactFlowProvider,
  useReactFlow,
  type Edge,
  type EdgeMouseHandler,
  type EdgeTypes,
  type Node,
  type NodeMouseHandler,
  type NodeTypes,
  type OnNodeDrag,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useScanStore } from '../../store/useScanStore';
import { useUIStore } from '../../store/useUIStore';
import { layoutWithElk } from '../../lib/layout-elk';
import type { LayoutPosition, NormalizedScanReport, TopologyEdge, TopologyNode } from '../../lib/models';
import { formatTopologyEdgeLabel, nodeDisplayTitle } from '../../lib/topology-display';
import { Icons } from '../icons/Icon';
import { TopologyNodeView } from './TopologyNodeView';
import { TopologyEdgeView } from './TopologyEdgeView';

const nodeTypes: NodeTypes = { iadNode: TopologyNodeView };
const edgeTypes: EdgeTypes = { iadEdge: TopologyEdgeView };
type InteractionMode = 'move' | 'pan';

function TopologyCanvas() {
  const { fitView, zoomIn, zoomOut } = useReactFlow();
  const normalized = useScanStore((state) => state.normalized);
  const topology = normalized?.topology;
  const isScanning = useScanStore((state) => state.isScanning);
  const scanError = useScanStore((state) => state.scanError);
  const settings = useUIStore((state) => state.settings);
  const layoutPositions = useUIStore((state) => state.layoutPositions);
  const setNodePosition = useUIStore((state) => state.setNodePosition);
  const resetLayoutPositions = useUIStore((state) => state.resetLayoutPositions);
  const selectNode = useUIStore((state) => state.selectNode);
  const selectEdge = useUIStore((state) => state.selectEdge);
  const selectedNodeId = useUIStore((state) => state.selectedNodeId);
  const selectedEdgeId = useUIStore((state) => state.selectedEdgeId);
  const [autoPositions, setAutoPositions] = useState<Record<string, LayoutPosition>>({});
  const [interactionMode, setInteractionMode] = useState<InteractionMode>('move');
  const [showLineLabels, setShowLineLabels] = useState(true);
  const [showMiniMap, setShowMiniMap] = useState(true);

  useEffect(() => {
    if (!topology) return;
    let cancelled = false;
    layoutWithElk(topology, settings.layoutEngine).then((positions) => {
      if (!cancelled) setAutoPositions(positions);
    });
    return () => {
      cancelled = true;
    };
  }, [settings.layoutEngine, topology]);

  useEffect(() => {
    if (!topology || topology.nodes.length === 0 || Object.keys(autoPositions).length === 0) return;
    const timer = window.setTimeout(() => void fitView({ padding: 0.24, duration: 180 }), 40);
    return () => window.clearTimeout(timer);
  }, [autoPositions, fitView, topology]);

  const nodes: Node<TopologyNode>[] = useMemo(() => {
    if (!topology) return [];
    return topology.nodes.map((node: TopologyNode) => ({
      id: node.id,
      type: 'iadNode',
      data: node,
      position: layoutPositions[node.id] ?? autoPositions[node.id] ?? node.position,
      selected: selectedNodeId === node.id,
      draggable: interactionMode === 'move',
    }));
  }, [autoPositions, interactionMode, layoutPositions, selectedNodeId, topology]);

  const edges: Edge<TopologyEdge>[] = useMemo(() => {
    if (!topology) return [];
    return topology.edges.map((edge: TopologyEdge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      type: 'iadEdge',
      data: { ...edge, showLabel: showLineLabels },
      selected: selectedEdgeId === edge.id,
    }));
  }, [selectedEdgeId, showLineLabels, topology]);

  const nodeLabels = useMemo(() => {
    const labels = new Map<string, string>();
    topology?.nodes.forEach((node) => labels.set(node.id, nodeDisplayTitle(node)));
    return labels;
  }, [topology]);

  const onNodeClick: NodeMouseHandler = (_, node) => selectNode(node.id);
  const onEdgeClick: EdgeMouseHandler = (_, edge) => selectEdge(edge.id);
  const onNodeDragStop: OnNodeDrag = (_, node) => setNodePosition(node.id, node.position);

  if (!topology || topology.nodes.length === 0) {
    return (
      <div className="flex h-full items-center justify-center p-8 text-center text-sm text-zinc-500">
        {isScanning ? (
          <span className="inline-flex items-center gap-3">
            <span className="h-4 w-4 animate-spin rounded-full border-2 border-zinc-400 border-t-transparent" />
            Scanning the network… this can take up to a minute.
          </span>
        ) : scanError ? (
          <div className="max-w-lg rounded-md border border-red-500/40 bg-red-500/10 p-4 text-left">
            <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-red-500">Scan failed</div>
            <div className="font-mono text-[11px] leading-relaxed text-zinc-600 dark:text-zinc-300">{scanError}</div>
          </div>
        ) : (
          <span>No scan loaded. Click <b>Run Scan</b> to map your network, or import a report.</span>
        )}
      </div>
    );
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      {normalized && <TopologyBanner normalized={normalized} />}
      <div className="flex h-11 items-center gap-4 border-b border-zinc-800 bg-zinc-950/80 px-4 text-xs text-zinc-400">
        <span className="font-mono uppercase tracking-[0.2em] text-zinc-500">Topology map</span>
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-2"><i className="h-px w-8 bg-zinc-500" /> physical only if proven</span>
          <span className="inline-flex items-center gap-2"><i className="h-px w-8 border-t border-dotted border-zinc-500" /> inferred</span>
        </div>
        <span className="ml-auto font-mono text-zinc-500">
          {normalized?.discoverySummary
            ? `${normalized.discoverySummary.devicesFound} devices discovered / ${normalized.discoverySummary.addressesScanned} addresses scanned`
            : `${topology.nodes.length} nodes / ${topology.edges.length} edges`}
        </span>
      </div>
      <div className="relative min-h-0 flex-1 bg-[radial-gradient(circle_at_1px_1px,rgba(148,163,184,.16)_1px,transparent_0)] [background-size:22px_22px]">
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
          onNodeClick={onNodeClick}
          onEdgeClick={onEdgeClick}
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
          <Controls showInteractive={false} position="bottom-right" />
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
      {topology.edges.length > 0 ? (
        <div className="shrink-0 overflow-x-auto border-t border-zinc-800 bg-zinc-950 px-4 py-2 font-mono text-[11px] text-zinc-500">
          <div className="whitespace-nowrap">
            {topology.edges.map((edge: TopologyEdge) => `${nodeLabels.get(edge.source) ?? edge.source} -> ${nodeLabels.get(edge.target) ?? edge.target}: ${formatTopologyEdgeLabel(edge)}`).join('  |  ')}
          </div>
        </div>
      ) : null}
    </div>
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
    </div>
  );
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
      <span className="hidden sm:inline">{label}</span>
    </button>
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
