package main

import (
//"testing"
)

type struct1 struct {
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
