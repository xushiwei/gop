package main

import (
	"fmt"
	"github.com/goplus/xgo/cl/internal/spx"
)

type Kai struct {
	spx.Sprite
	*MyGame
}
type MyGame struct {
	*spx.MyGame
}

func (this *MyGame) Main() {
	spx.Gopt_MyGame_Main(this)
}
func (this *Kai) Main() {
	this.OnStart(func() {
		x := 1
		spx.RepeatUntil(func() bool {
			return x > 10
		}, func() {
			fmt.Println("Hi")
			x++
		})
		spx.When(func() bool {
			return x == 11
		}, func() {
			fmt.Println("x = 11")
		})
		spx.ForEver(func() {
			this.Gop_Exec("step", 1)
		})
	})
}
func main() {
	new(MyGame).Main()
}
