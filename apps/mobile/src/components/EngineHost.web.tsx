import { useEffect } from "react";
import { workerTransport, type EngineTransport } from "@yourmove/chess/bots";
export function EngineHost({ onReady }: { onReady: (transport: EngineTransport) => void }) {
 useEffect(() => { const transport = workerTransport(); onReady(transport); return () => transport.dispose(); }, [onReady]);
 return null;
}
