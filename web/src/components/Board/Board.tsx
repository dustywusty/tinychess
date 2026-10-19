import { useId, useMemo } from "react";
import { Chess, type Square as ChessSquare } from "chess.js";
import { cellSquare } from "../../lib/board";
import type { Color, Square } from "../../types/chess";
import { Piece } from "../Piece";

export interface BoardProps {
  focusSquares?: string[];
  showRoutes?: boolean;
  arrows?: { from: string; to: string; label: string }[];
  fen: string; uci: string[]; perspective: Color; selected: Square | null; disabled?: boolean;
  onSquareClick: (square: Square) => void; onMove?: (from: Square, to: Square) => void; preview?: boolean;
}
const names: Record<string, string> = { p: "pawn", n: "knight", b: "bishop", r: "rook", q: "queen", k: "king" };
export function Board({ fen, uci, perspective, selected, disabled = false, onSquareClick, onMove, preview = false, arrows: inputArrows = [], showRoutes = true, focusSquares = [] }: BoardProps) {
  const markerId = useId().replace(/:/g, '');
  const arrows = inputArrows.filter((arrow, index, all) => all.findIndex(other => other.from === arrow.from && other.to === arrow.to) === index);
  const chess = useMemo(() => new Chess(fen), [fen]);
  const targets = useMemo(() => selected && !disabled ? chess.moves({ square: selected as ChessSquare, verbose: true }).map((move) => move.to) : [], [chess, selected, disabled]);
  const last = uci.at(-1);
  return <div id={preview ? undefined : "board"} data-testid={preview ? undefined : "board"} className={"chess-board" + (preview ? " preview-board" : "")} aria-label={preview ? undefined : "Chessboard"} aria-hidden={preview || undefined}>
    {showRoutes && arrows.length > 0 && <svg className="coach-arrows" viewBox="0 0 8 8" role="img" aria-label={arrows.map(a => `${a.label}: ${a.from} to ${a.to}`).join('. ')}>
      <defs>{arrows.map((arrow, i) => <marker key={i} id={`${markerId}-arrow-${i}`} markerWidth="3" markerHeight="3" refX="2" refY="1.5" orient="auto"><path d="M0,0 L3,1.5 L0,3" fill={arrow.label === 'Alternative' ? '#77629b' : '#447b68'} /></marker>)}</defs>
      {arrows.map((arrow, i) => {
        const point = (square: string) => { const x = square.charCodeAt(0) - 97, y = 8 - Number(square[1]); return perspective === 'white' ? [x + .5, y + .5] : [7.5 - x, 7.5 - y]; };
        const [x1, y1] = point(arrow.from), [targetX, targetY] = point(arrow.to);
        const length = Math.hypot(targetX - x1, targetY - y1);
        const x2 = targetX - (targetX - x1) / length * .27, y2 = targetY - (targetY - y1) / length * .27;
        return <line key={i} x1={x1} y1={y1} x2={x2} y2={y2} stroke={arrow.label === 'Alternative' ? '#77629b' : '#447b68'} strokeWidth=".085" strokeLinecap="round" strokeDasharray={arrow.label === 'Alternative' ? '.2 .1' : undefined} markerEnd={`url(#${markerId}-arrow-${i})`} />;
      })}
    </svg>}
    {Array.from({ length: 64 }, (_, index) => {
      const row = Math.floor(index / 8), column = index % 8;
      const square = cellSquare(row, column, perspective);
      const piece = chess.get(square as ChessSquare);
      const isLast = last?.slice(0, 2) === square || last?.slice(2, 4) === square;
      const inCheck = piece?.type === "k" && piece.color === chess.turn() && chess.isCheck();
      const departure = arrows.find(arrow => arrow.from === square);
      const arrival = arrows.find(arrow => arrow.to === square);
      const highlight = arrival ?? departure;
      const target = targets.includes(square as ChessSquare);
      return <button key={square} type="button" data-square={preview ? undefined : square} data-coach-focus={focusSquares.includes(square) || undefined} data-coach-square={highlight ? arrival ? "destination" : "origin" : undefined} data-coach-color={highlight?.label === "Alternative" ? "alternative" : highlight ? "played" : undefined} disabled={disabled} tabIndex={preview ? -1 : 0}
        aria-label={square + (piece ? ", " + (piece.color === "w" ? "white" : "black") + " " + names[piece.type] : ", empty") + (target ? ", legal move" : "")}
        aria-pressed={selected === square}
        className={["board-square", (row + column) % 2 === 0 ? "light-square" : "dark-square", isLast ? "last-move" : "", selected === square ? "selected-square" : "", inCheck ? "in-check" : ""].join(" ")}
        onClick={() => onSquareClick(square)} draggable={!disabled && !!piece}
        onDragStart={(event) => { event.dataTransfer.setData("text/square", square); event.dataTransfer.effectAllowed = "move"; }}
        onDragOver={(event) => { if (!disabled) event.preventDefault(); }}
        onDrop={(event) => { event.preventDefault(); const from = event.dataTransfer.getData("text/square"); if (/^[a-h][1-8]$/.test(from) && from !== square && !disabled) onMove?.(from as Square, square); }}>
        {!preview && column === 0 && <span className="board-rank">{square[1]}</span>}
        {!preview && row === 7 && <span className="board-file">{square[0]}</span>}
        {highlight && <span className={`coach-square-mark ${arrival ? 'arrival' : 'departure'}`} aria-hidden="true" />}
        {piece && <Piece piece={piece.color === "w" ? piece.type.toUpperCase() : piece.type} />}
        {target && <span className={piece ? "capture-target" : "move-target"} />}
      </button>;
    })}
  </div>;
}
