#!/usr/bin/env bash
set -euo pipefail
image_ref="${1:?Usage: bash scripts/smoke-image.sh IMAGE}"
container_id="$(docker run --detach --read-only --cap-drop ALL --security-opt no-new-privileges \
  --env PORT=8090 --publish 127.0.0.1::8090 "$image_ref")"
cleanup() {
  docker logs "$container_id"
  docker stop --time 15 "$container_id" >/dev/null
  docker rm "$container_id" >/dev/null
}
trap cleanup EXIT
address="$(docker port "$container_id" 8090/tcp)"
node scripts/smoke.mjs "http://$address"
docker exec "$container_id" /app/tinychess -healthcheck
test "$(docker inspect --format '{{.Config.User}}' "$container_id")" = "65532:65532"
