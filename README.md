# Polygeist

Sovereign multi-agent control plane and orchestration daemon.

Polygeist coordinates three decoupled agents through abstract interfaces:

- **vibeauracle** — structural codebase mutation
- **auracrab** — isolated sandbox verification
- **anyisland** — signed binary distribution

## Build

```bash
go build -o polygeist ./cmd/polygeist
```

## Run

```bash
# Band.ai control loop
export BAND_ROOM_ID=your-room
export BAND_TOKEN=your-token
./polygeist

# Local single-task mode
./polygeist --once "implement feature X" --workdir /path/to/repo
```

## Submodule workflow

Code changes belong in the standalone sibling repositories. Polygeist only tracks submodule SHAs:

```bash
git submodule update --remote --merge
```
