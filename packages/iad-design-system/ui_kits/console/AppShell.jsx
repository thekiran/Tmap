/* IAD UI kit — Application shell: sidebar, top status bar, content router,
   right details drawer, bottom log strip. Exposes window.AppShell. */
(function () {
  const { useState } = React;
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const { Badge, IconButton, Button, StatusDot } = NS;
  const I = window.Icons;
  const fmt = window.fmt;

  const NAV = [
    { id: 'dashboard', label: 'Dashboard', icon: 'dashboard' },
    { id: 'topology', label: 'Topology', icon: 'topology' },
    { id: 'devices', label: 'Devices', icon: 'devices' },
    { id: 'evidence', label: 'Evidence', icon: 'evidence' },
    { id: 'reports', label: 'Reports', icon: 'reports' },
    { id: 'settings', label: 'Settings', icon: 'settings' },
  ];

  function Sidebar({ active, onNav, scan }) {
    const counts = { devices: scan.devices.length, evidence: scan.evidence.length };
    return (
      React.createElement('aside', { style: {
        width: 'var(--sidebar-w)', flex: '0 0 auto', background: 'var(--ink-900)',
        borderRight: '1px solid var(--hairline)', display: 'flex', flexDirection: 'column',
      } },
        React.createElement('div', { style: { height: 'var(--topbar-h)', display: 'flex', alignItems: 'center', gap: 10, padding: '0 16px', borderBottom: '1px solid var(--hairline)' } },
          React.createElement('img', { src: '../../assets/mark-iad.svg', alt: 'IAD', style: { width: 26, height: 26 } }),
          React.createElement('span', { style: { fontFamily: 'var(--font-mono)', fontWeight: 700, letterSpacing: '1.5px', color: 'var(--fg-1)', fontSize: 15 } }, 'IAD'),
          React.createElement('span', { style: { fontFamily: 'var(--font-mono)', fontSize: 10, color: 'var(--fg-4)', marginLeft: 'auto' } }, 'v0.9'),
        ),
        React.createElement('nav', { style: { padding: 10, display: 'flex', flexDirection: 'column', gap: 2 } },
          NAV.map((n) => {
            const on = n.id === active;
            const Icon = I[n.icon];
            return React.createElement('button', {
              key: n.id, onClick: () => onNav(n.id),
              style: {
                display: 'flex', alignItems: 'center', gap: 11, padding: '9px 11px',
                borderRadius: 'var(--radius-md)', border: '1px solid ' + (on ? 'var(--hairline)' : 'transparent'),
                background: on ? 'var(--surface-2)' : 'transparent', cursor: 'pointer', textAlign: 'left',
                color: on ? 'var(--fg-1)' : 'var(--fg-3)', font: '500 var(--text-sm) var(--font-sans)',
                transition: 'background var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out)',
              },
            },
              React.createElement('span', { style: { display: 'inline-flex', color: on ? 'var(--accent-bright)' : 'var(--fg-4)' } }, React.createElement(Icon, { size: 17 })),
              n.label,
              (counts[n.id] != null) && React.createElement('span', { style: { marginLeft: 'auto', fontFamily: 'var(--font-mono)', fontSize: 10, color: 'var(--fg-4)' } }, counts[n.id]),
            );
          }),
        ),
        React.createElement('div', { style: { marginTop: 'auto', padding: 14, borderTop: '1px solid var(--hairline)' } },
          React.createElement('div', { style: { display: 'flex', alignItems: 'center', gap: 8, padding: '8px 10px', background: 'var(--ok-bg)', borderRadius: 'var(--radius-md)' } },
            React.createElement(I.shield, { size: 15, style: { color: 'var(--ok)' } }),
            React.createElement('span', { style: { fontSize: 11, color: 'var(--ok)', fontWeight: 600 } }, 'Safe mode · read-only'),
          ),
        ),
      )
    );
  }

  function TopMetric({ label, value, mono = true, tone, minW = 88, maxW }) {
    return React.createElement('div', { style: { display: 'flex', flexDirection: 'column', gap: 1, minWidth: minW, maxWidth: maxW, flex: '0 0 auto' } },
      React.createElement('span', { style: { font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)', fontSize: 9 } }, label),
      React.createElement('span', { style: { fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)', fontSize: 12, fontWeight: 600, color: tone || 'var(--fg-1)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' } }, value),
    );
  }

  function TopStatusBar({ scan, onImport, onExport }) {
    const c = scan.confidence;
    return React.createElement('header', { style: {
      height: 'var(--topbar-h)', flex: '0 0 auto', background: 'var(--ink-850)',
      borderBottom: '1px solid var(--hairline)', display: 'flex', alignItems: 'center',
      gap: 16, padding: '0 16px', overflow: 'hidden',
    } },
      React.createElement('div', { style: { display: 'flex', alignItems: 'center', gap: 9 } },
        React.createElement('div', { style: { display: 'flex', flexDirection: 'column', gap: 1 } },
          React.createElement('span', { style: { fontFamily: 'var(--font-mono)', fontSize: 9, color: 'var(--fg-4)', textTransform: 'uppercase', letterSpacing: '.1em' } }, 'Scan'),
          React.createElement('span', { style: { fontFamily: 'var(--font-mono)', fontSize: 12, color: 'var(--fg-1)', fontWeight: 600 } }, scan.scan_id),
        ),
        React.createElement(Badge, { tone: 'neutral', appearance: 'outline', size: 'sm', mono: true }, scan.mode),
      ),
      React.createElement('div', { style: { width: 1, height: 26, background: 'var(--hairline)' } }),
      React.createElement(TopMetric, { label: 'Time', value: fmt.time(scan.created_at), mono: false, minW: 96 }),
      React.createElement(TopMetric, { label: 'Interface', value: scan.detected_network_context.selected_interface.name, minW: 80 }),
      React.createElement(TopMetric, { label: 'Public IP', value: scan.public_ip.address, minW: 110 }),
      React.createElement(TopMetric, { label: 'ISP', value: scan.public_ip.asn + ' · ' + scan.public_ip.org, mono: false, minW: 120, maxW: 190 }),
      React.createElement('div', { style: { display: 'flex', alignItems: 'center', gap: 7, marginLeft: 'auto' } },
        React.createElement('div', { style: { display: 'flex', flexDirection: 'column', gap: 1, alignItems: 'flex-end' } },
          React.createElement('span', { style: { font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-4)', fontSize: 9 } }, 'Confidence'),
          React.createElement('div', { style: { display: 'flex', alignItems: 'center', gap: 6 } },
            React.createElement('span', { style: { fontFamily: 'var(--font-mono)', fontSize: 13, fontWeight: 700, color: fmt.bandColor(c) } }, fmt.pct(c)),
            React.createElement('span', { style: { fontSize: 9, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '.08em', color: fmt.bandColor(c) } }, fmt.bandWord(c)),
          ),
        ),
        React.createElement(Badge, { tone: scan.decision_quality === 'high' ? 'success' : scan.decision_quality === 'medium' ? 'warn' : 'neutral', size: 'sm', uppercase: true, mono: true }, scan.decision_quality + ' quality'),
      ),
      React.createElement('div', { style: { width: 1, height: 26, background: 'var(--hairline)' } }),
      React.createElement('div', { style: { display: 'flex', gap: 8 } },
        React.createElement(Button, { size: 'sm', variant: 'secondary', iconLeft: React.createElement(I.upload, { size: 14 }), onClick: onImport }, 'Import'),
        React.createElement(Button, { size: 'sm', variant: 'primary', iconLeft: React.createElement(I.download, { size: 14 }), onClick: onExport }, 'Export'),
      ),
    );
  }

  function LogStrip({ lines }) {
    return React.createElement('footer', { style: {
      height: 28, flex: '0 0 auto', background: 'var(--ink-900)', borderTop: '1px solid var(--hairline)',
      display: 'flex', alignItems: 'center', gap: 16, padding: '0 16px', overflow: 'hidden',
    } },
      React.createElement(StatusDot, { tone: 'success', size: 7 }),
      React.createElement('span', { style: { fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-3)', whiteSpace: 'nowrap' } }, lines),
    );
  }

  function AppShell({ scan }) {
    const [active, setActive] = useState('dashboard');
    const [drawer, setDrawer] = useState(null); // {kind, payload}
    const Screens = window.IAD_SCREENS || {};
    const Screen = Screens[active];

    const openDrawer = (kind, payload) => setDrawer({ kind, payload });
    const closeDrawer = () => setDrawer(null);

    const log = {
      dashboard: '09:41:22  scan complete · 10 probes · 8 devices · 6.84s',
      topology: '09:41:22  topology generated from context · 6 nodes · 7 edges · inferred where dashed',
      devices: '09:41:17  arp_sweep resolved 8 hosts on 192.168.1.0/24',
      evidence: '09:41:20  cpe_snmp blocked (timeout) · physical medium remains inferred',
      reports: 'ready · export NormalizedScanReport v1',
      settings: 'UI layout positions stored locally · scan data immutable',
    }[active];

    return React.createElement('div', { style: { display: 'flex', height: '100%', width: '100%', background: 'var(--bg-app)' } },
      React.createElement(Sidebar, { active, onNav: setActive, scan }),
      React.createElement('div', { style: { flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 } },
        React.createElement(TopStatusBar, { scan, onImport: () => setActive('reports'), onExport: () => setActive('reports') }),
        React.createElement('div', { style: { flex: 1, display: 'flex', minHeight: 0 } },
          React.createElement('main', { style: { flex: 1, minWidth: 0, overflow: 'auto' } },
            Screen ? React.createElement(Screen, { scan, openDrawer, drawer }) :
              React.createElement('div', { style: { padding: 40, color: 'var(--fg-3)' } }, 'Screen not loaded'),
          ),
          drawer && window.DetailsDrawer && React.createElement(window.DetailsDrawer, { drawer, onClose: closeDrawer, scan }),
        ),
        React.createElement(LogStrip, { lines: log }),
      ),
    );
  }

  window.AppShell = AppShell;
})();
