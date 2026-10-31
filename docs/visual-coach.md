# Visual coach preview

The web coach shows a compact report below the board after each player move in an eligible game. It is disabled by default.
Background analysis leaves the live board playable. The report shows a move grade, the played move’s score, and the engine’s preferred alternative.
Select a move card to inspect it on the board; expand **Ask the coach** for questions and exercises.
Bot-versus-bot games do not trigger automatic reports. Hints-only mode disables automatic answer reports.
The server checks seat ownership and permits coaching only in computer games or completed games.
The native Expo client does not yet include the coach UI.

## Local configuration

Install workspace dependencies with `corepack pnpm@9.15.0 install --frozen-lockfile`.
Build the native engine with `corepack pnpm@9.15.0 engine:native`.
The build requires Git, Node, `patch`, and a C++17 compiler. It downloads the pinned source and network.

If you use `.env.coach.local`, load it into the API terminal first:

```sh
set -a
source .env.coach.local
set +a
```

The file is ignored by Git and Docker builds. Keep provider credentials in the API environment.

Start the API:

```sh
COACH_ENABLED=1 \
COACH_ENGINE_PATH="$PWD/engine/arasan/generated/native/arasan" \
go run .
```

Start the web client in another terminal:

```sh
VITE_COACH_ENABLED=1 corepack pnpm@9.15.0 web
```

The existing browser opponent still requires the separate WASM build described in the README.
The native coach engine does not replace that opponent.

Optional server configuration:

| Variable | Purpose |
| --- | --- |
| `TYPESAFE_API_KEY` | Enables Jev language controls |
| `TYPESAFE_MODEL` | Selects a TypeSafe model; defaults to the documented `jev-latest` alias |
| `OPENAI_API_KEY` | Authorizes optional narration |
| `COACH_NARRATION_MODEL` | Explicit Responses API model for narration; no default |

The deterministic buttons work without provider credentials.
Missing narration configuration retains the verified lesson caption.
The **Why?** button streams an explanation when narration is configured.
Hidden-answer lessons use deterministic captions and never call the narrator.
Explicit hint requests constrain the action registry before Jev selects an action.
Language-routed explanations use the same narration path as the **Why?** button.

## Architecture and limits

`internal/coach` owns engine searches, verified lines, cached lessons, projections, budgets, and provider adapters.
`internal/handlers/coach.go` checks existing anonymous seat capabilities before every server action.
`coach_stream.go` adapts the service to AG-UI. The runtime discovery route exposes only static agent metadata.
The CopilotKit headless context connects that transport to `web/src/coach/useCoach.ts`.
The client reducer keeps review state outside the authoritative game store.

Each lesson retains the full position history prefix and its SHA-256 identity.
Later game moves do not replace a pinned lesson. Return reads the latest game store.
A new request or return invalidates delayed responses. Server revisions reject replayed mutations.
Reload starts a clean review session; lessons expire after 30 minutes.

A projection contains only permitted information. Revealed lines include validated local navigation frames.
Comparison, continuation, back, and return buttons make no inference calls. Revealed navigation also makes no network calls.
Line arrows use solid green for the played move and dashed purple for the alternative, with accessible text descriptions.
All arrows follow board orientation. The coach adds no motion.

The engine allows two concurrent processes, 16 MB hash tables, and a 1.5-second search budget.
A process has a five-second total deadline. Analysis runs until an explicit stop at 1.5 seconds to avoid the engine’s early depth-2 exit on a single restricted candidate. The analysis cache holds 256 entries for one fixed engine configuration.
The cache key includes the full move history and restricted candidate. It does not reuse another budget or engine instance.
The service retains at most 1,000 lessons and permits 60 requests per seat/game per minute. Routing and analysis within one request share one slot.
Repeated reviews can reuse cached analysis immediately. Providers have ten-second deadlines.

Exercise attempts never use the live move endpoint.
The engine analyzes each uncached legal answer before judging it.
At matching search depth, a centipawn result within 50 centipawns of the preference is accepted.
Matching preferred moves are accepted. Incompatible depth or score types produce an uncertain result, not a rejection.
Mate scores remain separate from centipawn scores.

Natural-language candidate exploration uses the displayed position, including earlier exploration and local navigation.
Exercise attempts always start from the lesson origin. Scores retain the reviewed mover’s perspective across branches.
Ambiguous requests offer legal moves with SAN and UCI labels.
Candidate exploration preserves the original played and alternative lines.
Rule-derived evidence links legal moves, captures, checks, and checkmate to exact line positions.
Each search records engine revision, elapsed search time, depth, and score perspective.
The coach does not infer tactical motifs such as forks or pins.

