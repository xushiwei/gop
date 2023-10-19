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

func Scope_elems(s *types.Scope) map[string]types.Object {
	panic("todo")
}

func Scope_insert(s *types.Scope, name string, obj types.Object) {
	panic("todo")
}

// resolve returns the Object represented by obj, resolving lazy
// objects as appropriate.
func Resolve(name string, obj types.Object) types.Object {
	panic("todo") // see types.resolve
}
