// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xtypes

import (
	"go/types"
)

// under returns the true expanded underlying type.
// If it doesn't exist, the result is Typ[Invalid].
// under must only be called when a type is known
// to be fully set up.
func Under(t types.Type) types.Type {
	if t, _ := t.(*types.Named); t != nil {
		return Named_under(t)
	}
	return t.Underlying()
}
