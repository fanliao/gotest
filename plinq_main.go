package main

import (
	"fmt"
	"github.com/fanliao/go-plinq"
	"strconv"
	"time"
)

type power struct {
	i int
	p int
}

func getChanSrc(src []interface{}) chan interface{} {
	chanSrc := make(chan interface{})
	go func() {
		for _, v := range src {
			chanSrc <- v
		}
		close(chanSrc)
	}()
	return chanSrc
}

func getIntChanSrc(src []int) chan int {
	chanSrc := make(chan int)
	go func() {
		for _, v := range src {
			chanSrc <- v
		}
		close(chanSrc)
	}()
	return chanSrc
}

func TestLinq() {
	time.Now()
	count := 20

	arrInts := make([]int, 0, 20)
	src1 := make([]interface{}, 0, 20)
	src2 := make([]interface{}, 0, 20)
	powers := make([]interface{}, 0, 20)
	for i := 0; i < count; i++ {
		arrInts = append(arrInts, i)
		src1 = append(src1, i)
		src2 = append(src2, i+count/2)
	}
	for i := count / 4; i < count/2; i++ {
		powers = append(powers, power{i, i * i})
		powers = append(powers, power{i, i * 100})
	}

	var whereFunc = func(v interface{}) bool {
		//var ss []int
		//_ = ss[2]
		i := v.(int)
		return i%2 == 0
	}
	_ = whereFunc
	var selectFunc = func(v interface{}) interface{} {
		i := v.(int)
		return "item" + strconv.Itoa(i)
	}
	_ = selectFunc
	var groupKeyFunc = func(v interface{}) interface{} {
		return v.(int) / 10
	}
	_ = groupKeyFunc

	var joinResultSelector = func(o interface{}, i interface{}) interface{} {
		if i == nil {
			return strconv.Itoa(o.(int))
		} else {
			o1, i1 := o.(int), i.(power)
			return strconv.Itoa(o1) + ";" + strconv.Itoa(i1.p)
		}
	}
	_ = joinResultSelector

	var groupJoinResultSelector = func(o interface{}, is []interface{}) interface{} {
		return plinq.KeyValue{o, is}
	}
	_ = groupJoinResultSelector

	testLinqOpr("Where opretion", func() ([]interface{}, error) {
		return plinq.From(src1).Where(whereFunc).Results()
	})

	//test where and select
	testLinqWithAllSource("Where and Select opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Where(whereFunc).Select(selectFunc)
	})

	pSrc := &src1
	q1 := plinq.From(pSrc).Where(whereFunc).Select(selectFunc)
	for i := count; i < count+10; i++ {
		src1 = append(src1, i)
	}
	rs1, err1 := q1.Results()
	fmt.Println("Where and Select from Pointer returns", rs1, err1, "\n")

	//test where and select with int slice
	dst, _ := plinq.From(arrInts).Where(whereFunc).Select(selectFunc).Results()
	fmt.Println("Int slice where select return", dst, "\n")

	dst, _ = plinq.From(getIntChanSrc(arrInts)).Where(whereFunc).Select(selectFunc).Results()
	fmt.Println("Int chan where select return", dst, "\n")

	//test group
	testLinqWithAllSource("Group opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.GroupBy(groupKeyFunc)
	}, func(dst []interface{}) {
		fmt.Println()
		for _, o := range dst {
			kv := o.(*plinq.KeyValue)
			fmt.Println("group get k=", kv.Key, ";v=", kv.Value, " ")
		}
	})

	//test group
	testLinqWithAllSource("Distinct opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Distinct()
	})

	//test left join
	testLinqWithAllSource("LeftJoin opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.LeftJoin(powers,
			func(o interface{}) interface{} { return o },
			func(i interface{}) interface{} { return i.(power).i },
			joinResultSelector)
	})

	//test left group join
	testLinqWithAllSource("LeftGroupJoin opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.LeftGroupJoin(powers,
			func(o interface{}) interface{} { return o },
			func(i interface{}) interface{} { return i.(power).i },
			groupJoinResultSelector)
	})

	//test union
	testLinqWithAllSource("Union opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Union(src2)
	})

	//test intersect
	testLinqWithAllSource("Intersect opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Intersect(src2)
	})

	//test except
	testLinqWithAllSource("Except opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Except(src2)
	})

	//test Concat
	testLinqWithAllSource("Concat opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Concat(src2)
	})

	//test Reverse
	testLinqWithAllSource("Reverse opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Reverse()
	})

	//test Aggregate
	testLinqAggWithAllSource("aggregate opretions", src1, func(q *plinq.Queryable) (interface{}, error) {
		return q.Sum()
	})

	//test Average
	testLinqAggWithAllSource("average opretions", src1, func(q *plinq.Queryable) (interface{}, error) {
		return q.Average()
	})

	//test Max
	testLinqAggWithAllSource("max opretions", src1, func(q *plinq.Queryable) (interface{}, error) {
		return q.Max()
	})

	//test Min
	testLinqAggWithAllSource("min opretions", src1, func(q *plinq.Queryable) (interface{}, error) {
		return q.Min()
	})

	//test aggregate multiple operation
	testLinqAggWithAllSource("aggregate multiple opretions", src1, func(q *plinq.Queryable) (interface{}, error) {
		return q.Aggregate(plinq.Sum, plinq.Count, plinq.Max, plinq.Min)
	})

	//test Skip
	testLinqWithAllSource("Skip opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Skip(12)
	})

	//test Skip
	testLinqWithAllSource("Skip all opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Skip(31)
	})

	//test Skip
	testLinqWithAllSource("Skip 0 opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Skip(-1)
	})

	//test Take
	testLinqWithAllSource("Take opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Take(12)
	})

	//test Skip
	testLinqWithAllSource("Take all opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Take(31)
	})

	//test Skip
	testLinqWithAllSource("Take 0 opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.Take(-1)
	})

	//test SkipWhile
	testLinqWithAllSource("SkipWhile opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.SkipWhile(func(v interface{}) bool { return v.(int) <= 11 })
	})

	//test TakeWhile
	testLinqWithAllSource("TakeWhile opretions", src1, func(q *plinq.Queryable) *plinq.Queryable {
		return q.TakeWhile(func(v interface{}) bool { return v.(int) <= 11 })
	})

	//TODO: don't support the mixed type in aggregate
	////test aggregate multiple operation
	//testLinqAggWithAllSource("aggregate multiple opretions with mixed type", src1, func(q *plinq.Queryable) (interface{}, error) {
	//	return plinq.From([]interface{}{1, int8(2), uint(3), float64(4.4)}).Aggregate(Sum, Count, Max, Min)
	//})

	//TODO: don't support the mixed type in aggregate
	//test aggregate multiple operation
	testLinqAggWithAllSource("aggregate multiple opretions with mixed type", src1, func(q *plinq.Queryable) (interface{}, error) {
		return plinq.From([]interface{}{0, 3, 6, 9}).Aggregate(plinq.Sum, plinq.Count, plinq.Max, plinq.Min)
	})
	myAgg := &plinq.AggregateOpretion{"",
		func(v interface{}, t interface{}) interface{} {
			v1, t1 := v.(power), t.(string)
			return t1 + "|{" + strconv.Itoa(v1.i) + ":" + strconv.Itoa(v1.p) + "}"
		}, func(t1 interface{}, t2 interface{}) interface{} {
			return t1.(string) + t2.(string)
		}}
	//test customized aggregate operation
	testLinqAggWithAllSource("customized aggregate opretions", powers, func(q *plinq.Queryable) (interface{}, error) {
		return q.Aggregate(myAgg)
	})

	//fmt.Print("distinctKvs return:")
	//concats, _ := plinq.From(src1).Concat(src2).Results()
	//kvs, e := distinctKVs(concats, &plinq.ParallelOption{numCPU, plinq.DEFAULTCHUNKSIZE, false})
	//if e == nil {
	//	for _, v := range kvs {
	//		fmt.Print(v, " ")
	//	}
	//	fmt.Println(", len=", len(kvs), "\n")
	//} else {
	//	fmt.Println(e.Error(), "\n")
	//}

	size := count / 4
	getCChunkSrc := func() chan *plinq.Chunk {
		chunkSrc := make(chan *plinq.Chunk)
		go func() {
			chunkSrc <- &plinq.Chunk{src1[3*size : 4*size], 3, 3 * size}
			chunkSrc <- &plinq.Chunk{src1[0:size], 0, 0}
			chunkSrc <- &plinq.Chunk{src1[2*size : 3*size], 2, 2 * size}
			chunkSrc <- &plinq.Chunk{src1[size : 2*size], 1, size}
			func() {
				defer func() {
					if e := recover(); e != nil {
						_ = e
					}
				}()
				close(chunkSrc)
			}()
			fmt.Println("close src------------------", chunkSrc)
		}()
		return chunkSrc
	}

	fmt.Println("chunkchansource TakeWhile")
	chunkSrc := getCChunkSrc()
	dst, err := plinq.From(chunkSrc).TakeWhile(func(v interface{}) bool { return v.(int) <= 11 }).Results()
	if err == nil {
		fmt.Println("chunkchansource TakeWhile return", dst)
		fmt.Println()
	} else {
		fmt.Println("chunkchansource TakeWhile get error:", err)
		fmt.Println()
	}

	fmt.Println("chunkchansource SkipWhile")
	chunkSrc = getCChunkSrc()
	dst, err = plinq.From(chunkSrc).SkipWhile(func(v interface{}) bool { return v.(int) <= 11 }).Results()
	if err == nil {
		fmt.Println("chunkchansource SkipWhile return", dst)
		fmt.Println()
	} else {
		fmt.Println("chunkchansource SkipWhile get error:", err)
		fmt.Println()
	}

	fmt.Println("chunkchansource Take 14")
	chunkSrc = getCChunkSrc()
	dst, err = plinq.From(chunkSrc).Take(14).Results()
	if err == nil {
		fmt.Println("chunkchansource Take 14 return", dst)
		fmt.Println()
	} else {
		fmt.Println("chunkchansource Take 14 get error:", err)
		fmt.Println()
	}

	fmt.Println("chunkchansource Skip 14")
	chunkSrc = getCChunkSrc()
	dst, err = plinq.From(chunkSrc).Skip(14).Results()
	if err == nil {
		fmt.Println("chunkchansource Skip 14 return", dst)
		fmt.Println()
	} else {
		fmt.Println("chunkchansource Skip 14 get error:", err)
		fmt.Println()
	}

}

