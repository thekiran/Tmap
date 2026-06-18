import React from 'react';

/**
 * StatusDot — a small filled dot encoding a state, optionally with a label.
 * Used for reachability, probe status, online/offline, layer legends.
 * The "pulse" option adds a calm breathing ring for live/active states only.
 */
export function StatusDot({
  tone = 'neutral',
  size = 8,
  pulse = false,
  label = null,
  style = {},
  ...rest
}) {
  const tones = {
    neutral: 'var(--neutral)',
    success: 'var(--ok)',
    warn: 'var(--warn)',
    danger: 'var(--danger)',
    info: 'var(--info)',
    accent: 'var(--accent-base)',
    blocked: 'var(--blocked)',
  };
  const c = tones[tone] || tones.neutral;

  const dot = (
    <span style={{ position: 'relative', display: 'inline-flex', width: size, height: size, flex: '0 0 auto' }}>
      {pulse && (
        <span
          style={{
            position: 'absolute',
            inset: 0,
            borderRadius: '50%',
            background: c,
            opacity: 0.5,
            animation: 'iad-dot-pulse 1.8s var(--ease-out) infinite',
          }}
        />
      )}
      <span style={{ width: size, height: size, borderRadius: '50%', background: c, position: 'relative' }} />
      <style>{'@keyframes iad-dot-pulse{0%{transform:scale(1);opacity:.5}70%{transform:scale(2.4);opacity:0}100%{opacity:0}}'}</style>
    </span>
  );

  if (label == null) return React.cloneElement(dot, { ...rest, style: { ...dot.props.style, ...style } });

  return (
    <span
      style={{ display: 'inline-flex', alignItems: 'center', gap: 7, color: 'var(--fg-2)', fontSize: 'var(--text-sm)', ...style }}
      {...rest}
    >
      {dot}
      <span>{label}</span>
    </span>
  );
}
