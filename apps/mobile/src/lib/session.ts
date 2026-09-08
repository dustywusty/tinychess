import AsyncStorage from "@react-native-async-storage/async-storage";
import { randomUUID } from "expo-crypto";
import { persistentIdentity } from "./persistentIdentity";

export const clientID = persistentIdentity(AsyncStorage, "yourmove.client-id", randomUUID);
