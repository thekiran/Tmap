# Architecture

This is the MVP slice of the larger "Internet Access Intelligence Platform". The
full target architecture (desktop UI, local API, SQLite, reports, update system,
ML) is described in the project brief; this document covers what is **built
today** and how it is meant to grow.

## Principle: evidence → score → confidence

The core idea is that no single signal reliably identifies an access type from
the client side. The engine therefore:

1. collects **independent pieces of evidence** (probes),
2. assigns **scores** to each candidate access type,
3. computes a **confidence** that reflects strength, agreement and corroboration,
4. and **explains** the verdict with the evidence behind it.

## Layers

```
┌──────────────────────────────────────────────────────────┐
│ CLI  (cmd/iad-agent)                                       │  prints summary,
│                                                            │  writes JSON
├──────────────────────────────────────────────────────────┤
│ Detection engine  (internal/detection, internal/scoring)  │  normalize, match,
│                                                            │  score, confidence,
│                                                            │  classify, explain
├──────────────────────────────────────────────────────────┤
│ Probe layer  (internal/probes)                            │  uniform ProbeResult
├──────────────────────────────────────────────────────────┤
│ OS/network helpers  (internal/system, internal/network)   │  cross-platform,
│                                                            │  build-tagged forks
├──────────────────────────────────────────────────────────┤
│ Shared contracts  (pkg/models)                            │  ProbeResult,
│                                                            │  ScanResult, types
└──────────────────────────────────────────────────────────┘
            ▲
            │  rules/  (YAML)  — access_rules, modem_fingerprints,
            │                    interface_patterns, isp_patterns
```

### Data contracts (`pkg/models`)

- `ProbeResult` — the single shape every probe emits (`probe_name`, `status`,
  `confidence`, `evidence`, `hints`, `errors`).
- `ScanResult` — the full output: verdict, category, confidence, per-type scores,
  alternatives, explanation, and the raw evidence.

Keeping these in one place means the future UI, local API and SQLite layer can
all consume the exact same JSON without redefining types.

### Probe layer (`internal/probes`)

Every probe implements:

```go
type Probe interface {
    Name() string
    Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error)
}
```

A `Runner` executes them concurrently with a per-probe timeout. A failing or
panicking probe is recorded as `failed` and never aborts the scan (graceful
degradation). `probes.Default(in)` selects the probe set from the scan mode and
online/offline flag.

### Detection engine (`internal/detection`, `internal/scoring`)

Pipeline:

1. **Normalizer** — tidy the discovered model string.
2. **Fingerprint matcher** — substring-match the model against
   `modem_fingerprints.yaml` (vendor, category, supported techs, access hints).
3. **Rule engine** — evaluate `access_rules.yaml`, `interface_patterns.yaml`,
   `isp_patterns.yaml`; each fired rule applies `add_score` deltas.
4. **Score engine** — accumulate points (incl. small banded latency), normalize.
5. **Classifier** — rank types; pick the leader + alternatives; map to a category.
6. **Confidence** — combine top *category* score, separation from the next
   category, corroboration (independent sources), and a fingerprint-match bonus.
7. **Decision layer** (`decision.go`) — downgrade the leader to `Unknown` (keeping
   the scores/alternatives) when confidence, top score or category margin is too
   low, or when no *strong physical evidence* exists; set `decision_quality` and
   `uncertainty_reasons`.
8. **Network context** (`context.go`) — assemble the factual situation (ISP,
   gateway chain, double-NAT, local access) reported regardless of the verdict.
9. **Explanation** — build the human-readable "why" (confident or uncertain).

### Cross-platform strategy

Pure-Go and standard-library code wherever possible. Genuinely OS-specific bits
are isolated behind build tags:

- `internal/network/dns_windows.go` (PowerShell, locale-independent) vs
  `dns_unix.go` (`/etc/resolv.conf`)
- `internal/system/traceroute_windows.go` (`tracert`) vs `traceroute_unix.go`
  (`traceroute`), with a shared parser in `traceroute_parse.go`

The default gateway uses `jackpal/gateway` (pure Go on all three OSes).

## Extending

- **New modem** → add an entry to `rules/modem_fingerprints.yaml`.
- **New ISP/tech pattern** → add a rule to `rules/access_rules.yaml` or
  `rules/isp_patterns.yaml`.
- **New evidence source** → implement `Probe`, add it to `probes.Default`. The
  engine and (future) UI need no changes because the `ProbeResult` shape is fixed.

## Planned (not yet built)

Tauri + React desktop UI, local HTTP API wrapping the runner, SQLite history,
SNMP probe, HTML/PDF reports, rule auto-update, ML-assisted classification.
