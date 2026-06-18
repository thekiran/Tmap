import type { CSSProperties, ReactNode, ButtonHTMLAttributes } from 'react';

/* Primitive UI components for the IAD console, ported from the design system.
   Inline styles reference CSS custom properties so theming/dark-light is free. */

type Tone = 'neutral' | 'accent' | 'success' | 'warn' | 'danger' | 'info' | 'blocked';

const TONES: Record<Tone, { c: string; bg: string; solid: string }> = {
  neutral: { c: 'var(--fg-2)', bg: 'var(--neutral-bg)', solid: 'var(--neutral)' },
  accent: { c: 'var(--accent-bright)', bg: 'var(--accent-ghost)', solid: 'var(--accent-base)' },
  success: { c: 'var(--ok)', bg: 'var(--ok-bg)', solid: 'var(--ok)' },
  warn: { c: 'var(--warn)', bg: 'var(--warn-bg)', solid: 'var(--warn)' },
  danger: { c: 'var(--danger)', bg: 'var(--danger-bg)', solid: 'var(--danger)' },
  info: { c: 'var(--info)', bg: 'var(--info-bg)', solid: 'var(--info)' },
  blocked: { c: 'var(--blocked)', bg: 'var(--blocked-bg)', solid: 'var(--blocked)' },
};

/* --------------------------------------------------------------- Button -- */
interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  iconLeft?: ReactNode;
  iconRight?: ReactNode;
  fullWidth?: boolean;
}
export function Button({ variant = 'secondary', size = 'md', iconLeft, iconRight, fullWidth, disabled, children, style, ...rest }: ButtonProps) {
  const sizes = {
    sm: { height: 28, padding: '0 10px', font: 'var(--text-xs)' },
    md: { height: 34, padding: '0 14px', font: 'var(--text-sm)' },
    lg: { height: 42, padding: '0 18px', font: 'var(--text-base)' },
  }[size];
  const variants: Record<string, CSSProperties> = {
    primary: { background: 'var(--accent-base)', color: 'var(--fg-on-accent)', border: '1px solid var(--accent-base)' },
    secondary: { background: 'var(--surface-2)', color: 'var(--fg-1)', border: '1px solid var(--hairline-strong)' },
    ghost: { background: 'transparent', color: 'var(--fg-2)', border: '1px solid transparent' },
    danger: { background: 'transparent', color: 'var(--danger)', border: '1px solid var(--danger-bg)' },
  };
  return (
    <button
      disabled={disabled}
      style={{
        display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 8,
        height: sizes.height, padding: sizes.padding, width: fullWidth ? '100%' : 'auto',
        font: `var(--fw-medium) ${sizes.font}/1 var(--font-sans)`, letterSpacing: '0.01em',
        borderRadius: 'var(--radius-md)', cursor: disabled ? 'not-allowed' : 'pointer',
        opacity: disabled ? 0.45 : 1, whiteSpace: 'nowrap', userSelect: 'none',
        transition: 'background var(--dur-fast) var(--ease-out), border-color var(--dur-fast) var(--ease-out)',
        ...variants[variant], ...style,
      }}
      {...rest}
    >
      {iconLeft && <span style={{ display: 'inline-flex' }}>{iconLeft}</span>}
      {children}
      {iconRight && <span style={{ display: 'inline-flex' }}>{iconRight}</span>}
    </button>
  );
}

/* ----------------------------------------------------------- IconButton -- */
interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  label: string;
  variant?: 'ghost' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  active?: boolean;
}
export function IconButton({ label, variant = 'ghost', size = 'md', active, disabled, children, style, ...rest }: IconButtonProps) {
  const d = { sm: 28, md: 34, lg: 40 }[size];
  const looks = {
    ghost: { background: active ? 'var(--surface-3)' : 'transparent', color: active ? 'var(--fg-1)' : 'var(--fg-2)', border: '1px solid ' + (active ? 'var(--hairline-strong)' : 'transparent') },
    outline: { background: active ? 'var(--surface-3)' : 'var(--surface-2)', color: 'var(--fg-1)', border: '1px solid var(--hairline-strong)' },
  };
  return (
    <button
      aria-label={label} title={label} aria-pressed={active || undefined} disabled={disabled}
      style={{
        display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: d, height: d,
        flex: '0 0 auto', borderRadius: 'var(--radius-md)', cursor: disabled ? 'not-allowed' : 'pointer',
        opacity: disabled ? 0.45 : 1, transition: 'background var(--dur-fast) var(--ease-out)',
        ...looks[variant], ...style,
      }}
      {...rest}
    >
      {children}
    </button>
  );
}

