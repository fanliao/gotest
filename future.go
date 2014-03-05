package main

import (
	"errors"
	"sync"
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

type pipe struct {
	pipeDoneTask, pipeFailTask func(v ...interface{}) *Future
	pipeFuture                 *Future
}

//Future代表一个异步任务
type Future struct {
	lock                 *sync.Mutex
	chIn, chOut          chan *futureResult
	dones, fails, always []func(v ...interface{})
	pipe
	targetFuture *Future
	r            *futureResult
}

//Get函数将一直阻塞直到任务完成,返回任务的结果
//如果Get多次，后续的Get将直接返回任务结果
func (this *Future) Get() ([]interface{}, bool) {
	if this.targetFuture != nil {
		return this.targetFuture.Get()
	}

	if fr, ok := <-this.chOut; ok {
		return fr.result, fr.ok
	} else {
		r, ok := this.r.result, this.r.ok
		return r, ok
	}
}

//Reslove表示任务正常完成
func (this *Future) Reslove(v ...interface{}) (e error) {
	defer func() {
		e = getError(recover())
	}()
	r := &futureResult{v, true}
	this.chIn <- r
	e = nil
	//a := new(sync.Once)
	//a.Do(func() {
	//	r := &futureResult{v, true}
	//	this.chIn <- r
	//	e = nil
	//})
	return
}

//Reject表示任务失败
func (this *Future) Reject(v ...interface{}) (e error) {
	defer func() {
		e = getError(recover())
	}()
	r := &futureResult{v, false}
	this.chIn <- r
	e = nil
	return
}

//添加一个任务成功完成的回调，如果任务已经成功完成，则直接执行回调函数
//传递给Done函数的参数与Reslove函数的参数相同
func (this *Future) Done(callback func(v ...interface{})) *Future {
	this.handleOneCallback(callback, CALLBACK_DONE)
	return this
}

//添加一个任务失败的回调，如果任务已经失败，则直接执行回调函数
//传递给Fail函数的参数与Reject函数的参数相同
func (this *Future) Fail(callback func(v ...interface{})) *Future {
	this.handleOneCallback(callback, CALLBACK_FAIL)
	return this
}

//添加一个回调函数，该函数将在任务完成后执行，无论成功或失败
//传递给Always回调的参数根据成功或失败状态，与Reslove或Reject函数的参数相同
func (this *Future) Always(callback func(v ...interface{})) *Future {
	this.handleOneCallback(callback, CALLBACK_ALWAYS)
	return this
}

//for then api, the new Future object will be return
//New future task object should be started after current future be done or failed
//链式添加异步任务，可以同时定制Done或Fail状态下的链式异步任务，并返回一个新的异步对象。如果对此对象执行Done，Fail，Always操作，则新的回调函数将会被添加到链式的异步对象中
//如果调用的参数超过2个，那第2个以后的参数将会被忽略
func (this *Future) Then(callbacks ...(func(v ...interface{}) *Future)) *Future {
	this.lock.Lock()
	defer this.lock.Unlock()
	if len(callbacks) == 0 ||
		(len(callbacks) == 1 && callbacks[0] == nil) ||
		(len(callbacks) > 1 && callbacks[0] == nil && callbacks[1] == nil) {
		return this
	}
	if this.r != nil {
		//this.pipeTasks = callbacks
		f := this

		if this.r.ok && callbacks[0] != nil {
			f = (callbacks[0])(this.r.result...)
		} else if !this.r.ok && len(callbacks) > 1 && callbacks[1] != nil {
			f = (callbacks[1])(this.r.result...)
		}
		return f
	} else {
		this.pipeDoneTask = callbacks[0]
		if len(callbacks) > 1 {
			this.pipeFailTask = callbacks[1]
		}
		this.pipeFuture = NewFuture()
		return this.pipeFuture
	}

}

//返回3个与链式调用相关的对象
func (this *Future) getPipe() (func(v ...interface{}) *Future, func(v ...interface{}) *Future, *Future) {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.pipeDoneTask, this.pipeFailTask, this.pipeFuture
}

//启动一个异步goroutine监控任务完成
func (this *Future) start() {
	r := <-this.chIn
	close(this.chIn)
	this.setResult(r)

	//让Get函数可以返回
	this.chOut <- r
	close(this.chOut)

	//任务完成后调用回调函数
	execCallback(r, this.dones, this.fails, this.always)

	//处理链式异步任务
	pipeDoneTask, pipeFailTask, pipeFuture := this.getPipe()
	if pipeDoneTask != nil || pipeFailTask != nil {
		//下面触发pipe的Future任务，但如果在之后调用pipeFuture的Done, Fail, Always，将直接转发到真正的Future对象
		var target *Future
		if r.ok && pipeDoneTask != nil {
			target = pipeDoneTask(r.result...)
		} else if !r.ok && pipeFailTask != nil {
			target = pipeFailTask(r.result...)
		}
		//target := this.pipeTask(this.r.result...)

		if target != nil {
			dones, fails, always := this.pipeFuture.dones, this.pipeFuture.fails, this.pipeFuture.always
			this.pipeFuture.dones = make([]func(v ...interface{}), 0, 0)
			this.pipeFuture.fails = make([]func(v ...interface{}), 0, 0)
			this.pipeFuture.always = make([]func(v ...interface{}), 0, 0)

			pipeFuture.targetFuture = target
			target.Done(func(v ...interface{}) {
				pipeFuture.Reslove(v...)
			}).Fail(func(v ...interface{}) {
				pipeFuture.Reject(v...)
			})
			f := target.batchCallback(dones, fails, always)
			if f != nil {
				f()
			}
		}
	}
}

//set this.r
func (this *Future) setResult(r *futureResult) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.r = r
}

