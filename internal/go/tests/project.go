// generate test files from a tel source file
package tests

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tel-lang/tel"
	"github.com/tel-lang/tel/ast/token"

	generator "github.com/tel-lang/tel/go/generator"
)

const testfileName = "render_test.go"

func CreateProject(testdataFS fs.ReadDirFS) (string, error) {
	dir, err := os.MkdirTemp("", "tel_test_*")
	if err != nil {
		return "", fmt.Errorf("failed to create test dir: %w", err)
	}
	if _, err := createFilesInDir(testdataFS, dir, "testdata"); err != nil {
		return dir, err
	}
	return dir, nil
}

func createFilesInDir(testdataFS fs.ReadDirFS, dest, dir string) (string, error) {
	files, err := testdataFS.ReadDir(dir)
	if err != nil {
		return dir, fmt.Errorf("failed to read embedded dir: %w", err)
	}
	for _, file := range files {
		if file.IsDir() {
			srcdir := filepath.Join(dir, file.Name())
			destdir := filepath.Join(dest, file.Name())
			if err := os.MkdirAll(destdir, 0777); err != nil {
				return dir, fmt.Errorf("failed to create dir: %w", err)
			}
			if _, err := createFilesInDir(testdataFS, destdir, srcdir); err != nil {
				return dir, fmt.Errorf("failed to copy dir %s: %w", dir, err)
			}
			continue
		}
		src := filepath.Join(dir, file.Name())
		data, err := os.ReadFile(src)
		if err != nil {
			return dir, fmt.Errorf("failed to read file: %w", err)
		}
		target := filepath.Join(dest, file.Name())
		err = os.WriteFile(target, data, 0660)
		if err != nil {
			return dir, fmt.Errorf("failed to copy file: %w", err)
		}
	}
	return dir, nil
}

func GenerateTest(root *os.Root, srcfile tel.File, flag tel.Flag) (err error) {
	testfilename := filepath.Join(filepath.Dir(srcfile.Path), testfileName)
	testfile, createErr := root.Create(testfilename)
	if createErr != nil {
		return createErr
	}
	defer func() {
		if closeErr := testfile.Close(); closeErr != nil {
			if err == nil {
				err = errors.Join(err, closeErr)
			}
		}
	}()

	rootFS := root.FS()
	fileFS := rootFS.(fs.ReadFileFS)

	toks := token.ScanInput(srcfile, tel.PreserveComment|flag)
	testdata := getTestDataFromComment(toks)
	if writeErr := writeTestFile(fileFS, testfile, testdata); writeErr != nil {
		return writeErr
	}

	return nil
}

func getTestDataFromComment(toks *token.File) string {
	for _, tok := range toks.Tokens {
		if tok.Kind != token.Comment {
			continue
		}

		cmt, ok := toks.Text(tok)
		if !ok {
			continue
		}

		if !strings.HasPrefix(cmt, ";value") {
			continue
		}

		cmt = strings.Replace(cmt, ";value", "", 1)
		cmt = strings.TrimLeft(cmt, " ")
		return cmt
	}
	return ""
}

func writeTestFile(testdataFS fs.ReadFileFS, w io.Writer, data string) error {
	model := map[string]any{
		"Data":              data,
		"ContextStructName": generator.ContextStructName,
		"RenderMethodName":  generator.RenderMethodName,
	}
	bytes, err := testdataFS.ReadFile("test.go.txt")
	if err != nil {
		return err
	}
	tmpl := template.Must(template.New("testcode").Parse(string(bytes)))
	if err := tmpl.Execute(w, model); err != nil {
		return err
	}
	return nil
}
