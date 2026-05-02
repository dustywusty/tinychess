const KEY = "tinychess:games:v1";

export interface RecentGameEntry {
  id: string;
  createdAt: number;
  lastSeen: number;
  lastSeenLocal: number;
  status: string;
  result: string;
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

interface SnapshotInput {
  status?: string;
  lastSeen?: number;
  pgn?: string;
}

export function recordGameSeen(id: string, snap: SnapshotInput): RecentGames {
  const games = loadRecentGames();
  const now = Date.now();
  const existing = games[id];
  const result = deriveResult(snap.status, snap.pgn);
  games[id] = {
    id,
    createdAt: existing?.createdAt ?? now,
    lastSeen: snap.lastSeen ?? now,
    lastSeenLocal: now,
    status: snap.status ?? existing?.status ?? "",
    result: result || existing?.result || "",
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

function deriveResult(status: string | undefined, pgn: string | undefined): string {
  if (!status) return "";
  if (pgn?.includes("1-0")) return "1-0";
  if (pgn?.includes("0-1")) return "0-1";
  if (pgn?.includes("1/2-1/2")) return "1/2-1/2";
  return "";
}
