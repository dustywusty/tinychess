import { useState } from "react";

const dismissalKey = "yourmove.coach-dismissed";

export function CoachCard() {
  const [dismissed, setDismissed] = useState(() => {
    try { return localStorage.getItem(dismissalKey) === "1"; }
    catch { return false; }
  });
  const dismiss = () => {
    setDismissed(true);
    try { localStorage.setItem(dismissalKey, "1"); }
    catch { /* Dismiss for this visit even if storage is unavailable. */ }
  };
  if (dismissed) return null;
  return <section className="coach-card" aria-labelledby="coach-title">
    <button type="button" className="coach-dismiss" aria-label="Dismiss chess coach" onClick={dismiss}>×</button>
    <div className="section-line coach-header"><span className="eyebrow">A LITTLE WISDOM</span><span className="soon-badge">COMING SOON</span></div>
    <div className="coach-content"><span className="coach-icon" aria-hidden="true">✳</span><div>
      <h2 id="coach-title">Meet your chess coach.</h2>
      <p>A nudge when you need it. A little more “aha” in every game.</p>
    </div></div>
  </section>;
}
