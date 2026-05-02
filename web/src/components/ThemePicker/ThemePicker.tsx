import { useUiStore } from "../../state/uiStore";
import { ACCENTS } from "../../lib/theme";

export function ThemePicker() {
  const theme = useUiStore((s) => s.theme);
  const accent = useUiStore((s) => s.accent);
  const setTheme = useUiStore((s) => s.setTheme);
  const setAccent = useUiStore((s) => s.setAccent);

  return (
    <div className="flex items-center gap-2">
      <div className="flex items-center gap-1" aria-label="Accent">
        {ACCENTS.map((c) => (
          <button
            key={c}
            type="button"
            aria-label={`Accent ${c}`}
            aria-pressed={c === accent}
            className={`w-5 h-5 rounded-full border ${
              c === accent ? "border-text" : "border-transparent"
            }`}
            style={{ background: c }}
            onClick={() => setAccent(c)}
          />
        ))}
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
