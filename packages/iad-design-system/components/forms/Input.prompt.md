Text field with optional leading icon and right addon. Mono variant for IP/MAC/ASN entry.

```jsx
<Input value={q} onChange={setQ} placeholder="Search devices…" iconLeft={<SearchIcon/>} />
<Input value={ip} onChange={setIp} mono placeholder="192.168.1.0/24" />
```

Controlled. `onChange` receives the string value directly. Sunken background sets it apart from card surfaces.
