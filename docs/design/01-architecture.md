# 01 — System Architecture

> Brief §A and Deliverable #1 (full architecture document) + #2 (module breakdown, see [README](README.md)).

## 1. Goals and constraints that shape the architecture

1. **Probabilistic, explainable, honest.** Every output is a calibrated
   distribution over access types + a first-class `Unknown`, with the evidence
   that produced it. This is enforced *structurally*: the verdict object cannot
   be serialized without `confidence_breakdown`, `score_contributions`, and
   (when applicable) `uncertainty_reasons` — the same discipline the MVP already
   applies in `agent/internal/detection/decision.go`.
2. **Many collectors, one schema.** Remote probes, browser, desktop agent,
   mobile, and CPE telemetry all produce the same `ProbeResult` envelope and a
   `ScanResult`. New evidence drops in without engine or UI rework (this is the
   existing probe contract, scaled).
3. **Two-tier classification: edge + cloud.** The agent can produce a verdict
   fully offline (rule baseline). The cloud re-classifies with the ML model and
   global enrichment when the user consents to upload. Neither tier is required
   for the other to function.
4. **Privacy and consent first.** No raw payloads, no unauthorized access, no
   third-party scanning. Consent state travels with every measurement. See
   [10-security-ethics.md](10-security-ethics.md).
5. **Global, not Turkey-only.** All ISP/region knowledge is data (rules,
   fingerprints, ASN tables, regional priors), never code. Regions are added by
   shipping data + labels, not by recompiling.

## 2. The 30,000-ft view

```
        COLLECTORS (evidence producers)                 CLOUD (classify + learn + serve)
 ┌───────────────────────────────────┐         ┌──────────────────────────────────────────────┐
 │ Remote-only probes (server-side)  │         │  API gateway  ── authn, consent, rate-limit    │
 │ Browser runner (WASM/JS)          │  HTTPS  │      │                                          │
 │ Desktop agent (Go, win/linux/mac) │ ──────► │  Ingest service ── validate, dedup, abuse gate │
 │ Mobile app (Android/iOS)          │  gRPC   │      │                                          │
 │ Authorized CPE readers (SNMP/...) │         │   Message queue (measurements, enrich, train)  │
 └───────────────────────────────────┘         │      │                                          │
            ▲        each emits                 │  Workers ──┬─ feature builder                  │
            │     ProbeResult/ScanResult        │            ├─ enrichment (ASN/BGP/rDNS/geo)     │
            │     (pkg/models JSON)              │            ├─ traceroute/path analyzer         │
            │                                    │            └─ classification orchestrator       │
            │                                    │      │                                          │
            │   verdict + evidence + report      │   Classifier  ── rule baseline → ML → calibrate │
            └────────────────────────────────────│                   → ambiguity/Unknown gate     │
                                                  │      │                                          │
                                                  │  Stores: Postgres (OLTP), TSDB (time-series),  │
                                                  │          object store (raw artifacts),         │
                                                  │          feature store, model registry         │
                                                  │      │                                          │
                                                  │  Reporting + Public API + Dashboard            │
                                                  └──────────────────────────────────────────────┘
                                                            │
                                                  ML training loop (offline) ◄── feedback labels
```

## 3. Components (Brief §A, one by one)

### 3.1 Collectors

A collector's only job is to gather evidence honestly and label its strength.
It never decides the access type alone; it emits `ProbeResult`s. Five families:

| Collector | Runtime | Representative evidence | Strongest signal it can get |
|---|---|---|---|
| **Remote-only** | Cloud workers + global probe fleet (RIPE-Atlas-style) | IP intel, ASN/BGP prefix, rDNS, multi-region ping, TCP-connect RTT, UDP probe, Paris traceroute, packet-pair/train dispersion, throughput, loaded latency, loss, jitter, TCP_INFO/retransmits | medium (rDNS/ASN tokens, CGNAT, path shape) |
| **Browser** | TS + WASM in page | NDT7 throughput, WebRTC/WebSocket RTT, loaded latency, `navigator` hints, RTT distribution | weak–medium |
| **Desktop agent** | Go binary (existing) | adapters, Wi-Fi RSSI/freq/width, Ethernet link speed, route table, DNS, NAT/CGNAT (STUN/PCP), UPnP modem model, optional Npcap capture (opt-in) | medium–strong (modem fingerprint, link speed) |
| **Mobile/FWA** | Android (Kotlin) first, iOS later | RSRP/RSRQ/RSSI/SINR, NR NSA/SA, cell ID (permitted), signal stability, handover/latency fluctuation | strong (radio type is authoritative for mobile) |
| **CPE telemetry** | Authorized readers | SNMP, UPnP/IGD, TR-064, vendor APIs; DSL line stats, DOCSIS channels, PON optical power/ONT model, WAN interface type | **strong (physical-layer ground truth)** |

