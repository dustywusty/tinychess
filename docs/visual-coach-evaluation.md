# Visual coach evaluation

Date: 2026-09-21. Status: deterministic, opt-in live evaluation, and container gates passed, with model limitations recorded below.

## Deterministic checks

| Check | Result |
| --- | --- |
| Go test suite | Passed |
| Coach and handler race tests | Passed |
| Workspace JavaScript tests | Passed |
| Web, mobile, and shared package type checks | Passed |
| Complete web production build, including engine asset checks | Passed; bundle-size warning remains |
| Browser Arasan WASM runtime | Legal moves, cancellation, and restart passed |
| Full web browser regression suite | 20 passed; one paid-provider test skipped |
| Current coach browser journeys | 4 tests passed, including injected engine/language errors |
| Native Arasan mate and restricted-candidate searches | Passed |
| Chrome coach journey with real native engine | Passed, 2.6 seconds in the recorded run |

The browser journey covers comparison, legal continuation, hidden hints, an exercise attempt, and return to game.
It asserts that revealed navigation sends no agent requests and exercise moves send no live move requests.
The final authoritative move history remains unchanged.
A second Chrome journey verifies incoming computer moves, immediate cached review, return to the latest board, and clean reload.
A third drops a response after server commitment, then retries the same request ID.
It verifies that retry reveals one hint and preserves the committed revision.
A fourth injects AG-UI engine and language errors, then checks retry, local navigation, and return.
These injected events test the UI contract; they do not establish real provider uptime or timeout behavior.

Unit tests cover stale results, replayed server revisions, ownership, game policy, cache access, good moves, and comparable alternative answers.
Rules tests include en passant, castling, promotion, mate, illegal lines, score bounds, and repetition-aware identity.
Stalemate and automatic repetition draws stop displayed continuations and produce position-linked draw evidence.
Native Arasan confirms a restricted stalemate candidate with a zero centipawn score.
Provider tests reject unregistered Jev actions and prevent narration calls for hidden-answer lessons.
Context tests cover nested exploration, Black/White score normalization, mate score perspective, and explicit hints with a misrouted model answer.
Evidence tests limit checkmate claims to revealed, rules-validated line steps.
Engine cancellation tests verify that stopped searches release process capacity.
Provider deadline tests verify that Jev and narration release concurrency capacity.
Interrupted narration tests restore the evidence caption after partial output and an incomplete, failed, or closed stream.

## Live evaluation

Local provider configuration was supplied and used without publishing credential values.
The first Jev run passed 23/26 routes. Prompt revisions and shared production/evaluation action definitions improved that to 26/26.
A stricter browser test then found that recent context could override the current navigation request.
After clarifying that distinction, the expanded corpus passed 28/28 routes, 2/2 candidate targets, and 2/2 clarification cases.
These are development-set results after prompt tuning, not an independent release benchmark.

The paid Chrome journey passed with actual Jev, Arasan, and narration through CopilotKit.
It checks lesson creation, alternative selection, exact forward/back step changes, exercise mode, and return to the game.
Forward/back and alternative commands produced no narration events.
The test buffers provider responses at the browser route to inspect complete AG-UI events; it does not measure time to first visible text.

