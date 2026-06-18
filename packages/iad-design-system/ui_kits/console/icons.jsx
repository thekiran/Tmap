/* IAD icon set — inline SVG, Lucide-style (24px grid, 2px stroke, round caps).
   Lucide is not bundled with the planned Tauri/React app, so the kit ships a
   small hand-picked subset matching Lucide's geometry. Exposed as window.Icons.
   Each is a function component taking {size, ...props}; stroke = currentColor. */
(function () {
  const h = React.createElement;
  function mk(paths) {
    return function Icon({ size = 18, style, ...rest }) {
      return h('svg', {
        viewBox: '0 0 24 24', width: size, height: size, fill: 'none',
        stroke: 'currentColor', strokeWidth: 2, strokeLinecap: 'round',
        strokeLinejoin: 'round', style: { display: 'block', ...style }, ...rest,
      }, paths.map((d, i) => h('path', { key: i, d })));
    };
  }
  function mkRaw(children) {
    return function Icon({ size = 18, style, ...rest }) {
      return h('svg', {
        viewBox: '0 0 24 24', width: size, height: size, fill: 'none',
        stroke: 'currentColor', strokeWidth: 2, strokeLinecap: 'round',
        strokeLinejoin: 'round', style: { display: 'block', ...style }, ...rest,
      }, children(h));
    };
  }

  window.Icons = {
    // nav
    dashboard: mk(['M3 3h7v7H3zM14 3h7v7h-7zM14 14h7v7h-7zM3 14h7v7H3z']),
    topology: mkRaw((h) => [
      h('circle', { key: 1, cx: 5, cy: 6, r: 2.4 }), h('circle', { key: 2, cx: 19, cy: 6, r: 2.4 }),
      h('circle', { key: 3, cx: 12, cy: 18, r: 2.4 }),
      h('path', { key: 4, d: 'M7 6h10M6.5 8 11 16M17.5 8 13 16' }),
    ]),
    devices: mk(['M3 5h18v11H3z', 'M8 21h8M12 16v5']),
    evidence: mk(['M9 3h6l3 4v14H6V3z', 'M9 12h6M9 16h4']),
    reports: mk(['M14 3v5h5', 'M14 3H6v18h12V8z', 'M9 13h6M9 17h6']),
    settings: mkRaw((h) => [
      h('circle', { key: 1, cx: 12, cy: 12, r: 3 }),
      h('path', { key: 2, d: 'M19.4 13.5a1.7 1.7 0 0 0 .3 1.9l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-2.9 1.2V21a2 2 0 1 1-4 0v-.2a1.7 1.7 0 0 0-2.9-1.2l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0-1.2-2.9H3a2 2 0 1 1 0-4h.2a1.7 1.7 0 0 0 1.2-2.9l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 2.9-1.2V3a2 2 0 1 1 4 0v.2a1.7 1.7 0 0 0 2.9 1.2l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.9Z' }),
    ]),
    // actions
    refresh: mk(['M21 12a9 9 0 1 1-3-6.7L21 8', 'M21 3v5h-5']),
    download: mk(['M12 3v12', 'm7 12 5 5 5-5', 'M5 21h14']),
    upload: mk(['M12 21V9', 'm7 12 5-5 5 5', 'M5 3h14']),
    copy: mkRaw((h) => [
      h('rect', { key: 1, x: 9, y: 9, width: 12, height: 12, rx: 2 }),
      h('path', { key: 2, d: 'M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1' }),
    ]),
    search: mkRaw((h) => [h('circle', { key: 1, cx: 11, cy: 11, r: 7 }), h('path', { key: 2, d: 'm21 21-4.3-4.3' })]),
    filter: mk(['M3 5h18l-7 8v6l-4-2v-4z']),
    close: mk(['M18 6 6 18M6 6l12 12']),
    chevronRight: mk(['m9 6 6 6-6 6']),
    chevronDown: mk(['m6 9 6 6 6-6']),
    external: mk(['M15 3h6v6', 'M10 14 21 3', 'M21 14v5a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5']),
    // topology controls
    zoomIn: mkRaw((h) => [h('circle', { key: 1, cx: 11, cy: 11, r: 7 }), h('path', { key: 2, d: 'm21 21-4.3-4.3M11 8v6M8 11h6' })]),
    zoomOut: mkRaw((h) => [h('circle', { key: 1, cx: 11, cy: 11, r: 7 }), h('path', { key: 2, d: 'm21 21-4.3-4.3M8 11h6' })]),
    fit: mk(['M3 8V5a2 2 0 0 1 2-2h3', 'M21 8V5a2 2 0 0 0-2-2h-3', 'M3 16v3a2 2 0 0 0 2 2h3', 'M21 16v3a2 2 0 0 1-2 2h-3']),
    reset: mk(['M3 12a9 9 0 1 0 9-9 9 9 0 0 0-6.4 2.6L3 8', 'M3 3v5h5']),
    layers: mk(['m12 2 9 5-9 5-9-5z', 'm3 12 9 5 9-5', 'm3 17 9 5 9-5']),
    lock: mkRaw((h) => [h('rect', { key: 1, x: 4, y: 11, width: 16, height: 10, rx: 2 }), h('path', { key: 2, d: 'M8 11V7a4 4 0 0 1 8 0v4' })]),
    // node / device types
    host: mkRaw((h) => [h('rect', { key: 1, x: 3, y: 4, width: 18, height: 12, rx: 2 }), h('path', { key: 2, d: 'M8 20h8M12 16v4' })]),
    router: mkRaw((h) => [h('rect', { key: 1, x: 2, y: 13, width: 20, height: 7, rx: 2 }), h('path', { key: 2, d: 'M6 17h.01M10 17h.01M14 8l2-2 2 2M16 6v7' })]),
    modem: mkRaw((h) => [h('rect', { key: 1, x: 2, y: 6, width: 20, height: 12, rx: 2 }), h('path', { key: 2, d: 'M6 18v2M18 18v2M6 10h.01M10 10h.01' })]),
    ap: mkRaw((h) => [h('path', { key: 1, d: 'M5 12.5a7 7 0 0 1 14 0M8 15a4 4 0 0 1 8 0' }), h('circle', { key: 2, cx: 12, cy: 18, r: 1.4 })]),
    switchIcon: mkRaw((h) => [h('rect', { key: 1, x: 3, y: 8, width: 18, height: 8, rx: 2 }), h('path', { key: 2, d: 'M7 12h.01M11 12h.01M15 12h.01' })]),
    server: mkRaw((h) => [h('rect', { key: 1, x: 3, y: 4, width: 18, height: 7, rx: 2 }), h('rect', { key: 2, x: 3, y: 13, width: 18, height: 7, rx: 2 }), h('path', { key: 3, d: 'M7 7.5h.01M7 16.5h.01' })]),
    printer: mkRaw((h) => [h('path', { key: 1, d: 'M6 9V3h12v6' }), h('rect', { key: 2, x: 4, y: 9, width: 16, height: 7, rx: 2 }), h('path', { key: 3, d: 'M7 16h10v5H7z' })]),
    mobile: mkRaw((h) => [h('rect', { key: 1, x: 7, y: 2, width: 10, height: 20, rx: 2 }), h('path', { key: 2, d: 'M11 18h2' })]),
    iot: mkRaw((h) => [h('circle', { key: 1, cx: 12, cy: 12, r: 3 }), h('path', { key: 2, d: 'M12 2v4M12 18v4M2 12h4M18 12h4' })]),
    globe: mkRaw((h) => [h('circle', { key: 1, cx: 12, cy: 12, r: 9 }), h('path', { key: 2, d: 'M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18' })]),
    unknown: mkRaw((h) => [h('circle', { key: 1, cx: 12, cy: 12, r: 9 }), h('path', { key: 2, d: 'M9.2 9a3 3 0 0 1 5.6 1c0 2-3 2.5-3 4' }), h('path', { key: 3, d: 'M12 17h.01' })]),
    // status / misc
    alert: mkRaw((h) => [h('path', { key: 1, d: 'M12 3 2 20h20z' }), h('path', { key: 2, d: 'M12 10v4M12 17h.01' })]),
    info: mkRaw((h) => [h('circle', { key: 1, cx: 12, cy: 12, r: 9 }), h('path', { key: 2, d: 'M12 11v5M12 8h.01' })]),
    check: mk(['M20 6 9 17l-5-5']),
    shield: mk(['M12 3 5 6v5c0 4 3 7 7 9 4-2 7-5 7-9V6z', 'M9.5 12l2 2 3.5-4']),
    plug: mk(['M9 2v6M15 2v6', 'M7 8h10v3a5 5 0 0 1-10 0z', 'M12 16v6']),
    activity: mk(['M3 12h4l3 8 4-16 3 8h4']),
  };
})();
