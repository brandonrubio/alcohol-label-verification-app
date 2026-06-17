#!/usr/bin/env bash

# Load backend environment variables into the current shell.
#
# Recommended:
#   source backend/scripts/load-env.sh
#
# From the backend directory:
#   source scripts/load-env.sh
#
# Optional one-liner (also works in subshells):
#   eval "$(backend/scripts/load-env.sh --export)"
#
# Custom file:
#   source backend/scripts/load-env.sh /path/to/.env

load_backend_env() {
  local env_file="${1:-}"

  if [[ -z "$env_file" ]]; then
    env_file="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/.env"
  fi

  if [[ ! -f "$env_file" ]]; then
    echo "load-env: file not found: $env_file" >&2
    echo "Copy backend/.env.example to backend/.env first." >&2
    return 1
  fi

  local line key value
  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line%$'\r'}"

    [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue

    if [[ ! "$line" =~ ^[[:space:]]*([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
      echo "load-env: skipping invalid line in $env_file" >&2
      continue
    fi

    key="${BASH_REMATCH[1]}"
    value="${BASH_REMATCH[2]}"

    if [[ "$value" =~ ^\".*\"$ ]]; then
      value="${value:1:${#value}-2}"
    elif [[ "$value" =~ ^\'.*\'$ ]]; then
      value="${value:1:${#value}-2}"
    fi

    export "${key}=${value}"
  done <"$env_file"
}

print_backend_env_exports() {
  local env_file="${1:-}"

  if [[ -z "$env_file" ]]; then
    env_file="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/.env"
  fi

  if [[ ! -f "$env_file" ]]; then
    echo "load-env: file not found: $env_file" >&2
    return 1
  fi

  local line key value
  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line%$'\r'}"

    [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue

    if [[ ! "$line" =~ ^[[:space:]]*([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
      continue
    fi

    key="${BASH_REMATCH[1]}"
    value="${BASH_REMATCH[2]}"

    if [[ "$value" =~ ^\".*\"$ ]]; then
      value="${value:1:${#value}-2}"
    elif [[ "$value" =~ ^\'.*\'$ ]]; then
      value="${value:1:${#value}-2}"
    fi

    printf 'export %q=%q\n' "$key" "$value"
  done <"$env_file"
}

env_file="${1:-}"

if [[ "${1:-}" == "--export" ]]; then
  print_backend_env_exports "${2:-}"
  exit $?
fi

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  cat >&2 <<'EOF'
This script is meant to be sourced into your current shell.

  source backend/scripts/load-env.sh

Or:

  eval "$(backend/scripts/load-env.sh --export)"
EOF
  exit 1
fi

load_backend_env "$env_file"
