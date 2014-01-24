package main

import (
	"fmt"
	"sync"
	"time"
)

//代表异步任务的结果
type futureResult struct {
	result []interface{}
	ok     bool
}

//Future代表一个异步任务
type Future struct {
	lock                 *sync.Mutex
	chIn, chOut          chan *futureResult
	dones, fails, always []func(v ...interface{})
	pipeTask             func(v ...interface{}) *Future
	pipeFuture           *Future
	targetFuture         *Future
	r                    *futureResult
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

	if this.targetFuture != nil {
		this.targetFuture.Done(callback)
		return this
	}
	if this.r != nil {
		if this.r.ok {
			callback(this.r.result...)
		}
	} else {
		this.dones = append(this.dones, callback)
	}
	return this
}

func (this *Future) Fail(callback func(v ...interface{})) *Future {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.targetFuture != nil {
		this.targetFuture.Fail(callback)
		return this
	}
	if this.r != nil {
		if !this.r.ok {
			callback(this.r.result...)
		}
	} else {
		this.fails = append(this.fails, callback)
	}
	return this
}

func (this *Future) Always(callback func(v ...interface{})) *Future {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.targetFuture != nil {
		this.targetFuture.Always(callback)
		return this
	}
	if this.r != nil {
		callback(this.r.result...)
	} else {
		this.always = append(this.always, callback)
	}
	return this
}

//
func (this *Future) Pipe(callback func(v ...interface{}) *Future) *Future {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.r != nil {
		this.pipeTask = callback
		f := this.pipeTask(this.r.result...)
		return f
	} else {
		this.pipeTask = callback
		this.pipeFuture = NewFuture()
		return this.pipeFuture
	}

}

func (this *Future) start() {
	r := <-this.chIn
	this.callback(r)
	if this.pipeTask == nil {
		this.chOut <- r
	} else {
		//下面触发pipe的Future任务，但如果在之后调用pipeFuture的Done, Fail, Always，如何处理？
		f := this.pipeTask(this.r.result...)

		forSlice(this.pipeFuture.dones, func(e func(v ...interface{})) { f.Done(e) })
		forSlice(this.pipeFuture.fails, func(e func(v ...interface{})) { f.Fail(e) })
		forSlice(this.pipeFuture.always, func(e func(v ...interface{})) { f.Always(e) })
		this.pipeFuture.targetFuture = f
		//f.Done(this.pipeFuture.dones...)
		//f.Fail(this.pipeFuture.fails...)
		//f.Always(this.pipeFuture.always...)
	}
	close(this.chOut)
	fmt.Println("is received")
}

func (this *Future) callback(r *futureResult) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.r = r
	fmt.Println("callback")

	var callbacks []func(v ...interface{})
	if r.ok {
		callbacks = this.dones

	} else {
		callbacks = this.fails

	}

	/*forFuncs := func(s []func(v ...interface{})) {
		for e := funcs.Front(); e != nil; e = e.Next() {
			f := e.Value.(func(v ...interface{}))
			f(r.result...)
		}
	}*/

	forFs := func(s []func(v ...interface{})) {
		forSlice(s, func(f func(v ...interface{})) { f(r.result...) })
	}

	forFs(callbacks)
	forFs(this.always)

}

func NewFuture() *Future {
	f := &Future{new(sync.Mutex),
		make(chan *futureResult, 1),
		make(chan *futureResult, 1),
		make([]func(v ...interface{}), 0, 8),
		make([]func(v ...interface{}), 0, 8),
		make([]func(v ...interface{}), 0, 4),
		nil, nil, nil, nil}
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

func forSlice(s []func(v ...interface{}), f func(func(v ...interface{}))) {
	for _, e := range s {
		f(e)
	}
}
