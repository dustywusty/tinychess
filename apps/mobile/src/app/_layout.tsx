import { Stack } from "expo-router";
import { StatusBar } from "expo-status-bar";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { ThemeProvider, colors } from "@/lib/theme";

export default function RootLayout() {
  return <SafeAreaProvider><ThemeProvider>
    <StatusBar style="dark" />
    <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }}>
      <Stack.Screen name="index" options={{ title: "Your Move" }} />
      <Stack.Screen name="g/[id]" options={{ title: "Game" }} />
    </Stack>
  </ThemeProvider></SafeAreaProvider>;
}
