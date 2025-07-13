package ast_test

import (
	"embed"
	"testing"
)

//go:embed testdata/resolver
var resolverBaseFS embed.FS

// resolverBasepath the path to the folder containing test source path
var resolverBasepath = "testdata/resolver"

func TestResolveValid(t *testing.T) {
	checkValidFiles(t, resolverBaseFS, resolverBasepath, ResolutionStage)
}

func TestResolveInvalid(t *testing.T) {
	checkInvalidFiles(t, resolverBaseFS, resolverBasepath, ResolutionStage)
}
