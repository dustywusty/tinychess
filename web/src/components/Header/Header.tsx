import { useEffect } from "react";
import { useUiStore } from "../../state/uiStore";
import { ThemePicker } from "../ThemePicker/ThemePicker";

interface Props {
  rightSlot?: React.ReactNode;
}

export function Header({ rightSlot }: Props) {
  const menuOpen = useUiStore((s) => s.menuOpen);
  const toggleMenu = useUiStore((s) => s.toggleMenu);
  const closeMenu = useUiStore((s) => s.closeMenu);

  useEffect(() => {
    if (!menuOpen) return;
    const onClick = (e: MouseEvent) => {
      const target = e.target as HTMLElement | null;
      if (!target?.closest(".header-actions, .header-menu-toggle")) {
        closeMenu();
      }
    };
    window.addEventListener("click", onClick);
    return () => window.removeEventListener("click", onClick);
  }, [menuOpen, closeMenu]);

  return (
    <header className="px-4 py-3 border-b border-[color:var(--btn-border,_rgba(255,255,255,0.1))] flex items-center justify-between bg-panel sticky top-0 z-20">
      <a href="/" className="text-base font-semibold">
        Tiny Chess
      </a>
      <button
        type="button"
        className="header-menu-toggle md:hidden w-9 h-9 rounded-md bg-[color:var(--accent)] text-bg font-bold"
        aria-label="Toggle menu"
        aria-expanded={menuOpen}
        onClick={toggleMenu}
      >
        ☰
      </button>
      <div
        className={`header-actions ${
          menuOpen ? "flex" : "hidden"
        } md:flex items-center gap-2 absolute md:static right-3 top-14 md:top-auto bg-panel md:bg-transparent border md:border-0 border-[color:var(--btn-border,_rgba(255,255,255,0.1))] p-2 md:p-0 rounded-xl flex-col md:flex-row min-w-[180px] md:min-w-0 shadow-lg md:shadow-none z-30`}
      >
        {rightSlot}
        <ThemePicker />
      </div>
    </header>
  );
}
