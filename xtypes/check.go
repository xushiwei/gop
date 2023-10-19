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
	"errors"
	"fmt"
	"go/constant"
	"go/types"

	"github.com/goplus/gop/ast"
	"github.com/goplus/gop/token"

	. "github.com/goplus/gop/xtypes/internal/errors"
	. "github.com/goplus/gop/xtypes/internal/xtypes"
)

// debugging/development support
const debug = false // leave on during development

// exprInfo stores information about an untyped expression.
type exprInfo struct {
	isLhs bool // expression is lhs operand of a shift with delayed type-check
	mode  operandMode
	typ   *types.Basic
	val   constant.Value // constant value; or nil (if not a constant)
}

// An environment represents the environment within which an object is
// type-checked.
type environment struct {
	decl          *declInfo              // package-level declaration whose init expression/function body is checked
	scope         *types.Scope           // top-most scope for lookups
	pos           token.Pos              // if valid, identifiers are looked up as if at position pos (used by Eval)
	iota          constant.Value         // value of iota in a constant declaration; nil otherwise
	errpos        positioner             // if set, identifier position of a constant with inherited initializer
	inTParamList  bool                   // set if inside a type parameter list
	sig           *types.Signature       // function signature if inside a function; nil otherwise
	isPanic       map[*ast.CallExpr]bool // set of panic call expressions (used for termination check)
	hasLabel      bool                   // set if a function makes use of labels (only ~1% of functions); unused outside functions
	hasCallOrRecv bool                   // set if an expression contains a function call or channel receive operation
}

// An importKey identifies an imported package by import path and source directory
// (directory containing the file containing the import). In practice, the directory
// may always be the same, or may not matter. Given an (import path, directory), an
// importer must always return the same package (but given two different import paths,
// an importer may still return the same package by mapping them to the same package
// paths).
type importKey struct {
	path, dir string
}

// A dotImportKey describes a dot-imported object in the given scope.
type dotImportKey struct {
	scope *types.Scope
	name  string
}

// An action describes a (delayed) action.
type action struct {
	f    func()      // action to be executed
	desc *actionDesc // action description; may be nil, requires debug to be set
}

// If debug is set, describef sets a printf-formatted description for action a.
// Otherwise, it is a no-op.
func (a *action) describef(pos positioner, format string, args ...any) {
	if debug {
		a.desc = &actionDesc{pos, format, args}
	}
}

// An actionDesc provides information on an action.
// For debugging only.
type actionDesc struct {
	pos    positioner
	format string
	args   []any
}

// A Checker maintains the state of the type checker.
// It must be created with NewChecker.
type Checker struct {
	// package information
	// (initialized by NewChecker, valid for the life-time of checker)
	conf *Config
	ctxt *Context // context for de-duplicating instances
	fset *token.FileSet
	pkg  *types.Package
	*Info
	nextID uint64                       // unique Id for type parameters (first valid Id is 1)
	objMap map[types.Object]*declInfo   // maps package-level objects and (non-interface) methods to declaration info
	impMap map[importKey]*types.Package // maps (import path, source directory) to (complete or fake) package
	valids instanceLookup               // valid *Named (incl. instantiated) types per the validType check

	// pkgPathMap maps package names to the set of distinct import paths we've
	// seen for that name, anywhere in the import graph. It is used for
	// disambiguating package names in error messages.
	//
	// pkgPathMap is allocated lazily, so that we don't pay the price of building
	// it on the happy path. seenPkgMap tracks the packages that we've already
	// walked.
	pkgPathMap map[string]map[string]bool
	seenPkgMap map[*types.Package]bool

	// information collected during type-checking of a set of package files
	// (initialized by Files, valid only for the duration of check.Files;
	// maps and lists are allocated on demand)
	files         []*ast.File                     // package files
	imports       []*types.PkgName                // list of imported packages
	dotImportMap  map[dotImportKey]*types.PkgName // maps dot-imported objects to the package they were dot-imported through
	recvTParamMap map[*ast.Ident]*types.TypeParam // maps blank receiver type parameters to their type
	brokenAliases map[*types.TypeName]bool        // set of aliases with broken (not yet determined) types
	unionTypeSets map[*types.Union]*_TypeSet      // computed type sets for union types
	mono          monoGraph                       // graph for detecting non-monomorphizable instantiation loops

	firstErr error                             // first error encountered
	methods  map[*types.TypeName][]*types.Func // maps package scope type names to associated non-blank (non-interface) methods
	untyped  map[ast.Expr]exprInfo             // map of expressions without final type
	delayed  []action                          // stack of delayed action segments; segments are processed in FIFO order
	objPath  []types.Object                    // path of object dependencies during type inference (for cycle reporting)
	cleaners []cleaner                         // list of types that may need a final cleanup at the end of type-checking

	// environment within which the current object is type-checked (valid only
	// for the duration of type-checking a specific object)
	environment

	// debugging
	indent int // indentation for tracing
}

