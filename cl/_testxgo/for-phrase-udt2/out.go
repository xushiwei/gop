package main

import "fmt"

type fooIter struct {
}
type foo struct {
}

func (p fooIter) Next() (key string, val int, ok bool) {
	return
}
func (p *foo) Gop_Enum() fooIter {
	return fooIter{}
}
func main() {
	for _xgo_it := new(foo).Gop_Enum(); ; {
		var _xgo_ok bool
		k, v, _xgo_ok := _xgo_it.Next()
		if !_xgo_ok {
			break
		}
		fmt.Println(k, v)
	}
}
