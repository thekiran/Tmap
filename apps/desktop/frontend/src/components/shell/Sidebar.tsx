import { useScanStore } from '../../store/useScanStore';
import { useUIStore, type ScreenId } from '../../store/useUIStore';
import { Icons } from '../icons/Icon';
import markUrl from '../../assets/mark-iad.svg';

const NAV: { id: ScreenId; label: string; icon: keyof typeof Icons }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: 'dashboard' },
  { id: 'topology', label: 'Topology', icon: 'topology' },
  { id: 'devices', label: 'Devices', icon: 'devices' },
  { id: 'evidence', label: 'Evidence', icon: 'evidence' },
  { id: 'reports', label: 'Reports', icon: 'reports' },
  { id: 'settings', label: 'Settings', icon: 'settings' },
];

export function Sidebar() {
  const screen = useUIStore((s) => s.screen);
  const setScreen = useUIStore((s) => s.setScreen);
  const scan = useScanStore((s) => s.normalized);
  const counts: Partial<Record<ScreenId, number>> = scan
    ? { devices: scan.devices.length, evidence: scan.evidence.length }
    : {};

  return (
    <aside style={{ width: 'var(--sidebar-w)', flex: '0 0 auto', background: 'var(--ink-900)', borderRight: '1px solid var(--hairline)', display: 'flex', flexDirection: 'column' }}>
      <div style={{ height: 'var(--topbar-h)', display: 'flex', alignItems: 'center', gap: 10, padding: '0 16px', borderBottom: '1px solid var(--hairline)' }}>
        <img src={markUrl} alt="IAD" style={{ width: 26, height: 26 }} />
        <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 700, letterSpacing: '1.5px', color: 'var(--fg-1)', fontSize: 15 }}>IAD</span>
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 10, color: 'var(--fg-4)', marginLeft: 'auto' }}>v0.9</span>
      </div>
      <nav style={{ padding: 10, display: 'flex', flexDirection: 'column', gap: 2 }}>
        {NAV.map((n) => {
          const on = n.id === screen;
          const IconC = Icons[n.icon];
          return (
            <button key={n.id} onClick={() => setScreen(n.id)}
              style={{ display: 'flex', alignItems: 'center', gap: 11, padding: '9px 11px', borderRadius: 'var(--radius-md)', border: '1px solid ' + (on ? 'var(--hairline)' : 'transparent'), background: on ? 'var(--surface-2)' : 'transparent', cursor: 'pointer', textAlign: 'left', color: on ? 'var(--fg-1)' : 'var(--fg-3)', font: '500 var(--text-sm) var(--font-sans)', transition: 'background var(--dur-fast) var(--ease-out)' }}>
              <span style={{ display: 'inline-flex', color: on ? 'var(--accent-bright)' : 'var(--fg-4)' }}><IconC size={17} /></span>
              {n.label}
              {counts[n.id] != null && <span style={{ marginLeft: 'auto', fontFamily: 'var(--font-mono)', fontSize: 10, color: 'var(--fg-4)' }}>{counts[n.id]}</span>}
            </button>
          );
        })}
      </nav>
      <div style={{ marginTop: 'auto', padding: 14, borderTop: '1px solid var(--hairline)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '8px 10px', background: 'var(--ok-bg)', borderRadius: 'var(--radius-md)' }}>
          <Icons.shield size={15} style={{ color: 'var(--ok)' }} />
          <span style={{ fontSize: 11, color: 'var(--ok)', fontWeight: 600 }}>Safe mode · read-only</span>
        </div>
      </div>
    </aside>
  );
}
