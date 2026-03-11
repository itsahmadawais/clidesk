package filesystem

import (
	"os"
	"path/filepath"
)

func DeleteEntry(path string, isDir bool) error {
	if isDir {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

func RenameEntry(oldPath, newName string) error {
	newPath := filepath.Join(filepath.Dir(oldPath), newName)
	return os.Rename(oldPath, newPath)
}

func CreateFile(dir, name string) error {
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	return f.Close()
}

func CreateDir(dir, name string) error {
	return os.MkdirAll(filepath.Join(dir, name), 0755)
}
