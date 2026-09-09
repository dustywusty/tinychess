export type Theme = "dark" | "light";
export const ACCENTS = ["#83a58b", "#a597c4", "#cf967d", "#4f9c98", "#7f9fc3", "#c58f9e"] as const;
export type Accent = (typeof ACCENTS)[number];
export const BOARD_THEMES = [
  { accent: ACCENTS[0], name: "Matcha", light: "#edf1de", dark: "#83a58b" },
  { accent: ACCENTS[1], name: "Lilac", light: "#f0eaf8", dark: "#a597c4" },
  { accent: ACCENTS[2], name: "Peach", light: "#faeddd", dark: "#cf967d" },
  { accent: ACCENTS[3], name: "Teal", light: "#e2f2ed", dark: "#4f9c98" },
  { accent: ACCENTS[4], name: "Sky", light: "#e7eff8", dark: "#7f9fc3" },
  { accent: ACCENTS[5], name: "Rose", light: "#f8e7ed", dark: "#c58f9e" },
];
export function loadTheme(): Theme {
  try { if (localStorage.getItem("theme") === "dark") return "dark"; } catch { /* Storage is optional. */ }
  return "light";
}
export function loadAccent(): string {
  try {
    const saved = localStorage.getItem("accent")?.toLowerCase();
    if (BOARD_THEMES.some((theme) => theme.accent === saved)) return saved!;
    if (saved === "#a78bfa") return ACCENTS[1];
    if (saved === "#f472b6" || saved === "#f59e0b") return ACCENTS[2];
  } catch { /* Storage is optional. */ }
  return ACCENTS[0];
}
export function applyTheme(theme: Theme, accent: string): void {
  const palette = BOARD_THEMES.find((item) => item.accent === accent) ?? BOARD_THEMES[0];
  document.documentElement.setAttribute("data-theme", theme);
  document.documentElement.style.setProperty("--accent", palette.accent);
  document.documentElement.style.setProperty("--sq1", palette.light);
  document.documentElement.style.setProperty("--sq2", palette.dark);
  try {
    localStorage.setItem("theme", theme);
    localStorage.setItem("accent", palette.accent);
  } catch { /* Storage is optional. */ }
}
