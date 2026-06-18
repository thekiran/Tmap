import { useImport } from '../../lib/useImport';
import { Button } from '../ui/primitives';
import { Icons } from '../icons/Icon';

/** Empty state shown when no scan is loaded. Offers import, agent run, and a
 *  clearly-labelled demo (the only path that loads the dev-only sample). */
export function EmptyState() {
  const { importViaDialog, runAgent, errors, busy } = useImport();
  return (
    <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 40 }}>
      <div style={{ maxWidth: 460, textAlign: 'center', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 18 }}>
        <div style={{ width: 64, height: 64, borderRadius: 'var(--radius-lg)', background: 'var(--surface-2)', border: '1px solid var(--hairline)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--fg-3)' }}>
          <Icons.topology size={30} />
        </div>
        <div>
          <h2 style={{ font: 'var(--type-h2)', marginBottom: 8 }}>No scan loaded</h2>
          <p style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-3)', lineHeight: 1.6 }}>
            Import an authorized <code style={{ fontFamily: 'var(--font-mono)' }}>iad-agent</code> scan report (JSON), or run the agent if it is installed. The report is validated before anything renders.
          </p>
        </div>
        <div style={{ display: 'flex', gap: 10 }}>
          <Button variant="primary" iconLeft={<Icons.upload size={15} />} onClick={importViaDialog} disabled={busy}>Import scan JSON</Button>
          <Button variant="secondary" iconLeft={<Icons.refresh size={15} />} onClick={() => runAgent('standard')} disabled={busy}>Run agent</Button>
        </div>
        {errors && (
          <div style={{ width: '100%', textAlign: 'left', padding: 12, background: 'var(--danger-bg)', borderRadius: 'var(--radius-md)', border: '1px solid var(--danger)' }}>
            <div style={{ fontSize: 'var(--text-xs)', fontWeight: 600, color: 'var(--danger)', marginBottom: 6 }}>Import failed</div>
            {errors.slice(0, 6).map((e, i) => (
              <div key={i} style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--fg-2)' }}>
                <span style={{ color: 'var(--fg-4)' }}>{e.path}:</span> {e.message}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
