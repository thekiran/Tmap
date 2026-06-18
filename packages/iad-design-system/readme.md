# IAD — Internet Access Detector · Design System

A design system for **IAD** (a.k.a. *iad-agent* / "Tmap"): a Go-based, authorized
network-intelligence tool that estimates how a host reaches the internet — the
**access technology** (Fiber / Cable / DSL / Fixed-Wireless / Mobile / Satellite…)
and the **network context** (interface, gateway chain, NAT topology, ISP/ASN) —
with a **calibrated confidence** and explicit honesty about what it cannot prove.

The product's defining trait is *epistemic honesty*: it distinguishes confirmed
evidence from inference, reports uncertainty as a first-class value, and never
overclaims. This system encodes that ethos into type, color, components, and a
read-only desktop console.

---

## Sources

This system was derived from the attached repository (read-only for the reader —
recorded here in case you have access):

- **GitHub:** `thekiran/Tmap` — https://github.com/thekiran/Tmap
  - `docs/design/` — the UI/UX brief, tech stack, API design, and physical-layer
    detection dossier (the UI was **specified but not yet built**).
  - `agent/` — the Go agent (detection adapters, probes). **Not modified.**

Because the visual UI did not exist, this system is an **original, faithful
realization** of the documented brief, not a copy of shipped screens. Explore the
repo to refine fidelity further.

> ### ⚠️ Font substitution — please confirm
> The repo ships **no brand fonts** (no UI was built). We selected a technical
> trio and load it from Google Fonts:
> **Space Grotesk** (display), **IBM Plex Sans** (UI/body), **JetBrains Mono**
> (all data readouts). If IAD has — or wants — official typefaces, send them and
> we'll swap `tokens/fonts.css`.

---

## Content fundamentals

How IAD writes. The voice is a **calm senior network engineer**: precise,
declarative, never breathless. Color and words both carry meaning, so neither is
wasted.

- **Honest, not confident-sounding.** Always pair a verdict with its confidence
  and its limits. *"Fiber (FTTH) — 74%, medium quality. Physical medium not
  directly confirmed."* Never *"Your connection is Fiber!"*
- **Name the doubt.** Uncertainty is a feature. Surfaces literally have an
  "Uncertainty" and "Why not certain" section. e.g. *"CPE management interface
  did not respond to SNMP."*
- **Confirmed vs inferred is sacred.** Two different words, two different line
  styles, two different tones. *"ARP confirmed"* vs *"Wi-Fi assoc (inferred)"* vs
  *"Unknown L2"*. Never label inference as confirmed.
