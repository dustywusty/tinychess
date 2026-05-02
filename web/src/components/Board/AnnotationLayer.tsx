import { useAnnotationStore } from "../../state/annotationStore";
import { squareToCell } from "../../lib/board";
import type { Color, Square } from "../../types/chess";

interface Props {
  perspective: Color;
}

const COLOR_MAP = {
  red: "rgba(239, 68, 68, 0.55)",
  green: "rgba(34, 197, 94, 0.55)",
  yellow: "rgba(234, 179, 8, 0.55)",
  blue: "rgba(59, 130, 246, 0.55)",
} as const;

const STROKE_MAP = {
  red: "#ef4444",
  green: "#22c55e",
  yellow: "#eab308",
  blue: "#3b82f6",
} as const;

/**
 * Reads the active annotation from annotationStore and renders highlights +
 * arrows over the main board. Sized as an absolute overlay with the same
 * 8x8 grid as Board so coordinates line up.
 */
export function AnnotationLayer({ perspective }: Props) {
  const annotations = useAnnotationStore((s) => s.annotations);
  const { squares, arrows } = annotations;

  if (squares.length === 0 && arrows.length === 0) return null;

  return (
    <div
      aria-hidden="true"
      className="absolute inset-0 pointer-events-none"
    >
      <div className="grid w-full h-full grid-cols-8 grid-rows-8">
        {squares.map((s, i) => {
          const { row, col } = squareToCell(s.square as Square, perspective);
          return (
            <div
              key={`sq-${i}-${s.square}`}
              style={{
                gridRow: row + 1,
                gridColumn: col + 1,
                background: COLOR_MAP[s.color],
              }}
            />
          );
        })}
      </div>
      <svg
        viewBox="0 0 8 8"
        preserveAspectRatio="none"
        className="absolute inset-0 w-full h-full"
      >
        {arrows.map((a, i) => {
          const from = squareToCell(a.from as Square, perspective);
          const to = squareToCell(a.to as Square, perspective);
          // Center of each cell, with origin at top-left.
          const x1 = from.col + 0.5;
          const y1 = from.row + 0.5;
          const x2 = to.col + 0.5;
          const y2 = to.row + 0.5;
          return (
            <line
              key={`ar-${i}-${a.from}-${a.to}`}
              x1={x1}
              y1={y1}
              x2={x2}
              y2={y2}
              stroke={STROKE_MAP[a.color]}
              strokeWidth={0.18}
              strokeLinecap="round"
              markerEnd={`url(#ann-arrow-${a.color})`}
            />
          );
        })}
        <defs>
          {(["red", "green", "yellow", "blue"] as const).map((c) => (
            <marker
              key={c}
              id={`ann-arrow-${c}`}
              viewBox="0 0 10 10"
              refX="6"
              refY="5"
              markerWidth="4"
              markerHeight="4"
              orient="auto-start-reverse"
            >
              <path d="M 0 0 L 10 5 L 0 10 z" fill={STROKE_MAP[c]} />
            </marker>
          ))}
        </defs>
      </svg>
    </div>
  );
}
