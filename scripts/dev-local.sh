#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_DIR="${TMPDIR:-/tmp}/calandar-dev"
APP_LOG_DIR="$REPO_ROOT/logs"
SUPABASE_LOG="$LOG_DIR/supabase.log"
API_LOG="$LOG_DIR/api.log"
WEB_LOG="$LOG_DIR/web.log"
DEFAULT_BACKEND_PORT="${BACKEND_PORT:-8080}"
DEFAULT_WEB_PORT="${WEB_PORT:-5173}"
BACKEND_PORT="$DEFAULT_BACKEND_PORT"
WEB_PORT="$DEFAULT_WEB_PORT"
BACKEND_URL=""
WEB_URL=""

API_PID=""
WEB_PID=""

usage() {
  cat <<USAGE
Usage: scripts/dev-local.sh

Start the local Hangout development stack:
  1. Supabase local services
  2. Gin API on 127.0.0.1:\$BACKEND_PORT (defaults to 8080, auto-falls back)
  3. Vite web app on 127.0.0.1:\$WEB_PORT (defaults to 5173, auto-falls back)
  4. Health checks for API direct access and Vite /api proxy

Logs:
  $APP_LOG_DIR
  $SUPABASE_LOG
  $API_LOG
  $WEB_LOG

Press Ctrl+C to stop the API and web processes started by this script.
Supabase is intentionally left running.
USAGE
}

log() {
  printf '[dev-local] %s\n' "$*"
}

die() {
  printf '[dev-local] ERROR: %s\n' "$*" >&2
  exit 1
}

cleanup() {
  local status=$?

  if [[ -n "${WEB_PID:-}" ]] && kill -0 "$WEB_PID" 2>/dev/null; then
    log "Stopping web process $WEB_PID"
    kill "$WEB_PID" 2>/dev/null || true
    wait "$WEB_PID" 2>/dev/null || true
  fi

  if [[ -n "${API_PID:-}" ]] && kill -0 "$API_PID" 2>/dev/null; then
    log "Stopping API process $API_PID"
    kill "$API_PID" 2>/dev/null || true
    wait "$API_PID" 2>/dev/null || true
  fi

  exit "$status"
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "Missing required command: $1"
}

port_is_in_use() {
  local port="$1"
  if command -v lsof >/dev/null 2>&1; then
    lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
    return
  fi

  die "Missing required command: lsof"
}

pick_free_port() {
  local requested_port="$1"
  local port="$requested_port"
  local max_port="${2:-65535}"

  while (( port <= max_port )); do
    if ! port_is_in_use "$port"; then
      printf '%s' "$port"
      return 0
    fi
    ((port++))
  done

  die "Could not find a free port starting from $requested_port"
}

configure_ports() {
  BACKEND_PORT="$(pick_free_port "$DEFAULT_BACKEND_PORT")"
  WEB_PORT="$(pick_free_port "$DEFAULT_WEB_PORT")"
  BACKEND_URL="http://127.0.0.1:$BACKEND_PORT"
  WEB_URL="http://127.0.0.1:$WEB_PORT"

  if [[ "$BACKEND_PORT" != "$DEFAULT_BACKEND_PORT" ]]; then
    log "Port $DEFAULT_BACKEND_PORT is busy, API will use $BACKEND_PORT"
  fi

  if [[ "$WEB_PORT" != "$DEFAULT_WEB_PORT" ]]; then
    log "Port $DEFAULT_WEB_PORT is busy, web will use $WEB_PORT"
  fi
}

load_supabase_env() {
  local status_output
  local key
  local value
  status_output="$(supabase status -o env 2>>"$SUPABASE_LOG")"

  while IFS= read -r line; do
    [[ "$line" =~ ^[A-Z_][A-Z0-9_]*= ]] || continue
    key="${line%%=*}"
    value="${line#*=}"
    value="${value%\"}"
    value="${value#\"}"
    export "$key=$value"
  done <<< "$status_output"

  [[ -n "${API_URL:-}" ]] || die "Supabase API_URL was not found in supabase status output."
  [[ -n "${DB_URL:-}" ]] || die "Supabase DB_URL was not found in supabase status output."
  [[ -n "${JWT_SECRET:-}" ]] || die "Supabase JWT_SECRET was not found in supabase status output."
  [[ -n "${ANON_KEY:-}" ]] || die "Supabase ANON_KEY was not found in supabase status output."
}

