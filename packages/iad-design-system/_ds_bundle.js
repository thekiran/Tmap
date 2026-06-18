/* @ds-bundle: {"format":3,"namespace":"IADInternetAccessDetectorDesignSystem_019e02","components":[{"name":"Badge","sourcePath":"components/core/Badge.jsx"},{"name":"Button","sourcePath":"components/core/Button.jsx"},{"name":"Card","sourcePath":"components/core/Card.jsx"},{"name":"IconButton","sourcePath":"components/core/IconButton.jsx"},{"name":"StatusDot","sourcePath":"components/core/StatusDot.jsx"},{"name":"ConfidenceBar","sourcePath":"components/data/ConfidenceBar.jsx"},{"name":"MetricStat","sourcePath":"components/data/MetricStat.jsx"},{"name":"ProbeStatusBadge","sourcePath":"components/data/ProbeStatusBadge.jsx"},{"name":"TierBadge","sourcePath":"components/data/TierBadge.jsx"},{"name":"Input","sourcePath":"components/forms/Input.jsx"},{"name":"SegmentedControl","sourcePath":"components/forms/SegmentedControl.jsx"},{"name":"Toggle","sourcePath":"components/forms/Toggle.jsx"}],"sourceHashes":{"components/core/Badge.jsx":"bbdedc3c5e3a","components/core/Button.jsx":"aeadbf0d1d49","components/core/Card.jsx":"b7ab397e4171","components/core/IconButton.jsx":"e2dc7b1218d7","components/core/StatusDot.jsx":"60261580aed0","components/data/ConfidenceBar.jsx":"d46b5c9883db","components/data/MetricStat.jsx":"6e809d1fd4be","components/data/ProbeStatusBadge.jsx":"d3aa5ca79c8e","components/data/TierBadge.jsx":"19fb1d7cc27c","components/forms/Input.jsx":"28bc5d1c2b0c","components/forms/SegmentedControl.jsx":"2ab90aa65818","components/forms/Toggle.jsx":"0a7ef8c4dcbc","ui_kits/console/AppShell.jsx":"8174f38a3f22","ui_kits/console/Dashboard.jsx":"716fb00894fa","ui_kits/console/DetailsDrawer.jsx":"7119b5870dff","ui_kits/console/Devices.jsx":"f29291e746ac","ui_kits/console/Evidence.jsx":"a178ec016a01","ui_kits/console/ReportsSettings.jsx":"c15e4bde95c8","ui_kits/console/Topology.jsx":"46ce1e39a093","ui_kits/console/data.js":"c7f7c70c58e6","ui_kits/console/format.js":"8804f9e2ec30","ui_kits/console/icons.jsx":"c614ac1628c4"},"inlinedExternals":[],"unexposedExports":[{"name":"band","sourcePath":"components/data/ConfidenceBar.jsx"}]} */

(() => {

const __ds_ns = (window.IADInternetAccessDetectorDesignSystem_019e02 = window.IADInternetAccessDetectorDesignSystem_019e02 || {});

const __ds_scope = {};

(__ds_ns.__errors = __ds_ns.__errors || []);

// components/core/Badge.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Badge — small label chip for statuses, counts, and categorical tags.
 * Tone maps to the semantic token set. Use "subtle" (default) for tinted-bg
 * chips, "solid" for high-emphasis, "outline" for quiet metadata.
 */
function Badge({
  tone = 'neutral',
  appearance = 'subtle',
  size = 'md',
  mono = false,
  uppercase = false,
  children,
  style = {},
  ...rest
}) {
  const tones = {
    neutral: {
      c: 'var(--fg-2)',
      bg: 'var(--neutral-bg)',
      solid: 'var(--neutral)'
    },
    accent: {
      c: 'var(--accent-bright)',
      bg: 'var(--accent-ghost)',
      solid: 'var(--accent-base)'
    },
    success: {
      c: 'var(--ok)',
      bg: 'var(--ok-bg)',
      solid: 'var(--ok)'
    },
    warn: {
      c: 'var(--warn)',
      bg: 'var(--warn-bg)',
      solid: 'var(--warn)'
    },
    danger: {
      c: 'var(--danger)',
      bg: 'var(--danger-bg)',
      solid: 'var(--danger)'
    },
    info: {
      c: 'var(--info)',
      bg: 'var(--info-bg)',
      solid: 'var(--info)'
    },
    blocked: {
      c: 'var(--blocked)',
      bg: 'var(--blocked-bg)',
      solid: 'var(--blocked)'
    }
  };
  const t = tones[tone] || tones.neutral;
  const sizes = {
    sm: {
      h: 18,
      px: 6,
      fs: 'var(--text-2xs)'
    },
    md: {
      h: 22,
      px: 8,
      fs: 'var(--text-xs)'
    }
  };
  const sz = sizes[size] || sizes.md;
  let look;
  if (appearance === 'solid') {
    look = {
      background: t.solid,
      color: 'var(--fg-on-accent)',
      border: '1px solid transparent'
    };
  } else if (appearance === 'outline') {
    look = {
      background: 'transparent',
      color: t.c,
      border: '1px solid var(--hairline-strong)'
    };
  } else {
    look = {
      background: t.bg,
      color: t.c,
      border: '1px solid transparent'
    };
  }
  return /*#__PURE__*/React.createElement("span", _extends({
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 5,
      height: sz.h,
      padding: `0 ${sz.px}px`,
      borderRadius: 'var(--radius-xs)',
      fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)',
      fontSize: sz.fs,
      fontWeight: 'var(--fw-medium)',
      lineHeight: 1,
      letterSpacing: uppercase ? 'var(--ls-caps)' : '0.01em',
      textTransform: uppercase ? 'uppercase' : 'none',
      whiteSpace: 'nowrap',
      ...look,
      ...style
    }
  }, rest), children);
}
Object.assign(__ds_scope, { Badge });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Badge.jsx", error: String((e && e.message) || e) }); }

// components/core/Button.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Button — primary action control for the IAD console.
 * Monochrome-first: the default ("secondary") is a neutral outline button;
 * "primary" uses the restrained accent; "ghost" is chromeless; "danger" only
 * for genuinely destructive/blocking actions (rare in a read-only tool).
 */
function Button({
  variant = 'secondary',
  size = 'md',
  iconLeft = null,
  iconRight = null,
  disabled = false,
  fullWidth = false,
  type = 'button',
  children,
  style = {},
  ...rest
}) {
  const sizes = {
    sm: {
      height: 28,
      padding: '0 10px',
      font: 'var(--text-xs)',
      gap: 6,
      radius: 'var(--radius-sm)'
    },
    md: {
      height: 34,
      padding: '0 14px',
      font: 'var(--text-sm)',
      gap: 8,
      radius: 'var(--radius-md)'
    },
    lg: {
      height: 42,
      padding: '0 18px',
      font: 'var(--text-base)',
      gap: 8,
      radius: 'var(--radius-md)'
    }
  };
  const s = sizes[size] || sizes.md;
  const variants = {
    primary: {
      background: 'var(--accent-base)',
      color: 'var(--fg-on-accent)',
      border: '1px solid var(--accent-base)'
    },
    secondary: {
      background: 'var(--surface-2)',
      color: 'var(--fg-1)',
      border: '1px solid var(--hairline-strong)'
    },
    ghost: {
      background: 'transparent',
      color: 'var(--fg-2)',
      border: '1px solid transparent'
    },
    danger: {
      background: 'transparent',
      color: 'var(--danger)',
      border: '1px solid var(--danger-bg)'
    }
  };
  const v = variants[variant] || variants.secondary;
  return /*#__PURE__*/React.createElement("button", _extends({
    type: type,
    disabled: disabled,
    "data-variant": variant,
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: s.gap,
      height: s.height,
      padding: s.padding,
      width: fullWidth ? '100%' : 'auto',
      font: `var(--fw-medium) ${s.font}/1 var(--font-sans)`,
      letterSpacing: '0.01em',
      borderRadius: s.radius,
      cursor: disabled ? 'not-allowed' : 'pointer',
      opacity: disabled ? 0.45 : 1,
      whiteSpace: 'nowrap',
      userSelect: 'none',
      transition: 'background var(--dur-fast) var(--ease-out), border-color var(--dur-fast) var(--ease-out), opacity var(--dur-fast) var(--ease-out)',
      ...v,
      ...style
    }
  }, rest), iconLeft && /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      width: '1em',
      height: '1em'
    }
  }, iconLeft), children, iconRight && /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      width: '1em',
      height: '1em'
    }
  }, iconRight));
}
Object.assign(__ds_scope, { Button });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Button.jsx", error: String((e && e.message) || e) }); }

// components/core/Card.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Card — the primary container surface. Optional header (title + eyebrow +
 * actions) and footer. Padding and emphasis are tunable. Depth on dark comes
 * from hairline borders + faint ambient shadow, never heavy drop shadows.
 */
function Card({
  title = null,
  eyebrow = null,
  actions = null,
  footer = null,
  padding = 'md',
  raised = false,
  interactive = false,
  children,
  style = {},
  bodyStyle = {},
  ...rest
}) {
  const pads = {
    none: 0,
    sm: 'var(--space-3)',
    md: 'var(--space-5)',
    lg: 'var(--space-6)'
  };
  const p = pads[padding] != null ? pads[padding] : pads.md;
  return /*#__PURE__*/React.createElement("section", _extends({
    style: {
      background: raised ? 'var(--surface-2)' : 'var(--surface-card)',
      border: '1px solid var(--hairline)',
      borderRadius: 'var(--radius-lg)',
      boxShadow: raised ? 'var(--shadow-sm)' : 'var(--shadow-xs)',
      display: 'flex',
      flexDirection: 'column',
      minWidth: 0,
      transition: interactive ? 'border-color var(--dur-fast) var(--ease-out), background var(--dur-fast) var(--ease-out)' : 'none',
      ...style
    }
  }, rest), (title || eyebrow || actions) && /*#__PURE__*/React.createElement("header", {
    style: {
      display: 'flex',
      alignItems: 'flex-start',
      justifyContent: 'space-between',
      gap: 'var(--space-3)',
      padding: `var(--space-4) ${typeof p === 'number' ? p : p} var(--space-3)`,
      paddingBottom: 'var(--space-3)',
      borderBottom: '1px solid var(--hairline)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      minWidth: 0
    }
  }, eyebrow && /*#__PURE__*/React.createElement("div", {
    style: {
      font: 'var(--type-overline)',
      letterSpacing: 'var(--ls-caps)',
      textTransform: 'uppercase',
      color: 'var(--fg-3)',
      marginBottom: 4
    }
  }, eyebrow), title && /*#__PURE__*/React.createElement("h3", {
    style: {
      font: 'var(--type-h3)',
      color: 'var(--fg-1)',
      margin: 0
    }
  }, title)), actions && /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-2)',
      flex: '0 0 auto'
    }
  }, actions)), /*#__PURE__*/React.createElement("div", {
    style: {
      padding: p,
      minWidth: 0,
      flex: 1,
      ...bodyStyle
    }
  }, children), footer && /*#__PURE__*/React.createElement("footer", {
    style: {
      padding: `var(--space-3) ${typeof p === 'number' ? p : p}`,
      borderTop: '1px solid var(--hairline)',
      color: 'var(--fg-3)',
      fontSize: 'var(--text-xs)'
    }
  }, footer));
}
Object.assign(__ds_scope, { Card });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Card.jsx", error: String((e && e.message) || e) }); }

// components/core/IconButton.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * IconButton — square, icon-only control. Used heavily in the topology toolbar
 * (zoom, fit, reset) and table row actions. Always pass an aria-label.
 */
function IconButton({
  variant = 'ghost',
  size = 'md',
  active = false,
  disabled = false,
  label,
  children,
  style = {},
  ...rest
}) {
  const dims = {
    sm: 28,
    md: 34,
    lg: 40
  };
  const d = dims[size] || dims.md;
  const base = {
    ghost: {
      background: active ? 'var(--surface-3)' : 'transparent',
      color: active ? 'var(--fg-1)' : 'var(--fg-2)',
      border: '1px solid ' + (active ? 'var(--hairline-strong)' : 'transparent')
    },
    outline: {
      background: active ? 'var(--surface-3)' : 'var(--surface-2)',
      color: 'var(--fg-1)',
      border: '1px solid var(--hairline-strong)'
    }
  };
  const v = base[variant] || base.ghost;
  return /*#__PURE__*/React.createElement("button", _extends({
    type: "button",
    "aria-label": label,
    "aria-pressed": active || undefined,
    disabled: disabled,
    title: label,
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      width: d,
      height: d,
      flex: '0 0 auto',
      borderRadius: 'var(--radius-md)',
      cursor: disabled ? 'not-allowed' : 'pointer',
      opacity: disabled ? 0.45 : 1,
      transition: 'background var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out), border-color var(--dur-fast) var(--ease-out)',
      ...v,
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      width: Math.round(d * 0.46),
      height: Math.round(d * 0.46)
    }
  }, children));
}
Object.assign(__ds_scope, { IconButton });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/IconButton.jsx", error: String((e && e.message) || e) }); }

// components/core/StatusDot.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * StatusDot — a small filled dot encoding a state, optionally with a label.
 * Used for reachability, probe status, online/offline, layer legends.
 * The "pulse" option adds a calm breathing ring for live/active states only.
 */
function StatusDot({
  tone = 'neutral',
  size = 8,
  pulse = false,
  label = null,
  style = {},
  ...rest
}) {
  const tones = {
    neutral: 'var(--neutral)',
    success: 'var(--ok)',
    warn: 'var(--warn)',
    danger: 'var(--danger)',
    info: 'var(--info)',
    accent: 'var(--accent-base)',
    blocked: 'var(--blocked)'
  };
  const c = tones[tone] || tones.neutral;
  const dot = /*#__PURE__*/React.createElement("span", {
    style: {
      position: 'relative',
      display: 'inline-flex',
      width: size,
      height: size,
      flex: '0 0 auto'
    }
  }, pulse && /*#__PURE__*/React.createElement("span", {
    style: {
      position: 'absolute',
      inset: 0,
      borderRadius: '50%',
      background: c,
      opacity: 0.5,
      animation: 'iad-dot-pulse 1.8s var(--ease-out) infinite'
    }
  }), /*#__PURE__*/React.createElement("span", {
    style: {
      width: size,
      height: size,
      borderRadius: '50%',
      background: c,
      position: 'relative'
    }
  }), /*#__PURE__*/React.createElement("style", null, '@keyframes iad-dot-pulse{0%{transform:scale(1);opacity:.5}70%{transform:scale(2.4);opacity:0}100%{opacity:0}}'));
  if (label == null) return React.cloneElement(dot, {
    ...rest,
    style: {
      ...dot.props.style,
      ...style
    }
  });
  return /*#__PURE__*/React.createElement("span", _extends({
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 7,
      color: 'var(--fg-2)',
      fontSize: 'var(--text-sm)',
      ...style
    }
  }, rest), dot, /*#__PURE__*/React.createElement("span", null, label));
}
Object.assign(__ds_scope, { StatusDot });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/StatusDot.jsx", error: String((e && e.message) || e) }); }

// components/data/ConfidenceBar.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * ConfidenceBar — the headline "how sure are we" control. Renders a labeled
 * track filled to a 0–1 confidence with the band color (Low/Med/High), plus
 * the numeric percentage and band word. Honest by design: low confidence is
 * calm gray, never red. Bands follow the IAD calibration:
 *   < 0.45 Low · 0.45–0.75 Medium · ≥ 0.75 High
 */
function band(value) {
  if (value >= 0.75) return 'high';
  if (value >= 0.45) return 'medium';
  return 'low';
}
function ConfidenceBar({
  value = 0,
  showLabel = true,
  showValue = true,
  label = 'Confidence',
  size = 'md',
  style = {},
  ...rest
}) {
  const v = Math.max(0, Math.min(1, value));
  const b = band(v);
  const colors = {
    low: {
      fg: 'var(--conf-low)',
      bg: 'var(--conf-low-bg)',
      word: 'Low'
    },
    medium: {
      fg: 'var(--conf-med)',
      bg: 'var(--conf-med-bg)',
      word: 'Medium'
    },
    high: {
      fg: 'var(--conf-high)',
      bg: 'var(--conf-high-bg)',
      word: 'High'
    }
  };
  const c = colors[b];
  const heights = {
    sm: 5,
    md: 7,
    lg: 9
  };
  const h = heights[size] || heights.md;
  const pct = Math.round(v * 100);
  return /*#__PURE__*/React.createElement("div", _extends({
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 6,
      minWidth: 0,
      ...style
    }
  }, rest), (showLabel || showValue) && /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'baseline',
      justifyContent: 'space-between',
      gap: 8
    }
  }, showLabel && /*#__PURE__*/React.createElement("span", {
    style: {
      font: 'var(--type-overline)',
      letterSpacing: 'var(--ls-caps)',
      textTransform: 'uppercase',
      color: 'var(--fg-3)'
    }
  }, label), showValue && /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      alignItems: 'baseline',
      gap: 6
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-mono)',
      fontVariantNumeric: 'tabular-nums',
      fontWeight: 'var(--fw-semibold)',
      fontSize: 'var(--text-sm)',
      color: c.fg
    }
  }, pct, "%"), /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-2xs)',
      color: c.fg,
      textTransform: 'uppercase',
      letterSpacing: 'var(--ls-caps)',
      fontWeight: 'var(--fw-semibold)'
    }
  }, c.word))), /*#__PURE__*/React.createElement("div", {
    role: "meter",
    "aria-valuenow": pct,
    "aria-valuemin": 0,
    "aria-valuemax": 100,
    "aria-label": `${label}: ${pct}% (${c.word})`,
    style: {
      height: h,
      borderRadius: 'var(--radius-pill)',
      background: 'var(--surface-3)',
      overflow: 'hidden'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      width: `${pct}%`,
      height: '100%',
      background: c.fg,
      borderRadius: 'var(--radius-pill)',
      transition: 'width var(--dur-meter) var(--ease-out)'
    }
  })));
}
Object.assign(__ds_scope, { band, ConfidenceBar });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/data/ConfidenceBar.jsx", error: String((e && e.message) || e) }); }

// components/data/MetricStat.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * MetricStat — a labeled value readout. The value uses the mono family with
 * tabular numerals so columns of stats align. Optional unit, secondary line,
 * and a small delta/qualifier. The workhorse of the dashboard cards.
 */
