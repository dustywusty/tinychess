import { useEffect, useId, useRef } from "react";
import { Chessboard } from "react-chessboard";
import { useAnnotationStore } from "../../state/annotationStore";
import type { InlineBoardProps as InlineBoardSchema } from "../../coach/schemas";

/**
 * Renders a single position from the chat. On mount and on viewport entry
 * pushes its `{ ply, annotations }` into annotationStore so the main board
 * mirrors whichever moment the user is looking at.
 *
 * Props match the InlineBoardProps skillet schema.
 */
export function InlineBoard(props: InlineBoardSchema) {
  const { fen, ply, caption, annotations } = props;
  const setAnnotation = useAnnotationStore((s) => s.setAnnotation);
  const ref = useRef<HTMLDivElement | null>(null);
  const messageId = useId();

  useEffect(() => {
    const node = ref.current;
    if (!node) return;
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting && entry.intersectionRatio >= 0.5) {
            setAnnotation({
              ply,
              annotations: annotations ?? { squares: [], arrows: [] },
              sourceMessageId: messageId,
            });
          }
        }
      },
      { threshold: [0, 0.5, 1] },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [ply, annotations, setAnnotation, messageId]);

  return (
    <div ref={ref} className="my-2 rounded-md bg-bg p-2 border border-[color:var(--btn-border,_rgba(255,255,255,0.1))]">
      {caption && (
        <div className="text-xs opacity-80 mb-1 font-medium">{caption}</div>
      )}
      <div className="aspect-square max-w-[280px]">
        <Chessboard position={fen} arePiecesDraggable={false} />
      </div>
      <div className="text-[10px] opacity-50 mt-1">ply {ply}</div>
    </div>
  );
}
