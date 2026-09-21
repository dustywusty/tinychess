export interface Arrow { from: string; to: string; label: string }
export interface MoveAssessment { grade: string; tone: string; played: string; best: string; playedScore: { kind: 'cp' | 'mate'; value: number }; bestScore: { kind: 'cp' | 'mate'; value: number }; summary: string; difference?: number }
export interface Lesson {
 currentPosition?: boolean;
 recommendation?: { move: string; summary: string; score: { kind: 'cp' | 'mate'; value: number } };
 assessment?: MoveAssessment;
 views?: Record<string, { caption: string; fen: string; arrows: Arrow[]; score: { kind: 'cp' | 'mate'; value: number } }[]>;
 id: string; revision: number; positionId: string; ply: number; fen: string;
 mode: string; branch: string; step: number; steps: number; caption: string;
 arrows: Arrow[]; reveal: number; quality: string;
 choices?: { id: string; label: string; move: string }[];
 score?: { kind: 'cp' | 'mate'; value: number };
}
export interface CoachState { revision: number; pending: boolean; lesson: Lesson | null; error: string }
export const initialCoach: CoachState = { revision: 0, pending: false, lesson: null, error: '' };
export type CoachEvent = { type: 'request' } | { type: 'return' } | { type: 'result'; revision: number; lesson: Lesson } | { type: 'error'; revision: number; error: string };
// The request revision invalidates delayed responses, including after returning live.
export function coachReducer(state: CoachState, event: CoachEvent): CoachState {
 if (event.type === 'return') return { ...initialCoach, revision: state.revision + 1 };
 if (event.type === 'request') return { ...state, revision: state.revision + 1, pending: true, error: '' };
 if (event.revision !== state.revision) return state;
 if (event.type === 'error') return { ...state, pending: false, error: event.error };
 if (event.lesson.mode === 'live') return { ...initialCoach, revision: state.revision + 1 };
 return { ...state, pending: false, lesson: event.lesson };
}

export function navigate(lesson: Lesson, action: string): Lesson | null {
 if (!lesson.views || lesson.reveal < 3) return null;
 const branch = ['played', 'alternative', 'compare'].includes(action) ? action : lesson.branch;
 const frames = lesson.views[branch]; if (!frames) return null;
 const step = action === 'before' ? 0 : action === 'next' ? Math.min(lesson.step + 1, frames.length - 1) : action === 'back' ? Math.max(lesson.step - 1, 0) : branch === 'compare' ? 0 : Math.min(1, frames.length - 1);
 const frame = frames[step];
 return { ...lesson, ...frame, caption: lesson.recommendation && step < 2 ? lesson.recommendation.summary : frame.caption, branch, step, steps: frames.length - 1, mode: 'reviewing' };
}

export function formatScore(score: { kind: 'cp' | 'mate'; value: number }): string {
 if (score.kind === 'cp') return `${(score.value / 100).toFixed(2)} pawns`;
 if (score.value === 0) return 'Checkmate';
 return score.value > 0 ? `Mate in ${score.value}` : `Faces mate in ${Math.abs(score.value)}`;
}
