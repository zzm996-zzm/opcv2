#!/usr/bin/env bash

set -euo pipefail

REMOTE_HOST="${DEPLOY_HOST:-prod}"
REMOTE_USER="${DEPLOY_USER:-}"
REMOTE_PORT="${DEPLOY_PORT:-}"
REMOTE_DIR="${DEPLOY_PATH:-/data/www/opcv2}"
REMOTE_COMMAND="${1:-update}"
shift || true

deploy_push="${DEPLOY_PUSH:-true}"
healthcheck_url="${DEPLOY_HEALTHCHECK_URL:-}"
healthcheck_attempts="${DEPLOY_HEALTHCHECK_ATTEMPTS:-12}"

case "$deploy_push" in
  1|true|TRUE|yes|YES)
    deploy_push=true
    ;;
  0|false|FALSE|no|NO)
    deploy_push=false
    ;;
  *)
    echo "DEPLOY_PUSH must be true or false" >&2
    exit 1
    ;;
esac

if [[ -n "$healthcheck_url" ]] && ! [[ "$healthcheck_attempts" =~ ^[1-9][0-9]*$ ]]; then
  echo "DEPLOY_HEALTHCHECK_ATTEMPTS must be a positive integer" >&2
  exit 1
fi

push_command=false
case "$REMOTE_COMMAND" in
  deploy|up|start|update|pull)
    push_command=true
    ;;
esac

if [[ "$deploy_push" == true && "$push_command" == true ]]; then
  if [[ -n "$(git status --porcelain)" ]]; then
    echo "working tree is not clean; commit changes before deployment" >&2
    exit 1
  fi
  current_branch="$(git symbolic-ref --quiet --short HEAD)" || {
    echo "deployment requires a checked-out branch" >&2
    exit 1
  }
  git push origin "HEAD:refs/heads/${current_branch}"
fi

quoted_args=()
for arg in "$@"; do
  quoted_args+=("$(printf '%q' "$arg")")
done
if ((${#quoted_args[@]} > 0)); then
  remote_args=" ${quoted_args[*]}"
else
  remote_args=""
fi

remote_dir_quoted="$(printf '%q' "$REMOTE_DIR")"
remote_command_quoted="$(printf '%q' "$REMOTE_COMMAND")"
remote_script=$(cat <<EOF
set -euo pipefail
cd ${remote_dir_quoted}
if [ -x ./deploy.sh ]; then
  ./deploy.sh ${remote_command_quoted}${remote_args}
else
  bash scripts/deploy_server.sh update --build
fi
EOF
)

ssh_args=()
if [[ -n "$REMOTE_USER" ]]; then
  ssh_args+=(-l "$REMOTE_USER")
fi
if [[ -n "$REMOTE_PORT" ]]; then
  ssh_args+=(-p "$REMOTE_PORT")
fi

if ((${#ssh_args[@]} > 0)); then
  ssh "${ssh_args[@]}" "$REMOTE_HOST" "bash -lc $(printf '%q' "$remote_script")"
else
  ssh "$REMOTE_HOST" "bash -lc $(printf '%q' "$remote_script")"
fi

if [[ -n "$healthcheck_url" ]]; then
  healthy=false
  for ((attempt = 1; attempt <= healthcheck_attempts; attempt++)); do
    if curl --fail --silent --show-error --max-time 10 "$healthcheck_url" >/dev/null; then
      healthy=true
      echo "health check passed: $healthcheck_url"
      break
    fi
    if ((attempt < healthcheck_attempts)); then
      sleep 5
    fi
  done
  if [[ "$healthy" != true ]]; then
    echo "health check failed after ${healthcheck_attempts} attempts: $healthcheck_url" >&2
    exit 1
  fi
fi
