package filesystem

import (
	"os/exec"
	"runtime"
)

// OpenFile opens a file or URL with the OS default application.
func OpenFile(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // linux and other unix-likes
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}
