# replay

Interactive CLI to walk through a Git repository commit by commit — forward and back, with an optional diff preview of what comes next.

## Installation

### via go install

```bash
go install github.com/anuchito/replay/cmd/replay@latest
```

> **Version shows `dev` after installing?**
> The Go module proxy caches releases and can lag a few minutes behind a new tag.
> Use an explicit version or bypass the proxy:
>
> ```bash
> # pin to a specific version
> go install github.com/anuchito/replay/cmd/replay@v1.1.3
>
> # or fetch directly from GitHub, skipping the proxy
> GOPROXY=direct go install github.com/anuchito/replay/cmd/replay@latest
> ```

### Build from source

```bash
git clone https://github.com/anuchito/replay.git
cd replay
make build          # builds ./replay, version stamped from git tag
make install-dev    # installs as replay-dev into GOBIN
```

`make install-dev` stamps the version from the current git tag when the tree is clean, or `dev` when there are uncommitted changes.

## Usage

Run inside any Git repository:

```bash
replay                        # pick a starting commit interactively
replay <start>                # replay from a commit to HEAD
replay <start> <end>          # replay a specific range
replay --version              # print version
replay --help                 # print help
```

## Controls

### Commit picker

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Ctrl+D` | Half page down |
| `Ctrl+U` | Half page up |
| `Enter` | Select commit |
| `q` | Quit |

### Replay mode

| Key | Action |
|-----|--------|
| `n` | Next commit |
| `p` | Previous commit |
| `d` | Toggle next-commit diff preview on/off |
| `q` / `Ctrl+C` | Quit and restore original branch |

### Diff preview (when `d` is on)

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down one line |
| `k` / `↑` | Scroll up one line |
| `Ctrl+D` | Half page down |
| `Ctrl+U` | Half page up |
| `Space` | Full page down |

## Notes

- Requires a clean working tree to start (no uncommitted changes)
- Original branch or HEAD is always restored on exit, even on Ctrl+C or error
- Diff preview shows the changes the **next** commit will introduce, before you apply it
