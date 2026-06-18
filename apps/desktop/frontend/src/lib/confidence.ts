/** Confidence helpers — the calibrated bands are a product-wide contract.
 *  < 0.45 Low · 0.45–0.75 Medium · ≥ 0.75 High. Uncertainty is calm, never red. */
import type { ConfidenceBand } from './scan-schema';

export const BAND_THRESHOLDS = { medium: 0.45, high: 0.75 } as const;

export function band(value: number): ConfidenceBand {
  if (value >= BAND_THRESHOLDS.high) return 'high';
  if (value >= BAND_THRESHOLDS.medium) return 'medium';
  return 'low';
}

export function bandWord(value: number): string {
  return { low: 'Low', medium: 'Medium', high: 'High' }[band(value)];
}

/** CSS custom-property name for a band's foreground color. */
export function bandColorVar(value: number): string {
  return { low: 'var(--conf-low)', medium: 'var(--conf-med)', high: 'var(--conf-high)' }[band(value)];
}

export function bandBgVar(value: number): string {
  return { low: 'var(--conf-low-bg)', medium: 'var(--conf-med-bg)', high: 'var(--conf-high-bg)' }[band(value)];
}

export function clamp01(v: number): number {
  return Math.max(0, Math.min(1, v));
}
