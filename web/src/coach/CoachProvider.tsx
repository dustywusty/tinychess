import { useEffect, useMemo, type ReactNode } from 'react';
import { CopilotKitContext, CopilotKitCoreReact, EMPTY_SET } from '@copilotkit/react-core/v2/context';

// Headless context keeps CopilotKit's DOM chat styles out of the existing board UI.
export function CoachProvider({ gameId, children }: { gameId: string; children: ReactNode }) {
 const core = useMemo(() => new CopilotKitCoreReact({ runtimeUrl: `/api/games/${encodeURIComponent(gameId)}/coach/runtime`, deferInitialConnection: true }), [gameId]);
 useEffect(() => { void core.connect(); }, [core]);
 const value = useMemo(() => ({ copilotkit: core, executingToolCallIds: EMPTY_SET }), [core]);
 return <CopilotKitContext.Provider value={value}>{children}</CopilotKitContext.Provider>;
}
