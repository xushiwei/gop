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

type monoGraph struct {
	vertices []monoVertex
	edges    []monoEdge

	// canon maps method receiver type parameters to their respective
	// receiver type's type parameters.
	canon map[*types.TypeParam]*types.TypeParam

	// nameIdx maps a defined type or (canonical) type parameter to its
	// vertex index.
	nameIdx map[*types.TypeName]int
}

type monoVertex struct {
	weight int // weight of heaviest known path to this vertex
	pre    int // previous edge (if any) in the above path
	len    int // length of the above path

	// obj is the defined type or type parameter represented by this
	// vertex.
	obj *types.TypeName
}

type monoEdge struct {
	dst, src int
	weight   int

	pos token.Pos
	typ types.Type
}
