package main

import (
	//"io"
	"testing"
)

func multiRtn() (int, int) {
	return 1, 2
}

func add1(i int) int {
	return i + 1
}

func TestMultiReturn(t *testing.T) {
	//compile error: too many arguments in call to add1
	//t.Log(add1(multiRtn()))

	m := make(map[int]int)
	m[1] = 1
	//为什么m[1]也是返回2个值，但可以使用add1直接调用
	t.Log(add1(m[1]))

	if j, ok := m[1]; ok {
		t.Log(j)
	}

	//同样不行
	//add1(io.WriteString(nil, ""))
}
