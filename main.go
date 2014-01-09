package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime/pprof"
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

	//if *cpuprofile != "" {
	f, err := os.Create("profile_file")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	benchmarkFastRWerGet(500000)
	benchmarkFastRWerGetValue(500000)
	//}
	//var s string
	//s = nil
	//fmt.Println(s)
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
