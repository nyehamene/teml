package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tel-lang/tel/cmd/tel/internal/cmderr"
)

func TestMain(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStdout string
		expectedStderr string
		expectedCode   int
	}{
		{
			name:           "no args prints usage",
			args:           []string{},
			expectedStderr: usageText,
			expectedCode:   cmderr.Usage,
		},
		{
			name:           `"help" prints usage`,
			args:           []string{"help"},
			expectedStdout: usageText,
			expectedCode:   0,
		},
		{
			name:           `"--help" prints usage`,
			args:           []string{"--help"},
			expectedStdout: usageText,
			expectedCode:   0,
		},
		{
			name:           `"-help" prints usage`,
			args:           []string{"-help"},
			expectedStdout: usageText,
			expectedCode:   0,
		},
		{
			name:           `"-h" prints usage`,
			args:           []string{"-h"},
			expectedStdout: usageText,
			expectedCode:   0,
		},
		{
			name:           `"version" prints version`,
			args:           []string{"version"},
			expectedStdout: version,
			expectedCode:   0,
		},
		{
			name:           `"--version" prints version`,
			args:           []string{"--version"},
			expectedStdout: version,
			expectedCode:   0,
		},
		{
			name:           `"-version" prints version`,
			args:           []string{"-version"},
			expectedStdout: version,
			expectedCode:   0,
		},
		{
			name:           `"-v" prints version`,
			args:           []string{"-v"},
			expectedStdout: version,
			expectedCode:   0,
		},
		{
			name:           `"generate --help" prints usage`,
			args:           []string{"generate", "--help"},
			expectedStdout: generateUsageText,
			expectedCode:   0,
		},
		{
			name:           `"generate -f -p" reports an error`,
			args:           []string{"generate", "-f", "foo.tel", "-p", "bar.tel"},
			expectedStderr: generateUsageText,
			expectedCode:   cmderr.Usage,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout := &strings.Builder{}
			stderr := &strings.Builder{}

			actualCode := run(stdout, stderr, test.args)

			if test.expectedCode != actualCode {
				t.Errorf("expected code %v got %v", test.expectedCode, actualCode)
			}
			if diff := cmp.Diff(test.expectedStdout, stdout.String()); diff != "" {
				t.Error(diff)
				t.Error("expected stdout:")
				t.Error(test.expectedStdout)
				t.Error("actual stdout:")
				t.Error(stdout.String())
			}
			if diff := cmp.Diff(test.expectedStderr, stderr.String()); diff != "" {
				t.Error(diff)
				t.Error("expected Stderr:")
				t.Error(test.expectedStderr)
				t.Error("actual Stderr:")
				t.Error(stderr.String())
			}
		})
	}
}
