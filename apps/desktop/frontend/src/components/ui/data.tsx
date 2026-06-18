import type { CSSProperties, ReactNode } from 'react';
import { band, bandColorVar, bandWord } from '../../lib/confidence';
import type { ProbeStatus, EvidenceClass } from '../../lib/scan-schema';
import { Overline } from './primitives';

/* Data-display primitives: confidence meter, metric readout, probe-status and
   tier badges. Ported from the design system, typed for the app. */

/* --------------------------------------------------------- ConfidenceBar -- */
export function ConfidenceBar({ value, label = 'Confidence', showLabel = true, showValue = true, size = 'md', style }: {
  value: number; label?: string; showLabel?: boolean; showValue?: boolean;
  size?: 'sm' | 'md' | 'lg'; style?: CSSProperties;
}) {
  const v = Math.max(0, Math.min(1, value));
  const color = bandColorVar(v);
  const h = { sm: 5, md: 7, lg: 9 }[size];
  const p = Math.round(v * 100);
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6, minWidth: 0, ...style }}>
      {(showLabel || showValue) && (
        <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', gap: 8 }}>
          {showLabel && <Overline>{label}</Overline>}
          {showValue && (
            <span style={{ display: 'inline-flex', alignItems: 'baseline', gap: 6 }}>
              <span style={{ fontFamily: 'var(--font-mono)', fontVariantNumeric: 'tabular-nums', fontWeight: 'var(--fw-semibold)', fontSize: 'var(--text-sm)', color }}>{p}%</span>
              <span style={{ fontSize: 'var(--text-2xs)', color, textTransform: 'uppercase', letterSpacing: 'var(--ls-caps)', fontWeight: 'var(--fw-semibold)' }}>{bandWord(v)}</span>
            </span>
          )}
        </div>
      )}
      <div role="meter" aria-valuenow={p} aria-valuemin={0} aria-valuemax={100} aria-label={`${label}: ${p}% ${band(v)}`}
        style={{ height: h, borderRadius: 'var(--radius-pill)', background: 'var(--surface-3)', overflow: 'hidden' }}>
        <div style={{ width: `${p}%`, height: '100%', background: color, borderRadius: 'var(--radius-pill)', transition: 'width var(--dur-meter) var(--ease-out)' }} />
      </div>
    </div>
  );
}

/* ----------------------------------------------------------- MetricStat -- */
export function MetricStat({ label, value, unit, secondary, tone = 'default', size = 'md', align = 'left', style }: {
  label: ReactNode; value: ReactNode; unit?: ReactNode; secondary?: ReactNode;
  tone?: 'default' | 'accent' | 'success' | 'warn' | 'danger' | 'muted';
  size?: 'sm' | 'md' | 'lg'; align?: 'left' | 'right'; style?: CSSProperties;
}) {
  const c = { default: 'var(--fg-1)', accent: 'var(--accent-bright)', success: 'var(--ok)', warn: 'var(--warn)', danger: 'var(--danger)', muted: 'var(--fg-3)' }[tone];
  const sz = { sm: 'var(--text-md)', md: 'var(--text-xl)', lg: 'var(--text-2xl)' }[size];
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4, alignItems: align === 'right' ? 'flex-end' : 'flex-start', textAlign: align, minWidth: 0, ...style }}>
      <Overline>{label}</Overline>
      <span style={{ display: 'inline-flex', alignItems: 'baseline', gap: 5, minWidth: 0, maxWidth: '100%' }}>
        <span style={{ fontFamily: 'var(--font-mono)', fontVariantNumeric: 'tabular-nums', fontWeight: 'var(--fw-semibold)', fontSize: sz, color: c, lineHeight: 1.1, letterSpacing: 'var(--ls-snug)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{value}</span>
        {unit && <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>{unit}</span>}
      </span>
      {secondary && <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.4 }}>{secondary}</span>}
    </div>
  );
}

