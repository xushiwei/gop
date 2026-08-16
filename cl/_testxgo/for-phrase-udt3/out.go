package main

import "fmt"

type foo struct {
}

func (p *foo) Gop_Enum() func(yield func(val string) bool) {
	return nil
}
func main() {
	fmt.Println(func() (_xgo_ret []string) {
		for v := range new(foo).Gop_Enum() {
			_xgo_ret = append(_xgo_ret, v)
		}
		return
	}())
}
