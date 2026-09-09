import type { BotId, StateEvent } from "@yourmove/protocol";
import { gameResult } from "@yourmove/chess";

const KEY = "tinychess:games:v1";

export interface RecentGameEntry {
  id: string;
  createdAt: number;
  lastSeen: number;
  lastSeenLocal: number;
  status: string;
  result: string;
  opponentType?: "human" | "bot";
  botId?: BotId;
  playerColor?: StateEvent["color"];
  role?: StateEvent["role"];
}

export type RecentGames = Record<string, RecentGameEntry>;

export function loadRecentGames(): RecentGames {
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return {};
    const parsed: unknown = JSON.parse(raw);
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return {};
    }
    return parsed as RecentGames;
  } catch {
    return {};
  }
}

export function saveRecentGames(games: RecentGames): void {
  try {
    localStorage.setItem(KEY, JSON.stringify(games));
  } catch {
    /* ignore */
  }
}

type SnapshotInput = Pick<StateEvent, "status" | "lastSeen" | "pgn" | "bot" | "color" | "role">;

export function recordGameSeen(id: string, snap: SnapshotInput): RecentGames {
  const games = loadRecentGames();
  const now = Date.now();
  const existing = games[id];
  const result = gameResult(snap.status, snap.pgn);
  games[id] = {
    id,
    createdAt: existing?.createdAt ?? now,
    lastSeen: snap.lastSeen ?? now,
    lastSeenLocal: now,
    status: snap.status ?? existing?.status ?? "",
    result: result || existing?.result || "",
    opponentType: snap.bot ? "bot" : "human",
    botId: snap.bot?.id,
    playerColor: snap.color !== undefined ? snap.color : existing?.playerColor,
    role: snap.role ?? existing?.role,
  };
  saveRecentGames(games);
  return games;
}

export function forgetGame(id: string): RecentGames {
  const games = loadRecentGames();
  delete games[id];
  saveRecentGames(games);
  return games;
}