func testLinqOpr(title string, linqFunc func() ([]interface{}, error), rsHandlers ...func([]interface{})) {
	fmt.Print(title, " ")
	var rsHanlder func([]interface{})
	if rsHandlers != nil && len(rsHandlers) > 0 {
		rsHanlder = rsHandlers[0]
	} else {
		rsHanlder = func(dst []interface{}) { fmt.Print(dst) }
	}
	if dst, err := linqFunc(); err == nil {
		fmt.Print("return ============:")
		rsHanlder(dst)
		fmt.Println("\n")
	} else {
		fmt.Println("get error:\n", err, "\n")
	}
}

func testLinqAgg(title string, aggFunc func() (interface{}, error)) {
	fmt.Print(title, " ")
	if dst, err := aggFunc(); err == nil {
		fmt.Printf("return:%v", dst)
		fmt.Println("\n")
	} else {
		fmt.Println("get error:\n", err, "\n")
	}
}

func testLinqWithAllSource(title string, listSrc []interface{}, query func(*plinq.Queryable) *plinq.Queryable, rsHandlers ...func([]interface{})) {
	testLinqOpr(title, func() ([]interface{}, error) {
		return query(plinq.From(listSrc)).SetSizeOfChunk(9).Results()
	}, rsHandlers...)
	testLinqOpr("Chan source use "+title, func() ([]interface{}, error) {
		return query(plinq.From(getChanSrc(listSrc))).SetSizeOfChunk(9).Results()
	}, rsHandlers...)
}

