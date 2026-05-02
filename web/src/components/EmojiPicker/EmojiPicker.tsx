import { useEffect, useRef, useState } from "react";
import "emoji-picker-element";
import { loadRecentEmojis, rememberEmoji } from "../../lib/recentEmojis";

const COOLDOWN_MS = 5000;

interface Props {
  disabled?: boolean;
  onSend: (emoji: string) => void | Promise<void>;
}

export function EmojiPicker({ disabled = false, onSend }: Props) {
  const [recent, setRecent] = useState<string[]>(() => loadRecentEmojis());
  const [open, setOpen] = useState(false);
  const [cooldownUntil, setCooldownUntil] = useState(0);
  const dialogRef = useRef<HTMLDialogElement | null>(null);
  const pickerRef = useRef<HTMLElement | null>(null);

  const inCooldown = cooldownUntil > Date.now();
  const buttonDisabled = disabled || inCooldown;

  // Stable ref to the latest send fn so the picker's emoji-click listener
  // doesn't need to re-attach on every render. The e2e test dispatches
  // emoji-click on the picker right after clicking the react button, so the
  // listener has to be present on mount, not gated on `open`.
  const sendRef = useRef<(emoji: string) => Promise<void>>(async () => {});
  sendRef.current = async (emoji: string) => {
    if (buttonDisabled) return;
    setRecent(rememberEmoji(emoji));
    setCooldownUntil(Date.now() + COOLDOWN_MS);
    setOpen(false);
    dialogRef.current?.close?.();
    await onSend(emoji);
  };

  // Open/close the native <dialog> element when the open flag flips.
  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (open) {
      if (typeof dialog.showModal === "function" && !dialog.open) {
        dialog.showModal();
      }
    } else if (dialog.open) {
      dialog.close();
    }
  }, [open]);

  // Attach the emoji-click listener once on mount.
  useEffect(() => {
    const picker = pickerRef.current;
    if (!picker) return;
    const handler = (ev: Event) => {
      const detail = (ev as CustomEvent<{ unicode?: string }>).detail;
      const unicode = detail?.unicode;
      if (unicode) void sendRef.current(unicode);
    };
    picker.addEventListener("emoji-click", handler);
    return () => picker.removeEventListener("emoji-click", handler);
  }, []);

  // Tick the cooldown UI back to enabled.
  useEffect(() => {
    if (!inCooldown) return;
    const t = setTimeout(() => setCooldownUntil(0), cooldownUntil - Date.now());
    return () => clearTimeout(t);
  }, [inCooldown, cooldownUntil]);

  return (
    <div className="flex items-center gap-2 flex-wrap">
      <div id="recent-emojis" className="flex items-center gap-1">
        {recent.slice(0, 5).map((e) => (
          <button
            key={e}
            type="button"
            disabled={buttonDisabled}
            className="text-xl px-1.5 py-0.5 rounded hover:bg-panel disabled:opacity-50"
            onClick={() => void sendRef.current(e)}
          >
            {e}
          </button>
        ))}
      </div>
      <button
        id="reactbtn"
        type="button"
        disabled={buttonDisabled}
        className="text-xl px-2 py-1 rounded-md border border-[color:var(--accent)] hover:opacity-90 disabled:opacity-50"
        onClick={() => setOpen(true)}
      >
        {inCooldown ? "…" : "🎉"}
      </button>
      <dialog
        id="emojiDialog"
        ref={dialogRef}
        onClose={() => setOpen(false)}
        className="bg-transparent backdrop:bg-black/40 p-0 rounded-xl border-none"
      >
        <emoji-picker
          id="emojiPicker"
          ref={(el: HTMLElement | null) => {
            pickerRef.current = el;
          }}
        />
      </dialog>
    </div>
  );
}
