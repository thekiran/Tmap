# Detection methods — and their limits

This document is deliberately honest about what each probe can and cannot tell
you. Overstating certainty would defeat the purpose of a confidence-based tool.

## The big caveat: you are behind the modem

From a **client OS** you usually cannot see the things that would prove an access
type outright:

- You **cannot** read the modem's WAN interface name (`ptm0` = VDSL, `atm0` =
  ADSL, `gpon0` = fiber). Those exist *inside* the modem. The
  `interface_patterns.yaml` rules only fire when such a name is surfaced via
  UPnP/SNMP/HTTP — which most consumer modems don't expose by default.
- You **cannot** read DSL line stats, optical power, or DOCSIS channel info
  without authenticating to the modem (which this tool does not do).
- **Double NAT** (a router behind the ISP modem) hides the modem one hop further;
  SSDP/UPnP won't reach it.

So the dominant client-side signals are the **modem model** (via UPnP) and the
**ISP identity** (via reverse-DNS / RDAP). When neither is conclusive, the honest
answer is a low-confidence lean or "Unknown".

## Probes

| Probe              | Mode        | What it observes                              | Strength |
|--------------------|-------------|-----------------------------------------------|----------|
| `adapter_probe`    | quick       | Network adapters; a cellular adapter → Mobile | weak     |
| `gateway_probe`    | quick       | Default gateway IP (context, reused by others)| weak     |
| `dns_probe`        | quick       | Configured DNS servers                         | weak     |
| `latency_probe`    | quick       | Avg latency + jitter (ICMP, TCP fallback)     | weak\*   |
| `upnp_probe`       | quick       | Modem manufacturer/model via SSDP             | strong   |
| `public_ip_probe`  | quick+online| Public IP; CGNAT (100.64/10) → Mobile/FWA     | medium   |
| `traceroute_probe` | deep+online | Path hops; early CGNAT hop → Mobile/FWA/WISP  | weak     |
| `asn_probe`        | deep+online | Reverse-DNS (PTR) + RDAP org; tech keywords   | medium   |

\* Latency is mostly corroborating. The one fairly distinctive case is very high
RTT (> 400 ms), which suggests geostationary satellite.

### `upnp_probe` (the workhorse)

Sends an SSDP `M-SEARCH` on the LAN, fetches the UPnP device description from the
`LOCATION` of any responder, and reads `manufacturer` / `modelName` /
`modelNumber` / `deviceType`. The model is what the fingerprint database and the
`router_model_contains` rules key on. LAN-only — runs even offline.

### `latency_probe` (privilege-aware)

Tries ICMP first (`pro-bing`). Raw ICMP needs elevation on Windows, so on failure
it falls back to timing TCP connects — meaning an unprivileged user still gets a
measurement. Online: targets a public resolver. Offline: targets the LAN gateway,
so no traffic leaves the local network.

### `asn_probe` (ISP fingerprinting)

Reverse-resolves the public IP and fetches its RDAP record. ISP PTR hostnames
frequently embed the technology (`...vdsl...`, `lte-...`, `ftth-...`). Brand names
(e.g. TTNet, Superonline) narrow the field even when the exact tech isn't encoded
— see the Turkish ISP starter rules in `isp_patterns.yaml`. Outbound calls, so
deep+online only.

## How evidence becomes a verdict

Each probe contributes `hints` and/or identifying `evidence`. The engine:

1. matches the model against `modem_fingerprints.yaml` (strongest signal),
2. applies the YAML rules (model / text / interface / hint conditions),
3. adds probe hints and the small banded latency contribution,
4. normalizes scores and ranks them,
5. computes confidence from top-**category** strength, separation from the next
   category, corroborating sources, and whether a known modem was fingerprinted.

A verdict with one weak source scores a low confidence on purpose.

## The decision layer (Unknown vs. committed)

Ranking by score is not enough — a top score of 0.07 should not become a
confident "DSL". On top of the scores the engine runs a **decision layer**
(`internal/detection/decision.go`) that downgrades the verdict to **`Unknown`**
(while keeping all scores and `alternatives`) when *any* of these hold:

- `confidence < 0.45`
- top **category** score `< 0.35`
- the gap between the top two **categories** `< 0.12` (too close to call)
- there is **no strong physical evidence** of the access type

"Strong physical evidence" means: a modem fingerprint match, a WAN interface name
(`ptm0`/`atm0`/`gpon`/`ont`/`lte`/…), DSL line stats, or ONT/GPON/DOCSIS markers.
PTR, ASN, public IP, latency and local Ethernet/Wi-Fi names are **not** strong —
they narrow the field and provide context, but never decide on their own.

Crucially, the margin is computed at the **category** level: DSL and VDSL being
close is *not* ambiguity (it's still confidently DSL); DSL vs. Fiber vs. Cable
being close *is*. Every result also carries a `decision_quality`
(`low`/`medium`/`high`) and, when uncertain, `uncertainty_reasons` plus a
`detected_network_context` (ISP, gateway chain, double-NAT, local access, …) so
the user sees what *was* established even when the type wasn't.

## Security boundary

Only the user's own network is inspected, using passive or standard diagnostic
methods. No modem logins, no brute force, no neighbor-network scanning, no
intrusive operator-infrastructure probing. External-service probes are gated by
`--online`.
