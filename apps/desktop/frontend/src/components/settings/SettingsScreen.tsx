import type { ReactNode } from 'react';
import { useUIStore } from '../../store/useUIStore';
import { useScanStore } from '../../store/useScanStore';
import { Card, Button, Badge, StatusDot } from '../ui/primitives';
import { Toggle, SegmentedControl } from '../ui/data';
import type { LayoutEngine, ThemeMode } from '../../lib/models';

function Row({ label, desc, children }: { label: string; desc?: string; children: ReactNode }) {
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

export function SettingsScreen() {
  const s = useUIStore((st) => st.settings);
  const update = useUIStore((st) => st.updateSettings);
  const setTheme = useUIStore((st) => st.setTheme);
  const resetLayout = useUIStore((st) => st.resetLayoutPositions);
  const safeMode = useScanStore((st) => st.normalized?.safeMode ?? true);

  return (
    <div className="app-scroll" style={{ flex: 1, overflow: 'auto' }}>
      <div style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 18, maxWidth: 720, margin: '0 auto' }}>
        <h1 style={{ font: 'var(--type-h1)' }}>Settings</h1>

        <Card eyebrow="Appearance" title="Theme & color">
          <Row label="Theme" desc="Dark is the default instrument surface.">
            <SegmentedControl size="sm" value={s.theme} onChange={(v: ThemeMode) => setTheme(v)} options={[{ value: 'dark', label: 'Dark' }, { value: 'light', label: 'Light' }]} />
          </Row>
          <Row label="Color mode" desc="Monochrome with status-only color. Locked in this build.">
            <Badge tone="neutral" appearance="outline" size="sm" mono>black_white</Badge>
          </Row>
        </Card>

        <Card eyebrow="Topology" title="Map & layout">
          <Row label="Layout engine" desc="ELK layered is recommended for hierarchy clarity.">
            <SegmentedControl size="sm" value={s.layoutEngine} onChange={(v: LayoutEngine) => update({ layoutEngine: v })} options={[{ value: 'elk_layered', label: 'Layered' }, { value: 'force', label: 'Force' }, { value: 'manual', label: 'Manual' }]} />
          </Row>
          <Row label="Show low-confidence edges" desc="Render edges below the 0.45 band, visually muted.">
            <Toggle checked={s.showLowConfidenceEdges} onChange={(v) => update({ showLowConfidenceEdges: v })} />
          </Row>
          <Row label="Show unknown L2 segments" desc="Inferred switches and the hosts hidden behind them.">
            <Toggle checked={s.showUnknownSegments} onChange={(v) => update({ showUnknownSegments: v })} />
          </Row>
          <Row label="Show ISP route context" desc="Gateway-chain hops beyond the home network.">
            <Toggle checked={s.showIspRouteContext} onChange={(v) => update({ showIspRouteContext: v })} />
          </Row>
          <Row label="Persist node positions" desc="Save manual layout to local UI state — never to scan data.">
            <Toggle checked={s.persistNodePositions} onChange={(v) => update({ persistNodePositions: v })} />
          </Row>
          <Row label="Reset UI layout positions" desc="Restore the generated layout. Does not touch evidence.">
            <Button size="sm" variant="danger" onClick={resetLayout}>Reset layout</Button>
          </Row>
        </Card>

        <Card eyebrow="Safety" title="Data integrity">
          <Row label="Safe mode" desc="Read-only topology. Scan evidence is immutable; no destructive actions.">
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <StatusDot tone={safeMode ? 'success' : 'warn'} />
              <span style={{ fontSize: 'var(--text-sm)', color: safeMode ? 'var(--ok)' : 'var(--warn)', fontWeight: 600 }}>{safeMode ? 'Enabled' : 'Off'}</span>
            </div>
          </Row>
        </Card>
      </div>
    </div>
  );
}
