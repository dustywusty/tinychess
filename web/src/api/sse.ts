import type { EmojiEvent, StateEvent } from "../types/events";
import { isEmojiEvent, isStateEvent } from "../types/events";

const RECONNECT_MIN_MS = 1000;
const RECONNECT_MAX_MS = 15000;

export interface SSEHandlers {
  onState?: (event: StateEvent) => void;
  onEmoji?: (event: EmojiEvent) => void;
  onOpen?: () => void;
  onError?: (error: Event) => void;
}

export interface SSESubscription {
  close: () => void;
}

export function subscribeSSE(
  gameId: string,
  clientId: string,
  handlers: SSEHandlers,
): SSESubscription {
  let es: EventSource | null = null;
  let timer: ReturnType<typeof setTimeout> | null = null;
  let closed = false;
  let backoff = RECONNECT_MIN_MS;

  const connect = () => {
    if (closed) return;
    const url = `/api/sse/${gameId}?clientId=${encodeURIComponent(clientId)}`;
    es = new EventSource(url);
    es.onopen = () => {
      backoff = RECONNECT_MIN_MS;
      handlers.onOpen?.();
    };
    es.onmessage = (ev) => {
      const raw = (ev.data ?? "").trim();
      if (!raw) return;
      let data: unknown;
      try {
        data = JSON.parse(raw);
      } catch {
        return;
      }
      if (isStateEvent(data)) handlers.onState?.(data);
      else if (isEmojiEvent(data)) handlers.onEmoji?.(data);
      // heartbeat ({}) ignored
    };
    es.onerror = (err) => {
      handlers.onError?.(err);
      es?.close();
      es = null;
      if (closed) return;
      const delay = backoff;
      backoff = Math.min(backoff * 2, RECONNECT_MAX_MS);
      timer = setTimeout(connect, delay);
    };
  };

  connect();

  return {
    close: () => {
      closed = true;
      if (timer) clearTimeout(timer);
      es?.close();
      es = null;
    },
  };
}
