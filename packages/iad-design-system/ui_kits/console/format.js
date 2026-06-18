/* IAD UI kit — formatting helpers. window.fmt */
window.fmt = {
  pct: (v) => Math.round(v * 100) + '%',
  band: (v) => (v >= 0.75 ? 'high' : v >= 0.45 ? 'medium' : 'low'),
  bandWord: (v) => (v >= 0.75 ? 'High' : v >= 0.45 ? 'Medium' : 'Low'),
  bandColor: (v) => (v >= 0.75 ? 'var(--conf-high)' : v >= 0.45 ? 'var(--conf-med)' : 'var(--conf-low)'),
  bandBg: (v) => (v >= 0.75 ? 'var(--conf-high-bg)' : v >= 0.45 ? 'var(--conf-med-bg)' : 'var(--conf-low-bg)'),
  time: (iso) => { try { return new Date(iso).toLocaleString('en-GB', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: 'short' }); } catch (e) { return iso; } },
  ago: (iso) => { const s = (Date.now() - new Date(iso).getTime()) / 1000; if (s < 60) return 'just now'; if (s < 3600) return Math.floor(s / 60) + 'm ago'; if (s < 86400) return Math.floor(s / 3600) + 'h ago'; return Math.floor(s / 86400) + 'd ago'; },
  reach: (r) => ({ self: { word: 'This host', tone: 'accent' }, reachable: { word: 'Reachable', tone: 'success' }, partial: { word: 'Partial', tone: 'warn' }, unreachable: { word: 'Unreachable', tone: 'danger' } }[r] || { word: r, tone: 'neutral' }),
  deviceIcon: (t) => ({ local_host: 'host', default_gateway: 'router', router: 'router', modem_cpe: 'modem', access_point: 'ap', mesh_node: 'ap', managed_switch: 'switchIcon', server: 'server', printer: 'printer', mobile: 'mobile', workstation: 'host', iot: 'iot', dns_server: 'server', isp_gateway: 'globe', public_internet: 'globe', unknown: 'unknown' }[t] || 'unknown'),
};
