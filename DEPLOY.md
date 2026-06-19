# Deploying Polygeist

Polygeist orchestrates **vibeauracle**, **auracrab**, and **anyisland** over Unix domain sockets (UDS) and connects to [Band.ai](https://docs.band.ai) for remote task dispatch.

## Architecture

```text
Band.ai (REST + Phoenix WebSocket)
        │
        ▼
   polygeist.sock ── polygeist (orchestrator)
        │                │
        │                ├── mutation  → vibeaura.sock (vibeauracle)
        │                ├── verify    → auracrab sandbox (local/docker)
        │                └── publish   → anyisland.sock (anyisland)
        │
   /run/agentic/  (shared UDS directory in Docker)
```

### UDS paths (default)

| Component   | Socket                                      | Env override        |
|-------------|---------------------------------------------|---------------------|
| anyisland   | `~/.anyisland/anyisland.sock`               | `ANYISLAND_SOCKET`  |
| vibeauracle | `~/.vibeauracle/vibeaura.sock`              | `VIBEAURA_SOCKET`   |
| polygeist   | `~/.polygeist/run/polygeist.sock`           | `POLYGEIST_SOCKET`  |

Set `AGENTIC_RUN_DIR=/run/agentic` to colocate all sockets (used by Docker).

Check polygeist health:

```bash
echo '{"op":"HEALTH"}' | nc -U ~/.polygeist/run/polygeist.sock
```

---

## Option 1 — Docker (fastest)

From the `polygeist` directory:

```bash
cp .env.example .env   # fill in Band credentials
docker compose up --build -d
```

Required env vars in `.env`:

```bash
BAND_API_KEY=your-agent-api-key
BAND_AGENT_ID=your-agent-uuid
BAND_CHAT_ID=your-chat-room-uuid
```

Services started:

| Service     | Role                                      |
|-------------|-------------------------------------------|
| anyisland   | Package manager daemon + UDS broker       |
| vibeauracle | Mutation engine daemon + UDS            |
| polygeist   | Band control loop + orchestrator UDS    |

Mount your repo into the stack:

```bash
docker compose run --rm -v "$PWD/../myproject:/workspace" polygeist \
  --once "fix the login bug" --workdir /workspace
```

---

## Option 2 — Native install via anyisland

```bash
anyisland install github.com/nathfavour/polygeist
```

Then start daemons:

```bash
anyisland daemon start
vibeaura daemon start
polygeist --chat "$BAND_CHAT_ID" --api-key "$BAND_API_KEY" --agent-id "$BAND_AGENT_ID"
```

---

## Option 3 — Build from source

```bash
# sibling repos
git clone --recursive https://github.com/nathfavour/polygeist
cd polygeist
go build -o polygeist ./cmd/polygeist

# start IPC daemons (separate terminals)
anyisland daemon start
vibeaura daemon start

# run
export BAND_API_KEY=...
export BAND_AGENT_ID=...
export BAND_CHAT_ID=...
./polygeist
```

---

## Band.ai setup

1. Register a remote agent in Band (Human API or dashboard).
2. Copy the **agent API key** and **agent ID**.
3. Add the agent to a chat room; copy the **chat room UUID**.
4. Export credentials (see above).

Polygeist uses:

- **REST** `https://app.band.ai/api/v1/agent` — send events, mark messages processed
- **WebSocket** `wss://app.band.ai/api/v1/socket/websocket` — Phoenix Channels, `message_created` events

On startup polygeist drains `/messages/next` for crash recovery, then listens on WebSocket with 30s heartbeats.

---

## Local task (no Band)

```bash
polygeist --once "add retry logic to the API client" --workdir /path/to/repo
```

Requires `vibeaura daemon start` for UDS mutation (falls back to CLI if UDS is down).

---

## auracrab (team integrations)

Run on the host for Telegram/Slack/Discord bridges:

```bash
auracrab start
```

auracrab talks to vibeauracle over `vibeaura.sock` and anyisland over `anyisland.sock`. It is optional for the core polygeist loop but required for engineering-team messaging integrations.

---

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `vibeauracle uds: connection refused` | Run `vibeaura daemon start` |
| `band websocket dial` failed | Check `BAND_API_KEY`, network, `app.band.ai` reachability |
| `chat ID required` | Set `BAND_CHAT_ID` or `--chat` |
| Sockets not shared in Docker | Ensure `agentic-run` volume is mounted on all services |
