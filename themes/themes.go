package themes

import "strings"

// Theme defines all color slots used throughout the UI.
type Theme struct {
	Name        string
	Title       string // "CLIDesk" heading
	Path        string // current path text
	SelectedBg  string // selected cell background
	SelectedFg  string // selected cell foreground
	Dir         string // unselected directory foreground
	SelectedDir string // selected directory foreground
	File        string // unselected file foreground
	Help        string // help bar text
	Border      string // outer border
	Empty       string // "(empty directory)" hint
}

// All contains every built-in theme in display order.
var All = []Theme{
	Default,
	Dracula,
	Nord,
	Gruvbox,
	Monokai,
	Catppuccin,
}

var Default = Theme{
	Name:        "default",
	Title:       "75",
	Path:        "245",
	SelectedBg:  "63",
	SelectedFg:  "255",
	Dir:         "117",
	SelectedDir: "159",
	File:        "252",
	Help:        "241",
	Border:      "63",
	Empty:       "241",
}

var Dracula = Theme{
	Name:        "dracula",
	Title:       "#BD93F9",
	Path:        "#6272A4",
	SelectedBg:  "#BD93F9",
	SelectedFg:  "#282A36",
	Dir:         "#8BE9FD",
	SelectedDir: "#282A36",
	File:        "#F8F8F2",
	Help:        "#6272A4",
	Border:      "#FF79C6",
	Empty:       "#6272A4",
}

var Nord = Theme{
	Name:        "nord",
	Title:       "#88C0D0",
	Path:        "#4C566A",
	SelectedBg:  "#5E81AC",
	SelectedFg:  "#ECEFF4",
	Dir:         "#81A1C1",
	SelectedDir: "#ECEFF4",
	File:        "#D8DEE9",
	Help:        "#4C566A",
	Border:      "#88C0D0",
	Empty:       "#4C566A",
}

var Gruvbox = Theme{
	Name:        "gruvbox",
	Title:       "#FABD2F",
	Path:        "#928374",
	SelectedBg:  "#D65D0E",
	SelectedFg:  "#FBF1C7",
	Dir:         "#83A598",
	SelectedDir: "#FBF1C7",
	File:        "#EBDBB2",
	Help:        "#7C6F64",
	Border:      "#D79921",
	Empty:       "#928374",
}

var Monokai = Theme{
	Name:        "monokai",
	Title:       "#A6E22E",
	Path:        "#75715E",
	SelectedBg:  "#F92672",
	SelectedFg:  "#F8F8F2",
	Dir:         "#66D9EF",
	SelectedDir: "#F8F8F2",
	File:        "#F8F8F2",
	Help:        "#75715E",
	Border:      "#A6E22E",
	Empty:       "#75715E",
}

var Catppuccin = Theme{
	Name:        "catppuccin",
	Title:       "#CBA6F7",
	Path:        "#6C7086",
	SelectedBg:  "#CBA6F7",
	SelectedFg:  "#1E1E2E",
	Dir:         "#89DCEB",
	SelectedDir: "#1E1E2E",
	File:        "#CDD6F4",
	Help:        "#585B70",
	Border:      "#CBA6F7",
	Empty:       "#585B70",
}

// Find returns the theme with the given name (case-insensitive).
// Returns Default if not found.
func Find(name string) Theme {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, t := range All {
		if t.Name == name {
			return t
		}
	}
	return Default
}

// Names returns a comma-separated list of all theme names.
func Names() string {
	names := make([]string, len(All))
	for i, t := range All {
		names[i] = t.Name
	}
	return strings.Join(names, ", ")
}
