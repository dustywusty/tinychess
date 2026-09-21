# Web coach acceptance audit

Scope: the supplied visual-coach handoff, limited to the web application.
Status: the web implementation, opt-in evaluation gates, and container validation passed.
The feature remains disabled by default. No production deployment was performed.

## Required journeys

| Handoff case | Current evidence | Status |
| --- | --- | --- |
| 1. Weak move, common origin, legal response, unchanged history | `visual-coach.spec.ts` uses a real native engine and verifies the final game history | Verified |
| 2. Honest treatment of a strong move | `TestGoodMoveAndAlternativeExercise` checks preferred and comparable moves | Verified in deterministic tests and the final live preferred-move narration sample |
| 3. Legal knight target or concrete clarification | Legal candidate enumeration uses the displayed position; context tests cover nested exploration | Live validation passed candidate and clarification cases; scope is the recorded small corpus |
| 4. Hint without hidden answers in board, text, or transport | Projection tests and adversarial routing tests remove scores, lines, evidence, and annotations | Deterministic projections verified; recorded live hint paraphrases pass |
| 5. Sandboxed legal exercise with sound alternative acceptance | Service tests cover the acceptance policy; Chrome asserts no live move requests | Verified under the documented compatible-depth policy |
| 6. Live updates, stale requests, cancellation, reconnect, repeated events | Chrome checks pinned review and reload; reducer and service tests check revisions; engine cancellation releases capacity | Verified for tested journeys, including a lost committed response retried with the same request ID |
| 7. Orientation, score perspective, special moves, draws/history | Chrome flips arrows; tests cover nested Black/White scores, mate, castling, en passant, promotion, and history identity | Rules regressions cover stalemate and fivefold repetition; native Arasan confirms the stalemate candidate |
| 8. Ownership, game policy, unregistered actions | Handler tests reject unauthorized requests and recheck ownership after delayed analysis | Verified for the existing anonymous seat trust model |
| 9. Provider failures preserve controls | Unit tests cover narration fallback, invalid Jev actions, hidden narration, and engine cancellation | Chrome verifies transport recovery and injected engine/language errors; unit tests verify provider deadlines and incomplete narration. Live providers also passed the recorded browser journey |
| 10. Buttons avoid inference; Jev navigation works without narration | Chrome proves local navigation has no network calls; deterministic router tests exercise language selection | Live Chrome verifies language navigation and no navigation narration; latest development corpus passes 28/28 routes |

## Other delivery requirements

| Requirement | Current evidence or remaining work |
| --- | --- |
| Repository decision record | `visual-coach-decisions.md` identifies stack, services, hosting, budgets, skill revision, and risks |
| TypeSafe skill and current API contract | Skill installed and locked; HTTP Choice adapter uses the inspected TypeSafe contract |
| Rules and engine remain chess authority | `ValidateLine` replays legal moves; evidence contains only rules facts and identified search results |
| Analysis provenance | Each native search records revision, root identity, root side, limits, elapsed time, depth, and status |
| Separate authoritative game and coach state | Web controller has no live move operation; game history stays in the existing store |
| Progressive reveal | Server projections gate every line, fact, score, and arrow; explicit hint phrases restrict model options |
| CopilotKit and grounded narration | Headless CopilotKit consumes Go AG-UI events; narration receives the selected evidence packet |
| Cost and latency controls | Engine/provider concurrency, per-seat budgets, cache limits, deadlines, token limits, and usage telemetry are implemented |
| Cache compatibility | Cache is scoped to one fixed engine configuration and keys the complete history plus restricted candidate |
| Feature flag and configuration | Web and API flags remain disabled by default; setup and price configuration are documented |
| Demo script | `visual-coach.md` contains the board review and exercise sequence |
| Deterministic verification | Full web production build, WASM runtime, native engine, Go race checks, and browser regressions pass |
| Live evaluation and measured thresholds | Live development evaluation: Jev routes 28/28, targets 2/2, clarification 2/2. Actual session latency and provider cost estimate recorded. Separate validation passed 19/20 routes. Final narration audit returned four generated explanations and two fallbacks with no observed unsupported claims or unit errors. Measured opt-in release gates are recorded |
| Deployment validation | Frontend image builds; ARM64 and AMD64 API images build and pass health/native-engine review checks. The ARM64 container with production web assets passes 20 browser tests |

The web-only scope does not require a native coach UI. The existing mobile application has no implementation changes.
