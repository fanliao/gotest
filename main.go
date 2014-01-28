package main

import (
	"flag"
	"fmt"
	//"log"
	//"os"
	//"runtime/pprof"
	"errors"
	"reflect"
	"time"
	"unsafe"
)

const ptrSize int = int(unsafe.Sizeof(int(0)))

type RWTestStruct1 struct {
	Id1   int
	Name1 string
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

type RWTestStruct2 struct {
	Id   int
	Name string
	Cash float32
	Date time.Time
	RWTestStruct1
}

func main() {
	fmt.Println("Hello World 1122!")
	//o := &RWTestStruct1{1, "test"}
	//rw := GetFastRWer(o)
	//fmt.Println(rw)

	var a *RWTestStruct2 = &RWTestStruct2{}
	fmt.Println(a)
	fmt.Println(uintptr(unsafe.Pointer(a)))

	b := unsafe.Pointer(&a)
	fmt.Println("b=", b)

	d := *((*[ptrSize]byte)(b))
	fmt.Println("d = ", d)
	fmt.Println("step 2")

	v := reflect.Indirect(reflect.ValueOf(*a))
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		fType := t.Field(i)
		f := v.Field(i)
		fmt.Println(fType.Name, f.Type().Size(), f.Type())
	}

	benchmarkFastRWerSetValueByName()
	flag.Parse()

	////if *cpuprofile != "" {
	//f, err := os.Create("profile_file")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()
	//benchmarkFastRWerGet(500000)
	//benchmarkFastRWerGetValue(500000)
	////}
	////var s string
	////s = nil
	////fmt.Println(s)

	testFuture()

	testCompare()

	c := make(chan int)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				s, ok := err.(error)
				if ok {
					e := errors.New(s.Error())
					fmt.Println(e.Error())
				} else {
					e := errors.New("")
					fmt.Println(e.Error())
				}
			}

		}()
		c <- 1
		time.Sleep(1 * time.Second)
		c <- 2

	}()

	fmt.Println(<-c)
	close(c)
	time.Sleep(2 * time.Second)
}

func benchmarkFastRWerGet(n int) {
	o := &RWTestStruct2{1, "test", 1.1, time.Now(), RWTestStruct1{}}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	//var id int
	//var name string
	for i := 0; i < n; i++ {
		_ = *((*int)(rw.Ptr(p, 0)))
		_ = *((*string)(rw.Ptr(p, 1)))
		_ = *((*float32)(rw.Ptr(p, 2)))
		_ = *((*time.Time)(rw.Ptr(p, 3)))
	}
	//b.Log(id, name)
}

func benchmarkFastRWerGetValue(n int) {
	o := &RWTestStruct2{1, "test", 1.1, time.Now(), RWTestStruct1{}}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	for i := 0; i < n; i++ {
		_ = rw.Value(p, 0)
		_ = rw.Value(p, 1)
		_ = rw.Value(p, 2)
		_ = rw.Value(p, 3)
	}

}

func benchmarkFastRWerSetValueByName() {
	o := &RWTestStruct2{}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	id := 1111111
	name := "test unsafe set, great!"
	cash := 22.22
	date := time.Now()
	ptr := &RWTestStruct1{1, "test"}
	//b.ResetTimer()
	//for i := 0; i < b.N; i++ {
	rw.SetValueByName(p, "Id", id)
	rw.SetValueByName(p, "Name", name)
	rw.SetValueByName(p, "Cash", cash)
	fmt.Println("SetValueByName", uintptr(p), date, uintptr(unsafe.Pointer(&date)))
	rw.SetValueByName(p, "Date", date)
	rw.SetValueByName(p, "Ptr", ptr)
	//}
}

type st1 struct {
	a int
}

type st2 struct {
	a int
	b map[int]int
}

func testCompare() {
	f := func() {

	}
	m1 := make(map[int]int)
	m2 := make(map[int]int)
	arr1 := [2]int{1, 2}
	arr2 := [2]int{1, 2}
	ch := make(chan int)
	sl1 := []int{1, 2}
	sl2 := []int{1, 2}
	var i interface{} = m1
	testdatas := [][]interface{}{{"a", "a"}, {1, 1}, {arr1, arr2}, {m1, m2}, {f, f}, {sl1, sl2}, {ch, ch}, {&m1, &m2}, {i, i}, {st1{}, st1{}}, {st2{}, st2{}}}
	for i, d := range testdatas {
		r := func(a interface{}, b interface{}) (r bool) {
			defer func() {
				if e := recover(); e != nil {
					fmt.Println(e)
					r = false
				}
			}()
			r = a == b
			return
		}(d[0], d[1])
		fmt.Println(reflect.TypeOf(d[0]), d[0], "=", d[1], r)
		fmt.Println(i, d)
		fmt.Println()
	}
	//测试结果：
	//uncomparable type：map, func, slice, 以及包含这些类型的struct
	//其他类型可以比较，并且比较的是变量的byte数组内容，所以2个不同的数组只要内容相同就是相等
}
