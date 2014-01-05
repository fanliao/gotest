package main

import (
	"reflect"
	"testing"
	"unsafe"
)

type RWTestStruct struct {
	Id   int
	Name string
}

func BenchmarkFastGeter(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := getFastRWer(o, p)
	var id int
	var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id = *((*int)(rw.Get(p, "Id")))
		name = *((*string)(rw.Get(p, "Name")))
	}
	b.StopTimer()
	b.Log(id, name)
}

func BenchmarkFastGeterByIndex(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := getFastRWer(o, p)
	var id int
	var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id = *((*int)(rw.GetbyIndex(p, 0)))
		name = *((*string)(rw.GetbyIndex(p, 1)))
	}
	b.StopTimer()
	b.Log(id, name)
}

func BenchmarkFastGeterByIndexButNoFunc(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := getFastRWer(o, p)
	meta := rw.GetStructMeta()
	var id int
	var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id = *((*int)(unsafe.Pointer(uintptr(p) + meta.FieldOffsetsByIndex[0])))
		name = *((*string)(unsafe.Pointer(uintptr(p) + meta.FieldOffsetsByIndex[1])))
	}
	b.StopTimer()
	b.Log(id, name)
}

func BenchmarkUnsafe(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	offsetId := unsafe.Offsetof(o.Id)
	OffsetName := unsafe.Offsetof(o.Name)
	var id int
	var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id = *((*int)(unsafe.Pointer(uintptr(p) + offsetId)))
		name = *((*string)(unsafe.Pointer(uintptr(p) + OffsetName)))
	}
	b.StopTimer()
	b.Log(id, name)
}

func BenchmarkUnsafeSet(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	offsetId := unsafe.Offsetof(o.Id)
	OffsetName := unsafe.Offsetof(o.Name)
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//id = *((*int)(unsafe.Pointer(uintptr(p) + offsetId)))
		//idPtr := ((*int)(unsafe.Pointer(uintptr(p) + offsetId)))
		//*idPtr = 11
		*((*int)(unsafe.Pointer(uintptr(p) + offsetId))) = 11
		//name = *((*string)(unsafe.Pointer(uintptr(p) + OffsetName)))
		*((*string)(unsafe.Pointer(uintptr(p) + OffsetName))) = "test unsafe set"

	}
	b.StopTimer()
	//id = o.Id
	//name = o.Name
	b.Log(o)
}

func BenchmarkReflectSet(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	v := reflect.ValueOf(o).Elem()
	//t := v.Type()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Field(0).SetInt(11)
		v.Field(1).SetString("test reflect set")
	}
	b.StopTimer()
	b.Log(o.Id, o.Name)
}

func BenchmarkGeter(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	var id int
	var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id = o.Id
		name = o.Name
	}
	b.StopTimer()
	b.Log(id, name)
}

func BenchmarkReflectGeter(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	v := reflect.ValueOf(o).Elem()
	//t := v.Type()
	var id int
	var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id = v.FieldByName("Id").Interface().(int)
		name = v.FieldByName("Name").Interface().(string)
	}
	b.StopTimer()
	b.Log(id, name)
}

func BenchmarkReflectGeterByIndex(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	v := reflect.ValueOf(o).Elem()
	//t := v.Type()
	var id int
	var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id = v.Field(0).Interface().(int)
		name = v.Field(1).Interface().(string)
	}
	b.StopTimer()
	b.Log(id, name)
}

func BenchmarFastSeter(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := getFastRWer(o, p)
	for i := 0; i < b.N; i++ {
		rw.Set(p, "Id", 2)
		rw.Set(p, "Name", "testSet")
	}
}
