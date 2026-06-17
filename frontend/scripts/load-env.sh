#!/usr/bin/env bash

# Load frontend environment variables into the current shell.
#
# Recommended:
#   source frontend/scripts/load-env.sh
#
# From the frontend directory:
#   source scripts/load-env.sh
#
# Optional one-liner:
#   eval "$(frontend/scripts/load-env.sh --export)"
#
# Custom file:
#   source frontend/scripts/load-env.sh /path/to/.env.local

resolve_frontend_env_file() {
  local env_file="${1:-}"
  local frontend_dir

  if [[ -n "$env_file" ]]; then
    printf '%s\n' "$env_file"
    return 0
  fi

  frontend_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

  if [[ -f "$frontend_dir/.env" ]]; then
    printf '%s\n' "$frontend_dir/.env"
    return 0
  fi

  if [[ -f "$frontend_dir/.env.local" ]]; then
    printf '%s\n' "$frontend_dir/.env.local"
    return 0
  fi

  return 1
}

load_frontend_env() {
  local env_file
  if ! env_file="$(resolve_frontend_env_file "${1:-}")"; then
    echo "load-env: no frontend env file found" >&2
    echo "Copy frontend/.env.example to frontend/.env or frontend/.env.local first." >&2
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

  echo "load-env: loaded $env_file" >&2
}

print_frontend_env_exports() {
  local env_file
  if ! env_file="$(resolve_frontend_env_file "${1:-}")"; then
    echo "load-env: no frontend env file found" >&2
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
  print_frontend_env_exports "${2:-}"
  exit $?
fi

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  cat >&2 <<'EOF'
This script is meant to be sourced into your current shell.

  source frontend/scripts/load-env.sh

Or:

  eval "$(frontend/scripts/load-env.sh --export)"
EOF
  exit 1
fi

load_frontend_env "$env_file"
