package source

import (
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	Path    string
	Name    string
	Content []byte
}

func NewFile(path string, content []byte) File {
	// TODO remove extension
	ext := filepath.Ext(path)
	name := filepath.Base(path)
	name = strings.Replace(name, ext, "", 1)
	f := File{
		Path:    path,
		Name:    name,
		Content: content,
	}
	return f
}

func OpenFile(path string) (File, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	return NewFile(path, content), nil
}
