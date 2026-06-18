/** Formatting helpers — pure, UI-agnostic. */
import type { Reachability, ProbeStatus } from './scan-schema';

export function pct(v: number): string {
  return Math.round(v * 100) + '%';
}

export function fixed(v: number, n = 1): string {
  return v.toFixed(n);
}

export function timeShort(iso?: string): string {
  if (!iso) return '—';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  return d.toLocaleString('en-GB', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: 'short' });
}

export function ago(iso?: string): string {
  if (!iso) return '—';
  const t = new Date(iso).getTime();
  if (isNaN(t)) return iso;
  const s = (Date.now() - t) / 1000;
  if (s < 60) return 'just now';
  if (s < 3600) return Math.floor(s / 60) + 'm ago';
  if (s < 86400) return Math.floor(s / 3600) + 'h ago';
  return Math.floor(s / 86400) + 'd ago';
}

/** IPv4 sort key. Non-v4 strings sort last but stable. */
export function ipKey(ip: string): number {
  const m = ip.match(/^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})/);
  if (!m) return Number.MAX_SAFE_INTEGER;
  return ((+m[1] * 256 + +m[2]) * 256 + +m[3]) * 256 + +m[4];
}

export const REACH_META: Record<Reachability, { word: string; tone: string }> = {
  self: { word: 'This host', tone: 'accent' },
  reachable: { word: 'Reachable', tone: 'success' },
  arp_only: { word: 'ARP only', tone: 'neutral' },
  partial: { word: 'Partial', tone: 'warn' },
  unreachable: { word: 'Unreachable', tone: 'danger' },
  unknown: { word: 'Unknown', tone: 'neutral' },
};

export const PROBE_STATUS_META: Record<ProbeStatus, { word: string; tone: string }> = {
  success: { word: 'Success', tone: 'success' },
  partial: { word: 'Partial', tone: 'warn' },
  no_data: { word: 'No data', tone: 'neutral' },
  skipped: { word: 'Skipped', tone: 'neutral' },
  failed: { word: 'Failed', tone: 'danger' },
  blocked: { word: 'Blocked', tone: 'blocked' },
  completed: { word: 'Completed', tone: 'success' },
};

/** Map a device/node type to an icon key (see components/icons/Icon.tsx). */
export function deviceIconKey(type: string): string {
  const map: Record<string, string> = {
    local_host: 'host', workstation: 'host', interface: 'plug', subnet: 'layers',
    default_gateway: 'router', router: 'router', modem_cpe: 'modem',
    access_point: 'ap', mesh_node: 'ap', repeater: 'ap',
    managed_switch: 'switch', unmanaged_switch_inferred: 'switch', unknown_l2_segment: 'unknown',
    server: 'server', dns_server: 'server', printer: 'printer', mobile: 'mobile', iot: 'iot',
    isp_gateway: 'globe', isp_route_hop: 'globe', public_internet: 'globe', unknown: 'unknown',
  };
  return map[type] ?? 'unknown';
}
