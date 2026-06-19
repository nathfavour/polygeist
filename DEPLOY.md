# Deploying Polygeist

Production install paths for the full agentic stack.

## Recommended — standalone install.sh

```bash
git clone --recursive https://github.com/nathfavour/polygeist
cd polygeist
chmod +x install.sh scripts/start-daemons.sh
./install.sh
```

This will:

1. `git submodule update --init --recursive`
2. Build **polygeist**, **vibeaura**, **auracrab**, and **anyisland**
3. Install all four to **`~/.local/bin`**
4. Configure UDS paths under **`~/.polygeist/run`**
5. Append env to your shell profile (`~/.zshrc` or `~/.profile`)

Then start daemons:

```bash
./scripts/start-daemons.sh
# or manually:
anyisland daemon start
vibeaura daemon start
polygeist --once "smoke test" --workdir .
```

---

## Via anyisland package manager

Install anyisland first, then:

```bash
anyisland install polygeist
# or
anyisland install github.com/nathfavour/polygeist
```

The polygeist manifest sets `track_submodules` and `recursive_install`, so anyisland will:

- `git clone --recursive`
- Build polygeist
- Build and install submodule binaries (vibeaura, auracrab, anyisland) alongside the primary binary

Install destination defaults to **`~/.local/bin`** when specified in the manifest.

---

## Docker

```bash
cp .env.example .env   # BAND_API_KEY, BAND_AGENT_ID, BAND_CHAT_ID
docker compose up --build -d
```

Shared UDS volume: `/run/agentic` on all services.

---

## UDS communication

| Component   | Socket (standalone default)     | Override env        |
|-------------|----------------------------------|---------------------|
| anyisland   | `$AGENTIC_RUN_DIR/anyisland.sock` | `ANYISLAND_SOCKET`  |
| vibeauracle | `$AGENTIC_RUN_DIR/vibeaura.sock`  | `VIBEAURA_SOCKET`   |
| polygeist   | `$AGENTIC_RUN_DIR/polygeist.sock` | `POLYGEIST_SOCKET`  |

After `./install.sh`, source `~/.config/polygeist/env`.

Health check:

```bash
echo '{"op":"HEALTH"}' | nc -U "$POLYGEIST_SOCKET"
```

---

## Band.ai production run

```bash
. ~/.config/polygeist/env
export BAND_API_KEY=...
export BAND_AGENT_ID=...
export BAND_CHAT_ID=...
polygeist
```

Uses `https://app.band.ai/api/v1/agent` (REST) and Phoenix WebSocket subscriptions.

---

## Requirements

- Go 1.25+
- Git
- `~/.local/bin` on PATH (install.sh configures this)
