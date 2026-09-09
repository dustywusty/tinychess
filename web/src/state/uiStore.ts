import { create } from "zustand";
import type { Square } from "../types/chess";
import { applyTheme, loadAccent, loadTheme, type Theme } from "../lib/theme";

export interface UiStore {
  selected: Square | null;
  emojiPickerOpen: boolean;
  menuOpen: boolean;
  theme: Theme;
  accent: string;
  // Last error/info message shown in #status
  status: string;
  statusError: boolean;
  // actions
  selectSquare: (sq: Square | null) => void;
  openEmojiPicker: () => void;
  closeEmojiPicker: () => void;
  toggleMenu: () => void;
  closeMenu: () => void;
  setTheme: (t: Theme) => void;
  setAccent: (a: string) => void;
  setStatus: (msg: string, isError?: boolean) => void;
  clearStatus: () => void;
}

const initialTheme = (() => {
  if (typeof window === "undefined") return "dark" as Theme;
  return loadTheme();
})();

const initialAccent = (() => {
  if (typeof window === "undefined") return "#6ee7ff";
  return loadAccent();
})();

if (typeof window !== "undefined") {
  applyTheme(initialTheme, initialAccent);
}

export const useUiStore = create<UiStore>((set) => ({
  selected: null,
  emojiPickerOpen: false,
  menuOpen: false,
  theme: initialTheme,
  accent: initialAccent,
  status: "",
  statusError: false,
  selectSquare: (sq) => set({ selected: sq }),
  openEmojiPicker: () => set({ emojiPickerOpen: true }),
  closeEmojiPicker: () => set({ emojiPickerOpen: false }),
  toggleMenu: () => set((s) => ({ menuOpen: !s.menuOpen })),
  closeMenu: () => set({ menuOpen: false }),
  setTheme: (t) =>
    set((s) => {
      applyTheme(t, s.accent);
      return { theme: t };
    }),
  setAccent: (a) =>
    set((s) => {
      applyTheme(s.theme, a);
      return { accent: a };
    }),
  setStatus: (msg, isError = false) =>
    set({ status: msg, statusError: isError }),
  clearStatus: () => set({ status: "", statusError: false }),
}));
