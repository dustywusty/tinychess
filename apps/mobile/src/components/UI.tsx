import type { ReactNode } from "react";
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from "react-native";
import { boardThemes, colors, useBoardTheme, type BoardTheme } from "@/lib/theme";

export function Button({ title, onPress, disabled = false, busy = false }: {
  title: string; onPress: () => void; disabled?: boolean; busy?: boolean;
}) {
  return <Pressable accessibilityRole="button" accessibilityState={{ disabled: disabled || busy, busy }} disabled={disabled || busy} onPress={onPress}
    style={({ pressed }) => [ui.button, (disabled || busy) && { opacity: 0.5 }, pressed && { opacity: 0.75 }]}>
    {busy ? <ActivityIndicator color={colors.ink} /> : <Text style={ui.buttonText}>{title}</Text>}
  </Pressable>;
}
export function ThemePicker() {
  const { theme, setTheme } = useBoardTheme();
  return <View style={ui.swatches}>{(Object.keys(boardThemes) as BoardTheme[]).map((key) => (
    <Pressable key={key} accessibilityRole="radio" accessibilityLabel={`${boardThemes[key].name} board`} aria-checked={theme === key} accessibilityState={{ checked: theme === key }}
      onPress={() => setTheme(key)} style={ui.swatchTarget}>
      <View style={[ui.swatch, { backgroundColor: boardThemes[key].dark }, theme === key && ui.swatchSelected]}>
        {theme === key && <Text style={ui.check}>✓</Text>}
      </View>
    </Pressable>
  ))}</View>;
}
export function CoachCard() {
  return <View style={ui.coach}>
    <View style={ui.row}><Text style={ui.eyebrow}>A LITTLE WISDOM</Text><View style={ui.badge}><Text style={ui.badgeText}>COMING SOON</Text></View></View>
    <View style={[ui.row, { gap: 16 }]}>
      <View style={ui.coachIcon}><Text style={{ fontSize: 28, color: colors.ink }}>✳</Text></View>
      <View style={{ flex: 1 }}><Text style={ui.cardTitle}>Meet your chess coach.</Text><Text style={ui.body}>A nudge when you need it. A little more “aha” in every game.</Text></View>
    </View>
  </View>;
}
export function ErrorMessage({ children }: { children: ReactNode }) {
  return <Text accessibilityRole="alert" accessibilityLiveRegion="polite" style={ui.error}>{children}</Text>;
}
export const ui = StyleSheet.create({
  row: { flexDirection: "row", alignItems: "center", justifyContent: "space-between" },
  eyebrow: { fontSize: 10, fontWeight: "700", letterSpacing: 1.6, color: colors.muted },
  body: { fontSize: 14, lineHeight: 21, color: colors.muted },
  cardTitle: { fontSize: 18, fontWeight: "600", letterSpacing: -0.5, color: colors.ink, marginBottom: 5 },
  button: { minHeight: 56, borderRadius: 18, backgroundColor: colors.mint, alignItems: "center", justifyContent: "center", paddingHorizontal: 20 },
  buttonText: { fontSize: 16, fontWeight: "700", color: colors.ink },
  swatches: { flexDirection: "row" },
  swatchTarget: { width: 44, height: 44, justifyContent: "center", alignItems: "center" },
  swatch: { width: 25, height: 25, borderRadius: 20, alignItems: "center", justifyContent: "center" },
  swatchSelected: { borderWidth: 2, borderColor: colors.ink },
  check: { color: "#fff", fontSize: 12, fontWeight: "800" },
  coach: { borderRadius: 24, backgroundColor: colors.lilac, padding: 20, gap: 18 },
  coachIcon: { width: 48, height: 48, borderRadius: 16, backgroundColor: "#F5F1FF", alignItems: "center", justifyContent: "center" },
  badge: { backgroundColor: "#F5F1FF", borderRadius: 6, paddingHorizontal: 7, paddingVertical: 5 },
  badgeText: { fontSize: 8, fontWeight: "800", color: "#665A7C", letterSpacing: 1 },
  error: { padding: 14, backgroundColor: colors.errorBg, color: colors.error, borderRadius: 14, lineHeight: 20, fontSize: 13 },
});
