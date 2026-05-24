#!/usr/bin/env bash
# Date: 2026-05-25
# Author: XinYang Li

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="${ROOT_DIR}/logs"
PID_DIR="${LOG_DIR}/pids"
BACKEND_PID_FILE="${PID_DIR}/backend.pid"
FRONTEND_PID_FILE="${PID_DIR}/frontend.pid"
BACKEND_LOG_FILE="${LOG_DIR}/backend.log"
FRONTEND_LOG_FILE="${LOG_DIR}/frontend.log"
CONDA_ENV_NAME="AgentHub"
DEFAULT_BACKEND_PORT="8080"
DEFAULT_FRONTEND_PORT="3000"

mkdir -p "${LOG_DIR}" "${PID_DIR}"

load_root_env() {
  local env_file="${ROOT_DIR}/.env"
  if [[ -f "${env_file}" ]]; then
    set -a
    # shellcheck disable=SC1090
    source "${env_file}"
    set +a
  fi
}

resolve_python_bin() {
  if ! command -v conda >/dev/null 2>&1; then
    echo "conda command not found. Please ensure conda is installed and available in PATH." >&2
    exit 1
  fi

  conda run -n "${CONDA_ENV_NAME}" python -c 'import sys; print(sys.executable)'
}

read_pid() {
  local pid_file="$1"
  if [[ -f "${pid_file}" ]]; then
    cat "${pid_file}"
  fi
}

is_process_running() {
  local pid="$1"
  if [[ -z "${pid}" ]]; then
    return 1
  fi
  kill -0 "${pid}" >/dev/null 2>&1
}

stop_one_service() {
  local service_name="$1"
  local pid_file="$2"
  local pid
  pid="$(read_pid "${pid_file}")"

  if [[ -z "${pid}" ]]; then
    echo "${service_name} is not running."
    return 0
  fi

  if ! is_process_running "${pid}"; then
    rm -f "${pid_file}"
    echo "${service_name} pid file was stale and has been removed."
    return 0
  fi

  kill "${pid}" >/dev/null 2>&1 || true

  for _ in {1..20}; do
    if ! is_process_running "${pid}"; then
      rm -f "${pid_file}"
      echo "${service_name} stopped."
      return 0
    fi
    sleep 0.5
  done

  kill -9 "${pid}" >/dev/null 2>&1 || true
  rm -f "${pid_file}"
  echo "${service_name} force stopped."
}

kill_processes_on_port() {
  local port="$1"
  local service_name="$2"

  local pids
  if command -v lsof >/dev/null 2>&1; then
    pids="$(lsof -ti tcp:"${port}" 2>/dev/null || true)"
  elif command -v ss >/dev/null 2>&1; then
    pids="$(ss -lptn "sport = :${port}" 2>/dev/null | awk -F'pid=' 'NR>1 && NF>1 {split($2, a, ","); print a[1]}' | sort -u)"
  elif command -v fuser >/dev/null 2>&1; then
    pids="$(fuser "${port}/tcp" 2>/dev/null | tr ' ' '\n' | sed '/^$/d' | sort -u)"
  else
    echo "No port inspection command available. Skipping port cleanup for ${service_name}." >&2
    return 0
  fi

  if [[ -z "${pids}" ]]; then
    echo "no process is using port ${port} for ${service_name}."
    return 0
  fi

  echo "killing ${service_name} processes on port ${port}: ${pids}"
  # shellcheck disable=SC2086
  kill ${pids} >/dev/null 2>&1 || true

  sleep 1

  local remaining_pids
  if command -v lsof >/dev/null 2>&1; then
    remaining_pids="$(lsof -ti tcp:"${port}" 2>/dev/null || true)"
  elif command -v ss >/dev/null 2>&1; then
    remaining_pids="$(ss -lptn "sport = :${port}" 2>/dev/null | awk -F'pid=' 'NR>1 && NF>1 {split($2, a, ","); print a[1]}' | sort -u)"
  else
    remaining_pids="$(fuser "${port}/tcp" 2>/dev/null | tr ' ' '\n' | sed '/^$/d' | sort -u)"
  fi
  if [[ -n "${remaining_pids}" ]]; then
    echo "force killing ${service_name} processes on port ${port}: ${remaining_pids}"
    # shellcheck disable=SC2086
    kill -9 ${remaining_pids} >/dev/null 2>&1 || true
  fi
}