## Deployment

The API Dockerfile builds native Arasan and includes its network and runtime notices.
It uses the Debian 12 C++ runtime image for the native engine's shared libraries.
The frontend Dockerfile accepts `--build-arg VITE_COACH_ENABLED=1`.
Set `COACH_ENABLED=1` and provider variables on the API service only after release validation.
For cost estimates, configure each deployed model's input and output prices per million tokens:

- `COACH_JEV_INPUT_USD_PER_MILLION` and `COACH_JEV_OUTPUT_USD_PER_MILLION`
- `COACH_NARRATION_INPUT_USD_PER_MILLION` and `COACH_NARRATION_OUTPUT_USD_PER_MILLION`

Provider logs separate latency, outcome, token counts, and estimated cost. Unknown usage or prices appear as `null`.
The checked-in deployment configuration keeps the preview disabled.

Docker access is now available. The frontend image and both ARM64 and AMD64 API images build successfully.
The container passed health, native-engine review, return, unchanged-history, and production web browser checks.
The native engine build uses Clang because the pinned source has GCC-incompatible ARM vector conversions.
The pinned Emscripten 4.0.14 toolchain produced the required WASM assets locally.
The complete web production build, WASM runtime tests, native engine tests, and web browser regression suite passed.

## Demo

1. Open a completed game containing `1. f3 e5 2. g4 Qh4#` as a seated player.
2. Select White's second move and click **Review move**.
3. Select **Your move**, **Alternative**, and **Compare**.
4. Select **Continue** to see a legal response.
5. Select **Try it**, then play a legal move on the board.
6. Select **Return to game** and verify the original result and move history.
7. With Jev configured, ask “Show their response” or “Go back to my game.”
8. With narration configured, select **Why?**.

## Verification and release status

See [evaluation results](visual-coach-evaluation.md), [implementation decisions](visual-coach-decisions.md), and the [acceptance audit](visual-coach-acceptance.md).
This is an opt-in preview. The evaluation report records live routing, narration, latency, cost estimates, and known limitations.
The recorded opt-in gates passed. Keep the flag disabled unless you explicitly enable this preview for your deployment.

## Move grades

Grades compare the played move with the engine choice from the same pre-move position.
They describe the move’s loss against that choice, not whether the player is already winning or losing.
The preferred move is labeled Best move. For two centipawn results at depth 8 or above, losses up to 0.5 pawns are Good move; above 0.5 is Inaccuracy, above 1 is Mistake, and above 2.5 is Blunder.
These are UI grading thresholds for a bounded search, not percentages or guarantees of perfect play.
Mate scores are handled separately. Shallow or incompatible results do not receive a confident numerical grade.
The primary explanation names the alternative and comparison; technical search details are collapsed.


## Ask for the next move

In a computer game, ask “What should I play next?” or select **Best move now**.
The coach analyzes the current live position, including the position before the first move.
It shows a suggestion and its evaluation. **Preview suggested move** changes only the review board.
Use **Return to game** to resume play and make your own move.

Follow-up explanations use the analyzed position. If the game advances, a notice offers
**Analyze latest position**. Asking for the best move now from a past review also starts
analysis of the latest live position. Requests to explain a past move still open its review.
Completed games have no next move; review a played move instead. Existing seat and game eligibility rules apply.

The sidebar is 400 pixels wide on large screens. Its status, coach, reactions, and move-history cards collapse independently.


## Jev teaching decisions

A language request asks Jev four Choice questions in one API call: the application action,
the teaching view, the line to focus on, and one suggested follow-up. The backend validates
all choices. It derives highlighted squares from legal engine lines; the model supplies no coordinates or code.

The available teaching views keep the current position, demonstrate a move, compare alternatives,
show a possible reply, or start an exercise. Explicit hidden-answer requests override presentation choices.
A request for a best move cannot silently become an exercise unless answers are hidden.
Unsupported requests and ambiguous move references do not change the board.

The sidebar shows the chosen view, focused square, conversation, and follow-up question.
CopilotKit continues to transport state and narration through its headless integration.
The custom conversation retains up to 12 messages while a lesson is open; the server retains the last six
question/action/view summaries to resolve follow-ups. Return clears the client conversation.
Exercise mode hides earlier chat answers. Browser-only navigation clears the model-selected focus.
