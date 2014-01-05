package main

import (
	"fmt"
	"unsafe"
)

type RWTestStruct1 struct {
	Id   int
	Name string
}

func main() {
	fmt.Println("Hello World 11!")
	o := &RWTestStruct1{1, "test"}
	p := unsafe.Pointer(o)
	rw := getFastRWer(o, p)
	fmt.Println(rw)
}
