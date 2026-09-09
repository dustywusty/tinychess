import { bots, type BotId } from "@yourmove/chess/bots";
import { useEffect, useRef, useState } from "react";

export function BotPicker({ open, busy, onClose, onStart }: { open: boolean; busy: boolean; onClose: () => void; onStart: (id: BotId, color: "w" | "b") => void }) {
 const dialog = useRef<HTMLDialogElement>(null);
 const [selected, setSelected] = useState<BotId>("pip");
 const [color, setColor] = useState<"w" | "b">("w");
 useEffect(() => { if (open) dialog.current?.showModal(); else dialog.current?.close(); }, [open]);
 return <dialog ref={dialog} className="bot-dialog" onClose={onClose} aria-labelledby="bot-title">
  <div className="section-line"><h2 id="bot-title">Meet your next opponent.</h2><button type="button" className="text-button" onClick={onClose} aria-label="Close opponent picker">✕</button></div>
  <div className="bot-roster" role="radiogroup" aria-label="Opponent">{bots.map(bot => <label key={bot.id} className={"bot-option" + (selected === bot.id ? " chosen" : "")}>
   <span className="bot-emoji" aria-hidden="true">{bot.emoji}</span><span><strong>{bot.name}</strong><small>{bot.description}</small></span>
   <input type="radio" name="bot" value={bot.id} checked={selected === bot.id} onChange={() => setSelected(bot.id)} />
  </label>)}</div>
  <fieldset className="bot-color"><legend>Your pieces</legend>{(["w", "b"] as const).map(value => <label key={value}><input type="radio" name="color" checked={color === value} onChange={() => setColor(value)} />{value === "w" ? "White" : "Black"}</label>)}</fieldset>
  <button type="button" className="primary-button" disabled={busy} onClick={() => onStart(selected, color)}>{busy ? "Opening your board…" : "Let’s play ↗"}</button>
 </dialog>;
}