with_sslmode_disabled() {
  local db_url="$1"
  if [[ "$db_url" == *\?* ]]; then
    printf '%s&sslmode=disable' "$db_url"
  else
    printf '%s?sslmode=disable' "$db_url"
  fi
}

wait_for_http() {
  local name="$1"
  local url="$2"
  local timeout_seconds="${3:-60}"
  local start
  local status
  local response_file

  start="$(date +%s)"
  response_file="$LOG_DIR/${name// /-}.response.json"

  while true; do
    status="$(curl -sS -o "$response_file" -w '%{http_code}' "$url" 2>/dev/null || true)"
    if [[ "$status" == "200" ]]; then
      log "$name is reachable: $url"
      return 0
    fi

    if [[ -n "${API_PID:-}" ]] && ! kill -0 "$API_PID" 2>/dev/null; then
      tail -n 80 "$API_LOG" >&2 || true
      die "API process exited before $name became reachable."
    fi

    if [[ -n "${WEB_PID:-}" ]] && ! kill -0 "$WEB_PID" 2>/dev/null; then
      tail -n 80 "$WEB_LOG" >&2 || true
      die "Web process exited before $name became reachable."
    fi

    if (( "$(date +%s)" - start >= timeout_seconds )); then
      log "Last HTTP status for $name: ${status:-none}"
      [[ -f "$response_file" ]] && cat "$response_file" >&2 || true
      die "Timed out waiting for $name at $url"
    fi

    sleep 1
  done
}

start_api() {
  local database_url
  local jwks_url
  database_url="$(with_sslmode_disabled "$DB_URL")"
  jwks_url="$API_URL/auth/v1/.well-known/jwks.json"

  log "Starting API on $BACKEND_URL"
  (
    cd "$REPO_ROOT"
    API_ADDR="127.0.0.1:$BACKEND_PORT" \
      DATABASE_URL="$database_url" \
      LOG_DIR="$APP_LOG_DIR" \
      LOG_LEVEL="${LOG_LEVEL:-DEBUG}" \
      LOG_TO_STDERR="${LOG_TO_STDERR:-false}" \
      LOG_ALSO_TO_STDERR="${LOG_ALSO_TO_STDERR:-true}" \
      SUPABASE_JWT_SECRET="$JWT_SECRET" \
      SUPABASE_JWKS_URL="$jwks_url" \
      go run ./apps/api/cmd/api
  ) >"$API_LOG" 2>&1 &
  API_PID=$!
}

start_web() {
  log "Starting web app on $WEB_URL"
  (
    cd "$REPO_ROOT/apps/web"
    VITE_SUPABASE_URL="$API_URL" \
      VITE_SUPABASE_ANON_KEY="$ANON_KEY" \
      VITE_API_PROXY_TARGET="$BACKEND_URL" \
      npm run dev -- --host 127.0.0.1 --port "$WEB_PORT" --strictPort
  ) >"$WEB_LOG" 2>&1 &
  WEB_PID=$!
}

monitor_processes() {
  log "Development stack is running."
  log "Web: $WEB_URL"
  log "API health: $BACKEND_URL/api/health"
  log "Glog files: $APP_LOG_DIR"
  log "Process logs: $LOG_DIR"
  log "Press Ctrl+C to stop API and web. Supabase will remain running."

  while true; do
    if ! kill -0 "$API_PID" 2>/dev/null; then
      tail -n 80 "$API_LOG" >&2 || true
      die "API process exited."
    fi

    if ! kill -0 "$WEB_PID" 2>/dev/null; then
      tail -n 80 "$WEB_LOG" >&2 || true
      die "Web process exited."
    fi

    sleep 2
  done
}

main() {
  if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    usage
    exit 0
  fi

  require_command supabase
  require_command go
  require_command npm
  require_command curl
  require_command lsof

  mkdir -p "$LOG_DIR"
  : >"$SUPABASE_LOG"
  : >"$API_LOG"
  : >"$WEB_LOG"

  trap 'exit 130' INT TERM
  trap cleanup EXIT

  configure_ports

  log "Starting Supabase local services"
  if ! (cd "$REPO_ROOT" && supabase start >"$SUPABASE_LOG" 2>&1); then
    tail -n 80 "$SUPABASE_LOG" >&2 || true
    die "Supabase failed to start."
  fi

  load_supabase_env
  start_api
  wait_for_http "API health" "$BACKEND_URL/api/health" 60

  start_web
  wait_for_http "Vite proxy health" "$WEB_URL/api/health" 60

  monitor_processes
}

main "$@"