function MetricStat({
  label,
  value,
  unit = null,
  secondary = null,
  tone = 'default',
  size = 'md',
  align = 'left',
  style = {},
  ...rest
}) {
  const tones = {
    default: 'var(--fg-1)',
    accent: 'var(--accent-bright)',
    success: 'var(--ok)',
    warn: 'var(--warn)',
    danger: 'var(--danger)',
    muted: 'var(--fg-3)'
  };
  const c = tones[tone] || tones.default;
  const sizes = {
    sm: {
      v: 'var(--text-md)',
      l: 'var(--text-2xs)'
    },
    md: {
      v: 'var(--text-xl)',
      l: 'var(--text-xs)'
    },
    lg: {
      v: 'var(--text-2xl)',
      l: 'var(--text-xs)'
    }
  };
  const sz = sizes[size] || sizes.md;
  return /*#__PURE__*/React.createElement("div", _extends({
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 4,
      alignItems: align === 'right' ? 'flex-end' : 'flex-start',
      textAlign: align,
      minWidth: 0,
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      font: 'var(--type-overline)',
      letterSpacing: 'var(--ls-caps)',
      textTransform: 'uppercase',
      color: 'var(--fg-3)',
      fontSize: sz.l
    }
  }, label), /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      alignItems: 'baseline',
      gap: 5,
      minWidth: 0
    }
  }, /*#__PURE__*/React.createElement("span", {
    className: "iad-num",
    style: {
      fontFamily: 'var(--font-mono)',
      fontVariantNumeric: 'tabular-nums',
      fontWeight: 'var(--fw-semibold)',
      fontSize: sz.v,
      color: c,
      lineHeight: 1.1,
      letterSpacing: 'var(--ls-snug)',
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      whiteSpace: 'nowrap'
    }
  }, value), unit && /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-mono)',
      fontSize: 'var(--text-xs)',
      color: 'var(--fg-3)'
    }
  }, unit)), secondary && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-xs)',
      color: 'var(--fg-3)',
      lineHeight: 1.4
    }
  }, secondary));
}
Object.assign(__ds_scope, { MetricStat });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/data/MetricStat.jsx", error: String((e && e.message) || e) }); }

// components/data/ProbeStatusBadge.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * ProbeStatusBadge — a fixed, opinionated badge for the IAD probe-status enum.
 * Each status has a locked tone + glyph so the same state always looks the same
 * across Evidence, Devices, and Reports. Statuses:
 *   success · partial · no_data · skipped · failed · blocked
 */
const MAP = {
  success: {
    word: 'Success',
    c: 'var(--ok)',
    bg: 'var(--ok-bg)'
  },
  partial: {
    word: 'Partial',
    c: 'var(--partial)',
    bg: 'var(--partial-bg)'
  },
  no_data: {
    word: 'No data',
    c: 'var(--neutral)',
    bg: 'var(--neutral-bg)'
  },
  skipped: {
    word: 'Skipped',
    c: 'var(--neutral)',
    bg: 'var(--neutral-bg)'
  },
  failed: {
    word: 'Failed',
    c: 'var(--danger)',
    bg: 'var(--danger-bg)'
  },
  blocked: {
    word: 'Blocked',
    c: 'var(--blocked)',
    bg: 'var(--blocked-bg)'
  }
};
function ProbeStatusBadge({
  status = 'no_data',
  size = 'md',
  style = {},
  ...rest
}) {
  const m = MAP[status] || MAP.no_data;
  const sz = size === 'sm' ? {
    h: 18,
    px: 7,
    dot: 5,
    fs: 'var(--text-2xs)'
  } : {
    h: 22,
    px: 9,
    dot: 6,
    fs: 'var(--text-xs)'
  };
  return /*#__PURE__*/React.createElement("span", _extends({
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 6,
      height: sz.h,
      padding: `0 ${sz.px}px`,
      borderRadius: 'var(--radius-xs)',
      background: m.bg,
      color: m.c,
      fontFamily: 'var(--font-mono)',
      fontSize: sz.fs,
      fontWeight: 'var(--fw-semibold)',
      letterSpacing: 'var(--ls-wide)',
      textTransform: 'uppercase',
      whiteSpace: 'nowrap',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      width: sz.dot,
      height: sz.dot,
      borderRadius: '50%',
      background: m.c,
      flex: '0 0 auto'
    }
  }), m.word);
}
Object.assign(__ds_scope, { ProbeStatusBadge });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/data/ProbeStatusBadge.jsx", error: String((e && e.message) || e) }); }

// components/data/TierBadge.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * TierBadge — labels an evidence tier / topology layer with its fixed hue.
 * Tiers are the categorical backbone of the product; each has ONE color used
 * everywhere (legend, edges context, evidence rows, layer toggles).
 *   physical · l2 · l3 · nat · isp
 */
const TIERS = {
  physical: {
    word: 'Physical',
    c: 'var(--tier-physical)',
    bg: 'var(--tier-physical-bg)'
  },
  l2: {
    word: 'L2 · Link',
    c: 'var(--tier-l2)',
    bg: 'var(--tier-l2-bg)'
  },
  l3: {
    word: 'L3 · Routing',
    c: 'var(--tier-l3)',
    bg: 'var(--tier-l3-bg)'
  },
  nat: {
    word: 'NAT',
    c: 'var(--tier-nat)',
    bg: 'var(--tier-nat-bg)'
  },
  isp: {
    word: 'ISP Route',
    c: 'var(--tier-isp)',
    bg: 'var(--tier-isp-bg)'
  }
};
function TierBadge({
  tier = 'l2',
  label = null,
  appearance = 'subtle',
  style = {},
  ...rest
}) {
  const t = TIERS[tier] || TIERS.l2;
  const look = appearance === 'solid' ? {
    background: t.c,
    color: 'var(--fg-on-accent)'
  } : appearance === 'dot' ? {
    background: 'transparent',
    color: 'var(--fg-2)',
    padding: 0,
    height: 'auto'
  } : {
    background: t.bg,
    color: t.c
  };
  if (appearance === 'dot') {
    return /*#__PURE__*/React.createElement("span", _extends({
      style: {
        display: 'inline-flex',
        alignItems: 'center',
        gap: 7,
        color: 'var(--fg-2)',
        fontSize: 'var(--text-sm)',
        ...style
      }
    }, rest), /*#__PURE__*/React.createElement("span", {
      style: {
        width: 9,
        height: 9,
        borderRadius: 3,
        background: t.c,
        flex: '0 0 auto'
      }
    }), label || t.word);
  }
  return /*#__PURE__*/React.createElement("span", _extends({
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 6,
      height: 22,
      padding: '0 9px',
      borderRadius: 'var(--radius-xs)',
      fontFamily: 'var(--font-mono)',
      fontSize: 'var(--text-2xs)',
      fontWeight: 'var(--fw-semibold)',
      letterSpacing: 'var(--ls-wide)',
      textTransform: 'uppercase',
      whiteSpace: 'nowrap',
      ...look,
      ...style
    }
  }, rest), label || t.word);
}
Object.assign(__ds_scope, { TierBadge });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/data/TierBadge.jsx", error: String((e && e.message) || e) }); }

// components/forms/Input.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Input — text field with optional leading icon and addon. Used for search,
 * filters, and Reports import paths. Mono variant for IP/MAC/ASN entry.
 */
function Input({
  value,
  onChange = () => {},
  placeholder = '',
  type = 'text',
  iconLeft = null,
  addonRight = null,
  size = 'md',
  mono = false,
  disabled = false,
  invalid = false,
  fullWidth = true,
  style = {},
  inputStyle = {},
  ...rest
}) {
  const h = size === 'sm' ? 30 : 36;
  return /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 8,
      width: fullWidth ? '100%' : 'auto',
      height: h,
      padding: '0 10px',
      background: 'var(--bg-sunken)',
      border: '1px solid ' + (invalid ? 'var(--danger)' : 'var(--hairline-strong)'),
      borderRadius: 'var(--radius-md)',
      opacity: disabled ? 0.5 : 1,
      transition: 'border-color var(--dur-fast) var(--ease-out)',
      ...style
    }
  }, iconLeft && /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      width: 15,
      height: 15,
      color: 'var(--fg-3)',
      flex: '0 0 auto'
    }
  }, iconLeft), /*#__PURE__*/React.createElement("input", _extends({
    value: value,
    onChange: e => onChange(e.target.value, e),
    placeholder: placeholder,
    type: type,
    disabled: disabled,
    style: {
      flex: 1,
      minWidth: 0,
      border: 'none',
      outline: 'none',
      background: 'transparent',
      color: 'var(--fg-1)',
      fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)',
      fontSize: size === 'sm' ? 'var(--text-xs)' : 'var(--text-sm)',
      ...inputStyle
    }
  }, rest)), addonRight && /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      flex: '0 0 auto',
      color: 'var(--fg-3)'
    }
  }, addonRight));
}
Object.assign(__ds_scope, { Input });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Input.jsx", error: String((e && e.message) || e) }); }

// components/forms/SegmentedControl.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * SegmentedControl — a compact single-select used for view switches
 * (Table / List), theme (Dark / Light), and layout engine (Layered / Force /
 * Manual). Options are { value, label, icon? }. Controlled.
 */
function SegmentedControl({
  options = [],
  value,
  onChange = () => {},
  size = 'md',
  fullWidth = false,
  style = {},
  ...rest
}) {
  const h = size === 'sm' ? 28 : 34;
  const fs = size === 'sm' ? 'var(--text-xs)' : 'var(--text-sm)';
  return /*#__PURE__*/React.createElement("div", _extends({
    role: "tablist",
    style: {
      display: 'inline-flex',
      width: fullWidth ? '100%' : 'auto',
      padding: 3,
      gap: 2,
      background: 'var(--surface-3)',
      border: '1px solid var(--hairline)',
      borderRadius: 'var(--radius-md)',
      ...style
    }
  }, rest), options.map(opt => {
    const selected = opt.value === value;
    return /*#__PURE__*/React.createElement("button", {
      key: opt.value,
      type: "button",
      role: "tab",
      "aria-selected": selected,
      onClick: () => onChange(opt.value),
      style: {
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 6,
        flex: fullWidth ? 1 : '0 0 auto',
        height: h,
        padding: '0 12px',
        borderRadius: 'var(--radius-sm)',
        border: 'none',
        cursor: 'pointer',
        fontFamily: 'var(--font-sans)',
        fontSize: fs,
        fontWeight: 'var(--fw-medium)',
        whiteSpace: 'nowrap',
        background: selected ? 'var(--surface-1)' : 'transparent',
        color: selected ? 'var(--fg-1)' : 'var(--fg-3)',
        boxShadow: selected ? 'var(--shadow-xs)' : 'none',
        transition: 'background var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out)'
      }
    }, opt.icon && /*#__PURE__*/React.createElement("span", {
      style: {
        display: 'inline-flex',
        width: '1em',
        height: '1em'
      }
    }, opt.icon), opt.label);
  }));
}
Object.assign(__ds_scope, { SegmentedControl });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/SegmentedControl.jsx", error: String((e && e.message) || e) }); }

// components/forms/Toggle.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/**
 * Toggle — accessible on/off switch. Used in Settings and the topology layer
 * panel. Monochrome track; the accent fills only when on.
 */
function Toggle({
  checked = false,
  onChange = () => {},
  disabled = false,
  label = null,
  description = null,
  size = 'md',
  id,
  style = {},
  ...rest
}) {
  const dims = size === 'sm' ? {
    w: 32,
    h: 18,
    k: 12
  } : {
    w: 38,
    h: 22,
    k: 16
  };
  const pad = (dims.h - dims.k) / 2;
  const sw = /*#__PURE__*/React.createElement("button", {
    type: "button",
    role: "switch",
    "aria-checked": checked,
    "aria-label": typeof label === 'string' ? label : undefined,
    disabled: disabled,
    id: id,
    onClick: () => !disabled && onChange(!checked),
    style: {
      position: 'relative',
      width: dims.w,
      height: dims.h,
      flex: '0 0 auto',
      borderRadius: 'var(--radius-pill)',
      border: '1px solid ' + (checked ? 'var(--accent-base)' : 'var(--hairline-strong)'),
      background: checked ? 'var(--accent-base)' : 'var(--surface-3)',
      cursor: disabled ? 'not-allowed' : 'pointer',
      opacity: disabled ? 0.45 : 1,
      padding: 0,
      transition: 'background var(--dur-base) var(--ease-out), border-color var(--dur-base) var(--ease-out)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      position: 'absolute',
      top: pad,
      left: checked ? dims.w - dims.k - pad - 1 : pad,
      width: dims.k,
      height: dims.k,
      borderRadius: '50%',
      background: checked ? 'var(--fg-on-accent)' : 'var(--fg-2)',
      transition: 'left var(--dur-base) var(--ease-out), background var(--dur-base) var(--ease-out)'
    }
  }));
  if (label == null && description == null) return React.cloneElement(sw, {
    style: {
      ...sw.props.style,
      ...style
    },
    ...rest
  });
  return /*#__PURE__*/React.createElement("label", _extends({
    htmlFor: id,
    style: {
      display: 'flex',
      alignItems: description ? 'flex-start' : 'center',
      gap: 'var(--space-3)',
      cursor: disabled ? 'not-allowed' : 'pointer',
      ...style
    }
  }, rest), sw, /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 2,
      minWidth: 0
    }
  }, label && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-sm)',
      color: 'var(--fg-1)',
      fontWeight: 'var(--fw-medium)'
    }
  }, label), description && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-xs)',
      color: 'var(--fg-3)',
      lineHeight: 1.4
    }
  }, description)));
}
Object.assign(__ds_scope, { Toggle });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Toggle.jsx", error: String((e && e.message) || e) }); }

