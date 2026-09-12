import type {
  Board8x8,
  Color,
  PieceCell,
  PieceLetter,
  Square,
} from "../types/chess";

export const PIECE_GLYPH: Record<PieceLetter, string> = {
  P: "♙",
  N: "♘",
  B: "♗",
  R: "♖",
  Q: "♕",
  K: "♔",
  p: "♟",
  n: "♞",
  b: "♝",
  r: "♜",
  q: "♛",
  k: "♚",
};

const SOLID_GLYPH_LOOKUP: Record<string, string> = {
  P: "♟",
  N: "♞",
  B: "♝",
  R: "♜",
  Q: "♛",
  K: "♚",
  p: "♟",
  n: "♞",
  b: "♝",
  r: "♜",
  q: "♛",
  k: "♚",
};

export function solidGlyph(piece: PieceLetter): string {
  return SOLID_GLYPH_LOOKUP[piece] ?? "";
}

export function isPieceLetter(c: string): c is PieceLetter {
  return c in PIECE_GLYPH;
}

export function parseFENBoard(fen: string): Board8x8 {
  const placement = fen.split(" ")[0] ?? "";
  const ranks = placement.split("/");
  const board: Board8x8 = [];
  for (const rank of ranks) {
    const row: PieceCell[] = [];
    for (const ch of rank) {
      if (/[1-8]/.test(ch)) {
        for (let i = 0; i < Number(ch); i++) row.push("");
      } else if (isPieceLetter(ch)) {
        row.push(ch);
      }
    }
    while (row.length < 8) row.push("");
    board.push(row);
  }
  while (board.length < 8) board.push(["", "", "", "", "", "", "", ""]);
  return board;
}

/**
 * Maps a (displayRow, displayCol) cell coordinate to algebraic notation,
 * accounting for board orientation. With white perspective, rank 8 is at the
 * top (displayRow 0); with black perspective, rank 1 is at the top.
 */
export function cellSquare(
  displayRow: number,
  displayCol: number,
  perspective: Color,
): Square {
  if (perspective === "white") {
    const file = String.fromCharCode("a".charCodeAt(0) + displayCol);
    const rank = String(8 - displayRow);
    return `${file}${rank}` as Square;
  }
  const file = String.fromCharCode("a".charCodeAt(0) + (7 - displayCol));
  const rank = String(displayRow + 1);
  return `${file}${rank}` as Square;
}

/**
 * Inverse of cellSquare: maps a square to its display position from the
 * given perspective.
 */
export function squareToCell(
  square: Square,
  perspective: Color,
): { row: number; col: number } {
  const file = square.charCodeAt(0) - "a".charCodeAt(0);
  const rank = Number(square[1]);
  if (perspective === "white") {
    return { row: 8 - rank, col: file };
  }
  return { row: rank - 1, col: 7 - file };
}

export function deriveLastMoveSquares(
  uci: string[],
): { from: Square; to: Square } | null {
  if (!uci || uci.length === 0) return null;
  const last = uci[uci.length - 1];
  if (!last || last.length < 4) return null;
  return {
    from: last.slice(0, 2) as Square,
    to: last.slice(2, 4) as Square,
  };
}

const STARTING_COUNTS: Record<PieceLetter, number> = {
  P: 8,
  N: 2,
  B: 2,
  R: 2,
  Q: 1,
  K: 1,
  p: 8,
  n: 2,
  b: 2,
  r: 2,
  q: 1,
  k: 1,
};

export interface CapturedSet {
  byWhite: PieceLetter[];
  byBlack: PieceLetter[];
}

/**
 * Returns pieces white has captured (lowercase letters) and pieces black has
 * captured (uppercase letters), derived from the FEN placement.
 */
export function capturedFromFEN(fen: string): CapturedSet {
  const placement = fen.split(" ")[0] ?? "";
  const counts: Record<PieceLetter, number> = {
    P: 0,
    N: 0,
    B: 0,
    R: 0,
    Q: 0,
    K: 0,
    p: 0,
    n: 0,
    b: 0,
    r: 0,
    q: 0,
    k: 0,
  };
  for (const ch of placement) {
    if (isPieceLetter(ch)) counts[ch]++;
  }
  const byWhite: PieceLetter[] = [];
  const byBlack: PieceLetter[] = [];
  (Object.keys(STARTING_COUNTS) as PieceLetter[]).forEach((p) => {
    const lost = STARTING_COUNTS[p] - counts[p];
    for (let i = 0; i < lost; i++) {
      if (p === p.toUpperCase()) byBlack.push(p);
      else byWhite.push(p);
    }
  });
  return { byWhite, byBlack };
}

/**
 * Wraps PGN tokens into ~lineLen-character lines, preserving move-number
 * groupings. Used for the moves panel.
 */
export function formatPGNLines(pgn: string, lineLen = 60): string[] {
  if (!pgn) return [];
  const tokens = pgn.split(/\s+/).filter(Boolean);
  const lines: string[] = [];
  let cur = "";
  for (const t of tokens) {
    if (cur.length + t.length + 1 > lineLen) {
      if (cur) lines.push(cur);
      cur = t;
    } else {
      cur = cur ? `${cur} ${t}` : t;
    }
  }
  if (cur) lines.push(cur);
  return lines;
}

export function turnFromFEN(fen: string): Color {
  const parts = fen.split(" ");
  return parts[1] === "b" ? "black" : "white";
}

/**
 * Normalizes whatever the server emits for color/turn fields ("w"/"b" today,
 * "white"/"black" historically) into a strict Color. Returns null on
 * unrecognized input so callers can preserve "no color assigned" semantics.
 */
export function normalizeColor(value: unknown): Color | null {
  if (value === null || value === undefined) return null;
  const v = String(value).toLowerCase();
  if (v === "w" || v === "white") return "white";
  if (v === "b" || v === "black") return "black";
  return null;
}
