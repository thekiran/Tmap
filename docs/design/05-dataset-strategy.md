# 05 — Dataset Strategy

> Brief §E. How to build a *labeled, global, privacy-respecting* dataset — the
> single hardest and most valuable part of the project. The model is only as
> honest as its labels.

## 1. The core problem

Ground truth for "what access technology is this" is **expensive and unevenly
available**. The strategy is a **label trust hierarchy**: collect many weak
labels cheaply, anchor them with few strong labels, and let the trust weight
flow into training. Never treat a self-reported plan name as equal to a DSL SNR
reading.

## 2. Label sources, by trust (highest → lowest)

| Source | Trust | How obtained | Caveats |
|---|---|---|---|
| **CPE-confirmed physical stats** | ★★★★★ | Authorized SNMP/TR-064/PON optical/DOCSIS readings give the medium directly | Sparse; only consenting users with readable CPE; the gold anchor set |
| **WAN interface name** (`ptm0`/`gpon0`/`lte`) | ★★★★★ | Surfaced via UPnP/IGD/TR-064 | Rare on consumer CPE |
| **Mobile radio type** (LTE/NR SA) | ★★★★★ | TelephonyManager on the mobile app | Authoritative for the mobile branch only |
| **ISP known-plan labels** | ★★★★ | Partner ISPs / verified plan import (e.g. "this account is FTTH 1Gbps") | Plan name ≠ delivered medium (FTTC sold as "fiber"); needs normalization |
| **Modem fingerprint (exclusive medium)** | ★★★★ | A DSL-only / DOCSIS-only modem model implies the medium | Multi-WAN/gateway models are ambiguous |
| **Self-reported labels** | ★★ | User picks "I have fiber/DSL/cable/5G home" in the UI | Users misremember/conflate; marketing terms; noisy but plentiful |
| **Synthetic lab measurements** | ★★★ (controlled) | Real links in a lab (ADSL2+, VDSL2, DOCSIS3.1, GPON, LTE/5G FWA, Starlink) under known conditions | Limited diversity; great for medium *signatures* and augmentation |
| **Public aggregate datasets** | ★★ (aggregate) | M-Lab (NDT) throughput/latency at scale; CAIDA; reverse-DNS corpora | Usually no per-host medium label → used for priors, distributions, weak supervision |
| **Active-probe networks** | ★★★ (path) | RIPE Atlas-style probes; own global fleet | Great for path/RTT; medium label only where probe host is known |

### Label normalization
A single **canonical label taxonomy = the `pkg/models` access types**. Every
source maps into it via a documented crosswalk (e.g. ISP "Fibernet" → resolve to
FTTH vs FTTC using line stats if available, else mark `Fiber` family only). Store
the **raw declared label + the normalized label + the source + trust weight**;
never overwrite the raw.

## 3. Weak supervision & label fusion

- Treat each source as a **labeling function** (Snorkel-style). A generative
  label model estimates per-source accuracy and produces **probabilistic
  training labels** with confidences — which the GBDT can train against directly.
- **Anchor with the gold set:** CPE/radio-confirmed rows calibrate the label
  model and are the *only* source allowed to define subtype ground truth.
- **Agreement boosts trust:** self-report "fiber" + symmetric throughput +
  GPON-token rDNS + low bufferbloat → high-trust Fiber label even without CPE.
- **Conflicts are data:** self-report "fiber" + DSL line stats from CPE →
  trust the CPE, flag the self-report error rate (it's a feature of that ISP's
  marketing). Track per-ISP self-report reliability.

## 4. Building the global, balanced corpus

The enemy is **regional and technology bias**: most volunteers cluster in a few
countries and on a few popular media, so a naive model "wins" by predicting the
local majority.

- **Stratified targets:** define minimum sample counts per
  (region × access-family × mode) cell; track fill rate on the dashboard.
- **Region-stratified training + group-aware splits:** split by network/household
  (group = ASN+prefix or device cluster) so rows from the same connection never
  straddle train/test (prevents leakage and inflated accuracy).
- **Regional bias correction:**
  - *Reweighting / importance sampling* to a target global population distribution
    (or report metrics per region and optimize macro-averaged, not micro).
  - *Per-region calibration* (see [04](04-classification-strategy.md)) so a
    confidence means the same thing in Turkey, Brazil, and Germany.
  - *Operator×region priors learned, capped:* priors narrow but never decide, so
    a model deployed in a new country degrades gracefully to `Unknown` rather than
    confidently wrong.
- **Active learning:** prioritize labeling requests where the model is *uncertain*
  or *OOD* (new ASN, new region, disagreement) — buys the most accuracy per label.
- **Seed regions then expand:** the MVP already has Turkish ISP starter rules
  (`isp_patterns.yaml`); the same pattern (rules + labels + priors as *data*)
  onboards each new country without code changes.

## 5. Synthetic & augmentation

- **Lab captures** of each medium under varied conditions (distance, congestion,
  weather for wireless/satellite) → ground-truth signatures for packet-train
  quantization, bufferbloat, LEO periodicity.
- **Augmentation:** jitter/scale performance features within physically plausible
  bounds; **modality dropout** (hide CPE/mobile features) to simulate the common
  remote-only case; synthesize `Unknown`/ambiguous samples by blending media so
  the abstain class is genuinely trained, not just a leftover.
- **Never synthesize Physical-tier "proof"** — synthetic data augments
  performance/path distributions, not the ground-truth gate.

## 6. Privacy & consent (dataset-specific; see also [10](10-security-ethics.md))

- **Consent is explicit, scoped, and revocable**, and the consent scope (e.g.
  "share for research", "allow CPE read") is stored with every measurement and
  enforced at training-set assembly (rows without research consent are excluded).
- **Minimize and anonymize at rest:** public IP is truncated/hashed for storage
  (keep ASN/prefix/geo, drop the host octets after enrichment); rDNS/ASN tokens
  are hashed in features; no payloads, ever. pcap (opt-in) is short-retention and
  never enters the shared corpus.
- **k-anonymity for aggregates & exports:** any released/aggregate dataset
  enforces a minimum group size per (ASN×region×type) cell; suppress small cells.
- **Right to erasure:** crypto-shred by `user_id`/`device_id`; labels derived from
  erased data are removed from the next training snapshot (snapshots are
  versioned and rebuildable).
- **Geographic compliance:** regional data residency for raw data; only
  de-identified aggregates cross borders. Document a data-handling DPIA per region.

## 7. Dataset versioning & governance

- **Immutable training snapshots** (snapshot id pinned in every `model_versions`
  row) so any model is reproducible and any erasure is auditable.
- **Datasheet for the dataset:** composition by region/family/source, known
  biases, consent basis, label-trust methodology — shipped publicly for an
  open-source product.
- **Continuous labels:** `feedback_labels` from the live product (confirm/correct
  in the UI, ISP plan imports, new CPE reads) flow into each snapshot, so the
  corpus and the model improve together. Active-learning requests target the
  cells the dashboard shows as weakest.

## 8. Minimum viable dataset (to start training)

- ~**A few thousand** gold-anchored rows (CPE/radio-confirmed) spread across the
  core families, plus tens of thousands of weak-labeled rows, across **≥5
  regions** and the **6 core families** (DSL, Fiber, Cable, FWA, Mobile,
  Enterprise) — enough to beat the rule baseline on macro-F1 and to calibrate.
- Below that, **ship the rule baseline only** and harvest labels; don't pretend a
  data-starved model is better than honest rules. (See the roadmap,
  [09](09-roadmap.md).)
