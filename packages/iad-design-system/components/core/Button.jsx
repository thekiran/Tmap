import React from 'react';

/**
 * Button — primary action control for the IAD console.
 * Monochrome-first: the default ("secondary") is a neutral outline button;
 * "primary" uses the restrained accent; "ghost" is chromeless; "danger" only
 * for genuinely destructive/blocking actions (rare in a read-only tool).
 */
export function Button({
  variant = 'secondary',
  size = 'md',
  iconLeft = null,
  iconRight = null,
  disabled = false,
  fullWidth = false,
  type = 'button',
  children,
  style = {},
  ...rest
}) {
  const sizes = {
    sm: { height: 28, padding: '0 10px', font: 'var(--text-xs)', gap: 6, radius: 'var(--radius-sm)' },
    md: { height: 34, padding: '0 14px', font: 'var(--text-sm)', gap: 8, radius: 'var(--radius-md)' },
    lg: { height: 42, padding: '0 18px', font: 'var(--text-base)', gap: 8, radius: 'var(--radius-md)' },
  };
  const s = sizes[size] || sizes.md;

  const variants = {
    primary: {
      background: 'var(--accent-base)',
      color: 'var(--fg-on-accent)',
      border: '1px solid var(--accent-base)',
    },
    secondary: {
      background: 'var(--surface-2)',
      color: 'var(--fg-1)',
      border: '1px solid var(--hairline-strong)',
    },
    ghost: {
      background: 'transparent',
      color: 'var(--fg-2)',
      border: '1px solid transparent',
    },
    danger: {
      background: 'transparent',
      color: 'var(--danger)',
      border: '1px solid var(--danger-bg)',
    },
  };
  const v = variants[variant] || variants.secondary;

  return (
    <button
      type={type}
      disabled={disabled}
      data-variant={variant}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: s.gap,
        height: s.height,
        padding: s.padding,
        width: fullWidth ? '100%' : 'auto',
        font: `var(--fw-medium) ${s.font}/1 var(--font-sans)`,
        letterSpacing: '0.01em',
        borderRadius: s.radius,
        cursor: disabled ? 'not-allowed' : 'pointer',
        opacity: disabled ? 0.45 : 1,
        whiteSpace: 'nowrap',
        userSelect: 'none',
        transition: 'background var(--dur-fast) var(--ease-out), border-color var(--dur-fast) var(--ease-out), opacity var(--dur-fast) var(--ease-out)',
        ...v,
        ...style,
      }}
      {...rest}
    >
      {iconLeft && <span style={{ display: 'inline-flex', width: '1em', height: '1em' }}>{iconLeft}</span>}
      {children}
      {iconRight && <span style={{ display: 'inline-flex', width: '1em', height: '1em' }}>{iconRight}</span>}
    </button>
  );
}
