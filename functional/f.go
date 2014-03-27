package main

import (
	"fmt"
	"time"
)

const (
	BATCH_SIZE = 50
)

type chunk struct {
	start int
	end   int
}

type queryable struct {
	source  chan interface{}
	actions []func(chan interface{}) chan interface{}
	datas   []interface{}
}

func (this queryable) Where(sure func(interface{}) bool) queryable {
	action := func(src chan interface{}) chan interface{} {
		return forChan(src, func(v interface{}, dst chan interface{}) {
			if sure(v) {
				dst <- v
			}
		})
	}
	this.actions = append(this.actions, action)
	return this
}

func (this queryable) Select(f func(interface{}) interface{}) queryable {
	action := func(src chan interface{}) chan interface{} {
		return forChan(src, func(v interface{}, dst chan interface{}) {
			dst <- f(v)
		})

	}
	this.actions = append(this.actions, action)
	return this
}

func (this queryable) Get() chan interface{} {
	src := this.source
	datas := make([]interface{}, 0, 10000)
	startChan := make(chan chunk, 10000)
	endChan := make(chan int)
	go func() {
		count, start, end := 0, 0, 0
		for v := range src {
			datas = append(datas, v)
			count++
			if count == BATCH_SIZE {
				end = len(datas) - 1
				startChan <- chunk{start, end}
				fmt.Println("send", start, end, datas[start:end+1])
				count, start = 0, end+1
			}
		}
		endChan <- 1
		close(startChan)
	}()
	<-endChan
	for _, action := range this.actions {
		src = action(src)
	}
	return src
}

func forChan(src chan interface{}, f func(interface{}, chan interface{})) chan interface{} {
	dst := make(chan interface{}, 1)
	go func() {
		for v := range src {
			f(v, dst)
		}
		close(dst)
	}()
	return dst
}

type queryableS struct {
	source  []interface{}
	actions []func([]interface{}) []interface{}
}

func (this queryableS) Where(sure func(interface{}) bool) queryableS {
	action := func(src []interface{}) []interface{} {
		dst := make([]interface{}, 0, len(this.source))
		forSlice(src, func(v interface{}) {
			if sure(v) {
				dst = append(dst, v)
			}
		})
		return dst
	}
	this.actions = append(this.actions, action)
	return this
}

func (this queryableS) Select(f func(interface{}) interface{}) queryableS {
	action := func(src []interface{}) []interface{} {
		dst := make([]interface{}, 0, len(this.source))
		forSlice(src, func(v interface{}) {
			dst = append(dst, f(v))
		})
		return dst
	}
	this.actions = append(this.actions, action)
	return this
}

func (this queryableS) Get() []interface{} {
	data := this.source
	for _, action := range this.actions {
		data = action(data)
	}
	return data
}

func forSlice(src []interface{}, f func(interface{})) {
	for v := range src {
		f(v)
	}
}

func main() {
	time.Now()
	count := 100
	src := make(chan interface{}, 1)
	go func() {
		for i := 0; i < count; i++ {
			src <- i
		}
		close(src)
	}()

	q := queryable{src, make([]func(chan interface{}) chan interface{}, 0, 1), nil}
	dst := q.Where(func(v interface{}) bool {
		i := v.(int)
		return i < 50
	}).Select(func(v interface{}) interface{} {
		return v.(int) * 100
	}).Get()

	for v := range dst {
		fmt.Print(v, " ")
	}
	fmt.Println()

	src1 := make([]interface{}, 0, 100)
	//go func() {
	for i := 0; i < count; i++ {
		src1 = append(src1, i)
	}
	//}()

	q1 := queryableS{src1, make([]func([]interface{}) []interface{}, 0, 1)}
	dst1 := q1.Where(func(v interface{}) bool {
		i := v.(int)
		//time.Sleep(10 * time.Nanosecond)
		return i < 50
	}).Select(func(v interface{}) interface{} {
		return v.(int) * 100
	}).Get()

	fmt.Println(dst1)

	fmt.Println("Hello World")
}
