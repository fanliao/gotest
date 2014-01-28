package main

import (
	"fmt"
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

	if k := v1.Type().Kind(); k == reflect.Map || k == reflect.Slice || k == reflect.Func || k == reflect.Struct {
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
			return false

		case reflect.Struct:
			for i := 0; i < v1.NumField(); i++ {
				if !compare(v1.Field(i).Interface(), v2.Field(i).Interface()) {
					return false
				}

			}
		}
		return true
	} else if a == b {
		return true
	} else {
		return false
	}
}