/* ------------------------------------------------------ ProbeStatusBadge -- */
const PROBE: Record<ProbeStatus, { word: string; c: string; bg: string }> = {
  success: { word: 'Success', c: 'var(--ok)', bg: 'var(--ok-bg)' },
  partial: { word: 'Partial', c: 'var(--partial)', bg: 'var(--partial-bg)' },
  no_data: { word: 'No data', c: 'var(--neutral)', bg: 'var(--neutral-bg)' },
  skipped: { word: 'Skipped', c: 'var(--neutral)', bg: 'var(--neutral-bg)' },
  failed: { word: 'Failed', c: 'var(--danger)', bg: 'var(--danger-bg)' },
  blocked: { word: 'Blocked', c: 'var(--blocked)', bg: 'var(--blocked-bg)' },
  completed: { word: 'Completed', c: 'var(--ok)', bg: 'var(--ok-bg)' },
};
export function ProbeStatusBadge({ status, size = 'md', style }: { status: ProbeStatus; size?: 'sm' | 'md'; style?: CSSProperties }) {
  const m = PROBE[status];
  const sz = size === 'sm' ? { h: 18, px: 7, dot: 5, fs: 'var(--text-2xs)' } : { h: 22, px: 9, dot: 6, fs: 'var(--text-xs)' };
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, height: sz.h, padding: `0 ${sz.px}px`, borderRadius: 'var(--radius-xs)', background: m.bg, color: m.c, fontFamily: 'var(--font-mono)', fontSize: sz.fs, fontWeight: 'var(--fw-semibold)', letterSpacing: 'var(--ls-wide)', textTransform: 'uppercase', whiteSpace: 'nowrap', ...style }}>
      <span style={{ width: sz.dot, height: sz.dot, borderRadius: '50%', background: m.c, flex: '0 0 auto' }} />{m.word}
    </span>
  );
}

/* -------------------------------------------------------------- TierBadge -- */
const TIERS: Record<string, { word: string; c: string; bg: string }> = {
  physical: { word: 'Physical', c: 'var(--tier-physical)', bg: 'var(--tier-physical-bg)' },
  l2: { word: 'L2 · Link', c: 'var(--tier-l2)', bg: 'var(--tier-l2-bg)' },
  l3: { word: 'L3 · Routing', c: 'var(--tier-l3)', bg: 'var(--tier-l3-bg)' },
  nat: { word: 'NAT', c: 'var(--tier-nat)', bg: 'var(--tier-nat-bg)' },
  isp: { word: 'ISP Route', c: 'var(--tier-isp)', bg: 'var(--tier-isp-bg)' },
  performance: { word: 'Performance', c: 'var(--tier-performance, var(--warn))', bg: 'var(--tier-performance-bg, var(--warn-bg))' },
};
export function TierBadge({ tier, label, appearance = 'subtle', style }: {
  tier: EvidenceClass | string; label?: ReactNode; appearance?: 'subtle' | 'solid' | 'dot'; style?: CSSProperties;
}) {
  const t = TIERS[tier] ?? TIERS.l2;
  if (appearance === 'dot') {
    return (
      <span style={{ display: 'inline-flex', alignItems: 'center', gap: 7, color: 'var(--fg-2)', fontSize: 'var(--text-sm)', ...style }}>
        <span style={{ width: 9, height: 9, borderRadius: 3, background: t.c, flex: '0 0 auto' }} />{label ?? t.word}
      </span>
    );
  }
  const look = appearance === 'solid' ? { background: t.c, color: 'var(--fg-on-accent)' } : { background: t.bg, color: t.c };
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', height: 22, padding: '0 9px', borderRadius: 'var(--radius-xs)', fontFamily: 'var(--font-mono)', fontSize: 'var(--text-2xs)', fontWeight: 'var(--fw-semibold)', letterSpacing: 'var(--ls-wide)', textTransform: 'uppercase', whiteSpace: 'nowrap', ...look, ...style }}>
      {label ?? t.word}
    </span>
  );
}

