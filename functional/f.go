package main

import (
	"fmt"
	"strconv"
	"time"
)

//queryable-------------------------------------------------------------------------------------

//type queryable struct {
//	source  chan interface{}
//	actions []func(chan interface{}) chan interface{}
//	datas   []interface{}
//}

//func (this queryable) Where(sure func(interface{}) bool) queryable {
//	action := func(src chan interface{}) chan interface{} {
//		return forChan(src, func(v interface{}, dst chan interface{}) {
//			if sure(v) {
//				dst <- v
//			}
//		})
//	}
//	this.actions = append(this.actions, action)
//	return this
//}

//func (this queryable) Select(f func(interface{}) interface{}) queryable {
//	action := func(src chan interface{}) chan interface{} {
//		return forChan(src, func(v interface{}, dst chan interface{}) {
//			dst <- f(v)
//		})

//	}
//	this.actions = append(this.actions, action)
//	return this
//}

//func (this queryable) Get() chan interface{} {
//	src := this.source
//	datas := make([]interface{}, 0, 10000)
//	startChan := make(chan chunk, 10000)
//	endChan := make(chan int)
//	go func() {
//		count, start, end := 0, 0, 0
//		for v := range src {
//			datas = append(datas, v)
//			count++
//			if count == BATCH_SIZE {
//				end = len(datas) - 1
//				startChan <- chunk{start, end}
//				fmt.Println("send", start, end, datas[start:end+1])
//				count, start = 0, end+1
//			}
//		}
//		endChan <- 1
//		close(startChan)
//	}()
//	<-endChan
//	for _, action := range this.actions {
//		src = action(src)
//	}
//	return src
//}

//func forChan(src chan interface{}, f func(interface{}, chan interface{})) chan interface{} {
//	dst := make(chan interface{}, 1)
//	go func() {
//		for v := range src {
//			f(v, dst)
//		}
//		close(dst)
//	}()
//	return dst
//}

type queryableS struct {
	source  []interface{}
	actions []func([]interface{}) []interface{}
}

func (this queryableS) Where(sure func(interface{}) bool) queryableS {
	action := func(src []interface{}) []interface{} {
		dst := make([]interface{}, 0, len(this.source))
		forSlice(src, func(v interface{}, out *[]interface{}) {
			if sure(v) {
				*out = append(*out, v)
			}
		}, &dst)
		return dst
	}
	this.actions = append(this.actions, action)
	return this
}

func (this queryableS) Select(f func(interface{}) interface{}) queryableS {
	action := func(src []interface{}) []interface{} {
		dst := make([]interface{}, 0, len(this.source))
		forSlice(src, func(v interface{}, out *[]interface{}) {
			*out = append(*out, f(v))
		}, &dst)
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

func main() {
	time.Now()
	count := 100
	//src := make(chan interface{}, 1)
	//go func() {
	//	for i := 0; i < count; i++ {
	//		src <- i
	//	}
	//	close(src)
	//}()

	//q := queryable{src, make([]func(chan interface{}) chan interface{}, 0, 1), nil}
	//dst := q.Where(func(v interface{}) bool {
	//	i := v.(int)
	//	return i < 50
	//}).Select(func(v interface{}) interface{} {
	//	return v.(int) * 100
	//}).Get()

	//for v := range dst {
	//	fmt.Print(v, " ")
	//}
	//fmt.Println()

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
		return i < 54
	}).Get()
	fmt.Println("dst1", dst1)

	//s := blockSource{src1, 2}
	//whereAct := where(func(v interface{}) bool {
	//	i := v.(int)
	//	return i%2 == 0
	//})
	dst := From(src1).Where(func(v interface{}) bool {
		i := v.(int)
		return i%2 == 0
	}).Select(func(v interface{}) interface{} {
		i := v.(int)
		return "item" + strconv.Itoa(i)
	}).Results()
	//dst := From(src1).Where(func(v interface{}) bool {
	//	i := v.(int)
	//	return i%2 == 0
	//}).Results()
	fmt.Println("dst", dst)

	//chSrc := make(chan *chunk)
	//go func() {
	//	chSrc <- &chunk{src1, 0, 24}
	//	chSrc <- &chunk{src1, 25, 49}
	//	chSrc <- &chunk{src1, 50, 74}
	//	chSrc <- &chunk{src1, 75, 99}
	//	chSrc <- nil
	//	fmt.Println("close src", chSrc)
	//}()

	////for v := range chSrc {
	////	fmt.Println(v)
	////}

	//cs := chunkSource{chSrc, 2}
	//dst = whereAct(cs)
	//fmt.Println("dst of chunk", (dst.(blockSource)).data)
	//fmt.Println("Hello World2")

	//fmt.Println("s" + strconv.Itoa(100000))

	s := []interface{}{0, 1, 2}
	fmt.Println(s[0:3])

}
