A labeled value readout — mono value with tabular numerals so columns align. Workhorse of dashboard cards.

```jsx
<MetricStat label="Downstream" value="412.6" unit="Mbps" />
<MetricStat label="Public IP" value="203.0.113.42" secondary="PTR: cpe-203-0-113-42.isp.net" />
```

Props: `tone`, `size` (`sm`/`md`/`lg`), `align`, `unit`, `secondary`. Keep values as already-formatted strings (use lib/format helpers).
