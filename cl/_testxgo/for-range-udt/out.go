package main

import "fmt"

type foo struct {
}

func (p *foo) Gop_Enum() func(yield func(key int, val string) bool) {
	return nil
}
func main() {
	for k, v := range new(foo).Gop_Enum() {
		fmt.Println(k, v)
	}
}
