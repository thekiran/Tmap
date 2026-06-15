# 09 — Implementation Roadmap

> Brief §I and Deliverable #8 (MVP task list for Codex). Starts from what exists
> today (the Go agent + detection engine) and grows to research-grade.

## Where we are (done)

The MVP detection core is built: cross-platform Go agent, probe layer, detection
engine (normalize → fingerprint → rule-score → classify → **decision/Unknown
gate** → explain), YAML rules, evidence tiers, confidence breakdown, score
contributions, deterministic fixture tests. This is the foundation; nothing below
replaces it.

## Phases

### MVP+ (harden the agent, add physical-layer depth) — *current focus*
- **Physical-layer property capture** (this is what the new requirement targets):
  parse and normalize fine-grained line properties from authorized CPE evidence —
  VDSL2 **profiles** (8a/8b/8c/8d/12a/12b/17a/30a/35b), vectoring, DSL
  attenuation/SNR-margin/attainable+sync rates/interleaving; DOCSIS version +
  OFDM/OFDMA + channels + power + MER; PON type + optical Rx/Tx + ONT model.
  → map to subtypes with `Physical` evidence strength. (See
  [12-physical-layer-detection.md](12-physical-layer-detection.md) and the
  `linestats` package.)
- Subtype layer in the classifier wired to those normalized fields.
- More modem fingerprints + ISP patterns across ≥3 regions (data, not code).
- Local HTTP API wrapping the runner; SQLite scan history.

### v1 (cloud + learning loop)
- Backend ingest API + queue/workers + Postgres/TimescaleDB ([08](08-database-schema.md)).
- Browser runner (NDT7 + WebRTC RTT + loaded latency).
- Rule baseline served in the cloud; `feedback_labels` capture in the UI.
- First calibrated LightGBM family-level model once the minimum dataset
  ([05 §8](05-dataset-strategy.md)) exists; per-region calibration; shadow → canary.
- End-user dashboard + per-scan HTML/PDF report.

### v2 (multi-mode + scale)
- Android app (radio stats, FWA-vs-mobile via cell stability).
- Authorized CPE telemetry mode hardened (SNMP/TR-064/vendor APIs, all opt-in).
- Hierarchical ML (domain→medium→family→subtype) + ambiguity/OOD gate + SHAP.
- Global probe fleet for multi-vantage remote-only; public aggregate API.
- Migrate hot path (ingest+serving) to Go/Rust per [02](02-tech-stack.md).

### Research-grade
- Active learning loop driven by uncertainty/OOD; weak-supervision label model.
- Packet-pair/train capacity & medium-signature analysis; LEO periodicity detector.
- Per-region datasheets, model cards, reproducible snapshots; open dataset release
  (k-anonymized).
- Continuous drift monitoring with auto-retrain triggers.

## Deliverable #8 — MVP+ task list (Codex-ready, small & testable)

Each task is scoped to be implementable and unit-tested in isolation. Tasks
marked **[built now]** are implemented in this session (physical-layer capture).

1. **[built now] `internal/linestats` package** — pure functions that take raw
   CPE evidence (maps/strings from TR-064/SNMP/UPnP-IGD) and return a normalized
   `LineProfile` struct (DSL/DOCSIS/PON). No I/O → fully unit-testable.
2. **[built now] VDSL2 profile parser** — recognize `8a/8b/8c/8d/12a/12b/17a/30a/35b`
   (and synonyms like "profile 35b", "VDSL2 35b", "Super Vectoring", "G.993.5"),
   attach band/MHz/vectoring metadata and a max-rate ceiling.
2b. **[built now] DSL line-stat parser** — attenuation, SNR margin, attainable &
    sync (up/down) rates, interleaving/fast-path, ADSL/ADSL2+/VDSL2/G.fast mode.
3. **[built now] DOCSIS parser** — version (2.0/3.0/3.1/4.0), OFDM/OFDMA, ds/us
   channel counts, power, SNR/MER.
4. **[built now] PON parser** — GPON/EPON/XGS-PON/10G-EPON, optical Rx/Tx dBm,
   ONT model.
5. **[built now] Subtype mapper** — `LineProfile` → access subtype +
   `EvidenceStrong` `Physical` contribution + human explanation line.
6. **[built now] Engine wiring** — surface `LineProfile` in `WANSignal` /
   `NetworkContext`, feed `ScoreContribution`s, and let it commit a subtype out of
   `Unknown` (it is Physical-tier).
7. **[built now] Fixtures + tests** — golden CPE evidence per technology
   (incl. a 35b super-vectoring line and a 17a line) asserting the normalized
   profile and the resulting subtype/explanation.
8. SNMP/TR-064 probes populate the raw evidence maps the parser consumes
   (probes already exist; ensure they pass through the relevant OIDs/fields).
9. Local HTTP API endpoint returning the `ScanResult` (wraps the runner).
10. SQLite history store keyed by `scan_id` (schema mirrors [08](08-database-schema.md)).

## Sequencing principle

Never let coverage outrun honesty: ship the **rule baseline + physical-layer
capture** first (high-precision, explainable, offline), harvest labels, and only
turn on the ML model in a region once it beats the baseline there
([06 §5](06-accuracy-expectations.md)).
