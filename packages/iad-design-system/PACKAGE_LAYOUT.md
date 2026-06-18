# IAD Design System Package Layout

Imported from the IAD Internet Access Detector Design System handoff bundle.

This package keeps the handoff source intact and separate from runtime app code:

```text
packages/iad-design-system/
  tokens/       design tokens
  assets/       logos and marks
  components/   reference component implementations
  guidelines/   product and design rules
  ui_kits/      full console UI kit prototypes
  styles.css    global reference styles
  readme.md     original handoff README
  SKILL.md      original design-system usage instructions
```

Use this package as the canonical visual source for the desktop console. Runtime
React code lives under `apps/desktop/frontend`.
