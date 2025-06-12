package token_test

import (
	"fmt"
	"testing"

	"github.com/eml-lang/teml/token"
)

func TestScan_string_error(t *testing.T) {
	source := []string{
		`"foo`,
		`"foo
		`,
		"-",
	}

	expected := token.Invalid

	for i, src := range source {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			f := token.Scan([]byte(src))
			kinds := getKinds(f.Tokens())

			for _, got := range kinds {
				if expected != got {
					t.Errorf("scan succeeded unexpectedly at %d", i)
				}
			}
		})
	}
}
