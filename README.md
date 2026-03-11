# sniffa

A terminal UI for watching and running Go tests. Sniffa monitors your project for file changes and reruns tests
automatically, with a sidebar for navigating test files and a live output panel.

![screenshot](https://github.com/user-attachments/assets/cee817ae-301d-47c2-bf5d-f647f30f11e0)

## Features

- Watches `.go` files for changes and reruns affected tests immediately
- Sidebar lists all `_test.go` files found in the project
- Navigate files with arrow keys or `j`/`k` — the output panel shows only that file's results
- Toggle files on/off with `space` to silence noisy or slow tests
- Runs tests per file using `go test -v -run` with extracted test function names, not the whole package
- Pass/fail state is reflected in color: green for passing, red for failing, blue for pending, gray for disabled
- Respects `maxDepth` to avoid descending into deeply nested directories

## Installation

```bash
go install github.com/jwc20/sniffa/cmd/sniffa@latest
```

Or build from source:

```bash
git clone https://github.com/jwc20/sniffa
cd sniffa
task build
```

The binary is placed in `./bin/sniffa`.

## Keybindings

| Key            | Action                      |
|----------------|-----------------------------|
| `↑` / `k`      | Move cursor up              |
| `↓` / `j`      | Move cursor down            |
| `space`        | Toggle selected file on/off |
| `q` / `ctrl+c` | Quit                        |

## How it works

On startup, sniffa:

1. Walks the given directories (up to `maxDepth = 10`) and finds all `_test.go` files
2. Parses each file to extract `TestXxx` function names
3. Runs all enabled tests immediately in parallel
4. Starts a filesystem watcher; when a `.go` file changes, only the tests in that package rerun

Tests are run as `go test -v -run ^(TestFoo|TestBar)$` scoped to the functions in the changed file rather than the
entire package.

## Dependencies

- [charm.land/bubbletea/v2](https://github.com/charmbracelet/bubbletea) — TUI framework
- [charm.land/lipgloss/v2](https://github.com/charmbracelet/lipgloss) — terminal styling
- [fsnotify/fsnotify](https://github.com/fsnotify/fsnotify) — filesystem watcher
- [fatih/color](https://github.com/fatih/color) — colored log output

## See also

- [gotestyourself/gotestsum](https://github.com/gotestyourself/gotestsum) — the file watcher implementation is adapted
  from gotestsum
- [lusingander/gotip](https://github.com/lusingander/gotip)
