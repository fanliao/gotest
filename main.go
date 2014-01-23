package main

import (
	"flag"
	"fmt"
	//"log"
	//"os"
	//"runtime/pprof"
	"container/list"
	"reflect"
	"sync"
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

	//}
	//var s string
	//s = nil
	//fmt.Println(s)
	testChan()
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

type futureResult struct {
	result []interface{}
	ok     bool
}

//Future代表一个异步任务
type Future struct {
	lock   *sync.Mutex
	chIn   chan *futureResult
	chOut  chan *futureResult
	dones  *list.List
	fails  *list.List
	always *list.List
	pipe   func(v ...interface{}) *Future
	r      *futureResult
}

//Get函数将一直阻塞直到任务完成
func (this *Future) Get() ([]interface{}, bool) {
	if r, ok := <-this.chOut; ok {
		return r.result, true
	} else {
		return nil, false
	}
}

func (this *Future) Reslove(v ...interface{}) {
	r := &futureResult{v, true}
	this.chIn <- r
	close(this.chIn)
}

func (this *Future) Reject(v ...interface{}) {
	r := &futureResult{v, false}
	this.chIn <- r
	close(this.chIn)
}

func (this *Future) Done(callback func(v ...interface{})) *Future {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.r != nil {
		if this.r.ok {
			callback(this.r.result...)
		}
	} else {
		this.dones.PushBack(callback)
	}
	return this
}

func (this *Future) Fail(callback func(v ...interface{})) *Future {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.r != nil {
		if !this.r.ok {
			callback(this.r.result...)
		}
	} else {
		this.fails.PushBack(callback)
	}
	return this
}

func (this *Future) Always(callback func(v ...interface{})) *Future {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.r != nil {
		callback(this.r.result...)
	} else {
		this.always.PushBack(callback)
	}
	return this
}

//Todo
func (this *Future) Pipe(callback func(v ...interface{}) *Future) *Future {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.r != nil {
		this.pipe = callback
		f := this.pipe(this.r.result...)
		return f
	} else {
		this.pipe = callback
		return nil
	}

}

func (this *Future) start() {
	r := <-this.chIn
	this.callback(r)
	this.chOut <- r
	close(this.chOut)
	fmt.Println("is received")
}

func (this *Future) callback(r *futureResult) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.r = r
	fmt.Println("callback")
	if r.ok {
		for e := this.dones.Front(); e != nil; e = e.Next() {
			f := e.Value.(func(v ...interface{}))
			f(r.result...)
		}
	} else {
		for e := this.fails.Front(); e != nil; e = e.Next() {
			f := e.Value.(func(v ...interface{}))
			f(r.result...)
		}
	}
	for e := this.always.Front(); e != nil; e = e.Next() {
		f := e.Value.(func(v ...interface{}))
		f(r.result...)
	}
}

func NewFuture() *Future {
	f := &Future{new(sync.Mutex),
		make(chan *futureResult, 1),
		make(chan *futureResult, 1),
		list.New(),
		list.New(),
		list.New(),
		nil, nil}
	go func() {
		f.start()
	}()
	return f
}

func When(fs ...*Future) *Future {
	f := NewFuture()
	go func() {
		rs := make([][]interface{}, len(fs))
		allOk := true
		for _, f := range fs {
			r, ok := f.Get()
			rs = append(rs, r)
			if !ok {
				allOk = false
			}
		}
		if allOk {
			f.Reslove(rs)
		} else {
			f.Reject(rs)
		}
	}()

	return f
}

func task() *Future {
	c := func(v ...interface{}) {
		fmt.Println("callback", v)
	}
	f := NewFuture().Done(c)
	go func() {
		time.Sleep(1 * time.Second)
		f.Reslove(10)
		fmt.Println("send done")
	}()
	fmt.Println("end start")
	return f
}
func testChan() {
	f := task()

	fmt.Println("begin receive")
	time.Sleep(2 * time.Second)
	r, ok := f.Get()
	fmt.Println("receive", r, ok)
	r, ok = f.Get()
	fmt.Println("receive again", r, ok)

}
