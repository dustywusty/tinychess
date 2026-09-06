import { useEffect, useRef, useState } from "react";
import { AccessibilityInfo, Animated, Easing, Platform, StyleSheet, Text, View, useWindowDimensions } from "react-native";
import type { EmojiEvent } from "@yourmove/protocol";

export function ReactionBurst({ reaction }: { reaction?: EmojiEvent }) {
  const { width, height } = useWindowDimensions();
  const scale = useRef(new Animated.Value(1)).current;
  const opacity = useRef(new Animated.Value(1)).current;
  const [visible, setVisible] = useState(false);
  const [reducedMotion, setReducedMotion] = useState(true);
  useEffect(() => {
    let active = true;
    void AccessibilityInfo.isReduceMotionEnabled().then((value) => { if (active) setReducedMotion(value); }).catch(() => {});
    const subscription = AccessibilityInfo.addEventListener("reduceMotionChanged", setReducedMotion);
    return () => { active = false; subscription.remove(); };
  }, []);
  useEffect(() => {
    if (!reaction) { setVisible(false); return; }
    setVisible(true); scale.setValue(1); opacity.setValue(1);
    const config = { duration: 1530, easing: Easing.inOut(Easing.ease), useNativeDriver: Platform.OS !== "web" };
    const animation = Animated.sequence([Animated.delay(270), Animated.parallel([
      Animated.timing(opacity, { ...config, toValue: 0 }),
      Animated.timing(scale, { ...config, toValue: reducedMotion ? 1 : 0.2 }),
    ])]);
    animation.start(({ finished }) => { if (finished) setVisible(false); });
    return () => animation.stop();
  }, [reaction, reducedMotion, scale, opacity]);
  if (!visible || !reaction) return null;
  return <View pointerEvents="none" accessibilityElementsHidden importantForAccessibility="no-hide-descendants" aria-hidden style={styles.overlay}>
    <Animated.View testID="reaction-burst" style={{ opacity, transform: [{ scale }] }}><Text style={{ fontSize: Math.min(260, width * 0.64, height * 0.4), lineHeight: Math.min(260, width * 0.64, height * 0.4) * 1.2 }}>{reaction.emoji}</Text></Animated.View>
  </View>;
}
const styles = StyleSheet.create({
  overlay: { position: "absolute", inset: 0, justifyContent: "center", alignItems: "center", zIndex: 100, pointerEvents: "none" },
});
