import { useUiStore } from "../../state/uiStore";
import { BOARD_THEMES } from "../../lib/theme";

export function ThemePicker() {
  const { theme, accent, setTheme, setAccent } = useUiStore();
  return <div className="theme-picker">
    <div className="swatches" role="group" aria-label="Board color">
      {BOARD_THEMES.map((item) => <button key={item.accent} type="button"
        aria-label={item.name + " board"} aria-pressed={item.accent === accent}
        className="swatch-target" onClick={() => setAccent(item.accent)}>
        <span className="swatch" style={{ background: item.accent }}>{item.accent === accent && "✓"}</span>
      </button>)}
    </div>
    <button type="button" className="theme-toggle" onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
      aria-label={theme === "dark" ? "Use light theme" : "Use dark theme"} title={theme === "dark" ? "Use light theme" : "Use dark theme"}>{theme === "dark" ? "☀" : "☾"}</button>
  </div>;
}
