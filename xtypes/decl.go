/*
 * Copyright (c) 2023 The GoPlus Authors (goplus.org). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package xtypes

import (
	"fmt"
	"go/constant"
	"go/types"

	"github.com/goplus/gop/ast"
	"github.com/goplus/gop/token"
	. "github.com/goplus/gop/xtypes/internal/errors"
	. "github.com/goplus/gop/xtypes/internal/xtypes"
)

func (check *Checker) reportAltDecl(obj types.Object) {
	if pos := obj.Pos(); pos.IsValid() {
		// We use "other" rather than "previous" here because
		// the first declaration seen may not be textually
		// earlier in the source.
		check.errorf(obj, DuplicateDecl, "\tother declaration of %s", obj.Name()) // secondary error, \t indented
	}
}

func (check *Checker) declare(scope *types.Scope, id *ast.Ident, obj types.Object, pos token.Pos) {
	// spec: "The blank identifier, represented by the underscore
	// character _, may be used in a declaration like any other
	// identifier but the declaration does not introduce a new
	// binding."
	if obj.Name() != "_" {
		if alt := scope.Insert(obj); alt != nil {
			check.errorf(obj, DuplicateDecl, "%s redeclared in this block", obj.Name())
			check.reportAltDecl(alt)
			return
		}
		Object_setScopePos(obj, pos)
	}
	if id != nil {
		check.recordDef(id, obj)
	}
}

// pathString returns a string of the form a->b-> ... ->g for a path [a, b, ... g].
func pathString(path []types.Object) string {
	var s string
	for i, p := range path {
		if i > 0 {
			s += "->"
		}
		s += p.Name()
	}
	return s
}

// objDecl type-checks the declaration of obj in its respective (file) environment.
// For the meaning of def, see Checker.definedType, in typexpr.go.
func (check *Checker) objDecl(obj types.Object, def *types.Named) {
	if check.conf._Trace && obj.Type() == nil {
		if check.indent == 0 {
			fmt.Println() // empty line between top-level objects for readability
		}
		check.trace(obj.Pos(), "-- checking %s (%s, objPath = %s)", obj, Object_color(obj), pathString(check.objPath))
		check.indent++
		defer func() {
			check.indent--
			check.trace(obj.Pos(), "=> %s (%s)", obj, Object_color(obj))
		}()
	}

	// Checking the declaration of obj means inferring its type
	// (and possibly its value, for constants).
	// An object's type (and thus the object) may be in one of
	// three states which are expressed by colors:
	//
	// - an object whose type is not yet known is painted white (initial color)
	// - an object whose type is in the process of being inferred is painted grey
	// - an object whose type is fully inferred is painted black
	//
	// During type inference, an object's color changes from white to grey
	// to black (pre-declared objects are painted black from the start).
	// A black object (i.e., its type) can only depend on (refer to) other black
	// ones. White and grey objects may depend on white and black objects.
	// A dependency on a grey object indicates a cycle which may or may not be
	// valid.
	//
	// When objects turn grey, they are pushed on the object path (a stack);
	// they are popped again when they turn black. Thus, if a grey object (a
	// cycle) is encountered, it is on the object path, and all the objects
	// it depends on are the remaining objects on that path. Color encoding
	// is such that the color value of a grey object indicates the index of
	// that object in the object path.

	// During type-checking, white objects may be assigned a type without
	// traversing through objDecl; e.g., when initializing constants and
	// variables. Update the colors of those objects here (rather than
	// everywhere where we set the type) to satisfy the color invariants.
	if Object_color(obj) == White && obj.Type() != nil {
		Object_setColor(obj, Black)
		return
	}

	switch Object_color(obj) {
	case White:
		assert(obj.Type() == nil)
		// All color values other than white and black are considered grey.
		// Because black and white are < grey, all values >= grey are grey.
		// Use those values to encode the object's index into the object path.
		Object_setColor(obj, Grey+Color(check.push(obj)))
		defer func() {
			Object_setColor(check.pop(), Black)
		}()

	case Black:
		assert(obj.Type() != nil)
		return

	default:
		// Color values other than white or black are considered grey.
		fallthrough

	case Grey:
		// We have a (possibly invalid) cycle.
		// In the existing code, this is marked by a non-nil type
		// for the object except for constants and variables whose
		// type may be non-nil (known), or nil if it depends on the
		// not-yet known initialization value.
		// In the former case, set the type to Typ[Invalid] because
		// we have an initialization cycle. The cycle error will be
		// reported later, when determining initialization order.
		// TODO(gri) Report cycle here and simplify initialization
		// order code.
		switch obj := obj.(type) {
		case *types.Const:
			if !check.validCycle(obj) || obj.Type() == nil {
				Const_setTyp(obj, types.Typ[types.Invalid])
			}

		case *types.Var:
			if !check.validCycle(obj) || obj.Type() == nil {
				Var_setTyp(obj, types.Typ[types.Invalid])
			}

		case *types.TypeName:
			if !check.validCycle(obj) {
				// break cycle
				// (without this, calling underlying()
				// below may lead to an endless loop
				// if we have a cycle for a defined
				// (*Named) type)
				TypeName_setTyp(obj, types.Typ[types.Invalid])
			}

		case *types.Func:
			if !check.validCycle(obj) {
				// Don't set obj.typ to Typ[Invalid] here
				// because plenty of code type-asserts that
				// functions have a *Signature type. Grey
				// functions have their type set to an empty
				// signature which makes it impossible to
				// initialize a variable with the function.
			}

		default:
			panic("unreachable")
		}
		assert(obj.Type() != nil)
		return
	}

	d := check.objMap[obj]
	if d == nil {
		check.dump("%v: %s should have been declared", obj.Pos(), obj)
		panic("unreachable")
	}

	// save/restore current environment and set up object environment
	defer func(env environment) {
		check.environment = env
	}(check.environment)
	check.environment = environment{
		scope: d.file,
	}

	// Const and var declarations must not have initialization
	// cycles. We track them by remembering the current declaration
	// in check.decl. Initialization expressions depending on other
	// consts, vars, or functions, add dependencies to the current
	// check.decl.
	switch obj := obj.(type) {
	case *types.Const:
		check.decl = d // new package-level const decl
		check.constDecl(obj, d.vtyp, d.init, d.inherited)
	case *types.Var:
		check.decl = d // new package-level var decl
		check.varDecl(obj, d.lhs, d.vtyp, d.init)
	case *types.TypeName:
		// invalid recursive types are detected via path
		check.typeDecl(obj, d.tdecl, def)
		check.collectMethods(obj) // methods can only be added to top-level types
	case *types.Func:
		// functions may be recursive - no need to track dependencies
		check.funcDecl(obj, d)
	default:
		panic("unreachable")
	}
}

// validCycle checks if the cycle starting with obj is valid and
// reports an error if it is not.
func (check *Checker) validCycle(obj types.Object) (valid bool) {
	// The object map contains the package scope objects and the non-interface methods.
	if debug {
		info := check.objMap[obj]
		inObjMap := info != nil && (info.fdecl == nil || info.fdecl.Recv == nil) // exclude methods
		isPkgObj := obj.Parent() == Package_scope(check.pkg)
		if isPkgObj != inObjMap {
			check.dump("%v: inconsistent object map for %s (isPkgObj = %v, inObjMap = %v)", obj.Pos(), obj, isPkgObj, inObjMap)
			panic("unreachable")
		}
	}

	// Count cycle objects.
	assert(Object_color(obj) >= Grey)
	start := Object_color(obj) - Grey // index of obj in objPath
	cycle := check.objPath[start:]
	tparCycle := false // if set, the cycle is through a type parameter list
	nval := 0          // number of (constant or variable) values in the cycle; valid if !generic
	ndef := 0          // number of type definitions in the cycle; valid if !generic
loop:
	for _, obj := range cycle {
		switch obj := obj.(type) {
		case *types.Const, *types.Var:
			nval++
		case *types.TypeName:
			// If we reach a generic type that is part of a cycle
			// and we are in a type parameter list, we have a cycle
			// through a type parameter list, which is invalid.
			if check.inTParamList && IsGeneric(obj.Type()) {
				tparCycle = true
				break loop
			}

			// Determine if the type name is an alias or not. For
			// package-level objects, use the object map which
			// provides syntactic information (which doesn't rely
			// on the order in which the objects are set up). For
			// local objects, we can rely on the order, so use
			// the object's predicate.
			// TODO(gri) It would be less fragile to always access
			// the syntactic information. We should consider storing
			// this information explicitly in the object.
			var alias bool
			if d := check.objMap[obj]; d != nil {
				alias = d.tdecl.Assign.IsValid() // package-level object
			} else {
				alias = obj.IsAlias() // function local object
			}
			if !alias {
				ndef++
			}
		case *types.Func:
			// ignored for now
		default:
			panic("unreachable")
		}
	}

	if check.conf._Trace {
		check.trace(obj.Pos(), "## cycle detected: objPath = %s->%s (len = %d)", pathString(cycle), obj.Name(), len(cycle))
		if tparCycle {
			check.trace(obj.Pos(), "## cycle contains: generic type in a type parameter list")
		} else {
			check.trace(obj.Pos(), "## cycle contains: %d values, %d type definitions", nval, ndef)
		}
		defer func() {
			if valid {
				check.trace(obj.Pos(), "=> cycle is valid")
			} else {
				check.trace(obj.Pos(), "=> error: cycle is invalid")
			}
		}()
	}

	if !tparCycle {
		// A cycle involving only constants and variables is invalid but we
		// ignore them here because they are reported via the initialization
		// cycle check.
		if nval == len(cycle) {
			return true
		}

		// A cycle involving only types (and possibly functions) must have at least
		// one type definition to be permitted: If there is no type definition, we
		// have a sequence of alias type names which will expand ad infinitum.
		if nval == 0 && ndef > 0 {
			return true
		}
	}

	check.cycleError(cycle)
	return false
}

// cycleError reports a declaration cycle starting with
// the object in cycle that is "first" in the source.
func (check *Checker) cycleError(cycle []types.Object) {
	// name returns the (possibly qualified) object name.
	// This is needed because with generic types, cycles
	// may refer to imported types. See go.dev/issue/50788.
	// TODO(gri) Thus functionality is used elsewhere. Factor it out.
	name := func(obj types.Object) string {
		// see types.packagePrefix
		return PackagePrefix(obj.Pkg(), check.qualifier) + obj.Name()
	}

	// TODO(gri) Should we start with the last (rather than the first) object in the cycle
	//           since that is the earliest point in the source where we start seeing the
	//           cycle? That would be more consistent with other error messages.
	i := firstInSrc(cycle)
	obj := cycle[i]
	objName := name(obj)
	// If obj is a type alias, mark it as valid (not broken) in order to avoid follow-on errors.
	tname, _ := obj.(*types.TypeName)
	if tname != nil && tname.IsAlias() {
		check.validAlias(tname, types.Typ[types.Invalid])
	}

	// report a more concise error for self references
	if len(cycle) == 1 {
		if tname != nil {
			check.errorf(obj, InvalidDeclCycle, "invalid recursive type: %s refers to itself", objName)
		} else {
			check.errorf(obj, InvalidDeclCycle, "invalid cycle in declaration: %s refers to itself", objName)
		}
		return
	}

	if tname != nil {
		check.errorf(obj, InvalidDeclCycle, "invalid recursive type %s", objName)
	} else {
		check.errorf(obj, InvalidDeclCycle, "invalid cycle in declaration of %s", objName)
	}
	for range cycle {
		check.errorf(obj, InvalidDeclCycle, "\t%s refers to", objName) // secondary error, \t indented
		i++
		if i >= len(cycle) {
			i = 0
		}
		obj = cycle[i]
		objName = name(obj)
	}
	check.errorf(obj, InvalidDeclCycle, "\t%s", objName)
}

// firstInSrc reports the index of the object with the "smallest"
// source position in path. path must not be empty.
func firstInSrc(path []types.Object) int {
	fst, pos := 0, path[0].Pos()
	for i, t := range path[1:] {
		if cmpPos(t.Pos(), pos) < 0 {
			fst, pos = i+1, t.Pos()
		}
	}
	return fst
}

type (
	decl interface {
		node() ast.Node
	}

	importDecl struct{ spec *ast.ImportSpec }
	constDecl  struct {
		spec      *ast.ValueSpec
		iota      int
		typ       ast.Expr
		init      []ast.Expr
		inherited bool
	}
	varDecl  struct{ spec *ast.ValueSpec }
	typeDecl struct{ spec *ast.TypeSpec }
	funcDecl struct{ decl *ast.FuncDecl }
)

func (d importDecl) node() ast.Node { return d.spec }
func (d constDecl) node() ast.Node  { return d.spec }
func (d varDecl) node() ast.Node    { return d.spec }
func (d typeDecl) node() ast.Node   { return d.spec }
func (d funcDecl) node() ast.Node   { return d.decl }

func (check *Checker) walkDecls(decls []ast.Decl, f func(decl)) {
	for _, d := range decls {
		check.walkDecl(d, f)
	}
}

func (check *Checker) walkDecl(d ast.Decl, f func(decl)) {
	switch d := d.(type) {
	case *ast.BadDecl:
		// ignore
	case *ast.GenDecl:
		var last *ast.ValueSpec // last ValueSpec with type or init exprs seen
		for iota, s := range d.Specs {
			switch s := s.(type) {
			case *ast.ImportSpec:
				f(importDecl{s})
			case *ast.ValueSpec:
				switch d.Tok {
				case token.CONST:
					// determine which initialization expressions to use
					inherited := true
					switch {
					case s.Type != nil || len(s.Values) > 0:
						last = s
						inherited = false
					case last == nil:
						last = new(ast.ValueSpec) // make sure last exists
						inherited = false
					}
					check.arityMatch(s, last)
					f(constDecl{spec: s, iota: iota, typ: last.Type, init: last.Values, inherited: inherited})
				case token.VAR:
					check.arityMatch(s, nil)
					f(varDecl{s})
				default:
					check.errorf(s, InvalidSyntaxTree, "invalid token %s", d.Tok)
				}
			case *ast.TypeSpec:
				f(typeDecl{s})
			default:
				check.errorf(s, InvalidSyntaxTree, "unknown ast.Spec node %T", s)
			}
		}
	case *ast.FuncDecl:
		f(funcDecl{d})
	default:
		check.errorf(d, InvalidSyntaxTree, "unknown ast.Decl node %T", d)
	}
}

func (check *Checker) constDecl(obj *types.Const, typ, init ast.Expr, inherited bool) {
	assert(obj.Type() == nil)

	// use the correct value of iota
	defer func(iota constant.Value, errpos positioner) {
		check.iota = iota
		check.errpos = errpos
	}(check.iota, check.errpos)
	check.iota = obj.Val()
	check.errpos = nil

	// provide valid constant value under all circumstances
	Const_setVal(obj, constant.MakeUnknown())

	// determine type, if any
	if typ != nil {
		t := check.typ(typ)
		if !IsConstType(t) {
			// don't report an error if the type is an invalid C (defined) type
			// (go.dev/issue/22090)
			if Under(t) != types.Typ[types.Invalid] {
				check.errorf(typ, InvalidConstType, "invalid constant type %s", t)
			}
			Const_setTyp(obj, types.Typ[types.Invalid])
			return
		}
		Const_setTyp(obj, t)
	}

	// check initialization
	var x operand
	if init != nil {
		if inherited {
			// The initialization expression is inherited from a previous
			// constant declaration, and (error) positions refer to that
			// expression and not the current constant declaration. Use
			// the constant identifier position for any errors during
			// init expression evaluation since that is all we have
			// (see issues go.dev/issue/42991, go.dev/issue/42992).
			check.errpos = atPos(obj.Pos())
		}
		check.expr(nil, &x, init)
	}
	check.initConst(obj, &x)
}

func (check *Checker) varDecl(obj *types.Var, lhs []*types.Var, typ, init ast.Expr) {
	assert(obj.Type() == nil)

	// determine type, if any
	if typ != nil {
		Var_setTyp(obj, check.varType(typ))
		// We cannot spread the type to all lhs variables if there
		// are more than one since that would mark them as checked
		// (see Checker.objDecl) and the assignment of init exprs,
		// if any, would not be checked.
		//
		// TODO(gri) If we have no init expr, we should distribute
		// a given type otherwise we need to re-evalate the type
		// expr for each lhs variable, leading to duplicate work.
	}

	// check initialization
	if init == nil {
		if typ == nil {
			// error reported before by arityMatch
			obj.typ = Typ[Invalid]
		}
		return
	}

	if lhs == nil || len(lhs) == 1 {
		assert(lhs == nil || lhs[0] == obj)
		var x operand
		check.expr(obj.typ, &x, init)
		check.initVar(obj, &x, "variable declaration")
		return
	}

	if debug {
		// obj must be one of lhs
		found := false
		for _, lhs := range lhs {
			if obj == lhs {
				found = true
				break
			}
		}
		if !found {
			panic("inconsistent lhs")
		}
	}

	// We have multiple variables on the lhs and one init expr.
	// Make sure all variables have been given the same type if
	// one was specified, otherwise they assume the type of the
	// init expression values (was go.dev/issue/15755).
	if typ != nil {
		for _, lhs := range lhs {
			lhs.typ = obj.typ
		}
	}

	check.initVars(lhs, []ast.Expr{init}, nil)
}

func (check *Checker) collectMethods(obj *types.TypeName) {
	// get associated methods
	// (Checker.collectObjects only collects methods with non-blank names;
	// Checker.resolveBaseTypeName ensures that obj is not an alias name
	// if it has attached methods.)
	methods := check.methods[obj]
	if methods == nil {
		return
	}
	delete(check.methods, obj)
	assert(!check.objMap[obj].tdecl.Assign.IsValid()) // don't use TypeName.IsAlias (requires fully set up object)

	// use an objset to check for name conflicts
	var mset objset

	// spec: "If the base type is a struct type, the non-blank method
	// and field names must be distinct."
	base, _ := obj.Type().(*types.Named) // shouldn't fail but be conservative
	if base != nil {
		assert(base.TypeArgs().Len() == 0) // collectMethods should not be called on an instantiated type

		// See go.dev/issue/52529: we must delay the expansion of underlying here, as
		// base may not be fully set-up.
		check.later(func() {
			check.checkFieldUniqueness(base)
		}).describef(obj, "verifying field uniqueness for %v", base)

		// Checker.Files may be called multiple times; additional package files
		// may add methods to already type-checked types. Add pre-existing methods
		// so that we can detect redeclarations.
		for i := 0; i < base.NumMethods(); i++ {
			m := base.Method(i)
			assert(m.Name() != "_")
			assert(mset.insert(m) == nil)
		}
	}

	// add valid methods
	for _, m := range methods {
		// spec: "For a base type, the non-blank names of methods bound
		// to it must be unique."
		assert(m.Name() != "_")
		if alt := mset.insert(m); alt != nil {
			if alt.Pos().IsValid() {
				check.errorf(m, DuplicateMethod, "method %s.%s already declared at %s", obj.Name(), m.Name(), alt.Pos())
			} else {
				check.errorf(m, DuplicateMethod, "method %s.%s already declared", obj.Name(), m.Name())
			}
			continue
		}

		if base != nil {
			base.AddMethod(m)
		}
	}
}

func (check *Checker) checkFieldUniqueness(base *types.Named) {
	if t, _ := Named_under(base).(*types.Struct); t != nil {
		var mset objset
		for i := 0; i < base.NumMethods(); i++ {
			m := base.Method(i)
			assert(m.Name() != "_")
			assert(mset.insert(m) == nil)
		}

		// Check that any non-blank field names of base are distinct from its
		// method names.
		for _, fld := range Struct_fields(t) {
			if fld.Name() != "_" {
				if alt := mset.insert(fld); alt != nil {
					// Struct fields should already be unique, so we should only
					// encounter an alternate via collision with a method name.
					_ = alt.(*types.Func)

					// For historical consistency, we report the primary error on the
					// method, and the alt decl on the field.
					check.errorf(alt, DuplicateFieldAndMethod, "field and method with the same name %s", fld.Name())
					check.reportAltDecl(fld)
				}
			}
		}
	}
}
