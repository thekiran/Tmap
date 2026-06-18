import React from 'react';

/**
 * TierBadge — labels an evidence tier / topology layer with its fixed hue.
 * Tiers are the categorical backbone of the product; each has ONE color used
 * everywhere (legend, edges context, evidence rows, layer toggles).
 *   physical · l2 · l3 · nat · isp
 */
const TIERS = {
  physical: { word: 'Physical', c: 'var(--tier-physical)', bg: 'var(--tier-physical-bg)' },
  l2: { word: 'L2 · Link', c: 'var(--tier-l2)', bg: 'var(--tier-l2-bg)' },
  l3: { word: 'L3 · Routing', c: 'var(--tier-l3)', bg: 'var(--tier-l3-bg)' },
  nat: { word: 'NAT', c: 'var(--tier-nat)', bg: 'var(--tier-nat-bg)' },
  isp: { word: 'ISP Route', c: 'var(--tier-isp)', bg: 'var(--tier-isp-bg)' },
};

export function TierBadge({ tier = 'l2', label = null, appearance = 'subtle', style = {}, ...rest }) {
  const t = TIERS[tier] || TIERS.l2;
  const look =
    appearance === 'solid'
      ? { background: t.c, color: 'var(--fg-on-accent)' }
      : appearance === 'dot'
      ? { background: 'transparent', color: 'var(--fg-2)', padding: 0, height: 'auto' }
      : { background: t.bg, color: t.c };

  if (appearance === 'dot') {
    return (
      <span style={{ display: 'inline-flex', alignItems: 'center', gap: 7, color: 'var(--fg-2)', fontSize: 'var(--text-sm)', ...style }} {...rest}>
        <span style={{ width: 9, height: 9, borderRadius: 3, background: t.c, flex: '0 0 auto' }} />
        {label || t.word}
      </span>
    );
  }

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 6,
        height: 22,
        padding: '0 9px',
        borderRadius: 'var(--radius-xs)',
        fontFamily: 'var(--font-mono)',
        fontSize: 'var(--text-2xs)',
        fontWeight: 'var(--fw-semibold)',
        letterSpacing: 'var(--ls-wide)',
        textTransform: 'uppercase',
        whiteSpace: 'nowrap',
        ...look,
        ...style,
      }}
      {...rest}
    >
      {label || t.word}
    </span>
  );
}
