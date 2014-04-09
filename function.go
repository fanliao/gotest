package main

import (
//"fmt"
)

type intsFunc func() []int

func (this intsFunc) result() []int {
	return this()
}

func (this intsFunc) itr() func() (int, bool) {
	is := this()
	i := 0
	return func() (v int, ok bool) {
		if i < len(is) {
			v, ok = is[i], true
			i = i + 1
			return
		} else {
			return 0, false
		}
	}

}

func (this intsFunc) add(b int) intsFunc {
	f := func() []int {
		is := this()
		for i, v := range is {
			is[i] = v + b
		}
		return is
	}
	return f
}

//func main() {
//	t := intsFunc (func() []int {
//		return []int{1, 2, 3}
//	})
//	t1 := intsFunc (func() []int {
//		return []int{1, 2, 3}
//	})
//	fmt.Println((t.add(1).add(2))())
//	next1 := t.itr()
//	next2 := t.add(1).add(2).itr()
//	for {
//		if v, ok := next1(); ok {
//			fmt.Println(v)
//		} else {
//			break
//		}
//	}
//	for {
//		if v, ok := next2(); ok {
//			fmt.Println(v)
//		} else {
//			break
//		}
//	}
//	fmt.Println(next1)
//	fmt.Println(next2)
//	fmt.Println(t1.add(5).add(2).itr())

//	f1 := func() {}
//	f2 := func() {}
//	fmt.Println(f1)
//	fmt.Println(f2)
//	fmt.Println("Hello World")
//}
