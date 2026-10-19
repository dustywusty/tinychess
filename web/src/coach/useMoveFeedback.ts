import { randomUUID } from "../lib/uuid.ts";
import { useEffect, useState } from 'react';
import type { Lesson } from './controller';

// Analyze the player's most recent move without changing the live board.
export function useMoveFeedback(gameId: string, clientId: string, moves: string[], color: string | null, enabled: boolean) {
 const [state, setState] = useState<{ lesson: Lesson | null; pending: boolean; error: string }>({ lesson: null, pending: false, error: '' });
 const side = color === 'black' ? 1 : 0;
 const ply = color ? moves.length - 1 - ((moves.length - 1 - side + 2) % 2) : -1;
 const position = ply >= 0 ? moves.slice(0, ply + 1).join(' ') : '';
 useEffect(() => {
  if (!enabled || !clientId || ply < 0) { setState({ lesson: null, pending: false, error: '' }); return; }
  const abort = new AbortController();
  setState({ lesson: null, pending: true, error: '' });
  const timer = window.setTimeout(async () => {
   try {
    const response = await fetch(`/api/games/${gameId}/coach`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, signal: abort.signal, body: JSON.stringify({ clientId, action: 'review', ply, requestId: randomUUID() }) });
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || 'Couldn’t review this move.');
    if (!abort.signal.aborted) setState({ lesson: data, pending: false, error: '' });
   } catch (error) { if (!abort.signal.aborted) setState({ lesson: null, pending: false, error: error instanceof Error ? error.message : 'Move review unavailable.' }); }
  }, 180);
  return () => { clearTimeout(timer); abort.abort(); };
 }, [gameId, clientId, position, ply, enabled]);
 return state;
}
