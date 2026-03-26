package main

import "github.com/goplus/xgo/cl/internal/overload/conflict"

var a, b conflict.Foo
var c = a.Mul(b)
var d = a.MulInt(100)
var e = conflict.Mul(a, b)
var f = conflict.MulInt(a, 100)
