package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/itsahmadawais/clidesk/filesystem"
	"github.com/itsahmadawais/clidesk/themes"

	"github.com/charmbracelet/lipgloss"
)

const CellWidth = 22

// ViewState is everything RenderGrid needs to produce a frame.
type ViewState struct {
	// ── Browser fields ────────────────────────────────────────────────────────
	Entries      []filesystem.Entry
	GitStatus    filesystem.GitStatus
	Cursor       int
	Width        int
	Height       int
	CurrentPath  string
	Theme        themes.Theme
	Mode         Mode
	InputText    string
	ShowFullHelp bool

	// ── Running-command fields (ModeRunning only) ─────────────────────────────
	RunCommand string
	RunOutput  []string
	RunDone    bool
	RunErr     error
	RunScroll  int // lines scrolled up from the bottom (0 = follow tail)
}

// gridStyles holds all Lipgloss styles derived from a Theme.
type gridStyles struct {
	title       lipgloss.Style
	path        lipgloss.Style
	selected    lipgloss.Style
	selectedDir lipgloss.Style
	dir         lipgloss.Style
	file        lipgloss.Style
	border      lipgloss.Style
}

func newGridStyles(t themes.Theme) gridStyles {
	return gridStyles{
		title: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.Title)),
		path:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Path)),

		selected: lipgloss.NewStyle().
			Background(lipgloss.Color(t.SelectedBg)).
			Foreground(lipgloss.Color(t.SelectedFg)).
			Bold(true).Padding(0, 1),

		selectedDir: lipgloss.NewStyle().
			Background(lipgloss.Color(t.SelectedBg)).
			Foreground(lipgloss.Color(t.SelectedDir)).
			Bold(true).Padding(0, 1),

		dir:  lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dir)).Padding(0, 1),
		file: lipgloss.NewStyle().Foreground(lipgloss.Color(t.File)).Padding(0, 1),

		border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.Border)).
			Padding(1, 2),
	}
}

func Columns(width int) int {
	inner := width - 8
	cols := inner / CellWidth
	if cols < 1 {
		cols = 1
	}
	return cols
}

func RenderGrid(vs ViewState) string {
	s := newGridStyles(vs.Theme)
	t := vs.Theme

	// ── Header ────────────────────────────────────────────────────────────────
	header := s.title.Render("CLIDesk") + "  " + s.path.Render(vs.CurrentPath)

	// ── Running mode: replace the whole body ──────────────────────────────────
	if vs.Mode == ModeRunning {
		cmdLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dir)).Render("▸ " + vs.RunCommand)
		runHeader := header + "  " + cmdLabel
		body := renderRunning(vs, t)
		content := runHeader + "\n\n" + body
		return s.border.Width(vs.Width - 4).Render(content)
	}

	cols := Columns(vs.Width)

	// ── File grid ─────────────────────────────────────────────────────────────
	var body strings.Builder

	if len(vs.Entries) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Empty)).Italic(true)
		body.WriteString(emptyStyle.Render("(empty directory)"))
	} else {
		rows := (len(vs.Entries) + cols - 1) / cols

		// Reserve lines for: header(1) + blank(1) + divider(1) + info bar(1) + blank(1) + bottom(2) + border padding(4)
		maxVisibleRows := vs.Height - 11
		if maxVisibleRows < 1 {
			maxVisibleRows = 1
		}

		cursorRow := vs.Cursor / cols
		startRow := 0
		if cursorRow >= maxVisibleRows {
			startRow = cursorRow - maxVisibleRows + 1
		}
		endRow := startRow + maxVisibleRows
		if endRow > rows {
			endRow = rows
		}

		for row := startRow; row < endRow; row++ {
			var rowParts []string
			for col := 0; col < cols; col++ {
				idx := row*cols + col
				if idx >= len(vs.Entries) {
					rowParts = append(rowParts, strings.Repeat(" ", CellWidth))
					continue
				}
				entry := vs.Entries[idx]

				// Build git badge suffix
				gitSuffix := ""
				if vs.GitStatus != nil {
					if xy, ok := vs.GitStatus[entry.Name]; ok {
						sym, colour := filesystem.GitBadge(xy)
						if sym != "" {
							gitSuffix = " " + lipgloss.NewStyle().Foreground(lipgloss.Color(colour)).Render(sym)
						}
					}
				}

				nameMax := CellWidth - 5
				if gitSuffix != "" {
					nameMax -= 2
				}
				label := fmt.Sprintf("%s %s", entry.Icon, truncate(entry.Name, nameMax))

				var cell string
				switch {
				case idx == vs.Cursor && entry.IsDir:
					cell = s.selectedDir.Width(CellWidth).Render(label) + gitSuffix
				case idx == vs.Cursor:
					cell = s.selected.Width(CellWidth).Render(label) + gitSuffix
				case entry.IsDir:
					cell = s.dir.Width(CellWidth).Render(label) + gitSuffix
				default:
					cell = s.file.Width(CellWidth).Render(label) + gitSuffix
				}
				rowParts = append(rowParts, cell)
			}
			body.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowParts...))
			body.WriteString("\n")
		}
	}

	// ── Divider ───────────────────────────────────────────────────────────────
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Help)).
		Render(strings.Repeat("─", vs.Width-12))

	// ── Info bar (selected item details) ─────────────────────────────────────
	infoBar := renderInfoBar(vs, t)

	// ── Bottom section: input bar or keybinding help ──────────────────────────
	bottom := renderBottom(vs, t)

	content := header + "\n\n" + body.String() + divider + "\n" + infoBar + "\n\n" + bottom

	return s.border.Width(vs.Width - 4).Render(content)
}

