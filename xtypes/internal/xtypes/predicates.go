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

// isGeneric reports whether a type is a generic, uninstantiated type
// (generic signatures are not included).
// TODO(gri) should we include signatures or assert that they are not present?
func IsGeneric(t types.Type) bool {
	panic("todo") // see types.isGeneric(t)
}

func IsConstType(t types.Type) bool { return IsBasic(t, types.IsConstType) }

// isBasic reports whether under(t) is a basic type with the specified info.
// If t is a type parameter the result is false; i.e.,
// isBasic does not look inside a type parameter.
func IsBasic(t types.Type, info types.BasicInfo) bool {
	u, _ := Under(t).(*types.Basic)
	return u != nil && u.Info()&info != 0
}

// isTyped reports whether t is typed; i.e., not an untyped
// constant or boolean. isTyped may be called with types that
// are not fully set up.
func IsTyped(t types.Type) bool {
	// isTyped is called with types that are not fully
	// set up. Must not call under()!
	b, _ := t.(*types.Basic)
	return b == nil || b.Info()&types.IsUntyped == 0
}
