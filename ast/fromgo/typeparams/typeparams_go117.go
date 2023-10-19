//go:build !go1.18
// +build !go1.18

/*
 * Copyright (c) 2022 The GoPlus Authors (goplus.org). All rights reserved.
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

package typeparams

import (
	"go/ast"
	"go/token"
)

func unsupported() {
	panic("type parameters are unsupported at this go version")
}

// IndexListExpr is a placeholder type, as type parameters are not supported at
// this Go version. Its methods panic on use.
type IndexListExpr struct {
	ast.Expr
	X       ast.Expr   // expression
	Lbrack  token.Pos  // position of "["
	Indices []ast.Expr // index expressions
	Rbrack  token.Pos  // position of "]"
}

func (*IndexListExpr) Pos() token.Pos { unsupported(); return token.NoPos }
func (*IndexListExpr) End() token.Pos { unsupported(); return token.NoPos }

// ForFuncType returns an empty field list, as type parameters are not
// supported at this Go version.
func ForFuncType(*ast.FuncType) *ast.FieldList {
	return nil
}

// ForTypeSpec returns an empty field list, as type parameters are not
// supported at this Go version.
func ForTypeSpec(*ast.TypeSpec) *ast.FieldList {
	return nil
}
