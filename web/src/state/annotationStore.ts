import { create } from "zustand";
import type { BoardAnnotation } from "../coach/schemas";

const EMPTY_ANNOTATION: BoardAnnotation = { squares: [], arrows: [] };

export interface AnnotationStore {
  /** Half-move number the main board should display alongside annotations. */
  activePly: number | null;
  /** Active annotation set; empty squares + arrows when nothing is highlighted. */
  annotations: BoardAnnotation;
  /** ID of the chat message that pushed the current annotation, for arbitration. */
  sourceMessageId: string | null;
  /** Push a new annotation set (called by InlineBoard on viewport entry). */
  setAnnotation: (input: {
    ply: number;
    annotations: BoardAnnotation;
    sourceMessageId: string;
  }) => void;
  /** Reset to no-op (called by tools / when leaving the review page). */
  clear: () => void;
}

export const useAnnotationStore = create<AnnotationStore>((set) => ({
  activePly: null,
  annotations: EMPTY_ANNOTATION,
  sourceMessageId: null,
  setAnnotation: ({ ply, annotations, sourceMessageId }) =>
    set({ activePly: ply, annotations, sourceMessageId }),
  clear: () =>
    set({
      activePly: null,
      annotations: EMPTY_ANNOTATION,
      sourceMessageId: null,
    }),
}));
