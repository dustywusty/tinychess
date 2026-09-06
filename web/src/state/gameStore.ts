import { create } from "zustand";
import type { Color } from "../types/chess";
import { START_FEN } from "../types/chess";
import type { StateEvent } from "@yourmove/protocol";
import { normalizeColor, turnFromFEN } from "../lib/board";

export interface GameStore {
  // Server-mirrored fields
  fen: string;
  turn: Color;
  status: string;
  pgn: string;
  uci: string[];
  watchers: number;
  lastSeen: number;
  // Per-client identity
  clientId: string;
  playerColor: Color | null;
  isSpectator: boolean;
  // Optimistic state (last applied locally; rolled back on server reject)
  optimisticUci: string | null;
  // Actions
  applyServerState: (event: StateEvent) => void;
  setClientId: (id: string) => void;
  setOptimistic: (uci: string | null) => void;
  reset: () => void;
}

const initial: Omit<
  GameStore,
  "applyServerState" | "setClientId" | "setOptimistic" | "reset"
> = {
  fen: START_FEN,
  turn: "white",
  status: "",
  pgn: "",
  uci: [],
  watchers: 0,
  lastSeen: 0,
  clientId: "",
  playerColor: null,
  isSpectator: true,
  optimisticUci: null,
};

export const useGameStore = create<GameStore>((set) => ({
  ...initial,
  applyServerState: (event) =>
    set((prev) => ({
      fen: event.fen,
      turn: normalizeColor(event.turn) ?? turnFromFEN(event.fen),
      status: event.status,
      pgn: event.pgn,
      uci: event.uci ?? [],
      watchers: event.watchers,
      lastSeen: event.lastSeen,
      clientId: event.clientId ?? prev.clientId,
      playerColor:
        event.color !== undefined
          ? normalizeColor(event.color)
          : prev.playerColor,
      isSpectator:
        event.role !== undefined
          ? event.role === "spectator"
          : prev.isSpectator,
      optimisticUci: null,
    })),
  setClientId: (id) => set({ clientId: id }),
  setOptimistic: (uci) => set({ optimisticUci: uci }),
  reset: () => set(initial),
}));