// ui_kits/console/AppShell.jsx
try { (() => {
/* IAD UI kit — Application shell: sidebar, top status bar, content router,
   right details drawer, bottom log strip. Exposes window.AppShell. */
(function () {
  const {
    useState
  } = React;
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const {
    Badge,
    IconButton,
    Button,
    StatusDot
  } = NS;
  const I = window.Icons;
  const fmt = window.fmt;
  const NAV = [{
    id: 'dashboard',
    label: 'Dashboard',
    icon: 'dashboard'
  }, {
    id: 'topology',
    label: 'Topology',
    icon: 'topology'
  }, {
    id: 'devices',
    label: 'Devices',
    icon: 'devices'
  }, {
    id: 'evidence',
    label: 'Evidence',
    icon: 'evidence'
  }, {
    id: 'reports',
    label: 'Reports',
    icon: 'reports'
  }, {
    id: 'settings',
    label: 'Settings',
    icon: 'settings'
  }];
  function Sidebar({
    active,
    onNav,
    scan
  }) {
    const counts = {
      devices: scan.devices.length,
      evidence: scan.evidence.length
    };
    return React.createElement('aside', {
      style: {
        width: 'var(--sidebar-w)',
        flex: '0 0 auto',
        background: 'var(--ink-900)',
        borderRight: '1px solid var(--hairline)',
        display: 'flex',
        flexDirection: 'column'
      }
    }, React.createElement('div', {
      style: {
        height: 'var(--topbar-h)',
        display: 'flex',
        alignItems: 'center',
        gap: 10,
        padding: '0 16px',
        borderBottom: '1px solid var(--hairline)'
      }
    }, React.createElement('img', {
      src: '../../assets/mark-iad.svg',
      alt: 'IAD',
      style: {
        width: 26,
        height: 26
      }
    }), React.createElement('span', {
      style: {
        fontFamily: 'var(--font-mono)',
        fontWeight: 700,
        letterSpacing: '1.5px',
        color: 'var(--fg-1)',
        fontSize: 15
      }
    }, 'IAD'), React.createElement('span', {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 10,
        color: 'var(--fg-4)',
        marginLeft: 'auto'
      }
    }, 'v0.9')), React.createElement('nav', {
      style: {
        padding: 10,
        display: 'flex',
        flexDirection: 'column',
        gap: 2
      }
    }, NAV.map(n => {
      const on = n.id === active;
      const Icon = I[n.icon];
      return React.createElement('button', {
        key: n.id,
        onClick: () => onNav(n.id),
        style: {
          display: 'flex',
          alignItems: 'center',
          gap: 11,
          padding: '9px 11px',
          borderRadius: 'var(--radius-md)',
          border: '1px solid ' + (on ? 'var(--hairline)' : 'transparent'),
          background: on ? 'var(--surface-2)' : 'transparent',
          cursor: 'pointer',
          textAlign: 'left',
          color: on ? 'var(--fg-1)' : 'var(--fg-3)',
          font: '500 var(--text-sm) var(--font-sans)',
          transition: 'background var(--dur-fast) var(--ease-out), color var(--dur-fast) var(--ease-out)'
        }
      }, React.createElement('span', {
        style: {
          display: 'inline-flex',
          color: on ? 'var(--accent-bright)' : 'var(--fg-4)'
        }
      }, React.createElement(Icon, {
        size: 17
      })), n.label, counts[n.id] != null && React.createElement('span', {
        style: {
          marginLeft: 'auto',
          fontFamily: 'var(--font-mono)',
          fontSize: 10,
          color: 'var(--fg-4)'
        }
      }, counts[n.id]));
    })), React.createElement('div', {
      style: {
        marginTop: 'auto',
        padding: 14,
        borderTop: '1px solid var(--hairline)'
      }
    }, React.createElement('div', {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        padding: '8px 10px',
        background: 'var(--ok-bg)',
        borderRadius: 'var(--radius-md)'
      }
    }, React.createElement(I.shield, {
      size: 15,
      style: {
        color: 'var(--ok)'
      }
    }), React.createElement('span', {
      style: {
        fontSize: 11,
        color: 'var(--ok)',
        fontWeight: 600
      }
    }, 'Safe mode · read-only'))));
  }
  function TopMetric({
    label,
    value,
    mono = true,
    tone,
    minW = 88,
    maxW
  }) {
    return React.createElement('div', {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 1,
        minWidth: minW,
        maxWidth: maxW,
        flex: '0 0 auto'
      }
    }, React.createElement('span', {
      style: {
        font: 'var(--type-overline)',
        letterSpacing: 'var(--ls-caps)',
        textTransform: 'uppercase',
        color: 'var(--fg-4)',
        fontSize: 9
      }
    }, label), React.createElement('span', {
      style: {
        fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)',
        fontSize: 12,
        fontWeight: 600,
        color: tone || 'var(--fg-1)',
        whiteSpace: 'nowrap',
        overflow: 'hidden',
        textOverflow: 'ellipsis'
      }
    }, value));
  }
  function TopStatusBar({
    scan,
    onImport,
    onExport
  }) {
    const c = scan.confidence;
    return React.createElement('header', {
      style: {
        height: 'var(--topbar-h)',
        flex: '0 0 auto',
        background: 'var(--ink-850)',
        borderBottom: '1px solid var(--hairline)',
        display: 'flex',
        alignItems: 'center',
        gap: 16,
        padding: '0 16px',
        overflow: 'hidden'
      }
    }, React.createElement('div', {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 9
      }
    }, React.createElement('div', {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 1
      }
    }, React.createElement('span', {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 9,
        color: 'var(--fg-4)',
        textTransform: 'uppercase',
        letterSpacing: '.1em'
      }
    }, 'Scan'), React.createElement('span', {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 12,
        color: 'var(--fg-1)',
        fontWeight: 600
      }
    }, scan.scan_id)), React.createElement(Badge, {
      tone: 'neutral',
      appearance: 'outline',
      size: 'sm',
      mono: true
    }, scan.mode)), React.createElement('div', {
      style: {
        width: 1,
        height: 26,
        background: 'var(--hairline)'
      }
    }), React.createElement(TopMetric, {
      label: 'Time',
      value: fmt.time(scan.created_at),
      mono: false,
      minW: 96
    }), React.createElement(TopMetric, {
      label: 'Interface',
      value: scan.detected_network_context.selected_interface.name,
      minW: 80
    }), React.createElement(TopMetric, {
      label: 'Public IP',
      value: scan.public_ip.address,
      minW: 110
    }), React.createElement(TopMetric, {
      label: 'ISP',
      value: scan.public_ip.asn + ' · ' + scan.public_ip.org,
      mono: false,
      minW: 120,
      maxW: 190
    }), React.createElement('div', {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 7,
        marginLeft: 'auto'
      }
    }, React.createElement('div', {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 1,
        alignItems: 'flex-end'
      }
    }, React.createElement('span', {
      style: {
        font: 'var(--type-overline)',
        letterSpacing: 'var(--ls-caps)',
        textTransform: 'uppercase',
        color: 'var(--fg-4)',
        fontSize: 9
      }
    }, 'Confidence'), React.createElement('div', {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 6
      }
    }, React.createElement('span', {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 13,
        fontWeight: 700,
        color: fmt.bandColor(c)
      }
    }, fmt.pct(c)), React.createElement('span', {
      style: {
        fontSize: 9,
        fontWeight: 700,
        textTransform: 'uppercase',
        letterSpacing: '.08em',
        color: fmt.bandColor(c)
      }
    }, fmt.bandWord(c)))), React.createElement(Badge, {
      tone: scan.decision_quality === 'high' ? 'success' : scan.decision_quality === 'medium' ? 'warn' : 'neutral',
      size: 'sm',
      uppercase: true,
      mono: true
    }, scan.decision_quality + ' quality')), React.createElement('div', {
      style: {
        width: 1,
        height: 26,
        background: 'var(--hairline)'
      }
    }), React.createElement('div', {
      style: {
        display: 'flex',
        gap: 8
      }
    }, React.createElement(Button, {
      size: 'sm',
      variant: 'secondary',
      iconLeft: React.createElement(I.upload, {
        size: 14
      }),
      onClick: onImport
    }, 'Import'), React.createElement(Button, {
      size: 'sm',
      variant: 'primary',
      iconLeft: React.createElement(I.download, {
        size: 14
      }),
      onClick: onExport
    }, 'Export')));
  }
  function LogStrip({
    lines
  }) {
    return React.createElement('footer', {
      style: {
        height: 28,
        flex: '0 0 auto',
        background: 'var(--ink-900)',
        borderTop: '1px solid var(--hairline)',
        display: 'flex',
        alignItems: 'center',
        gap: 16,
        padding: '0 16px',
        overflow: 'hidden'
      }
    }, React.createElement(StatusDot, {
      tone: 'success',
      size: 7
    }), React.createElement('span', {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 11,
        color: 'var(--fg-3)',
        whiteSpace: 'nowrap'
      }
    }, lines));
  }
  function AppShell({
    scan
  }) {
    const [active, setActive] = useState('dashboard');
    const [drawer, setDrawer] = useState(null); // {kind, payload}
    const Screens = window.IAD_SCREENS || {};
    const Screen = Screens[active];
    const openDrawer = (kind, payload) => setDrawer({
      kind,
      payload
    });
    const closeDrawer = () => setDrawer(null);
    const log = {
      dashboard: '09:41:22  scan complete · 10 probes · 8 devices · 6.84s',
      topology: '09:41:22  topology generated from context · 6 nodes · 7 edges · inferred where dashed',
      devices: '09:41:17  arp_sweep resolved 8 hosts on 192.168.1.0/24',
      evidence: '09:41:20  cpe_snmp blocked (timeout) · physical medium remains inferred',
      reports: 'ready · export NormalizedScanReport v1',
      settings: 'UI layout positions stored locally · scan data immutable'
    }[active];
    return React.createElement('div', {
      style: {
        display: 'flex',
        height: '100%',
        width: '100%',
        background: 'var(--bg-app)'
      }
    }, React.createElement(Sidebar, {
      active,
      onNav: setActive,
      scan
    }), React.createElement('div', {
      style: {
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        minWidth: 0
      }
    }, React.createElement(TopStatusBar, {
      scan,
      onImport: () => setActive('reports'),
      onExport: () => setActive('reports')
    }), React.createElement('div', {
      style: {
        flex: 1,
        display: 'flex',
        minHeight: 0
      }
    }, React.createElement('main', {
      style: {
        flex: 1,
        minWidth: 0,
        overflow: 'auto'
      }
    }, Screen ? React.createElement(Screen, {
      scan,
      openDrawer,
      drawer
    }) : React.createElement('div', {
      style: {
        padding: 40,
        color: 'var(--fg-3)'
      }
    }, 'Screen not loaded')), drawer && window.DetailsDrawer && React.createElement(window.DetailsDrawer, {
      drawer,
      onClose: closeDrawer,
      scan
    })), React.createElement(LogStrip, {
      lines: log
    })));
  }
  window.AppShell = AppShell;
})();
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/console/AppShell.jsx", error: String((e && e.message) || e) }); }

// ui_kits/console/Dashboard.jsx
try { (() => {
/* IAD UI kit — Dashboard screen. Registers window.IAD_SCREENS.dashboard */
(function () {
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const {
    Card,
    MetricStat,
    ConfidenceBar,
    Badge,
    TierBadge,
    ProbeStatusBadge,
    StatusDot
  } = NS;
  const I = window.Icons;
  const fmt = window.fmt;
  window.IAD_SCREENS = window.IAD_SCREENS || {};
  const Eyebrow = ({
    children
  }) => /*#__PURE__*/React.createElement("div", {
    style: {
      font: 'var(--type-overline)',
      letterSpacing: 'var(--ls-caps)',
      textTransform: 'uppercase',
      color: 'var(--fg-3)',
      marginBottom: 8
    }
  }, children);
  function VerdictCard({
    scan
  }) {
    const c = scan.classification_confidence;
    return /*#__PURE__*/React.createElement("div", {
      style: {
        gridColumn: 'span 2',
        background: 'var(--surface-card)',
        border: '1px solid var(--hairline)',
        borderRadius: 'var(--radius-lg)',
        padding: 24,
        display: 'flex',
        flexDirection: 'column',
        gap: 18,
        position: 'relative',
        overflow: 'hidden'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        position: 'absolute',
        inset: 0,
        backgroundImage: 'radial-gradient(var(--grid-line) 1px, transparent 1px)',
        backgroundSize: '22px 22px',
        opacity: 0.6,
        pointerEvents: 'none'
      }
    }), /*#__PURE__*/React.createElement("div", {
      style: {
        position: 'relative',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
        gap: 20
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        flex: 1,
        minWidth: 0
      }
    }, /*#__PURE__*/React.createElement(Eyebrow, null, "Access type decision"), /*#__PURE__*/React.createElement("div", {
      style: {
        font: 'var(--type-verdict)',
        color: 'var(--fg-1)',
        letterSpacing: 'var(--ls-tight)'
      }
    }, scan.primary_type), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        gap: 8,
        marginTop: 10,
        alignItems: 'center',
        flexWrap: 'wrap'
      }
    }, /*#__PURE__*/React.createElement(Badge, {
      tone: "neutral",
      appearance: "outline",
      mono: true,
      size: "sm"
    }, scan.category), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)'
      }
    }, "estimated \xB7 physical medium not directly confirmed"))), /*#__PURE__*/React.createElement("div", {
      style: {
        width: 200,
        flex: '0 0 auto'
      }
    }, /*#__PURE__*/React.createElement(ConfidenceBar, {
      value: c,
      label: "Classification"
    }), /*#__PURE__*/React.createElement("div", {
      style: {
        marginTop: 12
      }
    }, /*#__PURE__*/React.createElement(ConfidenceBar, {
      value: scan.context_confidence,
      label: "Network context"
    })))), /*#__PURE__*/React.createElement("div", {
      style: {
        position: 'relative',
        display: 'flex',
        gap: 8,
        flexWrap: 'wrap',
        borderTop: '1px solid var(--hairline)',
        paddingTop: 16
      }
    }, scan.candidates.map((c, i) => /*#__PURE__*/React.createElement("div", {
      key: i,
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        padding: '6px 10px',
        background: i === 0 ? 'var(--accent-ghost)' : 'var(--bg-sunken)',
        border: '1px solid ' + (i === 0 ? 'var(--accent-ring)' : 'var(--hairline)'),
        borderRadius: 'var(--radius-md)'
      }
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-xs)',
        fontWeight: 600,
        color: i === 0 ? 'var(--accent-bright)' : 'var(--fg-2)'
      }
    }, c.type), /*#__PURE__*/React.createElement("span", {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 11,
        color: fmt.bandColor(c.score)
      }
    }, fmt.pct(c.score))))));
  }
  function UncertaintyCard({
    scan
  }) {
    return /*#__PURE__*/React.createElement(Card, {
      eyebrow: "Why not certain",
      title: "Uncertainty",
      style: {
        gridColumn: 'span 2'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 11
      }
    }, scan.uncertainty_reasons.map((r, i) => /*#__PURE__*/React.createElement("div", {
      key: i,
      style: {
        display: 'flex',
        gap: 10,
        alignItems: 'flex-start'
      }
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        color: 'var(--warn)',
        flex: '0 0 auto',
        marginTop: 1
      }
    }, /*#__PURE__*/React.createElement(I.info, {
      size: 15
    })), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-2)',
        lineHeight: 1.45
      }
    }, r)))));
  }
  function GatewayChainCard({
    scan
  }) {
    return /*#__PURE__*/React.createElement(Card, {
      eyebrow: "Path to ISP",
      title: "Gateway chain",
      style: {
        gridColumn: 'span 2'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column'
      }
    }, scan.gateway_chain.map((g, i) => /*#__PURE__*/React.createElement("div", {
      key: i,
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '9px 0',
        borderBottom: i < scan.gateway_chain.length - 1 ? '1px solid var(--hairline)' : 'none'
      }
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 11,
        color: 'var(--fg-4)',
        width: 18
      }
    }, g.hop), /*#__PURE__*/React.createElement("span", {
      style: {
        width: 8,
        height: 8,
        borderRadius: '50%',
        background: g.private ? 'var(--fg-4)' : 'var(--tier-isp)',
        flex: '0 0 auto'
      }
    }), /*#__PURE__*/React.createElement("span", {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-1)',
        minWidth: 116
      }
    }, g.ip), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-2)'
      }
    }, g.label), g.note && /*#__PURE__*/React.createElement(Badge, {
      tone: "warn",
      size: "sm"
    }, g.note), /*#__PURE__*/React.createElement("span", {
      style: {
        marginLeft: 'auto',
        fontFamily: 'var(--font-mono)',
        fontSize: 11,
        color: 'var(--fg-3)'
      }
    }, g.rtt_ms, " ms")))));
  }
  function ConfidenceBreakdownCard({
    scan
  }) {
    return /*#__PURE__*/React.createElement(Card, {
      eyebrow: "What moved the needle",
      title: "Confidence breakdown",
      style: {
        gridColumn: 'span 2'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 13
      }
    }, scan.confidence_breakdown.map((f, i) => {
      const up = f.direction === 'up';
      const mag = Math.min(1, Math.abs(f.contribution) / 0.3);
      return /*#__PURE__*/React.createElement("div", {
        key: i,
        style: {
          display: 'flex',
          flexDirection: 'column',
          gap: 5
        }
      }, /*#__PURE__*/React.createElement("div", {
        style: {
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          gap: 8
        }
      }, /*#__PURE__*/React.createElement("span", {
        style: {
          fontSize: 'var(--text-sm)',
          color: 'var(--fg-1)',
          fontWeight: 500
        }
      }, f.factor), /*#__PURE__*/React.createElement("span", {
        style: {
          fontFamily: 'var(--font-mono)',
          fontSize: 12,
          color: up ? 'var(--ok)' : 'var(--danger)',
          fontWeight: 600
        }
      }, up ? '+' : '−', Math.abs(f.contribution).toFixed(2))), /*#__PURE__*/React.createElement("div", {
        style: {
          height: 4,
          borderRadius: 'var(--radius-pill)',
          background: 'var(--surface-3)',
          overflow: 'hidden',
          display: 'flex',
          justifyContent: up ? 'flex-start' : 'flex-end'
        }
      }, /*#__PURE__*/React.createElement("div", {
        style: {
          width: mag * 100 + '%',
          height: '100%',
          background: up ? 'var(--ok)' : 'var(--danger)',
          borderRadius: 'var(--radius-pill)'
        }
      })), /*#__PURE__*/React.createElement("span", {
        style: {
          fontSize: 'var(--text-xs)',
          color: 'var(--fg-3)',
          lineHeight: 1.4
        }
      }, f.detail));
    })));
  }
  function NextProbesCard({
    scan
  }) {
    return /*#__PURE__*/React.createElement(Card, {
      eyebrow: "Raise confidence",
      title: "Next best probes",
      style: {
        gridColumn: 'span 2'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 10
      }
    }, scan.next_best_probes.map((p, i) => /*#__PURE__*/React.createElement("div", {
      key: i,
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: 12,
        background: 'var(--bg-sunken)',
        border: '1px solid var(--hairline)',
        borderRadius: 'var(--radius-md)'
      }
    }, /*#__PURE__*/React.createElement(TierBadge, {
      tier: p.tier,
      appearance: "dot"
    }), /*#__PURE__*/React.createElement("div", {
      style: {
        minWidth: 0,
        flex: 1
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-1)',
        fontWeight: 500
      }
    }, p.name), /*#__PURE__*/React.createElement("div", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)'
      }
    }, "requires: ", p.requires)), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'flex-end',
        gap: 1
      }
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 13,
        color: 'var(--ok)',
        fontWeight: 600
      }
    }, "+", fmt.pct(p.gain)), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 9,
        color: 'var(--fg-4)',
        textTransform: 'uppercase',
        letterSpacing: '.08em'
      }
    }, "est. gain"))))));
  }
  function MiniCard({
    eyebrow,
    children
  }) {
    return /*#__PURE__*/React.createElement(Card, {
      padding: "md"
    }, /*#__PURE__*/React.createElement(Eyebrow, null, eyebrow), children);
  }
  function Dashboard({
    scan
  }) {
    const ni = scan.detected_network_context.selected_interface;
    const p = scan.performance;
    return /*#__PURE__*/React.createElement("div", {
      style: {
        padding: 22,
        display: 'flex',
        flexDirection: 'column',
        gap: 16,
        maxWidth: 1180,
        margin: '0 auto'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'baseline',
        justifyContent: 'space-between'
      }
    }, /*#__PURE__*/React.createElement("h1", {
      style: {
        font: 'var(--type-h1)'
      }
    }, "Overview"), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)'
      }
    }, "Scanned ", fmt.ago(scan.created_at), " \xB7 ", scan.duration_ms / 1000, "s \xB7 ", scan.evidence.length, " probes")), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'grid',
        gridTemplateColumns: 'repeat(4, 1fr)',
        gap: 16
      }
    }, /*#__PURE__*/React.createElement(VerdictCard, {
      scan: scan
    }), /*#__PURE__*/React.createElement(UncertaintyCard, {
      scan: scan
    }), /*#__PURE__*/React.createElement(MiniCard, {
      eyebrow: "Local interface"
    }, /*#__PURE__*/React.createElement(MetricStat, {
      label: ni.name + ' · IPv4',
      value: ni.ipv4,
      secondary: '/' + ni.prefix + ' · ' + ni.link_speed_mbps + ' Mbps link · DHCP'
    }), /*#__PURE__*/React.createElement("div", {
      style: {
        marginTop: 10,
        fontFamily: 'var(--font-mono)',
        fontSize: 11,
        color: 'var(--fg-3)'
      }
    }, ni.mac)), /*#__PURE__*/React.createElement(MiniCard, {
      eyebrow: "Public IP"
    }, /*#__PURE__*/React.createElement(MetricStat, {
      label: "Address",
      value: scan.public_ip.address,
      secondary: 'PTR ✓ ' + scan.public_ip.ptr
    })), /*#__PURE__*/React.createElement(MiniCard, {
      eyebrow: "ISP / ASN"
    }, /*#__PURE__*/React.createElement(MetricStat, {
      label: scan.public_ip.asn,
      value: scan.public_ip.org,
      secondary: scan.public_ip.city + ', ' + scan.public_ip.country
    })), /*#__PURE__*/React.createElement(MiniCard, {
      eyebrow: "Throughput"
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        gap: 18
      }
    }, /*#__PURE__*/React.createElement(MetricStat, {
      label: "Down",
      value: p.downstream_mbps,
      unit: "Mbps"
    }), /*#__PURE__*/React.createElement(MetricStat, {
      label: "Up",
      value: p.upstream_mbps,
      unit: "Mbps"
    })), /*#__PURE__*/React.createElement("div", {
      style: {
        marginTop: 8,
        fontFamily: 'var(--font-mono)',
        fontSize: 11,
        color: 'var(--fg-3)'
      }
    }, p.latency_ms, " ms \xB7 jitter ", p.jitter_ms, " ms \xB7 loss ", p.loss_pct, "%")), /*#__PURE__*/React.createElement(MiniCard, {
      eyebrow: "IPv4 NAT status"
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 8
      }
    }, /*#__PURE__*/React.createElement(Badge, {
      tone: "warn",
      appearance: "solid"
    }, "CGNAT"), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-2)'
      }
    }, scan.nat_topology.layers, " layers")), /*#__PURE__*/React.createElement("div", {
      style: {
        marginTop: 8,
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)',
        lineHeight: 1.4
      }
    }, scan.nat_topology.note)), /*#__PURE__*/React.createElement(MiniCard, {
      eyebrow: "IPv6 status"
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 8
      }
    }, /*#__PURE__*/React.createElement(StatusDot, {
      tone: "success"
    }), /*#__PURE__*/React.createElement("span", {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-1)'
      }
    }, scan.ipv6_context.global_address)), /*#__PURE__*/React.createElement("div", {
      style: {
        marginTop: 8,
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)',
        lineHeight: 1.4
      }
    }, scan.ipv6_context.note)), /*#__PURE__*/React.createElement(MiniCard, {
      eyebrow: "Network status"
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 7
      }
    }, /*#__PURE__*/React.createElement(StatusDot, {
      tone: "success",
      label: "Internet reachable (IPv4 + IPv6)"
    }), /*#__PURE__*/React.createElement(StatusDot, {
      tone: "warn",
      label: "Inbound blocked by CGNAT"
    }), /*#__PURE__*/React.createElement(StatusDot, {
      tone: "success",
      label: scan.devices.length + ' LAN devices discovered'
    }))), /*#__PURE__*/React.createElement(GatewayChainCard, {
      scan: scan
    }), /*#__PURE__*/React.createElement(ConfidenceBreakdownCard, {
      scan: scan
    }), /*#__PURE__*/React.createElement(NextProbesCard, {
      scan: scan
    }), /*#__PURE__*/React.createElement(Card, {
      eyebrow: "Advisories",
      title: "Warnings",
      style: {
        gridColumn: 'span 2'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 9
      }
    }, scan.warnings.map((w, i) => /*#__PURE__*/React.createElement("div", {
      key: i,
      style: {
        display: 'flex',
        gap: 10,
        alignItems: 'flex-start',
        padding: 11,
        background: w.level === 'warn' ? 'var(--warn-bg)' : 'var(--info-bg)',
        borderRadius: 'var(--radius-md)'
      }
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        color: w.level === 'warn' ? 'var(--warn)' : 'var(--info)',
        flex: '0 0 auto'
      }
    }, w.level === 'warn' ? /*#__PURE__*/React.createElement(I.alert, {
      size: 15
    }) : /*#__PURE__*/React.createElement(I.info, {
      size: 15
    })), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-2)',
        lineHeight: 1.45
      }
    }, w.text)))))));
  }
  window.IAD_SCREENS.dashboard = Dashboard;
})();
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/console/Dashboard.jsx", error: String((e && e.message) || e) }); }