Design notes:
- Each collector ships a **capability manifest** (which probes it ran, why
  others were skipped). The classifier weights evidence by *what was possible*,
  so a browser verdict is not penalized for lacking DSL stats it could never
  read.
- Collectors are **rate-limited and consent-gated at the source** and again at
  ingest (defense in depth).
- The desktop/CPE/mobile collectors reuse the agent's `Runner` model:
  concurrent probes, per-probe timeout, a failing probe degrades gracefully and
  is recorded as `failed` — never aborts the scan.

### 3.2 Backend API

Two surfaces behind one gateway:

- **Client API** (collectors → cloud): submit measurements, fetch a verdict,
  post feedback labels, register devices. gRPC for the agent/mobile (typed,
  streaming, compact); REST/JSON mirror for browser and third parties.
- **Public API** (read, for integrators): "classify this measurement bundle",
  "what's the access type for AS X / prefix Y aggregate", model metadata. Keyed,
  quota'd, versioned (`/v1`). Full schemas in [07-api-design.md](07-api-design.md).

The gateway owns: TLS termination, authn (API keys / OAuth device flow / mTLS
for the fleet), **consent verification**, global + per-key rate limits, request
size caps, and schema validation. It is intentionally thin — it never
classifies; it routes to ingest.

### 3.3 Queue / worker system

- **Why a queue:** measurement bundles arrive bursty (a country-wide test
  campaign), enrichment hits rate-limited third-party services (RDAP, BGP
  feeds), and classification should not block the HTTP request. Ingest writes a
  durable record and enqueues; workers do the slow work; the client polls or
  gets a webhook/WebSocket push when the verdict is ready.
- **Topics/queues:** `measurements.raw`, `enrich.asn`, `enrich.path`,
  `features.build`, `classify`, `report.render`, `labels.train`. Dead-letter
  queue per topic; idempotent consumers keyed by `measurement_id`.
- **Worker types** (stateless, horizontally scaled):
  - *Validator/deduper* — schema + plausibility checks, drop replays.
  - *Enrichment* — ASN/BGP/rDNS/geo/IXP/anycast, all cached (see 3.6).
  - *Path analyzer* — Paris-traceroute parsing, first/second-hop RTT,
    CGNAT-hop detection, hop-pattern features.
  - *Feature builder* — assembles the feature vector (see
    [03-feature-engineering.md](03-feature-engineering.md)) and writes it to the
    feature store.
  - *Classification orchestrator* — runs rule baseline → ML → calibration →
    ambiguity gate; persists a `prediction` row.
  - *Report renderer* — HTML/PDF on demand.

### 3.4 Classifier (rule baseline + ML)

Two-stage, designed so the rule baseline is always a sufficient fallback:

```
features ─► RULE BASELINE (the agent's YAML engine, server-side)
                │ produces an initial score vector + hard physical-evidence flags
                ▼
            ML MODEL (hierarchical LightGBM/XGBoost/CatBoost ensemble)
                │ refines the distribution using the full global feature set
                ▼
            CALIBRATION (isotonic / temperature per region+mode)
                │ turns raw scores into trustworthy probabilities
                ▼
            AMBIGUITY + UNKNOWN GATE (the decision layer, generalized)
                │ if no strong physical evidence, or top-2 categories too close,
                │ or calibrated confidence below the mode's floor → Unknown
                ▼
            VERDICT  (distribution + confidence_breakdown + score_contributions
                      + uncertainty_reasons + SHAP-derived explanation)
```

The rule baseline and the ML model **vote into the same score space**
(`pkg/models` access-type keys), so their disagreement is itself a feature and
an audit signal. Details in [04-classification-strategy.md](04-classification-strategy.md).

### 3.5 Database schema (OLTP)

PostgreSQL is the source of truth for entities and verdicts:
`users`, `devices`, `measurements`, `probes`, `traceroutes`, `speed_tests`,
`signal_samples`, `cpe_stats`, `predictions`, `model_versions`,
`feedback_labels`. Full DDL in [08-database-schema.md](08-database-schema.md).
Design choices: append-only measurement facts, soft-delete + crypto-shred for
privacy, every prediction references the `model_versions.id` that produced it
(reproducibility), and `feedback_labels` is the spine of the training loop.

### 3.6 Time-series & enrichment stores

