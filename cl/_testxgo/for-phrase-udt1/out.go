package main

import "fmt"

type foo struct {
}

func (p *foo) Gop_Enum(c func(val string)) {
}
func main() {
	new(foo).Gop_Enum(func(v string) {
		fmt.Println(v)
	})
}
