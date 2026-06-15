# Access-Type Classifier Redesign

This design makes Unknown a correct outcome when physical WAN evidence is absent.
The classifier improves certainty by collecting stronger evidence, not by raising
weak scores.

## Evidence Hierarchy

Tier A - direct physical WAN evidence:

- TR-181 interface stack: `Device.DSL.Line`, `Device.DSL.Channel`, `Device.PTM.Link`, `Device.ATM.Link`, `Device.Optical.Interface`, `Device.Ethernet.Link`, `Device.Cellular.Interface`.
- UPnP `WANCommonInterfaceConfig.GetCommonLinkProperties`: `WANAccessType`, layer-1 bitrates, physical link status.
- TR-064 WAN/DSL services: `WANDSLInterfaceConfig`, PTM/ATM/VDSL/ADSL fields, link status.
- SNMP opt-in: IF-MIB `ifType`, `ifDescr`, `ifName`, DSL/VDSL2/DOCSIS/optical/cellular MIB objects.
- Parsed line profile: VDSL2 profile, ADSL mode, DOCSIS version/channels, PON/optical type.

Tier B - CPE/device model evidence:

- HTTP/HTTPS fingerprint, UPnP device description, TLS cert CN/SAN, auth realm, favicon hash, MAC OUI.
- Local CPE model database mapping model patterns to supported WAN types.
- This can produce probable confidence, but is not final proof of the active WAN.

Tier C - local topology evidence:

- Default gateway, route table, traceroute private hops, gateway chain, double NAT, DHCP/NAT-PMP/PCP, IPv4/IPv6 architecture.

Tier D - performance evidence:

- Idle latency, jitter, packet loss, loaded latency, optional throughput profile.
- Never direct proof.

Tier E - regional/operator evidence:

- ASN, PTR/reverse DNS, ISP brand, known regional deployment hints.
- Never direct proof.

## Confidence Algorithm

Rules:

- Tier A direct and unambiguous evidence can confirm with confidence 0.85-0.98.
- Tier B model evidence is capped at 0.85 and produces `probable`, not `confirmed`.
- If only Tier C/D/E exists, confidence is capped at 0.40 and the result stays Unknown.
- If the top two categories differ by less than 0.12, return Unknown unless Tier A resolves it.
- Direct physical evidence wins over weaker hints.
- Conflicts cap confidence and may force `conflicting_evidence`.
- Latency, PTR, ASN, public IP, IPv6, CGNAT, and local Ethernet adapter names are never direct proof.

## Probe Design

- `gateway_discovery_probe`: read route table, find default gateway and active interface, ignore virtual adapters unless they are the only active route, identify private gateway chain and roles.
- `gateway_reachability_probe`: only private gateway candidates; check ICMP/TCP 80, 443, 1900, 5000, 7547; no public IP scans and no brute force.
- `upnp_discovery_probe`: SSDP M-SEARCH on local network only; parse device XML and IGD/WAN services.
- `upnp_wan_common_interface_probe`: read-only SOAP `GetCommonLinkProperties`; maps DSL, Cable, Ethernet, Other carefully. Ethernet becomes EthernetWAN, not Fiber.
- `tr064_probe`: descriptors and safe read-only calls only; authenticated calls require user credentials; auth-required without credentials is no direct evidence.
- `http_fingerprint_probe`: private gateway HTTP/HTTPS only; no login attempts; model evidence only.
- `cpe_model_database`: local JSON database in `rules/cpe_model_database.json`.
- `snmp_opt_in_probe`: disabled by default; requires user-provided read-only credentials; never guesses communities.
- `tr181_interface_stack_probe`: consumes supported local management API/user-provided CPE access; classifies only from explicit physical layer.
- `performance_profile_probe`: weak context only.

## Classification Rules

