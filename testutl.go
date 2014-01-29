package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

const ptrSize1 = unsafe.Sizeof((*byte)(nil))

func compare(a interface{}, b interface{}) bool {
	v1, v2 := reflect.ValueOf(a), reflect.ValueOf(b)

	if a == nil && b == nil {
		return true
	} else if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}

	if v1.Type() != v2.Type() {
		return false
	}

	if k := v1.Type().Kind(); k == reflect.Ptr || k == reflect.Map || k == reflect.Slice || k == reflect.Func || k == reflect.Struct {
		switch k {
		case reflect.Map:
			if len(v1.MapKeys()) == 0 && len(v2.MapKeys()) == 0 {
				return true
			}
			for _, k := range v1.MapKeys() {
				if v1.MapIndex(k) != v2.MapIndex(k) {
					return false
				}
			}
			for _, k := range v2.MapKeys() {
				if v2.MapIndex(k) != v1.MapIndex(k) {
					return false
				}
			}
			return true
		case reflect.Slice:
			if v1.Len() != v2.Len() {
				return false
			} else {
				for i := 0; i < v1.Len(); i++ {
					if v1.Index(i).Interface() != v2.Index(i).Interface() {
						fmt.Println("compare s", v1.Index(i).Interface(), v2.Index(i).Interface())
						return false
					}
				}
				return true
			}
		case reflect.Func:
			addr1, addr2 := InterfaceToPtr1(a), InterfaceToPtr1(b)
			return addr1 == addr2

		case reflect.Struct:
			addr1, addr2 := InterfaceToPtr1(a), InterfaceToPtr1(b)
			//Each interface{} variable takes up 2 words in memory:
			//one word for the type of what is contained,
			//the other word for either the contained data or a pointer to it.
			//so if data size is more than one word, addr1 be a pointer
			//otherwise addr1 be the data
			if v1.Type().Size() > ptrSize1 {
				return compareBytes(addr1, addr2, v1.Type().Size())
			} else {
				return addr1 == addr2
			}
			//for i := 0; i < v1.NumField(); i++ {
			//	if !compare(v1.Field(i).Interface(), v2.Field(i).Interface()) {
			//		return false
			//	}

			//}
		case reflect.Ptr:
			if v1.Elem().Type().Kind() == reflect.Struct {
				//fmt.Println("struct", reflect.Indirect(v1).Interface(), reflect.Indirect(v2).Interface())
				return compareBytes(reflect.Indirect(v1).UnsafeAddr(), reflect.Indirect(v2).UnsafeAddr(), v1.Elem().Type().Size())
			} else {
				return a == b
			}
		}
		return false
	} else if a == b {
		return true
	} else {
		return false
	}
}

func compareBytes(addr1 uintptr, addr2 uintptr, size uintptr) bool {
	for i := 0; uintptr(i) < size; i++ {
		//fmt.Println(addr1+uintptr(i), addr2+uintptr(i))
		//fmt.Println(*((*byte)(unsafe.Pointer(addr1 + uintptr(i)))), *((*byte)(unsafe.Pointer(addr2 + uintptr(i)))))
		if *((*byte)(unsafe.Pointer(addr1 + uintptr(i)))) != *((*byte)(unsafe.Pointer(addr2 + uintptr(i)))) {
			return false
		}
	}
	return true
}

// interfaceHeader is the header for an interface{} value. it is copied from unsafe.emptyInterface
type interfaceHeader1 struct {
	typ  uintptr
	word uintptr
}

func InterfaceToPtr1(i interface{}) uintptr {
	s := *((*interfaceHeader1)(unsafe.Pointer(&i)))
	return s.word
}
