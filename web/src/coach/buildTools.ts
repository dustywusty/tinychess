import { useTool } from "@hashbrownai/react";
import { s } from "@hashbrownai/core";
import { evalPosition } from "../lib/engine";
import { useAnnotationStore } from "../state/annotationStore";
import { BoardAnnotation } from "./schemas";

const MAIN_BOARD_SOURCE = "tool:set_main_board_annotation";

interface CoachToolsArgs {
  gameId: string;
  pgn: string;
  /** FEN at each ply, indexed by half-move number. */
  positions: Record<number, string>;
  /** Total ply count for bounds checks. */
  plyCount: number;
}

/**
 * Builds the browser-side tools the coach can call. Returns the tool
 * descriptors that should be passed to `useUiChat({ tools: ... })`.
 *
 * Tools fall into two groups:
 * - Read state: get_game_metadata, get_position_at
 * - Drive UI: set_main_board_annotation, clear_main_board_annotation
 * - Engine: eval_position (Stockfish-WASM, lazy-loaded)
 */
export function useCoachTools({ gameId, pgn, positions, plyCount }: CoachToolsArgs) {
  const setAnnotation = useAnnotationStore((s) => s.setAnnotation);
  const clearAnnotation = useAnnotationStore((s) => s.clear);

  const getMetadata = useTool({
    name: "get_game_metadata",
    description:
      "Return high-level metadata about the game under review (gameId, PGN, total ply count).",
    schema: s.object("No input", {}),
    handler: async () => ({ gameId, pgn, plyCount }),
    deps: [gameId, pgn, plyCount],
  });

  const getPositionAt = useTool({
    name: "get_position_at",
    description:
      "Return the FEN at a given half-move number (ply). Ply 0 is the starting position; subsequent plies follow the move list.",
    schema: s.object("Position lookup input", {
      ply: s.integer("Half-move number, 0-indexed"),
    }),
    handler: async (input) => {
      const fen = positions[input.ply];
      if (!fen) {
        throw new Error(`No position cached for ply ${input.ply}`);
      }
      return { ply: input.ply, fen };
    },
    deps: [positions],
  });

  const evalPositionTool = useTool({
    name: "eval_position",
    description:
      "Run Stockfish-WASM against a FEN and return its eval (centipawns from white's POV) and best move (UCI). depth defaults to 14.",
    schema: s.object("Eval input", {
      fen: s.string("Position to evaluate, FEN"),
      depth: s.anyOf([s.integer("Search depth, default 14"), s.nullish()]),
    }),
    handler: async (input) => {
      const depth = input.depth ?? 14;
      const result = await evalPosition(input.fen, depth);
      return {
        fen: result.fen,
        depth: result.depth,
        evalCp: result.evalCp,
        mateIn: result.mateIn,
        bestMove: result.bestMove,
      };
    },
    deps: [],
  });

  const setMainBoardAnnotation = useTool({
    name: "set_main_board_annotation",
    description:
      "Highlight squares and draw arrows on the main board on the left of the review page. Replaces any existing annotation.",
    schema: s.object("Annotation input", {
      ply: s.integer("Ply to display while annotation is active"),
      annotations: BoardAnnotation,
    }),
    handler: async (input) => {
      setAnnotation({
        ply: input.ply,
        annotations: input.annotations,
        sourceMessageId: MAIN_BOARD_SOURCE,
      });
      return { ok: true };
    },
    deps: [setAnnotation],
  });

  const clearMainBoardAnnotation = useTool({
    name: "clear_main_board_annotation",
    description: "Clear any highlights and arrows from the main board.",
    schema: s.object("No input", {}),
    handler: async () => {
      clearAnnotation();
      return { ok: true };
    },
    deps: [clearAnnotation],
  });

  return [
    getMetadata,
    getPositionAt,
    evalPositionTool,
    setMainBoardAnnotation,
    clearMainBoardAnnotation,
  ];
}
