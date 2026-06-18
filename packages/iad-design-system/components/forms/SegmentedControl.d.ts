import { ReactNode, CSSProperties } from 'react';

export interface SegmentOption {
  value: string;
  label: ReactNode;
  icon?: ReactNode;
}

/**
 * Compact single-select for view/theme/layout switches. Controlled via
 * value + onChange.
 */
export interface SegmentedControlProps {
  options: SegmentOption[];
  value: string;
  onChange: (value: string) => void;
  size?: 'sm' | 'md';
  fullWidth?: boolean;
  style?: CSSProperties;
}

export function SegmentedControl(props: SegmentedControlProps): JSX.Element;