// ui_kits/console/DetailsDrawer.jsx
try { (() => {
/* IAD UI kit — right-side details drawer. window.DetailsDrawer
   Renders device / node / edge / probe detail. Read-only: no delete actions. */
(function () {
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const {
    Badge,
    IconButton,
    ConfidenceBar,
    ProbeStatusBadge,
    TierBadge,
    StatusDot
  } = NS;
  const I = window.Icons;
  const fmt = window.fmt;
  const Row = ({
    k,
    v,
    mono = true
  }) => /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      justifyContent: 'space-between',
      gap: 14,
      padding: '8px 0',
      borderBottom: '1px solid var(--hairline)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      font: 'var(--type-overline)',
      letterSpacing: 'var(--ls-caps)',
      textTransform: 'uppercase',
      color: 'var(--fg-4)',
      whiteSpace: 'nowrap'
    }
  }, k), /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)',
      fontSize: 'var(--text-sm)',
      color: 'var(--fg-1)',
      textAlign: 'right',
      wordBreak: 'break-word'
    }
  }, v));
  const Section = ({
    title,
    children
  }) => /*#__PURE__*/React.createElement("div", {
    style: {
      marginTop: 18
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      font: 'var(--type-overline)',
      letterSpacing: 'var(--ls-caps)',
      textTransform: 'uppercase',
      color: 'var(--fg-3)',
      marginBottom: 8
    }
  }, title), children);
  function DeviceBody({
    d,
    scan
  }) {
    const r = fmt.reach(d.reachability);
    const Icon = I[fmt.deviceIcon(d.type)];
    const ev = scan.evidence.filter(e => e.evidence_class === 'l2' || e.evidence_class === 'l3').slice(0, 3);
    return /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        marginBottom: 4
      }
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        width: 40,
        height: 40,
        borderRadius: 'var(--radius-md)',
        background: 'var(--surface-3)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: 'var(--fg-2)',
        flex: '0 0 auto'
      }
    }, /*#__PURE__*/React.createElement(Icon, {
      size: 20
    })), /*#__PURE__*/React.createElement("div", {
      style: {
        minWidth: 0
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        font: 'var(--type-h3)',
        color: 'var(--fg-1)'
      }
    }, d.hostname !== '—' ? d.hostname : d.ip), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        gap: 6,
        marginTop: 4
      }
    }, /*#__PURE__*/React.createElement(Badge, {
      tone: r.tone,
      size: "sm"
    }, r.word), /*#__PURE__*/React.createElement(Badge, {
      tone: "neutral",
      appearance: "outline",
      mono: true,
      size: "sm"
    }, d.type)))), /*#__PURE__*/React.createElement(Section, {
      title: "Confidence"
    }, /*#__PURE__*/React.createElement(ConfidenceBar, {
      value: d.confidence,
      label: "Identification"
    })), /*#__PURE__*/React.createElement(Section, {
      title: "Addresses"
    }, /*#__PURE__*/React.createElement(Row, {
      k: "IPv4",
      v: d.ip
    }), /*#__PURE__*/React.createElement(Row, {
      k: "MAC",
      v: d.mac
    }), /*#__PURE__*/React.createElement(Row, {
      k: "Vendor",
      v: d.vendor,
      mono: false
    }), /*#__PURE__*/React.createElement(Row, {
      k: "Hostname",
      v: d.hostname
    })), /*#__PURE__*/React.createElement(Section, {
      title: "Role & services"
    }, /*#__PURE__*/React.createElement(Row, {
      k: "Role",
      v: d.role,
      mono: false
    }), /*#__PURE__*/React.createElement(Row, {
      k: "Source probe",
      v: d.source
    }), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        gap: 6,
        flexWrap: 'wrap',
        marginTop: 10
      }
    }, d.services.length && d.services[0] !== '—' ? d.services.map((s, i) => /*#__PURE__*/React.createElement(Badge, {
      key: i,
      tone: "neutral",
      mono: true,
      size: "sm"
    }, s)) : /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-4)'
      }
    }, "No services fingerprinted"))), /*#__PURE__*/React.createElement(Section, {
      title: "Evidence"
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 7
      }
    }, ev.map(e => /*#__PURE__*/React.createElement("div", {
      key: e.id,
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        padding: '7px 9px',
        background: 'var(--bg-sunken)',
        borderRadius: 'var(--radius-sm)'
      }
    }, /*#__PURE__*/React.createElement(ProbeStatusBadge, {
      status: e.status,
      size: "sm"
    }), /*#__PURE__*/React.createElement("span", {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 11,
        color: 'var(--fg-2)'
      }
    }, e.probe_name))))), /*#__PURE__*/React.createElement(Section, {
      title: "Limitations"
    }, /*#__PURE__*/React.createElement("p", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)',
        lineHeight: 1.5
      }
    }, "Identity inferred from ", d.source.toUpperCase(), " responses on the local broadcast domain. Device role is a best-effort classification and may be wrong for multi-purpose or virtualized hosts.")));
  }
  function ProbeBody({
    e
  }) {
    return /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: 6
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        font: 'var(--type-h3)',
        color: 'var(--fg-1)',
        fontFamily: 'var(--font-mono)'
      }
    }, e.probe_name), /*#__PURE__*/React.createElement(ProbeStatusBadge, {
      status: e.status
    })), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        gap: 6
      }
    }, /*#__PURE__*/React.createElement(TierBadge, {
      tier: e.evidence_class
    }), /*#__PURE__*/React.createElement(Badge, {
      tone: "neutral",
      appearance: "outline",
      mono: true,
      size: "sm"
    }, e.ts)), e.status !== 'no_data' && e.status !== 'skipped' && e.status !== 'blocked' && e.status !== 'failed' && /*#__PURE__*/React.createElement(Section, {
      title: "Confidence"
    }, /*#__PURE__*/React.createElement(ConfidenceBar, {
      value: e.confidence,
      label: "Probe confidence"
    })), /*#__PURE__*/React.createElement(Section, {
      title: "Reason"
    }, /*#__PURE__*/React.createElement("p", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-2)',
        lineHeight: 1.5
      }
    }, e.reason)), e.limitations && /*#__PURE__*/React.createElement(Section, {
      title: "Limitations"
    }, /*#__PURE__*/React.createElement("p", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)',
        lineHeight: 1.5
      }
    }, e.limitations)), /*#__PURE__*/React.createElement(Section, {
      title: "Raw evidence"
    }, /*#__PURE__*/React.createElement("pre", {
      style: {
        margin: 0,
        padding: 12,
        background: 'var(--ink-900)',
        border: '1px solid var(--hairline)',
        borderRadius: 'var(--radius-md)',
        fontFamily: 'var(--font-mono)',
        fontSize: 11,
        color: 'var(--fg-2)',
        lineHeight: 1.5,
        overflow: 'auto',
        whiteSpace: 'pre-wrap'
      }
    }, JSON.stringify(e.data, null, 2))));
  }
  function EdgeBody({
    edge
  }) {
    return /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
      style: {
        font: 'var(--type-h3)',
        color: 'var(--fg-1)',
        marginBottom: 6
      }
    }, edge.label), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        gap: 6
      }
    }, /*#__PURE__*/React.createElement(Badge, {
      tone: edge.kind === 'confirmed' ? 'success' : edge.kind === 'inferred' ? 'warn' : 'neutral',
      size: "sm",
      uppercase: true,
      mono: true
    }, edge.kind), edge.tier && /*#__PURE__*/React.createElement(TierBadge, {
      tier: edge.tier
    })), /*#__PURE__*/React.createElement(Section, {
      title: "Relationship"
    }, /*#__PURE__*/React.createElement(Row, {
      k: "From",
      v: edge.fromLabel,
      mono: false
    }), /*#__PURE__*/React.createElement(Row, {
      k: "To",
      v: edge.toLabel,
      mono: false
    }), /*#__PURE__*/React.createElement(Row, {
      k: "Type",
      v: edge.type
    })), /*#__PURE__*/React.createElement(Section, {
      title: "Basis"
    }, /*#__PURE__*/React.createElement("p", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-2)',
        lineHeight: 1.5
      }
    }, edge.basis)), edge.kind !== 'confirmed' && /*#__PURE__*/React.createElement(Section, {
      title: "Caveat"
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        gap: 9,
        padding: 11,
        background: 'var(--warn-bg)',
        borderRadius: 'var(--radius-md)'
      }
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        color: 'var(--warn)',
        flex: '0 0 auto'
      }
    }, /*#__PURE__*/React.createElement(I.alert, {
      size: 15
    })), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-2)',
        lineHeight: 1.45
      }
    }, "This link is inferred, not confirmed. The physical path may differ \u2014 traceroute hops are L3 routers, not switches."))));
  }
  function DetailsDrawer({
    drawer,
    onClose,
    scan
  }) {
    let title = 'Details';
    let body = null;
    if (drawer.kind === 'device') {
      title = 'Device';
      body = /*#__PURE__*/React.createElement(DeviceBody, {
        d: drawer.payload,
        scan: scan
      });
    } else if (drawer.kind === 'probe') {
      title = 'Probe';
      body = /*#__PURE__*/React.createElement(ProbeBody, {
        e: drawer.payload
      });
    } else if (drawer.kind === 'node') {
      title = drawer.payload.device ? 'Device' : 'Node';
      body = drawer.payload.device ? /*#__PURE__*/React.createElement(DeviceBody, {
        d: drawer.payload.device,
        scan: scan
      }) : /*#__PURE__*/React.createElement(ProbeBody, {
        e: drawer.payload
      });
    } else if (drawer.kind === 'edge') {
      title = 'Edge evidence';
      body = /*#__PURE__*/React.createElement(EdgeBody, {
        edge: drawer.payload
      });
    }
    return /*#__PURE__*/React.createElement("aside", {
      style: {
        width: 360,
        flex: '0 0 auto',
        background: 'var(--surface-card)',
        borderLeft: '1px solid var(--hairline)',
        display: 'flex',
        flexDirection: 'column',
        minHeight: 0
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        height: 'var(--topbar-h)',
        flex: '0 0 auto',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0 16px',
        borderBottom: '1px solid var(--hairline)'
      }
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        font: 'var(--type-overline)',
        letterSpacing: 'var(--ls-caps)',
        textTransform: 'uppercase',
        color: 'var(--fg-3)'
      }
    }, title), /*#__PURE__*/React.createElement(IconButton, {
      label: "Close",
      onClick: onClose
    }, /*#__PURE__*/React.createElement(I.close, {
      size: 16
    }))), /*#__PURE__*/React.createElement("div", {
      style: {
        flex: 1,
        overflow: 'auto',
        padding: 18
      }
    }, body));
  }
  window.DetailsDrawer = DetailsDrawer;
})();
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/console/DetailsDrawer.jsx", error: String((e && e.message) || e) }); }

