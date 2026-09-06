import { useEffect, useRef, useState } from "react";
import "emoji-picker-element";
import emojiDataURL from "emoji-picker-element-data/en/emojibase/data.json?url";
import { loadRecentEmojis, rememberEmoji } from "../../lib/recentEmojis";
import { useUiStore } from "../../state/uiStore";

const COOLDOWN_MS = 5000;

interface Props {
  disabled?: boolean;
  onSend: (emoji: string) => Promise<boolean>;
}

export function EmojiPicker({ disabled = false, onSend }: Props) {
  const [recent, setRecent] = useState<string[]>(() => loadRecentEmojis());
  const [open, setOpen] = useState(false);
  const [cooldownUntil, setCooldownUntil] = useState(0);
  const [sending, setSending] = useState(false);
  const sendLock = useRef(false);
  const theme = useUiStore((state) => state.theme);
  const dialogRef = useRef<HTMLDialogElement | null>(null);
  const pickerRef = useRef<HTMLElement | null>(null);

  const inCooldown = cooldownUntil > Date.now();
  const buttonDisabled = disabled || inCooldown || sending;

  // Stable ref to the latest send fn so the picker's emoji-click listener
  // doesn't need to re-attach on every render. The e2e test dispatches
  // emoji-click on the picker right after clicking the react button, so the
  // listener has to be present on mount, not gated on `open`.
  const sendRef = useRef<(emoji: string) => Promise<void>>(async () => {});
  sendRef.current = async (emoji: string) => {
    if (buttonDisabled || sendLock.current) return;
    sendLock.current = true;
    setSending(true);
    setOpen(false);
    dialogRef.current?.close?.();
    try {
      if (await onSend(emoji)) {
        setRecent(rememberEmoji(emoji));
        setCooldownUntil(Date.now() + COOLDOWN_MS);
      }
    } finally { sendLock.current = false; setSending(false); }
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
    <div className="emoji-controls">
      <div id="recent-emojis" className="quick-reactions">
        {[...new Set([...recent, "👋", "🤔", "🔥", "😂", "👏", "🤝"])].slice(0, 6).map((e) => (
          <button
            key={e}
            type="button"
            disabled={buttonDisabled}
            className="reaction-button"
            aria-label={`Send ${e}`}
            onClick={() => void sendRef.current(e)}
          >
            {e}
          </button>
        ))}
      </div>
      <div className="reaction-picker-footer"><span>{recent.length ? "Your recent favorites" : "A few favorites"}</span><button
        id="reactbtn"
        type="button"
        disabled={buttonDisabled}
        className="more-reactions text-button"
        aria-label="More emoji"
        onClick={() => setOpen(true)}
      >
        {sending ? "Sending…" : inCooldown ? "A little breather…" : "More emoji  +"}
      </button>
      </div>
      <dialog
        id="emojiDialog"
        ref={dialogRef}
        onClose={() => setOpen(false)}
        className="emoji-dialog"
        aria-label="Choose an emoji"
      >
        <div className="emoji-picker-heading"><h2>Say it your way.</h2><button type="button" className="dialog-close" aria-label="Close emoji picker" onClick={() => setOpen(false)}>×</button></div>
        <emoji-picker
          id="emojiPicker"
          className={theme}
          data-source={emojiDataURL}
          ref={(el: HTMLElement | null) => {
            pickerRef.current = el;
          }}
        />
      </dialog>
    </div>
  );
}
