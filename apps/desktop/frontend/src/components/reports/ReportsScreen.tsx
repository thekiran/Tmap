import { useState, type ReactNode } from 'react';
import { useScanStore } from '../../store/useScanStore';
import { Card, Button, Badge } from '../ui/primitives';
import { Icons } from '../icons/Icon';
import { useImport } from '../../lib/useImport';
import { pct } from '../../lib/format';
import { bandWord } from '../../lib/confidence';

async function save(name: string, content: string) {
  if (window.go?.main?.App?.SaveExport) {
    await window.go.main.App.SaveExport(name, content);
    return;
  }
  // Browser fallback download
  const blob = new Blob([content], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url; a.download = name; a.click();
  URL.revokeObjectURL(url);
}

export function ReportsScreen() {
  const raw = useScanStore((s) => s.raw);
  const scan = useScanStore((s) => s.normalized)!;
  const { importViaDialog } = useImport();
  const [copied, setCopied] = useState(false);

  const summary = [
    `IAD scan ${scan.scanId} — ${scan.isUnknown ? 'Unknown access type' : scan.primaryType} (${pct(scan.confidence)} ${bandWord(scan.confidence)} confidence, ${scan.decisionQuality} quality)`,
    scan.publicIp ? `ISP ${scan.publicIp.asn ?? '?'} ${scan.publicIp.org ?? ''} · Public IP ${scan.publicIp.address ?? '?'}` : 'No public IP resolved',
    `${scan.nat ? `NAT: ${(scan.nat.type ?? 'nat').toUpperCase()} (${scan.nat.layers ?? '?'} layers)` : 'NAT: unknown'} · IPv6 ${scan.ipv6?.available ? 'native' : 'none'} · ${scan.devices.length} LAN devices`,
  ].join('\n');

  const summaryMd = `# IAD scan ${scan.scanId}\n\n- **Access type:** ${scan.isUnknown ? 'Unknown' : scan.primaryType}\n- **Confidence:** ${pct(scan.confidence)} (${bandWord(scan.confidence)}), decision quality ${scan.decisionQuality}\n- **ISP:** ${scan.publicIp?.asn ?? '?'} ${scan.publicIp?.org ?? ''}\n- **Public IP:** ${scan.publicIp?.address ?? '?'}\n- **NAT:** ${scan.nat?.type ?? 'unknown'}\n- **Devices:** ${scan.devices.length}\n\n## Uncertainty\n${scan.uncertaintyReasons.map((r) => `- ${r}`).join('\n')}\n`;

  const copy = () => { navigator.clipboard?.writeText(summary).catch(() => {}); setCopied(true); setTimeout(() => setCopied(false), 1600); };

  const Action = ({ icon, title, desc, btn, onClick, primary }: { icon: ReactNode; title: string; desc: string; btn: string; onClick: () => void; primary?: boolean }) => (
    <div style={{ display: 'flex', alignItems: 'center', gap: 14, padding: 16, background: 'var(--surface-card)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-lg)' }}>
      <span style={{ width: 38, height: 38, borderRadius: 'var(--radius-md)', background: 'var(--surface-3)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--fg-2)', flex: '0 0 auto' }}>{icon}</span>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 600 }}>{title}</div>
        <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>{desc}</div>
      </div>
      <Button size="sm" variant={primary ? 'primary' : 'secondary'} onClick={onClick}>{btn}</Button>
    </div>
  );

  return (
    <div className="app-scroll" style={{ flex: 1, overflow: 'auto' }}>
      <div style={{ padding: 22, display: 'flex', flexDirection: 'column', gap: 14, maxWidth: 860, margin: '0 auto' }}>
        <h1 style={{ font: 'var(--type-h1)' }}>Reports</h1>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          <Action icon={<Icons.upload size={19} />} title="Import scan JSON" desc="Validated against the schema (Zod) before loading." btn="Choose file" onClick={importViaDialog} />
          <Action icon={<Icons.download size={19} />} title="Export full report" desc="The immutable scan evidence as JSON. UI layout positions are excluded." btn="Export JSON" primary onClick={() => save(`${scan.scanId}.json`, JSON.stringify(raw, null, 2))} />
          <Action icon={<Icons.reports size={19} />} title="Export scan summary" desc="One-page human-readable summary (Markdown)." btn="Export .md" onClick={() => save(`${scan.scanId}-summary.md`, summaryMd)} />
          <Action icon={<Icons.copy size={19} />} title="Copy diagnostic summary" desc="Three-line summary for tickets and chat." btn={copied ? 'Copied ✓' : 'Copy'} onClick={copy} />
        </div>

        <Card eyebrow="Preview" title="Diagnostic summary" style={{ marginTop: 4 }}>
          <pre style={{ margin: 0, fontFamily: 'var(--font-mono)', fontSize: 12, color: 'var(--fg-2)', lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>{summary}</pre>
        </Card>

        <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: 16, background: 'var(--bg-sunken)', border: '1px dashed var(--hairline-strong)', borderRadius: 'var(--radius-lg)', opacity: 0.85 }}>
          <Badge tone="neutral" appearance="outline" size="sm" uppercase mono>Planned</Badge>
          <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-3)' }}>Compare two scan reports — not yet implemented. This placeholder is intentionally non-functional.</span>
        </div>
      </div>
    </div>
  );
}
