import { useState } from "react";
import { Board } from "./components/Board/Board";
import { START_FEN, type Square } from "./types/chess";

export default function App() {
  const [selected, setSelected] = useState<Square | null>(null);
  const [fen, setFen] = useState(START_FEN);
  const [uci, setUci] = useState<string[]>([]);

  const handleSquareClick = (sq: Square) => {
    if (!selected) {
      setSelected(sq);
      return;
    }
    if (selected === sq) {
      setSelected(null);
      return;
    }
    handleMove(selected, sq);
  };

  const handleMove = (from: Square, to: Square) => {
    setUci((prev) => [...prev, `${from}${to}`]);
    setSelected(null);
    // Phase 1 preview only: keep the FEN frozen at start; PR 3 wires real
    // optimistic state + server roundtrip.
    void fen;
    void setFen;
  };

  return (
    <main className="min-h-screen flex flex-col items-center justify-center bg-bg text-text gap-6 p-6">
      <header className="w-full max-w-md text-center space-y-1">
        <h1 className="text-2xl font-semibold">Tiny Chess</h1>
        <p className="text-xs opacity-70">
          PR 2 preview — Board + types + stores. Game flow lands in PR 3.
        </p>
      </header>
      <div className="w-full max-w-md">
        <Board
          fen={fen}
          uci={uci}
          perspective="white"
          selected={selected}
          onSquareClick={handleSquareClick}
          onMove={handleMove}
        />
      </div>
    </main>
  );
}
