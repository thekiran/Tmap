import { ReactNode, CSSProperties } from 'react';

/**
 * Square, icon-only control for toolbars and row actions. Always requires an
 * accessible label. Supports a pressed/active state for toggle toolbars.
 */
export interface IconButtonProps {
  variant?: 'ghost' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  /** Pressed/selected state (e.g. an active layer toggle). */
  active?: boolean;
  disabled?: boolean;
  /** Required — used for aria-label and tooltip. */
  label: string;
  onClick?: (e: React.MouseEvent<HTMLButtonElement>) => void;
  /** The icon node. */
  children?: ReactNode;
  style?: CSSProperties;
}

export function IconButton(props: IconButtonProps): JSX.Element;
