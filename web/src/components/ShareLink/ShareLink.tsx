import { useEffect, useRef, useState } from "react";

export function ShareLink() {
  const [notice, setNotice] = useState("");
  const timer = useRef<ReturnType<typeof setTimeout>>();
  useEffect(() => () => clearTimeout(timer.current), []);
  const share = async () => {
    try {
      if (navigator.share) await navigator.share({ title: "Your Move", text: "A little chess?", url: window.location.href });
      else { await navigator.clipboard.writeText(window.location.href); setNotice("Game link copied."); }
    } catch (error) {
      if (!(error instanceof Error && error.name === "AbortError")) setNotice("Copy the game address from your browser to invite a friend.");
    }
    clearTimeout(timer.current);
    timer.current = setTimeout(() => setNotice(""), 5000);
  };
  return <div className="share-control"><button type="button" className="outline-button" onClick={() => void share()}>Invite a friend <span aria-hidden="true">↗</span></button>{notice && <span role="status" className="share-notice">{notice}</span>}</div>;
}
