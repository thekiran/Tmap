# Internet Access Intelligence Platform — Design Dossier

This folder is the full product/technical design for scaling the current MVP
(the cross-platform Go `iad-agent`) into a **global, production-grade system
that estimates a host's internet access technology with a calibrated
probabilistic confidence** — and that is honest about when it cannot tell.

> **Reality constraint (non-negotiable, baked into every document).**
> Remote-only measurements **cannot** identify an access technology with 100%
> certainty. Every verdict is a probability distribution with a calibrated
> confidence, explainable evidence, explicit ambiguity warnings, and a
> first-class `Unknown` class. This is already how the MVP behaves
> (`decision_quality`, `uncertainty_reasons`, `Unknown` with visible
> candidates); the platform generalizes it, it does not abandon it.

## How this maps to the brief

| Brief section | Document | Deliverable(s) covered |
|---|---|---|
| A. System architecture | [01-architecture.md](01-architecture.md) | #1 Architecture doc, #2 Module breakdown |
| B. Technology stack | [02-tech-stack.md](02-tech-stack.md) | — |
| C. Feature engineering | [03-feature-engineering.md](03-feature-engineering.md) | #5 Detection feature table |
| D. Classification strategy | [04-classification-strategy.md](04-classification-strategy.md) | #4 ML pipeline, #6 Rule-based pseudocode |
| E. Dataset strategy | [05-dataset-strategy.md](05-dataset-strategy.md) | — |
| F. Accuracy expectations | [06-accuracy-expectations.md](06-accuracy-expectations.md) | — |
| G. API design | [07-api-design.md](07-api-design.md) | #7 API schema |
| H. Database design | [08-database-schema.md](08-database-schema.md) | #3 Data model |
| I. Implementation roadmap | [09-roadmap.md](09-roadmap.md) | #8 MVP task list for Codex |
| J. Security & ethics | [10-security-ethics.md](10-security-ethics.md) | #9 Production hardening checklist |
| K. Deliverables | this index + all docs | #10 UI/UX brief → [11-ui-ux-brief.md](11-ui-ux-brief.md) |
| **Implemented feature** | [12-physical-layer-detection.md](12-physical-layer-detection.md) | VDSL2 profiles (8a/12a/17a/35b), DSL/DOCSIS/PON line capture — **already code** |

## What already exists (the starting point)

The repository root already ships the **MVP detection core** described in
[../architecture.md](../architecture.md) and [../detection-methods.md](../detection-methods.md):

- A Go agent (`agent/`, module `github.com/thekiran/iad`) with a probe layer,
  a detection engine (normalize → fingerprint-match → rule-score → classify →
  confidence → **decision layer** → explain), YAML rules, and deterministic
  fixture tests.
- Shared data contracts in `agent/pkg/models` (`ProbeResult`, `ScanResult`,
  `AccessCandidate`, `NetworkContext`, `ConfidenceBreakdown`,
  `ScoreContribution`, evidence-strength tiers).
- A philosophy — *no single signal decides; score, calibrate, and show why* —
  that the rest of this design simply scales to the world.

**Design rule for everything below: reuse those contracts.** The cloud
backend, the browser runner, the mobile app, and the ML pipeline all serialize
to / from the same JSON shapes the agent already emits, extended (never forked)
where a new mode needs new fields.

## Deliverable #2 — Module breakdown (platform-wide)

```
iad/                                  monorepo root
├── agent/                            EXISTING Go agent (desktop + CLI collector)
│   ├── cmd/iad-agent/                CLI entrypoint
│   ├── internal/probes/              one ProbeResult per evidence source
│   ├── internal/detection/           normalize→match→score→classify→decide→explain
│   ├── internal/scoring/             YAML rule engine + code-side weights
│   ├── internal/system,network/      build-tagged OS forks (win/linux/darwin)
│   └── pkg/models/                   SHARED CONTRACTS (the canonical schema)
│
├── collectors/                       NEW — non-desktop evidence sources
│   ├── browser/                      TS/WASM web runner (NDT7, WebRTC, loaded latency)
│   ├── mobile-android/              Kotlin app (TelephonyManager, WifiInfo)
│   ├── mobile-ios/                   Swift app (CoreTelephony, NEHotspotNetwork)
│   └── cpe/                          authorized SNMP/TR-064/UPnP-IGD/vendor-API readers
│
├── backend/                          NEW — cloud ingest + classification service
│   ├── api/                          public + client REST/gRPC (the contract in §G)
│   ├── ingest/                       measurement intake, validation, dedup, abuse gate
│   ├── workers/                      queue consumers: feature build, traceroute parse, geo/ASN enrich
│   ├── classifier/                   rule-baseline + ML serving + calibration + ambiguity
│   ├── enrichment/                   ASN/BGP/rDNS/geo/IXP/anycast lookups (cached)
│   └── reporting/                    HTML/PDF reports, aggregate stats, public API
│
├── ml/                               NEW — offline training + evaluation
│   ├── pipelines/                    feature store build, label join, train, calibrate
│   ├── models/                       LightGBM/XGBoost/CatBoost artifacts + cards
│   ├── eval/                         calibration, per-region/per-class metrics, drift
│   └── registry/                     model_versions, SHAP explainers, promotion gates
│
├── dashboard/                        NEW — operator + end-user web UI (§K #10)
├── infra/                            NEW — IaC, deploy, observability, secrets
├── rules/                            EXISTING YAML rules/fingerprints (hot-reloadable)
├── docs/                             architecture + detection methods + THIS dossier
└── tests/                            fixtures + golden + integration + dataset slices
```

Each top-level module is independently deployable and testable. The **only**
hard coupling allowed between modules is `pkg/models` (and its language ports):
the schema is the contract. See [01-architecture.md](01-architecture.md) for
data flow and ownership boundaries.

## Reading order

1. [01-architecture.md](01-architecture.md) — the shape of the whole system.
2. [04-classification-strategy.md](04-classification-strategy.md) — the brain.
3. [03-feature-engineering.md](03-feature-engineering.md) — what the brain eats.
4. [09-roadmap.md](09-roadmap.md) — how we get there from today's MVP.

Everything else is reference for the team building a specific slice.
