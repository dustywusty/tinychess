import type { ReactElement } from "react";

interface UserMessageProps {
  content: string;
}

export function UserMessage({ content }: UserMessageProps) {
  return (
    <div className="self-end max-w-[85%] rounded-2xl rounded-br-md bg-[color:var(--accent)] text-bg px-3 py-2 text-sm">
      {content}
    </div>
  );
}

interface AssistantMessageProps {
  content: string;
  ui: ReactElement[] | null;
}

export function AssistantMessage({ content, ui }: AssistantMessageProps) {
  return (
    <div className="self-start max-w-[95%] space-y-2">
      {content && (
        <div className="rounded-2xl rounded-bl-md bg-panel px-3 py-2 text-sm whitespace-pre-wrap">
          {content}
        </div>
      )}
      {ui && ui.length > 0 && (
        <div className="space-y-2">
          {ui.map((node, i) => (
            <div key={i}>{node}</div>
          ))}
        </div>
      )}
    </div>
  );
}

export function ErrorMessage({ message }: { message: string }) {
  return (
    <div className="self-start max-w-[85%] rounded-md bg-[color:var(--err)]/20 text-[color:var(--err)] px-3 py-2 text-xs">
      {message}
    </div>
  );
}
