import { bots, type BotId } from "@yourmove/chess/bots";
import { useEffect, useRef, useState } from "react";
import type { BotSettings } from "@yourmove/protocol";

export function BotPicker({ open, busy, onClose, onStart }: { open: boolean; busy: boolean; onClose: () => void; onStart: (id: BotId, color: "w" | "b", settings: BotSettings) => void }) {
 const dialog = useRef<HTMLDialogElement>(null);
 const [selected, setSelected] = useState<BotId>("pip");
 const [color, setColor] = useState<"w" | "b">("w");
 const [battle, setBattle] = useState(false);
 const [whiteBot, setWhiteBot] = useState<BotId>("pip");
 const [delay, setDelay] = useState<number>();
 useEffect(() => { if (open) dialog.current?.showModal(); else dialog.current?.close(); }, [open]);
 return <dialog ref={dialog} className="bot-dialog" onClose={onClose} aria-labelledby="bot-title">
  <div className="section-line"><h2 id="bot-title">{battle ? "Watch a bot battle." : "Meet your next opponent."}</h2><button type="button" className="text-button" onClick={onClose} aria-label="Close opponent picker">✕</button></div>
  {!battle && <div className="bot-roster" role="radiogroup" aria-label="Opponent">{bots.map(bot => <label key={bot.id} className={"bot-option" + (selected === bot.id ? " chosen" : "")}>
   <span className="bot-emoji" aria-hidden="true">{bot.emoji}</span><span><strong>{bot.name}</strong><small>{bot.description}</small></span>
   <input type="radio" name="bot" value={bot.id} checked={selected === bot.id} onChange={() => setSelected(bot.id)} />
  </label>)}</div>}
  {!battle && <fieldset className="bot-color"><legend>Your pieces</legend>{(["w", "b"] as const).map(value => <label key={value}><input type="radio" name="color" checked={color === value} onChange={() => setColor(value)} />{value === "w" ? "White" : "Black"}</label>)}</fieldset>}
  <details className="bot-advanced">
   <summary>Advanced settings</summary>
   <label className="bot-battle-toggle"><input type="checkbox" checked={battle} onChange={event => setBattle(event.target.checked)} />Bot vs. bot</label>
   <p>Watch two computer opponents play a test game.</p>
   {battle && ([{ id: "white-bot", label: "White bot", value: whiteBot, set: setWhiteBot }, { id: "black-bot", label: "Black bot", value: selected, set: setSelected }]).map(side => <div className="bot-select" key={side.id}><label htmlFor={side.id}>{side.label}</label><select id={side.id} value={side.value} onChange={event => side.set(event.target.value as BotId)}>{bots.map(bot => <option key={bot.id} value={bot.id}>{bot.emoji} {bot.name}</option>)}</select></div>)}
   <label className="bot-delay">Time between moves <output>{(delay ?? 1500) === 0 ? "As fast as possible" : `${(delay ?? 1500) / 1000} s`}</output>
    <input type="range" min="0" max="5000" step="250" value={delay ?? 1500} onChange={event => setDelay(Number(event.target.value))} aria-label="Time between moves" aria-valuetext={`${(delay ?? 1500) / 1000} seconds`} />
   </label>
   <p>Minimum interval; thinking may take longer.{!battle && delay === undefined ? " Regular games use each bot’s natural pace until you adjust this." : ""}</p>
  </details>
  <button type="button" className="primary-button" disabled={busy} onClick={() => onStart(selected, battle ? "w" : color, { playerBotId: battle ? whiteBot : undefined, moveDelayMs: delay ?? (battle ? 1500 : undefined) })}>{busy ? "Opening your board…" : battle ? "Start bot battle ↗" : "Let’s play ↗"}</button>
 </dialog>;
}