get_backend_port() {
  load_root_env
  echo "${AGENTHUB_BACKEND_PORT:-${DEFAULT_BACKEND_PORT}}"
}

get_frontend_port() {
  load_root_env
  if [[ -n "${FRONTEND_PORT:-}" ]]; then
    echo "${FRONTEND_PORT}"
    return 0
  fi

  if [[ -n "${PORT:-}" ]]; then
    echo "${PORT}"
    return 0
  fi

  echo "${DEFAULT_FRONTEND_PORT}"
}

start_backend() {
  local backend_port
  backend_port="$(get_backend_port)"
  kill_processes_on_port "${backend_port}" "backend"

  local existing_pid
  existing_pid="$(read_pid "${BACKEND_PID_FILE}")"
  if is_process_running "${existing_pid}"; then
    echo "backend is already running with pid ${existing_pid}."
    return 0
  fi

  local python_bin
  python_bin="$(resolve_python_bin)"

  load_root_env
  export AGENTHUB_PYTHON_BIN="${python_bin}"
  export AGENTHUB_AGENT_ROOT_DIR="${ROOT_DIR}/agent"
  export AGENTHUB_FRONTEND_ORIGIN="${AGENTHUB_FRONTEND_ORIGIN:-http://192.168.139.155:3000}"
  export AGENTHUB_BACKEND_PORT="${AGENTHUB_BACKEND_PORT:-8080}"

  nohup bash -lc "cd '${ROOT_DIR}/backend' && env GOCACHE=/tmp/agenthub-go-cache go run ./cmd/server" \
    >>"${BACKEND_LOG_FILE}" 2>&1 &

  echo $! >"${BACKEND_PID_FILE}"
  echo "backend started with pid $(cat "${BACKEND_PID_FILE}")."
}

start_frontend() {
  local frontend_port
  frontend_port="$(get_frontend_port)"
  kill_processes_on_port "${frontend_port}" "frontend"

  local existing_pid
  existing_pid="$(read_pid "${FRONTEND_PID_FILE}")"
  if is_process_running "${existing_pid}"; then
    echo "frontend is already running with pid ${existing_pid}."
    return 0
  fi

  load_root_env
  export NEXT_PUBLIC_API_BASE_URL="${NEXT_PUBLIC_API_BASE_URL:-http://192.168.139.155:8080}"
  export HOSTNAME="${HOSTNAME:-0.0.0.0}"
  export PORT="${frontend_port}"

  nohup bash -lc "cd '${ROOT_DIR}/frontend' && npm run dev" \
    >>"${FRONTEND_LOG_FILE}" 2>&1 &

  echo $! >"${FRONTEND_PID_FILE}"
  echo "frontend started with pid $(cat "${FRONTEND_PID_FILE}")."
}

show_status() {
  local backend_pid frontend_pid
  backend_pid="$(read_pid "${BACKEND_PID_FILE}")"
  frontend_pid="$(read_pid "${FRONTEND_PID_FILE}")"

  if is_process_running "${backend_pid}"; then
    echo "backend: running (pid ${backend_pid})"
  else
    echo "backend: stopped"
  fi

  if is_process_running "${frontend_pid}"; then
    echo "frontend: running (pid ${frontend_pid})"
  else
    echo "frontend: stopped"
  fi

  echo "backend log: ${BACKEND_LOG_FILE}"
  echo "frontend log: ${FRONTEND_LOG_FILE}"
}

start_all() {
  start_backend
  start_frontend
  show_status
}

stop_all() {
  stop_one_service "frontend" "${FRONTEND_PID_FILE}"
  stop_one_service "backend" "${BACKEND_PID_FILE}"
}

restart_all() {
  local backend_port frontend_port
  backend_port="$(get_backend_port)"
  frontend_port="$(get_frontend_port)"

  kill_processes_on_port "${frontend_port}" "frontend"
  kill_processes_on_port "${backend_port}" "backend"
  stop_all
  start_all
}

usage() {
  cat <<'EOF'
Usage: ./dev.sh <command>

Commands:
  start
  stop
  restart
  status
EOF
}

main() {
  local command="${1:-}"

  case "${command}" in
    start)
      start_all
      ;;
    stop)
      stop_all
      ;;
    restart)
      restart_all
      ;;
    status)
      show_status
      ;;
    *)
      usage
      exit 1
      ;;
  esac
}

main "$@"
