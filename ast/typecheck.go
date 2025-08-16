package ast

import (
	"errors"
	"fmt"
)

type typechecker struct {
	context *entityContext
}

func typecheckPackage(ctx *entityContext) {
	assert(ctx != nil, "context is nil")
	assert(ctx.pkg != nil, "package is nil")
	assert(ctx.scope != nil, "scope is nil")

	t := typechecker{context: ctx}

	t.checkUsing()

	// bind property types
	for _, m := range t.context.modules {
		for _, tmpl := range m.templates {
			t.checkProperties(tmpl)
			err := m.scope.rebind(tmpl.Name, tmpl)
			if err != nil {
				t.errorVar(err, tmpl.Var)
			}
		}
	}

	// check template statement
	for _, m := range t.context.modules {
		for _, tmpl := range m.templates {
			t.checkStmts(tmpl, tmpl.decl.Stmts)
		}
	}
}

func (t *typechecker) checkUsing() {
	assert(t != nil, "typechecker is nil")

	// TBD(feat): when the kind of using.from expr is a type then using.Idents must have only one identifier (alias)
	// TBD(feat): the using.from expression must either be a module or a type (enum or template)

	// check using declaration type
	for _, m := range t.context.modules {
		for _, u := range m.decl.Usings {
			from, exists := t.resolveScope(m.scope, u.From)
			if !exists {
				// error should have already been reported in resolveScope
				continue
			}

			for _, alias := range u.Aliases {
				entity, exists := from.lookup(alias.Name)
				if !exists {
					t.errorVar(ErrUndeclared, alias)
					continue
				}
				err := m.scope.bind(alias.Name, entity)
				if err != nil {
					t.errorVar(err, alias)
				}
			}
		}
	}
}

func (t *typechecker) checkProperties(entity *TypeTemplate) {
	assert(t != nil, "typechecker is nil")
	assert(entity != nil, "template is nil")
	assert(entity.scope != nil, "scope is nil")

	for _, field := range entity.fields {
		fieldType, err := t.resolveType(entity.scope, field.decl.Type)
		if err != nil {
			t.errorVar(err, field.Var)
			continue
		}

		field.Type = fieldType

		// NOTE(self): entity names have already been bound in the resolver, therefore, rebind is used to update it
		// NOTE(self): no need to rebind since the field pointer is update directly above
		// if err := entity.scope.rebind(field.Name, field); err != nil {
		// 	t.errorVar(err, field.Var)
		// }

		switch typeKind := fieldType.typeKind(); typeKind {
		case KindBool,
			KindNumber,
			KindEnum,
			KindString:
			// okay

		case KindHTML,
			KindModule,
			KindPackage:

			t.errorVar(ErrInvalidFieldType, field.Var)

		case KindField:
			// unreachable
			panic("unreachable")

		case KindTemplate:
			typeid := fieldType.(*TypeTemplate)
			if typeid.Kind == TemplateDocument {
				t.errorVar(ErrInvalidFieldType, field.Var)
			}
			if typesEqual(typeid, entity) {
				t.errorVar(ErrRecursiveDefinition, field.Var)
			}

		case KindInvalid:
			// TBD(fix): return proper error
			t.errorVar(ErrGeneric, field.Var)

		default:
			panic(fmt.Sprintf("unexpected ast.TypeKind: %#v", typeKind))
		}
	}
}

func (t *typechecker) checkStmts(tm Entity, stmts []Stmt) {
	assert(t != nil, "typechecker is nil")
	assert(tm != nil, "template entity is nil")

	for i, stmt := range stmts {
		t.checkStmt(tm, &stmt)
		stmts[i] = stmt
	}
}

func (t *typechecker) checkStmt(entity Entity, stmt *Stmt) {
	entityScope, exists := getScope(entity)
	if !exists {
		t.error(ErrNotObject)
		return
	}

	switch node := (*stmt).(type) {
	case ComponentElement:
		entity := t.checkElementTag(entityScope, node.Tag)
		typeid := mustCastType[*TypeTemplate](entity)
		t.checkParameters(typeid, node.Parameters)
		t.checkAttributes(entityScope, node.Attributes)

	case Cond:
		// TBD(feat): get the location of the expression in the source
		location := Var{}

		condTypeID, err := t.resolveType(entityScope, node.Target)
		if err != nil {
			t.errorVar(err, location)
			break
		}
		if condTypeID.typeKind() != KindEnum {
			t.errorVar(ErrNotEnum, location)
		}
		for i, c := range node.Cases {
			brandTypeID, err := t.resolveType(entityScope, c.Cond)
			if err != nil {
				t.errorVar(err, location)
			}
			if !typesEqual(condTypeID, brandTypeID) {
				t.errorVar(ErrMismatchType, location)
			}
			t.checkStmt(entity, &c.Branch)
			node.Cases[i] = c
		}

	case Element:
		element, transformed := t.transformElement(entityScope, node)
		if !transformed {
			// NOTE(): an error should have already been reported in transformElement()
			return
		}
		t.checkStmt(entity, &element)
		*stmt = element

	case If:
		// TBD(feat): get the location of the expression in the source
		location := Var{}

		ctypeid, err := t.resolveType(entityScope, node.Cond)
		if err != nil {
			t.errorVar(err, location)
			break
		}
		if ctypeid.typeKind() != KindBool {
			t.errorVar(ErrNotBool, location)
		}
		if node.Else != nil {
			t.checkStmt(entity, &node.Else)
		}
		t.checkStmt(entity, &node.Then)

	case HTMLElement:
		entity := t.checkElementTag(entityScope, node.Tag)
		typeid := mustCastType[*HTMLElementType](entity)
		t.checkAttributes(entityScope, node.Attributes)
		t.checkStmts(typeid, node.Children)

	case NumberElement:
		t.checkElementTag(entityScope, node.Tag)
		t.checkAttributes(entityScope, node.Attributes)

	case PropertyElement:
		entity := t.checkElementTag(entityScope, node.Tag)
		typeid := mustCastType[*TypeField](entity)
		t.checkAttributes(entityScope, node.Attributes)
		t.checkStmts(typeid, node.Children)

	case StringElement:
		t.checkElementTag(entityScope, node.Tag)
		t.checkAttributes(entityScope, node.Attributes)

	case Text:
		// TBD(feat): check string template

	case TextGroup:
		// TBD(feat): check string template

	default:
		panic(fmt.Sprintf("unexpected ast.Stmt: %#v", node))
	}
}

