import { CSSProperties } from 'react';

/**
 * Fixed badge for the IAD probe-status enum. Each status has a locked tone so
 * the same state reads identically across Evidence, Devices, and Reports.
 */
export type ProbeStatus = 'success' | 'partial' | 'no_data' | 'skipped' | 'failed' | 'blocked';

export interface ProbeStatusBadgeProps {
  status: ProbeStatus;
  size?: 'sm' | 'md';
  style?: CSSProperties;
}

export function ProbeStatusBadge(props: ProbeStatusBadgeProps): JSX.Element;
