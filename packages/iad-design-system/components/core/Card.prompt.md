The primary container surface. Optional header (eyebrow + title + actions) and footer; depth comes from hairline borders, not shadow.

```jsx
<Card eyebrow="Local interface" title="Ethernet" actions={<IconButton label="Copy"><CopyIcon/></IconButton>}>
  …body…
</Card>
```

Props: `padding` (`none`/`sm`/`md`/`lg`), `raised`, `interactive` (hover affordance for clickable cards). Compose MetricStat, ConfidenceBar, tables, etc. inside.
