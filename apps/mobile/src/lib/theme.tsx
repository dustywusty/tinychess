import AsyncStorage from "@react-native-async-storage/async-storage";
import { createContext, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from "react";

const lightColors = {
  background: "#F8F7F2", surface: "#FFFFFF", ink: "#252B28", muted: "#727970",
  line: "#E3E6DE", mint: "#D9F38D", lilac: "#E6DFFA", coral: "#F7C5AF",
  error: "#A1362C", errorBg: "#FCEAE4",
  soft: "#EEF0E7", buttonInk: "#252B28", kicker: "#61764D", coachIcon: "#F5F1FF",
  coachInk: "#4B435A", badgeInk: "#665A7C", turnBg: "#EAF0DE", turnInk: "#536943",
};
export type AppColors = typeof lightColors;
export type ColorMode = "light" | "dark";
const darkColors: AppColors = {
  ...lightColors, background: "#1D2521", surface: "#27332C", ink: "#F3F4E9", muted: "#B2BCAE",
  line: "#3F4D41", soft: "#2E3C31", lilac: "#494054", kicker: "#BDD49C",
  error: "#FFC4B5", errorBg: "#51322B", coachIcon: "#625570", coachInk: "#EEE6FA",
  badgeInk: "#EEE6FA", turnBg: "#35462F", turnInk: "#CCE0B1",
};
export const boardThemes = {
  mint: { name: "Matcha", light: "#EDF1DE", dark: "#83A58B", selected: "#D7EE89" },
  lilac: { name: "Lilac", light: "#F0EAF8", dark: "#A597C4", selected: "#E4EE9D" },
  coral: { name: "Peach", light: "#FAEDDD", dark: "#CF967D", selected: "#E4EE9D" },
  teal: { name: "Teal", light: "#E2F2ED", dark: "#4F9C98", selected: "#D7EE89" },
  sky: { name: "Sky", light: "#E7EFF8", dark: "#7F9FC3", selected: "#E4EE9D" },
  rose: { name: "Rose", light: "#F8E7ED", dark: "#C58F9E", selected: "#E4EE9D" },
};
export type BoardTheme = keyof typeof boardThemes;
const ThemeContext = createContext({ theme: "mint" as BoardTheme, setTheme: (_: BoardTheme) => {},
  mode: "light" as ColorMode, setMode: (_: ColorMode) => {}, colors: lightColors });
export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setValue] = useState<BoardTheme>("mint");
  const changed = useRef(false);
  const [mode, setModeValue] = useState<ColorMode>("light");
  const modeChanged = useRef(false);
  useEffect(() => {
    void AsyncStorage.getItem("yourmove.board-theme").then((saved) => {
      if (!changed.current && saved && Object.hasOwn(boardThemes, saved)) setValue(saved as BoardTheme);
    }).catch(() => {});
    void AsyncStorage.getItem("yourmove.color-mode").then((saved) => {
      if (!modeChanged.current && (saved === "light" || saved === "dark")) setModeValue(saved);
    }).catch(() => {});
  }, []);
  const setTheme = (value: BoardTheme) => {
    changed.current = true;
    setValue(value);
    void AsyncStorage.setItem("yourmove.board-theme", value).catch(() => {});
  };
  const setMode = (value: ColorMode) => {
    modeChanged.current = true;
    setModeValue(value);
    void AsyncStorage.setItem("yourmove.color-mode", value).catch(() => {});
  };
  return <ThemeContext.Provider value={{ theme, setTheme, mode, setMode, colors: mode === "dark" ? darkColors : lightColors }}>{children}</ThemeContext.Provider>;
}
export const useBoardTheme = () => useContext(ThemeContext);
export function useThemedStyles<T>(factory: (colors: AppColors) => T): T {
  const { colors } = useBoardTheme();
  return useMemo(() => factory(colors), [factory, colors]);
}
