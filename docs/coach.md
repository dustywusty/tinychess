# Chess coach

The coach is a position-aware product feature, not a chatbot.

## Responsibilities

Stockfish supplies chess truth: evaluation, mate score, best move, MultiPV,
principal variations, and move-loss classification. The coach policy decides
whether a position deserves attention and which information may be revealed.
An optional LLM turns a small structured analysis into concise teaching copy.
The LLM never selects or validates moves.

```text
current FEN + recent moves
          |
       Stockfish
          |
 normalized analysis + interesting-position classifier
          |
 deterministic policy + cached progressive artifacts
          |
 optional bounded provider call
```

## Supported intents

`HINT`, `ANOTHER_HINT`, `WHY_BAD`, `WHY_GOOD`, `EXPLAIN_POSITION`,
`EXPLAIN_LAST_MOVE`, `WHAT_SHOULD_I_NOTICE`, `CHESS_RULE`, `OFF_TOPIC`, and
`REPEATED`.

`packages/coach` routes a deliberately small phrase vocabulary. Unknown,
off-topic, empty, and prompts longer than 240 Unicode code points receive canned
responses without inference. The transport should eventually offer explicit
buttons for most intents so free-text routing becomes exceptional.

## Per-position policy

- At most two distinct contextual questions
- At most three hints
- Identical normalized questions are canned as repeated
- A FEN/position-key change resets counters
- After limits, the response asks the player to decide and move

Hints reveal concept, attention, then direction. A position artifact contains
all three; serving later hints is a cache read, not another model call.

## Cost and failure controls

- Analyze only interesting positions; routine good moves are silent
- Cache by normalized position, engine build/options, analysis depth, player
  perspective, strength band, and artifact schema version
- Send FEN, a few moves, engine lines, and structured state—not chat history
- Generate observation, three hints, and explanations in one bounded call
- Enforce provider timeout, small output limit, and schema validation
- Record input/output tokens, latency, provider/model, cache hit, and estimated
  cost for every attempted call
- If Stockfish or the LLM is unavailable, gameplay continues and the coach uses
  a terse local fallback

CopilotKit can later render or transport approved coach events on React and
React Native. It does not own intent routing, limits, engine access, or usage.
