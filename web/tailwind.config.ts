import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        accent: "var(--accent)",
        bg: "var(--bg)",
        panel: "var(--panel)",
        text: "var(--text)",
        primary: "var(--primary)",
        tertiary: "var(--tertiary)",
        sq1: "var(--sq1)",
        sq2: "var(--sq2)",
        sq3: "var(--sq3)",
      },
    },
  },
  plugins: [],
} satisfies Config;
