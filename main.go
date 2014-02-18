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

type RWTestStruct3 struct {
	Id1   int
	Name1 string
	a     int
	RWTestStruct2
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
	//fmt.Println("Hello World 1122!")
	////o := &RWTestStruct1{1, "test"}
	////rw := GetFastRWer(o)
	////fmt.Println(rw)

	//var a *RWTestStruct2 = &RWTestStruct2{}
	//fmt.Println(a)
	//fmt.Println(uintptr(unsafe.Pointer(a)))

	//b := unsafe.Pointer(&a)
	//fmt.Println("b=", b)

	//d := *((*[ptrSize]byte)(b))
	//fmt.Println("d = ", d)
	//fmt.Println("step 2")

	//v := reflect.Indirect(reflect.ValueOf(*a))
	//t := v.Type()

	//for i := 0; i < t.NumField(); i++ {
	//	fType := t.Field(i)
	//	f := v.Field(i)
	//	fmt.Println(fType.Name, f.Type().Size(), f.Type())
	//}

	//benchmarkFastRWerSetValueByName()
	//flag.Parse()

	//f := 22.22
	//fmt.Printf("float64 二进制：%b\n", f)
	//fmt.Printf("float32 二进制：%b\n", float32(f))
	//////if *cpuprofile != "" {
	////f, err := os.Create("profile_file")
	////if err != nil {
	////	log.Fatal(err)
	////}
	////pprof.StartCPUProfile(f)
	////defer pprof.StopCPUProfile()
	//benchmarkFastRWerGet(500000)
	////benchmarkFastRWerGetValue(500000)
	//////}
	//////var s string
	//////s = nil
	//////fmt.Println(s)

	////testFuture()

	i := 12
	fmt.Println("type of unsafe.pointer is", reflect.TypeOf(unsafe.Pointer(&i)).Kind())
	//testStructUnsafeCode()
	testCompare()
	testKind()

	testRWUnexported()
	testUnexportedFields(RWTestStruct3{})

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

func testUnexportedFields(o interface{}) {
	fmt.Println("testUnexportedFields")
	fs := faceToStruct(o)
	s := *((*structType)(unsafe.Pointer(fs.typ)))
	f := s.fields[3]
	var name string
	if f.name == nil {
		name = *(f.typ.string)
	} else {
		name = *(f.name)
	}
	fmt.Println(len(s.fields))
	fmt.Println(f.typ.string, f.name, name, "\n")
}

func faceAreEqual(a interface{}, b interface{}) (r bool) {
	fmt.Println("faceAreEqual?", a == b)
	return a == b
}

func testKind() {
	fmt.Println("testKind", flagRO, flagIndir, flagAddr, flagMethod,
		flagKindShift, flagKindWidth, flagKindMask, flagMethodShift, "\n")
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

func testRWUnexported() {
	fmt.Println()
	s1 := RWTestStruct1{10, "hello", 20}
	p := unsafe.Pointer(&s1)
	//rwer1 := GetFastRWer_bak(&s1)
	rwer2 := GetFastRWer(s1)
	printRWer := func(rw *FastRW) {
		for i := 0; i < len(rw.FieldNamesByIndex); i++ {
			fmt.Println(rw.FieldNamesByIndex[i], "size is", rw.FieldsByIndex[i].typ.size, "offset is", rw.FieldsByIndex[i].offset) //.FieldOffsetsByIndex[i])
		}
	}
	//fmt.Println("GetFastRWer")
	////反射不能获取未导出的字段
	//printRWer(rwer1)
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

type st struct {
	a int
}

type st1 struct {
	A int
	b int
}

type st2 struct {
	A int
	b map[int]int
}

type st3 struct {
	A int
	s *st3
}

func testCompare() {
	//f := func() {}
	//m1 := make(map[int]int)
	//m2 := make(map[int]int)
	//m3 := map[int]int{1: 1, 2: 2}
	//m4 := map[int]int{1: 1, 2: 2}
	//arr1 := [2]int{1, 2}
	//arr2 := [2]int{1, 2}
	//ch := make(chan int)
	//sl1 := []int{1, 2}
	//sl2 := []int{1, 2}
	//var i1 interface{} = m1
	//var i2 interface{} = m2
	//st11 := st1{2, 3454}
	//st12 := st1{2, 3454}

	//st21, st22 := st2{}, st2{}
	//st21.b = map[int]int{1: 1, 2: 2}
	//st22.b = map[int]int{1: 1, 2: 2}

	st31, st32 := st3{1, nil}, st3{1, nil}
	st31.s = &st31
	st32.s = &st32

	testdatas := [][]interface{}{
		//{nil, nil, false, true}, //测试nil与nil
		//{nil, "a", false, false},
		//{"a", "a", false, true},
		//{1, "1", false, false}, //不进行类型转换
		//{1, 1, false, true},
		//{arr1, arr2, false, true}, //equals比较内容，所以返回true
		//{m1, m2, false, true},
		//{m3, m4, false, true},
		//{f, f, false, true},
		//{sl1, sl2, false, true},
		//{ch, ch, false, true},
		//{&m1, &m2, false, false},
		//{&m1, &m2, true, true}, //*
		//{&m1, &m1, false, true},
		//{&m1, &m1, true, true}, //*
		//{i1, i1, false, true},
		//{i1, i2, false, true},
		//{i1, i2, true, true},
		//{st{64}, st{64}, false, true},
		//{st11, st12, false, true},
		//{st1{2, 200}, st1{2, 100}, false, false},
		//{st2{1, m1}, st2{1, m1}, false, true},
		//{st2{1, m1}, st2{1, m2}, true, true},
		//{st21, st22, false, false}, //不进行深度比较的情况下，st21与st22应该不相等，因为2个包含的map对象不同，虽然map的内容相同
		//{st21, st22, true, true},   //进行深度比较的情况下，st21与st22应该相等，因为2个包含的map的内容相同
		//{&st11, &st12, false, false},
		//{&st11, &st12, true, true},
		//{&st31, &st32, false, false},
		{&st31, &st32, true, true}, //检查递归关联的处理
		{st31, st32, true, true},   //检查递归关联的处理
	}
	checkError := func(e interface{}) bool {
		//err := e.(error)
		if e != nil {
			fmt.Println("error return false2", e)
			return true
		} else {
			return false
		}
	}
	f1 := func(a interface{}, b interface{}, deep bool, test func(a interface{}, b interface{}, deep bool) (bool, bool)) (r bool, valid bool) {
		defer func() {
			if e := recover(); checkError(e) {
				fmt.Println("error return false")
				r = false
			}
		}()
		//r = a == b
		r, valid = test(a, b, deep)
		return
	}

	f3 := func(title string, test func(a interface{}, b interface{}, deep bool) (bool, bool)) {
		fmt.Println(title)
		for i, d := range testdatas {
			r, valid := f1(d[0], d[1], d[2].(bool), test)
			if valid && !AreEqual(r, d[3], nil) {
				fmt.Println(i, reflect.TypeOf(d[0]), d[0], "=", d[1], r)
			}
		}
		fmt.Println()
	}

	f3("test equal function", func(a interface{}, b interface{}, deep bool) (bool, bool) {
		return equals(a, b, deep), true
	})

	f3("test reflect.DeepEqual function", func(a interface{}, b interface{}, deep bool) (bool, bool) {
		if deep {
			return reflect.DeepEqual(a, b), true
		} else {
			return false, false
		}
	})

	//fmt.Println(st21, st22, reflect.DeepEqual(st21, st22))
	//error, but why DeepEuqal don't throw error?
	//fmt.Println(i1, i2, i1 == i2)

	//fmt.Println("test reflect.DeepEqual function")
	//for i, d := range testdatas {
	//	r := reflect.DeepEqual(d[0], d[1])
	//	fmt.Println(i, reflect.TypeOf(d[0]), d[0], "=", d[1], r)
	//	AreEqual(r, d[2], nil)
	//	fmt.Println()
	//}

	//测试结果：
	//uncomparable type：map, func, slice, 以及包含这些类型的struct
	//其他类型可以比较，并且比较的是变量的byte数组内容，所以2个不同的数组只要内容相同就是相等
}
