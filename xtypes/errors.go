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
	"go/types"
	"runtime"
	"strconv"
	"strings"

	"github.com/goplus/gop/ast"
	"github.com/goplus/gop/token"
	. "github.com/goplus/gop/xtypes/internal/errors"
)

func assert(p bool) {
	if !p {
		msg := "assertion failed"
		// Include information about the assertion location. Due to panic recovery,
		// this location is otherwise buried in the middle of the panicking stack.
		if _, file, line, ok := runtime.Caller(1); ok {
			msg = fmt.Sprintf("%s:%d: %s", file, line, msg)
		}
		panic(msg)
	}
}

// An error_ represents a type-checking error.
// To report an error_, call Checker.report.
type error_ struct {
	desc []errorDesc
	code Code
	soft bool // TODO(gri) eventually determine this from an error code
}

func (err *error_) empty() bool {
	return err.desc == nil
}

func (err *error_) pos() token.Pos {
	if err.empty() {
		return token.NoPos
	}
	return err.desc[0].posn.Pos()
}

func (err *error_) msg(fset *token.FileSet, qf types.Qualifier) string {
	if err.empty() {
		return "no error"
	}
	var buf strings.Builder
	for i := range err.desc {
		p := &err.desc[i]
		if i > 0 {
			fmt.Fprint(&buf, "\n\t")
			if p.posn.Pos().IsValid() {
				fmt.Fprintf(&buf, "%s: ", fset.Position(p.posn.Pos()))
			}
		}
		buf.WriteString(sprintf(fset, qf, false, p.format, p.args...))
	}
	return buf.String()
}

// String is for testing.
func (err *error_) String() string {
	if err.empty() {
		return "no error"
	}
	return fmt.Sprintf("%d: %s", err.pos(), err.msg(nil, nil))
}

// errorf adds formatted error information to err.
// It may be called multiple times to provide additional information.
func (err *error_) errorf(at token.Pos, format string, args ...interface{}) {
	err.desc = append(err.desc, errorDesc{atPos(at), format, args})
}

func (check *Checker) qualifier(pkg *types.Package) string {
	// Qualify the package unless it's the package being type-checked.
	if pkg != check.pkg {
		if check.pkgPathMap == nil {
			check.pkgPathMap = make(map[string]map[string]bool)
			check.seenPkgMap = make(map[*types.Package]bool)
			check.markImports(check.pkg)
		}
		// If the same package name was used by multiple packages, display the full path.
		if len(check.pkgPathMap[pkg.Name()]) > 1 {
			return strconv.Quote(pkg.Path())
		}
		return pkg.Name()
	}
	return ""
}

// markImports recursively walks pkg and its imports, to record unique import
// paths in pkgPathMap.
func (check *Checker) markImports(pkg *types.Package) {
	if check.seenPkgMap[pkg] {
		return
	}
	check.seenPkgMap[pkg] = true

	forName, ok := check.pkgPathMap[pkg.Name()]
	if !ok {
		forName = make(map[string]bool)
		check.pkgPathMap[pkg.Name()] = forName
	}
	forName[pkg.Path()] = true

	for _, imp := range pkg.Imports() {
		check.markImports(imp)
	}
}

// An errorDesc describes part of a type-checking error.
type errorDesc struct {
	posn   positioner
	format string
	args   []interface{}
}

func sprintf(fset *token.FileSet, qf types.Qualifier, tpSubscripts bool, format string, args ...any) string {
	return fmt.Sprintf(format, args...) // TODO: see types.sprintf
}

func (check *Checker) trace(pos token.Pos, format string, args ...any) {
	fmt.Printf("%s:\t%s%s\n",
		check.fset.Position(pos),
		strings.Repeat(".  ", check.indent),
		sprintf(check.fset, check.qualifier, true, format, args...),
	)
}

// dump is only needed for debugging
func (check *Checker) dump(format string, args ...any) {
	fmt.Println(sprintf(check.fset, check.qualifier, true, format, args...))
}

