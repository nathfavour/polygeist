# Polygeist

Sovereign multi-agent control plane. Polygeist listens for tasks, runs them through a three-phase loop, and ships results back to your team.

```
Band.ai room ──► polygeist ──► vibeauracle ──► auracrab ──► anyisland
                     │              │              │            │
                     │           mutate         verify       publish
                     └──────────── orchestrates all three via interfaces
```

## The three components

Polygeist does not embed these tools. It orchestrates them as independent agents through abstract interfaces. Each lives in its own repository and is tracked here as a Git submodule.

### [vibeauracle](https://github.com/nathfavour/vibeauracle)

A CLI agentic harness — think Claude Code for your terminal. It reads your codebase, applies structural mutations, runs tools, and drives the engineering loop directly from the shell.

**Role in polygeist:** Phase 1 — mutation. Receives a task payload and changes the target workspace.

### [anyisland](https://github.com/nathfavour/anyisland)

An agentic package manager. Agents use it to install tools, share binaries, and communicate resources across the fleet without manual setup.

**Role in polygeist:** Phase 3 — distribution. Signs release hashes, updates manifests, and publishes immutable artifacts after verification passes.

### [auracrab](https://github.com/nathfavour/auracrab)

Like OpenClaw — a harness that connects to the outside world. It leverages vibeauracle under the hood and handles integration with Telegram, Slack, Discord, and engineering teams. The butler that keeps agents reachable where humans already work.

**Role in polygeist:** Phase 2 — verification. Runs the isolated compiler matrix and test runner inside a sandbox before anything ships.

---

## Quick start

### Option A — install everything with anyisland (recommended)

```bash
anyisland install github.com/nathfavour/polygeist
```

With `track_submodules` enabled, anyisland pulls all three components and installs `polygeist`, `vibeauracle`, `auracrab`, and `anyisland` into your island bin tree.

### Option B — build from source

```bash
git clone --recursive https://github.com/nathfavour/polygeist
cd polygeist
go build -o polygeist ./cmd/polygeist
```

---

## Run

### Band.ai control loop

```bash
export BAND_ROOM_ID=your-room
export BAND_TOKEN=your-token
polygeist
```

Polygeist connects to `wss://api.band.ai/v1/ws`, listens on your room, and executes the full loop for every incoming task.

### Local single task

No Band account needed — run one task and exit:

```bash
polygeist --once "fix the auth middleware" --workdir /path/to/your/repo
```

### Flags

| Flag | Env var | Description |
|---|---|---|
| `--room` | `BAND_ROOM_ID` | Band.ai room to subscribe to |
| `--token` | `BAND_TOKEN` | Band.ai API token |
| `--workdir` | — | Workspace directory (default: `.`) |
| `--once` | — | Run a single payload locally and exit |
| `--version` | — | Print version |

---

## Control loop

Every task flows through three phases:

1. **Mutation** — vibeauracle applies changes to the codebase
2. **Verification** — auracrab runs tests in an isolated sandbox
3. **Distribution** — anyisland signs and publishes the release metadata back to the room

If any phase fails, polygeist logs the error to the Band room and stops.

---

## Repository layout

```
polygeist/          ← you are here (orchestrator)
├── vibeauracle/    ← submodule tracker
├── auracrab/       ← submodule tracker
└── anyisland/      ← submodule tracker
```

The sibling repos (`vibeauracle/`, `auracrab/`, `anyisland/` alongside `polygeist/` in a parent directory) are the canonical source. **Never commit code inside the nested submodule folders here** — push changes to the standalone repos, then advance the pointers:

```bash
git submodule update --remote --merge
```

Compilation uses a `go.work` file committed only at the root of this repository.

---

## Requirements

- Go 1.25+
- `git` (submodule sync)
- `vibeaura` on `PATH` (mutation phase)
- `docker` optional (sandbox verification falls back to local execution)
