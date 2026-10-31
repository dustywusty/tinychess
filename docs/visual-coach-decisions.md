# Visual coach implementation decisions

Date: 2026-09-21. Status: web preview implemented, disabled by default.

## Repository findings

The web client uses React 18, Vite, Zustand, and a custom accessible board in `web/src/components/Board/Board.tsx`.
The Expo client uses a separate native board. Both clients use the Go API and SSE game updates.
There are no clocks, accounts, entitlements, CopilotKit dependencies, or agent runtime in this repository.
Anonymous client IDs identify seats. The coach must use that existing trust model without assigning new seats.
Go chess rules own authoritative moves. Client chess.js validates board interaction.
`packages/coach` contains an unused phrase router and hint policy. Its Stockfish design notes predate the current Arasan integration.

Arasan is pinned to `361840f407b588252a48b25bb3000c751a470c89` under MIT terms.
Its current WASM adapter already preserves move history and supports bounded searches and candidate restrictions.
At discovery, the API image contained only a static Go binary. Browser analysis cannot enforce hidden-answer projection.

## Implementation plan

1. Add server analysis, verified lines, and a web review controller. Reuse the main board and existing seat authorization.
2. Add a bounded Jev HTTP adapter. Keep buttons deterministic and validate selected action IDs in application code.
3. Connect a supported CopilotKit transport to the controller. Add optional evidence-bound narration with a deterministic fallback.
4. Add projected hints, sandbox exercises, concurrency tests, provider tests, and release documentation.

## Decisions

Start with the web client. Keep Expo gameplay intact; native coach UI requires separate platform verification.
Run a pinned native Arasan process on the API host. Configure its executable explicitly and never construct shell commands from requests.
Limit engine concurrency, search time, provider timeout, prompt length, and requests per seat.
Analyze only on demand. Keep lessons in bounded, expiring server memory; reconnect starts a fresh review.
Authorize every request before analysis or cached lesson access. Permit seated players in computer games and completed games only.
Use the full move prefix and its hash as position identity. Preserve historical review when later moves arrive.
Use typed mate and centipawn scores. Never calculate centipawn loss from a mate score.
Return only the current reveal projection. Exercise answers and unrevealed lines stay on the server.
Keep review state separate from Zustand game state. Review actions cannot call the move API.
Provider configuration stays on the server. No new billing or persistence schema is required.

Target budgets are 1.5 seconds per engine search, two concurrent searches, and 10 seconds per semantic request.
These are limits, not measured latency guarantees. Provider costs require live measurements before release.

## Integration risks

The existing anonymous identity is a bearer capability, not account authentication.
The deployment must include the native engine and network before enabling the feature.
Native CopilotKit rendering is not assumed to support web DOM components.
Live model accuracy, narration claims, and cost need credentials and a reproducible evaluation run.
The feature flag must remain disabled until release checks pass.

## TypeSafe provenance

Installed with `npx skills add typesafe-ai/skills --skill typesafe-ai --agent codex --yes`.
The recorded skill hash is `9cd84c5e535dec8dec59917c110f9c00b4a61faadb86b432ec7e41051170af12` in `skills-lock.json`.
The implementation reads the current TypeSafe HTTP API, Choice, model, and function-calling documentation.
Choice fits bounded action selection. Score and Noul are not permission checks and are unnecessary for deterministic navigation.

References: [TypeSafe API](https://docs.typesafe.ai/api), [Choice](https://docs.typesafe.ai/primitives/choice),
[function calling](https://docs.typesafe.ai/cookbooks/function_calling), and [CopilotKit tools](https://docs.copilotkit.ai/reference/hooks/useFrontendTool).

The web adapter uses CopilotKit 1.73.0 with its headless context and a Go implementation of runtime discovery and AG-UI run events.
Its transitive AG-UI client is 0.0.59. The full chat UI is not imported because its bundled styles conflict with Tailwind 3.
Optional narration uses the [Responses API](https://developers.openai.com/api/docs/guides/streaming-responses) and receives only revealed evidence.
