# 03 — Feature Engineering

> Brief §C and Deliverable #5 (detection feature table).

Features are grouped by the **evidence tier** the agent already uses
(`Physical` > `Device` > `Network` > `Performance`, plus `Regional`). The tier
matters as much as the value: **only Physical-tier evidence may push a verdict
out of `Unknown`** into a confident commit. Everything else narrows, calibrates,
and contextualizes. This mirrors `agent/internal/scoring/weights.go` and the
decision layer; the ML model gets all of it, but the ambiguity gate
(see [04](04-classification-strategy.md)) still requires Physical evidence to
commit.

## 1. Engineering principles

- **Tier every feature.** Each feature carries its evidence class so the model
  *and* the rule gate can treat "physical proof" differently from "weak hint".
- **Encode availability, not just value.** A missing feature is informative
  (a browser can't read DSL stats). Use explicit `*_present` flags + the
  collector capability manifest; never silently impute physical features.
- **Robust statistics over raw samples.** Use percentiles (p50/p90/p95/p99),
  IQR, MAD, and ratios — not means — because access links produce heavy-tailed,
  skewed latency/throughput.
- **Make ratios and shapes, not absolutes.** Throughput *asymmetry* and latency
  *distribution shape* generalize across countries and plans far better than
  absolute Mbps/ms (which are confounded by tier and distance).
- **Multi-vantage, not single-vantage.** RTT and path features are computed
  across ≥3 measurement servers; the *minimum* RTT and the *spread* are both
  features (min ≈ propagation/access floor; spread ≈ routing/congestion).
- **No PII in features.** Tokens from rDNS/ASN are hashed/bucketed categorical
  signals, not stored strings, in the feature vector (raw text stays in the
  evidence record for explanation only). See [10](10-security-ethics.md).

## 2. Deliverable #5 — Detection feature table

Legend — **Tier:** P=Physical, D=Device, N=Network, Perf=Performance,
R=Regional. **Modes:** R=remote, B=browser, A=desktop agent, M=mobile, C=CPE.
**Discriminates:** the access types this feature helps separate.

### 2.1 Throughput & capacity

| # | Feature | Tier | Modes | What it discriminates | Notes |
|---|---|---|---|---|---|
| 1 | Download throughput (p50, peak, sustained) | Perf | R B A M | tier/coarse separation | absolute, confounded — use as support |
| 2 | Upload throughput (p50, peak) | Perf | R B A M | DSL/Cable (asym) vs Fiber/Ethernet (sym) | |
| 3 | **Up/Down asymmetry ratio** = up/down | Perf | R B A M | **symmetric Fiber/Ethernet vs asymmetric DSL/Cable/FWA** | one of the strongest perf features |
| 4 | Throughput stability (CoV over repeated tests) | Perf | R A M | wireless/shared (Cable/WISP/Mobile) vs dedicated (Fiber/Ethernet) | needs time series |
| 5 | Capacity step/quantization (e.g. 100/300/1000 plan steps) | Perf | R A | provisioned-rate fingerprints (DSL sync caps, plan tiers) | |
| 6 | Time-of-day throughput variation (diurnal) | Perf | R A M | contended shared media (Cable/WISP/Mobile) vs dedicated | TSDB feature |

### 2.2 Latency & delay structure

| # | Feature | Tier | Modes | Discriminates | Notes |
|---|---|---|---|---|---|
| 7 | **Idle/min RTT** (per vantage, and global min) | Perf | R B A M | **GEO satellite (≥500ms), Mobile (40–100ms), Fiber/Cable (low), DSL (interleaving adds ms)** | min RTT ≈ access+propagation floor |
| 8 | Loaded latency (RTT under saturating load) | Perf | R B A M | bufferbloat-prone links (DSL/Cable/old CPE) | NDT7 / dual-flow |
| 9 | **Bufferbloat** = loaded − idle (and ratio) | Perf | R B A M | DSL/Cable/legacy vs well-managed fiber/AQM | classic discriminator |
| 10 | Latency percentiles p50/p90/p99 & IQR | Perf | R B A M | jitter regime per medium | |
| 11 | **Jitter / delay variation** (std, MAD) | Perf | R B A M | wireless (Wi-Fi/Mobile/WISP) vs wired | |
| 12 | Latency distribution *shape* (skew, kurtosis, bimodality) | Perf | R A M | scheduled-radio (LTE/NR HARQ, DOCSIS request-grant) signatures | bimodality hints at MAC scheduling |
| 13 | RTT vs distance residual (measured − great-circle floor) | Perf | R | satellite/long-haul vs local access | uses vantage geo |
| 14 | Latency-under-load recovery time | Perf | R A | AQM presence (fq_codel) vs dumb buffer | |

### 2.3 Loss, retransmission, reliability

| # | Feature | Tier | Modes | Discriminates | Notes |
|---|---|---|---|---|---|
| 15 | Packet loss rate (idle and loaded) | Perf | R B A M | wireless/satellite/contended vs wired | |
| 16 | Loss burstiness (gap model) | Perf | R A | radio fading (Mobile/WISP) vs congestion drop | |
| 17 | TCP retransmission rate | Perf | R A C | lossy/long-RTT links | from TCP_INFO/pcap |
| 18 | TCP_INFO: RTT, RTTVAR, cwnd, reorder, delivery rate | Perf | R A | path quality, BDP regime | where available |
| 19 | Reordering rate | Perf | R | multipath/bonded links (LTE+DSL, SD-WAN) | |

### 2.4 Packet-pair / packet-train (capacity & medium)

| # | Feature | Tier | Modes | Discriminates | Notes |
|---|---|---|---|---|---|
| 20 | **Packet-pair dispersion** → bottleneck capacity estimate | Perf | R A | corroborates provisioned rate; quantization → DSL/PON | needs UDP/raw timing |
| 21 | Packet-train inter-arrival distribution | Perf | R A | **slotted radio (LTE/NR/DOCSIS) shows quantized inter-arrivals** vs smooth wired | strong medium signature |
| 22 | Dispersion variance across trains | Perf | R A | shared/contended vs dedicated | |
| 23 | Serialization-delay signature (size-vs-delay slope) | Perf | R A | low-rate links (legacy DSL/T1) | |

### 2.5 Path / traceroute structure

| # | Feature | Tier | Modes | Discriminates | Notes |
|---|---|---|---|---|---|
| 24 | **First-hop RTT** | Perf/N | R A | Wi-Fi (variable) vs Ethernet (sub-ms); access floor | |
| 25 | **Second-hop RTT** (first ISP aggregation hop) | Perf/N | R A | DSLAM/CMTS/OLT/eNB proximity | |
| 26 | Hop count to first public IP | N | R A | CGNAT depth, double-NAT | |
| 27 | **Early-path CGNAT hop (100.64/10)** | N | R A | **Mobile/FWA/some WISP & cable** | medium signal |
| 28 | rDNS tokens on hops (`dsl`, `vdsl`, `ftth`, `cmts`, `bras`, `lte`, `gpon`) | N | R A | per-medium aggregation naming | tokenized/hashed |
| 29 | Hop RTT plateau pattern (where latency jumps) | N | R | access vs backbone boundary | |
| 30 | Paris-traceroute path multiplicity (ECMP/bonding) | N | R | bonded/SD-WAN/multipath | |
| 31 | IXP/anycast presence near edge | N | R | CDN-fronted vs direct | enrichment |

### 2.6 Addressing, NAT, routing identity

| # | Feature | Tier | Modes | Discriminates | Notes |
|---|---|---|---|---|---|
| 32 | **CGNAT detected** (100.64/10, STUN-vs-local mismatch) | N | R A M | Mobile/FWA/WISP lean | STUN/PCP from agent |
| 33 | Double-NAT topology (gateway chain length) | N | A | user router behind ISP CPE | from gateway-chain probe |
| 34 | IPv6 availability + prefix size (/64, /56, /48) | N | R A | ISP type & provisioning style | |
| 35 | DNS64/NAT64 presence | N | A M | mobile/CGNAT networks | |
| 36 | Public-IP stability across time | N | R A | static (Enterprise/Fiber) vs dynamic/CGNAT (Mobile) | |
| 37 | Reverse-DNS PTR full token set | N | R A | ISP + often medium | tokenized |

### 2.7 ASN / BGP / operator identity (Regional + Network)

| # | Feature | Tier | Modes | Discriminates | Notes |
|---|---|---|---|---|---|
| 38 | **ASN** (origin) + holder org | N/R | R A | operator → narrows plausible media via priors | |
| 39 | BGP prefix length & stability | N | R | enterprise PA/PI vs consumer pools | |
| 40 | ASN type (eyeball/transit/enterprise/mobile) | R | R | mobile-only ASN → Mobile prior | from ASN classification feeds |
| 41 | **Operator×region access priors** (P(type \| ASN, country)) | R | R A | strong regional prior, **never decisive alone** | learned from labels |
| 42 | rDNS/ASN technology keyword match | N | R A | direct medium tokens | from `isp_patterns.yaml` |
| 43 | Known-CGNAT / known-FWA / known-satellite ASN flags | R | R | Starlink/LTE-home/WISP operators | curated + learned |

### 2.8 Device / CPE identity (Device + Physical)

| # | Feature | Tier | Modes | Discriminates | Notes |
|---|---|---|---|---|---|
| 44 | **Modem/router fingerprint match** (model→supported techs) | D→P | A C | **strong; a DSL-only modem ⇒ DSL** | the workhorse; `modem_fingerprints.yaml` |
| 45 | UPnP/SSDP device type & manufacturer | D | A | vendor → medium lean | |
| 46 | HTTP banner / login realm / favicon hash / TLS CN | D | A C | CPE family identification | gateway-device probe |
| 47 | Gateway MAC OUI vendor | D | A | CPE vendor | |
| 48 | **WAN interface name** (`ptm0`/`atm0`/`gpon0`/`ont`/`lte`) | **P** | A C | **near-authoritative medium** | only if CPE surfaces it |

### 2.9 Physical-layer CPE stats (Physical — ground truth)

| # | Feature | Tier | Modes | Discriminates | Notes |
|---|---|---|---|---|---|
| 49 | **DSL line: attenuation, SNR margin, attainable/sync rate, interleaving, DSL mode** | **P** | C | **ADSL/ADSL2+/VDSL2/G.fast subtype** | authoritative when authorized |
| 50 | **DOCSIS: ds/us channels, OFDM/OFDMA, power, SNR/MER, modem mode** | **P** | C | **DOCSIS 2.0/3.0/3.1/4.0 subtype** | authoritative |
| 51 | **PON/ONT: GPON/EPON/XGS-PON markers, optical Rx/Tx power, ONT model** | **P** | C | **FTTH PON subtype** | authoritative |
| 52 | WAN access-type field (TR-064/IGD `WANAccessType`) | P | C | DSL/Ethernet/POTS class | |
| 53 | Ethernet link speed (10/100/1000/2500/10000) | D/P | A | Active-Ethernet/Fiber/Enterprise lean; rules out low-rate DSL | local NIC only ⇒ Device not Physical-of-WAN |

### 2.10 Wireless / cellular signal (Physical for the radio link)

| # | Feature | Tier | Modes | Discriminates | Notes |
|---|---|---|---|---|---|
| 54 | **Connection type: LTE / NR NSA / NR SA / 3G / 2G** | **P** | M | **authoritative for Mobile subtype** | TelephonyManager |
| 55 | RSRP, RSRQ, RSSI, RSSNR/SINR | P | M | signal quality; FWA-fixed vs mobile-moving stability | |
| 56 | Band / ARFCN / mmWave-vs-Sub6 | P | M | 5G mmWave vs Sub-6, FWA band plans | |
| 57 | Cell ID / handover events over time | N/P | M | **mobile (handovers) vs FWA (stationary, stable cell)** | permitted only; key FWA-vs-mobile signal |
| 58 | Signal stability over time (RSRP variance) | Perf | M | stationary FWA vs moving mobile | TSDB |
| 59 | **Wi-Fi: SSID/BSSID class, RSSI, frequency band (2.4/5/6GHz), channel width, PHY rate** | D/P | A | identifies *local* hop is Wi-Fi; hotspot/mesh/campus class | local link, not WAN |
| 60 | Local medium = Wi-Fi vs Ethernet vs Cellular | D | A M | separates the access-link question from the LAN hop | capability context |

### 2.11 Temporal / behavioral

| # | Feature | Tier | Modes | Discriminates | Notes |
|---|---|---|---|---|---|
| 61 | Speed stability over repeated runs | Perf | R A M | dedicated vs shared/wireless | |
| 62 | Diurnal latency/throughput pattern | Perf | R A | contended media | needs ≥1 day of samples |
| 63 | Session RTT drift / re-route frequency | N | R A | mobile/satellite handover vs stable wired | |
| 64 | LEO-satellite periodicity (~15s reconfig dips, latency 20–60ms) | Perf | R A | **LEO (Starlink) vs GEO vs terrestrial** | distinctive periodic signature |

### 2.12 Meta / quality features (about the measurement itself)

| # | Feature | Tier | Modes | Use |
|---|---|---|---|---|
| 65 | Collector capability manifest (which probes ran/were possible) | meta | all | weights evidence by what was achievable |
| 66 | Rule-baseline vs ML agreement | meta | cloud | disagreement → lower confidence, audit flag |
| 67 | Evidence-completeness score | meta | all | drives the Unknown gate and "next best probe" |
| 68 | Number of independent corroborating sources | meta | all | core input to confidence (already in MVP) |

## 3. Derived / composite features (built by the feature worker)

- **`asym_class`** = bucket(up/down ratio) ∈ {strong-asym, mild-asym, symmetric}.
- **`bufferbloat_grade`** = A–F from loaded−idle (TCP/CoDel-style scale).
- **`radio_quantization_score`** from packet-train inter-arrival regularity.
- **`access_floor_rtt`** = min over vantages of min RTT (propagation/access).
- **`cgnat_topology`** ∈ {none, single-NAT-public, cgnat, double-NAT, cgnat+double}.
- **`physical_evidence_present`** (bool) — gates the Unknown decision.
- **`mobile_stationarity`** from cell-ID change rate + RSRP variance (FWA vs mobile).
- **`operator_prior_vector`** = calibrated P(type | ASN, country) from labels.

## 4. Feature hygiene & leakage controls

- **No label leakage:** self-reported plan name is a *label/feedback* input, never
  a training feature (it would leak the answer). CPE-confirmed ground truth is a
  label, not a feature, during training; at *inference* CPE physical stats are
  legitimate features because they're measured, not declared.
- **Region balancing:** keep `country`/`region` as features but train with
  region-stratified sampling and report per-region metrics so the model can't win
  globally by memorizing one country's distribution (see
  [05](05-dataset-strategy.md) and [06](06-accuracy-expectations.md)).
- **Stable encodings:** ASN/token features are hashed to fixed buckets with a
  versioned vocabulary so a new ISP doesn't reshape the feature space mid-model.
- **Train/serve parity:** the same feature-builder code path (or a frozen spec)
  runs offline and online; the feature store stores exactly what serving sees.
