package icons

import "strings"

func GetIcon(name string, isDir bool) string {
	if isDir {
		return "📁"
	}

	dot := strings.LastIndex(name, ".")
	if dot == -1 {
		return "📄"
	}
	ext := strings.ToLower(name[dot:])

	switch ext {
	case ".go":
		return "🐹"
	case ".md", ".txt", ".rst", ".log":
		return "📝"
	case ".js", ".mjs", ".cjs":
		return "📜"
	case ".ts", ".tsx":
		return "📘"
	case ".jsx":
		return "⚛️"
	case ".py":
		return "🐍"
	case ".html", ".htm":
		return "🌐"
	case ".css", ".scss", ".sass", ".less":
		return "🎨"
	case ".json":
		return "🔧"
	case ".yaml", ".yml":
		return "⚙️"
	case ".toml", ".ini", ".cfg", ".conf":
		return "🔩"
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".ico":
		return "🖼️"
	case ".svg":
		return "✏️"
	case ".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv":
		return "🎬"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return "🎵"
	case ".zip", ".tar", ".gz", ".bz2", ".rar", ".7z", ".xz":
		return "🗜️"
	case ".pdf":
		return "📕"
	case ".doc", ".docx":
		return "📘"
	case ".xls", ".xlsx":
		return "📗"
	case ".ppt", ".pptx":
		return "📙"
	case ".exe", ".msi":
		return "⚙️"
	case ".sh", ".bash", ".zsh", ".fish":
		return "🐚"
	case ".bat", ".cmd", ".ps1":
		return "🖥️"
	case ".rs":
		return "🦀"
	case ".java", ".class", ".jar":
		return "☕"
	case ".c", ".h":
		return "🔵"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "🔷"
	case ".rb":
		return "💎"
	case ".php":
		return "🐘"
	case ".swift":
		return "🍎"
	case ".kt", ".kts":
		return "🟣"
	case ".lua":
		return "🌙"
	case ".r":
		return "📊"
	case ".sql":
		return "🗃️"
	case ".dockerfile", ".containerfile":
		return "🐳"
	case ".env":
		return "🔐"
	case ".gitignore", ".gitattributes":
		return "🙈"
	case ".lock":
		return "🔒"
	case ".sum":
		return "🔑"
	default:
		return "📄"
	}
}