- **Person:** address the operator as **you** ("inbound connections will not reach
  this host"); describe the system/data in the **third person** ("the host sits
  behind CGNAT"). Avoid "I".
- **Case:** Sentence case for prose and titles. **UPPERCASE mono** only for small
  labels, eyebrows, status enums, and tier names (`SUCCESS`, `NO_DATA`,
  `L2 · LINK`). Product name is **IAD** (all caps).
- **Numbers are exact and typed.** `412.6 Mbps`, `11.4 ms`, `−67 dBm`, `AS3320`,
  `192.168.1.0/24`. Units are spelled, monospaced, and slightly muted.
- **Enums are lowercase snake_case** in data and chips: `success`, `no_data`,
  `cgnat`, `route_hop`, `unknown_l2_connection`.
- **No marketing, no hype, no emoji.** A network tool doesn't celebrate. Microcopy
  is short and instrumental: *"scroll = zoom · drag node = reposition"*.
- **Empty/blocked states explain themselves.** *"No LLDP frames observed —
  consumer gear rarely emits LLDP; absence is not evidence."*

---

## Visual foundations

A **clean black/white network instrument**. Think oscilloscope, not dashboard
template. Color is reserved for meaning; the surface itself is quiet.

- **Palette / vibe.** Neutral near-black surfaces layered by elevation
  (`#0A0A0B → #18181B → #27272C`), a white→gray text ramp, and a **single
  restrained accent** (blue `#3B82F6`) used *only* for selection, focus, and the
  one primary action on a screen. No blue/purple "AI dashboard" wash, no neon, no
  glassmorphism, no glow clichés.
- **Color is meaning, never decoration.** Three meaningful scales:
  *Confidence* (Low gray → Medium amber → High green — uncertainty is calm, never
  red), *Probe status* (success/partial/no_data/skipped/failed/blocked), and
  *Evidence tiers* (one fixed hue each: physical/L2/L3/NAT/ISP), used sparingly.
- **Type.** Space Grotesk for headings and the verdict; IBM Plex Sans for body/UI;
  JetBrains Mono with **tabular numerals** for every IP, MAC, ASN, and metric so
  columns align. Small labels are uppercase mono with wide tracking.
- **Topology lines encode certainty by STYLE, not color:** solid = confirmed
  (LLDP/CDP/SNMP/ARP), dashed = inferred, dotted = unknown L2, thin = route hop,
  muted = low-confidence; NAT and ISP crossings get labeled boundary chips. This
  is the signature brand motif (see the *Topology edge language* card).
- **Backgrounds.** A subtle **radial dot-grid** ("instrument surface") behind maps
  and the verdict card; otherwise flat neutral fills. No imagery, no photography,
  no illustration — the data is the picture.
- **Depth.** Comes from **1px hairline borders** (`rgba(255,255,255,.09)`) plus a
  faint ambient shadow — not heavy drop shadows. Cards: `surface-1` fill, hairline
  border, `radius-lg` (12px), `shadow-xs`.
- **Corner radii.** Tight and instrument-like: 4 (chips) · 6 (inputs) · 8
  (buttons/controls) · 12 (cards) · 16 (dialogs). Pills only for meters/toggles.
- **Spacing.** 4px base grid; dense but breathable. Information-dense tables and
  cards, generous outer gutters.
- **Motion.** Calm and short. `--ease-out` (cubic-bezier(.16,1,.3,1)),
  120–200ms for hovers, ~800ms for the confidence-meter fill. No bounces, no
  infinite decorative loops. `prefers-reduced-motion` respected.
- **States.** Hover = surface lightens one step + border strengthens. Press =
  no shrink (instrument feel). Selection = accent border + faint accent ring.
  Focus = 2px accent outline, 2px offset. Disabled = 45% opacity.
- **Transparency/blur.** Used only for the dialog scrim (`--overlay`) and tinted
  status backgrounds (color at ~14% alpha). No frosted glass.
- **Imagery tone.** N/A — there is no photography. The only "imagery" is the node
  graph and the dot-grid texture, both monochrome.
- **Theme.** Dark-first; a complete light theme ships under `[data-theme="light"]`
  (white/off-white surfaces, black/gray text, same accents at higher contrast).

---

## Iconography

- **Style:** a small, hand-picked **Lucide-style** set — 24px grid, 2px stroke,
  round caps/joins, `currentColor`. Lucide is *not* bundled with the planned
  Tauri/React app, so the kit ships its own inline-SVG subset
  (`ui_kits/console/icons.jsx`, `window.Icons`) matching Lucide's geometry. If you
  adopt Lucide proper, the names map 1:1.
- **Usage:** icons are **functional, not decorative** — nav items, device/node
  types (router, modem, AP, switch, server, printer, mobile, IoT, globe, unknown),
  toolbar actions (zoom/fit/reset/layers/lock), and status glyphs
  (alert/info/check/shield). One icon per concept; never two icons competing.
- **Device/node icons** are mapped from `device.type` via `fmt.deviceIcon()` so the
  same type always gets the same glyph across Devices, Topology, and the drawer.
- **No emoji. No unicode-as-icon** except a few typographic marks used as data
  (`↑ ↓` sort arrows, `✓` confirmations, `−` for negatives/signal). 
- **Logo:** a node-graph glyph (a hub probing four links) + a JetBrains Mono
  "IAD" wordmark — `assets/logo-iad.svg` (uses `currentColor`) and a standalone
  app mark `assets/mark-iad.svg` (accent-blue hub on near-black).

---

## Index / manifest

**Root**
- `styles.css` — global entry (import-only). Consumers link this one file.
- `tokens/` — `colors.css`, `typography.css`, `spacing.css`, `base.css`, `fonts.css`.
- `assets/` — `logo-iad.svg`, `mark-iad.svg`.
- `SKILL.md` — Agent-Skill manifest (works in Claude Code).

**Components** (`window.IADInternetAccessDetectorDesignSystem_019e02`)
- `components/core/` — `Button`, `IconButton`, `Badge`, `StatusDot`, `Card`.
- `components/data/` — `MetricStat`, `ConfidenceBar` (+`band()`), `ProbeStatusBadge`, `TierBadge`.
- `components/forms/` — `Toggle`, `SegmentedControl`, `Input`.

Each component directory has a `.jsx`, a `.d.ts` (props), a `.prompt.md` (usage),
and a `@dsCard` HTML specimen.

**Foundations** (`guidelines/`) — specimen cards for the Design System tab:
Colors (surfaces, text, accent, confidence, status, tiers), Type (display, body,
mono), Spacing (scale, radii, elevation), Brand (logo, **edge language**, nodes,
texture).

**UI kit** (`ui_kits/console/`) — the full desktop console recreation
(Dashboard · Topology · Devices · Evidence · Reports · Settings). See its
`README.md`.

---

*Built from the `thekiran/Tmap` design dossier. The Go agent is unchanged; this
system covers the frontend visual language and component library only.*
