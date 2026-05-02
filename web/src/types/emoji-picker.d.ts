import type { DetailedHTMLProps, HTMLAttributes } from "react";

declare global {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace JSX {
    interface IntrinsicElements {
      "emoji-picker": DetailedHTMLProps<
        HTMLAttributes<HTMLElement>,
        HTMLElement
      >;
    }
  }
}

export {};