// brokenAlias records that alias doesn't have a determined type yet.
// It also sets alias.typ to Typ[Invalid].
func (check *Checker) brokenAlias(alias *types.TypeName) {
	if check.brokenAliases == nil {
		check.brokenAliases = make(map[*types.TypeName]bool)
	}
	check.brokenAliases[alias] = true
	TypeName_setTyp(alias, types.Typ[types.Invalid])
}

// validAlias records that alias has the valid type typ (possibly Typ[Invalid]).
func (check *Checker) validAlias(alias *types.TypeName, typ types.Type) {
	delete(check.brokenAliases, alias)
	TypeName_setTyp(alias, typ)
}

// isBrokenAlias reports whether alias doesn't have a determined type yet.
func (check *Checker) isBrokenAlias(alias *types.TypeName) bool {
	return alias.Type() == types.Typ[types.Invalid] && check.brokenAliases[alias]
}

// later pushes f on to the stack of actions that will be processed later;
// either at the end of the current statement, or in case of a local constant
// or variable declaration, before the constant or variable is in scope
// (so that f still sees the scope before any new declarations).
// later returns the pushed action so one can provide a description
// via action.describef for debugging, if desired.
func (check *Checker) later(f func()) *action {
	i := len(check.delayed)
	check.delayed = append(check.delayed, action{f: f})
	return &check.delayed[i]
}

// push pushes obj onto the object path and returns its index in the path.
func (check *Checker) push(obj types.Object) int {
	check.objPath = append(check.objPath, obj)
	return len(check.objPath) - 1
}

// pop pops and returns the topmost object from the object path.
func (check *Checker) pop() types.Object {
	i := len(check.objPath) - 1
	obj := check.objPath[i]
	check.objPath[i] = nil
	check.objPath = check.objPath[:i]
	return obj
}

type cleaner interface {
	cleanup()
}

// NewChecker returns a new Checker instance for a given package.
// Package files may be added incrementally via checker.Files.
func NewChecker(conf *Config, fset *token.FileSet, pkg *types.Package, info *Info) *Checker {
	// make sure we have a configuration
	if conf == nil {
		conf = new(Config)
	}

	// make sure we have an info struct
	if info == nil {
		info = new(Info)
	}

	// Note: clients may call NewChecker with the Unsafe package, which is
	// globally shared and must not be mutated. Therefore NewChecker must not
	// mutate *pkg.
	//
	// (previously, pkg.goVersion was mutated here: go.dev/issue/61212)

	return &Checker{
		conf:   conf,
		ctxt:   conf.Context,
		fset:   fset,
		pkg:    pkg,
		Info:   info,
		objMap: make(map[types.Object]*declInfo),
		impMap: make(map[importKey]*types.Package),
	}
}

