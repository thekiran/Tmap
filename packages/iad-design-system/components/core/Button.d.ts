import { ReactNode, CSSProperties } from 'react';

/**
 * Primary action control for the IAD console. Monochrome-first; color is
 * reserved for the accent (primary) and destructive (danger) variants.
 */
export interface ButtonProps {
  /** Visual emphasis. Default "secondary" (neutral outline). */
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger';
  /** Control height. Default "md". */
  size?: 'sm' | 'md' | 'lg';
  /** Optional leading icon (sized to 1em). */
  iconLeft?: ReactNode;
  /** Optional trailing icon (sized to 1em). */
  iconRight?: ReactNode;
  disabled?: boolean;
  /** Stretch to container width. */
  fullWidth?: boolean;
  type?: 'button' | 'submit' | 'reset';
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  children?: ReactNode;
  style?: CSSProperties;
}

export function Button(props: ButtonProps): JSX.Element;
