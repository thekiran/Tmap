# IAD Desktop Console — UI kit

A high-fidelity recreation of the planned **IAD / iad-agent** Windows desktop app
(Wails + React + React Flow). It renders a single authorized scan
(`NormalizedScanReport`) as an interactive, **read-only** network console.

> The agent's UI was specified but not yet built (see the Tmap repo's
> `docs/design/` dossier). This kit realizes that brief in the IAD design system.

## Run / preview
Open `index.html`. It loads the design-system bundle (`../../_ds_bundle.js`) and
mounts the app shell into a fixed desktop surface.

## Screens
- **Dashboard** — verdict + calibrated confidence, candidate access types, local
  interface, public IP / ISP / ASN, IPv4 NAT (CGNAT), IPv6, gateway chain,
  confidence breakdown, next-best probes, warnings.
- **Topology** — read-only SVG map: pan (drag bg), zoom (wheel), drag nodes to
  reposition, select nodes/edges, layer toggles (L2 / L3 / NAT / ISP / unknown /
  low-confidence), legend. No create/delete; layout positions are UI state only.
  *(Production app uses React Flow + ELK; this is a faithful recreation of that
  read-only mode.)*
- **Devices** — inventory with search, type/reachability filters, sortable table
  and list views; click a row to open the details drawer.
- **Evidence** — probe explorer: status enum, tier, confidence, limitations, raw
  JSON. Flags the "success but empty evidence → normalize to `no_data`" case.
- **Reports** — import / export JSON, summary export, copy diagnostic summary;
  scan-compare is an explicit non-functional placeholder.
- **Settings** — theme, layout engine, layer defaults, persist/reset positions,
  safe-mode indicator.

## Files
| File | Role |
|---|---|
| `index.html` | Entry; loads bundle + all screen scripts, mounts `AppShell`. |
| `data.js` | One realistic `NormalizedScanReport` fixture (`window.IAD_SCAN`). |
| `format.js` | Formatting helpers (`window.fmt`). |
| `icons.jsx` | Inline-SVG icon set (`window.Icons`), Lucide-style. |
| `AppShell.jsx` | Sidebar, top status bar, content router, log strip. |
| `DetailsDrawer.jsx` | Right-side device / node / edge / probe panel. |
| `Dashboard.jsx` · `Topology.jsx` · `Devices.jsx` · `Evidence.jsx` · `ReportsSettings.jsx` | Screens. |

## Design rules honored
Read-only (no delete, no fake devices/edges); inferred ≠ confirmed (dashed vs
solid lines, "(inferred)" labels); traceroute hops shown as L3 routers, not
switches; ISP-internal topology marked unobservable; UI layout state kept
separate from immutable scan evidence; monochrome theme with status-only color.
