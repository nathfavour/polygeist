# Polygeist

Sovereign multi-agent control plane. Polygeist listens for tasks, runs them through a three-phase loop, and ships results back to your team.

```
Band.ai room ──► polygeist ──► vibeauracle ──► auracrab ──► anyisland
                     │              │              │            │
                     │           mutate         verify       publish
                     └──────────── orchestrates all three via UDS
```

## The three components

### [vibeauracle](https://github.com/nathfavour/vibeauracle)

CLI agentic harness (Claude Code–style). Phase 1 — mutation via UDS (`vibeaura.sock`).

### [anyisland](https://github.com/nathfavour/anyisland)

Agentic package manager. Phase 3 — distribution via UDS (`anyisland.sock`).

### [auracrab](https://github.com/nathfavour/auracrab)

OpenClaw-style team bridge (Telegram, Slack, Discord). Phase 2 — sandbox verification.

---

## Install (production)

```bash
git clone --recursive https://github.com/nathfavour/polygeist
cd polygeist
./install.sh
./scripts/start-daemons.sh
```

Installs **polygeist**, **vibeaura**, **auracrab**, and **anyisland** to `~/.local/bin`.

Or via anyisland:

```bash
anyisland install polygeist
```

Full details: **[DEPLOY.md](DEPLOY.md)**

---

## Run

```bash
. ~/.config/polygeist/env
anyisland daemon start
vibeaura daemon start

# Band.ai
export BAND_API_KEY=... BAND_AGENT_ID=... BAND_CHAT_ID=...
polygeist

# Local task
polygeist --once "fix auth middleware" --workdir /path/to/repo
```

---

## Submodule workflow

Push changes to standalone repos (`vibeauracle`, `auracrab`, `anyisland`), then:

```bash
git submodule update --remote --merge
```
