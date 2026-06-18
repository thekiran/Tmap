Square icon-only control for toolbars (topology zoom/fit/reset) and table row actions. Always pass `label` — it drives both `aria-label` and the tooltip.

```jsx
<IconButton label="Fit view"><FitIcon /></IconButton>
<IconButton label="L2 layer" active variant="outline"><LayersIcon /></IconButton>
```

Variants: `ghost` (default), `outline`. Use `active` for toggle toolbars (sets `aria-pressed`). Sizes `sm`/`md`/`lg`.
