package main

import (
	"github.com/ahmetalpbalkan/go-linq"
	"math"
	"runtime"
	"strconv"
	"testing"
)

const (
	count    int = 10000
	MAXPROCS int = 4
)

var (
	arr     []interface{} = make([]interface{}, count, count)
	arrUser []interface{} = make([]interface{}, count, count)
)

func init() {
	runtime.GOMAXPROCS(MAXPROCS)
	for i := 0; i < count; i++ {
		arr[i] = i
		arrUser[i] = user{i, "user" + strconv.Itoa(i)}
	}
}

type user struct {
	id   int
	name string
}

func where1(v interface{}) bool {
	i := v.(int)
	//time.Sleep(10 * time.Nanosecond)
	return i%2 == 0
}

func whereUser(v interface{}) bool {
	u := v.(user)
	//time.Sleep(10 * time.Nanosecond)
	return u.id%2 == 0
}

func selectUser(v interface{}) interface{} {
	u := v.(user)
	return strconv.Itoa(u.id) + "/" + u.name
}

func select1(v interface{}) interface{} {
	return v.(int) + 9999
}

func select2(v interface{}) interface{} {
	return math.Sin(math.Cos(math.Pow(float64(v.(int)), 2)))
}

func BenchmarkSyncWhere(b *testing.B) {
	for i := 0; i < b.N; i++ {
		q := queryableS{arrUser, make([]func([]interface{}) []interface{}, 0, 1)}
		dst := q.Where(whereUser).Get()

		if len(dst) != count/2 {
			b.Fail()
			b.Error("size is ", len(dst))
		}
	}
}

func BenchmarkBlockSourceWhere(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dst := From(arrUser).Where(whereUser).Results()
		if len(dst) != count/2 {
			b.Fail()
			b.Log("arr=", arr)
			b.Error("size is ", len(dst))
			b.Log("dst=", dst)
		}
	}
}

func BenchmarkGoLinqWhere(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dst, _ := linq.From(arrUser).Where(func(i linq.T) (bool, error) {
			v := i.(user)
			return v.id%2 == 0, nil
		}).Results()
		if len(dst) != count/2 {
			b.Fail()
			b.Error("size is ", len(dst))
		}
	}
}

func BenchmarkGoLinqParallelWhere(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dst, _ := linq.From(arrUser).AsParallel().Where(func(i linq.T) (bool, error) {
			v := i.(user)
			return v.id%2 == 0, nil
		}).Results()
		if len(dst) != count/2 {
			b.Fail()
			b.Error("size is ", len(dst))
		}
	}
}

func BenchmarkBlockSourceSelectWhere(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dst := From(arrUser).Where(whereUser).Select(selectUser).Results()
		if len(dst) != count/2 {
			b.Fail()
			b.Log("arr=", arr)
			b.Error("size is ", len(dst))
			b.Log("dst=", dst)
		}
	}
}

func BenchmarkGoLinqSelectWhere(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dst, _ := linq.From(arrUser).Where(func(i linq.T) (bool, error) {
			v := i.(user)
			return v.id%2 == 0, nil
		}).Select(func(v linq.T) (linq.T, error) {
			u := v.(user)
			return strconv.Itoa(u.id) + "/" + u.name, nil
		}).Results()
		if len(dst) != count/2 {
			b.Fail()
			b.Error("size is ", len(dst))
		}
	}
}

func BenchmarkGoLinqParallelSelectWhere(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dst, _ := linq.From(arrUser).AsParallel().Where(func(i linq.T) (bool, error) {
			v := i.(user)
			return v.id%2 == 0, nil
		}).Select(func(v linq.T) (linq.T, error) {
			u := v.(user)
			return strconv.Itoa(u.id) + "/" + u.name, nil
		}).Results()
		if len(dst) != count/2 {
			b.Fail()
			b.Error("size is ", len(dst))
		}
	}
}
