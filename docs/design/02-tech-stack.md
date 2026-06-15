# 02 — Recommended Technology Stack

> Brief §B. Two coherent alternatives. The project already commits to **Go for
> the desktop/CPE collector** (it exists and works), so both stacks keep that and
> differ mainly in the *backend + ML serving* tier.

## Decision summary

| Concern | Python-heavy stack | Go/Rust performance stack |
|---|---|---|
| Best when | Iterating fast on ML, research-grade experimentation, small/medium fleet | High-throughput global fleet, low per-measurement cost, tight latency SLOs |
| Backend API | FastAPI (Python) | Go (chi/connect-go) or Rust (axum) |
| Workers | Celery / Arq / Dramatiq | Go workers / Rust + Tokio |
| ML training | Python (LightGBM/XGBoost/CatBoost, scikit-learn, Optuna) | **Same** — training is Python in both |
| ML serving | In-process Python, or ONNX Runtime | ONNX Runtime (Go/Rust bindings) or Treelite-compiled trees |
| Collectors | Go agent (exists) + TS browser + Kotlin/Swift mobile | **Same** |
| Verdict for | Teams whose center of gravity is data science | Teams optimizing infra cost & latency at scale |

**Recommendation:** *Start Python-heavy* (fastest path from MVP to a learning
system; training is Python regardless), and **migrate the hot path
(ingest + serving) to Go/Rust** once volume justifies it. Export models to ONNX
from day one so serving is portable between the two without retraining.

---

## Alternative A — Python-heavy stack

| Layer | Choice | Why |
|---|---|---|
| Client/Public API | **FastAPI** + Pydantic v2 + Uvicorn/Gunicorn | Schemas == validation; OpenAPI for free; async; mirrors `pkg/models` cleanly |
| gRPC (agent/mobile) | grpcio + protobuf (or Connect) | Typed, compact, streaming for signal samples |
| Queue/broker | **Redis Streams** (start) → **RabbitMQ/Kafka** (scale) | Streams are trivial early; Kafka when you replay history for re-classification |
| Workers | **Arq** or **Dramatiq** (asyncio-friendly) | Simpler than Celery, async-native for enrichment fan-out |
| OLTP DB | **PostgreSQL 16** + **TimescaleDB** ext | One engine for relational + time-series; hypertables for `signal_samples` |
| Cache/enrichment | **Redis** | ASN/rDNS/geo cache, rate-limit counters |
| Object store | **S3 / MinIO** | Raw traceroutes, packet-pair arrays, opt-in pcap, reports |
| Feature store | Parquet on S3 + Postgres metadata (start); **Feast** (later) | Train/serve parity without premature infra |
| ML training | **LightGBM** (primary), XGBoost, CatBoost; scikit-learn; **Optuna** (HPO); **SHAP** | Gradient-boosted trees dominate tabular; SHAP gives per-prediction explanations |
| Calibration | scikit-learn isotonic / Platt; temperature scaling | Per (region × mode) calibrators |
| Experiment tracking | **MLflow** | Runs, params, metrics, artifacts → feeds `model_versions` |
| ML serving | In-process LightGBM, or **ONNX Runtime** behind FastAPI | ONNX decouples training lib from serving |
| Dashboard | **React + TypeScript + Vite**, Tailwind, Recharts/visx | Same FE in both stacks |
| Reporting | WeasyPrint / Playwright (HTML→PDF) | Reuse dashboard components for reports |
| Orchestration | **Prefect** or Airflow | Schedule training, drift checks, enrichment refresh |
| Infra | Docker + Kubernetes; Terraform; GitHub Actions | Standard, portable |
| Observability | OpenTelemetry → Prometheus + Grafana + Loki/Tempo | Traces keyed by `measurement_id` |

**Pros:** fastest research-to-production loop; one language for ingest→features
→train→serve; richest ML ecosystem. **Cons:** higher CPU/RAM per request;
GIL-bound concurrency needs more workers; serving cost grows with volume.

---

## Alternative B — Go / Rust performance-oriented stack

| Layer | Choice | Why |
|---|---|---|
| Client/Public API | **Go** (chi + connect-go) or **Rust** (axum + tower) | Low latency, high concurrency, small footprint |
| gRPC | connect-go / tonic (Rust) | First-class, shares protobuf with agent |
| Queue/broker | **NATS JetStream** or **Kafka** (redpanda) | High-throughput, durable, replayable |
| Workers | **Go** goroutines / **Rust** + Tokio | Cheap concurrency for enrichment fan-out and path parsing |
| OLTP DB | **PostgreSQL 16** + TimescaleDB / **ClickHouse** for TSDB at scale | ClickHouse for billions of `signal_samples` |
| Cache | Redis / in-proc | Same role |
| Object store | S3 / MinIO | Same |
| Feature store | Parquet + a Go/Rust feature builder writing the same schema | Train/serve parity via shared schema, not shared code |
| ML training | **Python** (LightGBM/XGBoost/CatBoost) — unchanged | Don't fight the ecosystem; training stays Python |
| ML serving | **ONNX Runtime** (Go `onnxruntime_go` / Rust `ort`) or **Treelite** compiled trees, or **leaves** (pure-Go GBDT inference) | Microsecond tree inference, no Python in the hot path |
| Calibration | Export calibrator as a tiny ONNX/coefficients blob, apply in Go/Rust | Keep serving single-language |
| Dashboard | React + TypeScript | Same |
| Reporting | Headless Chromium (Playwright) service, or a Rust PDF lib | Same outcome |
| Orchestration | Argo Workflows / Temporal (Go) | Durable workflows for training + drift |
| Infra/Observability | K8s, Terraform, OTel → Prometheus/Grafana | Same |

**Pros:** lowest latency and per-measurement cost; trivial concurrency; the
agent and backend share a language (Go) and protobuf contracts; the rule
baseline is *literally the agent's engine* compiled into the service. **Cons:**
slower ML iteration ergonomics; serving needs an ONNX/Treelite export step;
fewer turnkey data-science tools server-side (mitigated: training is still
Python).

---

## Shared across both stacks (don't re-decide these)

- **Collectors:** Go agent (exists), **TypeScript + WASM** browser runner,
  **Kotlin** Android then **Swift** iOS. The browser runner integrates an
  **NDT7** client and **WebRTC** data channels for UDP-like RTT.
- **Contract:** `pkg/models` (Go) is canonical; generate the TS and Kotlin/Swift
  types from a single **protobuf/JSON-Schema** source of truth so all collectors
  stay in lockstep. Adding a field is additive and backward-compatible.
- **Training is Python in both.** Only *serving* differs. Always export to
  **ONNX** (and keep the native model) so you can move serving between stacks
  without retraining and so the model artifact outlives the framework.
- **Rules stay YAML**, hot-reloadable, shipped as data — same files the agent
  uses, evaluated identically server-side.

## Migration path (A → B)

1. Ship A. Get data, labels, and a calibrated model.
2. Export the model to ONNX; stand up a Go/Rust serving sidecar; shadow it
   against the Python server (compare distributions, not just argmax).
3. Move ingest + feature building to Go/Rust workers once they match.
4. Keep training, eval, and the dashboard in Python/TS — they're not the hot
   path and the ecosystem is worth more there.
