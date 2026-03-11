package ui

// Mode represents the current interaction state of the app.
type Mode int

const (
	ModeNormal        Mode = iota // default browsing
	ModeCommand                   // ":" — typing a shell command
	ModeConfirmDelete             // "d" — waiting for y/n
	ModeRename                    // "r" — typing a new name
	ModeNewItem                   // "n" — typing name for new file/folder
	ModeRunning                   // command is running, output shown in-app
)
