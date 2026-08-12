import { Platform } from "react-native";
import type {
  CommandResponse,
  CreateGameResponse,
  MoveResponse,
  StateEvent,
} from "@yourmove/protocol";

const developmentHost = Platform.OS === "android" ? "10.0.2.2" : "localhost";
export const apiURL = process.env.EXPO_PUBLIC_API_URL ?? `http://${developmentHost}:8080`;

async function json<T>(response: Response): Promise<T> {
  if (!response.ok) {
    throw new Error(`Your Move API returned ${response.status}`);
  }
  return (await response.json()) as T;
}

export async function createGame(): Promise<CreateGameResponse> {
  return json(await fetch(`${apiURL}/api/games`, { method: "POST" }));
}

export async function getSnapshot(gameID: string, clientID: string): Promise<StateEvent> {
  const query = new URLSearchParams({ clientId: clientID });
  return json(await fetch(`${apiURL}/api/games/${encodeURIComponent(gameID)}/snapshot?${query}`));
}

export async function makeMove(gameID: string, uci: string, clientID: string): Promise<MoveResponse> {
  return json(
    await fetch(`${apiURL}/api/games/${encodeURIComponent(gameID)}/move`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ uci, clientId: clientID }),
    }),
  );
}

export async function reactToGame(gameID: string, emoji: string, clientID: string): Promise<CommandResponse> {
  return json(
    await fetch(`${apiURL}/api/games/${encodeURIComponent(gameID)}/react`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ emoji, sender: clientID }),
    }),
  );
}
