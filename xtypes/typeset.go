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

	"github.com/goplus/gop/token"
)

// A _TypeSet represents the type set of an interface.
// Because of existing language restrictions, methods can be "factored out"
// from the terms. The actual type set is the intersection of the type set
// implied by the methods and the type set described by the terms and the
// comparable bit. To test whether a type is included in a type set
// ("implements" relation), the type must implement all methods _and_ be
// an element of the type set described by the terms and the comparable bit.
// If the term list describes the set of all types and comparable is true,
// only comparable types are meant; in all other cases comparable is false.
type _TypeSet struct {
	methods    []*types.Func // all methods of the interface; sorted by unique ID
	terms      termlist      // type terms of the type set
	comparable bool          // invariant: !comparable || terms.isAll()
}

// IsEmpty reports whether type set s is the empty set.
func (s *_TypeSet) IsEmpty() bool { return s.terms.isEmpty() }

// IsAll reports whether type set s is the set of all types (corresponding to the empty interface).
func (s *_TypeSet) IsAll() bool { return s.IsMethodSet() && len(s.methods) == 0 }

// IsMethodSet reports whether the interface t is fully described by its method set.
func (s *_TypeSet) IsMethodSet() bool { return !s.comparable && s.terms.isAll() }

// computeInterfaceTypeSet may be called with check == nil.
func computeInterfaceTypeSet(check *Checker, pos token.Pos, ityp *types.Interface) *_TypeSet {
	panic("todo")
}