// ui_kits/console/Devices.jsx
try { (() => {
/* IAD UI kit — Devices screen. Inventory with search / filters / sort,
   table + list views, click to open details. window.IAD_SCREENS.devices */
(function () {
  const {
    useState,
    useMemo
  } = React;
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const {
    Input,
    SegmentedControl,
    Badge,
    ConfidenceBar,
    StatusDot
  } = NS;
  const I = window.Icons;
  const fmt = window.fmt;
  const TYPES = ['all', 'default_gateway', 'access_point', 'server', 'printer', 'mobile', 'iot', 'unknown'];
  const REACH = ['all', 'reachable', 'partial', 'unreachable'];
  function ConfPill({
    v
  }) {
    return /*#__PURE__*/React.createElement("span", {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 12,
        fontWeight: 600,
        color: fmt.bandColor(v)
      }
    }, fmt.pct(v));
  }
  function Devices({
    scan,
    openDrawer,
    drawer
  }) {
    const [q, setQ] = useState('');
    const [type, setType] = useState('all');
    const [reach, setReach] = useState('all');
    const [view, setView] = useState('table');
    const [sort, setSort] = useState({
      k: 'ip',
      dir: 1
    });
    const selId = drawer && drawer.kind === 'device' ? drawer.payload.id : null;
    const rows = useMemo(() => {
      let r = scan.devices.filter(d => {
        if (type !== 'all' && d.type !== type) return false;
        if (reach !== 'all' && d.reachability !== reach && !(reach === 'reachable' && d.reachability === 'self')) return false;
        if (q) {
          const s = (d.ip + d.hostname + d.vendor + d.mac + d.role).toLowerCase();
          if (!s.includes(q.toLowerCase())) return false;
        }
        return true;
      });
      const ipNum = ip => ip.split('.').reduce((a, o) => a * 256 + (+o || 0), 0);
      r = [...r].sort((a, b) => {
        let av, bv;
        if (sort.k === 'ip') {
          av = ipNum(a.ip);
          bv = ipNum(b.ip);
        } else if (sort.k === 'confidence') {
          av = a.confidence;
          bv = b.confidence;
        } else {
          av = String(a[sort.k]);
          bv = String(b[sort.k]);
          return av.localeCompare(bv) * sort.dir;
        }
        return (av - bv) * sort.dir;
      });
      return r;
    }, [scan, q, type, reach, sort]);
    const toggleSort = k => setSort(s => s.k === k ? {
      k,
      dir: -s.dir
    } : {
      k,
      dir: 1
    });
    const Th = ({
      k,
      children,
      w
    }) => /*#__PURE__*/React.createElement("th", {
      onClick: () => toggleSort(k),
      style: {
        textAlign: 'left',
        padding: '0 12px',
        height: 34,
        font: 'var(--type-overline)',
        letterSpacing: 'var(--ls-caps)',
        textTransform: 'uppercase',
        color: 'var(--fg-4)',
        cursor: 'pointer',
        whiteSpace: 'nowrap',
        width: w,
        userSelect: 'none'
      }
    }, children, sort.k === k && /*#__PURE__*/React.createElement("span", {
      style: {
        marginLeft: 4,
        color: 'var(--accent-bright)'
      }
    }, sort.dir > 0 ? '↑' : '↓'));
    const Select = ({
      value,
      onChange,
      options,
      label
    }) => /*#__PURE__*/React.createElement("label", {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 7
      }
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        font: 'var(--type-overline)',
        letterSpacing: 'var(--ls-caps)',
        textTransform: 'uppercase',
        color: 'var(--fg-4)'
      }
    }, label), /*#__PURE__*/React.createElement("select", {
      value: value,
      onChange: e => onChange(e.target.value),
      style: {
        height: 30,
        background: 'var(--bg-sunken)',
        color: 'var(--fg-1)',
        border: '1px solid var(--hairline-strong)',
        borderRadius: 'var(--radius-sm)',
        padding: '0 8px',
        fontFamily: 'var(--font-sans)',
        fontSize: 'var(--text-xs)'
      }
    }, options.map(o => /*#__PURE__*/React.createElement("option", {
      key: o,
      value: o
    }, o === 'all' ? 'All' : o.replace(/_/g, ' ')))));
    return /*#__PURE__*/React.createElement("div", {
      style: {
        padding: 22,
        display: 'flex',
        flexDirection: 'column',
        gap: 14,
        maxWidth: 1180,
        margin: '0 auto'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'baseline',
        justifyContent: 'space-between'
      }
    }, /*#__PURE__*/React.createElement("h1", {
      style: {
        font: 'var(--type-h1)'
      }
    }, "Devices"), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)'
      }
    }, rows.length, " of ", scan.devices.length, " on 192.168.1.0/24")), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        flexWrap: 'wrap'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        width: 260
      }
    }, /*#__PURE__*/React.createElement(Input, {
      value: q,
      onChange: setQ,
      placeholder: "Search IP, host, vendor, MAC\u2026",
      iconLeft: /*#__PURE__*/React.createElement(I.search, {
        size: 15
      }),
      size: "sm"
    })), /*#__PURE__*/React.createElement(Select, {
      label: "Type",
      value: type,
      onChange: setType,
      options: TYPES
    }), /*#__PURE__*/React.createElement(Select, {
      label: "Reach",
      value: reach,
      onChange: setReach,
      options: REACH
    }), /*#__PURE__*/React.createElement("div", {
      style: {
        marginLeft: 'auto'
      }
    }, /*#__PURE__*/React.createElement(SegmentedControl, {
      size: "sm",
      value: view,
      onChange: setView,
      options: [{
        value: 'table',
        label: 'Table'
      }, {
        value: 'list',
        label: 'List'
      }]
    }))), view === 'table' ? /*#__PURE__*/React.createElement("div", {
      style: {
        border: '1px solid var(--hairline)',
        borderRadius: 'var(--radius-lg)',
        overflow: 'hidden',
        background: 'var(--surface-card)'
      }
    }, /*#__PURE__*/React.createElement("table", {
      style: {
        width: '100%',
        borderCollapse: 'collapse'
      }
    }, /*#__PURE__*/React.createElement("thead", null, /*#__PURE__*/React.createElement("tr", {
      style: {
        borderBottom: '1px solid var(--hairline)',
        background: 'var(--bg-sunken)'
      }
    }, /*#__PURE__*/React.createElement(Th, {
      k: "ip",
      w: "150"
    }, "IP"), /*#__PURE__*/React.createElement(Th, {
      k: "hostname"
    }, "Hostname"), /*#__PURE__*/React.createElement(Th, {
      k: "type",
      w: "150"
    }, "Type"), /*#__PURE__*/React.createElement(Th, {
      k: "vendor"
    }, "Vendor"), /*#__PURE__*/React.createElement(Th, {
      k: "reachability",
      w: "120"
    }, "Reach"), /*#__PURE__*/React.createElement(Th, {
      k: "confidence",
      w: "120"
    }, "Confidence"))), /*#__PURE__*/React.createElement("tbody", null, rows.map(d => {
      const r = fmt.reach(d.reachability);
      const Icon = I[fmt.deviceIcon(d.type)];
      const on = d.id === selId;
      return /*#__PURE__*/React.createElement("tr", {
        key: d.id,
        onClick: () => openDrawer('device', d),
        style: {
          borderBottom: '1px solid var(--hairline)',
          cursor: 'pointer',
          background: on ? 'var(--accent-ghost)' : 'transparent'
        }
      }, /*#__PURE__*/React.createElement("td", {
        style: {
          padding: '11px 12px',
          fontFamily: 'var(--font-mono)',
          fontSize: 'var(--text-sm)',
          color: 'var(--fg-1)'
        }
      }, d.ip), /*#__PURE__*/React.createElement("td", {
        style: {
          padding: '11px 12px'
        }
      }, /*#__PURE__*/React.createElement("div", {
        style: {
          display: 'flex',
          alignItems: 'center',
          gap: 9
        }
      }, /*#__PURE__*/React.createElement("span", {
        style: {
          color: 'var(--fg-3)',
          flex: '0 0 auto'
        }
      }, /*#__PURE__*/React.createElement(Icon, {
        size: 16
      })), /*#__PURE__*/React.createElement("span", {
        style: {
          fontSize: 'var(--text-sm)',
          color: 'var(--fg-1)'
        }
      }, d.hostname !== '—' ? d.hostname : /*#__PURE__*/React.createElement("span", {
        style: {
          color: 'var(--fg-4)'
        }
      }, "\u2014")))), /*#__PURE__*/React.createElement("td", {
        style: {
          padding: '11px 12px',
          fontFamily: 'var(--font-mono)',
          fontSize: 'var(--text-xs)',
          color: 'var(--fg-3)'
        }
      }, d.type), /*#__PURE__*/React.createElement("td", {
        style: {
          padding: '11px 12px',
          fontSize: 'var(--text-sm)',
          color: 'var(--fg-2)'
        }
      }, d.vendor), /*#__PURE__*/React.createElement("td", {
        style: {
          padding: '11px 12px'
        }
      }, /*#__PURE__*/React.createElement(Badge, {
        tone: r.tone,
        size: "sm"
      }, r.word)), /*#__PURE__*/React.createElement("td", {
        style: {
          padding: '11px 12px'
        }
      }, /*#__PURE__*/React.createElement("div", {
        style: {
          display: 'flex',
          alignItems: 'center',
          gap: 8
        }
      }, /*#__PURE__*/React.createElement("div", {
        style: {
          flex: 1,
          height: 4,
          background: 'var(--surface-3)',
          borderRadius: 'var(--radius-pill)',
          overflow: 'hidden'
        }
      }, /*#__PURE__*/React.createElement("div", {
        style: {
          width: fmt.pct(d.confidence),
          height: '100%',
          background: fmt.bandColor(d.confidence),
          borderRadius: 'var(--radius-pill)'
        }
      })), /*#__PURE__*/React.createElement(ConfPill, {
        v: d.confidence
      }))));
    }))), rows.length === 0 && /*#__PURE__*/React.createElement("div", {
      style: {
        padding: 40,
        textAlign: 'center',
        color: 'var(--fg-4)',
        fontSize: 'var(--text-sm)'
      }
    }, "No devices match these filters.")) : /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'grid',
        gridTemplateColumns: 'repeat(2, 1fr)',
        gap: 12
      }
    }, rows.map(d => {
      const r = fmt.reach(d.reachability);
      const Icon = I[fmt.deviceIcon(d.type)];
      return /*#__PURE__*/React.createElement("div", {
        key: d.id,
        onClick: () => openDrawer('device', d),
        style: {
          display: 'flex',
          alignItems: 'center',
          gap: 12,
          padding: 14,
          background: d.id === selId ? 'var(--accent-ghost)' : 'var(--surface-card)',
          border: '1px solid ' + (d.id === selId ? 'var(--accent-ring)' : 'var(--hairline)'),
          borderRadius: 'var(--radius-lg)',
          cursor: 'pointer'
        }
      }, /*#__PURE__*/React.createElement("span", {
        style: {
          width: 38,
          height: 38,
          borderRadius: 'var(--radius-md)',
          background: 'var(--surface-3)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'var(--fg-2)',
          flex: '0 0 auto'
        }
      }, /*#__PURE__*/React.createElement(Icon, {
        size: 19
      })), /*#__PURE__*/React.createElement("div", {
        style: {
          minWidth: 0,
          flex: 1
        }
      }, /*#__PURE__*/React.createElement("div", {
        style: {
          display: 'flex',
          alignItems: 'center',
          gap: 8
        }
      }, /*#__PURE__*/React.createElement("span", {
        style: {
          fontSize: 'var(--text-sm)',
          color: 'var(--fg-1)',
          fontWeight: 600
        }
      }, d.hostname !== '—' ? d.hostname : d.ip), /*#__PURE__*/React.createElement(Badge, {
        tone: r.tone,
        size: "sm"
      }, r.word)), /*#__PURE__*/React.createElement("div", {
        style: {
          fontFamily: 'var(--font-mono)',
          fontSize: 11,
          color: 'var(--fg-3)',
          marginTop: 2
        }
      }, d.ip, " \xB7 ", d.role)), /*#__PURE__*/React.createElement(ConfPill, {
        v: d.confidence
      }));
    })));
  }
  window.IAD_SCREENS = window.IAD_SCREENS || {};
  window.IAD_SCREENS.devices = Devices;
})();
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/console/Devices.jsx", error: String((e && e.message) || e) }); }

// ui_kits/console/Evidence.jsx
try { (() => {
/* IAD UI kit — Evidence screen. Probe explorer with status, tier, raw JSON.
   Flags the "success but empty evidence" anti-pattern. window.IAD_SCREENS.evidence */
(function () {
  const {
    useState
  } = React;
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const {
    ProbeStatusBadge,
    TierBadge,
    Badge,
    ConfidenceBar,
    Input
  } = NS;
  const I = window.Icons;
  const STATUSES = ['all', 'success', 'partial', 'no_data', 'skipped', 'failed', 'blocked'];
  function isEmptyEvidence(e) {
    if (!e.data) return true;
    const vals = Object.values(e.data);
    return vals.length === 0 || vals.every(v => v == null || v === '' || v === 0 || v === false);
  }
  function Evidence({
    scan,
    openDrawer,
    drawer
  }) {
    const [status, setStatus] = useState('all');
    const [q, setQ] = useState('');
    const [open, setOpen] = useState(scan.evidence[0].id);
    const selId = drawer && drawer.kind === 'probe' ? drawer.payload.id : null;
    const rows = scan.evidence.filter(e => {
      if (status !== 'all' && e.status !== status) return false;
      if (q && !(e.probe_name + e.reason).toLowerCase().includes(q.toLowerCase())) return false;
      return true;
    });
    const counts = STATUSES.slice(1).map(s => ({
      s,
      n: scan.evidence.filter(e => e.status === s).length
    }));
    return /*#__PURE__*/React.createElement("div", {
      style: {
        padding: 22,
        display: 'flex',
        flexDirection: 'column',
        gap: 14,
        maxWidth: 1000,
        margin: '0 auto'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'baseline',
        justifyContent: 'space-between'
      }
    }, /*#__PURE__*/React.createElement("h1", {
      style: {
        font: 'var(--type-h1)'
      }
    }, "Evidence"), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)'
      }
    }, scan.evidence.length, " probes \xB7 ", scan.evidence.filter(e => e.status === 'success').length, " success")), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        flexWrap: 'wrap'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        width: 240
      }
    }, /*#__PURE__*/React.createElement(Input, {
      value: q,
      onChange: setQ,
      placeholder: "Search probes\u2026",
      iconLeft: /*#__PURE__*/React.createElement(I.search, {
        size: 15
      }),
      size: "sm"
    })), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        gap: 6,
        flexWrap: 'wrap'
      }
    }, STATUSES.map(s => {
      const on = status === s;
      const n = s === 'all' ? scan.evidence.length : (counts.find(c => c.s === s) || {}).n;
      return /*#__PURE__*/React.createElement("button", {
        key: s,
        onClick: () => setStatus(s),
        style: {
          display: 'inline-flex',
          alignItems: 'center',
          gap: 6,
          height: 28,
          padding: '0 10px',
          borderRadius: 'var(--radius-sm)',
          cursor: 'pointer',
          border: '1px solid ' + (on ? 'var(--hairline-strong)' : 'transparent'),
          background: on ? 'var(--surface-2)' : 'transparent',
          color: on ? 'var(--fg-1)' : 'var(--fg-3)',
          fontFamily: 'var(--font-mono)',
          fontSize: 11,
          textTransform: 'uppercase',
          letterSpacing: '.06em'
        }
      }, s, /*#__PURE__*/React.createElement("span", {
        style: {
          color: 'var(--fg-4)'
        }
      }, n));
    }))), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 10
      }
    }, rows.map(e => {
      const isOpen = open === e.id;
      const emptyWarn = e.status === 'success' && isEmptyEvidence(e);
      return /*#__PURE__*/React.createElement("div", {
        key: e.id,
        style: {
          border: '1px solid ' + (e.id === selId ? 'var(--accent-ring)' : 'var(--hairline)'),
          borderRadius: 'var(--radius-lg)',
          background: 'var(--surface-card)',
          overflow: 'hidden'
        }
      }, /*#__PURE__*/React.createElement("div", {
        onClick: () => setOpen(isOpen ? null : e.id),
        style: {
          display: 'flex',
          alignItems: 'center',
          gap: 12,
          padding: '13px 16px',
          cursor: 'pointer'
        }
      }, /*#__PURE__*/React.createElement("span", {
        style: {
          color: 'var(--fg-4)',
          transform: isOpen ? 'rotate(90deg)' : 'none',
          transition: 'transform var(--dur-fast) var(--ease-out)',
          display: 'inline-flex'
        }
      }, /*#__PURE__*/React.createElement(I.chevronRight, {
        size: 15
      })), /*#__PURE__*/React.createElement("span", {
        style: {
          fontFamily: 'var(--font-mono)',
          fontSize: 'var(--text-sm)',
          color: 'var(--fg-1)',
          fontWeight: 600,
          minWidth: 150
        }
      }, e.probe_name), /*#__PURE__*/React.createElement(ProbeStatusBadge, {
        status: e.status,
        size: "sm"
      }), /*#__PURE__*/React.createElement(TierBadge, {
        tier: e.evidence_class,
        appearance: "dot"
      }), /*#__PURE__*/React.createElement("span", {
        style: {
          fontSize: 'var(--text-sm)',
          color: 'var(--fg-3)',
          flex: 1,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap'
        }
      }, e.reason), /*#__PURE__*/React.createElement("span", {
        style: {
          fontFamily: 'var(--font-mono)',
          fontSize: 11,
          color: 'var(--fg-4)'
        }
      }, e.ts)), isOpen && /*#__PURE__*/React.createElement("div", {
        style: {
          padding: '0 16px 16px 42px',
          display: 'flex',
          flexDirection: 'column',
          gap: 14
        }
      }, emptyWarn && /*#__PURE__*/React.createElement("div", {
        style: {
          display: 'flex',
          gap: 10,
          padding: 11,
          background: 'var(--warn-bg)',
          borderRadius: 'var(--radius-md)'
        }
      }, /*#__PURE__*/React.createElement("span", {
        style: {
          color: 'var(--warn)',
          flex: '0 0 auto'
        }
      }, /*#__PURE__*/React.createElement(I.alert, {
        size: 15
      })), /*#__PURE__*/React.createElement("span", {
        style: {
          fontSize: 'var(--text-xs)',
          color: 'var(--fg-2)',
          lineHeight: 1.45
        }
      }, "Probe completed but returned no useful evidence. Consider normalizing this as ", /*#__PURE__*/React.createElement("b", {
        style: {
          fontFamily: 'var(--font-mono)'
        }
      }, "no_data"), ".")), e.confidence > 0 && /*#__PURE__*/React.createElement("div", {
        style: {
          maxWidth: 300
        }
      }, /*#__PURE__*/React.createElement(ConfidenceBar, {
        value: e.confidence,
        label: "Probe confidence",
        size: "sm"
      })), e.limitations && /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
        style: {
          font: 'var(--type-overline)',
          letterSpacing: 'var(--ls-caps)',
          textTransform: 'uppercase',
          color: 'var(--fg-4)',
          marginBottom: 5
        }
      }, "Limitations"), /*#__PURE__*/React.createElement("p", {
        style: {
          fontSize: 'var(--text-xs)',
          color: 'var(--fg-3)',
          lineHeight: 1.5
        }
      }, e.limitations)), /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
        style: {
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 6
        }
      }, /*#__PURE__*/React.createElement("span", {
        style: {
          font: 'var(--type-overline)',
          letterSpacing: 'var(--ls-caps)',
          textTransform: 'uppercase',
          color: 'var(--fg-4)'
        }
      }, "Raw evidence"), /*#__PURE__*/React.createElement(Badge, {
        tone: "neutral",
        appearance: "outline",
        mono: true,
        size: "sm"
      }, e.evidence_class)), /*#__PURE__*/React.createElement("pre", {
        style: {
          margin: 0,
          padding: 12,
          background: 'var(--ink-900)',
          border: '1px solid var(--hairline)',
          borderRadius: 'var(--radius-md)',
          fontFamily: 'var(--font-mono)',
          fontSize: 11,
          color: 'var(--fg-2)',
          lineHeight: 1.55,
          overflow: 'auto',
          whiteSpace: 'pre-wrap'
        }
      }, JSON.stringify(e.data, null, 2)))));
    })));
  }
  window.IAD_SCREENS = window.IAD_SCREENS || {};
  window.IAD_SCREENS.evidence = Evidence;
})();
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/console/Evidence.jsx", error: String((e && e.message) || e) }); }

