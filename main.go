package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/itsahmadawais/clidesk/filesystem"
	"github.com/itsahmadawais/clidesk/themes"
	"github.com/itsahmadawais/clidesk/ui"
)

// ── Message types ─────────────────────────────────────────────────────────────

type cmdDoneMsg struct{ err error }

// outputLineMsg carries one line of captured command output.
type outputLineMsg string

// outputDoneMsg signals the command has exited (all output already delivered).
type outputDoneMsg struct{ err error }

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	// Browser state
	currentPath  string
	entries      []filesystem.Entry
	gitStatus    filesystem.GitStatus
	cursor       int
	width        int
	height       int
	err          error
	statusMsg    string
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

const maxOutputLines = 10_000 // cap stored lines to avoid unbounded memory

func initialModel(startTheme themes.Theme) model {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	entries, fsErr := filesystem.ReadDir(cwd)

	idx := 0
	for i, t := range themes.All {
		if t.Name == startTheme.Name {
			idx = i
			break
		}
	}

	return model{
		currentPath: cwd,
		entries:     entries,
		gitStatus:   filesystem.LoadGitStatus(cwd),
		cursor:      0,
		width:       80,
		height:      24,
		err:         fsErr,
		themeIdx:    idx,
	}
}

func (m model) theme() themes.Theme { return themes.All[m.themeIdx] }

func (m model) Init() tea.Cmd { return nil }

func (m *model) refreshDir() {
	entries, err := filesystem.ReadDir(m.currentPath)
	if err == nil {
		m.entries = entries
		if m.cursor >= len(m.entries) && len(m.entries) > 0 {
			m.cursor = len(m.entries) - 1
		}
	}
	m.gitStatus = filesystem.LoadGitStatus(m.currentPath)
}

func (m *model) navigateTo(path string) bool {
	entries, err := filesystem.ReadDir(path)
	if err != nil {
		m.statusMsg = fmt.Sprintf("Cannot open: %v", err)
		return false
	}
	m.currentPath = path
	m.entries = entries
	m.cursor = 0
	m.err = nil
	m.gitStatus = filesystem.LoadGitStatus(path)
	return true
}

func (m model) selectedEntry() *filesystem.Entry {
	if len(m.entries) == 0 {
		return nil
	}
	e := m.entries[m.cursor]
	return &e
}

func (m model) selectedPath() string {
	e := m.selectedEntry()
	if e == nil {
		return m.currentPath
	}
	return filepath.Join(m.currentPath, e.Name)
}

// ── Command streaming ─────────────────────────────────────────────────────────

// waitForOutput returns a Cmd that blocks until the next item arrives on ch.
func waitForOutput(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-ch
	}
}

// startCommand launches cmdStr as a subprocess, pipes stdout+stderr into the
// model's output buffer, and switches to ModeRunning.
func (m model) startCommand(cmdStr string) (model, tea.Cmd) {
	m.mode = ui.ModeRunning
	m.runCmdStr = cmdStr
	m.runOutput = nil
	m.runDone = false
	m.runErr = nil
	m.runScroll = 0

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	cmd.Dir = m.currentPath

	// Pipe both stdout and stderr into a single reader.
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pr.Close()
		pw.Close()
		m.mode = ui.ModeNormal
		m.statusMsg = fmt.Sprintf("Failed to start: %v", err)
		return m, nil
	}

	m.runCmd = cmd
	ch := make(chan tea.Msg, 512)
	m.outputChan = ch

	// Goroutine 1: wait for the process, then close the write-end of the pipe.
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
		pw.Close()
	}()

	// Goroutine 2: scan lines from the read-end, forward to ch, then send done.
	go func() {
		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 512*1024), 512*1024)
		for scanner.Scan() {
			ch <- outputLineMsg(scanner.Text())
		}
		pr.Close()
		err := <-errCh
		ch <- outputDoneMsg{err: err}
	}()

	return m, waitForOutput(ch)
}

// killRunning terminates the running process (whole process tree on Windows).
func (m *model) killRunning() {
	if m.runCmd == nil || m.runCmd.Process == nil {
		return
	}
	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/T", "/PID",
			fmt.Sprintf("%d", m.runCmd.Process.Pid)).Run()
	} else {
		m.runCmd.Process.Kill()
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	// ── Streamed output ───────────────────────────────────────────────────────
	case outputLineMsg:
		m.runOutput = append(m.runOutput, string(msg))
		if len(m.runOutput) > maxOutputLines {
			m.runOutput = m.runOutput[len(m.runOutput)-maxOutputLines:]
		}
		// Keep requesting the next line.
		return m, waitForOutput(m.outputChan)

	case outputDoneMsg:
		m.runDone = true
		m.runErr = msg.err
		m.runCmd = nil
		return m, nil

	case tea.KeyMsg:
		// ── Running mode keys ─────────────────────────────────────────────────
		if m.mode == ui.ModeRunning {
			return m.handleRunningKey(msg)
		}
		// ── Text-input modes ──────────────────────────────────────────────────
		if m.mode != ui.ModeNormal {
			return m.handleInputKey(msg)
		}
		// ── Normal browsing ───────────────────────────────────────────────────
		return m.handleNormalKey(msg)
	}

	return m, nil
}

func (m model) handleRunningKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		if !m.runDone {
			m.killRunning()
		}

	case "up":
		m.runScroll++

	case "down":
		if m.runScroll > 0 {
			m.runScroll--
		}

	case "q":
		// Kill if still running, then quit CLIDesk.
		if !m.runDone {
			m.killRunning()
		}
		return m, tea.Quit

	default:
		// Any other key returns to the browser once the process has finished.
		if m.runDone {
			m.mode = ui.ModeNormal
			m.runOutput = nil
			m.runCmd = nil
			m.outputChan = nil
			m.refreshDir()
		}
	}
	return m, nil
}

