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

}

func TestException(t *testing.T) {

}

func TestAny(t *testing.T) {

}

func TestWhen(t *testing.T) {

}
