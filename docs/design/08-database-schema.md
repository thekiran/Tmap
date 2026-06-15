# 08 — Database Schema

> Brief §H and Deliverable #3 (data model). PostgreSQL 16 + TimescaleDB. Column
> types and JSONB blobs mirror `pkg/models` so ingest is a near-direct map from
> the API bundle.

## Design rules

- **Facts are append-only.** `measurements` and its children are immutable once
  written; corrections are new rows / `feedback_labels`, never updates.
- **Reproducibility:** every `predictions` row pins the `model_versions.id` and
  the feature-schema hash that produced it.
- **Privacy by construction:** no raw payloads; public IP stored truncated +
  hashed; consent scope on every measurement; crypto-shred by `user_id`.
- **Time-series go to hypertables** (`signal_samples`, repeated `speed_tests`).
- **JSONB for the long tail** (full probe evidence, raw traceroute) with promoted
  columns for anything queried/joined/feature-built.

## DDL

```sql
-- Enums -------------------------------------------------------------------
CREATE TYPE collector_mode AS ENUM ('remote','browser','agent','mobile','cpe');
CREATE TYPE probe_status   AS ENUM ('success','failed','skipped');
CREATE TYPE evidence_tier  AS ENUM ('Physical','Device','Network','Performance','Regional');
CREATE TYPE label_source   AS ENUM ('self_report','isp_plan','cpe_confirmed','radio_confirmed','fingerprint','synthetic');
CREATE TYPE decision_quality AS ENUM ('low','medium','high');

-- 1. users ----------------------------------------------------------------
-- Pseudonymous account. PII minimized; crypto-shred deletes the key, voiding
-- all linkage, satisfying erasure without rewriting fact tables.
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_ref    TEXT UNIQUE,                 -- OAuth subject hash, nullable for anon
    consent_scopes  TEXT[] NOT NULL DEFAULT '{}',-- classify | store | research
    consent_policy_version TEXT,
    region          TEXT,                         -- coarse, for residency/compliance
    crypto_key_id   UUID,                         -- per-user key; drop to shred
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ                   -- soft delete
);

-- 2. devices --------------------------------------------------------------
CREATE TABLE devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    kind            collector_mode NOT NULL,      -- which collector family
    collector       TEXT,                         -- "iad-agent", "iad-android", ...
    collector_version TEXT,
    os              TEXT,
    capability_manifest JSONB NOT NULL DEFAULT '{}', -- what this device CAN probe
    device_token_hash TEXT,                        -- auth, hashed
    first_seen      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen       TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX ON devices (user_id);

-- 3. measurements ---------------------------------------------------------
-- One scan bundle. The factual network context lives here (mirrors
-- models.NetworkContext); detailed evidence is in child tables.
CREATE TABLE measurements (
    id              UUID PRIMARY KEY,             -- == client measurement_id (idempotent)
    device_id       UUID REFERENCES devices(id) ON DELETE CASCADE,
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    mode            collector_mode NOT NULL,
    schema_version  TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    consent_scopes  TEXT[] NOT NULL,
    consent_allow_cpe BOOLEAN NOT NULL DEFAULT false,
    -- network context (promoted for joins/features) --------------------
    isp             TEXT,
    country         TEXT,
    region          TEXT,
    asn             TEXT,
    bgp_org         TEXT,
    public_ip_trunc INET,                          -- host octets dropped
    public_ip_hash  TEXT,                          -- salted hash for dedup/linkage
    ptr             TEXT,
    cgnat           BOOLEAN,
    double_nat_possible BOOLEAN,
    gateway_chain   TEXT[],
    local_access    TEXT,                          -- ethernet | wifi | cellular
    link_speed_mbps INT,
    router_model    TEXT,
    fingerprint_matched BOOLEAN,
    network_context JSONB NOT NULL DEFAULT '{}',   -- full models.NetworkContext
    raw_blob_ref    TEXT,                          -- object-store pointer (optional)
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX ON measurements (asn, country);
CREATE INDEX ON measurements (device_id, created_at DESC);
CREATE INDEX ON measurements USING gin (network_context);

-- 4. probes ---------------------------------------------------------------
-- One row per ProbeResult. Uniform shape; evidence as JSONB.
CREATE TABLE probes (
    id              BIGSERIAL PRIMARY KEY,
    measurement_id  UUID NOT NULL REFERENCES measurements(id) ON DELETE CASCADE,
    probe_name      TEXT NOT NULL,
    status          probe_status NOT NULL,
    confidence      REAL,
    evidence_tier   evidence_tier,                 -- strongest tier this probe yields
    evidence        JSONB NOT NULL DEFAULT '{}',
    hints           TEXT[],
    errors          TEXT[]
);
CREATE INDEX ON probes (measurement_id);
CREATE INDEX ON probes (probe_name);

-- 5. traceroutes ----------------------------------------------------------
CREATE TABLE traceroutes (
    id              BIGSERIAL PRIMARY KEY,
    measurement_id  UUID NOT NULL REFERENCES measurements(id) ON DELETE CASCADE,
    vantage         TEXT,                          -- which measurement server / probe
    method          TEXT,                          -- paris | classic | mtr
    target          TEXT,
    hop_count       INT,
    first_hop_rtt_ms  REAL,
    second_hop_rtt_ms REAL,
    cgnat_hop_index INT,                            -- -1 if none
    path_tokens     TEXT[],                         -- dsl/vdsl/cmts/bras/gpon/lte found in rDNS
    hops            JSONB NOT NULL,                 -- [{idx, ip_trunc, rtt_ms[], ptr_tokens}]
    raw_blob_ref    TEXT
);
CREATE INDEX ON traceroutes (measurement_id);

-- 6. speed_tests ----------------------------------------------------------
-- A single throughput/latency-under-load run. Repeated runs over time also
-- feed the hypertable (7) for stability/diurnal features.
CREATE TABLE speed_tests (
    id              BIGSERIAL PRIMARY KEY,
    measurement_id  UUID NOT NULL REFERENCES measurements(id) ON DELETE CASCADE,
    method          TEXT,                          -- ndt7 | packet_pair | http
    server          TEXT,
    download_mbps   REAL,
    upload_mbps     REAL,
    asymmetry_ratio REAL,                           -- up/down
    idle_latency_ms REAL,
    loaded_latency_ms REAL,
    bufferbloat_ms  REAL,                           -- loaded - idle
    jitter_ms       REAL,
    packet_loss_pct REAL,
    retransmit_pct  REAL,
    tcp_info        JSONB,                          -- rtt, rttvar, cwnd, delivery_rate...
    pktpair_capacity_mbps REAL,
    pkttrain_quantization REAL,
    measured_at     TIMESTAMPTZ NOT NULL
);
CREATE INDEX ON speed_tests (measurement_id);

-- 7. signal_samples (TIME-SERIES) -----------------------------------------
-- Wi-Fi / cellular signal sampled over time. TimescaleDB hypertable: range
-- queries for stability, variance, handover-rate, diurnal features are cheap.
CREATE TABLE signal_samples (
    measurement_id  UUID NOT NULL REFERENCES measurements(id) ON DELETE CASCADE,
    device_id       UUID REFERENCES devices(id) ON DELETE CASCADE,
    ts              TIMESTAMPTZ NOT NULL,
    medium          TEXT NOT NULL,                  -- wifi | cellular
    -- cellular
    radio_type      TEXT,                           -- LTE | NR_NSA | NR_SA | UMTS | GSM
    rsrp_dbm        REAL, rsrq_db REAL, rssi_dbm REAL, sinr_db REAL,
    band            TEXT, arfcn INT, cell_id TEXT,  -- cell_id only if permitted
    -- wifi
    ssid_class      TEXT,                           -- home|hotspot|campus|mesh (no raw SSID)
    bssid_oui       TEXT,
    freq_band       TEXT,                           -- 2.4 | 5 | 6 GHz
    channel_width_mhz INT, phy_rate_mbps REAL,
    PRIMARY KEY (measurement_id, ts, medium)
);
SELECT create_hypertable('signal_samples','ts', if_not_exists => TRUE);
CREATE INDEX ON signal_samples (device_id, ts DESC);

-- 8. cpe_stats ------------------------------------------------------------
-- Authorized physical-layer ground truth (the gold mode). Source proves auth.
CREATE TABLE cpe_stats (
    id              BIGSERIAL PRIMARY KEY,
    measurement_id  UUID NOT NULL REFERENCES measurements(id) ON DELETE CASCADE,
    source          TEXT NOT NULL,                  -- snmp | tr064 | upnp_igd | vendor_api
    authorized      BOOLEAN NOT NULL,               -- consent.allow_cpe_read was true
    wan_access_type TEXT,                            -- DSL | Ethernet | GPON | ...
    wan_iface_name  TEXT,                            -- ptm0 | atm0 | gpon0 | ont | lte
    -- DSL
    dsl_mode        TEXT,                            -- ADSL2+ | VDSL2 | G.fast
    dsl_attn_db     REAL, dsl_snr_margin_db REAL,
    dsl_attainable_kbps INT, dsl_sync_down_kbps INT, dsl_sync_up_kbps INT,
    dsl_interleaving TEXT,
    -- DOCSIS
    docsis_version  TEXT,                            -- 2.0 | 3.0 | 3.1 | 4.0
    docsis_ds_channels INT, docsis_us_channels INT,
    docsis_ofdm BOOLEAN, docsis_ofdma BOOLEAN,
    docsis_power_dbmv REAL, docsis_snr_mer_db REAL,
    -- PON
    pon_type        TEXT,                            -- GPON | EPON | XGS-PON | 10G-EPON
    ont_model       TEXT, optical_rx_dbm REAL, optical_tx_dbm REAL,
    raw_stats       JSONB NOT NULL DEFAULT '{}',
    collected_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX ON cpe_stats (measurement_id);

-- 9. predictions ----------------------------------------------------------
-- A verdict. Immutable; a re-classification (new model) writes a NEW row.
CREATE TABLE predictions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    measurement_id  UUID NOT NULL REFERENCES measurements(id) ON DELETE CASCADE,
    model_version_id UUID NOT NULL REFERENCES model_versions(id),
    rule_baseline_version TEXT,
    feature_schema_hash TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    primary_type    TEXT NOT NULL,                  -- may be 'Unknown'
    category        TEXT NOT NULL,
    subtype         TEXT,
    classification_confidence REAL NOT NULL,
    context_confidence REAL,
    decision_quality decision_quality NOT NULL,
    is_unknown      BOOLEAN GENERATED ALWAYS AS (category = 'Unknown') STORED,
    distribution    JSONB NOT NULL,                 -- calibrated [{category,type,subtype,prob}]
    candidates      JSONB NOT NULL,                 -- AccessCandidate[]
    confidence_breakdown JSONB NOT NULL,            -- ConfidenceBreakdown
    evidence_strength JSONB,                        -- EvidenceStrengthSummary
    score_contributions JSONB NOT NULL,             -- ScoreContribution[]
    uncertainty_reasons TEXT[],
    next_best_probes JSONB,
    shap_top_features JSONB,
    explanation     TEXT[]
);
CREATE INDEX ON predictions (measurement_id, created_at DESC);
CREATE INDEX ON predictions (model_version_id);
CREATE INDEX ON predictions (category) WHERE is_unknown = false;

-- 10. model_versions ------------------------------------------------------
CREATE TABLE model_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT UNIQUE NOT NULL,           -- "iad-clf-2026.06.1"
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    algo            TEXT,                            -- lightgbm-ensemble | rule-baseline
    code_commit     TEXT,
    dataset_snapshot_id TEXT NOT NULL,              -- immutable training snapshot
    feature_schema_hash TEXT NOT NULL,
    calibration     JSONB,                          -- per (mode x region) calibrators
    metrics         JSONB NOT NULL,                 -- per-region/class acc, ECE, coverage
    model_card_ref  TEXT,                            -- object-store doc
    onnx_ref        TEXT,
    status          TEXT NOT NULL DEFAULT 'shadow', -- shadow | canary | active | retired
    promoted_at     TIMESTAMPTZ
);

-- 11. feedback_labels -----------------------------------------------------
-- The training spine. Raw declared label kept verbatim + normalized canonical.
CREATE TABLE feedback_labels (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    measurement_id  UUID REFERENCES measurements(id) ON DELETE SET NULL,
    prediction_id   UUID REFERENCES predictions(id) ON DELETE SET NULL,
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    source          label_source NOT NULL,
    declared_label  TEXT,                            -- raw, never overwritten
    norm_category   TEXT, norm_type TEXT, norm_subtype TEXT,
    trust_weight    REAL NOT NULL DEFAULT 0.5,       -- by source; CPE/radio ~1.0
    consent_scopes  TEXT[] NOT NULL,                 -- must include 'research' to train
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    notes           TEXT
);
CREATE INDEX ON feedback_labels (measurement_id);
CREATE INDEX ON feedback_labels (source, norm_category);
```

## Notes for implementers

- **Training-set assembly** is a query: `feedback_labels` JOIN `measurements`
  JOIN child evidence WHERE `'research' = ANY(consent_scopes)` AND not deleted,
  snapshotted immutably and referenced by `model_versions.dataset_snapshot_id`.
- **Re-classification** of history (new model) inserts new `predictions` rows;
  old ones stay for audit and A/B comparison. `is_unknown` is a generated column
  so the operator dashboard can chart Unknown-rate cheaply.
- **At ClickHouse scale**, `signal_samples` and high-volume `speed_tests` move to
  ClickHouse (same columns), Postgres keeps entities + predictions + labels.
- **Retention/erasure:** raw blobs and pcap have short TTLs in the object store;
  `deleted_at` + dropping `users.crypto_key_id` shreds linkage; the next dataset
  snapshot excludes shredded rows.
