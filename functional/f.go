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
		//dst := make([]interface{}, 0, len(this.source))
		//mapSlice(src, func(v interface{}, out *[]interface{}) {
		//	if sure(v) {
		//		*out = append(*out, v)
		//	}
		//}, &dst)
		dst := filterSlice(src, sure)
		return dst
	}
	this.actions = append(this.actions, action)
	return this
}

func (this queryableS) Select(f func(interface{}) interface{}) queryableS {
	action := func(src []interface{}) []interface{} {
		//dst := make([]interface{}, 0, len(this.source))
		//mapSlice(src, func(v interface{}, out *[]interface{}) {
		//	*out = append(*out, f(v))
		//}, &dst)
		dst := mapSlice(src, f, nil)
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

type power struct {
	i int
	p int
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

	arrInts := make([]int, 0, 100)
	src1 := make([]interface{}, 0, 100)
	pow1 := make([]interface{}, 0, 100)
	//go func() {
	for i := 0; i < count; i++ {
		arrInts = append(arrInts, i)
		src1 = append(src1, i)
	}
	for i := 10; i < count-30; i++ {
		pow1 = append(pow1, power{i, i * i})
		pow1 = append(pow1, power{i, i * 100})
	}
	//}()

	q1 := queryableS{src1, make([]func([]interface{}) []interface{}, 0, 1)}
	dst1 := q1.Where(func(v interface{}) bool {
		i := v.(int)
		//time.Sleep(10 * time.Nanosecond)
		return i < 54
	}).Get()
	fmt.Println("dst1", dst1)

	dst := From(src1).Where(func(v interface{}) bool {
		i := v.(int)
		return i%2 == 0
	}).Select(func(v interface{}) interface{} {
		i := v.(int)
		return "item" + strconv.Itoa(i)
	}).Results()
	fmt.Println("where select get dst", dst)

	dst = From(arrInts).Where(func(v interface{}) bool {
		i := v.(int)
		return i%2 == 0
	}).Select(func(v interface{}) interface{} {
		i := v.(int)
		return "item" + strconv.Itoa(i)
	}).Results()
	fmt.Println("Int slice where select get dst", dst)

	dst = From(src1).GroupBy(func(v interface{}) interface{} {
		return v.(int) / 10
	}).Results()
	for _, o := range dst {
		kv := o.(*keyValue)
		fmt.Println("group get k=", kv.key, ";v=", kv.value, " ")
	}
	fmt.Println("")

	dst = From(src1).Join(pow1,
		func(o interface{}) interface{} { return o },
		func(i interface{}) interface{} { return i.(power).i },
		func(o interface{}, i interface{}) interface{} {
			o1, i1 := o.(int), i.(power)
			return strconv.Itoa(o1) + ";" + strconv.Itoa(i1.p)
		}).Results()
	fmt.Println("join ", src1)
	fmt.Println("with", pow1)
	fmt.Println("join return", dst)

	chSrc := make(chan *chunk)
	go func() {
		chSrc <- &chunk{src1[0:25], 0}
		chSrc <- &chunk{src1[25:50], 25}
		chSrc <- &chunk{src1[50:75], 50}
		chSrc <- &chunk{src1[75:100], 75}
		chSrc <- nil
		fmt.Println("close src", chSrc)
	}()

	//for v := range chSrc {
	//	fmt.Println(v)
	//}

	dst = From(chSrc).Where(func(v interface{}) bool {
		i := v.(int)
		return i%2 == 0
	}).Select(func(v interface{}) interface{} {
		i := v.(int)
		return "item" + strconv.Itoa(i)
	}).KeepOrder(true).Results()
	fmt.Println("dst of chunk", dst)
	fmt.Println("Hello World2")

	//fmt.Println("s" + strconv.Itoa(100000))

	a := []interface{}{3, 2, 1, 4, 5, 6, 7, 10, 9, 8, 7, 6}
	avl := NewAvlTree(func(a interface{}, b interface{}) int {
		a1, b1 := a.(int), b.(int)
		if a1 < b1 {
			return -1
		} else if a1 == b1 {
			return 0
		} else {
			return 1
		}
	})
	for i := 0; i < len(a); i++ {
		avl.Insert(a[i])
	}
	//_ = taller
	//result := make([]interface{}, 0, 10)
	//avlToSlice(tree, &result)
	result := avl.ToSlice()
	fmt.Println("avl result=", result, "count=", avl.count)

}
