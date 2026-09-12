#!/usr/bin/env bash
set -euo pipefail
: "${RUNNER_TEMP:?Set RUNNER_TEMP to a temporary directory}"
export PORT="${PORT:-18091}"
export EXPO_PUBLIC_API_URL="http://127.0.0.1:$PORT"
web_url="$EXPO_PUBLIC_API_URL"
mobile_url=http://127.0.0.1:8082
pids=()
# shellcheck disable=SC2329 # Invoked by the EXIT trap.
cleanup() {
  for pid in "${pids[@]}"; do kill "$pid" 2>/dev/null || true; done
}
trap 'cleanup' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

go build -o "$RUNNER_TEMP/bot-api" .
"$RUNNER_TEMP/bot-api" -static-dir web/dist > "$RUNNER_TEMP/bot-api.log" 2>&1 &
pids+=("$!")
pnpm --filter @yourmove/mobile exec expo start --web --port 8082 > "$RUNNER_TEMP/bot-mobile.log" 2>&1 &
pids+=("$!")
ready=false
for attempt in $(seq 1 60); do
  echo "Waiting for browser servers ($attempt/60)"
  if curl -fsS "$EXPO_PUBLIC_API_URL/healthz" > /dev/null && curl -fsS "$web_url/" > /dev/null && curl -fsS "$mobile_url/" > /dev/null; then
    ready=true
    break
  fi
  sleep 1
done
if [ "$ready" != true ]; then
  echo 'Browser servers did not become ready within 60 attempts.' >&2
  exit 1
fi

# Run both suites together and retain either failure, even if the other passes.
PW_BASE_URL="$web_url" pnpm --filter @yourmove/web exec playwright test &
web_tests=$!
pids+=("$web_tests")
PW_BASE_URL="$mobile_url" pnpm --filter @yourmove/mobile exec playwright test &
mobile_tests=$!
pids+=("$mobile_tests")
result=0
wait "$web_tests" || result=1
wait "$mobile_tests" || result=1
exit "$result"
