import AsyncStorage from "@react-native-async-storage/async-storage";
import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from "react";

export const colors = {
  background: "#F8F7F2", surface: "#FFFFFF", ink: "#252B28", muted: "#727970",
  line: "#E3E6DE", mint: "#D9F38D", lilac: "#E6DFFA", coral: "#F7C5AF",
  error: "#A1362C", errorBg: "#FCEAE4",
};
export const boardThemes = {
  mint: { name: "Matcha", light: "#EDF1DE", dark: "#83A58B", selected: "#D7EE89" },
  lilac: { name: "Lilac", light: "#F0EAF8", dark: "#A597C4", selected: "#E4EE9D" },
  coral: { name: "Peach", light: "#FAEDDD", dark: "#CF967D", selected: "#E4EE9D" },
};
export type BoardTheme = keyof typeof boardThemes;
const ThemeContext = createContext({ theme: "mint" as BoardTheme, setTheme: (_: BoardTheme) => {} });
export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setValue] = useState<BoardTheme>("mint");
  const changed = useRef(false);
  useEffect(() => {
    void AsyncStorage.getItem("yourmove.board-theme").then((saved) => {
      if (!changed.current && saved && Object.hasOwn(boardThemes, saved)) setValue(saved as BoardTheme);
    }).catch(() => {});
  }, []);
  const setTheme = (value: BoardTheme) => {
    changed.current = true;
    setValue(value);
    void AsyncStorage.setItem("yourmove.board-theme", value).catch(() => {});
  };
  return <ThemeContext.Provider value={{ theme, setTheme }}>{children}</ThemeContext.Provider>;
}
export const useBoardTheme = () => useContext(ThemeContext);
