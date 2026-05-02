import { useCallback, useEffect, useRef, useState } from "react";

export interface ActiveReaction {
  id: string;
  emoji: string;
  origin: "self" | "remote";
}

const REACTION_DURATION_MS = 1600;

export function useReactions() {
  const [active, setActive] = useState<ActiveReaction[]>([]);
  const timers = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

  const show = useCallback((emoji: string, origin: "self" | "remote" = "self") => {
    const id = crypto.randomUUID();
    setActive((rs) => [...rs, { id, emoji, origin }]);
    const timeout = setTimeout(() => {
      setActive((rs) => rs.filter((r) => r.id !== id));
      timers.current.delete(id);
    }, REACTION_DURATION_MS);
    timers.current.set(id, timeout);
  }, []);

  useEffect(() => {
    const t = timers.current;
    return () => {
      t.forEach((timeout) => clearTimeout(timeout));
      t.clear();
    };
  }, []);

  return { active, show };
}
