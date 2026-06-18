import React from 'react';

/**
 * ProbeStatusBadge — a fixed, opinionated badge for the IAD probe-status enum.
 * Each status has a locked tone + glyph so the same state always looks the same
 * across Evidence, Devices, and Reports. Statuses:
 *   success · partial · no_data · skipped · failed · blocked
 */
const MAP = {
  success: { word: 'Success', c: 'var(--ok)', bg: 'var(--ok-bg)' },
  partial: { word: 'Partial', c: 'var(--partial)', bg: 'var(--partial-bg)' },
  no_data: { word: 'No data', c: 'var(--neutral)', bg: 'var(--neutral-bg)' },
  skipped: { word: 'Skipped', c: 'var(--neutral)', bg: 'var(--neutral-bg)' },
  failed: { word: 'Failed', c: 'var(--danger)', bg: 'var(--danger-bg)' },
  blocked: { word: 'Blocked', c: 'var(--blocked)', bg: 'var(--blocked-bg)' },
};

export function ProbeStatusBadge({ status = 'no_data', size = 'md', style = {}, ...rest }) {
  const m = MAP[status] || MAP.no_data;
  const sz = size === 'sm' ? { h: 18, px: 7, dot: 5, fs: 'var(--text-2xs)' } : { h: 22, px: 9, dot: 6, fs: 'var(--text-xs)' };
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 6,
        height: sz.h,
        padding: `0 ${sz.px}px`,
        borderRadius: 'var(--radius-xs)',
        background: m.bg,
        color: m.c,
        fontFamily: 'var(--font-mono)',
        fontSize: sz.fs,
        fontWeight: 'var(--fw-semibold)',
        letterSpacing: 'var(--ls-wide)',
        textTransform: 'uppercase',
        whiteSpace: 'nowrap',
        ...style,
      }}
      {...rest}
    >
      <span style={{ width: sz.dot, height: sz.dot, borderRadius: '50%', background: m.c, flex: '0 0 auto' }} />
      {m.word}
    </span>
  );
}