- Fiber confirmed by active optical/PON/ONT evidence, not by Ethernet adapter or low latency.
- VDSL confirmed by VDSL2-LINE-MIB, TR-181 DSL/PTM/VDSL2, TR-064 VDSL/PTM, or active DSL interface evidence.
- ADSL confirmed by ADSL-LINE-MIB, `ifType adsl(94)`, TR-181 ADSL mode, or explicit TR-064 DSL indicators.
- Cable/DOCSIS confirmed by DOCSIS MIB, docsCable interface types, or active cable modem telemetry.
- FWA/4G/5G confirmed by LTE/NR/WWAN CPE interface or cellular modem model with active interface evidence.
- Satellite confirmed by known terminal/CPE or satellite interface evidence.
- WISP confirmed by fixed-wireless CPE/radio bridge evidence.
- EthernetWAN confirmed by Ethernet WAN access type or interface stack, but it is not automatically Fiber.

## JSON Schema

The Go model now emits the legacy flat fields plus these UI/API fields:

```json
{
  "classification": {
    "primary_type": "Unknown | Fiber | VDSL | ADSL | DSL | Cable | FWA | Satellite | WISP | EthernetWAN",
    "subtype": null,
    "confidence": 0.0,
    "decision_quality": "high | medium | low",
    "state": "confirmed | probable | possible | insufficient_evidence | conflicting_evidence",
    "safe_to_display_as_final": false
  },
  "candidates": [
    {
      "type": "Fiber",
      "score": 0.07,
      "evidence_strength": "weak",
      "supporting_evidence": [],
      "missing_evidence": []
    }
  ],
  "evidence_tiers": {
    "direct_physical": {},
    "device_model": {},
    "topology": {},
    "performance": {},
    "regional": {}
  },
  "data_quality": {"has_conflicts": false, "conflicts": []},
  "conflicts": [],
  "next_best_probes": [],
  "ui": {"headline": "", "summary": "", "badges": [], "warnings": []}
}
```

## UI Rules

- `confirmed`, confidence >= 0.85, direct evidence: `Detected: VDSL2`.
- 0.60-0.84: `Probable: VDSL2`.
- 0.40-0.59: `Possible: VDSL2`.
- Below 0.40 or no physical/model evidence: `Unknown - not enough physical evidence`.
- Equal weak scores are shown as candidates only.
- Badges: Direct WAN evidence, CPE model found, UPnP available, TR-064 available, SNMP opt-in available, Performance only, Operator hint only.

## Pseudocode

```text
classifyAccessType(results):
  bag = mergeProbeResults(results)
  scores = scoreRulesAndHints(bag)
  conflicts = detectConflicts(results, bag, scores)
  direct = hasTierA(bag)
  model = hasTierB(bag)

  confidence = distributionConfidence(scores)
  if direct:
    confidence = max(confidence, directEvidenceConfidence(bag))
    confidence = min(confidence, 0.98)
  else if model:
    confidence = min(max(confidence, modelConfidence(bag)), 0.85)
  else:
    confidence = min(confidence, 0.40)

  if conflicts.high:
    confidence = min(confidence, 0.55)

  if no scores:
    return Unknown(insufficient_evidence)
  if topCategoryScore < 0.35:
    return Unknown(insufficient_evidence)
  if categoryMargin < 0.12 and not direct:
    return Unknown(insufficient_evidence)
  if not direct and not model:
    return Unknown(insufficient_evidence)
  if conflicts.classification:
    return Unknown(conflicting_evidence)

  return verdict(stateFrom(confidence, direct), safeFinal = direct and confidence >= 0.85)
```

```text
mergeProbeResults(results):
  for each successful probe:
    normalize hints and evidence into tiers
    merge gateway devices by IP
    keep model evidence separate from direct physical evidence
    parse CPE key/value telemetry into line_profile
    never promote local adapter, ASN, PTR, latency, CGNAT, or public IP to Tier A
  return evidence_bag
```

```text
detectConflicts(results, bag, scores):
  compare gateway reachability reports by source
  compare traceroute chain against gateway_chain_probe
  compare NAT/double-NAT facts
  flag Ethernet WAN direct evidence conflicting with high DSL/Fiber weaker scores
  return conflicts with field, values, severity, and effect
```

```text
nextBestProbes(bag):
  if no direct WAN evidence:
    suggest UPnP WANCommonInterfaceConfig, TR-064, TR-181, SNMP opt-in
  if no model:
    suggest HTTP fingerprint and UPnP description
  if gateway chain unclear:
    suggest gateway reachability/topology probes
  always include safety note and expected evidence
```