func (t *typechecker) checkElementTag(scope *Scope, tag Var) Entity {
	entity, exists := scope.lookup(tag.Name)
	if !exists {
		t.errorVar(ErrUndeclared, tag)
	}

	field := false
	tagEntity := entity

check:
	switch typeid := tagEntity.(type) {
	case *TypeNumber, *TypeString:
		if !field {
			t.errorVar(ErrInvalidElementTag, tag)
		}

	case *TypeBool,
		*TypeEnum,
		*TypeModule,
		*TypePackage:

		t.errorVar(ErrInvalidElementTag, tag)

	case *HTMLElementType:
		// okay

	case *TypeField:
		tagEntity = typeid.Type
		field = true
		goto check

	case *TypeTemplate:
		if typeid.Kind == TemplateDocument {
			t.errorVar(ErrInvalidElementTag, tag)
		}

	default:
		panic(fmt.Sprintf("unexpected ast.Entity: %#v", typeid))
	}

	return entity
}

func (t *typechecker) checkAttributes(scope *Scope, attrs []Attr) {
	// TBD(feat): check entity attributes
	// NOTE attributes key and value must be know at compile time
}

func (t *typechecker) checkParameters(entity *TypeTemplate, nodes []ParameterDecl) {
	assert(entity != nil, "entity is nil")

	parameters := make(map[string]*TypeField, len(nodes))

	for _, node := range nodes {
		_, exists := entity.scope.lookupCurrent(node.Name)
		if !exists {
			t.errorVar(ErrUndeclared, node.Var)
			continue
		}

		typeid, err := t.resolveType(entity.scope, node.Value)
		if err != nil {
			t.errorVar(err, node.Var)
		}

		parameter := &TypeField{
			Var:  node.Var,
			Type: typeid,
		}
		parameters[node.Name] = parameter
	}

	for _, field := range entity.fields {
		parameter, ok := parameters[field.Name]
		if !ok {
			t.errorVar(ErrMissingParameter, field.Var)
			continue
		}
		if !typesEqual(field.Type, parameter.Type) {
			t.errorVar(ErrMismatchType, parameter.Var)
		}
	}
}

func (t *typechecker) checkExpr(scope *Scope, expr Expr) (Entity, bool) {
	switch node := expr.(type) {
	case Number:
		return builtinNumber, true

	case Bool:
		return builtinBool, true

	case String:
		return builtinString, true

	case IfExpr:
		panic(errors.ErrUnsupported)

	case CondExpr:
		panic(errors.ErrUnsupported)

	case StringTemplate:
		panic(errors.ErrUnsupported)

	case Var:
		entity, exists := scope.lookup(node.Name)
		if !exists {
			t.errorVar(ErrUndeclared, node)
			return nil, false
		}
		return entity, true

	case MemberAccess:
		objectScope, exists := t.resolveScope(scope, node.Object)
		if !exists {
			t.error(fmt.Errorf("%w: %v", ErrUndeclaredNamespace, node))
			return nil, false
		}
		return t.checkExpr(objectScope, node.Member)

	default:
		panic(fmt.Sprintf("unexpected ast.Expr: %#v", node))
	}
}

func (t *typechecker) resolveScope(scope *Scope, expr Expr) (*Scope, bool) {
	entity, exists := t.checkExpr(scope, expr)
	if !exists {
		return nil, false
	}
	return getScope(entity)
}

func (t *typechecker) resolveType(scope *Scope, expr Expr) (Type, error) {
	entity, exists := t.checkExpr(scope, expr)
	if !exists {
		return nil, ErrUndeclared
	}

	switch typeid := entity.(type) {
	case *HTMLElementType:
		return typeid, nil
	case *TypeBool:
		return typeid, nil
	case *TypeEnum:
		return typeid, nil
	case *TypeField:
		if typeid.Type == nil {
			return nil, ErrGeneric
		}
		return typeid.Type, nil
	case *TypeModule:
		return typeid, nil
	case *TypeNumber:
		return typeid, nil
	case *TypePackage:
		return typeid, nil
	case *TypeString:
		return typeid, nil
	case *TypeTemplate:
		return typeid, nil
	default:
		panic(fmt.Sprintf("unexpected ast.Entity: %#v", typeid))
	}
}

func (t *typechecker) asVar(_ *Scope, expr Expr) (Var, bool) {
	switch node := expr.(type) {
	case Var:
		return node, true
	case MemberAccess:
		return node.Member, true
	case Number, Bool, String, IfExpr, CondExpr, StringTemplate:
		return Var{}, false
	default:
		panic(fmt.Sprintf("unexpected ast.Expr: %#v", node))
	}
}
