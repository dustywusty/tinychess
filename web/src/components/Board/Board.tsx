import { useMemo } from "react";
import {
  PIECE_GLYPH,
  cellSquare,
  deriveLastMoveSquares,
  parseFENBoard,
  solidGlyph,
} from "../../lib/board";
import type { Color, PieceLetter, Square } from "../../types/chess";

export interface BoardProps {
  fen: string;
  uci: string[];
  perspective: Color;
  selected: Square | null;
  disabled?: boolean;
  onSquareClick: (square: Square) => void;
  onMove?: (from: Square, to: Square) => void;
}

export function Board({
  fen,
  uci,
  perspective,
  selected,
  disabled = false,
  onSquareClick,
  onMove,
}: BoardProps) {
  const board = useMemo(() => parseFENBoard(fen), [fen]);
  const lastMove = useMemo(() => deriveLastMoveSquares(uci), [uci]);

  return (
    <div
      id="board"
      data-testid="board"
      className="grid w-full aspect-square grid-cols-8 grid-rows-8 select-none rounded-md overflow-hidden"
    >
      {Array.from({ length: 8 }).map((_, displayRow) =>
        Array.from({ length: 8 }).map((__, displayCol) => {
          const sq = cellSquare(displayRow, displayCol, perspective);
          const fenRow = perspective === "white" ? displayRow : 7 - displayRow;
          const fenCol = perspective === "white" ? displayCol : 7 - displayCol;
          const piece = board[fenRow]?.[fenCol] ?? "";
          return (
            <Cell
              key={sq}
              square={sq}
              displayRow={displayRow}
              displayCol={displayCol}
              piece={piece || null}
              perspective={perspective}
              selected={selected === sq}
              isLastMoveFrom={lastMove?.from === sq}
              isLastMoveTo={lastMove?.to === sq}
              disabled={disabled}
              onClick={onSquareClick}
              onMove={onMove}
            />
          );
        }),
      )}
    </div>
  );
}

interface CellProps {
  square: Square;
  displayRow: number;
  displayCol: number;
  piece: PieceLetter | null;
  perspective: Color;
  selected: boolean;
  isLastMoveFrom: boolean;
  isLastMoveTo: boolean;
  disabled: boolean;
  onClick: (square: Square) => void;
  onMove?: (from: Square, to: Square) => void;
}

function Cell({
  square,
  displayRow,
  displayCol,
  piece,
  perspective,
  selected,
  isLastMoveFrom,
  isLastMoveTo,
  disabled,
  onClick,
  onMove,
}: CellProps) {
  const isLight = (displayRow + displayCol) % 2 === 1;
  const showFile = displayRow === 7;
  const showRank = displayCol === 0;
  const fileLabel = showFile
    ? perspective === "white"
      ? String.fromCharCode("a".charCodeAt(0) + displayCol)
      : String.fromCharCode("a".charCodeAt(0) + (7 - displayCol))
    : "";
  const rankLabel = showRank
    ? perspective === "white"
      ? String(8 - displayRow)
      : String(displayRow + 1)
    : "";
  const glyph = piece ? solidGlyph(piece) : "";
  const isWhitePiece = piece && piece === piece.toUpperCase();

  const baseClass = isLight ? "bg-sq1" : "bg-sq2";
  const classes = [
    "relative flex items-center justify-center",
    baseClass,
    selected ? "outline outline-3 -outline-offset-[3px] outline-accent" : "",
    isLastMoveFrom || isLastMoveTo
      ? "shadow-[inset_0_0_0_3px_var(--accent)]"
      : "",
    disabled ? "cursor-default" : "cursor-pointer",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div
      data-square={square}
      className={classes}
      onClick={disabled ? undefined : () => onClick(square)}
      draggable={!disabled && !!piece}
      onDragStart={(e) => {
        if (disabled || !piece) return;
        e.dataTransfer.setData("text/square", square);
        e.dataTransfer.effectAllowed = "move";
      }}
      onDragOver={(e) => {
        if (disabled) return;
        e.preventDefault();
        e.dataTransfer.dropEffect = "move";
      }}
      onDrop={(e) => {
        if (disabled) return;
        e.preventDefault();
        const from = e.dataTransfer.getData("text/square") as Square;
        if (from && from !== square) {
          onMove?.(from, square);
        }
      }}
    >
      {fileLabel && (
        <span className="absolute bottom-0 right-1 text-[10px] opacity-50 leading-none">
          {fileLabel}
        </span>
      )}
      {rankLabel && (
        <span className="absolute top-0 left-1 text-[10px] opacity-50 leading-none">
          {rankLabel}
        </span>
      )}
      {glyph && (
        <span
          className={`text-[clamp(22px,6vw,54px)] leading-none ${
            isWhitePiece ? "text-white" : "text-black"
          }`}
          style={
            isWhitePiece
              ? {
                  WebkitTextStroke: "1.5px var(--sq2)",
                  paintOrder: "stroke fill",
                }
              : undefined
          }
          aria-label={piece ? `${isWhitePiece ? "white" : "black"} ${piece}` : undefined}
        >
          {glyph}
        </span>
      )}
      {/* hidden full-character glyph for accessibility/screen readers */}
      {piece && (
        <span className="sr-only">{PIECE_GLYPH[piece]}</span>
      )}
    </div>
  );
}
