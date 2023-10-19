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

// This file contains a definition of the type-checking context; an opaque type
// that may be supplied by users during instantiation.
//
// Contexts serve two purposes:
//  - reduce the duplication of identical instances
//  - short-circuit instantiation cycles
//
// For the latter purpose, we must always have a context during instantiation,
// whether or not it is supplied by the user. For both purposes, it must be the
// case that hashing a pointer-identical type produces consistent results
// (somewhat obviously).
//
// However, neither of these purposes require that our hash is perfect, and so
// this was not an explicit design goal of the context type. In fact, due to
// concurrent use it is convenient not to guarantee de-duplication.
//
// Nevertheless, in the future it could be helpful to allow users to leverage
// contexts to canonicalize instances, and it would probably be possible to
// achieve such a guarantee.

// A Context is an opaque type checking context. It may be used to share
// identical type instances across type-checked packages or calls to
// Instantiate. Contexts are safe for concurrent use.
//
// The use of a shared context does not guarantee that identical instances are
// deduplicated in all cases.
type Context = types.Context

// NewContext creates a new Context.
func NewContext() *Context {
	return types.NewContext()
}
