import React from 'react';

/**
 * Input — text field with optional leading icon and addon. Used for search,
 * filters, and Reports import paths. Mono variant for IP/MAC/ASN entry.
 */
export function Input({
  value,
  onChange = () => {},
  placeholder = '',
  type = 'text',
  iconLeft = null,
  addonRight = null,
  size = 'md',
  mono = false,
  disabled = false,
  invalid = false,
  fullWidth = true,
  style = {},
  inputStyle = {},
  ...rest
}) {
  const h = size === 'sm' ? 30 : 36;
  return (
    <div
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 8,
        width: fullWidth ? '100%' : 'auto',
        height: h,
        padding: '0 10px',
        background: 'var(--bg-sunken)',
        border: '1px solid ' + (invalid ? 'var(--danger)' : 'var(--hairline-strong)'),
        borderRadius: 'var(--radius-md)',
        opacity: disabled ? 0.5 : 1,
        transition: 'border-color var(--dur-fast) var(--ease-out)',
        ...style,
      }}
    >
      {iconLeft && (
        <span style={{ display: 'inline-flex', width: 15, height: 15, color: 'var(--fg-3)', flex: '0 0 auto' }}>{iconLeft}</span>
      )}
      <input
        value={value}
        onChange={(e) => onChange(e.target.value, e)}
        placeholder={placeholder}
        type={type}
        disabled={disabled}
        style={{
          flex: 1,
          minWidth: 0,
          border: 'none',
          outline: 'none',
          background: 'transparent',
          color: 'var(--fg-1)',
          fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)',
          fontSize: size === 'sm' ? 'var(--text-xs)' : 'var(--text-sm)',
          ...inputStyle,
        }}
        {...rest}
      />
      {addonRight && <span style={{ display: 'inline-flex', alignItems: 'center', flex: '0 0 auto', color: 'var(--fg-3)' }}>{addonRight}</span>}
    </div>
  );
}
