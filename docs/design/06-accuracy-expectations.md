# 06 — Accuracy Expectations

> Brief §F. **Honest, defensible ranges** — the whole product thesis is to *not*
> overclaim. These are engineering targets and realistic ceilings, not
> guarantees, and they assume the calibration and Unknown-gate discipline in
> [04](04-classification-strategy.md).

## 1. How to read these numbers

- Accuracy is reported as **selective accuracy at a coverage** — i.e. accuracy
  *among the predictions we choose to commit to*, plus what fraction we commit to
  (the rest are honest `Unknown`). A mode that can't see physical evidence has
  **low coverage by design**; that's a feature, not a failure.
- **Family-level** (DSL vs Fiber vs Cable vs FWA vs Mobile vs Satellite vs
  Enterprise) is far more achievable than **subtype-level** (ADSL2+ vs VDSL2,
  DOCSIS 3.0 vs 3.1, GPON vs XGS-PON).
- Numbers are **per-region calibrated**; a brand-new region starts lower and
  improves as labels arrive. Mobile and satellite have natural advantages
  (radio type / RTT signature are near-authoritative).

## 2. Family-level accuracy by mode (target ranges)

| Mode | Coverage (committed, non-Unknown) | Family selective accuracy | Subtype selective accuracy | Notes |
|---|---|---|---|---|
| **Remote-only** | 40–65% | **70–85%** | 35–55% | Strong for Mobile/Satellite (RTT, CGNAT, ASN); weak for DSL-vs-Fiber-vs-Cable behind a modem. Honest `Unknown` carries the rest. |
| **Browser-only** | 35–55% | **65–80%** | 30–50% | Similar to remote but better RTT/bufferbloat from the client; sandbox blocks local hints. |
| **Desktop agent** | 60–80% | **82–92%** | 50–70% | UPnP modem fingerprint + link speed + NAT topology lift family accuracy and sometimes subtype. |
| **Mobile agent** | 75–90% | **90–97%** (Mobile vs not) | 70–88% (LTE/5G NSA/SA) | Radio type is near-authoritative; FWA-vs-handheld via cell stability. |
| **Router/CPE telemetry** | 80–95% | **93–99%** | **85–97%** | Physical-layer ground truth (DSL stats, DOCSIS version, PON optical) — the gold mode. Limited by access permission, not by signal. |
| **Combined (agent + CPE + multi-vantage remote)** | 80–95% | **95–99%** | **88–97%** | Approaches ground truth; the few errors are genuinely ambiguous deployments (FTTB-then-copper, bonded hybrids). |

> **No mode reaches 100%, ever, on remote-only or browser**, and the product must
> never display it. Even CPE telemetry can be fooled by hybrid/bonded links and
> mislabeled CPE. The honest ceiling for *remote-only family* classification is
> roughly the mid-80s **with substantial `Unknown` coverage carved out** — push
> coverage higher and accuracy falls.

## 3. Where accuracy comes from / breaks down

**Strong, reliable discriminations (high accuracy):**
- Mobile vs fixed (radio type, CGNAT, ASN-type, packet-train quantization).
- GEO satellite (RTT ≥ 500ms is nearly decisive); LEO via periodic 20–60ms
  signature + Starlink ASN.
- Anything with CPE physical stats or a WAN interface name.

**Genuinely hard (where `Unknown`/low confidence is the honest answer):**
- **DSL vs Fiber vs Cable from remote-only/browser behind a NAT modem** — the
  classic confound; performance overlaps heavily on modern plans.
- **FTTC/FTTN vs VDSL2** — same copper last drop; often only marketing differs.
- **4G/5G FWA vs mobile handheld** — same radio; separated only by cell
  stability over time (needs the mobile agent + duration).
- **Subtypes generally** without physical evidence (DOCSIS version, PON variant,
  DSL profile) — abstain rather than guess.
- **Hybrid/bonded** (Fiber+DSL, LTE+DSL, SD-WAN) — flag the hybrid signature; do
  not force a single leaf.

## 4. Why these are realistic (grounding)

- Published medium-classification research and operator experience converge on
  the same story: **performance features alone give family-level signal, not
  certainty**; physical/operator evidence is what makes it sharp. The ranges above
  reflect that — modest remote-only ceilings, high CPE/mobile ceilings.
- The MVP today behaves exactly this way: behind a modem with only weak evidence
  it returns `Unknown` at ~12% confidence with visible candidates rather than a
  confident guess. The platform improves *coverage* and *calibration*, not the
  laws of physics.

## 5. Promotion gates (how accuracy claims stay honest)

A model version is only promoted if, on the held-out, region-stratified set:

1. **No per-region family-accuracy regression** vs the current production model
   (macro-averaged across regions, and no single major region drops > 2pts).
2. **Calibration**: ECE ≤ target per (mode × region) — a stated confidence must
   match observed accuracy.
3. **Selective accuracy at target coverage** meets the band in §2 for each mode;
   if it can only hit the accuracy by abstaining more, the **Unknown-rate must
   stay within an allowed band** (no silently dumping hard cases into `Unknown`
   to fake accuracy).
4. **Beats the rule baseline** on macro-F1 in every region it serves; otherwise
   that region keeps the rule baseline.
5. **OOD behavior**: on synthetic out-of-distribution inputs, it abstains, not
   commits.

## 6. What we report to users (and never report)

- **Always:** a calibrated confidence, the top alternatives, the evidence behind
  it, and "what would raise confidence" (next-best probe).
- **Never:** a bare label, "100% / certain", or a subtype the evidence can't
  support. A confident *family* with "subtype undetermined" is correct and honest;
  a falsely precise subtype is a product bug.
- **Operator dashboard:** live per-region/per-ASN accuracy from incoming labels,
  Unknown-rate, calibration drift — so claims are continuously audited, not
  asserted once.
