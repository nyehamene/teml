package token_test

import (
	"fmt"
	"testing"

	"github.com/tel-lang/tel"
	"github.com/tel-lang/tel/ast/token"
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
			f := token.Scan([]byte(src), "test.tel", tel.PreserveComment)
			kinds := getKinds(f.Tokens)

			for _, got := range kinds {
				if expected != got {
					t.Errorf("expected %s but got %s at %d", expected, got, i)
				}
			}
		})
	}
}
