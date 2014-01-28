package main

import (
	"reflect"
)

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

	if k := v1.Type().Kind(); k == reflect.Map || k == reflect.Slice || k == reflect.Func {
		return false
	} else if a == b {
		return true
	} else {
		return false
	}
}