// initFiles initializes the files-specific portion of checker.
// The provided files must all belong to the same package.
func (check *Checker) initFiles(files []*ast.File) {
	// start with a clean slate (check.Files may be called multiple times)
	check.files = nil
	check.imports = nil
	check.dotImportMap = nil

	check.firstErr = nil
	check.methods = nil
	check.untyped = nil
	check.delayed = nil
	check.objPath = nil
	check.cleaners = nil

	// determine package name and collect valid files
	pkg := check.pkg
	for _, file := range files {
		switch name := file.Name.Name; pkg.Name() {
		case "":
			if name != "_" {
				pkg.SetName(name)
			} else {
				check.error(file.Name, BlankPkgName, "invalid package name _")
			}
			fallthrough

		case name:
			check.files = append(check.files, file)

		default:
			check.errorf(atPos(file.Package), MismatchedPkgName, "package %s; expected %s", name, pkg.Name())
			// ignore this file
		}
	}
}

// A bailout panic is used for early termination.
type bailout struct{}

func (check *Checker) handleBailout(err *error) {
	switch p := recover().(type) {
	case nil, bailout:
		// normal return or early exit
		*err = check.firstErr
	default:
		// re-panic
		panic(p)
	}
}

// Files checks the provided files as part of the checker's package.
func (check *Checker) Files(files []*ast.File) error { return check.checkFiles(files) }

var errBadCgo = errors.New("cannot use FakeImportC and go115UsesCgo together")

func (check *Checker) checkFiles(files []*ast.File) (err error) {
	if check.pkg == types.Unsafe {
		// Defensive handling for Unsafe, which cannot be type checked, and must
		// not be mutated. See https://go.dev/issue/61212 for an example of where
		// Unsafe is passed to NewChecker.
		return nil
	}

	if check.conf.FakeImportC && check.conf.go115UsesCgo {
		return errBadCgo
	}

	defer check.handleBailout(&err)

	print := func(msg string) {
		if check.conf._Trace {
			fmt.Println()
			fmt.Println(msg)
		}
	}

	print("== initFiles ==")
	check.initFiles(files)

	print("== collectObjects ==")
	check.collectObjects()

	print("== packageObjects ==")
	check.packageObjects()

	print("== processDelayed ==")
	check.processDelayed(0) // incl. all functions

	print("== cleanup ==")
	check.cleanup()

	print("== initOrder ==")
	check.initOrder()

	if !check.conf.DisableUnusedImportCheck {
		print("== unusedImports ==")
		check.unusedImports()
	}

	print("== recordUntyped ==")
	check.recordUntyped()

	if check.firstErr == nil {
		// TODO(mdempsky): Ensure monomorph is safe when errors exist.
		check.monomorph()
	}

	check.pkg.MarkComplete()

	// no longer needed - release memory
	check.imports = nil
	check.dotImportMap = nil
	check.pkgPathMap = nil
	check.seenPkgMap = nil
	check.recvTParamMap = nil
	check.brokenAliases = nil
	check.unionTypeSets = nil
	check.ctxt = nil

	// TODO(rFindley) There's more memory we should release at this point.
	return
}

func (check *Checker) recordDef(id *ast.Ident, obj types.Object) {
	assert(id != nil)
	if m := check.Defs; m != nil {
		m[id] = obj
	}
}

func (check *Checker) recordUse(id *ast.Ident, obj types.Object) {
	assert(id != nil)
	assert(obj != nil)
	if m := check.Uses; m != nil {
		m[id] = obj
	}
}

func (check *Checker) recordImplicit(node ast.Node, obj types.Object) {
	assert(node != nil)
	assert(obj != nil)
	if m := check.Implicits; m != nil {
		m[node] = obj
	}
}

func (check *Checker) recordScope(node ast.Node, scope *types.Scope) {
	assert(node != nil)
	assert(scope != nil)
	if m := check.Scopes; m != nil {
		m[node] = scope
	}
}
