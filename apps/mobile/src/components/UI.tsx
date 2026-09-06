import type { ReactNode } from "react";
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from "react-native";
import { useBoardTheme, useThemedStyles, type AppColors } from "@/lib/theme";

export function Button({ title, onPress, disabled = false, busy = false }: {
  title: string; onPress: () => void; disabled?: boolean; busy?: boolean;
}) {
  const { colors } = useBoardTheme();
  const ui = useUI();
  return <Pressable accessibilityRole="button" accessibilityState={{ disabled: disabled || busy, busy }} disabled={disabled || busy} onPress={onPress}
    style={({ pressed }) => [ui.button, (disabled || busy) && { opacity: 0.5 }, pressed && { opacity: 0.75 }]}>
    {busy ? <ActivityIndicator color={colors.buttonInk} /> : <Text style={ui.buttonText}>{title}</Text>}
  </Pressable>;
}
export function CoachCard() {
  const { colors } = useBoardTheme();
  const ui = useUI();
  return <View style={ui.coach}>
    <View style={ui.row}><Text style={[ui.eyebrow, { color: colors.coachInk }]}>A LITTLE WISDOM</Text><View style={ui.badge}><Text style={ui.badgeText}>COMING SOON</Text></View></View>
    <View style={[ui.row, { gap: 16 }]}>
      <View style={ui.coachIcon}><Text style={{ fontSize: 28, color: colors.coachInk }}>✳</Text></View>
      <View style={{ flex: 1 }}><Text style={[ui.cardTitle, { color: colors.coachInk }]}>Meet your chess coach.</Text><Text style={[ui.body, { color: colors.coachInk }]}>A nudge when you need it. A little more “aha” in every game.</Text></View>
    </View>
  </View>;
}
export function ErrorMessage({ children }: { children: ReactNode }) {
  const ui = useUI();
  return <Text accessibilityRole="alert" accessibilityLiveRegion="polite" style={ui.error}>{children}</Text>;
}
export const useUI = () => useThemedStyles(createStyles);
const createStyles = (colors: AppColors) => StyleSheet.create({
  row: { flexDirection: "row", alignItems: "center", justifyContent: "space-between" },
  eyebrow: { fontSize: 10, fontWeight: "700", letterSpacing: 1.6, color: colors.muted },
  body: { fontSize: 14, lineHeight: 21, color: colors.muted },
  cardTitle: { fontSize: 18, fontWeight: "600", letterSpacing: -0.5, color: colors.ink, marginBottom: 5 },
  button: { minHeight: 56, borderRadius: 18, backgroundColor: colors.mint, alignItems: "center", justifyContent: "center", paddingHorizontal: 20 },
  buttonText: { fontSize: 16, fontWeight: "700", color: colors.buttonInk },
  coach: { borderRadius: 24, backgroundColor: colors.lilac, padding: 20, gap: 18 },
  coachIcon: { width: 48, height: 48, borderRadius: 16, backgroundColor: colors.coachIcon, alignItems: "center", justifyContent: "center" },
  badge: { backgroundColor: colors.coachIcon, borderRadius: 6, paddingHorizontal: 7, paddingVertical: 5 },
  badgeText: { fontSize: 8, fontWeight: "800", color: colors.badgeInk, letterSpacing: 1 },
  error: { padding: 14, backgroundColor: colors.errorBg, color: colors.error, borderRadius: 14, lineHeight: 20, fontSize: 13 },
});
