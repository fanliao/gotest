package main

import (
	"testing"
	"time"
)

func TestDone(t *testing.T) {

	order := make([]int, 0, 10)
	f := NewFuture().Done(func(v ...interface{}) {
		t.Log(v...)
		if len(v) != 2 || (v[0]).(int) != 10 || (v[1]).(string) != "ok" {
			t.Error("expect [10, 'ok'], actrul: ")
			t.Error(v...)
		}
	})

	go func() {
		time.Sleep(500 * time.Millisecond)
		f.Reslove(10, "ok")
		t.Log("Resloved")
		order = append(order, 1)
	}()

	time.Sleep(1 * time.Second)
	order = append(order, 2)
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Error("expect [1, 2], actrul: ")
		t.Error(order[0], order[1])
	}
	t.Log("End")
}
