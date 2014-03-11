package main

import (
	"errors"
	//"fmt"
	"sync"
	"time"
)

const (
	CALLBACK_DONE callbackType = iota
	CALLBACK_FAIL
	CALLBACK_ALWAYS
)

type callbackType int

type stringer interface {
	String() string
}

type anyFutureResult struct {
	result []interface{}
	i      int
}

//代表异步任务的结果
type futureResult struct {
	result []interface{}
	ok     bool
}

type pipe struct {
	pipeDoneTask, pipeFailTask func(v ...interface{}) *PromiseValue
	pipeFuture                 *Future
}

//Future代表一个异步任务
type Future struct {
	onceEnd *sync.Once
	*PromiseValue
	//onceThen             *sync.Once
	//lock                 *sync.Mutex
	//chOut                chan *futureResult
	//dones, fails, always []func(v ...interface{})
	//pipe
	//r          *futureResult
	isPromised bool
}

func (this *Future) Promise() *PromiseValue {
	return this.PromiseValue
}

//添加一个任务成功完成时的回调，如果任务已经成功完成，则直接执行回调函数
//传递给Done函数的参数与Reslove函数的参数相同
func (this *Future) Done(callback func(v ...interface{})) *Future {
	this.PromiseValue.Done(callback) // .handleOneCallback(callback, CALLBACK_DONE)
	return this
}

//添加一个任务失败时的回调，如果任务已经失败，则直接执行回调函数
//传递给Fail函数的参数与Reject函数的参数相同
func (this *Future) Fail(callback func(v ...interface{})) *Future {
	this.PromiseValue.Fail(callback) //.handleOneCallback(callback, CALLBACK_FAIL)
	return this
}

//添加一个回调函数，该函数将在任务完成后执行，无论成功或失败
//传递给Always回调的参数根据成功或失败状态，与Reslove或Reject函数的参数相同
func (this *Future) Always(callback func(v ...interface{})) *Future {
	this.PromiseValue.Always(callback)
	//this.handleOneCallback(callback, CALLBACK_ALWAYS)
	return this
}

//Reslove表示任务正常完成
func (this *Future) Reslove(v ...interface{}) (e error) {
	return this.end(&futureResult{v, true})
}

//Reject表示任务失败
func (this *Future) Reject(v ...interface{}) (e error) {
	return this.end(&futureResult{v, false})
}

//Promise代表一个异步任务
type PromiseValue struct {
	onceThen             *sync.Once
	lock                 *sync.Mutex
	chOut                chan *futureResult
	dones, fails, always []func(v ...interface{})
	pipe
	r *futureResult
}

//Get函数将一直阻塞直到任务完成,并返回任务的结果
//如果任务已经完成，后续的Get将直接返回任务结果
func (this *PromiseValue) Get() ([]interface{}, bool) {
	if fr, ok := <-this.chOut; ok {
		return fr.result, fr.ok
	} else {
		r, ok := this.r.result, this.r.ok
		return r, ok
	}
}

//Get函数将一直阻塞直到任务完成,并返回任务的结果
//如果任务已经完成，后续的Get将直接返回任务结果
func (this *PromiseValue) GetOrTimeout(mm int) ([]interface{}, bool, bool) {
	select {
	case <-time.After((time.Duration)(mm) * time.Millisecond):
		return nil, false, true
	case fr, ok := <-this.chOut:
		if ok {
			return fr.result, fr.ok, false
		} else {
			r, ok := this.r.result, this.r.ok
			return r, ok, false
		}

	}
}

//添加一个任务成功完成时的回调，如果任务已经成功完成，则直接执行回调函数
//传递给Done函数的参数与Reslove函数的参数相同
func (this *PromiseValue) Done(callback func(v ...interface{})) *PromiseValue {
	this.handleOneCallback(callback, CALLBACK_DONE)
	return this
}

//添加一个任务失败时的回调，如果任务已经失败，则直接执行回调函数
//传递给Fail函数的参数与Reject函数的参数相同
func (this *PromiseValue) Fail(callback func(v ...interface{})) *PromiseValue {
	this.handleOneCallback(callback, CALLBACK_FAIL)
	return this
}

//添加一个回调函数，该函数将在任务完成后执行，无论成功或失败
//传递给Always回调的参数根据成功或失败状态，与Reslove或Reject函数的参数相同
func (this *PromiseValue) Always(callback func(v ...interface{})) *PromiseValue {
	this.handleOneCallback(callback, CALLBACK_ALWAYS)
	return this
}

//for then api, the new Future object will be return
//New future task object should be started after current future be done or failed
//链式添加异步任务，可以同时定制Done或Fail状态下的链式异步任务，并返回一个新的异步对象。如果对此对象执行Done，Fail，Always操作，则新的回调函数将会被添加到链式的异步对象中
//如果调用的参数超过2个，那第2个以后的参数将会被忽略
//Then只能调用一次，第一次后的调用将被忽略
func (this *PromiseValue) Then(callbacks ...(func(v ...interface{}) *PromiseValue)) (result *PromiseValue, ok bool) {
	if len(callbacks) == 0 ||
		(len(callbacks) == 1 && callbacks[0] == nil) ||
		(len(callbacks) > 1 && callbacks[0] == nil && callbacks[1] == nil) {
		result = this
		return
	}

	this.onceThen.Do(func() {
		this.lock.Lock()
		defer this.lock.Unlock()
		if this.r != nil {
			f := this

			if this.r.ok && callbacks[0] != nil {
				f = (callbacks[0](this.r.result...))
			} else if !this.r.ok && len(callbacks) > 1 && callbacks[1] != nil {
				f = (callbacks[1](this.r.result...))
			}
			result = f
		} else {
			this.pipeDoneTask = callbacks[0]
			if len(callbacks) > 1 {
				this.pipeFailTask = callbacks[1]
			}
			this.pipeFuture = NewFuture()
			result = this.pipeFuture.Promise()
		}
		ok = true
	})
	return
}

