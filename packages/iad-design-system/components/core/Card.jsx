import React from 'react';

/**
 * Card — the primary container surface. Optional header (title + eyebrow +
 * actions) and footer. Padding and emphasis are tunable. Depth on dark comes
 * from hairline borders + faint ambient shadow, never heavy drop shadows.
 */
export function Card({
  title = null,
  eyebrow = null,
  actions = null,
  footer = null,
  padding = 'md',
  raised = false,
  interactive = false,
  children,
  style = {},
  bodyStyle = {},
  ...rest
}) {
  const pads = { none: 0, sm: 'var(--space-3)', md: 'var(--space-5)', lg: 'var(--space-6)' };
  const p = pads[padding] != null ? pads[padding] : pads.md;

  return (
    <section
      style={{
        background: raised ? 'var(--surface-2)' : 'var(--surface-card)',
        border: '1px solid var(--hairline)',
        borderRadius: 'var(--radius-lg)',
        boxShadow: raised ? 'var(--shadow-sm)' : 'var(--shadow-xs)',
        display: 'flex',
        flexDirection: 'column',
        minWidth: 0,
        transition: interactive
          ? 'border-color var(--dur-fast) var(--ease-out), background var(--dur-fast) var(--ease-out)'
          : 'none',
        ...style,
      }}
      {...rest}
    >
      {(title || eyebrow || actions) && (
        <header
          style={{
            display: 'flex',
            alignItems: 'flex-start',
            justifyContent: 'space-between',
            gap: 'var(--space-3)',
            padding: `var(--space-4) ${typeof p === 'number' ? p : p} var(--space-3)`,
            paddingBottom: 'var(--space-3)',
            borderBottom: '1px solid var(--hairline)',
          }}
        >
          <div style={{ minWidth: 0 }}>
            {eyebrow && (
              <div
                style={{
                  font: 'var(--type-overline)',
                  letterSpacing: 'var(--ls-caps)',
                  textTransform: 'uppercase',
                  color: 'var(--fg-3)',
                  marginBottom: 4,
                }}
              >
                {eyebrow}
              </div>
            )}
            {title && (
              <h3 style={{ font: 'var(--type-h3)', color: 'var(--fg-1)', margin: 0 }}>{title}</h3>
            )}
          </div>
          {actions && <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', flex: '0 0 auto' }}>{actions}</div>}
        </header>
      )}
      <div style={{ padding: p, minWidth: 0, flex: 1, ...bodyStyle }}>{children}</div>
      {footer && (
        <footer
          style={{
            padding: `var(--space-3) ${typeof p === 'number' ? p : p}`,
            borderTop: '1px solid var(--hairline)',
            color: 'var(--fg-3)',
            fontSize: 'var(--text-xs)',
          }}
        >
          {footer}
        </footer>
      )}
    </section>
  );
}
