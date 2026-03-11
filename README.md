# CLIDesk

> A desktop-style file explorer that lives entirely in your terminal.

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)]()
[![GitHub Repo](https://img.shields.io/badge/GitHub-itsahmadawais%2Fclidesk-181717?logo=github)](https://github.com/itsahmadawais/clidesk)

```
╭──────────────────────────────────────────────────────────────────╮
│  CLIDesk  /home/user/projects                                    │
│                                                                  │
│  📁 api          📁 frontend      📁 docs          📁 scripts   │
│  📁 .github      🐹 main.go       🔧 config.json   📝 README.md │
│  🗜️ release.zip  🔒 go.sum        📕 design.pdf    🐚 deploy.sh │
│  ────────────────────────────────────────────────────────────    │
│  📁 frontend  ·  directory  ·  Mar 11, 2026  16:04              │
│                                                                  │
│  [↑↓←→] Move  ·  [↵] Open  ·  [⌫] Back  ·  [:] Run  ·  [Q] Quit│
╰──────────────────────────────────────────────────────────────────╯
```

---

## Why CLIDesk?

Developers spend most of their time in the terminal, but navigating the filesystem with `ls` means constantly switching between a mental map and a flat text list. CLIDesk brings the **visual clarity of a desktop file browser** into the terminal — icon grid, file details, git status, and a built-in command runner — without ever leaving the CLI.

---

## Features

- **Icon grid layout** — files and folders rendered as icons with labels, columns auto-adjust to terminal width
- **40+ file type icons** — Go, Python, Rust, Node, images, archives, configs, and more
- **Built-in command runner** — press `:` to run `npm run dev`, `git log`, `go build` — output streams live inside CLIDesk, never drops to the normal terminal
- **Git status indicators** — modified `●`, staged `●`, untracked `?`, deleted `✗` shown inline per file
- **File operations** — delete, rename, create files and folders without leaving the app
- **6 built-in themes** — Default, Dracula, Nord, Gruvbox, Monokai, Catppuccin; cycle with `T` or set via flag
- **File info bar** — size, modification date, and git status for the selected item
- **CD-on-quit** — navigate to a folder in CLIDesk, quit, and your terminal moves there (via shell wrapper)
- **Cross-platform** — Windows, macOS, Linux

---

## Installation

### Via Go (recommended)

```bash
go install github.com/itsahmadawais/clidesk@latest
```

> Requires Go 1.21+. The binary is placed in `$GOPATH/bin` (usually already in your `$PATH`).

### Build from source

```bash
git clone https://github.com/itsahmadawais/clidesk.git
cd clidesk
go build -o clidesk .

# Then move it somewhere in your PATH:
# Windows:   move clidesk.exe C:\Windows\System32\
# macOS/Linux: sudo mv clidesk /usr/local/bin/
```

---

## Shell Integration (CD-on-Quit)

Without this step, CLIDesk works as a viewer. With it, quitting CLIDesk moves your terminal to whatever directory you navigated to — making it a genuine navigation tool.

### PowerShell (Windows)

Add to your `$PROFILE`:

```powershell
. "C:\path\to\clidesk\shell\clidesk.ps1"
```

Reload: `. $PROFILE`

### Bash / Zsh (macOS & Linux)

Add to `~/.bashrc` or `~/.zshrc`:

```bash
source /path/to/clidesk/shell/clidesk.sh
```

Reload: `source ~/.bashrc`

---

## Usage

```bash
clidesk                     # open in current directory
clidesk --theme dracula     # start with a specific theme
clidesk --help              # show all flags
```

---

## Keyboard Controls

### Browsing

| Key | Action |
|-----|--------|
| `↑` `↓` `←` `→` | Navigate the grid |
| `Enter` | Open folder / open file with default app |
| `Backspace` | Go to parent directory |
| `Q` | Quit |

### Commands

| Key | Action |
|-----|--------|
| `:` | Open command input — run any shell command in the current directory |
| `↑` `↓` | Scroll output while a command is running |
| `Ctrl+C` | Stop the running command |
| Any key | Return to browser after a command finishes |

### File Operations

| Key | Action |
|-----|--------|
| `D` | Delete selected file or folder (confirms before deleting) |
| `R` | Rename selected item |
| `N` | New file (type name) or folder (type name ending with `/`) |

### UI

| Key | Action |
|-----|--------|
| `T` | Cycle to next theme |
| `?` | Toggle expanded help bar |

---

## Themes

| Name | Character |
|---|---|
| `default` | Blue / indigo |
| `dracula` | Purple + pink |
| `nord` | Icy blue |
| `gruvbox` | Warm orange / gold |
| `monokai` | Neon green + pink |
| `catppuccin` | Soft pastel mauve |

```bash
clidesk --theme nord
```

---

## Command Runner

Press `:` to open the command bar. Any shell command runs in the **current directory** and streams its output live inside CLIDesk:

```
↵ run  ·  Esc cancel
  :  npm run dev█
```

While running:

```
╭────────────────────────────────────────────────────────╮
│  CLIDesk  /projects/myapp  ▸ npm run dev               │
│                                                        │
│    VITE v5.2.11  ready in 328 ms                      │
│    ➜  Local:   http://localhost:5174/                  │
│    ➜  Network: use --host to expose                    │
│                                                        │
│  ────────────────────────────────────────────────────  │
│  ● Running  ·  ↑↓ scroll  ·  Ctrl+C stop              │
╰────────────────────────────────────────────────────────╯
```

After the command exits, press any key to return to the browser. The directory listing refreshes automatically.

---

## Project Structure

```
clidesk/
├── main.go            # App state, event handling, command streaming
├── ui/
│   ├── grid.go        # All rendering (grid, info bar, running view, help)
│   ├── mode.go        # App mode enum
│   └── navigation.go  # Cursor movement logic
├── filesystem/
│   ├── reader.go      # Directory reading with metadata
│   ├── opener.go      # Open files with OS default app
│   ├── gitstat.go     # Git status detection
│   └── fileops.go     # Delete, rename, create
├── icons/icons.go     # 40+ file type → emoji mappings
├── themes/themes.go   # Theme definitions
└── shell/             # CD-on-quit wrappers (PS1 + sh)
```

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for a full technical deep-dive.

---

## Roadmap

- [ ] File preview panel (text, images)
- [ ] Fuzzy search (`/` to filter)
- [ ] Multi-select with batch operations
- [ ] Tabs for multiple directories
- [ ] SSH / remote directory browsing
- [ ] Homebrew tap and Scoop bucket for one-line install

---

## Contributing

Pull requests are welcome. For large changes, please open an issue first.

```bash
git clone https://github.com/itsahmadawais/clidesk.git
cd clidesk
go run .          # run in dev mode
go test ./...     # run tests
```

---

## License

MIT — see [LICENSE](LICENSE) for details.
