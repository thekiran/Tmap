Compact single-select for view switches (Table/List), theme (Dark/Light), and layout engine.

```jsx
<SegmentedControl
  value={view}
  onChange={setView}
  options={[{value:'table',label:'Table'},{value:'list',label:'List'}]} />
```

Controlled. Options take `{ value, label, icon? }`. Use `fullWidth` to stretch across a panel.
