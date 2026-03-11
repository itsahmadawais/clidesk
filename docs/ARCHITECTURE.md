# CLIDesk — Architecture & Technical Overview

## What Is CLIDesk?

CLIDesk is a terminal-based file explorer written in Go. It replaces the traditional `ls` / `dir` output with a **desktop-style icon grid**, fully keyboard-driven, built entirely inside the terminal using a TUI (Terminal User Interface) framework.

---

## Tech Stack

| Layer | Tool | Purpose |
|---|---|---|
| Language | [Go](https://go.dev/) 1.25+ | Core implementation |
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Event loop, model-view-update pattern |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal colors, borders, layout |
| Filesystem | `os`, `path/filepath` | Directory reading, file operations |
| Process I/O | `os/exec`, `io.Pipe`, `bufio` | Streaming command output |

---

## Project Structure

```
clidesk/
├── main.go                  # Bubble Tea model, all state, event handling
│
├── ui/
│   ├── mode.go              # App mode enum (Normal, Running, Command, etc.)
│   ├── grid.go              # All rendering: grid, info bar, running view, help bar
│   └── navigation.go        # Pure cursor movement functions
│
├── filesystem/
│   ├── reader.go            # Read directory → []Entry (sorted, with metadata)
│   ├── opener.go            # Open a file with the OS default app
│   ├── gitstat.go           # Git status detection via git status --porcelain
│   └── fileops.go           # Delete, rename, create file/folder
│
├── icons/
│   └── icons.go             # File extension → emoji icon mapping (40+ types)
│
├── themes/
│   └── themes.go            # Theme struct + 6 built-in themes
│
├── shell/
│   ├── clidesk.ps1          # PowerShell wrapper (enables cd-on-quit)
│   └── clidesk.sh           # Bash/Zsh wrapper (enables cd-on-quit)
│
├── docs/
│   ├── ARCHITECTURE.md      # This file
│   └── DISTRIBUTION.md      # Release and packaging guide
│
├── go.mod
├── go.sum
└── README.md
```

---

## Architecture: Bubble Tea MVU Pattern

CLIDesk follows Bubble Tea's **Model-View-Update** (MVU) pattern, also known as The Elm Architecture.

```
┌─────────────────────────────────────────────────────────┐
│                      Bubble Tea Loop                    │
│                                                         │
│   ┌──────────┐    Msg     ┌──────────┐    tea.Cmd      │
│   │  Update  │ ─────────► │  Model   │ ──────────────► │
│   │          │ ◄───────── │  (state) │                 │
│   └──────────┘   (model,  └──────────┘                 │
│         │         cmd)         │                       │
│         │                      │                       │
│         ▼                      ▼                       │
│   ┌──────────┐           ┌──────────┐                  │
│   │  Events  │           │   View   │ ──► terminal     │
│   │ keyboard │           │ (string) │                  │
│   │ window   │           └──────────┘                  │
│   │ output   │                                         │
│   └──────────┘                                         │
└─────────────────────────────────────────────────────────┘
```

### Model (`main.go`)

The `model` struct is the single source of truth for all application state:

```go
type model struct {
    // Browser state
    currentPath  string
    entries      []filesystem.Entry
    gitStatus    filesystem.GitStatus
    cursor       int
    width, height int
    themeIdx     int
    mode         ui.Mode
    inputText    string
    showFullHelp bool

    // Running-command state
    runCmdStr  string
    runOutput  []string
    runScroll  int
    runCmd     *exec.Cmd
    runDone    bool
    runErr     error
    outputChan chan tea.Msg
}
```

### App Modes

| Mode | Trigger | Description |
|---|---|---|
| `ModeNormal` | default | File browsing |
| `ModeCommand` | `:` | Typing a shell command |
| `ModeRunning` | Enter in ModeCommand | Command executing, output streaming |
| `ModeConfirmDelete` | `D` | Waiting for Y/N |
| `ModeRename` | `R` | Typing new filename |
| `ModeNewItem` | `N` | Typing new file/folder name |

---

## File Grid Rendering (`ui/grid.go`)

### Layout

```
╭──────────────────────────────────────────────────────╮
│  CLIDesk  /path/to/directory                         │  ← header
│                                                      │
│  📁 Desktop    📁 Documents   📁 Projects            │  ─┐
│  📁 Downloads  📝 notes.txt   🐹 main.go             │   │ file grid
│  📕 resume.pdf 🔧 config.json 🗜️ archive.zip        │  ─┘
│  ─────────────────────────────────────────────────   │  ← divider
│  📄 main.go  ·  5.2 KB  ·  Mar 11, 2026  ·  ● mod  │  ← info bar
│                                                      │
│  [↑↓←→] Move  ·  [↵] Open  ·  [:] Run  ·  [Q] Quit │  ← help bar
╰──────────────────────────────────────────────────────╯
```

### Column Calculation

Columns are calculated dynamically from the terminal width:

```go
cols = (terminalWidth - 8) / CellWidth   // CellWidth = 22
```

### Grid Scrolling

Only the visible window of rows is rendered. The cursor row is always kept on screen:

```go
cursorRow = cursor / cols
startRow  = max(0, cursorRow - maxVisibleRows + 1)
endRow    = min(rows, startRow + maxVisibleRows)
```

---

## Embedded Command Runner (`ModeRunning`)

When a command is run, CLIDesk **never switches to the normal terminal**. Output is piped into the model and rendered inside the TUI.

### Stream Architecture

```
              ┌─────────────────────────────────────────────┐
              │  cmd.Stdout ──► io.PipeWriter                │
              │  cmd.Stderr ──►        │                     │
              │                        ▼                     │
              │               io.PipeReader                  │
              │                        │                     │
              │             bufio.Scanner (goroutine)        │
              │                        │                     │
              │              chan tea.Msg (buffered 512)      │
              │                        │                     │
              │            waitForOutput() tea.Cmd            │
              │                        │                     │
              │              Bubble Tea Update loop           │
              │                        │                     │
              │           append to model.runOutput          │
              └─────────────────────────────────────────────┘
```

### Sequence

1. User presses `:`, types `npm run dev`, presses Enter
2. `startCommand()` creates `io.Pipe`, assigns both ends to `cmd.Stdout`/`cmd.Stderr`
3. Two goroutines start: one waits for `cmd.Wait()`, the other scans lines from the pipe
4. Each line becomes an `outputLineMsg` sent to Bubble Tea's event loop
5. `Update()` appends to `model.runOutput`, requests next line via `waitForOutput()`
6. `View()` renders the latest N lines that fit on screen
7. On exit: `outputDoneMsg` arrives, status bar switches to `✓ Done` or `✗ Exited with error`
8. Any key press returns to browser with directory refreshed

---

## Git Status (`filesystem/gitstat.go`)

Runs `git status --porcelain -u` in the current directory and parses the two-character XY status codes into a `map[string]string` keyed by filename.

| Code | Meaning | Badge |
|---|---|---|
| `A ` | Staged new file | `●` green |
| `M ` or ` M` | Modified | `●` yellow |
| `??` | Untracked | `?` red |
| `D ` or ` D` | Deleted | `✗` red |

The badge appears next to the file name in the grid cell and in the info bar.

---

## Themes (`themes/themes.go`)

Each theme defines 10 color slots:

```go
type Theme struct {
    Name, Title, Path           string  // header area
    SelectedBg, SelectedFg      string  // selected cell
    Dir, SelectedDir            string  // directory cells
    File                        string  // regular file cells
    Help, Border, Empty         string  // chrome
}
```

Built-in themes: `default`, `dracula`, `nord`, `gruvbox`, `monokai`, `catppuccin`.

Selected via `--theme <name>` flag or cycled live with the `T` key.

---

## CD-on-Quit (`shell/`)

A fundamental limitation of any TUI tool is that it cannot change the parent shell's working directory directly (processes cannot affect their parent). CLIDesk solves this with a thin shell wrapper:

1. The wrapper calls `clidesk.exe --print-dir <tmpfile>`
2. On quit, CLIDesk writes `model.currentPath` to the temp file
3. The wrapper reads the path and calls `cd`

**PowerShell** (`shell/clidesk.ps1`) — source from `$PROFILE`  
**Bash/Zsh** (`shell/clidesk.sh`) — source from `~/.bashrc` or `~/.zshrc`

---

## Key Design Decisions

| Decision | Rationale |
|---|---|
| No PTY for command output | PTY requires platform-specific code (ConPTY on Windows). `io.Pipe` works everywhere and covers the majority of use cases. |
| Value-type model with channel | Bubble Tea requires a copyable model. Go channels are reference types so they copy by pointer, keeping a single shared channel across model copies. |
| Dirs listed before files | Standard file explorer convention; developers expect folders first. |
| `tea.WithAltScreen()` | Keeps CLIDesk's UI isolated from the scrollback buffer; quitting cleanly restores the terminal. |
| 10,000 line cap on output | Prevents unbounded memory growth for long-running processes. |
