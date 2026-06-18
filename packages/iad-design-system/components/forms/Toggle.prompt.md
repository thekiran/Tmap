Accessible on/off switch. Monochrome track; accent fills only when on. Used in Settings and the topology layer panel.

```jsx
<Toggle checked={showLowConf} onChange={setShowLowConf}
        label="Show low-confidence edges"
        description="Render edges below the 0.45 band, visually muted." />
```

Pass `label`/`description` for a full settings row, or use bare for a compact inline switch. Controlled via `checked` + `onChange`.
