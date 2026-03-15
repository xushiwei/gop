package main

import (
	"fmt"
	"github.com/qiniu/x/stringslice"
	"github.com/qiniu/x/stringutil"
)

func main() {
	a := []string{"hello", "world", "123"}
	fmt.Println(stringslice.Capitalize(a))
	fmt.Println(stringutil.Contains("param-required required", "required"))
	fmt.Println(stringutil.Contains("param-required required", "param"))
}
