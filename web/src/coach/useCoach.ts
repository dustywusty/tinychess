import { useAgent } from '@copilotkit/react-core/v2/headless';
import { useCallback, useEffect, useRef, useState } from 'react';
import { coachReducer, initialCoach, navigate, type Lesson, type CoachEvent } from './controller';

export function useCoach(gameId: string, clientId: string) {
 const { agent } = useAgent({ agentId: 'coach' });
 const [state, setState] = useState(initialCoach);
 const current = useRef(initialCoach);
 const pending = useRef<AbortController | null>(null);
 const failedRequest = useRef<Record<string, unknown> | null>(null);
 const dispatch = useCallback((event: CoachEvent) => { current.current = coachReducer(current.current, event); setState(current.current); }, []);
 const returnToGame = useCallback(() => { pending.current?.abort(); agent.abortRun(); failedRequest.current = null; dispatch({ type: 'return' }); }, [dispatch, agent]);
 useEffect(() => { returnToGame(); return () => pending.current?.abort(); }, [gameId, clientId, returnToGame]);
 const command = useCallback(async (action: string, args: Record<string, unknown> = {}) => {
  if (action === 'return') { returnToGame(); return; }
  const visible = current.current.lesson;
  if (visible && !args.noSpoilers && ['played', 'alternative', 'compare', 'before', 'next', 'back'].includes(action)) {
   const next = navigate(visible, action);
   if (next) { pending.current?.abort(); failedRequest.current = null; dispatch({ type: 'request' }); dispatch({ type: 'result', revision: current.current.revision, lesson: next }); return; }
  }
  if (action === "retry" && !failedRequest.current) return;
  pending.current?.abort(); const abort = new AbortController(); pending.current = abort;
  abort.signal.addEventListener("abort", () => agent.abortRun(), { once: true });
  const lesson = current.current.lesson;
  const request = action === 'retry' ? failedRequest.current! : { clientId, action, requestId: crypto.randomUUID(), lessonId: lesson?.id ?? '', revision: lesson?.revision ?? 0, branch: lesson?.branch ?? '', step: lesson?.step ?? 0, ...args };
  dispatch({ type: 'request' }); const revision = current.current.revision;
  const fail = (message: string) => {
   if (abort.signal.aborted) return;
   const expired = /lesson expired|stale review|reviewed position changed/.test(message);
   failedRequest.current = expired ? {
    clientId, action: lesson?.currentPosition || request.action === 'best_now' ? 'best_now' : 'review',
    ply: lesson?.ply ?? request.ply ?? 0, requestId: crypto.randomUUID(),
    noSpoilers: !!request.noSpoilers || (!!lesson && lesson.reveal < 3),
   } : request;
   dispatch({ type: 'error', revision, error: expired ? 'This analysis expired. Refresh it to continue.' : message });
  };
  try {
   agent.setState({}); agent.setMessages([]);
   let streamedCaption = "";
   await agent.runAgent({ forwardedProps: { coach: request } }, {
    onStateSnapshotEvent: ({ event }) => {
     if (!abort.signal.aborted) { failedRequest.current = null; dispatch({ type: 'result', revision, lesson: event.snapshot.lesson as Lesson }); }
    },
    onTextMessageContentEvent: ({ event }) => {
     streamedCaption += event.delta;
     if (!abort.signal.aborted && current.current.revision === revision && current.current.lesson) dispatch({ type: 'result', revision, lesson: { ...current.current.lesson, caption: streamedCaption } });
    },
    onRunErrorEvent: ({ event }) => fail(event.message),
   });
  } catch (error) {
   fail(error instanceof Error ? error.message : 'Coach unavailable.');
  }
 }, [gameId, clientId, dispatch, returnToGame, agent]);
 const openLesson = useCallback((lesson: Lesson, branch = "played") => {
  pending.current?.abort(); agent.abortRun(); failedRequest.current = null;
  dispatch({ type: "request" });
  dispatch({ type: "result", revision: current.current.revision, lesson: navigate(lesson, branch) ?? lesson });
 }, [agent, dispatch]);
 return { ...state, command, returnToGame, openLesson };
}
