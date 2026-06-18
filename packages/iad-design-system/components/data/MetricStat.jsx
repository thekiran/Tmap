import React from 'react';

/**
 * MetricStat — a labeled value readout. The value uses the mono family with
 * tabular numerals so columns of stats align. Optional unit, secondary line,
 * and a small delta/qualifier. The workhorse of the dashboard cards.
 */
export function MetricStat({
  label,
  value,
  unit = null,
  secondary = null,
  tone = 'default',
  size = 'md',
  align = 'left',
  style = {},
  ...rest
}) {
  const tones = {
    default: 'var(--fg-1)',
    accent: 'var(--accent-bright)',
    success: 'var(--ok)',
    warn: 'var(--warn)',
    danger: 'var(--danger)',
    muted: 'var(--fg-3)',
  };
  const c = tones[tone] || tones.default;

  const sizes = {
    sm: { v: 'var(--text-md)', l: 'var(--text-2xs)' },
    md: { v: 'var(--text-xl)', l: 'var(--text-xs)' },
    lg: { v: 'var(--text-2xl)', l: 'var(--text-xs)' },
  };
  const sz = sizes[size] || sizes.md;

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 4,
        alignItems: align === 'right' ? 'flex-end' : 'flex-start',
        textAlign: align,
        minWidth: 0,
        ...style,
      }}
      {...rest}
    >
      <span
        style={{
          font: 'var(--type-overline)',
          letterSpacing: 'var(--ls-caps)',
          textTransform: 'uppercase',
          color: 'var(--fg-3)',
          fontSize: sz.l,
        }}
      >
        {label}
      </span>
      <span style={{ display: 'inline-flex', alignItems: 'baseline', gap: 5, minWidth: 0 }}>
        <span
          className="iad-num"
          style={{
            fontFamily: 'var(--font-mono)',
            fontVariantNumeric: 'tabular-nums',
            fontWeight: 'var(--fw-semibold)',
            fontSize: sz.v,
            color: c,
            lineHeight: 1.1,
            letterSpacing: 'var(--ls-snug)',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {value}
        </span>
        {unit && (
          <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)', color: 'var(--fg-3)' }}>{unit}</span>
        )}
      </span>
      {secondary && (
        <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.4 }}>{secondary}</span>
      )}
    </div>
  );
}
