package tel

import (
	"path/filepath"
	"strings"
)

type FileSet struct {
	Path  string
	Files []File
}

type File struct {
	Path    string
	Name    string
	Content []byte
}

// NewFile returns a File named name located in dir containing the following content
func NewFile(dir, name string, content []byte) File {
	filename := strings.Replace(filepath.Base(name), filepath.Ext(name), "", 1)
	f := File{
		Path:    dir,
		Name:    filename,
		Content: content,
	}
	return f
}

// NewFileSet returns a FileSet for files in the directory stored in path
func NewFileSet(path string, files ...File) FileSet {
	return FileSet{Path: path, Files: files}
}

// Add creates a File named name containing content and adds it to f
func (f *FileSet) Add(name string, content []byte) {
	file := NewFile(f.Path, name, content)
	f.Files = append(f.Files, file)
}

// Add creates a File named name containing content and adds it to f
func (f *FileSet) AddString(name, content string) {
	file := NewFile(f.Path, name, []byte(content))
	f.Files = append(f.Files, file)
}
