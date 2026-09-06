import { useMemo, useState } from "react";
import { Chess } from "chess.js";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { squareAt, type Color } from "@/lib/chess";
import { boardThemes, useBoardTheme } from "@/lib/theme";
import { Piece } from "./Piece";

const names: Record<string, string> = { p: "pawn", n: "knight", b: "bishop", r: "rook", q: "queen", k: "king" };

export function ChessBoard({ fen, perspective = "white", selected = null, disabled = true, onSquare, destinations = [], lastMove, preview = false }: {
  fen: string; perspective?: Color; selected?: string | null; disabled?: boolean;
  onSquare?: (square: string) => void; destinations?: string[]; lastMove?: string; preview?: boolean;
}) {
  const chess = useMemo(() => new Chess(fen), [fen]);
  const { theme } = useBoardTheme();
  const palette = boardThemes[theme];
  const [width, setWidth] = useState(320);
  return <View accessibilityElementsHidden={preview} importantForAccessibility={preview ? "no-hide-descendants" : "auto"}
    onLayout={(event) => setWidth(event.nativeEvent.layout.width)} style={styles.board}>
    {Array.from({ length: 64 }, (_, index) => {
      const row = Math.floor(index / 8), column = index % 8;
      const square = squareAt(row, column, perspective);
      const piece = chess.get(square);
      const light = (row + column) % 2 === 0;
      const isLast = lastMove?.slice(0, 2) === square || lastMove?.slice(2, 4) === square;
      const target = destinations.includes(square);
      const inCheck = piece?.type === "k" && piece.color === chess.turn() && chess.isCheck();
      return <Pressable key={square} accessibilityRole="button"
        accessibilityLabel={square + (piece ? ", " + (piece.color === "w" ? "white" : "black") + " " + names[piece.type] : ", empty") + (target ? ", legal move" : "") + (inCheck ? ", in check" : "")}
        accessibilityState={{ selected: selected === square, disabled }} disabled={disabled} onPress={() => onSquare?.(square)}
        style={[styles.square, { backgroundColor: inCheck ? "#EEAF97" : selected === square ? palette.selected : light ? palette.light : palette.dark }]}>
        {isLast && <View pointerEvents="none" style={styles.lastMove} />}
        {!preview && column === 0 && <Text style={[styles.coordinate, styles.rank, { color: light ? "#4A6353" : "#FFFFFF" }]}>{square[1]}</Text>}
        {!preview && row === 7 && <Text style={[styles.coordinate, styles.file, { color: light ? "#4A6353" : "#FFFFFF" }]}>{square[0]}</Text>}
        {piece && <Piece piece={piece.color === "w" ? piece.type.toUpperCase() : piece.type} size={width / 8 * 0.83} />}
        {target && <View pointerEvents="none" style={piece ? styles.capture : styles.dot} />}
      </Pressable>;
    })}
  </View>;
}
const styles = StyleSheet.create({
  board: { width: "100%", aspectRatio: 1, flexDirection: "row", flexWrap: "wrap", borderRadius: 12, overflow: "hidden" },
  square: { width: "12.5%", height: "12.5%", alignItems: "center", justifyContent: "center" },
  coordinate: { position: "absolute", fontSize: 8, fontWeight: "700" },
  rank: { top: 3, left: 4 }, file: { bottom: 2, right: 4 },
  dot: { position: "absolute", width: "24%", height: "24%", borderRadius: 30, backgroundColor: "#253D3550" },
  capture: { position: "absolute", width: "94%", height: "94%", borderRadius: 50, borderWidth: 4, borderColor: "#253D3570" },
  lastMove: { position: "absolute", inset: 0, backgroundColor: "#EBF69250", borderWidth: 2, borderColor: "#D7ED9380" },
});
