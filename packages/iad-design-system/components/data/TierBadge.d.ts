import { ReactNode, CSSProperties } from 'react';

/**
 * Labels an evidence tier / topology layer with its fixed hue. Each tier has
 * one consistent color used across legend, edges, evidence, and toggles.
 */
export type Tier = 'physical' | 'l2' | 'l3' | 'nat' | 'isp';

export interface TierBadgeProps {
  tier: Tier;
  /** Override the default label text. */
  label?: ReactNode;
  /** subtle = tinted chip (default), solid = filled, dot = dot + text. */
  appearance?: 'subtle' | 'solid' | 'dot';
  style?: CSSProperties;
}

export function TierBadge(props: TierBadgeProps): JSX.Element;
