import { ReactNode, CSSProperties } from 'react';

/**
 * A labeled value readout. Value renders in the mono family with tabular
 * numerals so stacked stats align. The workhorse of dashboard cards.
 */
export interface MetricStatProps {
  /** Uppercase mono label above the value. */
  label: ReactNode;
  value: ReactNode;
  /** Small trailing unit (Mbps, ms, dBm…). */
  unit?: ReactNode;
  /** Caption line below the value. */
  secondary?: ReactNode;
  tone?: 'default' | 'accent' | 'success' | 'warn' | 'danger' | 'muted';
  size?: 'sm' | 'md' | 'lg';
  align?: 'left' | 'right';
  style?: CSSProperties;
}

export function MetricStat(props: MetricStatProps): JSX.Element;
