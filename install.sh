#!/usr/bin/env bash
# Polygeist installer — bootstraps anyisland, then installs the full stack.
#
# Recommended (single command, no git clone):
#   curl -sSL https://raw.githubusercontent.com/nathfavour/polygeist/main/install.sh | bash
#
# From a local monorepo checkout:
#   ./install.sh
set -euo pipefail

ANYISLAND_INSTALL_URL="https://raw.githubusercontent.com/nathfavour/anyisland/main/install.sh"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
RUN_DIR="${AGENTIC_RUN_DIR:-${HOME}/.polygeist/run}"
CONFIG_DIR="${HOME}/.config/polygeist"
ENV_FILE="${CONFIG_DIR}/env"
MARKER="# polygeist agentic stack"

log() { printf '==> %s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

ensure_path() {
  case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *) export PATH="${INSTALL_DIR}:${PATH}" ;;
  esac
}

ensure_anyisland() {
  ensure_path
  if command -v anyisland >/dev/null 2>&1; then
    return 0
  fi
  log "anyisland not found — bootstrapping package manager"
  curl -fsSL "${ANYISLAND_INSTALL_URL}" | bash
  ensure_path
  command -v anyisland >/dev/null 2>&1 || die "anyisland install failed (is ${INSTALL_DIR} on PATH?)"
}

configure_polygeist_runtime() {
  mkdir -p "${INSTALL_DIR}" "${RUN_DIR}" "${CONFIG_DIR}"

  cat > "${ENV_FILE}" <<EOF
# Source this file: . "${ENV_FILE}"
export AGENTIC_RUN_DIR="${RUN_DIR}"
export ANYISLAND_SOCKET="${RUN_DIR}/anyisland.sock"
export VIBEAURA_SOCKET="${RUN_DIR}/vibeaura.sock"
export POLYGEIST_SOCKET="${RUN_DIR}/polygeist.sock"
export ANYISLAND_BIN_DIR="${INSTALL_DIR}"
export PATH="${INSTALL_DIR}:\${PATH}"
EOF

  profile="${HOME}/.profile"
  if [ -n "${ZSH_VERSION:-}" ] || [ -f "${HOME}/.zshrc" ]; then
    profile="${HOME}/.zshrc"
  fi

  if ! grep -q "${MARKER}" "${profile}" 2>/dev/null; then
    log "Adding PATH and UDS env to ${profile}"
    cat >> "${profile}" <<EOF

${MARKER}
. "${ENV_FILE}"
EOF
  fi
}

install_via_anyisland() {
  ensure_anyisland
  log "Installing polygeist via anyisland"
  anyisland install polygeist
  configure_polygeist_runtime

  log "Done."
  cat <<EOF

Installed via anyisland to ${INSTALL_DIR}:
  polygeist, vibeaura, auracrab, anyisland

UDS runtime directory: ${RUN_DIR}

Start daemons:
  . "${ENV_FILE}"
  anyisland daemon start
  vibeaura daemon start
  polygeist --once "hello" --workdir .

Restart your shell or run: . "${ENV_FILE}"
EOF
}

install_from_source() {
  local repo_root="$1"

  need() {
    command -v "$1" >/dev/null 2>&1 || die "missing dependency: $1"
  }

  need git
  need go

  mkdir -p "${INSTALL_DIR}" "${RUN_DIR}" "${CONFIG_DIR}"

  log "Syncing git submodules (recursive)"
  git -C "${repo_root}" submodule update --init --recursive

  for dir in anyisland vibeauracle auracrab; do
    [ -d "${repo_root}/${dir}" ] || die "missing submodule: ${dir} (use git clone --recursive)"
  done

  configure_polygeist_runtime

  install_bin() {
    local src="$1" name="$2"
    install -m 0755 "${src}" "${INSTALL_DIR}/${name}"
    log "installed ${name} -> ${INSTALL_DIR}/${name}"
  }

  log "Building anyisland"
  ( cd "${repo_root}/anyisland" && CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/anyisland-build ./cmd/anyisland )
  install_bin /tmp/anyisland-build anyisland

  log "Building vibeaura"
  ( cd "${repo_root}/vibeauracle/cmd/vibeaura" && CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/vibeaura-build . )
  install_bin /tmp/vibeaura-build vibeaura

  log "Building auracrab"
  ( cd "${repo_root}/auracrab" && CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/auracrab-build ./cmd/auracrab )
  install_bin /tmp/auracrab-build auracrab

  log "Building polygeist"
  ( cd "${repo_root}" && CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/polygeist-build ./cmd/polygeist )
  install_bin /tmp/polygeist-build polygeist

  rm -f /tmp/anyisland-build /tmp/vibeaura-build /tmp/auracrab-build /tmp/polygeist-build

  log "Done."
  cat <<EOF

Installed to ${INSTALL_DIR}:
  polygeist, vibeaura, auracrab, anyisland

UDS runtime directory: ${RUN_DIR}

Start daemons (or use scripts/start-daemons.sh):
  . "${ENV_FILE}"
  anyisland daemon start
  vibeaura daemon start
  polygeist --once "hello" --workdir .

Restart your shell or run: . "${ENV_FILE}"
EOF
}

repo_root=""
if [ -n "${BASH_SOURCE[0]:-}" ] && [ -f "${BASH_SOURCE[0]}" ]; then
  repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
fi

if [ -n "${repo_root}" ] && [ -d "${repo_root}/.git" ] && [ -d "${repo_root}/anyisland" ]; then
  install_from_source "${repo_root}"
else
  install_via_anyisland
fi
