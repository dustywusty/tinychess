import AsyncStorage from "@react-native-async-storage/async-storage";

let pendingID: Promise<string> | null = null;
export function clientID(): Promise<string> {
  pendingID ??= (async () => {
    const saved = await AsyncStorage.getItem("yourmove.client-id");
    if (saved) return saved;
    const id = `mobile-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 14)}`;
    await AsyncStorage.setItem("yourmove.client-id", id);
    return id;
  })().catch((error) => { pendingID = null; throw error; });
  return pendingID;
}
