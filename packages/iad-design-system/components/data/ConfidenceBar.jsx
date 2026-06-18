import React from 'react';

/**
 * ConfidenceBar — the headline "how sure are we" control. Renders a labeled
 * track filled to a 0–1 confidence with the band color (Low/Med/High), plus
 * the numeric percentage and band word. Honest by design: low confidence is
 * calm gray, never red. Bands follow the IAD calibration:
 *   < 0.45 Low · 0.45–0.75 Medium · ≥ 0.75 High
 */
export function band(value) {
  if (value >= 0.75) return 'high';
  if (value >= 0.45) return 'medium';
  return 'low';
}

export function ConfidenceBar({
  value = 0,
  showLabel = true,
  showValue = true,
  label = 'Confidence',
  size = 'md',
  style = {},
  ...rest
}) {
  const v = Math.max(0, Math.min(1, value));
  const b = band(v);
  const colors = {
    low: { fg: 'var(--conf-low)', bg: 'var(--conf-low-bg)', word: 'Low' },
    medium: { fg: 'var(--conf-med)', bg: 'var(--conf-med-bg)', word: 'Medium' },
    high: { fg: 'var(--conf-high)', bg: 'var(--conf-high-bg)', word: 'High' },
  };
  const c = colors[b];
  const heights = { sm: 5, md: 7, lg: 9 };
  const h = heights[size] || heights.md;
  const pct = Math.round(v * 100);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6, minWidth: 0, ...style }} {...rest}>
      {(showLabel || showValue) && (
        <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between', gap: 8 }}>
          {showLabel && (
            <span
              style={{
                font: 'var(--type-overline)',
                letterSpacing: 'var(--ls-caps)',
                textTransform: 'uppercase',
                color: 'var(--fg-3)',
              }}
            >
              {label}
            </span>
          )}
          {showValue && (
            <span style={{ display: 'inline-flex', alignItems: 'baseline', gap: 6 }}>
              <span
                style={{
                  fontFamily: 'var(--font-mono)',
                  fontVariantNumeric: 'tabular-nums',
                  fontWeight: 'var(--fw-semibold)',
                  fontSize: 'var(--text-sm)',
                  color: c.fg,
                }}
              >
                {pct}%
              </span>
              <span style={{ fontSize: 'var(--text-2xs)', color: c.fg, textTransform: 'uppercase', letterSpacing: 'var(--ls-caps)', fontWeight: 'var(--fw-semibold)' }}>
                {c.word}
              </span>
            </span>
          )}
        </div>
      )}
      <div
        role="meter"
        aria-valuenow={pct}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-label={`${label}: ${pct}% (${c.word})`}
        style={{ height: h, borderRadius: 'var(--radius-pill)', background: 'var(--surface-3)', overflow: 'hidden' }}
      >
        <div
          style={{
            width: `${pct}%`,
            height: '100%',
            background: c.fg,
            borderRadius: 'var(--radius-pill)',
            transition: 'width var(--dur-meter) var(--ease-out)',
          }}
        />
      </div>
    </div>
  );
}