//执行回调函数，利用lock保证访问this.r是线程安全的
func (this *Future) execCallback(r *futureResult) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.r = r
	execCallback(r, this.dones, this.fails, this.always)
}

//执行回调函数
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

//批量添加回调函数
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

//处理单个回调函数的添加请求p
func (this *Future) handleOneCallback(callback func(v ...interface{}), t callbackType) {
	f := this.addOneCallback(callback, t)
	if f != nil {
		f()
	}
}

//添加一个回调函数
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

//添加回调函数的框架函数
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

//异步执行一个函数。如果最后一个返回值为bool，则将认为此值代表异步任务成功或失败。如果函数抛出error，则认为异步任务失败
func Submit(action func() []interface{}) *Future {
	fu := NewFuture()

	go func() {
		defer func() {
			if e := recover(); e != nil {
				fu.Reject(e)
			}
		}()

		r := action()
		if l := len(r); l > 0 {
			if done, ok := r[l-1].(bool); ok {
				if done {
					fu.Reslove(r[:l-1]...)
				} else {
					fu.Reject(r[:l-1]...)
				}

			}
		}
		fu.Reslove(r...)
	}()

	return fu
}

func Submit0(action func()) *Future {
	return Submit(func() []interface{} {
		action()
		return make([]interface{}, 0, 0)
	})
}

//Factory function for future
func NewFuture() *Future {
	f := &Future{new(sync.Mutex),
		make(chan *futureResult, 1),
		make(chan *futureResult, 1),
		make([]func(v ...interface{}), 0, 8),
		make([]func(v ...interface{}), 0, 8),
		make([]func(v ...interface{}), 0, 4),
		pipe{}, nil, nil}
	go func() {
		f.start()
	}()
	return f
}

//产生一个新的Future，如果列表中任意1个Future完成，则Future完成
func Any(fs ...*Future) *Future {
	nf := NewFuture()

	for _, f := range fs {
		f.Done(func(v ...interface{}) {
			nf.Reslove(v...)
		}).Fail(func(v ...interface{}) {
			nf.Reject(v...)
		})
	}

	return nf
}

//产生一个新的Future，如果列表中所有Future都成功完成，则Future成功完成，否则失败
func When(fs ...*Future) *Future {
	f := NewFuture()
	go func() {
		rs := make([][]interface{}, len(fs))
		allOk := true
		for _, f := range fs {
			r, ok := f.Get()
			r = append(r, ok)
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

func forSlice(s []func(v ...interface{}), f func(func(v ...interface{}))) {
	for _, e := range s {
		f(e)
	}
}
