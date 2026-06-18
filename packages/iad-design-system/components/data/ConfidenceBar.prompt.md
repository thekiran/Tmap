The headline confidence meter — a labeled track filled to a 0–1 value, colored by band. Low confidence is calm gray (never red); the product is honest about uncertainty.

```jsx
<ConfidenceBar value={0.82} label="Overall confidence" />
<ConfidenceBar value={0.31} label="Access type" size="sm" />
```

Bands: `<0.45` Low (gray) · `0.45–0.75` Medium (amber) · `≥0.75` High (green). Exports a `band(value)` helper. Pair with an explanation when value is low — never imply certainty the data doesn't support.
