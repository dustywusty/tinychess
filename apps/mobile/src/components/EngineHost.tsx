import { useEffect, useRef, useState } from "react";
import { View } from "react-native";
import { WebView } from "react-native-webview";
import { Asset } from "expo-asset";
import { File } from "expo-file-system";
import type { EngineTransport } from "@yourmove/chess/bots";

/** Owns a bundled, offline WebView. It never receives game credentials. */
export function EngineHost({ onReady }: { onReady: (transport: EngineTransport) => void }) {
 const ref = useRef<WebView>(null);
 const [html, setHtml] = useState<string>();
 const [error, setError] = useState<Error>();
 const lines = useRef<(line: string) => void>(() => {});
 const failures = useRef<(error: Error) => void>(() => {});
 const pending = useRef<string[]>([]);
 useEffect(() => {
  let active = true;
  const transport: EngineTransport = {
   send(command) {
    const script = `window.arasan ? window.arasan.commands.push(${JSON.stringify(command)}) : (window.pendingCommands ||= []).push(${JSON.stringify(command)}); true;`;
    if (ref.current) ref.current.injectJavaScript(script); else pending.current.push(script);
   },
   subscribe(line, fail) { lines.current = line; failures.current = fail; return () => { lines.current = () => {}; failures.current = () => {}; }; },
   dispose() { if (active) { setHtml(undefined); ref.current?.stopLoading(); } },
  };
  onReady(transport);
  void (async () => {
   const asset = await Asset.fromModule(require("../../assets/engine/arasan.html")).downloadAsync();
   if (!asset.localUri) throw new Error("Computer engine asset is missing.");
   const text = await new File(asset.localUri).text();
   if (active) setHtml(text);
  })().catch(() => { if (active) setError(new Error("Computer engine failed to load.")); });
  return () => { active = false; };
 }, [onReady]);
 useEffect(() => { if (error) failures.current(error); }, [error]);
 return <View pointerEvents="none" accessibilityElementsHidden importantForAccessibility="no-hide-descendants" style={{ position: "absolute", width: 1, height: 1, opacity: 0, overflow: "hidden" }}>
  {html && <WebView ref={ref} source={{ html }} originWhitelist={["about:blank"]} javaScriptEnabled scrollEnabled={false}
   onShouldStartLoadWithRequest={request => request.url === "about:blank"}
   onLoadEnd={() => { for (const script of pending.current.splice(0)) ref.current?.injectJavaScript(script); }}
   onMessage={event => lines.current(event.nativeEvent.data)}
   onError={() => failures.current(new Error("Computer engine failed to load."))}
   onContentProcessDidTerminate={() => failures.current(new Error("Computer engine stopped."))}
   onRenderProcessGone={() => failures.current(new Error("Computer engine stopped."))} />}
 </View>;
}
