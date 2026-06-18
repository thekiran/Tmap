import { ReactNode, CSSProperties } from 'react';

/**
 * Accessible on/off switch (role="switch"). Used in Settings and the topology
 * layer panel. Optional inline label + description.
 */
export interface ToggleProps {
  checked: boolean;
  onChange: (next: boolean) => void;
  disabled?: boolean;
  label?: ReactNode;
  description?: ReactNode;
  size?: 'sm' | 'md';
  id?: string;
  style?: CSSProperties;
}

export function Toggle(props: ToggleProps): JSX.Element;
