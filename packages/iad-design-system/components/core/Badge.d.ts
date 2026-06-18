import { ReactNode, CSSProperties } from 'react';

/**
 * Small label chip for statuses, counts, and categorical tags. Tone maps to the
 * semantic color tokens; color is meaningful, not decorative.
 */
export interface BadgeProps {
  tone?: 'neutral' | 'accent' | 'success' | 'warn' | 'danger' | 'info' | 'blocked';
  /** subtle = tinted bg (default), solid = filled, outline = quiet metadata. */
  appearance?: 'subtle' | 'solid' | 'outline';
  size?: 'sm' | 'md';
  /** Render label in the mono family (good for IDs, counts, enums). */
  mono?: boolean;
  /** Uppercase + tracking — for status enums like SUCCESS / NO_DATA. */
  uppercase?: boolean;
  children?: ReactNode;
  style?: CSSProperties;
}

export function Badge(props: BadgeProps): JSX.Element;
