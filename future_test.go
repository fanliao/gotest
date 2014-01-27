package main

import (
	"testing"
)

func TestDone(t *testing.T) {
	c := func(v ...interface{}) {
		fmt.Println("callback", v)
	}
	f := NewFuture().Done(c)

	go func() {
		time.Sleep(0.5 * time.Second)
		f.Reslove(10)
		fmt.Println("send done")
	}()

	fmt.Println("end start")
	return f

}
