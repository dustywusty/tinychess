# Coach (Phase 2) — implementation notes

This file tracks deviations from the handoff plan and the Hashbrown API
behaviors we relied on, so future Hashbrown bumps can be reviewed quickly.

## Hashbrown version

`@hashbrownai/core@0.4.1`, `@hashbrownai/react@0.4.1`. The React peer says
`>=18 <20`, so we're holding React at 18 (matches the locked decision).

## Schema deviations from the handoff plan

The plan called for `s.optional(...)`. The actual schema lib doesn't export
`optional` — the equivalent is `s.anyOf([type, s.nullish()])`. See
`src/coach/schemas.ts` for the pattern.

`s.string` / `s.integer` / `s.array` / `s.object` / `s.enumeration` all take a
description string as the first argument. Verified against the SDK type defs
under `node_modules/@hashbrownai/core/src/schema/base.d.ts`.

## Tool runtime

We use `useTool({ name, description, schema, handler, deps })`. Tools are
resolved client-side; the handler is async and receives the parsed schema
input. Hashbrown manages the `tool_use` round-trip — the model emits a tool
call in OpenAI delta format, the React side runs the handler, and the result
is fed back as a `tool_result`.

The Go backend (`internal/coach`) doesn't see tool calls directly — it just
streams the model's text + tool_use deltas in the OpenAI-shaped frame format
that Hashbrown consumes. (PR 9 ships text-only streaming; tool_use translation
from Anthropic's content blocks to OpenAI deltas is a follow-up if we hit
issues with browser-side tool calling.)

## Annotation bridge

`src/state/annotationStore.ts` is the bridge between Hashbrown chat messages
and the main board. `InlineBoard` (chat-side) pushes via `setAnnotation` on
viewport entry; `AnnotationLayer` (board-side) reads via `useAnnotationStore`.

Tool calls (`set_main_board_annotation`) write to the same store with
`sourceMessageId: "tool:set_main_board_annotation"` so we can distinguish
chat-driven vs tool-driven annotations if it matters later.
