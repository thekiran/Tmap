# 10 — Security & Ethics

> Brief §J and Deliverable #9 (production hardening checklist). These rules are
> not advisory — they are enforced in code, schema, and process. The MVP already
> embodies the core stance (own-network only, no logins, `--offline` kill switch).

## 1. Non-negotiable principles

1. **Consent before collection.** No measurement without explicit, scoped,
   revocable consent; the scope travels with every record and is enforced at
   ingest *and* at training-set assembly.
2. **Own network only.** The system inspects only the user's own connection. No
   scanning of unrelated/third-party networks, no neighbor discovery beyond the
   user's own LAN gateway chain.
3. **No unauthorized access, ever.** No router-panel logins by guessing, no
   brute forcing, no default-credential attempts, no exploitation. CPE telemetry
   requires the user to authorize it (and often provide credentials *themselves*);
   absence of authorization means absence of that probe — not a workaround.
4. **No third-party probing of operator infrastructure.** Traceroute and pings go
   to the user's own path / public measurement servers, rate-limited.
5. **Minimize, anonymize, and forget.** Collect the least that classifies; strip
   PII at the source; short retention for raw artifacts; honor erasure.
6. **Honesty as ethics.** Never display false certainty. `Unknown` and calibrated
   confidence are an ethical commitment, not just a UX choice.

## 2. Consent model

- **Granular scopes:** `classify` (compute a verdict, ephemeral), `store`
  (persist the measurement), `research` (may enter the training corpus),
  `allow_cpe_read` (authorize SNMP/TR-064/vendor-API physical reads). Each is
  independently grantable and revocable.
- **Transparency:** before any probe runs, the UI explains in plain language what
  is collected, why, where it goes, and how to revoke. A machine-readable
  `policy_version` is recorded.
- **Default-deny for sensitive probes:** packet capture (Npcap/libpcap), CPE
  reads, and cell-ID collection are off unless explicitly enabled per session.
- **Revocation = deletion:** revoking `research`/`store` triggers crypto-shred and
  removal from the next dataset snapshot.

## 3. Data protection

- **No payloads.** Measurements are metadata/statistics only.
- **IP minimization:** enrich (ASN/geo/rDNS) at the edge, then store the public IP
  **truncated** + a salted hash (for dedup/linkage), never the full host address
  long-term.
- **Token hashing:** rDNS/ASN tokens are hashed/bucketed in features; raw text
  kept only transiently for the explanation, then dropped/anonymized.
- **Encryption:** TLS in transit (mTLS for the fleet); encryption at rest;
  per-user crypto key enabling shred-on-erasure.
- **k-anonymity** on every aggregate/exported dataset; suppress small cells.
- **Regional residency:** raw data stays in-region; only de-identified aggregates
  cross borders. Per-region DPIA.

## 4. Probe safety (technical guardrails)

- **Rate limiting** at three layers: collector (local token bucket per target),
  gateway (per-key/global), and enrichment (respect third-party RDAP/BGP limits).
- **Scope clamps in code:** traceroute/ping targets restricted to the user's path
  or an allow-listed measurement-server set; LAN discovery restricted to the
  user's own gateway chain (the agent already does this).
- **Graceful degradation:** a denied/failed probe is recorded and skipped; it
  never escalates or retries aggressively.
- **Offline kill switch:** `--offline` (exists) guarantees no probe contacts any
  external service; the cloud has an equivalent per-tenant "local-only" mode.

## 5. Deliverable #9 — Production hardening checklist

**Application / API**
- [ ] All inputs schema-validated (`/v1`); request size caps; reject unknown
      consent/policy versions.
- [ ] Idempotency enforced via `measurement_id`; replay-safe ingest.
- [ ] AuthN: API keys + OAuth device flow + mTLS fleet; short-lived tokens.
- [ ] AuthZ: per-key scopes/quotas; tenant isolation; consent re-checked server-side.
- [ ] Rate limits + abuse detection (anomalous submission volume per key/ASN).
- [ ] Output never includes raw PII; predictions reference a `model_version`.

**Data**
- [ ] PII minimization verified (IP truncation/hash) by automated test.
- [ ] Encryption at rest + in transit; key rotation; per-user crypto-shred path
      tested end-to-end (erase → excluded from next snapshot).
- [ ] Retention TTLs enforced on raw blobs/pcap; lifecycle policies on object store.
- [ ] k-anonymity gate on aggregate endpoints (unit-tested with small cells).
- [ ] Backups encrypted; restore tested; PITR for Postgres.

**Infra / Ops**
- [ ] IaC (Terraform) reviewed; least-privilege IAM; secrets in a vault, never in
      env files or images.
- [ ] Network policy: workers can reach only required services; egress allow-list
      for enrichment.
- [ ] Container image scanning + SBOM; dependency pinning; `govulncheck`/`pip-audit`
      in CI.
- [ ] Structured logs scrubbed of PII; audit log for consent changes & CPE reads.
- [ ] Observability: ingest lag, classify latency, Unknown-rate, per-region
      accuracy, error budgets + alerts.
- [ ] DR runbook; multi-AZ; graceful ML-serving-down → rule-baseline fallback
      verified.

**ML / Safety**
- [ ] Promotion gates ([06 §5](06-accuracy-expectations.md)) enforced in CI before
      any model goes live.
- [ ] Calibration (ECE) monitored in prod; drift alerts trigger retrain.
- [ ] OOD detector forces `Unknown` on out-of-distribution inputs (tested).
- [ ] Model cards + dataset datasheets published; reproducible snapshots.

**Process / Legal**
- [ ] Privacy policy + DPIA per region; lawful basis documented.
- [ ] Coordinated vulnerability disclosure policy; security contact.
- [ ] Open-source license + contribution guidelines reaffirming the
      "no unauthorized access / own-network-only" boundary.

## 6. Abuse & misuse resistance

- The product must not become a network-recon tool. Guardrails: own-network
  scope clamps, no arbitrary-target scanning surface in the API, rate limits, and
  refusing to operate against targets the user can't demonstrate they control.
- Aggregate endpoints expose distributions, never individuals; small cells
  suppressed; no reverse-lookup of a specific household's technology.
