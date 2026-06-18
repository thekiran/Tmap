import { ReactNode, CSSProperties } from 'react';

/**
 * Text field with optional leading icon and right addon. Use the mono variant
 * for IP / MAC / ASN entry.
 */
export interface InputProps {
  value: string;
  onChange: (value: string, e?: React.ChangeEvent<HTMLInputElement>) => void;
  placeholder?: string;
  type?: string;
  iconLeft?: ReactNode;
  addonRight?: ReactNode;
  size?: 'sm' | 'md';
  mono?: boolean;
  disabled?: boolean;
  invalid?: boolean;
  fullWidth?: boolean;
  style?: CSSProperties;
  inputStyle?: CSSProperties;
}

export function Input(props: InputProps): JSX.Element;
