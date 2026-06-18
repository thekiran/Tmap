import { ReactNode, CSSProperties } from 'react';

/**
 * A small filled dot encoding a state, optionally with a trailing label.
 * Used for reachability, probe status, and layer legends.
 */
export interface StatusDotProps {
  tone?: 'neutral' | 'success' | 'warn' | 'danger' | 'info' | 'accent' | 'blocked';
  /** Diameter in px. Default 8. */
  size?: number;
  /** Calm breathing ring — reserve for genuinely live/active states. */
  pulse?: boolean;
  /** Optional trailing text label. */
  label?: ReactNode;
  style?: CSSProperties;
}

export function StatusDot(props: StatusDotProps): JSX.Element;