- **TSDB** (TimescaleDB hypertable on the same Postgres, or ClickHouse at
  scale) for `signal_samples` and repeated `speed_tests` — anything sampled over
  time (RSRP trend, speed stability, time-of-day variation, loaded-latency
  series). These are *features* (stability, variance, diurnal pattern), so they
  live where range queries are cheap.
- **Enrichment cache** (Redis + a periodically-refreshed Postgres table) for
  ASN→org, prefix→ASN (from a BGP feed / RouteViews / RIPE RIS), rDNS, geo,
  IXP membership, known-anycast ranges, and the CGNAT ranges (100.64/10). Cached
  because they're slow, rate-limited, and change slowly.
- **Object store** (S3-compatible) for bulky raw artifacts: full traceroute
  JSON, packet-pair timing arrays, optional pcap (opt-in, retention-limited),
  rendered reports. Postgres stores a pointer + hash, not the blob.

### 3.7 ML pipeline

Offline, reproducible, gated. Feature store build → label join (from
`feedback_labels` + CPE-confirmed ground truth) → train hierarchical models →
calibrate → evaluate (per region, per class, calibration curves, slice drift) →
register in `model_versions` → **shadow** in production → promote behind a flag.
SHAP explainers are versioned alongside each model. Full design in
[04-classification-strategy.md](04-classification-strategy.md) and
[05-dataset-strategy.md](05-dataset-strategy.md).

### 3.8 Dashboard, reporting, public API

- **Dashboard** — two audiences: end users ("what is my connection, and why,
  and how sure are you") and operators/researchers (fleet health, accuracy by
  region/ASN, drift, label queue, model comparison). Brief in
  [11-ui-ux-brief.md](11-ui-ux-brief.md).
- **Reporting** — per-scan HTML/PDF (verdict, evidence table, confidence
  breakdown, "what would raise confidence"), and aggregate exports for
  researchers (privacy-reviewed).
- **Public API** — the read surface above; the same verdict schema, plus
  aggregate endpoints that never expose an individual.

## 4. Data flow (end-to-end, one measurement)

1. A collector runs its probe set (consent + rate-limit checked locally),
   producing a `ScanResult` with `ProbeResult` evidence and a *local* rule
   verdict.
2. If the user consents to upload, the collector POSTs the bundle to the client
   API. The gateway authenticates, re-checks consent, rate-limits, validates.
3. Ingest persists `measurements` + child rows (`probes`, `traceroutes`,
   `speed_tests`, `signal_samples`, `cpe_stats`), writes raw blobs to object
   store, enqueues `enrich`/`features.build`.
4. Workers enrich (ASN/BGP/rDNS/geo), parse paths, and build the feature vector
   into the feature store.
5. The classification orchestrator runs rule baseline → ML → calibration →
   ambiguity gate and writes a `predictions` row (referencing `model_versions`).
6. The client gets the verdict (poll/push). The dashboard and reports read it.
7. If the user later confirms/corrects ("my plan is actually VDSL2"), that
   becomes a `feedback_labels` row → the next training cycle.

## 5. Cross-cutting concerns

- **Versioning:** `pkg/models` schema version travels in every bundle; the API
  is `/v1`; every prediction pins its `model_version`. Old collectors keep
  working (additive fields only).
- **Idempotency & dedup:** `measurement_id` (client-generated UUID) is the
  idempotency key end-to-end.
- **Observability:** structured logs, traces across queue hops (trace id =
  `measurement_id`), and golden dashboards for ingest lag, classify latency,
  Unknown-rate, and per-region accuracy.
- **Failure posture:** ML serving down → rule baseline still answers (lower
  confidence, flagged). Enrichment down → classify on what's available, mark
  missing-evidence. A collector probe failing never fails the scan.
- **Multi-region:** stateless API + workers deploy per region; the global probe
  fleet is inherently multi-region (it's a feature: multi-vantage RTT). Postgres
  primary + read replicas; TSDB and object store regional with async replication.

## 6. Why this shape (trade-offs)

- **Queue between ingest and classify** adds latency (seconds) but buys burst
  absorption, ret/backpressure on rate-limited enrichment, and a clean place to
  re-classify historical data when a new model ships. Worth it.
- **Rule baseline kept in production forever** (not just as a bootstrap) costs
  maintenance but guarantees an explainable floor and a safety net when ML is
  uncertain or unavailable — and it's the only thing that can run on-device
  offline. Worth it.
- **One schema across all modes** constrains each collector (they must speak
  `pkg/models`) but is the single biggest force multiplier: the engine, UI,
  storage, and ML all share it. Non-negotiable.
