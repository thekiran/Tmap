# 04 — Classification Strategy

> Brief §D and Deliverables #4 (ML pipeline design) and #6 (rule-based classifier
> pseudocode). Generalizes the MVP's existing pipeline
> (`agent/internal/detection`) into a learning system without abandoning its
> honesty guarantees.

## 0. The contract every classifier obeys

Whatever produces a verdict — rule baseline or ML — must emit:

- a **distribution** over access types (summing to 1, including `Unknown`),
- a **calibrated confidence** (so 0.8 means right ~80% of the time),
- a **hierarchical** decomposition (category → type → subtype),
- `score_contributions` (every point traced to a probe + evidence class),
- `uncertainty_reasons` when not committing,
- and it must **respect the Physical-evidence gate**: no commit out of `Unknown`
  without Physical-tier evidence, no matter how confident the ML "feels".

## 1. Hierarchical classification

Classify top-down; each level is its own calibrated model, and lower levels only
run (and only count) when the parent is confident. This matches `AccessCandidate`
(`category` → `type` → `subtype`) in `pkg/models`.

```
Level 0  domain:    Fixed | Mobile | Satellite | Enterprise        (+ Unknown)
Level 1  medium:    Wired | Wireless           (within Fixed)       (+ Unknown)
Level 2  family:    DSL | Fiber | Cable | FWA | WISP | Ethernet     (+ Unknown)
Level 3  subtype:   ADSL2+/VDSL2/G.fast | FTTH-GPON/XGS-PON/AE |
                    DOCSIS3.0/3.1/4.0 | 4G-FWA/5G-FWA | LTE/5G-NSA/5G-SA | GEO/MEO/LEO
```

Rules:
- **Confidence is the product down the chosen path** (Level0 × Level1 × … ),
  then re-calibrated. A confident family with an uncertain subtype yields a
  confident *family* verdict and an explicit "subtype undetermined" — never a
  falsely-precise leaf.