//完成一个任务
func (this *Future) end(r *futureResult) (e error) { //r *futureResult) {
	defer func() {
		e = getError(recover())
	}()
	if this.isPromised {
		e = errors.New("The future is promised")
	}
	e = errors.New("Cannoy resolve/reject more than once")
	this.onceEnd.Do(func() {
		//r := <-this.chIn
		this.setResult(r)

		//让Get函数可以返回
		this.chOut <- r
		close(this.chOut)

		//任务完成后调用回调函数
		execCallback(r, this.dones, this.fails, this.always)

		this.startPipe()
		e = nil
	})
	return
}

//返回与链式调用相关的对象
func (this *PromiseValue) getPipe(isResolved bool) (func(v ...interface{}) *PromiseValue, *Future) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if isResolved {
		return this.pipeDoneTask, this.pipeFuture
	} else {
		return this.pipeFailTask, this.pipeFuture
	}
}

//set this.r
func (this *Future) setResult(r *futureResult) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.r = r
}

func (this *PromiseValue) startPipe() {
	//处理链式异步任务
	pipeTask, pipeFuture := this.getPipe(this.r.ok)
	var f *PromiseValue
	if pipeTask != nil {
		f = pipeTask(this.r.result...)
	} else {
		f = this
	}

	f.Done(func(v ...interface{}) {
		pipeFuture.Reslove(v...)
	}).Fail(func(v ...interface{}) {
		pipeFuture.Reject(v...)
	})

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

//处理单个回调函数的添加请求
func (this *PromiseValue) handleOneCallback(callback func(v ...interface{}), t callbackType) {
	if callback == nil {
		return
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
	if f := this.addCallback(pendingAction, finalAction); f != nil {
		f()
	}
}

//添加回调函数的框架函数
func (this *PromiseValue) addCallback(pendingAction func(), finalAction func(*futureResult)) func() {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.r == nil {
		pendingAction()
		return nil
	} else {
		r := this.r
		return func() {
			finalAction(r)
		}
	}
}

//异步执行一个函数。如果最后一个返回值为bool，则将认为此值代表异步任务成功或失败。如果函数抛出error，则认为异步任务失败
func Start(action func() []interface{}) *Future {
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
		} else {
			fu.Reslove(r...)
		}
	}()

	return fu //.Promise()
}

func Start0(action func()) *Future {
	return Start(func() []interface{} {
		action()
		return make([]interface{}, 0, 0)
	})
}

func Wrap(value interface{}) *Future {
	fu := NewFuture()
	fu.Reslove(value)
	return fu
}

//Factory function for future
func NewFuture() *Future {
	f := &Future{new(sync.Once), &PromiseValue{new(sync.Once), new(sync.Mutex),
		make(chan *futureResult, 1),
		make([]func(v ...interface{}), 0, 8),
		make([]func(v ...interface{}), 0, 8),
		make([]func(v ...interface{}), 0, 4),
		pipe{}, nil}, false,
	}
	return f
}

//产生一个新的Future，如果列表中任意1个Future完成，则Future完成, 否则将触发Reject，参数为包含所有Future的Reject返回值的slice
func Any(fs ...*Future) *Future {
	nf := NewFuture()
	errors := make([]interface{}, len(fs), len(fs))
	chFails := make(chan anyFutureResult)

	for i, f := range fs {
		k := i
		f.Done(func(v ...interface{}) {
			nf.Reslove(v...)
		}).Fail(func(v ...interface{}) {
			chFails <- anyFutureResult{v, k}
		})
	}

	if len(fs) == 0 {
		nf.Reslove()
	} else {
		go func() {
			j := 0
			for {
				select {
				case r := <-chFails:
					errors[r.i] = r.result
					if j++; j == len(fs) {
						nf.Reject(errors...)
						break
					}
				case _ = <-nf.chOut:
					break
				}
			}
			//fmt.Println("exit any loop")
		}()
	}
	return nf
}

//产生一个新的Future，如果列表中所有Future都成功完成，则Future成功完成，否则失败
func When(fs ...*Future) *Future {
	f := NewFuture()
	if len(fs) == 0 {
		f.Reslove()
	} else {
		go func() {
			rs := make([]interface{}, len(fs))
			allOk := true
			for i, f := range fs {
				r, ok := f.Get()
				r = append(r, ok)
				rs[i] = r
				if !ok {
					allOk = false
				}
			}
			if allOk {
				f.Reslove(rs...)
			} else {
				f.Reject(rs...)
			}
		}()
	}

	return f
}

func forSlice(s []func(v ...interface{}), f func(func(v ...interface{}))) {
	for _, e := range s {
		f(e)
	}
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
