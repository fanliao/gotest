package main

import (
	"testing"
	"time"
)

func TestDoneAlways(t *testing.T) {

	order := make([]int, 0, 10)
	f := NewFuture().Done(func(v ...interface{}) {
		order = append(order, 2)
		//t.Log(v...)
		AreEqual(v, []interface{}{10, "ok"}, t)
	}).Always(func(v ...interface{}) {
		order = append(order, 4)
		//t.Log(v...)
		AreEqual(v, []interface{}{10, "ok"}, t)
	}).Done(func(v ...interface{}) {
		order = append(order, 3)
		//t.Log(v...)
		AreEqual(v, []interface{}{10, "ok"}, t)
	})

	go func() {
		time.Sleep(500 * time.Millisecond)
		f.Reslove(10, "ok")
		t.Log("Resloved")
		order = append(order, 1)
	}()

	_, _ = f.Get()
	AreEqual(order, []int{1, 2, 3, 4}, t)
	t.Log("End")
}

func TestFailAlways(t *testing.T) {
	order := make([]int, 0, 10)
	f := NewFuture().Fail(func(v ...interface{}) {
		order = append(order, 2)
		//t.Log(v...)
		AreEqual(v, []interface{}{10, "fail"}, t)
	}).Always(func(v ...interface{}) {
		order = append(order, 4)
		//t.Log(v...)
		AreEqual(v, []interface{}{10, "fail"}, t)
	}).Fail(func(v ...interface{}) {
		order = append(order, 3)
		//t.Log(v...)
		AreEqual(v, []interface{}{10, "fail"}, t)
	})

	go func() {
		time.Sleep(500 * time.Millisecond)
		f.Reject(10, "fail")
		t.Log("Rejected")
		order = append(order, 1)
	}()

	_, _ = f.Get()
	AreEqual(order, []int{1, 2, 3, 4}, t)
	t.Log("End")
}

func TestGet(t *testing.T) {
	order := make([]int, 0, 10)
	getfutrue := func() *Future {
		f := NewFuture().Done(func(v ...interface{}) {
			time.Sleep(200 * time.Millisecond)
			order = append(order, 3)
			//t.Log(v...)
			AreEqual(v, []interface{}{10, "ok"}, t)
		}).Fail(func(v ...interface{}) {
			time.Sleep(200 * time.Millisecond)
			order = append(order, 3)
			//t.Log(v...)
			AreEqual(v, []interface{}{10, "fail"}, t)
		})
		return f
	}

	f := getfutrue()
	go func() {
		time.Sleep(500 * time.Millisecond)
		f.Reslove(10, "ok")
		order = append(order, 1)
	}()

	order = append(order, 0)
	r, ok := f.Get()
	order = append(order, 2)
	//Done callback will be run in another gorouter, so order is 0,1,2
	AreEqual(order, []int{0, 1, 2}, t)
	AreEqual(r, []interface{}{10, "ok"}, t)
	AreEqual(ok, true, t)
	time.Sleep(500 * time.Millisecond)
	//Done callback has run be done, so order is 0,1,2,3
	AreEqual(order, []int{0, 1, 2, 3}, t)

	r, ok = f.Get()
	//test get again
	AreEqual(r, []interface{}{10, "ok"}, t)
	AreEqual(ok, true, t)

	order = make([]int, 0, 10)
	f = getfutrue()

	go func() {
		time.Sleep(500 * time.Millisecond)
		f.Reject(10, "fail")
		t.Log("Resloved")
		order = append(order, 1)
	}()

	order = append(order, 0)
	r, ok = f.Get()
	order = append(order, 2)
	//Done callback will be run in another gorouter, so order is 0,1,2
	AreEqual(order, []int{0, 1, 2}, t)
	AreEqual(r, []interface{}{10, "fail"}, t)
	AreEqual(ok, false, t)
	time.Sleep(500 * time.Millisecond)
	//Done callback has run be done, so order is 0,1,2,3
	AreEqual(order, []int{0, 1, 2, 3}, t)

	r, ok = f.Get()
	//test get again
	AreEqual(r, []interface{}{10, "fail"}, t)
	AreEqual(ok, false, t)

}

func TestThen(t *testing.T) {
	order := make([]int, 0, 10)
	f := NewFuture().Done(func(v ...interface{}) {
		time.Sleep(200 * time.Millisecond)
		order = append(order, 3)
		AreEqual(v, []interface{}{10, "ok"}, t)
	}).Then(func(v ...interface{}) *Future {
		AreEqual(v, []interface{}{10, "ok"}, t)
		f1 := NewFuture().Done(func(v ...interface{}) {
			time.Sleep(200 * time.Millisecond)
			order = append(order, 3)
			AreEqual(v, []interface{}{10, "ok"}, t)
		})
		go func() {
			time.Sleep(500 * time.Millisecond)
			f1.Reslove(v[0].(int)*2, v[1].(string)+"2")
			order = append(order, 1)
		}()
		return f1
	})

	go func() {
		time.Sleep(500 * time.Millisecond)
		f.Reslove(10, "ok")
		t.Log("Resloved")
		order = append(order, 1)
	}()

	order = append(order, 0)
	_, _ = f.Get()

}

func TestException(t *testing.T) {

}

func TestAny(t *testing.T) {

}

func TestWhen(t *testing.T) {

}
