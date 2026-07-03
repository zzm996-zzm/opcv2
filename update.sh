#!/usr/bin/env bash

set -euo pipefail

REMOTE_HOST="${DEPLOY_HOST:-prod}"
REMOTE_DIR="${DEPLOY_PATH:-/data/www/opcv2}"
REMOTE_COMMAND="${1:-update}"
shift || true

quoted_args=()
for arg in "$@"; do
  quoted_args+=("$(printf '%q' "$arg")")
done
if ((${#quoted_args[@]} > 0)); then
  remote_args=" ${quoted_args[*]}"
else
  remote_args=""
fi

remote_script=$(cat <<EOF
set -euo pipefail
cd ${REMOTE_DIR}
if [ -x ./deploy.sh ]; then
  ./deploy.sh $(printf '%q' "$REMOTE_COMMAND")${remote_args}
else
  bash scripts/deploy_server.sh update --build
fi
EOF
)

ssh "$REMOTE_HOST" "bash -lc $(printf '%q' "$remote_script")"
