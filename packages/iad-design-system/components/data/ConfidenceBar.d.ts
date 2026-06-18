import { CSSProperties } from 'react';

/**
 * The headline "how sure are we" control: a labeled track filled to a 0–1
 * confidence, colored by band (Low gray / Medium amber / High green) with the
 * percentage and band word. Low confidence is deliberately calm, never red.
 * Bands: <0.45 Low · 0.45–0.75 Medium · ≥0.75 High.
 */
export interface ConfidenceBarProps {
  /** Confidence 0–1 (clamped). */
  value: number;
  showLabel?: boolean;
  showValue?: boolean;
  label?: string;
  size?: 'sm' | 'md' | 'lg';
  style?: CSSProperties;
}

export function ConfidenceBar(props: ConfidenceBarProps): JSX.Element;
/** Returns the band name for a 0–1 confidence value. */
export function band(value: number): 'low' | 'medium' | 'high';
