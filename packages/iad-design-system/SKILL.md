---
name: iad-design
description: Use this skill to generate well-branded interfaces and assets for IAD (Internet Access Detector / iad-agent) — a network-intelligence tool that estimates internet access technology and topology with calibrated confidence. Contains design guidelines, color/type/spacing tokens, fonts, logo assets, reusable React components, and a full desktop-console UI kit for prototyping or production work.
user-invocable: true
---

Read `readme.md` in this skill first — it covers the product context, content
fundamentals (voice, casing, confirmed-vs-inferred), visual foundations
(black/white instrument theme, confidence/tier color systems, the topology edge
language), and iconography. Then explore the other files as needed.

## What's here
- `styles.css` — link this one file to inherit all tokens, fonts, and base styles.
- `tokens/` — colors, typography, spacing, base, fonts (CSS custom properties).
- `components/` — React primitives (Button, Badge, Card, ConfidenceBar,
  ProbeStatusBadge, TierBadge, Toggle, SegmentedControl, Input, …). Each has a
  `.prompt.md` with usage.
- `guidelines/` — foundation specimen cards (color, type, spacing, brand).
- `ui_kits/console/` — a full Wails-style desktop console recreation
  (Dashboard, Topology, Devices, Evidence, Reports, Settings) you can copy from.
- `assets/` — `logo-iad.svg`, `mark-iad.svg`.

## How to use it
- **Visual artifacts** (mocks, slides, throwaway prototypes): copy the assets and
  the tokens you need out into static HTML files and link `styles.css`. Reuse the
  patterns in `ui_kits/console/` rather than inventing new ones.
- **Production code:** read the rules here and the component `.prompt.md` files to
  become an expert in the IAD brand; lift exact token values, the edge-style
  language, and the confidence/tier color systems.

## Non-negotiable brand rules
- Honest by default: always pair a verdict with its confidence and limitations.
- Confirmed ≠ inferred: solid lines/`confirmed` wording vs dashed/`inferred` vs
  dotted/`unknown`. Never label inference as confirmed.
- Color is meaning, not decoration. Monochrome surfaces; one restrained blue
  accent; status-only color. No neon, glassmorphism, emoji, or AI-gradient wash.
- Mono + tabular numerals for every IP/MAC/ASN/metric.

If invoked with no other guidance, ask the user what they want to build or design,
ask a few focused questions, then act as an expert IAD designer who outputs either
HTML artifacts or production code depending on the need.
