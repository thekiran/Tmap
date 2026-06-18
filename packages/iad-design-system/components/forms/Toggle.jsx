import React from 'react';

/**
 * Toggle — accessible on/off switch. Used in Settings and the topology layer
 * panel. Monochrome track; the accent fills only when on.
 */
export function Toggle({
  checked = false,
  onChange = () => {},
  disabled = false,
  label = null,
  description = null,
  size = 'md',
  id,
  style = {},
  ...rest
}) {
  const dims = size === 'sm' ? { w: 32, h: 18, k: 12 } : { w: 38, h: 22, k: 16 };
  const pad = (dims.h - dims.k) / 2;

  const sw = (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={typeof label === 'string' ? label : undefined}
      disabled={disabled}
      id={id}
      onClick={() => !disabled && onChange(!checked)}
      style={{
        position: 'relative',
        width: dims.w,
        height: dims.h,
        flex: '0 0 auto',
        borderRadius: 'var(--radius-pill)',
        border: '1px solid ' + (checked ? 'var(--accent-base)' : 'var(--hairline-strong)'),
        background: checked ? 'var(--accent-base)' : 'var(--surface-3)',
        cursor: disabled ? 'not-allowed' : 'pointer',
        opacity: disabled ? 0.45 : 1,
        padding: 0,
        transition: 'background var(--dur-base) var(--ease-out), border-color var(--dur-base) var(--ease-out)',
      }}
    >
      <span
        style={{
          position: 'absolute',
          top: pad,
          left: checked ? dims.w - dims.k - pad - 1 : pad,
          width: dims.k,
          height: dims.k,
          borderRadius: '50%',
          background: checked ? 'var(--fg-on-accent)' : 'var(--fg-2)',
          transition: 'left var(--dur-base) var(--ease-out), background var(--dur-base) var(--ease-out)',
        }}
      />
    </button>
  );

  if (label == null && description == null) return React.cloneElement(sw, { style: { ...sw.props.style, ...style }, ...rest });

  return (
    <label
      htmlFor={id}
      style={{ display: 'flex', alignItems: description ? 'flex-start' : 'center', gap: 'var(--space-3)', cursor: disabled ? 'not-allowed' : 'pointer', ...style }}
      {...rest}
    >
      {sw}
      <span style={{ display: 'flex', flexDirection: 'column', gap: 2, minWidth: 0 }}>
        {label && <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-1)', fontWeight: 'var(--fw-medium)' }}>{label}</span>}
        {description && <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-3)', lineHeight: 1.4 }}>{description}</span>}
      </span>
    </label>
  );
}
