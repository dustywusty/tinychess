import { useCallback, useEffect, useRef, useState } from "react";

export interface ActiveReaction {
  id: string;
  emoji: string;
  origin: "self" | "remote";
}

const REACTION_DURATION_MS = 1800;

export function useReactions() {
  const [active, setActive] = useState<ActiveReaction[]>([]);
  const timer = useRef<ReturnType<typeof setTimeout>>();

  const show = useCallback((emoji: string, origin: "self" | "remote" = "self") => {
    const id = crypto.randomUUID();
    clearTimeout(timer.current);
    setActive([{ id, emoji, origin }]);
    timer.current = setTimeout(() => setActive([]), REACTION_DURATION_MS);
  }, []);

  useEffect(() => () => clearTimeout(timer.current), []);

  return { active, show };
}
