import { bots, MAX_BOT_ELO, MIN_BOT_ELO, type BotId } from "@yourmove/chess/bots";
import { useEffect, useRef, useState, type CSSProperties } from "react";
import type { BotSettings } from "@yourmove/protocol";

// Range fill for the custom slider track.
const fill = (value: number, min: number, max: number) => ({ "--fill": `${(value - min) / (max - min) * 100}%` }) as CSSProperties;

export function BotPicker({ open, busy, onClose, onStart }: { open: boolean; busy: boolean; onClose: () => void; onStart: (id: BotId, color: "w" | "b", settings: BotSettings) => void }) {
 const dialog = useRef<HTMLDialogElement>(null);
 const [selected, setSelected] = useState<BotId>("pip");
 const [color, setColor] = useState<"w" | "b">("w");
 const [custom, setCustom] = useState(false);
 const [elo, setElo] = useState(1500);
 const [battle, setBattle] = useState(false);
 const [whiteBot, setWhiteBot] = useState<BotId>("pip");
 const [delay, setDelay] = useState<number>();
 const pace = delay ?? 1500;
 useEffect(() => { if (open) dialog.current?.showModal(); else dialog.current?.close(); }, [open]);
 return <dialog ref={dialog} className="bot-dialog" onClose={onClose} aria-labelledby="bot-title">
  <div className="section-line"><h2 id="bot-title">{battle ? "Watch a bot battle." : "A worthy opponent."}</h2><button type="button" className="text-button" onClick={onClose} aria-label="Close opponent picker">✕</button></div>
  {!battle && <p className="bot-intro">Five personalities. Practice anywhere, even offline.</p>}
  {!battle && <div className="bot-roster" role="radiogroup" aria-label="Opponent">{bots.map(bot => <label key={bot.id} className={"bot-option" + (selected === bot.id ? " chosen" : "")}>
   <span className="bot-emoji" aria-hidden="true">{bot.emoji}</span><span><strong>{bot.name}</strong><small>{bot.description}</small></span>
   <input type="radio" name="bot" value={bot.id} checked={selected === bot.id} onChange={() => setSelected(bot.id)} />
  </label>)}</div>}
  {!battle && <fieldset className="bot-color"><legend className="eyebrow">YOUR PIECES</legend>{(["w", "b"] as const).map(value => <label key={value} className={color === value ? "chosen" : ""}>
   <input type="radio" name="color" aria-label={value === "w" ? "White" : "Black"} checked={color === value} onChange={() => setColor(value)} /><span aria-hidden="true">{value === "w" ? "♔" : "♚"}</span>{value === "w" ? "White" : "Black"}
  </label>)}</fieldset>}
  {!battle && <label className="bot-switch">Custom difficulty<input type="checkbox" role="switch" checked={custom} onChange={event => setCustom(event.target.checked)} /></label>}
  {!battle && custom && <div className="bot-range">
   <label className="eyebrow" htmlFor="bot-elo">ELO TARGET · {elo}</label>
   <input id="bot-elo" type="range" min={MIN_BOT_ELO} max={MAX_BOT_ELO} step="50" value={elo} style={fill(elo, MIN_BOT_ELO, MAX_BOT_ELO)} onChange={event => setElo(Number(event.target.value))} aria-label="Elo target" aria-valuetext={`Elo ${elo}`} />
   <p>Approximate engine strength, not a measured player rating. Uses the chosen strength instead of personality difficulty.</p>
  </div>}
  <details className="bot-advanced">
   <summary>Advanced settings</summary>
   <label className="bot-battle-toggle"><input type="checkbox" checked={battle} onChange={event => setBattle(event.target.checked)} />Bot vs. bot</label>
   <p>Watch two computer opponents play a test game.</p>
   {battle && ([{ id: "white-bot", label: "White bot", value: whiteBot, set: setWhiteBot }, { id: "black-bot", label: "Black bot", value: selected, set: setSelected }]).map(side => <div className="bot-select" key={side.id}><label htmlFor={side.id}>{side.label}</label><select id={side.id} value={side.value} onChange={event => side.set(event.target.value as BotId)}>{bots.map(bot => <option key={bot.id} value={bot.id}>{bot.emoji} {bot.name}</option>)}</select></div>)}
   <div className="bot-range">
    <label className="eyebrow" htmlFor="bot-delay">MINIMUM MOVE INTERVAL · <output>{pace === 0 ? "AS FAST AS POSSIBLE" : `${pace / 1000}S`}</output></label>
    <input id="bot-delay" type="range" min="0" max="5000" step="250" value={pace} style={fill(pace, 0, 5000)} onChange={event => setDelay(Number(event.target.value))} aria-label="Minimum move interval" aria-valuetext={`${pace / 1000} seconds`} />
    <p>Thinking may take longer.{!battle && delay === undefined ? " Regular games use each bot’s natural pace until you adjust this." : ""}</p>
   </div>
  </details>
  <button type="button" className="primary-button" disabled={busy} onClick={() => onStart(selected, battle ? "w" : color, { playerBotId: battle ? whiteBot : undefined, moveDelayMs: delay ?? (battle ? 1500 : undefined), targetElo: !battle && custom ? elo : undefined })}>{busy ? "Opening your board…" : battle ? "Start bot battle ↗" : "Let’s play ↗"}</button>
 </dialog>;
}
