package ast

import (
	"fmt"

	"github.com/tel-lang/tel"
)

type Package struct {
	Modules  []*Module
	Name     string
	Path     string
	FullName string
	hasError bool
}

func LoadPackage(ctx tel.Context, fileset tel.FileSet, pkg *Package) {
	builtinScope := newBuiltinScope(ctx)
	scope := newScope(builtinScope)

	context := &entityContext{
		Context: ctx,
		pkg:     pkg,
		scope:   scope,
	}

	loadPackage(ctx, fileset, pkg)
	resolvePackage(context)
	typecheckPackage(context)
}

func loadPackage(ctx tel.Context, fileset tel.FileSet, pkg *Package) {
	totalFiles := len(fileset.Files)
	pkg.Modules = make([]*Module, totalFiles)
	for i, src := range fileset.Files {
		m := loadModule(ctx, src)
		pkg.Modules[i] = m

		if m.HasError() {
			pkg.hasError = true
		}
		if i > 0 && pkg.Name != m.Package.ID {
			err := fmt.Errorf("%w in %s", ErrMismatchPackageName, src.Name)
			pkg.errorVar(err, m.Package.Var)
		} else {
			pkg.Name = m.Package.ID
		}
	}
}
