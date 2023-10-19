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
	"go/constant"
	"go/token"
	"go/types"
)

// color encodes the color of an object (see Checker.objDecl for details).
type Color uint32

// An object may be painted in one of three colors.
// Color values other than white or black are considered grey.
const (
	White Color = iota
	Black
	Grey // must be > white and black
)

func (c Color) String() string {
	switch c {
	case White:
		return "white"
	case Black:
		return "black"
	default:
		return "grey"
	}
}

// color returns the object's color.
func Object_color(types.Object) Color {
	panic("todo")
}

// setColor sets the object's color. It must not be white.
func Object_setColor(obj types.Object, color Color) {
	panic("todo")
}

// order reflects a package-level object's source order: if object
// a is before object b in the source, then a.order() < b.order().
// order returns a value > 0 for package-level objects; it returns
// 0 for all other objects (including objects in file scopes).
func Object_order(types.Object) uint32 {
	panic("todo")
}

// setOrder sets the order number of the object. It must be > 0.
func Object_setOrder(types.Object, uint32) {
	panic("todo")
}

// setScopePos sets the start position of the scope for this Object.
func Object_setScopePos(obj types.Object, pos token.Pos) {
	panic("todo")
}

func Const_setVal(obj *types.Const, val constant.Value) {
	panic("todo")
}

func Const_setTyp(obj *types.Const, typ types.Type) {
	panic("todo")
}

func Var_setTyp(obj *types.Var, typ types.Type) {
	panic("todo")
}

func TypeName_setTyp(obj *types.TypeName, typ types.Type) {
	panic("todo")
}

func Func_setParent(obj *types.Func, parent *types.Scope) {
	panic("todo")
}

func Func_setHasPtrRecv_(obj *types.Func, hasPtrRecv_ bool) {
	panic("todo")
}

func PkgName_setUsed(pkgName *types.PkgName, used bool) {
	panic("todo")
}

func PackagePrefix(pkg *types.Package, qf types.Qualifier) string {
	if pkg == nil {
		return ""
	}
	var s string
	if qf != nil {
		s = qf(pkg)
	} else {
		s = pkg.Path()
	}
	if s != "" {
		s += "."
	}
	return s
}
