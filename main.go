package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"
	"unsafe"
)

type RWTestStruct1 struct {
	Id   int
	Name string
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

type RWTestStruct struct {
	Id   int
	Name string
	Cash float32
	Date time.Time
}

func main() {
	fmt.Println("Hello World 1111!")
	//o := &RWTestStruct1{1, "test"}
	//rw := GetFastRWer(o)
	//fmt.Println(rw)

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
	o := &RWTestStruct{1, "test", 1.1, time.Now()}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	//var id int
	//var name string
	for i := 0; i < n; i++ {
		_ = *((*int)(rw.GetPtr(p, 0)))
		_ = *((*string)(rw.GetPtr(p, 1)))
		_ = *((*float32)(rw.GetPtr(p, 2)))
		_ = *((*time.Time)(rw.GetPtr(p, 3)))
	}
	//b.Log(id, name)
}

func benchmarkFastRWerGetValue(n int) {
	o := &RWTestStruct{1, "test", 1.1, time.Now()}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	for i := 0; i < n; i++ {
		_ = rw.GetValue(p, 0)
		_ = rw.GetValue(p, 1)
		_ = rw.GetValue(p, 2)
		_ = rw.GetValue(p, 3)
	}

}
