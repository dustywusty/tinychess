#!/usr/bin/env bash
# Forced SSH command: this key can deploy this repository's images, not run a shell.
set -euo pipefail
read -r -a args <<< "${SSH_ORIGINAL_COMMAND:-}"
if [[ ${#args[@]} != 4 || ${args[0]:-} != deploy ]]; then
  echo 'Expected: deploy BACKEND_DIGEST FRONTEND_DIGEST COMMIT' >&2
  exit 1
fi
if ! [[ "${args[1]}" =~ ^ghcr.io/dustywusty/tinychess@sha256:[a-f0-9]{64}$ ]] ||
   ! [[ "${args[2]}" =~ ^ghcr.io/dustywusty/tinychess@sha256:[a-f0-9]{64}$ ]] ||
   ! [[ "${args[3]}" =~ ^[a-f0-9]{40}$ ]]; then
  echo 'Only pinned Tinychess images and a full commit are accepted.' >&2
  exit 1
fi
exec sudo -n /opt/yourmove/deploy.sh "${args[@]:1}"
