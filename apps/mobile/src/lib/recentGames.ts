import AsyncStorage from "@react-native-async-storage/async-storage";
import type { BotId, StateEvent } from "@yourmove/protocol";
import { gameResult } from "@yourmove/chess";

export type RecentGame = { id: string; fen: string; status: string; moves: number; updatedAt: number; opponentType?: "human" | "bot"; botId?: BotId; result?: string; playerColor?: StateEvent["color"]; role?: StateEvent["role"] };
const key = "yourmove.recent-games";
let writing = Promise.resolve();
export async function recentGames(): Promise<RecentGame[]> {
  try {
    const data: unknown = JSON.parse(await AsyncStorage.getItem(key) ?? "[]");
    return Array.isArray(data) ? data.filter((game): game is RecentGame =>
      typeof game?.id === "string" && typeof game.fen === "string" &&
      typeof game.status === "string" && typeof game.moves === "number" && typeof game.updatedAt === "number") : [];
  } catch { return []; }
}
export function rememberGame(id: string, state: StateEvent) {
  writing = writing.then(async () => {
    const games = await recentGames();
    await AsyncStorage.setItem(key, JSON.stringify([
      { id, fen: state.fen, status: state.status, moves: state.uci?.length ?? 0, updatedAt: Date.now(), opponentType: state.bot ? "bot" : "human", botId: state.bot?.id,
        result: gameResult(state.status, state.pgn), playerColor: state.color, role: state.role },
      ...games.filter((game) => game.id !== id),
    ].slice(0, 8)));
  }).catch(() => {});
}
