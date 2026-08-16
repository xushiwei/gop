package main

import "fmt"

type foo struct {
}

func (p *foo) Gop_Enum() func(yield func(val string) bool) {
	return nil
}
func main() {
	for v := range new(foo).Gop_Enum() {
		fmt.Println(v)
	}
}
