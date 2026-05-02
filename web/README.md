# tinychess web

The React + TypeScript SPA. Phase 1 (in progress) extracts the inline React
from `internal/templates/game.html` into a properly tooled Vite project.

## Dev

```sh
pnpm install
pnpm dev          # Vite at :5173, proxies /api/* and /new to :8080
```

In a second terminal, run the Go server:

```sh
make run          # :8080
```

## Build

```sh
pnpm build        # outputs to web/dist
```

`make build` invokes this transitively via the `web-build` target before
compiling the Go binary.

## Layout

```
web/
  src/
    api/            HTTP + SSE wrappers
    components/     UI components (Board, GameStatus, EmojiPicker, ShareLink, Chat)
    state/          Zustand stores (gameStore, uiStore, annotationStore)
    pages/          Game, Home, Review (Phase 2)
    types/          chess + SSE event types
    lib/            session, theme, recentGames, recentEmojis utils
    main.tsx        entrypoint
    App.tsx         root
    index.css       tailwind + theme variables
```

See `MIGRATION_NOTES.md` for the parity checklist.