The recorded journey took 12.2 seconds. Six Jev calls took 6907, 118, 96, 122, 102, and 121 ms.
One narration call took 1271 ms. The initial Jev outlier is included rather than omitted from the report.
Jev used 8196 input tokens; narration used 568 input and 62 output tokens.
At the inspected prices, estimated provider cost was $0.000344232 for Jev and $0.0003264 for narration, or $0.000670632 total.
This excludes engine hosting, network costs, taxes, and any account-specific discounts.
Prices were checked on 2026-09-21: [TypeSafe models](https://docs.typesafe.ai/models) and [GPT-4.1 mini](https://developers.openai.com/api/docs/models/gpt-4.1-mini).

All three narration cases returned text in both runs. The second run retained the preferred move as good and qualified shallow analysis.
The weak-move explanation still used general mating-attack wording beyond the supplied concrete continuation; further evidence review is required.
No unsupported-claim rate is claimed from this small sample.
Provider token counts and latency were recorded locally. Runtime monetary telemetry remains null until price settings are configured; the session estimate above was calculated from recorded usage and inspected public prices.
The browser duration is not an inference latency measurement.

The reproducible corpus is `internal/coach/testdata/jev-corpus.json`.
The 28 cases include navigation paraphrases, ambiguous knight captures, explicit hints, unsupported commands, and requests to bypass the action registry.
They also cover recent interaction, good-move wording, uncertain evidence, and commands without an active lesson.
The shared knight position is replayed from legal moves before evaluation.
Run the paid experiment with:

```sh
COACH_LIVE_EVAL=1 go test ./internal/coach -run TestLiveJevCorpus -v
```

The adapter logs provider token usage, latency, and outcome. With explicit price configuration, it also logs estimated cost.
Missing usage or prices remain unknown instead of reporting a zero cost.
The experiment reports route accuracy, candidate target accuracy, and clarification accuracy separately.
It logs each request, expected answer, selected answer, error, and elapsed time. Inspect every mismatch before enabling language controls.
The small corpus is a smoke evaluation, not a release-quality benchmark.

The narration audit uses real engine evidence for a weak move, a preferred move, and an uncertainty question.
It logs each evidence packet beside the generated text. Review each chess claim against that packet.
A successful API response alone does not establish factual correctness.

Run both experiments after filling the local configuration:

```sh
set -a
source .env.coach.local
set +a
COACH_LIVE_EVAL=1 COACH_TEST_ENGINE="$PWD/engine/arasan/generated/native/arasan" \
  go test ./internal/coach -run 'TestLive(JevCorpus|NarrationEvidenceAudit)' -v \
  | tee /tmp/yourmove-live-evaluation.log
```

Before release, add further real positions and multi-turn adversarial spoiler sessions.
Measure route and target accuracy separately. Audit narration claims against supplied evidence.
Choose confidence and release thresholds from those measurements. No cookbook confidence cutoff is used.

## Additional validation and narration audit

A separate 20-case wording set passed 19/20 routes without further routing changes.
The miss was a harmless clarification instead of Compare for “Put both options on the starting board.”
Both candidate targets and both ambiguous-target cases passed. This is a small validation sample, not a population accuracy estimate.
Run it with `COACH_VALIDATION_SET=1 COACH_LIVE_EVAL=1 go test ./internal/coach -run TestLiveJevCorpus -v`.

The six-case narration audit exposed an invented fork and two score-unit errors in its first run.
After stronger instructions and a streaming guard, the fork request used the deterministic fallback and the sampled score units were correct.
The second run still overstated a mate score as a forced sequence. The guard now also rejects explicit forced-outcome phrases.
The final guarded live run passed: four generated explanations and two evidence fallbacks.
Manual inspection of these six final texts found no unsupported chess claims or incorrect score units.
The uncertainty question used a conservative fallback. This small sample does not establish general factual reliability.

Narration holds incomplete sentences before emitting them. Unsupported motif and forced-outcome phrases trigger the evidence caption.
This is a conservative lexical guard, not a general fact checker. Negated mentions may also fall back.
It does not establish correctness of claims outside its narrow vocabulary.

## Container validation

Docker access was restored. The frontend image and ARM64 API image built successfully.
The container passed its health check and a real native-engine review without provider credentials.
The complete production web build served by that container passed 20 browser tests; the paid provider journey was skipped in this run.
The API image filesystem was checked: `.env.coach.local` was absent.
The AMD64 API image also built and passed health, native-engine review, return, and unchanged-history checks under emulation.
The Linux ARM build uses Clang to support vector conversions in the pinned Arasan source.
The missing WASM assets were built with Emscripten 4.0.14. The complete production web build now passes.
The requested scope is web-only. Reload returns to a clean live view. Whole-game reports are outside the handoff scope.

See the [requirement-by-requirement audit](visual-coach-acceptance.md) for remaining verification and implementation work.

## Measured opt-in release gates

The frozen validation set requires at least 95% route accuracy and no misses in candidate, clarification, hint, or unsupported cases.
The observed result meets that gate: 19/20 overall, with the only miss asking for clarification instead of Compare.
The development corpus still requires every case to match; its recorded result is 28/28.
These are application release gates, not model confidence thresholds or statistical guarantees.

Narration requires no observed unsupported claims or score-unit errors in the six audited final texts.
Fallback captions are allowed and reported separately; they are not counted as generated explanations.
The recorded final sample meets this gate with four generated explanations and two fallbacks.
The lexical guard remains incomplete, so wider production monitoring is still needed.

A typical six-command session should remain below $0.01 in provider costs and each provider stage must finish within its 10-second deadline.
The measured session cost estimate was $0.000670632; the slowest observed stage in that session was 6.907 seconds.
These measurements support the opt-in feature only; they are not an SLA.
Both supported API architectures passed image builds and health checks. No images were pushed and no production deployment was performed.

## Board-integrated feedback revision

The coach now appears below the board, matching the existing theme. Background review follows the player’s latest move without opening review mode.
A Chrome regression verifies that opponent replies do not replace the player’s report, later player moves update it, and opening/closing a variation preserves the live game.
Desktop and phone screenshots were inspected; no horizontal overflow was observed.
The full browser suite passed 21 tests with one paid-provider test skipped. Production build and coach/handler race tests passed.
Native integration verifies that a restricted candidate now receives at least one second of analysis and reaches depth 8 or above instead of stopping at depth 2.
Live narration now receives the move assessment. The weak-move sample correctly explained that g4 allows Qh4# and named Nc3 as the alternative.
Earlier container validation predates this UI revision; no new image has been published.
