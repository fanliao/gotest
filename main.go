package main

import (
	"fmt"
)

type RWTestStruct1 struct {
	Id   int
	Name string
}

func main() {
	fmt.Println("Hello World 11!")
	o := &RWTestStruct1{1, "test"}
	rw := GetFastRWer(o)
	fmt.Println(rw)
}
