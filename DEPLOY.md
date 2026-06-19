# Deploying Polygeist

Production install paths for the full agentic stack.

## Recommended — auracrab + stack install

```bash
curl -sSL https://raw.githubusercontent.com/nathfavour/auracrab/main/install.sh | bash
auracrab stack install polygeist
```

Step 1 installs **auracrab** and **vibeaura**. Step 2 installs **polygeist** and **anyisland** as internal libraries (via anyisland official packages — no `github.com/nathfavour/polygeist` URL needed).

This will:

1. Bootstrap **anyisland** if missing
2. `anyisland install polygeist` (recursive submodules: vibeauracle, auracrab, anyisland)
3. Install binaries to **`~/.local/bin`**
4. Configure UDS paths under **`~/.polygeist/run`**
5. Append env to your shell profile (`~/.zshrc` or `~/.profile`)

Then start daemons:

```bash
. ~/.config/polygeist/env
anyisland daemon start
vibeaura daemon start
polygeist --once "smoke test" --workdir .
```

List other installable libraries (go, node, git, docker, …):

```bash
auracrab stack list
auracrab stack install go
```

---

## From source — standalone install.sh

```bash
git clone --recursive https://github.com/nathfavour/polygeist
cd polygeist
chmod +x install.sh scripts/start-daemons.sh
./install.sh
./scripts/start-daemons.sh
```

Same outcome as stack install, but builds from a local monorepo checkout.

---

## Via anyisland directly

If auracrab is already installed:

```bash
anyisland install polygeist
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
