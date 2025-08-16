package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tel-lang/tel/cmd/tel/generatecmd"
	"github.com/tel-lang/tel/cmd/tel/internal/cmderr"
)

//go:embed usage.txt
var usageText string

//go:embed generatecmd/usage.txt
var generateUsageText string

const version = "0.0.1-alpha"

func main() {
	stdout := os.Stdout
	stderr := os.Stderr

	if len(os.Args) == 0 {
		mustWriteString(stderr, usageText)
	}

	args := os.Args[1:]
	code := run(stdout, stderr, args)
	if code != 0 {
		os.Exit(code)
	}
}

func run(stdout, stderr io.Writer, args []string) (code int) {
	if len(args) == 0 {
		mustWriteString(stderr, usageText)
		return cmderr.Usage
	}
	cmd := args[0]
	switch cmd {
	case "help", "--help", "-help", "-h":
		mustWriteString(stdout, usageText)
		return 0
	case "version", "--version", "-version", "-v":
		mustWriteString(stdout, version)
		return 0
	case "generate":
		return generateCmd(stdout, stderr, args[1:])
	}
	mustWriteString(stderr, usageText)
	return cmderr.Usage
}

func generateCmd(stdout, stderr io.Writer, args []string) (code int) {
	var helpFlag bool
	var packageFlag, fileFlag string

	f := flag.NewFlagSet("generate", flag.ExitOnError)
	f.BoolVar(&helpFlag, "help", false, "")
	f.StringVar(&packageFlag, "p", "", "")
	f.StringVar(&fileFlag, "f", "", "")

	if err := f.Parse(args); err != nil {
		mustWriteString(stderr, generateUsageText)
		return cmderr.Usage
	}

	if helpFlag {
		mustWriteString(stdout, generateUsageText)
		return
	}

	var (
		mode generatecmd.Mode
		path string
	)

	// get the selected mode
	{
		flags := map[generatecmd.Mode]string{
			generatecmd.ModeFile:    fileFlag,
			generatecmd.ModePackage: packageFlag,
		}
		count := 0
		for m, f := range flags {
			if f == "" {
				continue
			}
			count += 1
			mode = m
		}
		// defaults to processing a package
		if count == 0 {
			mode = generatecmd.ModePackage
		} else if count > 1 {
			mustWriteString(stderr, generateUsageText)
			return cmderr.Usage
		}
	}

	root := mustOpenRoot()
	cmd := generatecmd.Arguments{
		Stdout: stdout,
		Stderr: stderr,
		Path:   path,
		Root:   root,
		Mode:   mode,
	}

	err := generatecmd.Run(cmd)
	if err != nil {
		return cmderr.Failed
	}

	return
}

func mustOpenRoot() *os.Root {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		panic(err)
	}
	return root
}

func mustWriteString(w io.Writer, s string) {
	_, err := fmt.Fprint(w, s)
	if err != nil {
		panic(err)
	}
}
