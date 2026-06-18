import React from 'react';

/**
 * Badge — small label chip for statuses, counts, and categorical tags.
 * Tone maps to the semantic token set. Use "subtle" (default) for tinted-bg
 * chips, "solid" for high-emphasis, "outline" for quiet metadata.
 */
export function Badge({
  tone = 'neutral',
  appearance = 'subtle',
  size = 'md',
  mono = false,
  uppercase = false,
  children,
  style = {},
  ...rest
}) {
  const tones = {
    neutral: { c: 'var(--fg-2)', bg: 'var(--neutral-bg)', solid: 'var(--neutral)' },
    accent:  { c: 'var(--accent-bright)', bg: 'var(--accent-ghost)', solid: 'var(--accent-base)' },
    success: { c: 'var(--ok)', bg: 'var(--ok-bg)', solid: 'var(--ok)' },
    warn:    { c: 'var(--warn)', bg: 'var(--warn-bg)', solid: 'var(--warn)' },
    danger:  { c: 'var(--danger)', bg: 'var(--danger-bg)', solid: 'var(--danger)' },
    info:    { c: 'var(--info)', bg: 'var(--info-bg)', solid: 'var(--info)' },
    blocked: { c: 'var(--blocked)', bg: 'var(--blocked-bg)', solid: 'var(--blocked)' },
  };
  const t = tones[tone] || tones.neutral;

  const sizes = {
    sm: { h: 18, px: 6, fs: 'var(--text-2xs)' },
    md: { h: 22, px: 8, fs: 'var(--text-xs)' },
  };
  const sz = sizes[size] || sizes.md;

  let look;
  if (appearance === 'solid') {
    look = { background: t.solid, color: 'var(--fg-on-accent)', border: '1px solid transparent' };
  } else if (appearance === 'outline') {
    look = { background: 'transparent', color: t.c, border: '1px solid var(--hairline-strong)' };
  } else {
    look = { background: t.bg, color: t.c, border: '1px solid transparent' };
  }

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 5,
        height: sz.h,
        padding: `0 ${sz.px}px`,
        borderRadius: 'var(--radius-xs)',
        fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)',
        fontSize: sz.fs,
        fontWeight: 'var(--fw-medium)',
        lineHeight: 1,
        letterSpacing: uppercase ? 'var(--ls-caps)' : '0.01em',
        textTransform: uppercase ? 'uppercase' : 'none',
        whiteSpace: 'nowrap',
        ...look,
        ...style,
      }}
      {...rest}
    >
      {children}
    </span>
  );
}
