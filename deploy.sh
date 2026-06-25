#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_COMMAND="${1:-update}"
shift || true

case "$DEPLOY_COMMAND" in
  deploy|up|start)
    exec bash "$ROOT_DIR/scripts/deploy_server.sh" deploy --build "$@"
    ;;
  update|pull)
    exec bash "$ROOT_DIR/scripts/deploy_server.sh" update --build "$@"
    ;;
  restart|down|logs|ps|status|config)
    exec bash "$ROOT_DIR/scripts/deploy_server.sh" "$DEPLOY_COMMAND" "$@"
    ;;
  *)
    cat <<'EOF' >&2
Usage:
  ./deploy.sh [update|deploy|restart|down|logs|ps|config] [service]

Default:
  ./deploy.sh

The default command pulls latest code and rebuilds/restarts server Docker Compose services.
EOF
    exit 1
    ;;
esac
