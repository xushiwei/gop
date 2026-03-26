package conflict

const XGoPackage = true

type Foo struct {
}

const XGoo_Foo_Mul = ".Mul,.MulInt"

func (a Foo) MulInt(b int) Foo {
	return a
}

func (a Foo) Mul(b Foo) Foo {
	return a
}

const XGoo_Mul = "Mul,MulInt"

func Mul(a, b Foo) Foo {
	return a
}

func MulInt(a Foo, b int) Foo {
	return a
}
