export type Theme = "dark" | "light";

export const ACCENTS = [
  "#6ee7ff",
  "#a78bfa",
  "#f472b6",
  "#f59e0b",
  "#10b981",
] as const;

export type Accent = (typeof ACCENTS)[number];

const THEME_KEY = "theme";
const ACCENT_KEY = "accent";

export function loadTheme(): Theme {
  try {
    const raw = localStorage.getItem(THEME_KEY);
    if (raw === "light") return "light";
  } catch {
    /* ignore */
  }
  return "dark";
}

export function loadAccent(): string {
  try {
    const raw = localStorage.getItem(ACCENT_KEY);
    if (raw && /^#[0-9a-f]{6}$/i.test(raw)) return raw;
  } catch {
    /* ignore */
  }
  return ACCENTS[0];
}

export function applyTheme(theme: Theme, accent: string): void {
  document.documentElement.setAttribute("data-theme", theme);
  document.documentElement.style.setProperty("--accent", accent);
  try {
    localStorage.setItem(THEME_KEY, theme);
    localStorage.setItem(ACCENT_KEY, accent);
  } catch {
    /* ignore */
  }
}
