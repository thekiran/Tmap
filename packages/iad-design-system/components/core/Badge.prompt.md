Small label chip for statuses, counts, and categorical tags. Tone is semantic — pick the tone that matches meaning, never for decoration.

```jsx
<Badge tone="success">Reachable</Badge>
<Badge tone="warn" uppercase mono>partial</Badge>
<Badge tone="neutral" appearance="outline" mono>ASN 3320</Badge>
```

Tones: `neutral` `accent` `success` `warn` `danger` `info` `blocked`. Appearances: `subtle` (tinted bg, default), `solid`, `outline`. `mono` + `uppercase` are ideal for probe-status enums.
