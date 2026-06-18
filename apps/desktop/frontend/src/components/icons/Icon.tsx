import type { CSSProperties } from 'react';

/** Inline-SVG icon set (Lucide-style: 24px grid, 2px stroke, round caps).
 *  Self-contained so the desktop app has no icon-font/CDN dependency. */
export interface IconProps {
  size?: number;
  style?: CSSProperties;
  className?: string;
}

const stroke = (size = 18, style?: CSSProperties, className?: string) =>
  ({
    viewBox: '0 0 24 24', width: size, height: size, fill: 'none',
    stroke: 'currentColor', strokeWidth: 2, strokeLinecap: 'round' as const,
    strokeLinejoin: 'round' as const, style: { display: 'block', ...style }, className,
  });

const path = (paths: string[]) => function Icon({ size, style, className }: IconProps) {
  return (
    <svg {...stroke(size, style, className)}>
      {paths.map((d: string, i: number) => (<path key={i} d={d} />))}
    </svg>
  );
};

// Generic raw builder for icons that need circles/rects.
function raw(children: (P: typeof svgEls) => JSX.Element[]) {
  return function Icon({ size, style, className }: IconProps) {
    return <svg {...stroke(size, style, className)}>{children(svgEls)}</svg>;
  };
}
const svgEls = {
  c: (cx: number, cy: number, r: number, key: number) => <circle key={key} cx={cx} cy={cy} r={r} />,
  r: (x: number, y: number, w: number, h: number, rx: number, key: number) => <rect key={key} x={x} y={y} width={w} height={h} rx={rx} />,
  p: (d: string, key: number) => <path key={key} d={d} />,
};

export const Icons = {
  dashboard: path(['M3 3h7v7H3zM14 3h7v7h-7zM14 14h7v7h-7zM3 14h7v7H3z']),
  topology: raw((P) => [P.c(5, 6, 2.4, 1), P.c(19, 6, 2.4, 2), P.c(12, 18, 2.4, 3), P.p('M7 6h10M6.5 8 11 16M17.5 8 13 16', 4)]),
  devices: path(['M3 5h18v11H3z', 'M8 21h8M12 16v5']),
  evidence: path(['M9 3h6l3 4v14H6V3z', 'M9 12h6M9 16h4']),
  reports: path(['M14 3v5h5', 'M14 3H6v18h12V8z', 'M9 13h6M9 17h6']),
  settings: raw((P) => [P.c(12, 12, 3, 1), P.p('M19.4 13.5a1.7 1.7 0 0 0 .3 1.9l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-2.9 1.2V21a2 2 0 1 1-4 0v-.2a1.7 1.7 0 0 0-2.9-1.2l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0-1.2-2.9H3a2 2 0 1 1 0-4h.2a1.7 1.7 0 0 0 1.2-2.9l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 2.9-1.2V3a2 2 0 1 1 4 0v.2a1.7 1.7 0 0 0 2.9 1.2l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.9Z', 2)]),

  refresh: path(['M21 12a9 9 0 1 1-3-6.7L21 8', 'M21 3v5h-5']),
  download: path(['M12 3v12', 'm7 12 5 5 5-5', 'M5 21h14']),
  upload: path(['M12 21V9', 'm7 12 5-5 5 5', 'M5 3h14']),
  copy: raw((P) => [P.r(9, 9, 12, 12, 2, 1), P.p('M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1', 2)]),
  search: raw((P) => [P.c(11, 11, 7, 1), P.p('m21 21-4.3-4.3', 2)]),
  close: path(['M18 6 6 18M6 6l12 12']),
  chevronRight: path(['m9 6 6 6-6 6']),
  zoomIn: raw((P) => [P.c(11, 11, 7, 1), P.p('m21 21-4.3-4.3M11 8v6M8 11h6', 2)]),
  zoomOut: raw((P) => [P.c(11, 11, 7, 1), P.p('m21 21-4.3-4.3M8 11h6', 2)]),
  fit: path(['M3 8V5a2 2 0 0 1 2-2h3', 'M21 8V5a2 2 0 0 0-2-2h-3', 'M3 16v3a2 2 0 0 0 2 2h3', 'M21 16v3a2 2 0 0 1-2 2h-3']),
  reset: path(['M3 12a9 9 0 1 0 9-9 9 9 0 0 0-6.4 2.6L3 8', 'M3 3v5h5']),
  layers: path(['m12 2 9 5-9 5-9-5z', 'm3 12 9 5 9-5', 'm3 17 9 5 9-5']),
  lock: raw((P) => [P.r(4, 11, 16, 10, 2, 1), P.p('M8 11V7a4 4 0 0 1 8 0v4', 2)]),

  host: raw((P) => [P.r(3, 4, 18, 12, 2, 1), P.p('M8 20h8M12 16v4', 2)]),
  router: raw((P) => [P.r(2, 13, 20, 7, 2, 1), P.p('M6 17h.01M10 17h.01M14 8l2-2 2 2M16 6v7', 2)]),
  modem: raw((P) => [P.r(2, 6, 20, 12, 2, 1), P.p('M6 18v2M18 18v2M6 10h.01M10 10h.01', 2)]),
  ap: raw((P) => [P.p('M5 12.5a7 7 0 0 1 14 0M8 15a4 4 0 0 1 8 0', 1), P.c(12, 18, 1.4, 2)]),
  switch: raw((P) => [P.r(3, 8, 18, 8, 2, 1), P.p('M7 12h.01M11 12h.01M15 12h.01', 2)]),
  server: raw((P) => [P.r(3, 4, 18, 7, 2, 1), P.r(3, 13, 18, 7, 2, 2), P.p('M7 7.5h.01M7 16.5h.01', 3)]),
  printer: raw((P) => [P.p('M6 9V3h12v6', 1), P.r(4, 9, 16, 7, 2, 2), P.p('M7 16h10v5H7z', 3)]),
  mobile: raw((P) => [P.r(7, 2, 10, 20, 2, 1), P.p('M11 18h2', 2)]),
  iot: raw((P) => [P.c(12, 12, 3, 1), P.p('M12 2v4M12 18v4M2 12h4M18 12h4', 2)]),
  globe: raw((P) => [P.c(12, 12, 9, 1), P.p('M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18', 2)]),
  plug: path(['M9 2v6M15 2v6', 'M7 8h10v3a5 5 0 0 1-10 0z', 'M12 16v6']),
  unknown: raw((P) => [P.c(12, 12, 9, 1), P.p('M9.2 9a3 3 0 0 1 5.6 1c0 2-3 2.5-3 4', 2), P.p('M12 17h.01', 3)]),

  alert: raw((P) => [P.p('M12 3 2 20h20z', 1), P.p('M12 10v4M12 17h.01', 2)]),
  info: raw((P) => [P.c(12, 12, 9, 1), P.p('M12 11v5M12 8h.01', 2)]),
  check: path(['M20 6 9 17l-5-5']),
  shield: path(['M12 3 5 6v5c0 4 3 7 7 9 4-2 7-5 7-9V6z', 'M9.5 12l2 2 3.5-4']),
} as const;

export type IconKey = keyof typeof Icons;

/** Render an icon by key (used for device/node-type icons). */
export function Icon({ name, ...props }: IconProps & { name: IconKey }) {
  const C = Icons[name] ?? Icons.unknown;
  return <C {...props} />;
}