// renderInfoBar shows details for the currently selected entry.
func renderInfoBar(vs ViewState, t themes.Theme) string {
	if len(vs.Entries) == 0 {
		return ""
	}
	entry := vs.Entries[vs.Cursor]
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Help))
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Title))
	sep := muted.Render("  ·  ")

	name := accent.Render(entry.Icon + " " + entry.Name)

	var sizeStr string
	if entry.IsDir {
		sizeStr = muted.Render("directory")
	} else {
		sizeStr = muted.Render(humanSize(entry.Size))
	}

	var dateStr string
	if !entry.ModTime.IsZero() {
		dateStr = muted.Render(entry.ModTime.Format("Jan 02, 2006  15:04"))
	}

	var gitStr string
	if vs.GitStatus != nil {
		if xy, ok := vs.GitStatus[entry.Name]; ok {
			sym, colour := filesystem.GitBadge(xy)
			gitStr = lipgloss.NewStyle().Foreground(lipgloss.Color(colour)).Render(sym + " " + gitLabel(xy))
		}
	}

	parts := []string{name, sizeStr}
	if dateStr != "" {
		parts = append(parts, dateStr)
	}
	if gitStr != "" {
		parts = append(parts, gitStr)
	}
	return strings.Join(parts, sep)
}

func gitLabel(xy string) string {
	if len(xy) < 2 {
		return ""
	}
	switch {
	case xy == "??":
		return "untracked"
	case xy[0] == 'A':
		return "added"
	case xy[0] == 'M' || xy[1] == 'M':
		return "modified"
	case xy[0] == 'R':
		return "renamed"
	case xy[0] == 'D' || xy[1] == 'D':
		return "deleted"
	default:
		return "changed"
	}
}

// renderBottom renders the context-sensitive area below the divider.
// Input modes use two lines: hint on top, input field below.
func renderBottom(vs ViewState, t themes.Theme) string {
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.Title))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Help))
	hint := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Help)).Italic(true)
	cur := lipgloss.NewStyle().Foreground(lipgloss.Color(t.SelectedBg)).Render("█")

	hintSep := "  " + muted.Render("·") + "  "

	switch vs.Mode {
	case ModeCommand:
		hintLine := hint.Render("↵ run") + hintSep + hint.Render("Esc cancel")
		inputLine := accent.Render("  :  ") + muted.Render(vs.InputText) + cur
		return hintLine + "\n" + inputLine

	case ModeRename:
		hintLine := hint.Render("↵ confirm") + hintSep + hint.Render("Esc cancel")
		inputLine := accent.Render("  Rename →  ") + muted.Render(vs.InputText) + cur
		return hintLine + "\n" + inputLine

	case ModeNewItem:
		hintLine := hint.Render("end name with / to create a directory") + hintSep + hint.Render("↵ confirm") + hintSep + hint.Render("Esc cancel")
		inputLine := accent.Render("  New →  ") + muted.Render(vs.InputText) + cur
		return hintLine + "\n" + inputLine

	case ModeConfirmDelete:
		name := ""
		if len(vs.Entries) > 0 {
			name = vs.Entries[vs.Cursor].Name
		}
		hintLine := muted.Render("This cannot be undone.")
		confirmLine := accent.Render("  Delete  ") + muted.Render("\""+name+"\"?") +
			"    " + accent.Render("[Y]") + muted.Render(" Yes") +
			hintSep + accent.Render("[N]") + muted.Render(" No")
		return hintLine + "\n" + confirmLine

	default:
		return renderHelpBar(vs.Theme, vs.ShowFullHelp)
	}
}

