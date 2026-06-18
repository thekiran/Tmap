Primary action control for the IAD console — monochrome-first, with the accent reserved for the `primary` variant and color reserved for `danger`.

```jsx
<Button variant="primary" iconLeft={<RefreshIcon />}>Re-run scan</Button>
<Button>Import JSON</Button>
<Button variant="ghost" size="sm">Cancel</Button>
```

Variants: `primary` (accent fill), `secondary` (neutral outline — default), `ghost` (chromeless), `danger` (destructive only). Sizes: `sm` / `md` / `lg`. Supports `iconLeft` / `iconRight`, `fullWidth`, `disabled`. In this read-only tool, prefer `secondary`/`ghost`; reserve `danger` for blocking/reset actions.