## TypeScript Data Model

```ts
type PrimaryType =
  | "Unknown" | "Fiber" | "VDSL" | "ADSL" | "DSL" | "Cable"
  | "FWA" | "Satellite" | "WISP" | "EthernetWAN";

type ClassificationState =
  | "confirmed" | "probable" | "possible"
  | "insufficient_evidence" | "conflicting_evidence";

interface Classification {
  primary_type: PrimaryType;
  subtype: string | null;
  confidence: number;
  decision_quality: "high" | "medium" | "low";
  state: ClassificationState;
  safe_to_display_as_final: boolean;
}

interface EvidenceTier {
  present: boolean;
  strength: "none" | "weak" | "medium" | "strong";
  confidence?: number;
  sources?: string[];
  evidence?: string[];
}

interface AccessCandidate {
  category?: string;
  type: string;
  subtype?: string;
  score: number;
  confidence?: number;
  evidence_strength: "none" | "weak" | "medium" | "strong";
  supporting_evidence: string[];
  missing_evidence: string[];
}

interface DataConflict {
  field: string;
  values: Array<{ source: string; value: unknown }>;
  severity: "low" | "medium" | "high";
  effect: "downgrade_confidence" | "set_conflicting_evidence";
}
```

## Sample Weak Output

```json
{
  "classification": {
    "primary_type": "Unknown",
    "subtype": null,
    "confidence": 0.19,
    "decision_quality": "low",
    "state": "insufficient_evidence",
    "safe_to_display_as_final": false
  },
  "candidates": [
    {
      "type": "Fiber",
      "score": 0.07,
      "evidence_strength": "weak",
      "supporting_evidence": ["performance context", "operator/regional context"],
      "missing_evidence": ["TR-181/TR-064/SNMP/UPnP WAN physical-layer evidence"]
    },
    {
      "type": "VDSL",
      "score": 0.07,
      "evidence_strength": "weak",
      "supporting_evidence": ["performance context", "operator/regional context"],
      "missing_evidence": ["TR-181/TR-064/SNMP/UPnP WAN physical-layer evidence"]
    }
  ],
  "evidence_tiers": {
    "direct_physical": {"present": false, "strength": "none"},
    "device_model": {"present": false, "strength": "none"},
    "topology": {"present": true, "strength": "weak"},
    "performance": {"present": true, "strength": "weak"},
    "regional": {"present": true, "strength": "weak"}
  },
  "next_best_probes": [
    {"probe_name": "upnp_igd_deep_probe_v2", "expected_evidence": "WANAccessType and link properties"},
    {"probe_name": "tr064_probe_v2", "expected_evidence": "WAN/DSL service descriptors"},
    {"probe_name": "snmp_probe_opt_in", "expected_evidence": "DSL/DOCSIS/optical/cellular MIB objects"}
  ],
  "ui": {
    "headline": "Unknown - not enough physical evidence",
    "summary": "The strongest candidates are Fiber, VDSL, but the evidence is not strong enough to classify.",
    "badges": ["Performance only", "Operator hint only"],
    "warnings": ["No strong physical-layer evidence of the access type was found."]
  }
}
```

## Implementation Tasks

1. Keep Tier A/B/C/D/E separated in the model and scoring pipeline.
2. Ensure HTTP/model fingerprints never populate direct physical evidence.
3. Add TR-181 and real SNMP backends only behind explicit user authorization.
4. Expand `rules/cpe_model_database.json` as devices are validated.
5. Persist conflict records and monitor Unknown-rate over time.
6. Add UI rendering for `classification`, `evidence_tiers`, `data_quality`, and `next_best_probes`.
7. Add fixtures for more operators and CPEs, especially ambiguous ONT-bridge/router chains.

## Safety Rules

- Only scan local/private gateway IPs by default.
- Do not scan public IP ranges unless explicitly authorized for topology mapping.
- Do not brute force router logins or SNMP communities.
- Do not bypass authentication.
- SNMP is opt-in and read-only.
- TR-064 authenticated calls require user-provided credentials.
- All probes must use timeouts and bounded candidate lists.
