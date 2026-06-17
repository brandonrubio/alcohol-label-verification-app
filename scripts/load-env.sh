#!/usr/bin/env bash

# Load backend and frontend environment variables into the current shell.
#
# Recommended:
#   source scripts/load-env.sh
#
# Backend only:
#   source scripts/load-env.sh --backend
#
# Frontend only:
#   source scripts/load-env.sh --frontend

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

load_all_env() {
  source "$repo_root/backend/scripts/load-env.sh"
  source "$repo_root/frontend/scripts/load-env.sh"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  cat >&2 <<'EOF'
This script is meant to be sourced into your current shell.

  source scripts/load-env.sh
EOF
  exit 1
fi

case "${1:-}" in
  --backend)
    source "$repo_root/backend/scripts/load-env.sh"
    ;;
  --frontend)
    source "$repo_root/frontend/scripts/load-env.sh"
    ;;
  "")
    load_all_env
    ;;
  *)
    echo "load-env: unknown option: $1" >&2
    return 1
    ;;
esac