// ui_kits/console/ReportsSettings.jsx
try { (() => {
/* IAD UI kit — Reports + Settings screens.
   window.IAD_SCREENS.reports / .settings */
(function () {
  const {
    useState
  } = React;
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const {
    Card,
    Button,
    Toggle,
    SegmentedControl,
    Badge,
    StatusDot
  } = NS;
  const I = window.Icons;
  const fmt = window.fmt;
  function Reports({
    scan
  }) {
    const [copied, setCopied] = useState(false);
    const summary = `IAD scan ${scan.scan_id} — ${scan.primary_type} (${fmt.pct(scan.confidence)} ${fmt.bandWord(scan.confidence)} confidence, ${scan.decision_quality} quality)\nISP ${scan.public_ip.asn} ${scan.public_ip.org} · Public IP ${scan.public_ip.address}\nNAT: CGNAT (${scan.nat_topology.layers} layers) · IPv6 native · ${scan.devices.length} LAN devices`;
    const copy = () => {
      try {
        navigator.clipboard.writeText(summary);
      } catch (e) {}
      setCopied(true);
      setTimeout(() => setCopied(false), 1600);
    };
    const Action = ({
      icon,
      title,
      desc,
      btn,
      onClick,
      primary,
      disabled
    }) => /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 14,
        padding: 16,
        background: 'var(--surface-card)',
        border: '1px solid var(--hairline)',
        borderRadius: 'var(--radius-lg)'
      }
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        width: 38,
        height: 38,
        borderRadius: 'var(--radius-md)',
        background: 'var(--surface-3)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: 'var(--fg-2)',
        flex: '0 0 auto'
      }
    }, icon), /*#__PURE__*/React.createElement("div", {
      style: {
        flex: 1,
        minWidth: 0
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-1)',
        fontWeight: 600
      }
    }, title), /*#__PURE__*/React.createElement("div", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)'
      }
    }, desc)), /*#__PURE__*/React.createElement(Button, {
      size: "sm",
      variant: primary ? 'primary' : 'secondary',
      onClick: onClick,
      disabled: disabled
    }, btn));
    return /*#__PURE__*/React.createElement("div", {
      style: {
        padding: 22,
        display: 'flex',
        flexDirection: 'column',
        gap: 14,
        maxWidth: 860,
        margin: '0 auto'
      }
    }, /*#__PURE__*/React.createElement("h1", {
      style: {
        font: 'var(--type-h1)'
      }
    }, "Reports"), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 10
      }
    }, /*#__PURE__*/React.createElement(Action, {
      icon: /*#__PURE__*/React.createElement(I.upload, {
        size: 19
      }),
      title: "Import scan JSON",
      desc: "Validate against the NormalizedScanReport schema (Zod) before loading.",
      btn: "Choose file"
    }), /*#__PURE__*/React.createElement(Action, {
      icon: /*#__PURE__*/React.createElement(I.download, {
        size: 19
      }),
      title: "Export full report",
      desc: "The immutable scan evidence as JSON. UI layout positions are excluded.",
      btn: "Export JSON",
      primary: true
    }), /*#__PURE__*/React.createElement(Action, {
      icon: /*#__PURE__*/React.createElement(I.reports, {
        size: 19
      }),
      title: "Export scan summary",
      desc: "One-page human-readable summary (Markdown).",
      btn: "Export .md"
    }), /*#__PURE__*/React.createElement(Action, {
      icon: /*#__PURE__*/React.createElement(I.copy, {
        size: 19
      }),
      title: "Copy diagnostic summary",
      desc: "Three-line summary for tickets and chat.",
      btn: copied ? 'Copied ✓' : 'Copy',
      onClick: copy
    })), /*#__PURE__*/React.createElement(Card, {
      eyebrow: "Preview",
      title: "Diagnostic summary",
      style: {
        marginTop: 4
      }
    }, /*#__PURE__*/React.createElement("pre", {
      style: {
        margin: 0,
        fontFamily: 'var(--font-mono)',
        fontSize: 12,
        color: 'var(--fg-2)',
        lineHeight: 1.6,
        whiteSpace: 'pre-wrap'
      }
    }, summary)), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: 16,
        background: 'var(--bg-sunken)',
        border: '1px dashed var(--hairline-strong)',
        borderRadius: 'var(--radius-lg)',
        opacity: 0.85
      }
    }, /*#__PURE__*/React.createElement(Badge, {
      tone: "neutral",
      appearance: "outline",
      size: "sm",
      uppercase: true,
      mono: true
    }, "Planned"), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-3)'
      }
    }, "Compare two scan reports \u2014 not yet implemented. This placeholder is intentionally non-functional.")));
  }
  function SettingRow({
    label,
    desc,
    children
  }) {
    return /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 16,
        padding: '14px 0',
        borderBottom: '1px solid var(--hairline)'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        flex: 1,
        minWidth: 0
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-1)',
        fontWeight: 500
      }
    }, label), desc && /*#__PURE__*/React.createElement("div", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)',
        marginTop: 2
      }
    }, desc)), /*#__PURE__*/React.createElement("div", {
      style: {
        flex: '0 0 auto'
      }
    }, children));
  }
  function Settings({
    scan
  }) {
    const [s, setS] = useState({
      theme: 'dark',
      engine: 'elk_layered',
      lowconf: true,
      unknown: true,
      isp: true,
      persist: true
    });
    const set = (k, v) => setS(p => ({
      ...p,
      [k]: v
    }));
    return /*#__PURE__*/React.createElement("div", {
      style: {
        padding: 22,
        display: 'flex',
        flexDirection: 'column',
        gap: 18,
        maxWidth: 720,
        margin: '0 auto'
      }
    }, /*#__PURE__*/React.createElement("h1", {
      style: {
        font: 'var(--type-h1)'
      }
    }, "Settings"), /*#__PURE__*/React.createElement(Card, {
      eyebrow: "Appearance",
      title: "Theme & color"
    }, /*#__PURE__*/React.createElement(SettingRow, {
      label: "Theme",
      desc: "Dark is the default instrument surface."
    }, /*#__PURE__*/React.createElement(SegmentedControl, {
      size: "sm",
      value: s.theme,
      onChange: v => set('theme', v),
      options: [{
        value: 'dark',
        label: 'Dark'
      }, {
        value: 'light',
        label: 'Light'
      }]
    })), /*#__PURE__*/React.createElement(SettingRow, {
      label: "Color mode",
      desc: "Monochrome with status-only color. Locked in this build."
    }, /*#__PURE__*/React.createElement(Badge, {
      tone: "neutral",
      appearance: "outline",
      size: "sm",
      mono: true
    }, "black_white"))), /*#__PURE__*/React.createElement(Card, {
      eyebrow: "Topology",
      title: "Map & layout"
    }, /*#__PURE__*/React.createElement(SettingRow, {
      label: "Layout engine",
      desc: "ELK layered is recommended for hierarchy clarity."
    }, /*#__PURE__*/React.createElement(SegmentedControl, {
      size: "sm",
      value: s.engine,
      onChange: v => set('engine', v),
      options: [{
        value: 'elk_layered',
        label: 'Layered'
      }, {
        value: 'force',
        label: 'Force'
      }, {
        value: 'manual',
        label: 'Manual'
      }]
    })), /*#__PURE__*/React.createElement(SettingRow, {
      label: "Show low-confidence edges",
      desc: "Render edges below the 0.45 band, visually muted."
    }, /*#__PURE__*/React.createElement(Toggle, {
      checked: s.lowconf,
      onChange: v => set('lowconf', v)
    })), /*#__PURE__*/React.createElement(SettingRow, {
      label: "Show unknown L2 segments",
      desc: "Inferred switches and the hosts hidden behind them."
    }, /*#__PURE__*/React.createElement(Toggle, {
      checked: s.unknown,
      onChange: v => set('unknown', v)
    })), /*#__PURE__*/React.createElement(SettingRow, {
      label: "Show ISP route context",
      desc: "Gateway chain hops beyond the home network."
    }, /*#__PURE__*/React.createElement(Toggle, {
      checked: s.isp,
      onChange: v => set('isp', v)
    })), /*#__PURE__*/React.createElement(SettingRow, {
      label: "Persist node positions",
      desc: "Save manual layout to local UI state \u2014 never to scan data."
    }, /*#__PURE__*/React.createElement(Toggle, {
      checked: s.persist,
      onChange: v => set('persist', v)
    })), /*#__PURE__*/React.createElement(SettingRow, {
      label: "Reset UI layout positions",
      desc: "Restore the generated layout. Does not touch evidence."
    }, /*#__PURE__*/React.createElement(Button, {
      size: "sm",
      variant: "danger"
    }, "Reset layout"))), /*#__PURE__*/React.createElement(Card, {
      eyebrow: "Safety",
      title: "Data integrity"
    }, /*#__PURE__*/React.createElement(SettingRow, {
      label: "Safe mode",
      desc: "Read-only topology. Scan evidence is immutable; no destructive actions."
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 8
      }
    }, /*#__PURE__*/React.createElement(StatusDot, {
      tone: "success"
    }), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--ok)',
        fontWeight: 600
      }
    }, "Enabled")))));
  }
  window.IAD_SCREENS = window.IAD_SCREENS || {};
  window.IAD_SCREENS.reports = Reports;
  window.IAD_SCREENS.settings = Settings;
})();
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/console/ReportsSettings.jsx", error: String((e && e.message) || e) }); }

// ui_kits/console/Topology.jsx
try { (() => {
/* IAD UI kit — Topology screen. Read-only interactive map (pan / zoom / drag /
   select), layer toggles, legend. No create/delete — layout positions are UI
   state only, never written back to scan data. window.IAD_SCREENS.topology
   NOTE: the production app uses React Flow + ELK; this is a faithful SVG
   recreation of that read-only topology mode. */
(function () {
  const {
    useState,
    useRef,
    useCallback
  } = React;
  const NS = window.IADInternetAccessDetectorDesignSystem_019e02;
  const {
    IconButton,
    Toggle,
    Badge,
    TierBadge
  } = NS;
  const I = window.Icons;
  function buildGraph(scan) {
    const dev = id => scan.devices.find(d => d.id === id);
    const nodes = [{
      id: 'host',
      label: 'This host',
      sub: '192.168.1.24',
      icon: 'host',
      x: 430,
      y: 50,
      layers: ['l3'],
      device: dev('d-host'),
      accent: true
    }, {
      id: 'router',
      label: 'Home router',
      sub: '192.168.1.1',
      icon: 'router',
      x: 430,
      y: 168,
      layers: ['l2', 'l3'],
      device: dev('d-gw')
    }, {
      id: 'ap',
      label: 'Mesh AP',
      sub: '192.168.1.2',
      icon: 'ap',
      x: 150,
      y: 250,
      layers: ['l2'],
      device: dev('d-ap')
    }, {
      id: 'nas',
      label: 'NAS',
      sub: '192.168.1.30',
      icon: 'server',
      x: 280,
      y: 322,
      layers: ['l2'],
      device: dev('d-nas')
    }, {
      id: 'printer',
      label: 'Printer',
      sub: '192.168.1.41',
      icon: 'printer',
      x: 600,
      y: 248,
      layers: ['l2'],
      device: dev('d-print')
    }, {
      id: 'switch',
      label: 'Switch',
      sub: 'inferred · unmanaged',
      icon: 'switchIcon',
      x: 712,
      y: 330,
      layers: ['unknown'],
      inferred: true
    }, {
      id: 'segment',
      label: 'Unknown L2 segment',
      sub: '≥1 host behind',
      icon: 'unknown',
      x: 712,
      y: 438,
      layers: ['unknown'],
      inferred: true
    }, {
      id: 'cgnat',
      label: 'CGNAT gateway',
      sub: '100.64.12.1',
      icon: 'router',
      x: 430,
      y: 300,
      layers: ['l3', 'nat'],
      badge: 'NAT'
    }, {
      id: 'isp',
      label: 'ISP edge',
      sub: '203.0.113.1',
      icon: 'globe',
      x: 430,
      y: 412,
      layers: ['isp']
    }, {
      id: 'inet',
      label: 'Public internet',
      sub: 'AS3320',
      icon: 'globe',
      x: 430,
      y: 524,
      layers: ['isp']
    }];
    const edges = [{
      id: 'e1',
      from: 'host',
      to: 'router',
      type: 'local_interface',
      kind: 'confirmed',
      tier: 'l3',
      conf: 1.0,
      layers: ['l3'],
      accent: true,
      label: 'Local interface',
      basis: 'Default route via this NIC; same /24 as gateway.'
    }, {
      id: 'e2',
      from: 'router',
      to: 'nas',
      type: 'arp_confirmed',
      kind: 'confirmed',
      tier: 'l2',
      conf: 0.93,
      layers: ['l2'],
      label: 'ARP confirmed',
      basis: 'NAS answered ARP on the local broadcast domain.'
    }, {
      id: 'e3',
      from: 'router',
      to: 'printer',
      type: 'arp_confirmed',
      kind: 'confirmed',
      tier: 'l2',
      conf: 0.81,
      layers: ['l2'],
      label: 'ARP confirmed',
      basis: 'Printer answered ARP + advertised IPP via mDNS.'
    }, {
      id: 'e4',
      from: 'router',
      to: 'ap',
      type: 'wifi_association_inferred',
      kind: 'inferred',
      tier: 'l2',
      conf: 0.6,
      layers: ['l2'],
      label: 'Wi-Fi assoc (inferred)',
      basis: 'mDNS suggests a mesh repeater; bridge topology not confirmed.'
    }, {
      id: 'e5',
      from: 'router',
      to: 'switch',
      type: 'unknown_l2_connection',
      kind: 'unknown',
      tier: 'l2',
      conf: 0.3,
      layers: ['unknown'],
      label: 'Unknown L2',
      basis: 'MAC counts imply an unmanaged switch, but it is invisible to probes.'
    }, {
      id: 'e6',
      from: 'switch',
      to: 'segment',
      type: 'unknown_l2_connection',
      kind: 'unknown',
      tier: 'l2',
      conf: 0.22,
      layers: ['unknown'],
      label: 'Unknown L2',
      basis: 'At least one host sits beyond the inferred switch; count is a lower bound.'
    }, {
      id: 'e7',
      from: 'router',
      to: 'cgnat',
      type: 'upstream_private_gateway',
      kind: 'confirmed',
      tier: 'l3',
      conf: 0.88,
      layers: ['l3', 'nat'],
      boundary: 'NAT',
      label: 'NAT boundary',
      basis: 'Next hop is RFC 6598 (100.64/10): carrier-grade NAT.'
    }, {
      id: 'e8',
      from: 'cgnat',
      to: 'isp',
      type: 'route_hop',
      kind: 'confirmed',
      tier: 'isp',
      conf: 0.7,
      layers: ['l3', 'isp'],
      thin: true,
      label: 'Route hop',
      basis: 'Traceroute L3 hop — a router, not a physical switch.'
    }, {
      id: 'e9',
      from: 'isp',
      to: 'inet',
      type: 'isp_boundary',
      kind: 'confirmed',
      tier: 'isp',
      conf: 0.85,
      layers: ['isp'],
      boundary: 'ISP',
      label: 'ISP boundary',
      basis: 'First public hop; ISP-internal topology is not observable.'
    }];
    const byId = Object.fromEntries(nodes.map(n => [n.id, n]));
    edges.forEach(e => {
      e.fromLabel = byId[e.from].label;
      e.toLabel = byId[e.to].label;
    });
    return {
      nodes,
      edges
    };
  }
  const LAYERS = [{
    id: 'l2',
    label: 'L2 · Link',
    tier: 'l2'
  }, {
    id: 'l3',
    label: 'L3 · Routing',
    tier: 'l3'
  }, {
    id: 'nat',
    label: 'NAT',
    tier: 'nat'
  }, {
    id: 'isp',
    label: 'ISP route context',
    tier: 'isp'
  }, {
    id: 'unknown',
    label: 'Unknown segments',
    tier: null
  }, {
    id: 'lowconf',
    label: 'Low-confidence edges',
    tier: null
  }];
  function edgePath(a, b) {
    const mx = (a.x + b.x) / 2;
    return `M ${a.x} ${a.y} C ${mx} ${a.y}, ${mx} ${b.y}, ${b.x} ${b.y}`;
  }
  function Topology({
    scan,
    openDrawer,
    drawer
  }) {
    const graphRef = useRef(buildGraph(scan));
    const {
      edges
    } = graphRef.current;
    const [positions, setPositions] = useState(() => Object.fromEntries(graphRef.current.nodes.map(n => [n.id, {
      x: n.x,
      y: n.y
    }])));
    const [view, setView] = useState({
      x: 60,
      y: 20,
      k: 0.92
    });
    const [layers, setLayers] = useState({
      l2: true,
      l3: true,
      nat: true,
      isp: true,
      unknown: true,
      lowconf: true
    });
    const [sel, setSel] = useState(null);
    const svgRef = useRef(null);
    const drag = useRef(null);
    const nodes = graphRef.current.nodes.map(n => ({
      ...n,
      ...positions[n.id]
    }));
    const nodeVisible = n => n.layers.some(l => layers[l]) || n.id === 'host';
    const edgeVisible = e => {
      if (!e.layers.some(l => layers[l])) return false;
      if (e.conf < 0.45 && !layers.lowconf) return false;
      return true;
    };
    const onWheel = useCallback(ev => {
      ev.preventDefault();
      setView(v => {
        const k = Math.max(0.4, Math.min(2.2, v.k * (ev.deltaY < 0 ? 1.1 : 0.9)));
        const rect = svgRef.current.getBoundingClientRect();
        const cx = ev.clientX - rect.left,
          cy = ev.clientY - rect.top;
        const nx = cx - (cx - v.x) * (k / v.k);
        const ny = cy - (cy - v.y) * (k / v.k);
        return {
          x: nx,
          y: ny,
          k
        };
      });
    }, []);
    const onPointerDownBg = ev => {
      if (ev.target.closest('[data-node]') || ev.target.closest('[data-edge]')) return;
      setSel(null);
      drag.current = {
        mode: 'pan',
        sx: ev.clientX,
        sy: ev.clientY,
        ox: view.x,
        oy: view.y
      };
    };
    const onPointerDownNode = (ev, n) => {
      ev.stopPropagation();
      setSel({
        kind: 'node',
        id: n.id
      });
      drag.current = {
        mode: 'node',
        id: n.id,
        sx: ev.clientX,
        sy: ev.clientY,
        ox: positions[n.id].x,
        oy: positions[n.id].y,
        moved: false
      };
    };
    const onPointerMove = ev => {
      const d = drag.current;
      if (!d) return;
      if (d.mode === 'pan') setView(v => ({
        ...v,
        x: d.ox + (ev.clientX - d.sx),
        y: d.oy + (ev.clientY - d.sy)
      }));else if (d.mode === 'node') {
        const dx = (ev.clientX - d.sx) / view.k,
          dy = (ev.clientY - d.sy) / view.k;
        if (Math.abs(dx) > 1 || Math.abs(dy) > 1) d.moved = true;
        setPositions(p => ({
          ...p,
          [d.id]: {
            x: d.ox + dx,
            y: d.oy + dy
          }
        }));
      }
    };
    const onPointerUp = ev => {
      const d = drag.current;
      if (d && d.mode === 'node' && !d.moved) {
        const n = graphRef.current.nodes.find(x => x.id === d.id);
        openDrawer('node', n.device ? {
          device: n.device
        } : {
          probe_name: n.label
        });
      }
      drag.current = null;
    };
    const fitView = () => setView({
      x: 60,
      y: 20,
      k: 0.92
    });
    const resetLayout = () => setPositions(Object.fromEntries(graphRef.current.nodes.map(n => [n.id, {
      x: n.x,
      y: n.y
    }])));
    const tierColor = t => ({
      l2: 'var(--tier-l2)',
      l3: 'var(--fg-2)',
      nat: 'var(--tier-nat)',
      isp: 'var(--tier-isp)'
    })[t] || 'var(--fg-3)';
    const edgeStroke = e => {
      if (e.accent) return 'var(--accent-base)';
      if (e.conf < 0.45 && layers.lowconf) return 'var(--edge-muted)';
      if (e.kind === 'unknown') return 'var(--edge-unknown)';
      if (e.boundary) return tierColor(e.tier);
      if (e.kind === 'inferred') return 'var(--edge-inferred)';
      return 'var(--edge-confirmed)';
    };
    const edgeDash = e => e.kind === 'unknown' ? '2 7' : e.kind === 'inferred' ? '7 6' : e.conf < 0.45 ? '7 6' : 'none';
    return /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        height: '100%'
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '14px 22px',
        borderBottom: '1px solid var(--hairline)',
        flex: '0 0 auto'
      }
    }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("h1", {
      style: {
        font: 'var(--type-h2)'
      }
    }, "Topology"), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-3)'
      }
    }, "Generated from network context \xB7 6 confirmed \xB7 3 inferred links \xB7 read-only")), /*#__PURE__*/React.createElement(Badge, {
      tone: "neutral",
      appearance: "outline",
      size: "sm",
      mono: true
    }, /*#__PURE__*/React.createElement("span", {
      style: {
        display: 'inline-flex',
        marginRight: 5,
        verticalAlign: 'middle'
      }
    }, /*#__PURE__*/React.createElement(I.lock, {
      size: 11
    })), "No edits")), /*#__PURE__*/React.createElement("div", {
      style: {
        flex: 1,
        display: 'flex',
        minHeight: 0
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        flex: 1,
        position: 'relative',
        minWidth: 0,
        background: 'var(--ink-800)',
        backgroundImage: 'radial-gradient(var(--grid-line-strong) 1px, transparent 1px)',
        backgroundSize: 24 * view.k + 'px ' + 24 * view.k + 'px',
        backgroundPosition: view.x + 'px ' + view.y + 'px',
        overflow: 'hidden'
      }
    }, /*#__PURE__*/React.createElement("svg", {
      ref: svgRef,
      width: "100%",
      height: "100%",
      style: {
        position: 'absolute',
        inset: 0,
        cursor: drag.current && drag.current.mode === 'pan' ? 'grabbing' : 'grab'
      },
      onWheel: onWheel,
      onPointerDown: onPointerDownBg,
      onPointerMove: onPointerMove,
      onPointerUp: onPointerUp,
      onPointerLeave: onPointerUp
    }, /*#__PURE__*/React.createElement("g", {
      transform: `translate(${view.x} ${view.y}) scale(${view.k})`
    }, edges.filter(edgeVisible).map(e => {
      const a = positions[e.from],
        b = positions[e.to];
      const selected = sel && sel.kind === 'edge' && sel.id === e.id;
      const mx = (a.x + b.x) / 2,
        my = (a.y + b.y) / 2;
      return /*#__PURE__*/React.createElement("g", {
        key: e.id,
        "data-edge": e.id,
        style: {
          cursor: 'pointer'
        },
        onPointerDown: ev => {
          ev.stopPropagation();
          setSel({
            kind: 'edge',
            id: e.id
          });
          openDrawer('edge', e);
        }
      }, /*#__PURE__*/React.createElement("path", {
        d: edgePath(a, b),
        fill: "none",
        stroke: "transparent",
        strokeWidth: 14
      }), /*#__PURE__*/React.createElement("path", {
        d: edgePath(a, b),
        fill: "none",
        stroke: selected ? 'var(--accent-base)' : edgeStroke(e),
        strokeWidth: e.thin ? 1.3 : selected ? 2.6 : 1.8,
        strokeDasharray: edgeDash(e),
        strokeLinecap: "round"
      }), e.boundary && /*#__PURE__*/React.createElement("g", {
        transform: `translate(${mx} ${my})`
      }, /*#__PURE__*/React.createElement("rect", {
        x: -22,
        y: -9,
        width: 44,
        height: 18,
        rx: 4,
        fill: "var(--ink-850)",
        stroke: tierColor(e.tier),
        strokeWidth: 1
      }), /*#__PURE__*/React.createElement("text", {
        x: 0,
        y: 4,
        textAnchor: "middle",
        fontFamily: "var(--font-mono)",
        fontSize: 10,
        fontWeight: 700,
        fill: tierColor(e.tier),
        style: {
          letterSpacing: '.06em'
        }
      }, e.boundary)));
    }), nodes.filter(nodeVisible).map(n => {
      const Icon = I[n.icon];
      const selected = sel && sel.kind === 'node' && sel.id === n.id;
      return /*#__PURE__*/React.createElement("g", {
        key: n.id,
        "data-node": n.id,
        transform: `translate(${n.x} ${n.y})`,
        style: {
          cursor: 'grab'
        },
        onPointerDown: ev => onPointerDownNode(ev, n)
      }, /*#__PURE__*/React.createElement("g", {
        transform: "translate(-66 -26)"
      }, /*#__PURE__*/React.createElement("rect", {
        width: 132,
        height: 52,
        rx: 8,
        fill: "var(--node-fill)",
        stroke: selected ? 'var(--accent-base)' : n.accent ? 'var(--accent-ring)' : 'var(--node-stroke)',
        strokeWidth: selected ? 2 : 1,
        strokeDasharray: n.inferred ? '5 4' : 'none'
      }), selected && /*#__PURE__*/React.createElement("rect", {
        x: -3,
        y: -3,
        width: 138,
        height: 58,
        rx: 10,
        fill: "none",
        stroke: "var(--accent-base)",
        strokeWidth: 1,
        opacity: 0.35
      }), /*#__PURE__*/React.createElement("g", {
        transform: "translate(11 11)"
      }, /*#__PURE__*/React.createElement("rect", {
        width: 30,
        height: 30,
        rx: 6,
        fill: "var(--surface-3)"
      }), /*#__PURE__*/React.createElement("g", {
        transform: "translate(7 7)",
        color: n.accent ? 'var(--accent-bright)' : n.inferred ? 'var(--fg-3)' : 'var(--fg-2)'
      }, /*#__PURE__*/React.createElement(Icon, {
        size: 16
      }))), /*#__PURE__*/React.createElement("text", {
        x: 50,
        y: 22,
        fontFamily: "var(--font-sans)",
        fontSize: 12,
        fontWeight: 600,
        fill: "var(--fg-1)"
      }, n.label), /*#__PURE__*/React.createElement("text", {
        x: 50,
        y: 38,
        fontFamily: "var(--font-mono)",
        fontSize: 10,
        fill: "var(--fg-3)"
      }, n.sub), n.badge && /*#__PURE__*/React.createElement("g", {
        transform: "translate(104 7)"
      }, /*#__PURE__*/React.createElement("rect", {
        width: 20,
        height: 13,
        rx: 3,
        fill: "var(--tier-nat-bg)"
      }), /*#__PURE__*/React.createElement("text", {
        x: 10,
        y: 9.5,
        textAnchor: "middle",
        fontFamily: "var(--font-mono)",
        fontSize: 8,
        fontWeight: 700,
        fill: "var(--tier-nat)"
      }, n.badge))));
    }))), /*#__PURE__*/React.createElement("div", {
      style: {
        position: 'absolute',
        top: 14,
        right: 14,
        display: 'flex',
        flexDirection: 'column',
        gap: 6,
        background: 'var(--surface-card)',
        border: '1px solid var(--hairline)',
        borderRadius: 'var(--radius-md)',
        padding: 5
      }
    }, /*#__PURE__*/React.createElement(IconButton, {
      label: "Zoom in",
      onClick: () => setView(v => ({
        ...v,
        k: Math.min(2.2, v.k * 1.15)
      }))
    }, /*#__PURE__*/React.createElement(I.zoomIn, {
      size: 16
    })), /*#__PURE__*/React.createElement(IconButton, {
      label: "Zoom out",
      onClick: () => setView(v => ({
        ...v,
        k: Math.max(0.4, v.k * 0.87)
      }))
    }, /*#__PURE__*/React.createElement(I.zoomOut, {
      size: 16
    })), /*#__PURE__*/React.createElement(IconButton, {
      label: "Fit view",
      onClick: fitView
    }, /*#__PURE__*/React.createElement(I.fit, {
      size: 16
    })), /*#__PURE__*/React.createElement(IconButton, {
      label: "Reset layout",
      onClick: resetLayout
    }, /*#__PURE__*/React.createElement(I.reset, {
      size: 16
    }))), /*#__PURE__*/React.createElement("div", {
      style: {
        position: 'absolute',
        bottom: 12,
        left: 14,
        fontFamily: 'var(--font-mono)',
        fontSize: 10,
        color: 'var(--fg-4)'
      }
    }, "scroll = zoom \xB7 drag bg = pan \xB7 drag node = reposition \xB7 ", Math.round(view.k * 100), "%")), /*#__PURE__*/React.createElement("div", {
      style: {
        width: 232,
        flex: '0 0 auto',
        borderLeft: '1px solid var(--hairline)',
        background: 'var(--surface-card)',
        overflow: 'auto',
        padding: 16,
        display: 'flex',
        flexDirection: 'column',
        gap: 18
      }
    }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
      style: {
        font: 'var(--type-overline)',
        letterSpacing: 'var(--ls-caps)',
        textTransform: 'uppercase',
        color: 'var(--fg-3)',
        marginBottom: 12
      }
    }, "Layers"), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 12
      }
    }, LAYERS.map(l => /*#__PURE__*/React.createElement("div", {
      key: l.id,
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 9
      }
    }, /*#__PURE__*/React.createElement(Toggle, {
      size: "sm",
      checked: layers[l.id],
      onChange: v => setLayers(s => ({
        ...s,
        [l.id]: v
      }))
    }), l.tier ? /*#__PURE__*/React.createElement(TierBadge, {
      tier: l.tier,
      appearance: "dot",
      label: l.label
    }) : /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-sm)',
        color: 'var(--fg-2)'
      }
    }, l.label))))), /*#__PURE__*/React.createElement("div", {
      style: {
        borderTop: '1px solid var(--hairline)',
        paddingTop: 16
      }
    }, /*#__PURE__*/React.createElement("div", {
      style: {
        font: 'var(--type-overline)',
        letterSpacing: 'var(--ls-caps)',
        textTransform: 'uppercase',
        color: 'var(--fg-3)',
        marginBottom: 12
      }
    }, "Edge legend"), /*#__PURE__*/React.createElement("div", {
      style: {
        display: 'flex',
        flexDirection: 'column',
        gap: 10
      }
    }, [['Confirmed', 'var(--edge-confirmed)', 'none', 1.8], ['Inferred', 'var(--edge-inferred)', '7 6', 1.8], ['Unknown L2', 'var(--edge-unknown)', '2 7', 1.8], ['Route hop', 'var(--edge-confirmed)', 'none', 1], ['Low confidence', 'var(--edge-muted)', '7 6', 1.8]].map((r, i) => /*#__PURE__*/React.createElement("div", {
      key: i,
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 10
      }
    }, /*#__PURE__*/React.createElement("svg", {
      width: 40,
      height: 10,
      style: {
        flex: '0 0 auto'
      }
    }, /*#__PURE__*/React.createElement("line", {
      x1: 1,
      y1: 5,
      x2: 39,
      y2: 5,
      stroke: r[1],
      strokeWidth: r[3],
      strokeDasharray: r[2],
      strokeLinecap: "round"
    })), /*#__PURE__*/React.createElement("span", {
      style: {
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-2)'
      }
    }, r[0]))))))));
  }
  window.IAD_SCREENS = window.IAD_SCREENS || {};
  window.IAD_SCREENS.topology = Topology;
})();
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/console/Topology.jsx", error: String((e && e.message) || e) }); }

