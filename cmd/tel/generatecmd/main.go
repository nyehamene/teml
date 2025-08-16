package generatecmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tel-lang/tel"
	"github.com/tel-lang/tel/go/generator"
	"github.com/tel-lang/tel/go/transpiler"
)

type Arguments struct {
	Stdout io.Writer
	Stderr io.Writer
	Path   string
	Root   *os.Root
	Mode   Mode
	Flag   tel.Flag
}

//go:generate stringer -type=Mode -linecomment
type Mode int

const (
	ModePackage Mode = iota // Package
	ModeFile                // File
)

func Run(args Arguments) error {
	switch args.Mode {
	case ModeFile:
		return generateFile(args)
	case ModePackage:
		return generatePackage(args)
	default:
		panic(fmt.Sprintf("unexpected cmd.Mode: %#v", args.Mode))
	}
}

func generatePackage(args Arguments) error {
	dir := args.Path
	root := args.Root

	packagedir, err := root.Open(dir)
	if err != nil {
		return err
	}
	defer mustCloseFile(packagedir)

	entries, err := packagedir.ReadDir(0)
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	files := tel.NewFileSet(dir)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".tel") {
			continue
		}

		fullname := filepath.Join(dir, entry.Name())
		templateSource, err := root.ReadFile(fullname)
		if err != nil {
			return err
		}

		files.Add(entry.Name(), templateSource)
	}

	if err := generateSource(args, files); err != nil {
		return err
	}

	return nil
}

func generateTel(args Arguments, pkg string) error {
	basedir := args.Path
	if args.Mode == ModeFile {
		basedir = filepath.Dir(args.Path)
	}

	// create tel.go in path
	telSourceFilePath := filepath.Join(basedir, "tel.go")
	telSourceFile, err := args.Root.Create(telSourceFilePath)
	if err != nil {
		return err
	}
	defer mustCloseFile(telSourceFile)

	if err := generator.WriteTel(telSourceFile, pkg); err != nil {
		return err
	}

	return nil
}

func generateFile(args Arguments) error {
	sourceFile, err := args.Root.Open(args.Path)
	if err != nil {
		return err
	}
	defer mustCloseFile(sourceFile)

	sourceBytes, err := io.ReadAll(sourceFile)
	if err != nil {
		return err
	}

	dir := filepath.Dir(args.Path)
	name := filepath.Base(args.Path)
	files := tel.NewFileSet(dir)
	files.Add(name, sourceBytes)

	if err := generateSource(args, files); err != nil {
		return err
	}
	return nil
}

func generateSource(args Arguments, packageFiles tel.FileSet) error {
	flag := args.Flag
	root := args.Root

	ctx := tel.NewContextWithFlags(flag)
	// TBD(fix): handle transpiler error
	src := transpiler.ParseFile(ctx, packageFiles)
	if err := generateTel(args, src.Package); err != nil {
		return err
	}

	basedir := args.Path
	if args.Mode == ModeFile {
		basedir = filepath.Dir(args.Path)
	}

	for _, file := range src.Files {
		// create go source file
		destFilename := file.Name + ".go"
		destFilePath := filepath.Join(basedir, destFilename)
		destFile, err := root.Create(destFilePath)
		if err != nil {
			return err
		}
		defer mustCloseFile(destFile)

		if err := generator.WriteNamespace(destFile, file); err != nil {
			return err
		}
	}

	return nil
}

func mustCloseFile(f *os.File) {
	if err := f.Close(); err != nil {
		panic(err)
	}
}
