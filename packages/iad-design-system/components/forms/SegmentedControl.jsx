import React from 'react';

/**
 * SegmentedControl — a compact single-select used for view switches
 * (Table / List), theme (Dark / Light), and layout engine (Layered / Force /
 * Manual). Options are { value, label, icon? }. Controlled.
 */
export function SegmentedControl({
  options = [],
  value,
  onChange = () => {},
  size = 'md',
  fullWidth = false,
  style = {},
  ...rest
}) {
  const h = size === 'sm' ? 28 : 34;
  const fs = size === 'sm' ? 'var(--text-xs)' : 'var(--text-sm)';
  return (
    <div
      role="tablist"
      style={{
        display: 'inline-flex',
        width: fullWidth ? '100%' : 'auto',
        padding: 3,
        gap: 2,
        background: 'var(--surface-3)',
        border: '1px solid var(--hairline)',
        borderRadius: 'var(--radius-md)',
        ...style,
      }}
      {...rest}
    >
      {options.map((opt) => {
        const selected = opt.value === value;
        return (
          <button
            key={opt.value}
            type="button"
            role="tab"
            aria-selected={selected}
            onClick={() => onChange(opt.value)}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 6,
              flex: fullWidth ? 1 : '0 0 auto',
              height: h,
              padding: '0 12px',
              borderRadius: 'var(--radius-sm)',
              border: 'none',
              cursor: 'pointer',
              fontFamily: 'var(--font-sans)',
              fontSize: fs,
              fontWeight: 'var(--fw-medium)',
              whiteSpace: 'nowrap',
              background: selected ? 'var(--surface-1)' : 'transparent',
              color: selected ? 'var(--fg-1)' : 'var(--fg-3)',
              boxShadow: selected ? 'var(--shadow-xs)' : 'none',
              transition: 'background var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out)',
            }}
          >
            {opt.icon && <span style={{ display: 'inline-flex', width: '1em', height: '1em' }}>{opt.icon}</span>}
            {opt.label}
          </button>
        );
      })}
    </div>
  );
}
