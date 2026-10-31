#!/usr/bin/env bash
set -euo pipefail
cd /opt/yourmove
exec 9>/opt/yourmove/deploy.lock
flock -n 9 || { echo 'Another deployment is running.' >&2; exit 1; }
backend="${1:?Pass a backend image digest}"
frontend="${2:?Pass a frontend image digest}"
release="${3:?Pass the full Git commit}"
if ! [[ "$backend" =~ ^ghcr.io/dustywusty/tinychess@sha256:[a-f0-9]{64}$ ]] ||
   ! [[ "$frontend" =~ ^ghcr.io/dustywusty/tinychess@sha256:[a-f0-9]{64}$ ]] ||
   ! [[ "$release" =~ ^[a-f0-9]{40}$ ]]; then
  echo 'Only pinned Tinychess images and a full commit are accepted.' >&2
  exit 1
fi

docker pull "$backend"
docker pull "$frontend"
for image in "$backend" "$frontend"; do
  test "$(docker inspect --format '{{index .Config.Labels "org.opencontainers.image.revision"}}' "$image")" = "$release"
done

mkdir -p "releases/$release"
container_id=$(docker create "$frontend" /unused)
trap 'docker rm "$container_id" >/dev/null' EXIT
docker cp "$container_id:/site" "releases/$release/"
test -s "releases/$release/site/index.html"
test -s "releases/$release/site/arasan/arasan.wasm"
test -s "releases/$release/site/arasan/worker.js"
chmod -R a+rX "releases/$release"

if docker compose ps --status running --services | grep -qx db; then
  ./backup.sh
fi
cp -p .env .env.previous
python3 - "$backend" "$frontend" "$release" <<'PY'
from pathlib import Path
import sys
p = Path('.env')
values = dict(line.split('=', 1) for line in p.read_text().splitlines() if line and not line.startswith('#'))
values.update(zip(('TINYCHESS_IMAGE', 'FRONTEND_IMAGE', 'RELEASE'), sys.argv[1:]))
p.write_text(''.join(f'{key}={value}\n' for key, value in values.items()))
p.chmod(0o600)
PY

rollback() {
  cp -p .env.previous .env
  docker compose up -d --wait --wait-timeout 120 app
}
if ! docker compose config --quiet || ! docker compose up -d --wait --wait-timeout 120 app; then
  rollback
  exit 1
fi
if ! curl --fail --silent --show-error http://127.0.0.1:18080/api/version |
  python3 -c 'import json,sys; assert json.load(sys.stdin)["commit"] == sys.argv[1]' "$release"; then
  rollback
  exit 1
fi
echo "Deployed $release on 127.0.0.1:18080"
