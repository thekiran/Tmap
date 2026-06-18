/* IAD UI kit — Topology screen. Read-only interactive map (pan / zoom / drag /
   select), layer toggles, legend. No create/delete — layout positions are UI
   state only, never written back to scan data. window.IAD_SCREENS.topology
   NOTE: the production app uses React Flow + ELK; this is a faithful SVG
   recreation of that read-only topology mode. */
(function () {
  const { useState, useRef, useCallback } = React;
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const { IconButton, Toggle, Badge, TierBadge } = NS;
  const I = window.Icons;

  function buildGraph(scan) {
    const dev = (id) => scan.devices.find((d) => d.id === id);
    const nodes = [
      { id: 'host', label: 'This host', sub: '192.168.1.24', icon: 'host', x: 430, y: 50, layers: ['l3'], device: dev('d-host'), accent: true },
      { id: 'router', label: 'Home router', sub: '192.168.1.1', icon: 'router', x: 430, y: 168, layers: ['l2', 'l3'], device: dev('d-gw') },
      { id: 'ap', label: 'Mesh AP', sub: '192.168.1.2', icon: 'ap', x: 150, y: 250, layers: ['l2'], device: dev('d-ap') },
      { id: 'nas', label: 'NAS', sub: '192.168.1.30', icon: 'server', x: 280, y: 322, layers: ['l2'], device: dev('d-nas') },
      { id: 'printer', label: 'Printer', sub: '192.168.1.41', icon: 'printer', x: 600, y: 248, layers: ['l2'], device: dev('d-print') },
      { id: 'switch', label: 'Switch', sub: 'inferred · unmanaged', icon: 'switchIcon', x: 712, y: 330, layers: ['unknown'], inferred: true },
      { id: 'segment', label: 'Unknown L2 segment', sub: '≥1 host behind', icon: 'unknown', x: 712, y: 438, layers: ['unknown'], inferred: true },
      { id: 'cgnat', label: 'CGNAT gateway', sub: '100.64.12.1', icon: 'router', x: 430, y: 300, layers: ['l3', 'nat'], badge: 'NAT' },
      { id: 'isp', label: 'ISP edge', sub: '203.0.113.1', icon: 'globe', x: 430, y: 412, layers: ['isp'] },
      { id: 'inet', label: 'Public internet', sub: 'AS3320', icon: 'globe', x: 430, y: 524, layers: ['isp'] },
    ];
    const edges = [
      { id: 'e1', from: 'host', to: 'router', type: 'local_interface', kind: 'confirmed', tier: 'l3', conf: 1.0, layers: ['l3'], accent: true, label: 'Local interface', basis: 'Default route via this NIC; same /24 as gateway.' },
      { id: 'e2', from: 'router', to: 'nas', type: 'arp_confirmed', kind: 'confirmed', tier: 'l2', conf: 0.93, layers: ['l2'], label: 'ARP confirmed', basis: 'NAS answered ARP on the local broadcast domain.' },
      { id: 'e3', from: 'router', to: 'printer', type: 'arp_confirmed', kind: 'confirmed', tier: 'l2', conf: 0.81, layers: ['l2'], label: 'ARP confirmed', basis: 'Printer answered ARP + advertised IPP via mDNS.' },
      { id: 'e4', from: 'router', to: 'ap', type: 'wifi_association_inferred', kind: 'inferred', tier: 'l2', conf: 0.6, layers: ['l2'], label: 'Wi-Fi assoc (inferred)', basis: 'mDNS suggests a mesh repeater; bridge topology not confirmed.' },
      { id: 'e5', from: 'router', to: 'switch', type: 'unknown_l2_connection', kind: 'unknown', tier: 'l2', conf: 0.3, layers: ['unknown'], label: 'Unknown L2', basis: 'MAC counts imply an unmanaged switch, but it is invisible to probes.' },
      { id: 'e6', from: 'switch', to: 'segment', type: 'unknown_l2_connection', kind: 'unknown', tier: 'l2', conf: 0.22, layers: ['unknown'], label: 'Unknown L2', basis: 'At least one host sits beyond the inferred switch; count is a lower bound.' },
      { id: 'e7', from: 'router', to: 'cgnat', type: 'upstream_private_gateway', kind: 'confirmed', tier: 'l3', conf: 0.88, layers: ['l3', 'nat'], boundary: 'NAT', label: 'NAT boundary', basis: 'Next hop is RFC 6598 (100.64/10): carrier-grade NAT.' },
      { id: 'e8', from: 'cgnat', to: 'isp', type: 'route_hop', kind: 'confirmed', tier: 'isp', conf: 0.7, layers: ['l3', 'isp'], thin: true, label: 'Route hop', basis: 'Traceroute L3 hop — a router, not a physical switch.' },
      { id: 'e9', from: 'isp', to: 'inet', type: 'isp_boundary', kind: 'confirmed', tier: 'isp', conf: 0.85, layers: ['isp'], boundary: 'ISP', label: 'ISP boundary', basis: 'First public hop; ISP-internal topology is not observable.' },
    ];
    const byId = Object.fromEntries(nodes.map((n) => [n.id, n]));
    edges.forEach((e) => { e.fromLabel = byId[e.from].label; e.toLabel = byId[e.to].label; });
    return { nodes, edges };
  }

  const LAYERS = [
    { id: 'l2', label: 'L2 · Link', tier: 'l2' },
    { id: 'l3', label: 'L3 · Routing', tier: 'l3' },
    { id: 'nat', label: 'NAT', tier: 'nat' },
    { id: 'isp', label: 'ISP route context', tier: 'isp' },
    { id: 'unknown', label: 'Unknown segments', tier: null },
    { id: 'lowconf', label: 'Low-confidence edges', tier: null },
  ];

  function edgePath(a, b) {
    const mx = (a.x + b.x) / 2;
    return `M ${a.x} ${a.y} C ${mx} ${a.y}, ${mx} ${b.y}, ${b.x} ${b.y}`;
  }

  function Topology({ scan, openDrawer, drawer }) {
    const graphRef = useRef(buildGraph(scan));
    const { edges } = graphRef.current;
    const [positions, setPositions] = useState(() => Object.fromEntries(graphRef.current.nodes.map((n) => [n.id, { x: n.x, y: n.y }])));
    const [view, setView] = useState({ x: 60, y: 20, k: 0.92 });
    const [layers, setLayers] = useState({ l2: true, l3: true, nat: true, isp: true, unknown: true, lowconf: true });
    const [sel, setSel] = useState(null);
    const svgRef = useRef(null);
    const drag = useRef(null);

    const nodes = graphRef.current.nodes.map((n) => ({ ...n, ...positions[n.id] }));
    const nodeVisible = (n) => n.layers.some((l) => layers[l]) || n.id === 'host';
    const edgeVisible = (e) => {
      if (!e.layers.some((l) => layers[l])) return false;
      if (e.conf < 0.45 && !layers.lowconf) return false;
      return true;
    };

    const onWheel = useCallback((ev) => {
      ev.preventDefault();
      setView((v) => {
        const k = Math.max(0.4, Math.min(2.2, v.k * (ev.deltaY < 0 ? 1.1 : 0.9)));
        const rect = svgRef.current.getBoundingClientRect();
        const cx = ev.clientX - rect.left, cy = ev.clientY - rect.top;
        const nx = cx - (cx - v.x) * (k / v.k);
        const ny = cy - (cy - v.y) * (k / v.k);
        return { x: nx, y: ny, k };
      });
    }, []);

    const onPointerDownBg = (ev) => {
      if (ev.target.closest('[data-node]') || ev.target.closest('[data-edge]')) return;
      setSel(null);
      drag.current = { mode: 'pan', sx: ev.clientX, sy: ev.clientY, ox: view.x, oy: view.y };
    };
    const onPointerDownNode = (ev, n) => {
      ev.stopPropagation();
      setSel({ kind: 'node', id: n.id });
      drag.current = { mode: 'node', id: n.id, sx: ev.clientX, sy: ev.clientY, ox: positions[n.id].x, oy: positions[n.id].y, moved: false };
    };
    const onPointerMove = (ev) => {
      const d = drag.current; if (!d) return;
      if (d.mode === 'pan') setView((v) => ({ ...v, x: d.ox + (ev.clientX - d.sx), y: d.oy + (ev.clientY - d.sy) }));
      else if (d.mode === 'node') {
        const dx = (ev.clientX - d.sx) / view.k, dy = (ev.clientY - d.sy) / view.k;
        if (Math.abs(dx) > 1 || Math.abs(dy) > 1) d.moved = true;
        setPositions((p) => ({ ...p, [d.id]: { x: d.ox + dx, y: d.oy + dy } }));
      }
    };
    const onPointerUp = (ev) => {
      const d = drag.current;
      if (d && d.mode === 'node' && !d.moved) {
        const n = graphRef.current.nodes.find((x) => x.id === d.id);
        openDrawer('node', n.device ? { device: n.device } : { probe_name: n.label });
      }
      drag.current = null;
    };

    const fitView = () => setView({ x: 60, y: 20, k: 0.92 });
    const resetLayout = () => setPositions(Object.fromEntries(graphRef.current.nodes.map((n) => [n.id, { x: n.x, y: n.y }])));

    const tierColor = (t) => ({ l2: 'var(--tier-l2)', l3: 'var(--fg-2)', nat: 'var(--tier-nat)', isp: 'var(--tier-isp)' }[t] || 'var(--fg-3)');
    const edgeStroke = (e) => {
      if (e.accent) return 'var(--accent-base)';
      if (e.conf < 0.45 && layers.lowconf) return 'var(--edge-muted)';
      if (e.kind === 'unknown') return 'var(--edge-unknown)';
      if (e.boundary) return tierColor(e.tier);
      if (e.kind === 'inferred') return 'var(--edge-inferred)';
      return 'var(--edge-confirmed)';
    };
    const edgeDash = (e) => (e.kind === 'unknown' ? '2 7' : e.kind === 'inferred' ? '7 6' : (e.conf < 0.45 ? '7 6' : 'none'));

    return (
      <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 22px', borderBottom: '1px solid var(--hairline)', flex: '0 0 auto' }}>
          <div>
            <h1 style={{ font: 'var(--type-h2)' }}>Topology</h1>
            <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>Generated from network context · 6 confirmed · 3 inferred links · read-only</span>
          </div>
          <Badge tone="neutral" appearance="outline" size="sm" mono><span style={{ display: 'inline-flex', marginRight: 5, verticalAlign: 'middle' }}><I.lock size={11} /></span>No edits</Badge>
        </div>

        <div style={{ flex: 1, display: 'flex', minHeight: 0 }}>
          {/* Canvas */}
          <div style={{ flex: 1, position: 'relative', minWidth: 0, background: 'var(--ink-800)', backgroundImage: 'radial-gradient(var(--grid-line-strong) 1px, transparent 1px)', backgroundSize: 24 * view.k + 'px ' + 24 * view.k + 'px', backgroundPosition: view.x + 'px ' + view.y + 'px', overflow: 'hidden' }}>
            <svg ref={svgRef} width="100%" height="100%" style={{ position: 'absolute', inset: 0, cursor: drag.current && drag.current.mode === 'pan' ? 'grabbing' : 'grab' }}
              onWheel={onWheel} onPointerDown={onPointerDownBg} onPointerMove={onPointerMove} onPointerUp={onPointerUp} onPointerLeave={onPointerUp}>
              <g transform={`translate(${view.x} ${view.y}) scale(${view.k})`}>
                {edges.filter(edgeVisible).map((e) => {
                  const a = positions[e.from], b = positions[e.to];
                  const selected = sel && sel.kind === 'edge' && sel.id === e.id;
                  const mx = (a.x + b.x) / 2, my = (a.y + b.y) / 2;
                  return (
                    <g key={e.id} data-edge={e.id} style={{ cursor: 'pointer' }}
                      onPointerDown={(ev) => { ev.stopPropagation(); setSel({ kind: 'edge', id: e.id }); openDrawer('edge', e); }}>
                      <path d={edgePath(a, b)} fill="none" stroke="transparent" strokeWidth={14} />
                      <path d={edgePath(a, b)} fill="none" stroke={selected ? 'var(--accent-base)' : edgeStroke(e)} strokeWidth={e.thin ? 1.3 : selected ? 2.6 : 1.8} strokeDasharray={edgeDash(e)} strokeLinecap="round" />
                      {e.boundary && (
                        <g transform={`translate(${mx} ${my})`}>
                          <rect x={-22} y={-9} width={44} height={18} rx={4} fill="var(--ink-850)" stroke={tierColor(e.tier)} strokeWidth={1} />
                          <text x={0} y={4} textAnchor="middle" fontFamily="var(--font-mono)" fontSize={10} fontWeight={700} fill={tierColor(e.tier)} style={{ letterSpacing: '.06em' }}>{e.boundary}</text>
                        </g>
                      )}
                    </g>
                  );
                })}
                {nodes.filter(nodeVisible).map((n) => {
                  const Icon = I[n.icon];
                  const selected = sel && sel.kind === 'node' && sel.id === n.id;
                  return (
                    <g key={n.id} data-node={n.id} transform={`translate(${n.x} ${n.y})`} style={{ cursor: 'grab' }}
                      onPointerDown={(ev) => onPointerDownNode(ev, n)}>
                      <g transform="translate(-66 -26)">
                        <rect width={132} height={52} rx={8} fill="var(--node-fill)"
                          stroke={selected ? 'var(--accent-base)' : n.accent ? 'var(--accent-ring)' : 'var(--node-stroke)'}
                          strokeWidth={selected ? 2 : 1} strokeDasharray={n.inferred ? '5 4' : 'none'} />
                        {selected && <rect x={-3} y={-3} width={138} height={58} rx={10} fill="none" stroke="var(--accent-base)" strokeWidth={1} opacity={0.35} />}
                        <g transform="translate(11 11)">
                          <rect width={30} height={30} rx={6} fill="var(--surface-3)" />
                          <g transform="translate(7 7)" color={n.accent ? 'var(--accent-bright)' : n.inferred ? 'var(--fg-3)' : 'var(--fg-2)'}>
                            <Icon size={16} />
                          </g>
                        </g>
                        <text x={50} y={22} fontFamily="var(--font-sans)" fontSize={12} fontWeight={600} fill="var(--fg-1)">{n.label}</text>
                        <text x={50} y={38} fontFamily="var(--font-mono)" fontSize={10} fill="var(--fg-3)">{n.sub}</text>
                        {n.badge && <g transform="translate(104 7)"><rect width={20} height={13} rx={3} fill="var(--tier-nat-bg)" /><text x={10} y={9.5} textAnchor="middle" fontFamily="var(--font-mono)" fontSize={8} fontWeight={700} fill="var(--tier-nat)">{n.badge}</text></g>}
                      </g>
                    </g>
                  );
                })}
              </g>
            </svg>

            {/* Toolbar */}
            <div style={{ position: 'absolute', top: 14, right: 14, display: 'flex', flexDirection: 'column', gap: 6, background: 'var(--surface-card)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-md)', padding: 5 }}>
              <IconButton label="Zoom in" onClick={() => setView((v) => ({ ...v, k: Math.min(2.2, v.k * 1.15) }))}><I.zoomIn size={16} /></IconButton>
              <IconButton label="Zoom out" onClick={() => setView((v) => ({ ...v, k: Math.max(0.4, v.k * 0.87) }))}><I.zoomOut size={16} /></IconButton>
              <IconButton label="Fit view" onClick={fitView}><I.fit size={16} /></IconButton>
              <IconButton label="Reset layout" onClick={resetLayout}><I.reset size={16} /></IconButton>
            </div>
            <div style={{ position: 'absolute', bottom: 12, left: 14, fontFamily: 'var(--font-mono)', fontSize: 10, color: 'var(--fg-4)' }}>scroll = zoom · drag bg = pan · drag node = reposition · {Math.round(view.k * 100)}%</div>
          </div>

          {/* Right rail: layers + legend */}
          <div style={{ width: 232, flex: '0 0 auto', borderLeft: '1px solid var(--hairline)', background: 'var(--surface-card)', overflow: 'auto', padding: 16, display: 'flex', flexDirection: 'column', gap: 18 }}>
            <div>
              <div style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-3)', marginBottom: 12 }}>Layers</div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                {LAYERS.map((l) => (
                  <div key={l.id} style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
                    <Toggle size="sm" checked={layers[l.id]} onChange={(v) => setLayers((s) => ({ ...s, [l.id]: v }))} />
                    {l.tier ? <TierBadge tier={l.tier} appearance="dot" label={l.label} /> : <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-2)' }}>{l.label}</span>}
                  </div>
                ))}
              </div>
            </div>
            <div style={{ borderTop: '1px solid var(--hairline)', paddingTop: 16 }}>
              <div style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-3)', marginBottom: 12 }}>Edge legend</div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                {[['Confirmed', 'var(--edge-confirmed)', 'none', 1.8], ['Inferred', 'var(--edge-inferred)', '7 6', 1.8], ['Unknown L2', 'var(--edge-unknown)', '2 7', 1.8], ['Route hop', 'var(--edge-confirmed)', 'none', 1], ['Low confidence', 'var(--edge-muted)', '7 6', 1.8]].map((r, i) => (
                  <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <svg width={40} height={10} style={{ flex: '0 0 auto' }}><line x1={1} y1={5} x2={39} y2={5} stroke={r[1]} strokeWidth={r[3]} strokeDasharray={r[2]} strokeLinecap="round" /></svg>
                    <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-2)' }}>{r[0]}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  window.IAD_SCREENS = window.IAD_SCREENS || {};
  window.IAD_SCREENS.topology = Topology;
})();
