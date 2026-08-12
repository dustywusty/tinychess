import { Pressable, StyleSheet, Text, View } from "react-native";

type Color = "white" | "black";

const glyphs: Record<string, string> = {
  K: "♔", Q: "♕", R: "♖", B: "♗", N: "♘", P: "♙",
  k: "♚", q: "♛", r: "♜", b: "♝", n: "♞", p: "♟",
};

function boardFromFEN(fen: string): string[][] {
  const rows = fen.split(" ")[0]?.split("/") ?? [];
  return rows.map((row) => {
    const squares: string[] = [];
    for (const char of row) {
      const empty = Number(char);
      if (Number.isInteger(empty) && empty > 0) squares.push(...Array<string>(empty).fill(""));
      else squares.push(char);
    }
    return squares;
  });
}

function squareAt(row: number, column: number, perspective: Color): string {
  const rank = perspective === "white" ? 8 - row : row + 1;
  const file = perspective === "white" ? column : 7 - column;
  return `${String.fromCharCode(97 + file)}${rank}`;
}

export function ChessBoard({
  fen,
  perspective,
  selected,
  disabled,
  onSquare,
}: {
  fen: string;
  perspective: Color;
  selected: string | null;
  disabled: boolean;
  onSquare: (square: string) => void;
}) {
  const board = boardFromFEN(fen);
  return (
    <View style={styles.board}>
      {Array.from({ length: 8 }, (_, displayRow) =>
        Array.from({ length: 8 }, (_, displayColumn) => {
          const row = perspective === "white" ? displayRow : 7 - displayRow;
          const column = perspective === "white" ? displayColumn : 7 - displayColumn;
          const square = squareAt(displayRow, displayColumn, perspective);
          const piece = board[row]?.[column] ?? "";
          const light = (displayRow + displayColumn) % 2 === 0;
          return (
            <Pressable
              key={square}
              accessibilityLabel={square}
              disabled={disabled}
              onPress={() => onSquare(square)}
              style={[
                styles.square,
                light ? styles.light : styles.dark,
                selected === square && styles.selected,
              ]}
            >
              <Text style={styles.piece}>{glyphs[piece] ?? ""}</Text>
            </Pressable>
          );
        }),
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  board: { width: "100%", aspectRatio: 1, flexDirection: "row", flexWrap: "wrap", borderRadius: 8, overflow: "hidden" },
  square: { width: "12.5%", height: "12.5%", alignItems: "center", justifyContent: "center" },
  light: { backgroundColor: "#cbd5e1" },
  dark: { backgroundColor: "#475569" },
  selected: { borderWidth: 3, borderColor: "#22d3ee" },
  piece: { fontSize: 35, lineHeight: 40, color: "#0f172a" },
});
