#!/usr/bin/env bash
# Polygeist stack installer — recursive clone, build, and install to ~/.local/bin
set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
RUN_DIR="${AGENTIC_RUN_DIR:-${HOME}/.polygeist/run}"
CONFIG_DIR="${HOME}/.config/polygeist"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

log() { printf '==> %s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "missing dependency: $1"
}

need git
need go

mkdir -p "${INSTALL_DIR}" "${RUN_DIR}" "${CONFIG_DIR}"

if [ -d "${REPO_ROOT}/.git" ]; then
  log "Syncing git submodules (recursive)"
  git -C "${REPO_ROOT}" submodule update --init --recursive
else
  die "run install.sh from a polygeist git checkout (git clone --recursive https://github.com/nathfavour/polygeist)"
fi

for dir in anyisland vibeauracle auracrab; do
  [ -d "${REPO_ROOT}/${dir}" ] || die "missing submodule: ${dir} (use git clone --recursive)"
done

export AGENTIC_RUN_DIR="${RUN_DIR}"
export ANYISLAND_SOCKET="${RUN_DIR}/anyisland.sock"
export VIBEAURA_SOCKET="${RUN_DIR}/vibeaura.sock"
export POLYGEIST_SOCKET="${RUN_DIR}/polygeist.sock"
export ANYISLAND_BIN_DIR="${INSTALL_DIR}"

cat > "${CONFIG_DIR}/env" <<EOF
# Source this file: . "${CONFIG_DIR}/env"
export AGENTIC_RUN_DIR="${RUN_DIR}"
export ANYISLAND_SOCKET="${ANYISLAND_SOCKET}"
export VIBEAURA_SOCKET="${VIBEAURA_SOCKET}"
export POLYGEIST_SOCKET="${POLYGEIST_SOCKET}"
export ANYISLAND_BIN_DIR="${INSTALL_DIR}"
export PATH="${INSTALL_DIR}:\${PATH}"
EOF

install_bin() {
  local src="$1" name="$2"
  install -m 0755 "${src}" "${INSTALL_DIR}/${name}"
  log "installed ${name} -> ${INSTALL_DIR}/${name}"
}

log "Building anyisland"
( cd "${REPO_ROOT}/anyisland" && CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/anyisland-build ./cmd/anyisland )
install_bin /tmp/anyisland-build anyisland

log "Building vibeaura"
( cd "${REPO_ROOT}/vibeauracle/cmd/vibeaura" && CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/vibeaura-build . )
install_bin /tmp/vibeaura-build vibeaura

log "Building auracrab"
( cd "${REPO_ROOT}/auracrab" && CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/auracrab-build ./cmd/auracrab )
install_bin /tmp/auracrab-build auracrab

log "Building polygeist"
( cd "${REPO_ROOT}" && CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/polygeist-build ./cmd/polygeist )
install_bin /tmp/polygeist-build polygeist

rm -f /tmp/anyisland-build /tmp/vibeaura-build /tmp/auracrab-build /tmp/polygeist-build

MARKER="# polygeist agentic stack"
PROFILE="${HOME}/.profile"
if [ -n "${ZSH_VERSION:-}" ] || [ -f "${HOME}/.zshrc" ]; then
  PROFILE="${HOME}/.zshrc"
fi

if ! grep -q "${MARKER}" "${PROFILE}" 2>/dev/null; then
  log "Adding PATH and UDS env to ${PROFILE}"
  cat >> "${PROFILE}" <<EOF

${MARKER}
. "${CONFIG_DIR}/env"
EOF
fi

log "Done."
cat <<EOF

Installed to ${INSTALL_DIR}:
  polygeist, vibeaura, auracrab, anyisland

UDS runtime directory: ${RUN_DIR}

Start daemons (three terminals or use scripts/start-daemons.sh):
  anyisland daemon start
  vibeaura daemon start
  polygeist --once "hello" --workdir .

Or via anyisland after installing anyisland first:
  anyisland install github.com/nathfavour/polygeist

Restart your shell or run: . "${CONFIG_DIR}/env"
EOF
