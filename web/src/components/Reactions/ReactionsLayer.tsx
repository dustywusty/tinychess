import type { ActiveReaction } from "../../hooks/useReactions";

interface Props {
  reactions: ActiveReaction[];
}

export function ReactionsLayer({ reactions }: Props) {
  if (reactions.length === 0) return null;
  // The center "big emoji" shows for any incoming reaction.
  const latest = reactions[reactions.length - 1];
  return (
    <>
      <div key={latest.id} className="big-emoji" aria-hidden="true">
        {latest.emoji}
      </div>
    </>
  );
}
