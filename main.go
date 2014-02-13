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

type RWTestStruct1 struct {
	Id1   int
	Name1 string
	a     int
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

type RWTestStruct2 struct {
	Id   int
	Name string
	Cash float32
	Date time.Time
	RWTestStruct1
	Ptr *RWTestStruct2
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

	f := 22.22
	fmt.Printf("float64 二进制：%b\n", f)
	fmt.Printf("float32 二进制：%b\n", float32(f))
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

	//testFuture()

	i := 12
	fmt.Println("type of unsafe.pointer is", reflect.TypeOf(unsafe.Pointer(&i)).Kind())
	//testStructUnsafeCode()
	testCompare()
	testGetUnexported()
	//testSetUnexported()

	fmt.Println()
	s1 := RWTestStruct1{10, "hello", 20}
	p := unsafe.Pointer(&s1)
	rwer1 := GetFastRWer(&s1)
	rwer2 := GetFastRWer1(s1)
	printRWer := func(rw *FastRW) {
		for i := 0; i < len(rw.FieldNamesByIndex); i++ {
			fmt.Println(rw.FieldNamesByIndex[i], "size is", rw.FieldSizeByIndex[i], "offset is", rw.FieldOffsetsByIndex[i])
		}
	}
	fmt.Println("GetFastRWer")
	//反射不能获取未导出的字段
	printRWer(rwer1)
	fmt.Println("GetFastRWer1")
	//直接获取rtype可以
	printRWer(rwer2)
	//测试读取未导出字段
	fmt.Println(s1, "is {", *((*int)(rwer2.Ptr(p, 0))), *((*string)(rwer2.Ptr(p, 1))), *((*int)(rwer2.Ptr(p, 2))), "}")
	//测试写入未导出字段
	intF, strF := 11, "aaa"
	rwer2.SetPtr(p, 2, uintptr(unsafe.Pointer(&intF)))
	rwer2.SetPtr(p, 1, uintptr(unsafe.Pointer(&strF)))
	fmt.Println(s1, "is {", *((*int)(rwer2.Ptr(p, 0))), *((*string)(rwer2.Ptr(p, 1))), *((*int)(rwer2.Ptr(p, 2))), "}")
	fmt.Println()

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
	o := &RWTestStruct2{1, "test", 1.1, time.Now(), RWTestStruct1{}, nil}
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
	o := &RWTestStruct2{1, "test", 1.1, time.Now(), RWTestStruct1{}, nil}
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
	var cash float64 = 22.22
	date := time.Now()
	var ptr *RWTestStruct2 = nil //&RWTestStruct1{1, "test"}

	//rw.SetValue(p, 0, id)
	//rw.SetValue(p, 1, &name)
	//rw.SetValue(p, 2, cash)
	//rw.SetValue(p, 3, &date)
	//rw.SetValue(p, 4, ptr)

	rw.SetValueByName(p, "Id", id)
	rw.SetValueByName(p, "Name", &name)
	rw.SetValueByName(p, "Cash", float32(cash))
	rw.SetValueByName(p, "Date", &date)
	rw.SetValueByName(p, "Ptr", ptr)

	fmt.Println("SetValueByName", o)
	AreEqual(nil, o.Ptr, nil)
	printInterfaceLayout(nil)
	printInterfaceLayout(o.Ptr)
	faceAreEqual(nil, o.Ptr)
	fmt.Println("nil == ptr of nil?", o.Ptr == nil)
	printInterfaceLayout(RWTestStruct2{})

}

func faceAreEqual(a interface{}, b interface{}) (r bool) {
	fmt.Println("faceAreEqual?", a == b)
	return a == b
}

func printInterfaceLayout(a interface{}) {
	fmt.Println("printInterfaceLayout", a, "isnil?", a == nil)
	s := *((*interfaceHeader)(unsafe.Pointer(&a)))
	if uintptr(unsafe.Pointer(s.typ)) != 0 {
		fmt.Println(s, *(s.typ), *(*s.typ).string)
		if s.typ.Kind() == reflect.Struct {
			tt := (*structType)(unsafe.Pointer(s.typ))
			fmt.Println("struct layout", *tt, "\n")
		}
		p := s.typ.ptrToThis
		for uintptr(unsafe.Pointer(p)) != 0 {
			fmt.Println("this is pointer to", *(p), *p.string)
			p = p.ptrToThis
		}
		//if uintptr(unsafe.Pointer((s.typ.ptrToThis))) != 0 {
		//	fmt.Println("this is pointer to", *(s.typ.ptrToThis), *(*s.typ).string)
		//}
	} else {
		fmt.Println(s)
	}
	fmt.Println()

}

type st struct {
	a int
}

type st1 struct {
	a int
	b int
}

type st2 struct {
	a int
	b map[int]int
}

func testGetUnexported() {
	st11 := st1{2, 3454}
	//panic: reflect.Value.UnsafeAddr of unaddressable value
	aOffset := reflect.ValueOf(st11).Field(1)
	fmt.Println(aOffset.CanAddr(), aOffset.CanInterface(), aOffset.CanSet())
	//panic: reflect.Value.Interface: cannot return value obtained from unexported field or method
	//fmt.Println(aOffset.Interface())
}

func testSetUnexported() {
	st11 := st1{2, 3454}
	//panic: reflect.Value.UnsafeAddr of unaddressable value
	aOffset := reflect.ValueOf(st11).Field(1).UnsafeAddr()
	pint := (*int)(unsafe.Pointer(aOffset))
	*pint = 110
	fmt.Println(st11)
}

func testStructUnsafeCode() {
	st11 := st1{2, 3454}
	st12 := st1{2, 3454}
	fmt.Println("st", uintptr(unsafe.Pointer(&st11)), uintptr(unsafe.Pointer(&st12)))
	var i1, i2 interface{}
	i1, i2 = st11, st12
	fmt.Println(*((*interfaceHeader1)(unsafe.Pointer(&i1))),
		*((*interfaceHeader1)(unsafe.Pointer(&i2))))
	word1 := (*((*interfaceHeader1)(unsafe.Pointer(&i1)))).word
	ptr1 := uintptr(unsafe.Pointer(&st11))
	for i := 0; uintptr(i) < unsafe.Sizeof(st11); i++ {
		fmt.Println(*((*byte)(unsafe.Pointer(ptr1 + uintptr(i)))))
	}
	fmt.Println(*((*st1)(unsafe.Pointer(word1))))
	fmt.Println(i1, i2)
	fmt.Println()

	fmt.Println("direct compare struct")
	ptr := InterfaceToPtr1(st11)
	fmt.Println(ptr)
	for i := 0; uintptr(i) < unsafe.Sizeof(st11); i++ {
		fmt.Println(*((*byte)(unsafe.Pointer(ptr + uintptr(i)))))
	}

	fmt.Println()
}

func testCompare() {
	m11 := map[int]int{1: 1}
	m22 := map[int]int64{1: 1}
	fmt.Println(reflect.TypeOf(m11), reflect.TypeOf(m22), reflect.TypeOf(m11) == reflect.TypeOf(m22))

	f := func() {

	}
	m1 := make(map[int]int)
	m2 := make(map[int]int)
	m3 := map[int]int{1: 1, 2: 2}
	m4 := map[int]int{1: 1, 2: 2}
	arr1 := [2]int{1, 2}
	arr2 := [2]int{1, 2}
	ch := make(chan int)
	sl1 := []int{1, 2}
	sl2 := []int{1, 2}
	var i interface{} = m1
	st11 := st1{2, 3454}
	st12 := st1{2, 3454}

	testdatas := [][]interface{}{
		{nil, nil, true},
		{nil, "a", false},
		{"a", "a", true},
		{1, "1", false},
		{1, 1, true},
		{arr1, arr2, true},
		{m1, m2, true},
		{m3, m4, true},
		{f, f, true},
		{sl1, sl2, true},
		{ch, ch, true},
		{&m1, &m2, false},
		{&m1, &m1, true},
		{i, i, true},
		{st{64}, st{64}, true},
		{st11, st12, true},
		{st2{1, m1}, st2{1, m1}, true},
		{st2{1, m1}, st2{1, m2}, false},
		{&st11, &st12, false},
	}
	f1 := func(a interface{}, b interface{}) (r bool) {
		defer func() {
			if e := recover(); e != nil {
				fmt.Println(e)
				r = false
			}
		}()
		//r = a == b
		r = equals(a, b)
		return
	}
	for i, d := range testdatas {
		r := f1(d[0], d[1])
		fmt.Println(i, reflect.TypeOf(d[0]), d[0], "=", d[1], r)
		AreEqual(r, d[2], nil)
		fmt.Println()
	}

	//测试结果：
	//uncomparable type：map, func, slice, 以及包含这些类型的struct
	//其他类型可以比较，并且比较的是变量的byte数组内容，所以2个不同的数组只要内容相同就是相等
}
