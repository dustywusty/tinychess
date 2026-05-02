import { s } from "@hashbrownai/core";

const ArrowColor = s.enumeration("Arrow / highlight color", [
  "red",
  "green",
  "yellow",
  "blue",
]);

const HighlightedSquare = s.object("A highlighted square", {
  square: s.string("Algebraic, e.g. 'e4'"),
  color: ArrowColor,
});

const Arrow = s.object("An arrow from one square to another", {
  from: s.string("From square in algebraic"),
  to: s.string("To square in algebraic"),
  color: ArrowColor,
});

export const BoardAnnotation = s.object("Annotation overlay for a board", {
  squares: s.array("Highlighted squares (may be empty)", HighlightedSquare),
  arrows: s.array("Arrows (may be empty)", Arrow),
});

// Per-field record used by exposeComponent (which expects { name: schema, … }
// not an ObjectType). Hashbrown's lib uses anyOf+nullish to encode optional.
export const InlineBoardPropsSchema = {
  fen: s.string("Position to render in FEN notation"),
  ply: s.number(
    "Half-move number this position corresponds to in the game (0-indexed)",
  ),
  caption: s.string("Short caption above the board"),
  annotations: s.anyOf([BoardAnnotation, s.nullish()]),
  moveSequence: s.anyOf([
    s.array(
      "Optional sequence of UCI moves to animate from this position",
      s.string("UCI move, e.g. 'e2e4'"),
    ),
    s.nullish(),
  ]),
} as const;

// Composed ObjectType for use as a tool input (e.g.
// set_main_board_annotation accepts annotations: BoardAnnotation).
export const InlineBoardProps = s.object(
  "Props for an inline coach board",
  InlineBoardPropsSchema,
);

export const ReviewMoment = s.object("A single coaching moment in a review", {
  title: s.string(
    "One-line framing of the moment, e.g. 'Move 23: a tactical chance'",
  ),
  prose: s.string(
    "2-4 sentence explanation in second person, addressed to the player",
  ),
  board: InlineBoardProps,
  severity: s.enumeration("How significant the moment was", [
    "blunder",
    "mistake",
    "inaccuracy",
    "good_move",
    "lesson",
  ]),
});

// Inferred TypeScript types — used by InlineBoard / Message components.
export type BoardAnnotation = s.Infer<typeof BoardAnnotation>;
export type InlineBoardProps = s.Infer<typeof InlineBoardProps>;
export type ReviewMoment = s.Infer<typeof ReviewMoment>;
