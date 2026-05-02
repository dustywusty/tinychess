import { useState } from "react";

export function ShareLink() {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      /* clipboard unavailable */
    }
  };

  return (
    <button
      type="button"
      className="px-3 py-2 rounded-md bg-panel border border-[color:var(--accent)] hover:opacity-90 text-sm"
      onClick={handleCopy}
    >
      {copied ? "Copied" : "Copy link"}
    </button>
  );
}