/* ---------------------------------------------------------------- Badge -- */
export function Badge({ tone = 'neutral', appearance = 'subtle', size = 'md', mono, uppercase, children, style }: {
  tone?: Tone; appearance?: 'subtle' | 'solid' | 'outline'; size?: 'sm' | 'md';
  mono?: boolean; uppercase?: boolean; children: ReactNode; style?: CSSProperties;
}) {
  const t = TONES[tone];
  const sz = size === 'sm' ? { h: 18, px: 6, fs: 'var(--text-2xs)' } : { h: 22, px: 8, fs: 'var(--text-xs)' };
  const look = appearance === 'solid'
    ? { background: t.solid, color: 'var(--fg-on-accent)', border: '1px solid transparent' }
    : appearance === 'outline'
      ? { background: 'transparent', color: t.c, border: '1px solid var(--hairline-strong)' }
      : { background: t.bg, color: t.c, border: '1px solid transparent' };
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 5, height: sz.h, padding: `0 ${sz.px}px`,
      borderRadius: 'var(--radius-xs)', fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)',
      fontSize: sz.fs, fontWeight: 'var(--fw-medium)', lineHeight: 1,
      letterSpacing: uppercase ? 'var(--ls-caps)' : '0.01em', textTransform: uppercase ? 'uppercase' : 'none',
      whiteSpace: 'nowrap', ...look, ...style,
    }}>{children}</span>
  );
}

/* ------------------------------------------------------------ StatusDot -- */
export function StatusDot({ tone = 'neutral', size = 8, label, style }: {
  tone?: Tone; size?: number; label?: ReactNode; style?: CSSProperties;
}) {
  const c = TONES[tone].solid;
  const dot = <span style={{ width: size, height: size, borderRadius: '50%', background: c, flex: '0 0 auto' }} />;
  if (label == null) return dot;
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 7, color: 'var(--fg-2)', fontSize: 'var(--text-sm)', ...style }}>
      {dot}{label}
    </span>
  );
}

/* ----------------------------------------------------------------- Card -- */
export function Card({ title, eyebrow, actions, footer, padding = 'md', raised, children, style, bodyStyle }: {
  title?: ReactNode; eyebrow?: ReactNode; actions?: ReactNode; footer?: ReactNode;
  padding?: 'none' | 'sm' | 'md' | 'lg'; raised?: boolean; children?: ReactNode;
  style?: CSSProperties; bodyStyle?: CSSProperties;
}) {
  const p = { none: 0, sm: 'var(--space-3)', md: 'var(--space-5)', lg: 'var(--space-6)' }[padding];
  return (
    <section style={{
      background: raised ? 'var(--surface-2)' : 'var(--surface-card)', border: '1px solid var(--hairline)',
      borderRadius: 'var(--radius-lg)', boxShadow: raised ? 'var(--shadow-sm)' : 'var(--shadow-xs)',
      display: 'flex', flexDirection: 'column', minWidth: 0, ...style,
    }}>
      {(title || eyebrow || actions) && (
        <header style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 'var(--space-3)', padding: `var(--space-4) ${p} var(--space-3)`, borderBottom: '1px solid var(--hairline)' }}>
          <div style={{ minWidth: 0 }}>
            {eyebrow && <Overline style={{ marginBottom: 4 }}>{eyebrow}</Overline>}
            {title && <h3 style={{ font: 'var(--type-h3)', color: 'var(--fg-1)', margin: 0 }}>{title}</h3>}
          </div>
          {actions && <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', flex: '0 0 auto' }}>{actions}</div>}
        </header>
      )}
      <div style={{ padding: p, minWidth: 0, flex: 1, ...bodyStyle }}>{children}</div>
      {footer && <footer style={{ padding: `var(--space-3) ${p}`, borderTop: '1px solid var(--hairline)', color: 'var(--fg-3)', fontSize: 'var(--text-xs)' }}>{footer}</footer>}
    </section>
  );
}

/* ------------------------------------------------------------- Overline -- */
export function Overline({ children, style }: { children: ReactNode; style?: CSSProperties }) {
  return (
    <div style={{ font: 'var(--type-overline)', letterSpacing: 'var(--ls-caps)', textTransform: 'uppercase', color: 'var(--fg-3)', ...style }}>
      {children}
    </div>
  );
}