// Report records the error pointed to by errp, setting check.firstError if
// necessary.
func (check *Checker) report(errp *error_) {
	if errp.empty() {
		panic("empty error details")
	}

	msg := errp.msg(check.fset, check.qualifier)
	switch errp.code {
	case InvalidSyntaxTree:
		msg = "invalid AST: " + msg
	case 0:
		panic("no error code provided")
	}

	// If we have an URL for error codes, add a link to the first line.
	if errp.code != 0 && check.conf._ErrorURL != "" {
		u := fmt.Sprintf(check.conf._ErrorURL, errp.code)
		if i := strings.Index(msg, "\n"); i >= 0 {
			msg = msg[:i] + u + msg[i:]
		} else {
			msg += u
		}
	}

	span := spanOf(errp.desc[0].posn)
	e := types.Error{
		Fset: check.fset,
		Pos:  span.pos,
		Msg:  msg,
		Soft: errp.soft,
	}

	// Cheap trick: Don't report errors with messages containing
	// "invalid operand" or "invalid type" as those tend to be
	// follow-on errors which don't add useful information. Only
	// exclude them if these strings are not at the beginning,
	// and only if we have at least one error already reported.
	isInvalidErr := strings.Index(e.Msg, "invalid operand") > 0 || strings.Index(e.Msg, "invalid type") > 0
	if check.firstErr != nil && isInvalidErr {
		return
	}

	e.Msg = stripAnnotations(e.Msg)
	if check.errpos != nil {
		// If we have an internal error and the errpos override is set, use it to
		// augment our error positioning.
		// TODO(rFindley) we may also want to augment the error message and refer
		// to the position (pos) in the original expression.
		span := spanOf(check.errpos)
		e.Pos = span.pos
	}
	err := e

	if check.firstErr == nil {
		check.firstErr = err
	}

	if check.conf._Trace {
		pos := e.Pos
		msg := e.Msg
		check.trace(pos, "ERROR: %s", msg)
	}

	f := check.conf.Error
	if f == nil {
		panic(bailout{}) // report only first error
	}
	f(err)
}

// newErrorf creates a new error_ for later reporting with check.report.
func newErrorf(at positioner, code Code, format string, args ...any) *error_ {
	return &error_{
		desc: []errorDesc{{at, format, args}},
		code: code,
	}
}

func (check *Checker) error(at positioner, code Code, msg string) {
	check.report(newErrorf(at, code, "%s", msg))
}

func (check *Checker) errorf(at positioner, code Code, format string, args ...any) {
	check.report(newErrorf(at, code, format, args...))
}

func (check *Checker) softErrorf(at positioner, code Code, format string, args ...any) {
	err := newErrorf(at, code, format, args...)
	err.soft = true
	check.report(err)
}

// The positioner interface is used to extract the position of type-checker
// errors.
type positioner interface {
	Pos() token.Pos
}

// posSpan holds a position range along with a highlighted position within that
// range. This is used for positioning errors, with pos by convention being the
// first position in the source where the error is known to exist, and start
// and end defining the full span of syntax being considered when the error was
// detected. Invariant: start <= pos < end || start == pos == end.
type posSpan struct {
	start, pos, end token.Pos
}

func (e posSpan) Pos() token.Pos {
	return e.pos
}

// atPos wraps a token.Pos to implement the positioner interface.
type atPos token.Pos

func (s atPos) Pos() token.Pos {
	return token.Pos(s)
}

// spanOf extracts an error span from the given positioner. By default this is
// the trivial span starting and ending at pos, but this span is expanded when
// the argument naturally corresponds to a span of source code.
func spanOf(at positioner) posSpan {
	switch x := at.(type) {
	case nil:
		panic("nil positioner")
	case posSpan:
		return x
	case ast.Node:
		pos := x.Pos()
		return posSpan{pos, pos, x.End()}
	case *operand:
		if x.expr != nil {
			pos := x.Pos()
			return posSpan{pos, pos, x.expr.End()}
		}
		return posSpan{token.NoPos, token.NoPos, token.NoPos}
	default:
		pos := at.Pos()
		return posSpan{pos, pos, pos}
	}
}

// stripAnnotations removes internal (type) annotations from s.
func stripAnnotations(s string) string {
	var buf strings.Builder
	for _, r := range s {
		// strip #'s and subscript digits
		if r < '₀' || '₀'+10 <= r { // '₀' == U+2080
			buf.WriteRune(r)
		}
	}
	if buf.Len() < len(s) {
		return buf.String()
	}
	return s
}
