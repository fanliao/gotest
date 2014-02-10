package main

import (
//"testing"
)

type struct1 struct {
	i int
}

func (this struct1) GetI() int {
	return this.i
}

type struct2 struct {
	j int
	*struct1
}

func (this struct2) GetJ() int {
	return this.j
}

func (this struct1) getMethod() string {
	return "hello"
}

func getFunc() string {
	return "hello"
}

//func BenchmarkFuncCall(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		_ = getFunc()
//	}
//}

//func BenchmarkMethodCall(b *testing.B) {
//	o := struct1{}
//	for i := 0; i < b.N; i++ {
//		_ = o.getMethod()
//	}
//}
