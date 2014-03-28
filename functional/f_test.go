package main

import (
	"math"
	"runtime"
	"testing"
)

const count int = 10000

func init() {
	runtime.GOMAXPROCS(2)
}

func where1(v interface{}) bool {
	i := v.(int)
	//time.Sleep(10 * time.Nanosecond)
	return i < 5000
}

func select1(v interface{}) interface{} {
	return v.(int) + 9999
}

func select2(v interface{}) interface{} {
	return math.Sin(math.Cos(math.Pow(float64(v.(int)), 2)))
}

//func BenchmarkAsyncStep(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		src := make(chan interface{}, 1)
//		go func() {
//			for i := 0; i < count; i++ {
//				src <- i
//			}
//			close(src)
//		}()

//		q := queryable{src, make([]func(chan interface{}) chan interface{}, 0, 1)}
//		dst := q.Where(where1).Select(select1).Select(select2).Get()

//		j := 0
//		for v := range dst {
//			_ = v
//			j = j + 1
//		}
//		if j != 5000 {
//			b.Fail()
//			b.Error("size is ", j)
//		}
//	}
//}

func BenchmarkSyncStep(b *testing.B) {
	for i := 0; i < b.N; i++ {
		src := make([]interface{}, 0, count)
		//go func() {
		for i := 0; i < count; i++ {
			src = append(src, i)
		}
		//}()

		q := queryableS{src, make([]func([]interface{}) []interface{}, 0, 1)}
		dst := q.Where(where1).Select(select1).Select(select2).Get()

		j := 0
		for v := range dst {
			_ = v
			j = j + 1
		}
		if j != 5000 {
			b.Fail()
			b.Error("size is ", j)
		}
	}
}
