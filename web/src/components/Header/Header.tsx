import type { ReactNode } from "react";
import { Piece } from "../Piece";
import { ThemePicker } from "../ThemePicker/ThemePicker";

export function Header({ rightSlot }: { rightSlot?: ReactNode }) {
  return <header className="site-header">
    <a href="/" className="brand" aria-label="Your Move home"><span className="brand-mark"><Piece piece="n" size={27} /></span><span>your move<span className="brand-period">.</span></span></a>
    <div className="header-right">{rightSlot ?? <span className="brand-tag">CHESS, TOGETHER</span>}<ThemePicker /></div>
  </header>;
}
