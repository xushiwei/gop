package main

import "fmt"

type foo struct {
}

func (p *foo) Gop_Enum(c func(key int, val string)) {
}
func main() {
	new(foo).Gop_Enum(func(k int, v string) {
		fmt.Println(k, v)
	})
}
