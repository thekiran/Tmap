# 07 — API Design

> Brief §G and Deliverable #7. Schemas extend the existing `pkg/models` JSON
> (`ScanResult`, `ProbeResult`, `NetworkContext`, `AccessCandidate`,
> `ConfidenceBreakdown`, `ScoreContribution`) — the agent already emits most of
> this; the API just adds envelope, auth, consent, and async fields.

## 1. Conventions

- Versioned base path `/v1`. JSON over HTTPS; gRPC mirror for agent/mobile.
- `measurement_id` (client UUID) is the **idempotency key** everywhere.
- Every write carries a **consent block**; the gateway rejects writes without it.
- Additive evolution only: unknown fields ignored, new fields optional, so old
  collectors keep working. `schema_version` travels in every bundle.
- Timestamps RFC3339 UTC. All floats in documented units (ms, Mbps, %, dBm).

## 2. Endpoints

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/v1/measurements` | Submit a measurement bundle (collector → cloud). Async by default. |
| `GET` | `/v1/measurements/{id}` | Fetch bundle status + (when ready) the verdict. |
| `POST` | `/v1/classify` | **Synchronous** classify of an inline bundle (no storage; for integrators). |
| `GET` | `/v1/predictions/{id}` | Fetch a verdict by prediction id. |
| `POST` | `/v1/feedback` | Submit/correct a label (drives training). |
| `POST` | `/v1/devices` | Register a device (agent/mobile), returns device token. |
| `GET` | `/v1/aggregate/asn/{asn}` | Privacy-safe aggregate distribution for an ASN/region (no individuals). |
| `GET` | `/v1/models/current` | Active model version + card metadata. |

Auth: API key (integrators), OAuth device-flow token (agent/mobile/browser),
mTLS (the probe fleet). Per-key quotas + global rate limits at the gateway.

## 3. Input — `POST /v1/measurements` / `POST /v1/classify`

```json
{
  "schema_version": "1.0",
  "measurement_id": "9f1c2a7e-5b3d-4e21-9a44-2c7e8f0b1a33",
  "mode": "agent",                       // remote | browser | agent | mobile | cpe
  "created_at": "2026-06-15T09:31:22Z",
  "consent": {
    "granted": true,
    "scopes": ["classify", "store", "research"],   // research = may enter training corpus
    "policy_version": "2026-01",
    "allow_cpe_read": true,              // explicit authorization for CPE telemetry
    "subject_ref": "device:7be3...hash" // pseudonymous; no raw PII
  },
  "client": {
    "collector": "iad-agent",
    "collector_version": "1.4.0",
    "os": "windows",
    "capability_manifest": {            // what this collector COULD run (weights evidence)
      "can_read_cpe": true, "can_capture_packets": false,
      "can_traceroute": true, "can_radio_stats": false
    }
  },
  "network_context": {                  // factual situation (mirrors models.NetworkContext)
    "isp": "Example Telecom", "country": "TR", "asn": "AS9121",
    "public_ip_truncated": "85.96.0.0/16",      // host octets dropped at source
    "ptr": "host-85-96-x-x.example.net.tr",
    "cgnat": false, "gateway_chain": ["192.168.1.1"], "double_nat_possible": false,
    "local_access": "ethernet", "link_speed_mbps": 1000,
    "router_model": "Huawei HG8245H", "fingerprint_matched": true
  },
  "evidence": [                         // array of models.ProbeResult (unchanged shape)
    {
      "probe_name": "upnp_probe", "status": "success", "confidence": 0.9,
      "evidence": {"manufacturer": "Huawei", "model": "HG8245H", "device_type": "InternetGatewayDevice"},
      "hints": ["GPON ONT model"]
    },
    {
      "probe_name": "performance_profile", "status": "success", "confidence": 0.6,
      "evidence": {
        "idle_latency_ms": 3.1, "loaded_latency_ms": 9.4, "jitter_ms": 0.7,
        "packet_loss_pct": 0.0,
        "download_mbps": 920.0, "upload_mbps": 880.0, "asymmetry_ratio": 0.96,
        "packet_pair_capacity_mbps": 1000, "packet_train_quantization": 0.04
      }
    },
    {
      "probe_name": "cpe_stats", "status": "success", "confidence": 0.95,
      "evidence": {
        "source": "tr064", "wan_access_type": "GPON",
        "pon": {"ont_model": "HG8245H", "rx_power_dbm": -19.2, "tx_power_dbm": 2.1}
      },
      "hints": ["PON optical power present", "WAN access type GPON"]
    }
  ],
  "samples": {                          // optional time-series (signal_samples/speed_tests)
    "signal": [], "speed": []
  }
}
```

`/v1/classify` returns the verdict inline and synchronously; `/v1/measurements`
returns `202 Accepted` with a status URL and classifies asynchronously.

## 4. Output — the verdict (extends `models.ScanResult`)

```json
{
  "schema_version": "1.0",
  "prediction_id": "pred_01HZ...",
  "measurement_id": "9f1c2a7e-5b3d-4e21-9a44-2c7e8f0b1a33",
  "model_version": "iad-clf-2026.06.1",
  "rule_baseline_version": "rules-2026.06.0",
  "created_at": "2026-06-15T09:31:25Z",

  "primary_type": "FTTH",
  "category": "Fiber",
  "subtype": "GPON",

  "classification_confidence": 0.94,    // calibrated; the headline number
  "context_confidence": 0.97,           // confidence in ISP/NAT/topology facts
  "decision_quality": "high",           // low | medium | high
  "uncertainty_reasons": [],            // populated when not committing

  "distribution": [                     // calibrated, sums to ~1.0, includes Unknown
    {"category": "Fiber",  "type": "FTTH", "subtype": "GPON", "probability": 0.94},
    {"category": "Fiber",  "type": "FTTH", "subtype": "XGS-PON", "probability": 0.02},
    {"category": "Cable",  "type": "DOCSIS", "probability": 0.02},
    {"category": "Unknown", "probability": 0.02}
  ],

  "candidates": [                       // models.AccessCandidate tree (ranked)
    {"category": "Fiber", "type": "FTTH", "subtype": "GPON",
     "score": 0.91, "confidence": 0.94, "evidence_strength": "strong"}
  ],

  "confidence_breakdown": {             // models.ConfidenceBreakdown
    "classification": 0.94, "context": 0.97,
    "physical": 0.92, "device": 0.85, "network": 0.6, "performance": 0.55,
    "regional": 0.4, "penalty": 0.0
  },
  "evidence_strength": {                // strongest class observed per tier
    "physical": "strong", "device": "strong", "network": "medium",
    "performance": "medium", "regional": "weak"
  },

  "score_contributions": [              // models.ScoreContribution — full audit trail
    {"target": "FTTH", "category": "Fiber", "type": "FTTH",
     "amount": 0.45, "evidence_class": "Physical", "strength": "strong",
     "probe_name": "cpe_stats", "reason": "TR-064 WANAccessType=GPON + ONT optical power"},
    {"target": "Fiber", "amount": 0.30, "evidence_class": "Device", "strength": "strong",
     "probe_name": "upnp_probe", "reason": "GPON ONT model HG8245H fingerprint match"},
    {"target": "Fiber", "amount": 0.08, "evidence_class": "Performance", "strength": "medium",
     "probe_name": "performance_profile", "reason": "symmetric throughput (asym 0.96), low bufferbloat"}
  ],
  "explanation": [
    "Verdict rests on physical CPE evidence: TR-064 reports WANAccessType=GPON with ONT optical power.",
    "Corroborated by a GPON ONT modem fingerprint and symmetric ~920/880 Mbps throughput with low bufferbloat.",
    "Subtype GPON (not XGS-PON) inferred from ONT model and capacity ~1 Gbps."
  ],
  "shap_top_features": [                // ML explainability, mapped to tiers
    {"feature": "wan_access_type=GPON", "tier": "Physical", "contribution": 0.34},
    {"feature": "modem_fingerprint_exclusive", "tier": "Device", "contribution": 0.21},
    {"feature": "asymmetry_ratio", "tier": "Performance", "contribution": 0.07}
  ],

  "detected_network_context": { "isp": "Example Telecom", "country": "TR",
    "asn": "AS9121", "cgnat": false, "double_nat_possible": false,
    "local_access": "ethernet", "router_model": "Huawei HG8245H" },

  "next_best_probes": [],               // populated when uncertain — see below
  "warnings": []
}
```

### Uncertain example (the honest default)

```json
{
  "primary_type": "Unknown",
  "category": "Unknown",
  "classification_confidence": 0.12,
  "context_confidence": 0.78,
  "decision_quality": "low",
  "uncertainty_reasons": [
    "No strong physical evidence of the access type (behind NAT modem, CPE not readable).",
    "Top categories too close: Fiber 0.21 vs Cable 0.18 vs VDSL 0.17.",
    "Rule baseline and ML model disagree at family level."
  ],
  "distribution": [
    {"category": "Fiber", "probability": 0.27},
    {"category": "Cable", "probability": 0.24},
    {"category": "DSL", "type": "VDSL", "probability": 0.22},
    {"category": "Unknown", "probability": 0.27}
  ],
  "next_best_probes": [
    {"probe_name": "cpe_stats", "reason": "Reading WAN interface/line stats would likely resolve DSL vs Fiber vs Cable",
     "expected_evidence": "WANAccessType / DSL SNR / DOCSIS channels", "safety": "requires explicit CPE authorization"},
    {"probe_name": "desktop_agent", "reason": "UPnP modem fingerprint not available from this collector",
     "expected_evidence": "modem manufacturer/model", "safety": "LAN-only, no login"}
  ]
}
```

## 5. Feedback — `POST /v1/feedback`

```json
{
  "measurement_id": "9f1c2a7e-...",
  "prediction_id": "pred_01HZ...",
  "label_source": "self_report",       // self_report | isp_plan | cpe_confirmed | radio_confirmed
  "declared_label": "fiber 1Gbps",     // raw, as the user/ISP stated it
  "normalized": {"category": "Fiber", "type": "FTTH", "subtype": null},
  "trust_weight": 0.4,                  // server may override based on source
  "consent_scopes": ["research"],
  "notes": "user confirmed via ISP portal"
}
```

## 6. Aggregate (privacy-safe) — `GET /v1/aggregate/asn/AS9121?country=TR`

```json
{
  "asn": "AS9121", "country": "TR", "sample_size": 4120,   // suppressed if < k
  "distribution": [
    {"category": "Fiber", "share": 0.46},
    {"category": "DSL", "share": 0.38},
    {"category": "FWA", "share": 0.10},
    {"category": "Unknown", "share": 0.06}
  ],
  "k_anonymity_min": 50, "model_version": "iad-clf-2026.06.1"
}
```

## 7. Errors

Standard problem+json: `{ "type", "title", "status", "detail", "instance",
"measurement_id" }`. Notable codes: `409 measurement_exists` (idempotent replay,
returns existing prediction), `422 consent_required`, `429 rate_limited`
(+`Retry-After`), `413 payload_too_large`, `503 classifier_degraded` (verdict
still returned from rule baseline, `model_version: "rule-baseline"`,
`warnings:["ml_unavailable"]`).