/* --------------------------------------------------------------- Toggle -- */
export function Toggle({ checked, onChange, disabled, size = 'md', style }: {
  checked: boolean; onChange: (next: boolean) => void; disabled?: boolean; size?: 'sm' | 'md'; style?: CSSProperties;
}) {
  const d = size === 'sm' ? { w: 32, h: 18, k: 12 } : { w: 38, h: 22, k: 16 };
  const pad = (d.h - d.k) / 2;
  return (
    <button type="button" role="switch" aria-checked={checked} disabled={disabled}
      onClick={() => !disabled && onChange(!checked)}
      style={{ position: 'relative', width: d.w, height: d.h, flex: '0 0 auto', borderRadius: 'var(--radius-pill)', border: '1px solid ' + (checked ? 'var(--accent-base)' : 'var(--hairline-strong)'), background: checked ? 'var(--accent-base)' : 'var(--surface-3)', cursor: disabled ? 'not-allowed' : 'pointer', opacity: disabled ? 0.45 : 1, padding: 0, transition: 'background var(--dur-base) var(--ease-out)', ...style }}>
      <span style={{ position: 'absolute', top: pad, left: checked ? d.w - d.k - pad - 1 : pad, width: d.k, height: d.k, borderRadius: '50%', background: checked ? 'var(--fg-on-accent)' : 'var(--fg-2)', transition: 'left var(--dur-base) var(--ease-out)' }} />
    </button>
  );
}

/* ----------------------------------------------------- SegmentedControl -- */
export function SegmentedControl<T extends string>({ options, value, onChange, size = 'md', fullWidth, style }: {
  options: { value: T; label: ReactNode }[]; value: T; onChange: (v: T) => void;
  size?: 'sm' | 'md'; fullWidth?: boolean; style?: CSSProperties;
}) {
  const h = size === 'sm' ? 28 : 34;
  return (
    <div role="tablist" style={{ display: 'inline-flex', width: fullWidth ? '100%' : 'auto', padding: 3, gap: 2, background: 'var(--surface-3)', border: '1px solid var(--hairline)', borderRadius: 'var(--radius-md)', ...style }}>
      {options.map((opt) => {
        const on = opt.value === value;
        return (
          <button key={opt.value} type="button" role="tab" aria-selected={on} onClick={() => onChange(opt.value)}
            style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 6, flex: fullWidth ? 1 : '0 0 auto', height: h, padding: '0 12px', borderRadius: 'var(--radius-sm)', border: 'none', cursor: 'pointer', fontFamily: 'var(--font-sans)', fontSize: size === 'sm' ? 'var(--text-xs)' : 'var(--text-sm)', fontWeight: 'var(--fw-medium)', whiteSpace: 'nowrap', background: on ? 'var(--surface-1)' : 'transparent', color: on ? 'var(--fg-1)' : 'var(--fg-3)', boxShadow: on ? 'var(--shadow-xs)' : 'none', transition: 'background var(--dur-fast) var(--ease-out)' }}>
            {opt.label}
          </button>
        );
      })}
    </div>
  );
}

/* ---------------------------------------------------------------- Input -- */
export function Input({ value, onChange, placeholder, iconLeft, size = 'md', mono, invalid, fullWidth = true, style }: {
  value: string; onChange: (v: string) => void; placeholder?: string; iconLeft?: ReactNode;
  size?: 'sm' | 'md'; mono?: boolean; invalid?: boolean; fullWidth?: boolean; style?: CSSProperties;
}) {
  const h = size === 'sm' ? 30 : 36;
  return (
    <div style={{ display: 'inline-flex', alignItems: 'center', gap: 8, width: fullWidth ? '100%' : 'auto', height: h, padding: '0 10px', background: 'var(--bg-sunken)', border: '1px solid ' + (invalid ? 'var(--danger)' : 'var(--hairline-strong)'), borderRadius: 'var(--radius-md)', ...style }}>
      {iconLeft && <span style={{ display: 'inline-flex', width: 15, height: 15, color: 'var(--fg-3)', flex: '0 0 auto' }}>{iconLeft}</span>}
      <input value={value} onChange={(e) => onChange(e.target.value)} placeholder={placeholder}
        style={{ flex: 1, minWidth: 0, border: 'none', outline: 'none', background: 'transparent', color: 'var(--fg-1)', fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)', fontSize: size === 'sm' ? 'var(--text-xs)' : 'var(--text-sm)' }} />
    </div>
  );
}
