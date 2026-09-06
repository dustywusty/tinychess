import { useCallback, useEffect, useRef, useState } from "react";
import { Modal, Platform, Pressable, ScrollView, StyleSheet, Text, View, useWindowDimensions } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { boardThemes, useBoardTheme, useThemedStyles, type AppColors, type BoardTheme } from "@/lib/theme";

export function AppearanceMenu() {
  const { theme, setTheme, mode, setMode, colors } = useBoardTheme();
  const styles = useThemedStyles(createStyles);
  const [open, setOpen] = useState(false);
  const [anchor, setAnchor] = useState({ x: 0, y: 0 });
  const trigger = useRef<View>(null);
  const { width, height } = useWindowDimensions();
  const insets = useSafeAreaInsets();
  const panelWidth = Math.min(280, width - 32);
  const top = Math.max(insets.top + 12, Math.min(anchor.y, height - insets.bottom - 320));
  const left = Math.max(16, Math.min(anchor.x - panelWidth, width - panelWidth - 16));
  const close = useCallback(() => { setOpen(false); }, []);
  useEffect(() => {
    if (!open || Platform.OS !== "web") return;
    const escape = (event: KeyboardEvent) => { if (event.key === "Escape") close(); };
    window.addEventListener("keydown", escape);
    return () => window.removeEventListener("keydown", escape);
  }, [open, close]);
  return <>
    <Pressable ref={trigger} accessibilityRole="button" accessibilityLabel="Appearance"
      aria-expanded={open} accessibilityState={{ expanded: open }} style={styles.trigger}
      onPress={() => trigger.current?.measureInWindow((x, y, w, h) => { setAnchor({ x: x + w, y: y + h + 12 }); setOpen(true); })}>
      <View style={styles.mark} aria-hidden>{[boardThemes[theme].dark, "#E6DFFA", "#F7C5AF", colors.ink].map((backgroundColor, index) => <View key={index} style={[styles.markDot, { backgroundColor }]} />)}</View>
    </Pressable>
    <Modal visible={open} transparent animationType="fade" statusBarTranslucent onRequestClose={close} onDismiss={() => trigger.current?.focus()}>
      <View style={styles.overlay}>
        <Pressable accessibilityRole="button" accessibilityLabel="Dismiss appearance" onPress={close} style={StyleSheet.absoluteFill} />
        <ScrollView accessibilityViewIsModal accessibilityLabel="Appearance settings" style={[styles.panel, { top, left, width: panelWidth, maxHeight: height - top - insets.bottom - 16 }]} contentContainerStyle={styles.content}>
          <View style={styles.heading}><Text accessibilityRole="header" style={styles.title}>Make it yours.</Text><Pressable accessibilityRole="button" accessibilityLabel="Close appearance" onPress={close} style={styles.close}><Text style={styles.closeText}>×</Text></Pressable></View>
          <Text style={styles.label}>BOARD COLOR</Text>
          <View style={styles.swatches}>{(Object.keys(boardThemes) as BoardTheme[]).map((key) => <Pressable key={key}
            accessibilityRole="radio" accessibilityLabel={`${boardThemes[key].name} board`} aria-checked={theme === key} accessibilityState={{ checked: theme === key }}
            onPress={() => setTheme(key)} style={[styles.swatchTarget, theme === key && styles.selected]}>
            <View style={[styles.swatch, { backgroundColor: boardThemes[key].dark, borderColor: theme === key ? colors.ink : "transparent" }]}>{theme === key && <Text style={styles.check}>✓</Text>}</View>
            <Text style={styles.swatchName}>{boardThemes[key].name}</Text>
          </Pressable>)}</View>
          <Text style={styles.label}>APPEARANCE</Text>
          <View style={styles.modes}>{(["light", "dark"] as const).map((value) => <Pressable key={value}
            accessibilityRole="radio" accessibilityLabel={`Use ${value} theme`} aria-checked={mode === value} accessibilityState={{ checked: mode === value }}
            onPress={() => setMode(value)} style={[styles.mode, mode === value && styles.activeMode]}>
            <Text style={[styles.modeText, mode === value && styles.activeText]}>{value === "light" ? "☀  Light" : "☾  Dark"}</Text>
          </Pressable>)}</View>
        </ScrollView>
      </View>
    </Modal>
  </>;
}

const createStyles = (colors: AppColors) => StyleSheet.create({
  trigger: { width: 44, height: 44, borderWidth: 1, borderColor: colors.line, backgroundColor: colors.surface, borderRadius: 15, alignItems: "center", justifyContent: "center" },
  mark: { width: 19, height: 19, flexDirection: "row", flexWrap: "wrap", gap: 3, transform: [{ rotate: "-10deg" }] },
  markDot: { width: 8, height: 8, borderRadius: 4 },
  overlay: { flex: 1, backgroundColor: "#15231B18" },
  panel: { position: "absolute", backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.line, borderRadius: 24, boxShadow: "0 16px 48px #15231B24" },
  content: { padding: 16 },
  heading: { flexDirection: "row", alignItems: "center", justifyContent: "space-between", marginBottom: 12 },
  title: { fontSize: 18, fontWeight: "600", letterSpacing: -0.5, color: colors.ink },
  close: { width: 44, height: 44, alignItems: "center", justifyContent: "center", margin: -6 },
  closeText: { fontSize: 23, color: colors.muted },
  label: { fontSize: 10, fontWeight: "700", letterSpacing: 1.6, color: colors.muted },
  swatches: { flexDirection: "row", gap: 6, marginTop: 10, marginBottom: 20 },
  swatchTarget: { flex: 1, minHeight: 72, alignItems: "center", justifyContent: "center", gap: 8, borderWidth: 1, borderColor: "transparent", borderRadius: 15 },
  selected: { backgroundColor: colors.soft, borderColor: colors.line },
  swatch: { width: 25, height: 25, borderRadius: 20, borderWidth: 2, alignItems: "center", justifyContent: "center" },
  check: { color: "#fff", fontSize: 12, fontWeight: "800" },
  swatchName: { fontSize: 11, color: colors.ink },
  modes: { flexDirection: "row", gap: 4, borderRadius: 15, padding: 4, marginTop: 10, backgroundColor: colors.background, borderWidth: 1, borderColor: colors.line },
  mode: { flex: 1, minHeight: 44, justifyContent: "center", alignItems: "center", borderRadius: 11 },
  activeMode: { backgroundColor: colors.surface, boxShadow: "0 2px 7px #00000012" },
  modeText: { fontSize: 13, color: colors.muted },
  activeText: { color: colors.ink, fontWeight: "600" },
});
