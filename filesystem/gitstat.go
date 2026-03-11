package filesystem

import (
	"os/exec"
	"strings"
)

// GitStatus maps a filename to its two-character git status code (e.g. "M ", " M", "??").
type GitStatus map[string]string

// LoadGitStatus runs git status --porcelain in dir and returns a per-file status map.
// Returns nil if the directory is not a git repo or git is unavailable.
func LoadGitStatus(dir string) GitStatus {
	cmd := exec.Command("git", "status", "--porcelain", "-u")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	status := make(GitStatus)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if len(line) < 4 {
			continue
		}
		xy := line[:2]
		name := strings.TrimSpace(line[3:])
		// Renamed: "old -> new" — track the new name
		if idx := strings.Index(name, " -> "); idx != -1 {
			name = name[idx+4:]
		}
		name = strings.Trim(name, "\"")
		// Only keep the top-level segment (so "src/file.go" maps to "src")
		if slash := strings.IndexByte(name, '/'); slash != -1 {
			name = name[:slash]
		}
		if _, exists := status[name]; !exists {
			status[name] = xy
		}
	}
	return status
}

// GitBadge returns a short symbol and hex colour for a two-char git status code.
func GitBadge(xy string) (symbol, colour string) {
	if len(xy) < 2 {
		return "", ""
	}
	x, y := string(xy[0]), string(xy[1])
	switch {
	case xy == "??":
		return "?", "#FB4934" // untracked — red
	case x == "A" || x == "M" || x == "R":
		return "●", "#B8BB26" // staged — green
	case y == "M":
		return "●", "#FABD2F" // modified — yellow
	case x == "D" || y == "D":
		return "✗", "#FB4934" // deleted — red
	default:
		return "●", "#83A598" // other — teal
	}
}