// renderRunning renders the embedded command-output view.
func renderRunning(vs ViewState, t themes.Theme) string {
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Help))
	sep := "  " + muted.Render("·") + "  "

	// Lines available for output:
	// header(1) + blank(1) + divider(1) + status(1) + border-padding(4) = 8
	visibleLines := vs.Height - 8
	if visibleLines < 1 {
		visibleLines = 1
	}

	lines := vs.RunOutput
	total := len(lines)

	// Tail-follow unless scrolled up
	end := total - vs.RunScroll
	if end < 0 {
		end = 0
	}
	start := end - visibleLines
	if start < 0 {
		start = 0
	}

	var body strings.Builder
	for i := start; i < end; i++ {
		body.WriteString("  " + lines[i] + "\n")
	}
	// Pad with blank lines so the divider stays pinned to the same row
	for i := end - start; i < visibleLines; i++ {
		body.WriteString("\n")
	}

	divider := muted.Render(strings.Repeat("─", vs.Width-12))

	var statusLine string
	if vs.RunDone {
		if vs.RunErr != nil {
			statusLine = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FB4934")).Render("✗  Exited with error") +
				sep + muted.Render("any key to return")
		} else {
			statusLine = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#B8BB26")).Render("✓  Done") +
				sep + muted.Render("any key to return")
		}
	} else {
		scrollHint := ""
		if vs.RunScroll > 0 {
			scrollHint = sep + muted.Render(fmt.Sprintf("↑ %d lines up", vs.RunScroll))
		}
		statusLine = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.Title)).Render("● Running") +
			sep + muted.Render("↑↓ scroll") +
			sep + lipgloss.NewStyle().Foreground(lipgloss.Color(t.Title)).Render("Ctrl+C") + muted.Render(" stop") +
			scrollHint
	}

	return body.String() + divider + "\n" + statusLine
}

// renderHelpBar renders the normal-mode keybinding row.
func renderHelpBar(t themes.Theme, full bool) string {
	type binding struct{ key, action string }

	compact := []binding{
		{"↑↓←→", "Move"},
		{"↵", "Open"},
		{"⌫", "Back"},
		{":", "Run"},
		{"Q", "Quit"},
		{"?", "More"},
	}
	expanded := []binding{
		{"↑↓←→", "Move"},
		{"↵", "Open"},
		{"⌫", "Back"},
		{":", "Run"},
		{"D", "Delete"},
		{"R", "Rename"},
		{"N", "New"},
		{"T", t.Name + " →"},
		{"Q", "Quit"},
		{"?", "Less"},
	}

	bindings := compact
	if full {
		bindings = expanded
	}

	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(t.Title))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Help))
	sep := "  " + muted.Render("·") + "  "

	parts := make([]string, len(bindings))
	for i, b := range bindings {
		parts[i] = muted.Render("[") + keyStyle.Render(b.key) + muted.Render("]") + " " + muted.Render(b.action)
	}
	return strings.Join(parts, sep)
}

func humanSize(b int64) string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%d B", b)
	case b < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	case b < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/1024/1024)
	default:
		return fmt.Sprintf("%.1f GB", float64(b)/1024/1024/1024)
	}
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// Ensure time import is used (ModTime formatting).
var _ = time.Now
