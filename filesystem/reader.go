package filesystem

import (
	"os"
	"sort"
	"time"

	"github.com/itsahmadawais/clidesk/icons"
)

type Entry struct {
	Name    string
	IsDir   bool
	Icon    string
	Size    int64
	ModTime time.Time
}

func ReadDir(path string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs, files []Entry

	for _, de := range dirEntries {
		info, _ := de.Info()
		var size int64
		var modTime time.Time
		if info != nil {
			size = info.Size()
			modTime = info.ModTime()
		}

		entry := Entry{
			Name:    de.Name(),
			IsDir:   de.IsDir(),
			Icon:    icons.GetIcon(de.Name(), de.IsDir()),
			Size:    size,
			ModTime: modTime,
		}
		if de.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	return append(dirs, files...), nil
}