func testLinqAggWithAllSource(title string, listSrc []interface{}, agg func(*plinq.Queryable) (interface{}, error)) {
	testLinqAgg(title, func() (interface{}, error) {
		return agg(plinq.From(listSrc))
	})
	testLinqAgg("Chan source use "+title, func() (interface{}, error) {
		return agg(plinq.From(getChanSrc(listSrc)))
	})
}

func testAVL() {
	a := []interface{}{3, 2, 1, 4, 5, 6, 7, 10, 9, 8, 7, 6}
	avl := plinq.NewAvlTree(func(a interface{}, b interface{}) int {
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
	fmt.Println("avl result=", result)

}

//func testHash() {
//	printHash("user1" + strconv.Itoa(10))
//	printHash("user" + strconv.Itoa(110))
//	printHash("user" + strconv.Itoa(0))
//	printHash("user" + strconv.Itoa(0))
//	printHash(nil)
//	printHash(nil)
//	printHash(111.11)
//	printHash(111.11)
//	printHash([]int{1, 2, 0})
//	printHash([]int{1, 2, 0})
//	printHash(0)
//	printHash(0)
//	printHash([]interface{}{1, "user" + strconv.Itoa(2), 0})
//	printHash([]interface{}{1, "user" + strconv.Itoa(2), 0})
//	slice := []interface{}{5, "user" + strconv.Itoa(5)}
//	printHash(slice)
//	printHash(slice)
//	printHash(power{1, 1})
//	printHash(power{1, 1})
//}

//func printHash(data interface{}) {
//	fmt.Println("hash", data, hash64(data))
//}

//func TestSortChunks() {
//	cs := []interface{}{&plinq.Chunk{nil, 7}, &plinq.Chunk{nil, 14}, &plinq.Chunk{nil, 21}, &plinq.Chunk{nil, 0}}
//	fmt.Println("\nbefore expandChunks(),", true, ":")
//	for _, v := range cs {
//		fmt.Print(v.(*plinq.Chunk).Order, ", ")
//	}
//	fmt.Println()
//	result := expandChunks(cs, true)
//	fmt.Println("\nafter expandChunks():")
//	for _, v := range result {
//		fmt.Print(v.(*plinq.Chunk).Order, ", ")
//	}
//	fmt.Println()

//}