func (m model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.statusMsg = ""
	cols := ui.Columns(m.width)

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "t":
		m.themeIdx = (m.themeIdx + 1) % len(themes.All)

	case "?":
		m.showFullHelp = !m.showFullHelp

	case "left":
		m.cursor = ui.MoveLeft(m.cursor, cols)
	case "right":
		m.cursor = ui.MoveRight(m.cursor, len(m.entries), cols)
	case "up":
		m.cursor = ui.MoveUp(m.cursor, cols)
	case "down":
		m.cursor = ui.MoveDown(m.cursor, len(m.entries), cols)

	case "enter":
		if e := m.selectedEntry(); e != nil {
			if e.IsDir {
				m.navigateTo(filepath.Join(m.currentPath, e.Name))
			} else {
				if err := filesystem.OpenFile(m.selectedPath()); err != nil {
					m.statusMsg = fmt.Sprintf("Cannot open: %v", err)
				}
			}
		}

	case "backspace":
		parent := filepath.Dir(m.currentPath)
		if parent != m.currentPath {
			m.navigateTo(parent)
		}

	case ":":
		m.mode = ui.ModeCommand
		m.inputText = ""

	case "d", "D":
		if m.selectedEntry() != nil {
			m.mode = ui.ModeConfirmDelete
		}

	case "r", "R":
		if e := m.selectedEntry(); e != nil {
			m.mode = ui.ModeRename
			m.inputText = e.Name
		}

	case "n", "N":
		m.mode = ui.ModeNewItem
		m.inputText = ""
	}

	return m, nil
}

func (m model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = ui.ModeNormal
		m.inputText = ""
		return m, nil

	case tea.KeyBackspace:
		runes := []rune(m.inputText)
		if len(runes) > 0 {
			m.inputText = string(runes[:len(runes)-1])
		}
		return m, nil

	case tea.KeyRunes:
		m.inputText += string(msg.Runes)
		return m, nil

	case tea.KeySpace:
		m.inputText += " "
		return m, nil

	case tea.KeyEnter:
		return m.confirmInput()
	}

	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	return m, nil
}

func (m model) confirmInput() (tea.Model, tea.Cmd) {
	switch m.mode {

	case ui.ModeCommand:
		cmdStr := strings.TrimSpace(m.inputText)
		m.mode = ui.ModeNormal
		m.inputText = ""
		if cmdStr == "" {
			return m, nil
		}
		return m.startCommand(cmdStr)

	case ui.ModeConfirmDelete:
		m.mode = ui.ModeNormal
		if strings.ToLower(strings.TrimSpace(m.inputText)) == "y" {
			e := m.selectedEntry()
			if e != nil {
				if err := filesystem.DeleteEntry(m.selectedPath(), e.IsDir); err != nil {
					m.statusMsg = fmt.Sprintf("Delete failed: %v", err)
				} else {
					m.statusMsg = fmt.Sprintf("Deleted %s", e.Name)
					m.refreshDir()
				}
			}
		}
		m.inputText = ""
		return m, nil

	case ui.ModeRename:
		newName := strings.TrimSpace(m.inputText)
		m.mode = ui.ModeNormal
		m.inputText = ""
		if newName == "" {
			return m, nil
		}
		if err := filesystem.RenameEntry(m.selectedPath(), newName); err != nil {
			m.statusMsg = fmt.Sprintf("Rename failed: %v", err)
		} else {
			m.statusMsg = fmt.Sprintf("Renamed to %s", newName)
			m.refreshDir()
		}
		return m, nil

	case ui.ModeNewItem:
		name := strings.TrimSpace(m.inputText)
		m.mode = ui.ModeNormal
		m.inputText = ""
		if name == "" {
			return m, nil
		}
		var err error
		if strings.HasSuffix(name, "/") {
			err = filesystem.CreateDir(m.currentPath, strings.TrimSuffix(name, "/"))
		} else {
			err = filesystem.CreateFile(m.currentPath, name)
		}
		if err != nil {
			m.statusMsg = fmt.Sprintf("Create failed: %v", err)
		} else {
			m.statusMsg = fmt.Sprintf("Created %s", name)
			m.refreshDir()
		}
		return m, nil
	}

	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\nPress q to quit.\n", m.err)
	}

	view := ui.RenderGrid(ui.ViewState{
		// Browser
		Entries:      m.entries,
		GitStatus:    m.gitStatus,
		Cursor:       m.cursor,
		Width:        m.width,
		Height:       m.height,
		CurrentPath:  m.currentPath,
		Theme:        m.theme(),
		Mode:         m.mode,
		InputText:    m.inputText,
		ShowFullHelp: m.showFullHelp,
		// Running
		RunCommand: m.runCmdStr,
		RunOutput:  m.runOutput,
		RunDone:    m.runDone,
		RunErr:     m.runErr,
		RunScroll:  m.runScroll,
	})

	if m.statusMsg != "" {
		view += "\n  " + m.statusMsg
	}
	return view
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	themeName := flag.String("theme", "default",
		fmt.Sprintf("color theme (%s)", themes.Names()))
	chdirFile := flag.String("print-dir", "",
		"write final directory to this file on exit (used by shell wrapper)")
	flag.Parse()

	p := tea.NewProgram(
		initialModel(themes.Find(*themeName)),
		tea.WithAltScreen(),
	)

	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "CLIDesk error: %v\n", err)
		os.Exit(1)
	}

	if *chdirFile != "" {
		if m, ok := result.(model); ok {
			_ = os.WriteFile(*chdirFile, []byte(m.currentPath), 0600)
		}
	}
}
