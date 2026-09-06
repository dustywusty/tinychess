import { useMemo } from "react";
import { Chess, type Square as ChessSquare } from "chess.js";
import { cellSquare } from "../../lib/board";
import type { Color, Square } from "../../types/chess";
import { Piece } from "../Piece";

export interface BoardProps {
  fen: string; uci: string[]; perspective: Color; selected: Square | null; disabled?: boolean;
  onSquareClick: (square: Square) => void; onMove?: (from: Square, to: Square) => void; preview?: boolean;
}
const names: Record<string, string> = { p: "pawn", n: "knight", b: "bishop", r: "rook", q: "queen", k: "king" };
export function Board({ fen, uci, perspective, selected, disabled = false, onSquareClick, onMove, preview = false }: BoardProps) {
  const chess = useMemo(() => new Chess(fen), [fen]);
  const targets = useMemo(() => selected && !disabled ? chess.moves({ square: selected as ChessSquare, verbose: true }).map((move) => move.to) : [], [chess, selected, disabled]);
  const last = uci.at(-1);
  return <div id={preview ? undefined : "board"} data-testid={preview ? undefined : "board"} className={"chess-board" + (preview ? " preview-board" : "")} aria-label={preview ? undefined : "Chessboard"} aria-hidden={preview || undefined}>
    {Array.from({ length: 64 }, (_, index) => {
      const row = Math.floor(index / 8), column = index % 8;
      const square = cellSquare(row, column, perspective);
      const piece = chess.get(square as ChessSquare);
      const isLast = last?.slice(0, 2) === square || last?.slice(2, 4) === square;
      const inCheck = piece?.type === "k" && piece.color === chess.turn() && chess.isCheck();
      const target = targets.includes(square as ChessSquare);
      return <button key={square} type="button" data-square={preview ? undefined : square} disabled={disabled} tabIndex={preview ? -1 : 0}
        aria-label={square + (piece ? ", " + (piece.color === "w" ? "white" : "black") + " " + names[piece.type] : ", empty") + (target ? ", legal move" : "")}
        aria-pressed={selected === square}
        className={["board-square", (row + column) % 2 === 0 ? "light-square" : "dark-square", isLast ? "last-move" : "", selected === square ? "selected-square" : "", inCheck ? "in-check" : ""].join(" ")}
        onClick={() => onSquareClick(square)} draggable={!disabled && !!piece}
        onDragStart={(event) => { event.dataTransfer.setData("text/square", square); event.dataTransfer.effectAllowed = "move"; }}
        onDragOver={(event) => { if (!disabled) event.preventDefault(); }}
        onDrop={(event) => { event.preventDefault(); const from = event.dataTransfer.getData("text/square"); if (/^[a-h][1-8]$/.test(from) && from !== square && !disabled) onMove?.(from as Square, square); }}>
        {!preview && column === 0 && <span className="board-rank">{square[1]}</span>}
        {!preview && row === 7 && <span className="board-file">{square[0]}</span>}
        {piece && <Piece piece={piece.color === "w" ? piece.type.toUpperCase() : piece.type} />}
        {target && <span className={piece ? "capture-target" : "move-target"} />}
      </button>;
    })}
  </div>;
}
