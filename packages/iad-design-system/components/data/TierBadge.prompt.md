Labels an evidence tier / topology layer with its fixed hue. Tiers are the categorical backbone — one color each, used everywhere.

```jsx
<TierBadge tier="physical" />
<TierBadge tier="l2" appearance="dot" />
<TierBadge tier="isp" appearance="solid" />
```

Tiers: `physical` (coral) · `l2` (blue) · `l3` (violet) · `nat` (pink) · `isp` (teal). Appearances: `subtle` chip, `solid`, `dot`. Use the same tier color in legends and layer toggles for consistency.
