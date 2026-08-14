package main

import "fmt"

type fooIter struct {
	data *foo
	idx  int
}
type foo struct {
	key []int
	val []string
}

func (p *fooIter) Next() (key int, val string, ok bool) {
	if p.idx < len(p.data.key) {
		key, val, ok = p.data.key[p.idx], p.data.val[p.idx], true
		p.idx++
	}
	return
}
func (p *foo) Gop_Enum() *fooIter {
	return &fooIter{data: p}
}
func newFoo() *foo {
	return &foo{key: []int{3, 7}, val: []string{"Hi", "XGo"}}
}
func main() {
	for _xgo_it := newFoo().Gop_Enum(); ; {
		var _xgo_ok bool
		k, v, _xgo_ok := _xgo_it.Next()
		if !_xgo_ok {
			break
		}
		fmt.Println(k, v)
	}
}
