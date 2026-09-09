import { useEffect, useId, useRef, useState } from "react";
import { useUiStore } from "../../state/uiStore";
import { BOARD_THEMES } from "../../lib/theme";

export function ThemePicker() {
  const { theme, accent, setTheme, setAccent } = useUiStore();
  const [open, setOpen] = useState(false);
  const root = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const panelId = useId();
  useEffect(() => {
    if (!open) return;
    const outside = (event: PointerEvent) => {
      if (!root.current?.contains(event.target as Node)) setOpen(false);
    };
    const escape = (event: KeyboardEvent) => {
      if (event.key === "Escape") { setOpen(false); trigger.current?.focus(); }
    };
    document.addEventListener("pointerdown", outside);
    document.addEventListener("keydown", escape);
    return () => { document.removeEventListener("pointerdown", outside); document.removeEventListener("keydown", escape); };
  }, [open]);
  return <div className="appearance" ref={root} onBlur={(event) => {
    if (!event.currentTarget.contains(event.relatedTarget)) setOpen(false);
  }}>
    <button ref={trigger} type="button" className="appearance-trigger" aria-label="Appearance" title="Appearance"
      aria-expanded={open} aria-controls={panelId} onClick={() => setOpen(!open)}>
      <span className="appearance-mark" aria-hidden="true">{[accent, "#e6dffa", "#f7c5af", "var(--text)"].map((color, index) => <i key={index} style={{ background: color }} />)}</span>
    </button>
    {open && <section id={panelId} aria-label="Appearance settings" className="appearance-panel">
    <div className="appearance-heading"><h2>Make it yours.</h2><button type="button" className="appearance-close" aria-label="Close appearance" onClick={() => { setOpen(false); trigger.current?.focus(); }}>×</button></div>
    <p className="eyebrow">BOARD COLOR</p>
    <div className="swatches" role="group" aria-label="Board color">
      {BOARD_THEMES.map((item) => <button key={item.accent} type="button"
        aria-label={item.name + " board"} aria-pressed={item.accent === accent}
        className="swatch-target" onClick={() => setAccent(item.accent)}>
        <span className="swatch" style={{ background: item.accent }}>{item.accent === accent && "✓"}</span><span>{item.name}</span>
      </button>)}
    </div>
    <p className="eyebrow">APPEARANCE</p>
    <div className="mode-selector" role="group" aria-label="Color mode">{(["light", "dark"] as const).map((mode) => <button key={mode} type="button" aria-label={`Use ${mode} theme`} aria-pressed={theme === mode} onClick={() => setTheme(mode)}><span aria-hidden="true">{mode === "light" ? "☀" : "☾"}</span>{mode === "light" ? "Light" : "Dark"}</button>)}</div>
    </section>}
  </div>;
}
