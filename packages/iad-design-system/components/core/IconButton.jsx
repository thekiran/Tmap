import React from 'react';

/**
 * IconButton — square, icon-only control. Used heavily in the topology toolbar
 * (zoom, fit, reset) and table row actions. Always pass an aria-label.
 */
export function IconButton({
  variant = 'ghost',
  size = 'md',
  active = false,
  disabled = false,
  label,
  children,
  style = {},
  ...rest
}) {
  const dims = { sm: 28, md: 34, lg: 40 };
  const d = dims[size] || dims.md;

  const base = {
    ghost: {
      background: active ? 'var(--surface-3)' : 'transparent',
      color: active ? 'var(--fg-1)' : 'var(--fg-2)',
      border: '1px solid ' + (active ? 'var(--hairline-strong)' : 'transparent'),
    },
    outline: {
      background: active ? 'var(--surface-3)' : 'var(--surface-2)',
      color: 'var(--fg-1)',
      border: '1px solid var(--hairline-strong)',
    },
  };
  const v = base[variant] || base.ghost;

  return (
    <button
      type="button"
      aria-label={label}
      aria-pressed={active || undefined}
      disabled={disabled}
      title={label}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        width: d,
        height: d,
        flex: '0 0 auto',
        borderRadius: 'var(--radius-md)',
        cursor: disabled ? 'not-allowed' : 'pointer',
        opacity: disabled ? 0.45 : 1,
        transition: 'background var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out), border-color var(--dur-fast) var(--ease-out)',
        ...v,
        ...style,
      }}
      {...rest}
    >
      <span style={{ display: 'inline-flex', width: Math.round(d * 0.46), height: Math.round(d * 0.46) }}>
        {children}
      </span>
    </button>
  );
}
