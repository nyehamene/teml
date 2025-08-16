package generatecmd

import (
	"embed"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/tel-lang/tel/internal/go/tests"
)

//go:embed testdata/file/*
var fileFS embed.FS

//go:embed testdata/package/*
var packageFS embed.FS

func TestRun(t *testing.T) {
	t.Run("can generate in single-file mode", func(t *testing.T) {
		dir, err := tests.CreateProject(fileFS)
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create test dir: %w", err))
		}

		root, err := os.OpenRoot(dir)
		if err != nil {
			t.Fatal(fmt.Errorf("failed to open root: %w", err))
		}

		stdout := &strings.Builder{}
		stderr := &strings.Builder{}

		err = Run(Arguments{
			Stdout: stdout,
			Stderr: stderr,
			Path:   "file/template.tel",
			Root:   root,
			Mode:   ModeFile,
		})

		if err != nil {
			t.Fatal(err)
		}

		// TODO check that file was created
	})
	t.Run("can generate in package mode", func(t *testing.T) {
		dir, err := tests.CreateProject(packageFS)
		if err != nil {
			t.Fatal(fmt.Errorf("failed to create test dir: %w", err))
		}
		_, _ = fmt.Fprintln(os.Stdout, "Test dir:", dir)

		root, err := os.OpenRoot(dir)
		if err != nil {
			t.Fatal(fmt.Errorf("failed to open root: %w", err))
		}

		stdout := &strings.Builder{}
		stderr := &strings.Builder{}

		err = Run(Arguments{
			Stdout: stdout,
			Stderr: stderr,
			Path:   "package/demo",
			Root:   root,
			Mode:   ModePackage,
		})

		if err != nil {
			t.Fatal(err)
		}

		// TODO check that package was created
	})

	// TODO test workspace
}
