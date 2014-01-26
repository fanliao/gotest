package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type stringer interface {
	String() string
}

func getError(i interface{}) (e error) {
	if i != nil {
		switch v := i.(type) {
		case error:
			e = v
		case stringer:
			e = errors.New(v.String())
		default:
			e = errors.New("unknow error")
		}
	}
	return
}

type callbackType int

const (
	CALLBACK_DONE callbackType = iota
	CALLBACK_FAIL
	CALLBACK_ALWAYS
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

//Get函数将一直阻塞直到任务完成,返回任务的结果
//只能Get一次，多次Get将返回nil, false
func (this *Future) Get() (r []interface{}, ok bool) {
	if fr, ok := <-this.chOut; ok {
		r = fr.result
	} else {
		r = nil
	}
	return
}

//Reslove表示任务已经正常完成
func (this *Future) Reslove(v ...interface{}) (e error) {
	defer func() {
		e = getError(recover())
	}()
	r := &futureResult{v, true}
	this.chIn <- r
	close(this.chIn)
	e = nil
	return
}

func (this *Future) Reject(v ...interface{}) (e error) {
	defer func() {
		e = getError(recover())
	}()
	r := &futureResult{v, false}
	this.chIn <- r
	close(this.chIn)
	return
}

func (this *Future) Done(callback func(v ...interface{})) *Future {
	this.handleOneCallback(callback, CALLBACK_DONE)
	return this
}

func (this *Future) Fail(callback func(v ...interface{})) *Future {
	this.handleOneCallback(callback, CALLBACK_FAIL)
	return this
}

func (this *Future) Always(callback func(v ...interface{})) *Future {
	this.handleOneCallback(callback, CALLBACK_ALWAYS)
	return this
}

//
func (this *Future) Then(callback func(v ...interface{}) *Future) *Future {
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
	this.execCallback(r)
	if this.pipeTask == nil {
		this.chOut <- r
	} else {
		//下面触发pipe的Future任务，但如果在之后调用pipeFuture的Done, Fail, Always，如何处理？
		target := this.pipeTask(this.r.result...)

		f := target.batchCallback(this.pipeFuture.dones, this.pipeFuture.fails, this.pipeFuture.always)
		if f != nil {
			f()
		}
		this.pipeFuture.targetFuture = target
	}
	close(this.chOut)
	fmt.Println("is received")
}

func (this *Future) execCallback(r *futureResult) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.r = r
	fmt.Println("callback")
	execCallback(r, this.dones, this.fails, this.always)
}

func execCallback(r *futureResult, dones []func(v ...interface{}), fails []func(v ...interface{}), always []func(v ...interface{})) {

	var callbacks []func(v ...interface{})
	if r.ok {
		callbacks = dones
	} else {
		callbacks = fails

	}

	forFs := func(s []func(v ...interface{})) {
		forSlice(s, func(f func(v ...interface{})) { f(r.result...) })
	}

	forFs(callbacks)
	forFs(always)

}

func (this *Future) batchCallback(dones []func(v ...interface{}), fails []func(v ...interface{}), always []func(v ...interface{})) func() {
	proxyAction := func(t *Future) {
		f := t.batchCallback(dones, fails, always)
		if f != nil {
			f()
		}
	}
	pendingAction := func() {
		this.dones = append(this.dones, dones...)
		this.fails = append(this.fails, fails...)
		this.always = append(this.always, always...)
	}
	finalAction := func(r *futureResult) {
		execCallback(r, dones, fails, always)
	}
	return this.addCallback(proxyAction, pendingAction, finalAction)
}

func (this *Future) handleOneCallback(callback func(v ...interface{}), t callbackType) {
	f := this.addOneCallback(callback, CALLBACK_DONE)
	if f != nil {
		f()
	}
}

func (this *Future) addOneCallback(callback func(v ...interface{}), t callbackType) func() {
	proxyAction := func(target *Future) {
		target.addOneCallback(callback, t)
	}
	pendingAction := func() {
		switch t {
		case CALLBACK_DONE:
			this.dones = append(this.dones, callback)
		case CALLBACK_FAIL:
			this.fails = append(this.fails, callback)
		case CALLBACK_ALWAYS:
			this.always = append(this.always, callback)
		}
	}
	finalAction := func(r *futureResult) {
		if (t == CALLBACK_DONE && r.ok) ||
			(t == CALLBACK_FAIL && !r.ok) ||
			(t == CALLBACK_ALWAYS) {
			callback(r.result...)
		}
	}
	return this.addCallback(proxyAction, pendingAction, finalAction)
}

func (this *Future) addCallback(proxyAction func(*Future), pendingAction func(), finalAction func(*futureResult)) func() {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.targetFuture != nil {
		target := this.targetFuture
		return func() {
			proxyAction(target)
		}
	}

	if this.r == nil {
		pendingAction()
		//this.dones = append(this.dones, callback)
		return nil
	} else {
		r := this.r
		return func() {
			finalAction(r)
		}
	}
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

func Any(fs ...*Future) *Future {
	f := NewFuture()

	for _, f := range fs {
		f.Done(func(v ...interface{}) {
			f.Reslove(v...)
		}).Fail(func(v ...interface{}) {
			f.Reject(v...)
		})
	}

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
