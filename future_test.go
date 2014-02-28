package main

import (
	"testing"
	"time"
)

func TestDoneAlways(t *testing.T) {

	order := make([]int, 0, 10)
	task := func() []interface{} {
		time.Sleep(500 * time.Millisecond)
		order = append(order, 1)
		return []interface{}{10, "ok", true}
	}
	f := Submit(task).Done(func(v ...interface{}) {
		time.Sleep(200 * time.Millisecond)
		order = append(order, 3)
		AreEqual(v, []interface{}{10, "ok"}, t)
	}).Always(func(v ...interface{}) {
		order = append(order, 5)
		AreEqual(v, []interface{}{10, "ok"}, t)
	}).Done(func(v ...interface{}) {
		order = append(order, 4)
		AreEqual(v, []interface{}{10, "ok"}, t)
	})

	order = append(order, 0)
	r, ok := f.Get()
	order = append(order, 2)
	time.Sleep(500 * time.Millisecond)
	order = append(order, 6)
	//The code after Get() and the callback will be concurrent run
	//The always callback always run after all done or fail callbacks be done
	AreEqual(order, []int{0, 1, 2, 3, 4, 5, 6}, t)
	AreEqual(r, []interface{}{10, "ok"}, t)
	AreEqual(ok, true, t)
}

func TestFailAlways(t *testing.T) {
	order := make([]int, 0, 10)
	task := func() []interface{} {
		time.Sleep(500 * time.Millisecond)
		order = append(order, 1)
		return []interface{}{10, "fail", false}
	}
	f := Submit(task).Fail(func(v ...interface{}) {
		time.Sleep(200 * time.Millisecond)
		order = append(order, 3)
		AreEqual(v, []interface{}{10, "fail"}, t)
	}).Always(func(v ...interface{}) {
		order = append(order, 5)
		AreEqual(v, []interface{}{10, "fail"}, t)
	}).Fail(func(v ...interface{}) {
		order = append(order, 4)
		AreEqual(v, []interface{}{10, "fail"}, t)
	})

	order = append(order, 0)
	r, ok := f.Get()
	order = append(order, 2)
	time.Sleep(500 * time.Millisecond)
	order = append(order, 6)
	AreEqual(order, []int{0, 1, 2, 3, 4, 5, 6}, t)
	AreEqual(r, []interface{}{10, "fail"}, t)
	AreEqual(ok, false, t)
}

func TestThen(t *testing.T) {
	order := make([]int, 0, 10)
	taskDone := func() []interface{} {
		time.Sleep(500 * time.Millisecond)
		order = append(order, 1)
		return []interface{}{10, "ok", true}
	}
	taskFail := func() []interface{} {
		time.Sleep(500 * time.Millisecond)
		order = append(order, 1)
		return []interface{}{10, "fail", false}
	}

	SubmitWithCallback := func(task func() []interface{}) *Future {
		f := Submit(task).Done(func(v ...interface{}) {
			time.Sleep(200 * time.Millisecond)
			order = append(order, 2)
			AreEqual(v, []interface{}{10, "ok"}, t)
		}).Fail(func(v ...interface{}) {
			time.Sleep(200 * time.Millisecond)
			order = append(order, 2)
			AreEqual(v, []interface{}{10, "fail"}, t)
		}).Then(func(v ...interface{}) *Future {
			AreEqual(v, []interface{}{10, "ok"}, t)
			f1 := Submit(func() []interface{} {
				time.Sleep(500 * time.Millisecond)
				order = append(order, 3)
				return []interface{}{v[0].(int) * 2, v[1].(string) + "2", true}
			})
			return f1
		}, func(v ...interface{}) *Future {
			AreEqual(v, []interface{}{10, "fail"}, t)
			f1 := Submit(func() []interface{} {
				time.Sleep(500 * time.Millisecond)
				order = append(order, 3)
				return []interface{}{v[0].(int) * 2, v[1].(string) + "2", false}
			})
			return f1
		}).Done(func(v ...interface{}) {
			time.Sleep(100 * time.Millisecond)
			order = append(order, 5)
			AreEqual(v, []interface{}{20, "ok2"}, t)
		}).Fail(func(v ...interface{}) {
			time.Sleep(100 * time.Millisecond)
			order = append(order, 5)
			AreEqual(v, []interface{}{20, "fail2"}, t)
		})
		return f
	}

	//for then api, the new Future object will be return
	//New future task object should be started after current future be done or failed
	f := SubmitWithCallback(taskDone)
	order = append(order, 0)
	r, ok := f.Get()
	order = append(order, 4)
	time.Sleep(300 * time.Millisecond)
	order = append(order, 6)
	AreEqual(order, []int{0, 1, 2, 3, 4, 5, 6}, t)
	AreEqual(r, []interface{}{20, "ok2"}, t)
	AreEqual(ok, true, t)

	order = make([]int, 0, 10)
	f = SubmitWithCallback(taskFail)
	order = append(order, 0)
	r, ok = f.Get()
	order = append(order, 4)
	time.Sleep(300 * time.Millisecond)
	order = append(order, 6)
	AreEqual(order, []int{0, 1, 2, 3, 4, 5, 6}, t)
	AreEqual(r, []interface{}{20, "fail2"}, t)
	AreEqual(ok, false, t)

}

func TestException(t *testing.T) {

}

func TestAny(t *testing.T) {

}

func TestWhen(t *testing.T) {

}
