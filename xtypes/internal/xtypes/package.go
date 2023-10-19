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
	"go/types"
)

func Package_appendImport(pkg *types.Package, imp *types.Package) {
	pkg.SetImports(append(pkg.Imports(), imp))
}

func Package_scope(pkg *types.Package) *types.Scope {
	return pkg.Scope()
}

func Package_fake(pkg *types.Package) (fake bool) {
	panic("todo")
}

func Package_setFake(pkg *types.Package, fake bool) {
	panic("todo")
}

func Package_cgo(pkg *types.Package) (cgo bool) {
	panic("todo")
}

func Package_setCgo(pkg *types.Package, cgo bool) {
	panic("todo")
}
