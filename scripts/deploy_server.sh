#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-${ROOT_DIR}/docker-compose.server.yml}"
ENV_FILE="${ENV_FILE:-${ROOT_DIR}/.env.server}"
ENV_EXAMPLE_FILE="${ENV_EXAMPLE_FILE:-${ROOT_DIR}/.env.server.example}"

if [[ ! -f "$ENV_FILE" ]]; then
  if [[ -f "${ROOT_DIR}/.env" ]]; then
    ENV_FILE="${ROOT_DIR}/.env"
    echo "using existing env file: $ENV_FILE"
  elif [[ -f "$ENV_EXAMPLE_FILE" ]]; then
    cp "$ENV_EXAMPLE_FILE" "${ROOT_DIR}/.env.server"
    ENV_FILE="${ROOT_DIR}/.env.server"
    echo "bootstrapped env file from example: $ENV_FILE"
  else
    echo "missing env file: $ENV_FILE" >&2
    exit 1
  fi
fi

compose() {
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

usage() {
  cat <<'EOF'
Usage:
  bash scripts/deploy_server.sh deploy [--build]
  bash scripts/deploy_server.sh update [--build]
  bash scripts/deploy_server.sh down
  bash scripts/deploy_server.sh restart [service]
  bash scripts/deploy_server.sh logs [service]
  bash scripts/deploy_server.sh ps
  bash scripts/deploy_server.sh config

Commands:
  deploy   Start all server services, optionally rebuilding images
  update   Pull latest code, then restart services, optionally rebuilding images
  down     Stop and remove all server services
  restart  Restart all services or one named service
  logs     Follow logs for all services or one named service
  ps       Show current service status
  config   Render final docker compose config
EOF
}

cmd="${1:-deploy}"
shift || true

build_requested=false
service=""

for arg in "$@"; do
  case "$arg" in
    --build)
      build_requested=true
      ;;
    *)
      if [[ -z "$service" ]]; then
        service="$arg"
      else
        echo "unexpected argument: $arg" >&2
        usage
        exit 1
      fi
      ;;
  esac
done

up_args=(up -d --remove-orphans)
if [[ "$build_requested" == true ]]; then
  up_args=(up -d --build --remove-orphans)
fi

case "$cmd" in
  deploy|up)
    compose "${up_args[@]}"
    ;;
  update)
    git -C "$ROOT_DIR" pull --ff-only
    compose "${up_args[@]}"
    ;;
  down)
    compose down
    ;;
  restart)
    if [[ -n "$service" ]]; then
      compose restart "$service"
    else
      compose restart
    fi
    ;;
  logs)
    if [[ -n "$service" ]]; then
      compose logs -f --tail=200 "$service"
    else
      compose logs -f --tail=200
    fi
    ;;
  ps|status)
    compose ps
    ;;
  config)
    compose config
    ;;
  *)
    usage
    exit 1
    ;;
esac
