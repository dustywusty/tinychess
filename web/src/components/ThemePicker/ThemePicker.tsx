import { useEffect, useRef, useState } from "react";
import { useUiStore } from "../../state/uiStore";
import { ACCENTS } from "../../lib/theme";

export function ThemePicker() {
  const theme = useUiStore((s) => s.theme);
  const accent = useUiStore((s) => s.accent);
  const setTheme = useUiStore((s) => s.setTheme);
  const setAccent = useUiStore((s) => s.setAccent);

  const [open, setOpen] = useState(false);
  const wrapRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!open) return;
    const onClick = (e: MouseEvent) => {
      if (!wrapRef.current?.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    window.addEventListener("mousedown", onClick);
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("mousedown", onClick);
      window.removeEventListener("keydown", onKey);
    };
  }, [open]);

  return (
    <div className="flex items-center gap-2">
      <div className="relative" ref={wrapRef}>
        <button
          type="button"
          aria-label="Accent color"
          aria-haspopup="true"
          aria-expanded={open}
          className="w-5 h-5 rounded-full border border-text"
          style={{ background: accent }}
          onClick={() => setOpen((v) => !v)}
        />
        {open && (
          <div
            role="menu"
            className="absolute left-1/2 -translate-x-1/2 top-full mt-2 flex flex-col items-center gap-1.5 p-1.5 rounded-md bg-panel border border-[color:var(--btn-border,_rgba(255,255,255,0.1))] shadow-lg z-40"
          >
            {ACCENTS.map((c) => (
              <button
                key={c}
                type="button"
                role="menuitemradio"
                aria-label={`Accent ${c}`}
                aria-checked={c === accent}
                className={`w-5 h-5 rounded-full border ${
                  c === accent ? "border-text" : "border-transparent"
                }`}
                style={{ background: c }}
                onClick={() => {
                  setAccent(c);
                  setOpen(false);
                }}
              />
            ))}
          </div>
        )}
      </div>
      <button
        type="button"
        className="text-xs px-2 py-1 rounded bg-panel border border-[color:var(--btn-border,_rgba(255,255,255,0.1))]"
        onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
        aria-label="Toggle theme"
      >
        {theme === "dark" ? "☾" : "☀"}
      </button>
    </div>
  );
}
