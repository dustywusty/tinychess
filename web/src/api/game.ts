import type { StateEvent } from "../types/events";

export interface OkResponse {
  ok: boolean;
  error?: string;
}

export interface MoveResponse extends OkResponse {
  state?: Omit<StateEvent, "kind">;
}

export async function createGame(): Promise<{ id: string }> {
  const res = await fetch("/api/games", { method: "POST" });
  if (!res.ok) throw new Error(`createGame failed (${res.status})`);
  return (await res.json()) as { id: string };
}

export async function postMove(
  gameId: string,
  uci: string,
  clientId: string,
): Promise<MoveResponse> {
  const res = await fetch(`/api/games/${gameId}/move`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ uci, clientId }),
  });
  return (await res.json()) as MoveResponse;
}

export async function postReact(
  gameId: string,
  emoji: string,
  sender: string,
): Promise<OkResponse> {
  const res = await fetch(`/api/games/${gameId}/react`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ emoji, sender }),
  });
  return (await res.json()) as OkResponse;
}

export async function postRelease(
  gameId: string,
  clientId: string,
  targetId: string,
): Promise<OkResponse> {
  const res = await fetch(`/api/games/${gameId}/release`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ clientId, targetId }),
  });
  return (await res.json()) as OkResponse;
}
