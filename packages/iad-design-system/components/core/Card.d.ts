import { ReactNode, CSSProperties } from 'react';

/**
 * The primary container surface for the IAD console. Optional header
 * (eyebrow + title + actions) and footer. Depth comes from hairline borders,
 * not heavy shadows.
 */
export interface CardProps {
  title?: ReactNode;
  /** Small uppercase mono label above the title. */
  eyebrow?: ReactNode;
  /** Header-right slot (buttons, menus). */
  actions?: ReactNode;
  footer?: ReactNode;
  padding?: 'none' | 'sm' | 'md' | 'lg';
  /** Use the raised surface tone + slightly stronger shadow. */
  raised?: boolean;
  /** Animate border/background on hover (for clickable cards). */
  interactive?: boolean;
  children?: ReactNode;
  style?: CSSProperties;
  bodyStyle?: CSSProperties;
}

export function Card(props: CardProps): JSX.Element;
