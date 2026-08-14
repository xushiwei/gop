package main

import "fmt"

type foo struct {
}

func (p *foo) Gop_Enum(c func(val string)) {
}
func main() {
	fmt.Println(func() (_xgo_ret []string) {
		new(foo).Gop_Enum(func(v string) {
			_xgo_ret = append(_xgo_ret, v)
		})
		return
	}())
}
