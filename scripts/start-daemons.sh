#!/usr/bin/env bash
# Start the agentic daemon stack (anyisland + vibeauracle + optional polygeist)
set -euo pipefail

CONFIG="${HOME}/.config/polygeist/env"
if [ -f "${CONFIG}" ]; then
  # shellcheck disable=SC1090
  . "${CONFIG}"
fi

RUN_DIR="${AGENTIC_RUN_DIR:-${HOME}/.polygeist/run}"
mkdir -p "${RUN_DIR}"
PID_DIR="${HOME}/.polygeist/pids"
mkdir -p "${PID_DIR}"

start_bg() {
  local name=$1
  shift
  if [ -f "${PID_DIR}/${name}.pid" ] && kill -0 "$(cat "${PID_DIR}/${name}.pid")" 2>/dev/null; then
    echo "${name} already running (pid $(cat "${PID_DIR}/${name}.pid"))"
    return
  fi
  echo "starting ${name}..."
  "$@" &
  echo $! > "${PID_DIR}/${name}.pid"
}

start_bg anyisland anyisland daemon start
sleep 1
start_bg vibeaura vibeaura daemon start
sleep 1

if [ "${START_POLYGEIST:-0}" = "1" ]; then
  start_bg polygeist polygeist
fi

echo "daemons up. UDS dir: ${RUN_DIR}"
echo "health: echo '{\"op\":\"HEALTH\"}' | nc -U ${POLYGEIST_SOCKET:-${RUN_DIR}/polygeist.sock}"
