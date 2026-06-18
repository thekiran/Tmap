Fixed badge for the IAD probe-status enum. Locked tone + dot per status so the same state always reads the same.

```jsx
<ProbeStatusBadge status="success" />
<ProbeStatusBadge status="no_data" />
<ProbeStatusBadge status="blocked" size="sm" />
```

Statuses: `success` (green) · `partial` (amber) · `no_data` / `skipped` (gray) · `failed` (red) · `blocked` (violet). Never render `success` for a probe with empty evidence — normalize to `no_data` instead.
