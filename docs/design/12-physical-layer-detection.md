# 12 — Physical-Layer Line Detection (implemented)

> This is the one design document that is **already code**. It captures
> fine-grained line properties — the kind a modem's status page shows — and folds
> them into the verdict as the strongest (Physical-tier) evidence. It is the
> answer to "VDSL2 profile 8a/12a/17a/35b — detect that, and the equivalents for
> the other technologies, for real."

## 1. What it captures today

When authorized CPE telemetry exposes them, the agent now reads and normalizes:

| Medium | Properties captured |
|---|---|
| **DSL** | mode (ADSL / ADSL2 / ADSL2+ / VDSL2 / G.fast / SHDSL), ITU standard (G.992.x / G.993.2 / **G.993.5** / G.9701), **VDSL2 profile (8a, 8b, 8c, 8d, 12a, 12b, 17a, 30a, 35b)** with its nominal band (MHz), **vectoring / super-vectoring**, SNR margin (up/down), attenuation (up/down), attainable & sync rates (up/down), interleaving/fast path |
| **DOCSIS** | version (2.0 / 3.0 / 3.1 / 4.0, inferred from OFDM/OFDMA when not stated), OFDM (downstream), OFDMA (upstream), downstream/upstream channel counts, downstream power (dBmV), MER/SNR (dB) |
| **PON** | type (GPON / EPON / XG-PON / XGS-PON / 10G-EPON), ONT model, optical Rx/Tx power (dBm) |

These land in `models.LineProfile` (`pkg/models/lineprofile.go`), serialized under
`detected_network_context.line_profile` in every report, and printed by the CLI.

### The VDSL2 profile is a *mode*, not a speed
The system encodes the same truth the user described: `8a < 12a < 17a < 35b` by
band/rate **potential**, but the profile is the line's working mode — the
delivered rate depends on loop length, copper quality, SNR, DSLAM/MSAN support
and the plan. So the profile is reported alongside the **actual** SNR margin,
attenuation and sync rate, and the explanation explicitly notes "profile ≠
speed". `35b` is flagged as Super Vectoring (G.993.5, ~35 MHz).

## 2. Where the data comes from (authorized only)

Physical-layer stats live **inside the modem**, not on the client OS. They are
read only via standard, authorized CPE interfaces — never by logging in,
guessing credentials, or scanning:

- **TR-064** `WANDSLInterfaceConfig:GetInfo` / `WANCommonInterfaceConfig` (the
  primary source today; the `tr064_probe` now passes the full field set through).
- **SNMP** (opt-in) ADSL-LINE-MIB / DOCS-IF-MIB / vendor MIBs.
- **UPnP-IGD** `WANAccessType` and link properties.
- **Vendor APIs** where the user authorizes them.

The parser is **source-agnostic**: it takes a key/value map plus free text, so any
of these probes can feed it.

## 3. How it's built (`internal/linestats`)

```
CPE evidence ─► linestats.Parse(kv, text, source) ─► *models.LineProfile
                     │
                     ├─ detectMedium       dsl | docsis | pon  (PON/DOCSIS win over generic "dsl")
                     ├─ parseDSL           mode, standard, VDSL2 profile+band, vectoring, stats
                     ├─ parseDOCSIS        version (or inferred from OFDM/OFDMA), channels, power, MER
                     └─ parsePON           type, optical Rx/Tx, ONT model
                     │
linestats.AccessHints(lp) ─► [VDSL2, VDSL, DSL] / [DOCSIS, Cable] / [GPON, FTTH, Fiber] / ...
linestats.Summary(lp)     ─► human explanation lines
```

Design choices that make it *genuinely* work and stay testable:

- **Pure functions, no I/O.** Every parser is unit-tested from canned CPE
  maps/text (`internal/linestats/linestats_test.go`), so it is deterministic and
  identical on every OS.
- **Unit-correct numbers.** dB fields from TR-064 / ADSL-MIB are encoded in 0.1 dB
  as integers; `lookupTenthsDB` converts `"61" → 6.1 dB` but takes an explicit
  decimal (`"6.3 dB"`) literally — both forms are tested.
- **No hallucinated readings.** A bare profile token (`8a`) is only accepted with
  VDSL context; an unrelated string ("ticket #8a") yields nothing. "Super
  vectoring" sets the vectoring flag but never invents a profile number.
- **Returns nil honestly** when the CPE exposed nothing physical — nil means
  "could not read", never "absent".

## 4. How it changes the verdict

`linestats.AccessHints(lp)` returns the scoreable access-type keys
(`pkg/models` constants), most-specific first. The engine
(`internal/detection/engine.go`) feeds them in as **strong physical-layer
evidence**, so:

- the scoreboard commits the subtype (e.g. **VDSL2**) out of `Unknown`,
- `hasStrongPhysicalEvidence` is satisfied (this is true Physical-tier proof),
- `decision_quality` rises to medium/high,
- the explanation and CLI print the profile, rates, SNR margin and attenuation.

It does **not** bypass the honesty model: if no CPE data is read, `LineProfile`
is nil and the engine behaves exactly as before (weak evidence → `Unknown`).
Verified by the existing fixtures plus `engine_lineprofile_test.go`
(`scan_vdsl2_35b.json` → commits DSL/VDSL2 with profile 35b;
`scan_docsis31.json` → commits Cable/DOCSIS 3.1).

## 5. Extending it to more technologies (the recipe)

To capture fine-grained properties for any other technology, follow the same
four steps — no engine rework needed:

1. **Add fields** to the relevant struct in `pkg/models/lineprofile.go` (or add a
   new sub-profile struct, e.g. a `MobileProfile` for band/EARFCN/CA combos, a
   `SatelliteProfile` for beam/modcod). Keep them `omitempty`.
2. **Parse** them in `internal/linestats` from kv/text, reusing the `lookup*`
   helpers and adding focused regexes. Add a `detectMedium` branch if it's a new
   medium.
3. **Map** to access-type keys in `AccessHints` and add human lines in `Summary`.
4. **Test** with a canned fixture in `linestats_test.go` and (optionally) an
   end-to-end fixture under `tests/fixtures/` asserted by an engine test.

Concrete next targets (each is the "8a/12a/17a/35b for its medium"):

- **DSL**: per-band bitloading, retransmission (G.INP) state, vectoring group,
  RTX/SRA, power-management mode (L0/L2).
- **DOCSIS**: per-channel modulation (QAM-256/1024/4096), OFDM profile IDs,
  partial-service flags, low-latency DOCSIS (LLD) ASF.
- **PON**: split ratio, FEC state, T-CONT/GEM allocation, alarm/los thresholds;
  XGS-PON vs Combo-PON discrimination.
- **Mobile/FWA**: NR band + Sub-6/mmWave, carrier-aggregation combo, NSA/SA,
  RSRP/RSRQ/SINR trend, cell-stability (the FWA-vs-handheld signal).
- **Satellite**: LEO periodicity signature, beam/handover cadence (LEO vs GEO).

The principle is constant: **read what the line negotiated, normalize it, tier it
as Physical, and let it commit the subtype — but only when the CPE actually
exposes it, and always show the real stats next to the label.**
