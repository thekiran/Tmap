/* IAD UI kit — Reports + Settings screens.
   window.IAD_SCREENS.reports / .settings */
(function () {
  const { useState } = React;
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const { Card, Button, Toggle, SegmentedControl, Badge, StatusDot } = NS;
  const I = window.Icons;
  const fmt = window.fmt;

  function Reports({ scan }) {
    const [copied, setCopied] = useState(false);
    const summary = `IAD scan ${scan.scan_id} — ${scan.primary_type} (${fmt.pct(scan.confidence)} ${fmt.bandWord(scan.confidence)} confidence, ${scan.decision_quality} quality)\nISP ${scan.public_ip.asn} ${scan.public_ip.org} · Public IP ${scan.public_ip.address}\nNAT: CGNAT (${scan.nat_topology.layers} layers) · IPv6 native · ${scan.devices.length} LAN devices`;
    const copy = () => { try { navigator.clipboard.writeText(summary); } catch (e) {} setCopied(true); setTimeout(() => setCopied(false), 1600); };

    const Action = ({ icon, title, desc, btn, onClick, primary, disabled }) => (
      <div style={{ display: 'flex', alignItems: 'center', gap: 14, padding: 16, background: 'var(--surface-card)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-lg)' }}>
        <span style={{ width: 38, height: 38, borderRadius: 'var(--radius-md)', background: 'var(--surface-3)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--fg-2)', flex: '0 0 auto' }}>{icon}</span>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 600 }}>{title}</div>
          <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>{desc}</div>
        </div>
        <Button size="sm" variant={primary ? 'primary' : 'secondary'} onClick={onClick} disabled={disabled}>{btn}</Button>
      </div>
    );

    return (
      <div style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 14, maxWidth: 860, margin: '0 auto' }}>
        <h1 style={{ font: 'var(--type-h1)' }}>Reports</h1>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          <Action icon={<I.upload size={19} />} title="Import scan JSON" desc="Validate against the NormalizedScanReport schema (Zod) before loading." btn="Choose file" />
          <Action icon={<I.download size={19} />} title="Export full report" desc="The immutable scan evidence as JSON. UI layout positions are excluded." btn="Export JSON" primary />
          <Action icon={<I.reports size={19} />} title="Export scan summary" desc="One-page human-readable summary (Markdown)." btn="Export .md" />
          <Action icon={<I.copy size={19} />} title="Copy diagnostic summary" desc="Three-line summary for tickets and chat." btn={copied ? 'Copied ✓' : 'Copy'} onClick={copy} />
        </div>

        <Card eyebrow="Preview" title="Diagnostic summary" style={{ marginTop: 4 }}>
          <pre style={{ margin: 0, fontFamily: 'var(--font-mono)', fontSize: 12, color: 'var(--fg-2)', lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>{summary}</pre>
        </Card>

        <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: 16, background: 'var(--bg-sunken)', border: '1px dashed var(--hairline-strong)', borderRadius: 'var(--radius-lg)', opacity: 0.85 }}>
          <Badge tone="neutral" appearance="outline" size="sm" uppercase mono>Planned</Badge>
          <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-3)' }}>Compare two scan reports — not yet implemented. This placeholder is intentionally non-functional.</span>
        </div>
      </div>
    );
  }

  function SettingRow({ label, desc, children }) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', gap: 16, padding: '14px 0', borderBottom: '1px solid var(--hairline)' }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 500 }}>{label}</div>
          {desc && <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', marginTop: 2 }}>{desc}</div>}
        </div>
        <div style={{ flex: '0 0 auto' }}>{children}</div>
      </div>
    );
  }

  function Settings({ scan }) {
    const [s, setS] = useState({ theme: 'dark', engine: 'elk_layered', lowconf: true, unknown: true, isp: true, persist: true });
    const set = (k, v) => setS((p) => ({ ...p, [k]: v }));

    return (
      <div style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 18, maxWidth: 720, margin: '0 auto' }}>
        <h1 style={{ font: 'var(--type-h1)' }}>Settings</h1>

        <Card eyebrow="Appearance" title="Theme & color">
          <SettingRow label="Theme" desc="Dark is the default instrument surface.">
            <SegmentedControl size="sm" value={s.theme} onChange={(v) => set('theme', v)} options={[{ value: 'dark', label: 'Dark' }, { value: 'light', label: 'Light' }]} />
          </SettingRow>
          <SettingRow label="Color mode" desc="Monochrome with status-only color. Locked in this build.">
            <Badge tone="neutral" appearance="outline" size="sm" mono>black_white</Badge>
          </SettingRow>
        </Card>

        <Card eyebrow="Topology" title="Map & layout">
          <SettingRow label="Layout engine" desc="ELK layered is recommended for hierarchy clarity.">
            <SegmentedControl size="sm" value={s.engine} onChange={(v) => set('engine', v)} options={[{ value: 'elk_layered', label: 'Layered' }, { value: 'force', label: 'Force' }, { value: 'manual', label: 'Manual' }]} />
          </SettingRow>
          <SettingRow label="Show low-confidence edges" desc="Render edges below the 0.45 band, visually muted.">
            <Toggle checked={s.lowconf} onChange={(v) => set('lowconf', v)} />
          </SettingRow>
          <SettingRow label="Show unknown L2 segments" desc="Inferred switches and the hosts hidden behind them.">
            <Toggle checked={s.unknown} onChange={(v) => set('unknown', v)} />
          </SettingRow>
          <SettingRow label="Show ISP route context" desc="Gateway chain hops beyond the home network.">
            <Toggle checked={s.isp} onChange={(v) => set('isp', v)} />
          </SettingRow>
          <SettingRow label="Persist node positions" desc="Save manual layout to local UI state — never to scan data.">
            <Toggle checked={s.persist} onChange={(v) => set('persist', v)} />
          </SettingRow>
          <SettingRow label="Reset UI layout positions" desc="Restore the generated layout. Does not touch evidence.">
            <Button size="sm" variant="danger">Reset layout</Button>
          </SettingRow>
        </Card>

        <Card eyebrow="Safety" title="Data integrity">
          <SettingRow label="Safe mode" desc="Read-only topology. Scan evidence is immutable; no destructive actions.">
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}><StatusDot tone="success" /><span style={{ fontSize: 'var(--text-sm)', color: 'var(--ok)', fontWeight: 600 }}>Enabled</span></div>
          </SettingRow>
        </Card>
      </div>
    );
  }

  window.IAD_SCREENS = window.IAD_SCREENS || {};
  window.IAD_SCREENS.reports = Reports;
  window.IAD_SCREENS.settings = Settings;
})();
