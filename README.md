# Internet Access Detector

Estimates a host's **internet access type** (ADSL / VDSL / Fiber / Cable / WISP /
Mobile-FWA / Satellite / Enterprise) from multiple pieces of evidence, combined
with a scoring + confidence model. It never decides from a single signal: it
collects evidence, scores each access type, computes a confidence, and **always
shows why**.

This repository currently contains the **MVP detection core**: a cross-platform
Go agent with a CLI. The desktop UI (Tauri + React), local API, SQLite history,
SNMP, and HTML/PDF reports are planned later phases — the probe and engine
contracts are designed so they slot in without rework.

## Why a score instead of a yes/no?

Behind a modem, VDSL, Fiber, Cable and WISP can look identical from the client
side. So the engine produces a ranked result with a confidence — and a
**decision layer** that refuses to commit when the evidence is weak, contested,
or lacks anything that physically proves the access type. The candidates are
still shown, e.g.:

```
Tahmin: Unknown  (kategori: Unknown)
Güven: %12 | Karar kalitesi: low
Operatör: TurkTelekom
Ağ geçidi zinciri: [192.168.31.1 192.168.1.1]  [çift NAT olası]
UPnP modem bulundu: false | Fingerprint eşleşti: false
Skorlar:
  Fiber      0.07
  VDSL       0.07
  DSL        0.04
Neden kesin karar verilmedi:
  - Fiziksel erişim türünü gösteren güçlü kanıt yok
  - UPnP modem modeli bulunamadı
  - PTR kaydı erişim türü için kesin kanıt değil
```

`Unknown` with visible candidates is a feature: it tells you the evidence was
weak, instead of pretending to be certain. When a modem *is* fingerprinted (or
a WAN interface / DSL-GPON-DOCSIS marker is seen), the engine commits with
`decision_quality: medium`/`high`.

## Quick start

Requires Go 1.26+.

```sh
cd agent

# fast LAN-side scan
go run ./cmd/iad-agent --mode quick --out ../report.json

# deeper scan: adds traceroute + reverse-DNS/ASN analysis
go run ./cmd/iad-agent --mode deep  --out ../report-deep.json

# privacy mode: no probe contacts any external service
go run ./cmd/iad-agent --mode quick --offline
```

Build a standalone binary:

```sh
cd agent
go build -o iad-agent ./cmd/iad-agent      # add .exe on Windows
```

### Flags

| Flag        | Default   | Meaning                                                        |
|-------------|-----------|---------------------------------------------------------------|
| `--mode`    | `quick`   | `quick` (LAN-side, fast) or `deep` (adds traceroute + ASN)    |
| `--online`  | `true`    | Allow probes that contact external services                   |
| `--offline` | `false`   | Disable all online probes (overrides `--online`)              |
| `--rules`   | auto      | Directory with the YAML rule/fingerprint files                |
| `--out`     | (none)    | Write the full JSON report to this path                       |

## How it works

```
probes ──► evidence ──► detection engine ──► verdict + confidence + explanation
```

1. **Probes** each emit the same `ProbeResult` shape: adapters, gateway, DNS,
   latency/jitter, UPnP modem model, public IP, traceroute, reverse-DNS/ASN.
2. **Detection engine** normalizes the evidence, matches the modem against a
   fingerprint database, applies YAML scoring rules, ranks the access types,
   computes a confidence, and builds a human-readable explanation.

Rules and fingerprints live in [`rules/`](rules/) as YAML, so new modem models
or ISP patterns can be added **without recompiling**.

See [docs/architecture.md](docs/architecture.md) and
[docs/detection-methods.md](docs/detection-methods.md) for details (including an
honest account of what can and cannot be detected from a client OS).

## Layout

```
agent/            Go module (CLI + probes + detection engine)
  cmd/iad-agent/  CLI entrypoint
  internal/       system, network, probes, detection, scoring, report
  pkg/models/     shared data contracts (ProbeResult, ScanResult, access types)
rules/            YAML rules + modem/ISP fingerprint databases
tests/fixtures/   canned probe outputs for deterministic engine tests
docs/             architecture + detection methods
```

## Testing

Deterministic, no network required, identical on every OS:

```sh
cd agent
go test ./...
```

## Scope & security

The agent only inspects the **user's own network** using passive or standard
diagnostic checks (ping, traceroute, DNS, public UPnP/SSDP info). It does not
attempt logins, brute force, neighbor-network scanning, or any intrusive probing.
Probes that reach an external service (public IP, ASN, traceroute) are gated
behind `--online` and never run under `--offline`.