// ui_kits/console/data.js
try { (() => {
/* IAD UI kit — a single realistic NormalizedScanReport sample.
   This is a recreation fixture for the kit's screens, NOT production data.
   Models the shape described in docs/design/07-api-design.md: an honest scan
   that knows a lot about the LAN edge but is calibrated about the physical
   medium. Exposed as window.IAD_SCAN. */
window.IAD_SCAN = {
  scan_id: 'scan_01HV6Q8M2K',
  created_at: '2026-06-16T09:41:22Z',
  status: 'complete',
  mode: 'standard',
  // quick | standard | deep
  duration_ms: 6840,
  safe_mode: true,
  // ---- Headline decision -------------------------------------------------
  primary_type: 'Fiber (FTTH)',
  category: 'fixed_broadband',
  confidence: 0.74,
  // overall
  classification_confidence: 0.71,
  // which access type
  context_confidence: 0.83,
  // network context certainty
  decision_quality: 'medium',
  // low | medium | high
  uncertainty_reasons: ['Physical medium cannot be confirmed from inside the LAN — no DSL/DOCSIS modem stats exposed.', 'CPE management interface (192.168.100.1) did not respond to SNMP.', 'Downstream/upstream symmetry is consistent with fiber but also with some cable plans.'],
  candidates: [{
    type: 'Fiber (FTTH)',
    score: 0.74,
    note: 'Low latency, symmetric throughput, fiber-typical jitter floor.'
  }, {
    type: 'Cable (DOCSIS)',
    score: 0.41,
    note: 'Cannot rule out — no modem telemetry to confirm or deny.'
  }, {
    type: 'Fixed Wireless (FWA)',
    score: 0.12,
    note: 'Latency too low and too stable for typical FWA.'
  }, {
    type: 'VDSL',
    score: 0.06,
    note: 'Throughput exceeds VDSL2 profile ceilings.'
  }],
  // ---- Network context ---------------------------------------------------
  detected_network_context: {
    selected_interface: {
      name: 'Ethernet',
      type: 'ethernet',
      ipv4: '192.168.1.24',
      prefix: 24,
      mac: '9C:30:5B:A1:4F:02',
      mtu: 1500,
      gateway: '192.168.1.1',
      dns: ['192.168.1.1', '1.1.1.1']
    },
    link_speed_mbps: 1000,
    dhcp: true
  },
  gateway_chain: [{
    hop: 1,
    ip: '192.168.1.1',
    kind: 'default_gateway',
    rtt_ms: 1.1,
    label: 'Home router',
    private: true
  }, {
    hop: 2,
    ip: '100.64.12.1',
    kind: 'upstream_private_gateway',
    rtt_ms: 8.7,
    label: 'CGNAT gateway',
    private: true,
    note: 'RFC 6598 shared address space'
  }, {
    hop: 3,
    ip: '203.0.113.1',
    kind: 'isp_gateway',
    rtt_ms: 11.4,
    label: 'ISP edge',
    private: false
  }],
  nat_topology: {
    type: 'cgnat',
    layers: 2,
    public_reachable: false,
    note: 'Double NAT detected: local NAT behind carrier-grade NAT (100.64/10).'
  },
  ipv6_context: {
    available: true,
    global_address: '2a01:598:8000::5b2a',
    delegated_prefix: '2a01:598:8000::/56',
    note: 'Native dual-stack; IPv6 path avoids CGNAT.'
  },
  public_ip: {
    address: '203.0.113.42',
    ptr: 'cpe-203-0-113-42.cust.example-isp.net',
    asn: 'AS3320',
    org: 'Example ISP GmbH',
    city: 'Frankfurt',
    country: 'DE',
    geo_confidence: 0.6
  },
  performance: {
    downstream_mbps: 412.6,
    upstream_mbps: 198.3,
    latency_ms: 11.4,
    jitter_ms: 0.8,
    loss_pct: 0.0
  },
  // ---- Confidence breakdown ---------------------------------------------
  confidence_breakdown: [{
    factor: 'Latency & jitter profile',
    weight: 0.28,
    contribution: 0.24,
    direction: 'up',
    detail: 'Sub-12ms, jitter <1ms — fiber-typical.'
  }, {
    factor: 'Symmetric throughput',
    weight: 0.22,
    contribution: 0.16,
    direction: 'up',
    detail: '412↓/198↑ Mbps; high upstream favors fiber.'
  }, {
    factor: 'No modem telemetry',
    weight: 0.20,
    contribution: -0.12,
    direction: 'down',
    detail: 'CPE SNMP blocked; physical layer unverified.'
  }, {
    factor: 'ASN / ISP profile',
    weight: 0.18,
    contribution: 0.11,
    direction: 'up',
    detail: 'AS3320 deploys predominantly fiber in this region.'
  }, {
    factor: 'CGNAT presence',
    weight: 0.12,
    contribution: -0.03,
    direction: 'down',
    detail: 'Common across access types; weak signal.'
  }],
  next_best_probes: [{
    name: 'CPE SNMP walk',
    gain: 0.18,
    requires: 'CPE credentials or read community',
    tier: 'physical',
    detail: 'Would expose DOCSIS/GPON line stats and confirm medium.'
  }, {
    name: 'TR-069 / management VLAN',
    gain: 0.12,
    requires: 'ISP management access',
    tier: 'physical'
  }, {
    name: 'Sustained throughput test',
    gain: 0.06,
    requires: 'User consent (data usage)',
    tier: 'performance'
  }],
  warnings: [{
    level: 'warn',
    text: 'Double NAT (CGNAT) — inbound connections will not reach this host.'
  }, {
    level: 'info',
    text: 'IPv6 is native and bypasses CGNAT; prefer it for reachability.'
  }],
  // ---- Devices (LAN inventory) ------------------------------------------
  devices: [{
    id: 'd-host',
    ip: '192.168.1.24',
    mac: '9C:30:5B:A1:4F:02',
    vendor: 'Dell Inc.',
    hostname: 'WS-OPS-14',
    type: 'local_host',
    role: 'This host',
    reachability: 'self',
    confidence: 1.0,
    source: 'interface',
    services: ['—']
  }, {
    id: 'd-gw',
    ip: '192.168.1.1',
    mac: 'F0:9F:C2:1A:88:E0',
    vendor: 'AVM GmbH',
    hostname: 'fritz.box',
    type: 'default_gateway',
    role: 'Router / NAT',
    reachability: 'reachable',
    confidence: 0.98,
    source: 'arp',
    services: ['DNS', 'HTTP', 'HTTPS']
  }, {
    id: 'd-ap',
    ip: '192.168.1.2',
    mac: 'F0:9F:C2:1A:88:E1',
    vendor: 'AVM GmbH',
    hostname: 'repeater-og',
    type: 'access_point',
    role: 'Mesh Wi-Fi',
    reachability: 'reachable',
    confidence: 0.86,
    source: 'mdns',
    services: ['HTTP']
  }, {
    id: 'd-nas',
    ip: '192.168.1.30',
    mac: '00:11:32:7C:A9:01',
    vendor: 'Synology',
    hostname: 'nas-archive',
    type: 'server',
    role: 'File / NAS',
    reachability: 'reachable',
    confidence: 0.93,
    source: 'mdns',
    services: ['SMB', 'HTTPS', 'NFS']
  }, {
    id: 'd-print',
    ip: '192.168.1.41',
    mac: '3C:2A:F4:11:0D:7B',
    vendor: 'Brother',
    hostname: 'HL-L2350DW',
    type: 'printer',
    role: 'Printer',
    reachability: 'reachable',
    confidence: 0.81,
    source: 'mdns',
    services: ['IPP', 'HTTP']
  }, {
    id: 'd-phone',
    ip: '192.168.1.57',
    mac: 'A4:83:E7:5F:22:9C',
    vendor: 'Apple, Inc.',
    hostname: 'iphone-k',
    type: 'mobile',
    role: 'Mobile',
    reachability: 'reachable',
    confidence: 0.64,
    source: 'arp',
    services: []
  }, {
    id: 'd-iot',
    ip: '192.168.1.88',
    mac: 'D8:A0:11:42:6E:33',
    vendor: 'Espressif',
    hostname: '—',
    type: 'iot',
    role: 'IoT (sensor?)',
    reachability: 'partial',
    confidence: 0.38,
    source: 'arp',
    services: ['?']
  }, {
    id: 'd-unknown',
    ip: '192.168.1.103',
    mac: '5E:B2:9A:00:1F:44',
    vendor: '— (locally administered)',
    hostname: '—',
    type: 'unknown',
    role: 'Unknown',
    reachability: 'partial',
    confidence: 0.22,
    source: 'arp',
    services: []
  }],
  // ---- Evidence / probes -------------------------------------------------
  evidence: [{
    id: 'p-iface',
    probe_name: 'interface_enum',
    status: 'success',
    confidence: 1.0,
    ts: '09:41:16',
    evidence_class: 'l3',
    reason: 'Enumerated active interfaces and addresses.',
    limitations: 'OS-reported; does not confirm physical medium.',
    data: {
      interfaces: 2,
      selected: 'Ethernet',
      ipv4: '192.168.1.24/24',
      mac: '9C:30:5B:A1:4F:02'
    }
  }, {
    id: 'p-arp',
    probe_name: 'arp_sweep',
    status: 'success',
    confidence: 0.92,
    ts: '09:41:17',
    evidence_class: 'l2',
    reason: 'ARP-resolved 8 hosts on local subnet.',
    limitations: 'Only reaches the local broadcast domain.',
    data: {
      subnet: '192.168.1.0/24',
      responded: 8,
      mac_table_seen: false
    }
  }, {
    id: 'p-gw',
    probe_name: 'gateway_trace',
    status: 'success',
    confidence: 0.88,
    ts: '09:41:18',
    evidence_class: 'l3',
    reason: 'Traced 3-hop gateway chain to ISP edge.',
    limitations: 'Hops are L3 routers, not physical switches.',
    data: {
      hops: 3,
      cgnat: true
    }
  }, {
    id: 'p-dns',
    probe_name: 'public_ip_dns',
    status: 'success',
    confidence: 0.9,
    ts: '09:41:19',
    evidence_class: 'isp',
    reason: 'Resolved public IP, PTR, and ASN.',
    limitations: 'Geo from ASN registry; city-level only.',
    data: {
      ip: '203.0.113.42',
      asn: 'AS3320',
      ptr: 'cpe-203-0-113-42.cust.example-isp.net'
    }
  }, {
    id: 'p-snmp',
    probe_name: 'cpe_snmp',
    status: 'blocked',
    confidence: 0.0,
    ts: '09:41:20',
    evidence_class: 'physical',
    reason: 'CPE management host did not respond to SNMP (timeout).',
    limitations: 'Cannot read DOCSIS/GPON line stats; medium stays inferred.',
    data: {
      target: '192.168.100.1',
      community: 'public',
      result: 'timeout'
    }
  }, {
    id: 'p-lldp',
    probe_name: 'lldp_listen',
    status: 'no_data',
    confidence: 0.0,
    ts: '09:41:20',
    evidence_class: 'l2',
    reason: 'No LLDP/CDP frames observed in capture window.',
    limitations: 'Consumer gear rarely emits LLDP; absence is not evidence.',
    data: {
      window_s: 4,
      frames: 0
    }
  }, {
    id: 'p-perf',
    probe_name: 'perf_sample',
    status: 'partial',
    confidence: 0.6,
    ts: '09:41:21',
    evidence_class: 'performance',
    reason: 'Short latency/throughput sample taken.',
    limitations: 'Burst sample, not sustained; throughput is a lower bound.',
    data: {
      down_mbps: 412.6,
      up_mbps: 198.3,
      latency_ms: 11.4
    }
  }, {
    id: 'p-ipv6',
    probe_name: 'ipv6_probe',
    status: 'success',
    confidence: 0.85,
    ts: '09:41:21',
    evidence_class: 'l3',
    reason: 'Native IPv6 GUA + /56 delegation observed.',
    limitations: null,
    data: {
      gua: '2a01:598:8000::5b2a',
      prefix: '/56'
    }
  }, {
    id: 'p-mdns',
    probe_name: 'mdns_discovery',
    status: 'success',
    confidence: 0.78,
    ts: '09:41:21',
    evidence_class: 'l2',
    reason: 'Discovered service records for 4 hosts.',
    limitations: 'mDNS scope is the local link only.',
    data: {
      hosts: 4,
      services: ['_ipp', '_smb', '_http', '_raop']
    }
  }, {
    id: 'p-portscan',
    probe_name: 'service_probe',
    status: 'skipped',
    confidence: 0.0,
    ts: '—',
    evidence_class: 'l3',
    reason: 'Skipped in standard mode (safe mode).',
    limitations: 'Enable deep mode to fingerprint services.',
    data: {
      reason: 'safe_mode'
    }
  }],
  // ---- Topology (conservative, generated from context) -------------------
  topology: {
    generated: true,
    // not natively provided; derived from context + evidence
    layers: {
      l2: true,
      l3: true,
      nat: true,
      isp_route_context: true,
      unknown: true
    }
  }
};
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/console/data.js", error: String((e && e.message) || e) }); }

// ui_kits/console/format.js
try { (() => {
/* IAD UI kit — formatting helpers. window.fmt */
window.fmt = {
  pct: v => Math.round(v * 100) + '%',
  band: v => v >= 0.75 ? 'high' : v >= 0.45 ? 'medium' : 'low',
  bandWord: v => v >= 0.75 ? 'High' : v >= 0.45 ? 'Medium' : 'Low',
  bandColor: v => v >= 0.75 ? 'var(--conf-high)' : v >= 0.45 ? 'var(--conf-med)' : 'var(--conf-low)',
  bandBg: v => v >= 0.75 ? 'var(--conf-high-bg)' : v >= 0.45 ? 'var(--conf-med-bg)' : 'var(--conf-low-bg)',
  time: iso => {
    try {
      return new Date(iso).toLocaleString('en-GB', {
        hour: '2-digit',
        minute: '2-digit',
        day: '2-digit',
        month: 'short'
      });
    } catch (e) {
      return iso;
    }
  },
  ago: iso => {
    const s = (Date.now() - new Date(iso).getTime()) / 1000;
    if (s < 60) return 'just now';
    if (s < 3600) return Math.floor(s / 60) + 'm ago';
    if (s < 86400) return Math.floor(s / 3600) + 'h ago';
    return Math.floor(s / 86400) + 'd ago';
  },
  reach: r => ({
    self: {
      word: 'This host',
      tone: 'accent'
    },
    reachable: {
      word: 'Reachable',
      tone: 'success'
    },
    partial: {
      word: 'Partial',
      tone: 'warn'
    },
    unreachable: {
      word: 'Unreachable',
      tone: 'danger'
    }
  })[r] || {
    word: r,
    tone: 'neutral'
  },
  deviceIcon: t => ({
    local_host: 'host',
    default_gateway: 'router',
    router: 'router',
    modem_cpe: 'modem',
    access_point: 'ap',
    mesh_node: 'ap',
    managed_switch: 'switchIcon',
    server: 'server',
    printer: 'printer',
    mobile: 'mobile',
    workstation: 'host',
    iot: 'iot',
    dns_server: 'server',
    isp_gateway: 'globe',
    public_internet: 'globe',
    unknown: 'unknown'
  })[t] || 'unknown'
};
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/console/format.js", error: String((e && e.message) || e) }); }

// ui_kits/console/icons.jsx
try { (() => {
/* IAD icon set — inline SVG, Lucide-style (24px grid, 2px stroke, round caps).
   Lucide is not bundled with the planned Tauri/React app, so the kit ships a
   small hand-picked subset matching Lucide's geometry. Exposed as window.Icons.
   Each is a function component taking {size, ...props}; stroke = currentColor. */
(function () {
  const h = React.createElement;
  function mk(paths) {
    return function Icon({
      size = 18,
      style,
      ...rest
    }) {
      return h('svg', {
        viewBox: '0 0 24 24',
        width: size,
        height: size,
        fill: 'none',
        stroke: 'currentColor',
        strokeWidth: 2,
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        style: {
          display: 'block',
          ...style
        },
        ...rest
      }, paths.map((d, i) => h('path', {
        key: i,
        d
      })));
    };
  }
  function mkRaw(children) {
    return function Icon({
      size = 18,
      style,
      ...rest
    }) {
      return h('svg', {
        viewBox: '0 0 24 24',
        width: size,
        height: size,
        fill: 'none',
        stroke: 'currentColor',
        strokeWidth: 2,
        strokeLinecap: 'round',
        strokeLinejoin: 'round',
        style: {
          display: 'block',
          ...style
        },
        ...rest
      }, children(h));
    };
  }
  window.Icons = {
    // nav
    dashboard: mk(['M3 3h7v7H3zM14 3h7v7h-7zM14 14h7v7h-7zM3 14h7v7H3z']),
    topology: mkRaw(h => [h('circle', {
      key: 1,
      cx: 5,
      cy: 6,
      r: 2.4
    }), h('circle', {
      key: 2,
      cx: 19,
      cy: 6,
      r: 2.4
    }), h('circle', {
      key: 3,
      cx: 12,
      cy: 18,
      r: 2.4
    }), h('path', {
      key: 4,
      d: 'M7 6h10M6.5 8 11 16M17.5 8 13 16'
    })]),
    devices: mk(['M3 5h18v11H3z', 'M8 21h8M12 16v5']),
    evidence: mk(['M9 3h6l3 4v14H6V3z', 'M9 12h6M9 16h4']),
    reports: mk(['M14 3v5h5', 'M14 3H6v18h12V8z', 'M9 13h6M9 17h6']),
    settings: mkRaw(h => [h('circle', {
      key: 1,
      cx: 12,
      cy: 12,
      r: 3
    }), h('path', {
      key: 2,
      d: 'M19.4 13.5a1.7 1.7 0 0 0 .3 1.9l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-2.9 1.2V21a2 2 0 1 1-4 0v-.2a1.7 1.7 0 0 0-2.9-1.2l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0-1.2-2.9H3a2 2 0 1 1 0-4h.2a1.7 1.7 0 0 0 1.2-2.9l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 2.9-1.2V3a2 2 0 1 1 4 0v.2a1.7 1.7 0 0 0 2.9 1.2l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.9Z'
    })]),
    // actions
    refresh: mk(['M21 12a9 9 0 1 1-3-6.7L21 8', 'M21 3v5h-5']),
    download: mk(['M12 3v12', 'm7 12 5 5 5-5', 'M5 21h14']),
    upload: mk(['M12 21V9', 'm7 12 5-5 5 5', 'M5 3h14']),
    copy: mkRaw(h => [h('rect', {
      key: 1,
      x: 9,
      y: 9,
      width: 12,
      height: 12,
      rx: 2
    }), h('path', {
      key: 2,
      d: 'M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1'
    })]),
    search: mkRaw(h => [h('circle', {
      key: 1,
      cx: 11,
      cy: 11,
      r: 7
    }), h('path', {
      key: 2,
      d: 'm21 21-4.3-4.3'
    })]),
    filter: mk(['M3 5h18l-7 8v6l-4-2v-4z']),
    close: mk(['M18 6 6 18M6 6l12 12']),
    chevronRight: mk(['m9 6 6 6-6 6']),
    chevronDown: mk(['m6 9 6 6 6-6']),
    external: mk(['M15 3h6v6', 'M10 14 21 3', 'M21 14v5a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5']),
    // topology controls
    zoomIn: mkRaw(h => [h('circle', {
      key: 1,
      cx: 11,
      cy: 11,
      r: 7
    }), h('path', {
      key: 2,
      d: 'm21 21-4.3-4.3M11 8v6M8 11h6'
    })]),
    zoomOut: mkRaw(h => [h('circle', {
      key: 1,
      cx: 11,
      cy: 11,
      r: 7
    }), h('path', {
      key: 2,
      d: 'm21 21-4.3-4.3M8 11h6'
    })]),
    fit: mk(['M3 8V5a2 2 0 0 1 2-2h3', 'M21 8V5a2 2 0 0 0-2-2h-3', 'M3 16v3a2 2 0 0 0 2 2h3', 'M21 16v3a2 2 0 0 1-2 2h-3']),
    reset: mk(['M3 12a9 9 0 1 0 9-9 9 9 0 0 0-6.4 2.6L3 8', 'M3 3v5h5']),
    layers: mk(['m12 2 9 5-9 5-9-5z', 'm3 12 9 5 9-5', 'm3 17 9 5 9-5']),
    lock: mkRaw(h => [h('rect', {
      key: 1,
      x: 4,
      y: 11,
      width: 16,
      height: 10,
      rx: 2
    }), h('path', {
      key: 2,
      d: 'M8 11V7a4 4 0 0 1 8 0v4'
    })]),
    // node / device types
    host: mkRaw(h => [h('rect', {
      key: 1,
      x: 3,
      y: 4,
      width: 18,
      height: 12,
      rx: 2
    }), h('path', {
      key: 2,
      d: 'M8 20h8M12 16v4'
    })]),
    router: mkRaw(h => [h('rect', {
      key: 1,
      x: 2,
      y: 13,
      width: 20,
      height: 7,
      rx: 2
    }), h('path', {
      key: 2,
      d: 'M6 17h.01M10 17h.01M14 8l2-2 2 2M16 6v7'
    })]),
    modem: mkRaw(h => [h('rect', {
      key: 1,
      x: 2,
      y: 6,
      width: 20,
      height: 12,
      rx: 2
    }), h('path', {
      key: 2,
      d: 'M6 18v2M18 18v2M6 10h.01M10 10h.01'
    })]),
    ap: mkRaw(h => [h('path', {
      key: 1,
      d: 'M5 12.5a7 7 0 0 1 14 0M8 15a4 4 0 0 1 8 0'
    }), h('circle', {
      key: 2,
      cx: 12,
      cy: 18,
      r: 1.4
    })]),
    switchIcon: mkRaw(h => [h('rect', {
      key: 1,
      x: 3,
      y: 8,
      width: 18,
      height: 8,
      rx: 2
    }), h('path', {
      key: 2,
      d: 'M7 12h.01M11 12h.01M15 12h.01'
    })]),
    server: mkRaw(h => [h('rect', {
      key: 1,
      x: 3,
      y: 4,
      width: 18,
      height: 7,
      rx: 2
    }), h('rect', {
      key: 2,
      x: 3,
      y: 13,
      width: 18,
      height: 7,
      rx: 2
    }), h('path', {
      key: 3,
      d: 'M7 7.5h.01M7 16.5h.01'
    })]),
    printer: mkRaw(h => [h('path', {
      key: 1,
      d: 'M6 9V3h12v6'
    }), h('rect', {
      key: 2,
      x: 4,
      y: 9,
      width: 16,
      height: 7,
      rx: 2
    }), h('path', {
      key: 3,
      d: 'M7 16h10v5H7z'
    })]),
    mobile: mkRaw(h => [h('rect', {
      key: 1,
      x: 7,
      y: 2,
      width: 10,
      height: 20,
      rx: 2
    }), h('path', {
      key: 2,
      d: 'M11 18h2'
    })]),
    iot: mkRaw(h => [h('circle', {
      key: 1,
      cx: 12,
      cy: 12,
      r: 3
    }), h('path', {
      key: 2,
      d: 'M12 2v4M12 18v4M2 12h4M18 12h4'
    })]),
    globe: mkRaw(h => [h('circle', {
      key: 1,
      cx: 12,
      cy: 12,
      r: 9
    }), h('path', {
      key: 2,
      d: 'M3 12h18M12 3a14 14 0 0 1 0 18M12 3a14 14 0 0 0 0 18'
    })]),
    unknown: mkRaw(h => [h('circle', {
      key: 1,
      cx: 12,
      cy: 12,
      r: 9
    }), h('path', {
      key: 2,
      d: 'M9.2 9a3 3 0 0 1 5.6 1c0 2-3 2.5-3 4'
    }), h('path', {
      key: 3,
      d: 'M12 17h.01'
    })]),
    // status / misc
    alert: mkRaw(h => [h('path', {
      key: 1,
      d: 'M12 3 2 20h20z'
    }), h('path', {
      key: 2,
      d: 'M12 10v4M12 17h.01'
    })]),
    info: mkRaw(h => [h('circle', {
      key: 1,
      cx: 12,
      cy: 12,
      r: 9
    }), h('path', {
      key: 2,
      d: 'M12 11v5M12 8h.01'
    })]),
    check: mk(['M20 6 9 17l-5-5']),
    shield: mk(['M12 3 5 6v5c0 4 3 7 7 9 4-2 7-5 7-9V6z', 'M9.5 12l2 2 3.5-4']),
    plug: mk(['M9 2v6M15 2v6', 'M7 8h10v3a5 5 0 0 1-10 0z', 'M12 16v6']),
    activity: mk(['M3 12h4l3 8 4-16 3 8h4'])
  };
})();
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/console/icons.jsx", error: String((e && e.message) || e) }); }

__ds_ns.Badge = __ds_scope.Badge;

__ds_ns.Button = __ds_scope.Button;

__ds_ns.Card = __ds_scope.Card;

__ds_ns.IconButton = __ds_scope.IconButton;

__ds_ns.StatusDot = __ds_scope.StatusDot;

__ds_ns.ConfidenceBar = __ds_scope.ConfidenceBar;

__ds_ns.MetricStat = __ds_scope.MetricStat;

__ds_ns.ProbeStatusBadge = __ds_scope.ProbeStatusBadge;

__ds_ns.TierBadge = __ds_scope.TierBadge;

__ds_ns.Input = __ds_scope.Input;

__ds_ns.SegmentedControl = __ds_scope.SegmentedControl;

__ds_ns.Toggle = __ds_scope.Toggle;

})();
