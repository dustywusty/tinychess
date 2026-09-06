export interface EventStream {
  addEventListener(type: "message", listener: (event: { data: string | null }) => void): void;
  addEventListener(type: "error", listener: () => void): void;
  close(): void;
}

// The heartbeat watchdog also recovers silent disconnects after a network change.
export function liveConnection(create: () => EventStream, receive: (value: unknown) => void, disconnected: () => void,
  timing = { retry: 1000, maxRetry: 15000, heartbeat: 35000 }) {
  let closed = false;
  let stream: EventStream | null = null;
  let retry: ReturnType<typeof setTimeout> | undefined;
  let watchdog: ReturnType<typeof setTimeout> | undefined;
  let delay = timing.retry;
  let generation = 0;
  const reconnect = () => {
    if (closed || retry) return;
    generation++;
    clearTimeout(watchdog);
    stream?.close();
    disconnected();
    retry = setTimeout(() => { retry = undefined; connect(); }, delay);
    delay = Math.min(delay * 2, timing.maxRetry);
  };
  const arm = () => { clearTimeout(watchdog); watchdog = setTimeout(reconnect, timing.heartbeat); };
  const connect = () => {
    if (closed) return;
    const current = ++generation;
    try {
      stream = create();
      arm();
      stream.addEventListener("message", (event) => {
        if (closed || current !== generation) return;
        arm();
        delay = timing.retry;
        let value: unknown;
        try { value = JSON.parse(event.data ?? "{}"); } catch { return; }
        receive(value);
      });
      stream.addEventListener("error", () => { if (current === generation) reconnect(); });
    } catch { reconnect(); }
  };
  connect();
  return () => { closed = true; generation++; clearTimeout(retry); clearTimeout(watchdog); stream?.close(); };
}
