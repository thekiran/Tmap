# 11 — UI/UX Dashboard Design Brief

> Brief §K Deliverable #10. Two surfaces from one design system: the **end-user
> result** (what's my connection, and how sure are you?) and the **operator/
> researcher console** (fleet health, accuracy, drift, labels, models).

## 1. Design principles

1. **Confidence is the headline, not the label.** Never show a bare "Fiber".
   Show "Fiber — FTTH/GPON, 94% confident" with the confidence visually primary.
2. **Always answer "why".** Every verdict expands into its evidence (tiered) and
   a plain-language explanation — the `score_contributions` made visible.
3. **`Unknown` is a first-class, non-embarrassing state.** When uncertain, show
   the ranked candidates, what *was* established (ISP/NAT/context), and the
   single most useful next step ("authorize CPE read to resolve DSL vs Fiber").
4. **Progressive disclosure.** A glanceable answer up top; tiers, raw evidence,
   and JSON on demand. Never overwhelm; never hide.
5. **Privacy is visible.** Consent state, what was collected, and a one-click
   "delete my data" are always reachable.

## 2. End-user result screen

```
┌────────────────────────────────────────────────────────────┐
│  Your connection                                            │
│  ╭──────────────────────╮                                   │
│  │   FIBER · FTTH (GPON) │   ▓▓▓▓▓▓▓▓▓░  94% confident       │
│  ╰──────────────────────╯   decision quality: HIGH          │
│                                                              │
│  Also possible:  Cable 2% · XGS-PON 2% · Unknown 2%          │
│                                                              │
│  Why we think this  ▾                                        │
│   ● Physical  TR-064: WAN access type GPON, ONT optical OK   │
│   ● Device    Modem HG8245H matches a GPON ONT fingerprint   │
│   ● Perf      ~920/880 Mbps symmetric, low bufferbloat       │
│   ○ Network   ISP Example Telecom (AS9121), no CGNAT         │
│                                                              │
│  Network facts (high confidence even if type were unknown):  │
│   ISP · Country · ASN · Public IP (masked) · NAT · IPv6      │
│                                                              │
│  [ Re-run test ]  [ Download report (PDF) ]  [ Manage data ] │
└────────────────────────────────────────────────────────────┘
```

- **Confidence meter** uses color + number + a label (Low/Medium/High) — never
  100%. A calibrated 94% is described as "very likely", not "certain".
- **Tier dots** (●/○) show evidence *strength* per tier (Physical/Device/Network/
  Performance) at a glance — the `evidence_strength` summary.
- **Subtype honesty:** if family is confident but subtype isn't, show
  "Fiber (subtype undetermined)" rather than a guess.

### Uncertain variant
```
   UNKNOWN · evidence too weak to commit       12% · quality LOW
   Closest guesses:  Fiber 27 · Cable 24 · VDSL 22 · Unknown 27
   Why uncertain:
     – No physical evidence of the line behind your modem (NAT)
     – Top options too close to separate honestly
   Raise confidence:  ▸ Install the desktop agent (reads modem model)
                      ▸ Authorize CPE read (reveals WAN line type)
```

## 3. Physical-layer detail (when CPE telemetry is available)

When the agent/CPE mode captures line stats, expose them — this is the
high-trust, satisfying view (and where VDSL2 profiles surface):

```
  Line details (read from your modem, with permission)
   Technology:  VDSL2  ·  Profile 35b (Super Vectoring, ~35 MHz)
   Sync rate:   294 / 48 Mbps     Attainable: 310 / 52 Mbps
   SNR margin:  6.1 dB            Attenuation: 11.5 dB
   Mode:        G.993.5 vectoring · interleaved
```

- Profile chip (`35b`, `17a`, …) with a tooltip explaining band/MHz and that
  **profile ≠ speed** (real rate depends on loop length, copper quality, SNR).
- DOCSIS view: version, OFDM/OFDMA, channel counts, power, MER. PON view: type,
  optical Rx/Tx dBm, ONT model.

## 4. Operator / researcher console

- **Fleet overview:** scans/day by mode & region, Unknown-rate trend, ingest lag.
- **Accuracy board:** per-region / per-ASN / per-family selective accuracy from
  incoming labels; calibration (reliability) curves; coverage vs accuracy slider.
- **Model panel:** active vs shadow vs canary `model_versions`, side-by-side
  distribution comparison, promotion-gate status, SHAP global importances + drift.
- **Label queue:** active-learning picks (uncertain/OOD/disagreement) for review;
  per-ISP self-report reliability.
- **Drift monitors:** feature & prediction drift, with retrain triggers.
- **Data governance:** consent/erasure audit log, dataset snapshot browser,
  k-anonymity status on exports.

## 5. Reporting

- **Per-scan PDF/HTML:** verdict + confidence + tiered evidence table +
  "what would raise confidence" + (if CPE) the line-detail panel. Reuses the
  dashboard components.
- **Aggregate exports** (researchers): privacy-reviewed, k-anonymized
  distributions by region/ASN/family; includes the model version + datasheet link.

## 6. Visual system & accessibility

- Calm, technical, trustworthy. One accent per evidence tier (consistent
  everywhere). Confidence color scale is colorblind-safe and always paired with
  text. WCAG AA. Dark/light. Localizable from day one (the MVP already speaks
  Turkish + English in its output) — strings externalized, RTL-ready.
- Charts: distribution bars, reliability curves, time-series for stability/diurnal
  features, traceroute path view with per-hop RTT and token annotations.
