import { useMemo, useState, type FormEvent } from "react";
import { exposeComponent, useUiChat } from "@hashbrownai/react";
import { useCoachTools } from "../../coach/buildTools";
import { InlineBoardPropsSchema } from "../../coach/schemas";
import { InlineBoard } from "./InlineBoard";
import { AssistantMessage, ErrorMessage, UserMessage } from "./Message";

interface ChatProps {
  gameId: string;
  pgn: string;
  positions: Record<number, string>;
  plyCount: number;
  /** Inline FEN sequence from the SSE state (or storage). Drives initial moments. */
  initialPrompt: string;
}

export function Chat({ gameId, pgn, positions, plyCount, initialPrompt }: ChatProps) {
  const tools = useCoachTools({ gameId, pgn, positions, plyCount });

  const exposedInlineBoard = useMemo(
    () =>
      exposeComponent(InlineBoard, {
        name: "InlineBoard",
        description:
          "Renders a chess position with an optional annotation overlay and a caption above the board. Use this whenever you want to show the player a specific position.",
        props: InlineBoardPropsSchema,
      }),
    [],
  );

  const system = useMemo(
    () =>
      [
        "You are a chess coach reviewing a completed game with the player who just finished it.",
        "",
        "Walk through 2-4 critical moments where the evaluation swung significantly. For each moment:",
        "- Address the player in second person.",
        "- Be encouraging but honest — name the mistake clearly without being harsh.",
        "- Explain the chess reason (tactical or strategic), not just the eval delta.",
        "- Suggest the better continuation and why.",
        "- Render an InlineBoard component for the position so the player can see it.",
        "",
        "Use the get_position_at tool when you need a FEN for an arbitrary ply, eval_position to confirm Stockfish's recommendation, and set_main_board_annotation to highlight squares on the main board on the left while you explain.",
        "",
        "Keep individual replies tight (2-5 sentences per moment). Use standard algebraic notation for moves.",
      ].join("\n"),
    [],
  );

  const { messages, sendMessage, isReceiving } = useUiChat({
    model: "claude-opus-4-7",
    system,
    components: [exposedInlineBoard],
    tools,
    debugName: "review-coach",
  });

  const [draft, setDraft] = useState("");
  const [primed, setPrimed] = useState(false);

  // Send the initial review prompt once we have the game context loaded.
  if (!primed && initialPrompt && pgn) {
    setPrimed(true);
    void sendMessage({ role: "user", content: initialPrompt });
  }

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const text = draft.trim();
    if (!text || isReceiving) return;
    setDraft("");
    void sendMessage({ role: "user", content: text });
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 overflow-y-auto p-3 flex flex-col gap-2">
        {messages.map((msg, i) => {
          if (msg.role === "user") {
            const text =
              typeof msg.content === "string"
                ? msg.content
                : JSON.stringify(msg.content);
            return <UserMessage key={i} content={text} />;
          }
          if (msg.role === "assistant") {
            const text =
              typeof msg.content === "string" ? msg.content : "";
            return <AssistantMessage key={i} content={text} ui={msg.ui} />;
          }
          if (msg.role === "error") {
            return <ErrorMessage key={i} message={msg.content ?? "error"} />;
          }
          return null;
        })}
        {isReceiving && (
          <div className="text-xs opacity-60 self-start">…thinking</div>
        )}
      </div>
      <form
        onSubmit={handleSubmit}
        className="border-t border-[color:var(--btn-border,_rgba(255,255,255,0.1))] p-3 flex gap-2"
      >
        <input
          type="text"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          placeholder="Ask a follow-up question…"
          disabled={isReceiving}
          className="flex-1 px-3 py-2 rounded-md bg-bg border border-[color:var(--btn-border,_rgba(255,255,255,0.1))] text-sm"
        />
        <button
          type="submit"
          disabled={isReceiving || !draft.trim()}
          className="px-3 py-2 rounded-md bg-[color:var(--accent)] text-bg text-sm disabled:opacity-50"
        >
          Send
        </button>
      </form>
    </div>
  );
}