- **Margin is judged at the level that matters.** ADSL2+ vs VDSL2 being close is
  *not* ambiguity (still confidently DSL); DSL vs Fiber vs Cable being close *is*.
  (This is exactly the MVP's category-level margin rule.)
- **Hybrids** (Fiber+DSL, LTE+DSL, Satellite+LTE, SD-WAN) are modeled as a
  multi-label flag at Level 2 ("bonded/failover detected"), surfaced when
  multipath/reordering features + two plausible media co-fire — not forced into a
  single leaf.

## 2. Stage 1 — Rule-based first-pass classifier

The rule baseline is the floor: explainable, offline-capable, and the safety net
when ML is unavailable or uncertain. It is the **same YAML engine the agent
already runs** (`internal/scoring` + `rules/*.yaml`), executed server-side over
the full evidence set.

### Deliverable #6 — Rule-based classifier pseudocode

```pseudo
function rule_classify(evidence, capability_manifest, enrichment):
    # scores are points in pkg/models access-type keys; engine normalizes 100 -> 1.0
    scores      = zeros over ACCESS_TYPES
    contributions = []          # ScoreContribution[]  (auditable)
    physical    = false         # any Physical-tier evidence seen?
    strengths   = {Physical:none, Device:none, Network:none, Performance:none}

    # --- 2a. Strongest single signal: modem/CPE fingerprint ---
    model = normalize(evidence.cpe.model)
    fp = fingerprint_match(model, modem_fingerprints.yaml)
    if fp.matched:
        add(scores, fp.primary_type,  WEIGHT_FINGERPRINT_HINT, contributions, "fingerprint", Device)
        for t in fp.supported_techs:
            add(scores, t, WEIGHT_FINGERPRINT_SUPP, contributions, "fingerprint-supported", Device)
        strengths.Device = strong
        # a single-medium CPE (e.g. DSL-only modem) is physical proof of family
        if fp.medium_is_exclusive: physical = true; strengths.Physical = strong

    # --- 2b. Physical-layer ground truth (CPE telemetry, when authorized) ---
    for sig in evidence.wan_signals:           # DSL stats / DOCSIS / PON / WAN-iface
        add(scores, sig.implied_type, WEIGHT_STRONG_ACCESS_HINT, contributions, sig.source, Physical)
        physical = true; strengths.Physical = strong
        if sig.subtype: add(scores, sig.subtype, WEIGHT_STRONG_ACCESS_HINT, ..., Physical)

    # --- 2c. WAN interface name (ptm0/atm0/gpon/ont/lte) ---
    for iface in evidence.interface_names:
        for rule in interface_patterns.yaml where match(iface, rule):
            apply(rule.add_score, scores, contributions, "interface", Physical)
            physical = true

    # --- 2d. Mobile radio type is authoritative for the mobile branch ---
    if evidence.radio.connection_type in {LTE, NR_NSA, NR_SA}:
        add(scores, map_radio(evidence.radio), WEIGHT_STRONG_ACCESS_HINT, ..., Physical)
        physical = true; strengths.Physical = strong
        if stationary(evidence.radio):   # stable cell + low RSRP variance
            nudge(scores, FWA)           # FWA vs handheld-mobile

    # --- 2e. YAML access/ISP/text rules (medium signals) ---
    for rule in access_rules.yaml + isp_patterns.yaml:
        if any_group_matches(rule.if, evidence, enrichment):
            apply(rule.add_score, scores, contributions, rule.id, tier_of(rule))
            strengths[tier_of(rule)] = max(strengths[tier_of(rule)], rule.strength)

    # --- 2f. Network-tier hints: CGNAT, double-NAT, IPv6, path tokens ---
    if enrichment.cgnat:            nudge(scores, [Mobile, FWA, WISP], WEIGHT_PROBE_HINT, Network)
    for token in path_tokens(evidence.traceroute):     # dsl/vdsl/cmts/bras/gpon/lte
        nudge(scores, token.implied_type, WEIGHT_PROBE_HINT, Network)
    strengths.Network = grade(enrichment)

    # --- 2g. Performance bands: weak corroboration ONLY (tiny weights) ---
    apply_latency_bands(scores, evidence.performance)   # see weights.go: 2..15 pts
    apply_asymmetry(scores, evidence.performance)       # symmetric->Fiber/Eth, asym->DSL/Cable
    apply_bufferbloat(scores, evidence.performance)
    strengths.Performance = grade(evidence.performance)

    # --- 2h. Regional prior (narrow, never decide) ---
    prior = operator_region_prior(enrichment.asn, enrichment.country)
    blend_prior(scores, prior, cap = SMALL)             # cannot create Physical evidence

    normalize(scores)                                   # 100 pts -> 1.0
    candidates = build_hierarchy(scores)                # category/type/subtype tree
    confidence = rule_confidence(scores, strengths, corroboration_count(contributions))
    return Verdict(scores, candidates, confidence, contributions, strengths, physical)
```

The decision/Unknown gate (§5) runs *after* this, identically for rule and ML
output.

## 3. Stage 2 — ML classifier

### Why gradient-boosted trees
Tabular, heterogeneous, many missing values, non-linear thresholds (e.g. "RTT
≥ 500ms ⇒ GEO"), and a need for fast, explainable inference → **LightGBM**
(primary), **XGBoost**, **CatBoost** (great with high-cardinality categoricals
like ASN). Deep nets are not justified for this tabular regime and hurt
explainability. Use a **calibrated soft-voting ensemble** of the three (or
LightGBM alone for the MVP model).

### Hierarchical model layout
- One classifier per level (L0 domain, L1 medium, L2 family) — multiclass with an
  explicit `Unknown`/`abstain` class trained on genuinely ambiguous samples.
- Subtype (L3) models are per-family (DSL-subtype, Fiber-subtype, …) and run only
  when the family is confident. Subtype models lean heavily on Physical features
  (DSL stats, DOCSIS version, PON markers, radio type) — and when those are
  absent, they are *designed to abstain* rather than guess.
- The rule-baseline score vector is **fed in as features** to the ML model
  (stacking), so ML starts from the explainable prior and only moves the
  distribution where global data justifies it.

### Handling missing modalities
Train with **modality dropout** (randomly hide CPE/mobile/path features) so the
model is robust to the common case where most physical evidence is absent. The
`*_present` flags + capability manifest let the model learn "trust performance
features more when physical evidence is missing, but never claim a subtype".

## 4. Confidence calibration

Raw model scores are **not** probabilities. Calibrate per **(mode × region)**
because reliability differs wildly (CPE telemetry is sharp; remote-only is soft):

- **Method:** isotonic regression (enough data) or temperature scaling (sparse
  slices); for the hierarchical product, calibrate the *final path probability*.
- **Validation:** reliability diagrams + **ECE/MCE** per slice; the promotion gate
  in [06](06-accuracy-expectations.md) requires ECE below a threshold per region.
- **Confidence breakdown** (already a model field, `ConfidenceBreakdown`): report
  `classification` vs `context` separately — context (ISP/NAT/IPv6) can be high
  even when the type is `Unknown`. Calibrate them independently.
- **Abstention is calibrated too:** the Unknown gate's thresholds are chosen on a
  validation set to hit a target *selective accuracy* (accuracy on committed
  predictions) — e.g. "≥90% correct among everything we commit to". This is the
  knob that trades coverage for trust.

## 5. Ambiguity detection & the `Unknown` class

`Unknown` is a real, trained, first-class outcome — not a fallback bucket. A
verdict is downgraded to `Unknown` (keeping the full ranked candidates and
scores, exactly like the MVP) when **any** hold:

1. **No Physical-tier evidence** of the access type (the hard gate). Performance/
   ASN/PTR/latency narrow and contextualize but never commit on their own.
2. Calibrated top-1 **category** confidence `< floor(mode, region)`.
3. Top-2 **category** margin `< δ` (too close to call at the level that matters).
4. **Rule-vs-ML disagreement** beyond a threshold (the models contradict).
5. **Out-of-distribution**: the feature vector is far from training support
   (isolation-forest / Mahalanobis flag) → the model has no business committing.

Each firing reason becomes a human-readable `uncertainty_reason` and drives the
**`next_best_probes`** suggestion ("run the agent / authorize SNMP to read the
WAN interface — that would likely resolve DSL vs Fiber"). This is already the
shape of the MVP's `NextBestProbe`.

## 6. Explainability

- **Per-prediction:** SHAP values (TreeExplainer — exact and fast for GBDTs)
  mapped onto the evidence tiers, rendered as "the verdict rests mostly on:
  modem fingerprint (Physical, +0.34), symmetric throughput (Performance, +0.08),
  ASN prior (Regional, +0.05)". The `score_contributions` array already carries
  this audit trail; SHAP refines the ML portion.
- **Counterfactual / next-best:** "If the WAN interface had been readable, this
  would likely commit to Fiber." Drives `next_best_probes`.
- **Global:** SHAP summary + per-feature importance per model version, shipped in
  the **model card** and the operator dashboard (drift in importances = retrain
  signal).
- **Honesty rendering:** the UI always shows confidence + top alternatives +
  "why uncertain", never a bare label. See [11](11-ui-ux-brief.md).

## 7. Deliverable #4 — ML pipeline design

```
                         ┌────────────────────────────────────────────────┐
   feedback_labels ─────►│ 1. LABEL ASSEMBLY                                │
   CPE-confirmed truth ─►│   join labels to measurements; weight by source  │
   ISP known-plan ──────►│   trust (CPE > ISP-plan > self-report); dedup    │
                         └───────────────────┬──────────────────────────────┘
   measurements ───────► 2. FEATURE BUILD ───┘  (same code path as serving)
                            └─► feature store (Parquet/Feast) + schema version
                                          │
                         3. SPLIT  (region-stratified, time-based holdout,
                            group-by-network to prevent leakage across rows
                            from the same household/ASN)
                                          │
                         4. TRAIN  per level: LightGBM/XGBoost/CatBoost
                            + Optuna HPO + modality dropout + class weights
                            + explicit Unknown/abstain class
                                          │
                         5. CALIBRATE  isotonic/temperature per (mode×region)
                                          │
                         6. EVALUATE  per-class & per-region accuracy, macro-F1,
                            selective-accuracy@coverage, ECE/MCE, confusion,
                            OOD detector fit, SHAP summary  ──► MLflow
                                          │
                         7. GATE  promotion checks (no per-region regression,
                            ECE ≤ target, Unknown-rate within band) — see §F
                                          │
                         8. REGISTER  model_versions row + ONNX export +
                            SHAP explainer + model card + feature-schema hash
                                          │
                         9. SHADOW  serve alongside current; compare full
                            distributions on live traffic (not just argmax)
                                          │
                        10. PROMOTE  behind a flag, per-region canary; auto-
                            rollback on accuracy/Unknown-rate alarm
                                          │
                        11. MONITOR  drift (feature & prediction), per-region
                            accuracy from incoming labels → triggers retrain
```

Reproducibility: every `predictions` row pins `model_version` + feature-schema
hash; every model version pins its training data snapshot id, code commit, and
calibration set. You can always answer "why did we say that, with which model,
on which features".

## 8. Failure & fallback ladder (production)

1. ML + calibration healthy → calibrated hierarchical verdict.
2. ML up, OOD flag high → force `Unknown`, suggest next-best probes.
3. ML serving down → **rule baseline** answers, confidence capped, flagged
   `degraded:rule-only`.
4. Enrichment down → classify on local/physical evidence; mark missing-evidence;
   re-classify later from stored raw when enrichment recovers.
5. Nothing but weak evidence → `Unknown` with visible candidates + context (the
   MVP's current honest default).
