package main

import (
	//"fmt"
	"reflect"
	"unsafe"
)

type StructMeta struct {
	IsFixedSize         bool
	FieldOffsetsByName  map[string]uintptr
	FieldOffsetsByIndex []uintptr
	FieldNamesByIndex   []string
}

type FastRWer interface {
	GetStructMeta() StructMeta
	GetbyIndex(obj unsafe.Pointer, i int) unsafe.Pointer
	Get(obj unsafe.Pointer, fieldName string) unsafe.Pointer
	Set(obj unsafe.Pointer, filedName string, value interface{})
}

//FastRWer implement class
type defaultFastRWImpl struct {
	structMeta     StructMeta
	GetbyIndexImpl func(this defaultFastRWImpl, obj unsafe.Pointer, i int) unsafe.Pointer
	GetImpl        func(this defaultFastRWImpl, obj unsafe.Pointer, fieldName string) unsafe.Pointer
	SetImpl        func(this defaultFastRWImpl, obj unsafe.Pointer, fieldName string, value interface{})
}

func (this defaultFastRWImpl) GetbyIndex(obj unsafe.Pointer, i int) unsafe.Pointer {
	return this.GetbyIndexImpl(this, obj, i)
}

func (this defaultFastRWImpl) Get(obj unsafe.Pointer, fieldName string) unsafe.Pointer {
	return this.GetImpl(this, obj, fieldName)
}

func (this defaultFastRWImpl) Set(obj unsafe.Pointer, fieldName string, value interface{}) {
	this.SetImpl(this, obj, fieldName, value)
}

func (this defaultFastRWImpl) GetStructMeta() StructMeta {
	return this.structMeta
}

//factory function
func newFastRWImpl(structMeta StructMeta,
	getByIndex func(this defaultFastRWImpl, obj unsafe.Pointer, i int) unsafe.Pointer,
	get func(this defaultFastRWImpl, obj unsafe.Pointer, fieldName string) unsafe.Pointer,
	set func(this defaultFastRWImpl, obj unsafe.Pointer, fieldName string, value interface{})) defaultFastRWImpl {
	return defaultFastRWImpl{structMeta, getByIndex, get, set}
}

//Get a FastRWer implement class by a pointer of struct
//obj must be a pointer of struct value
func getFastRWer(obj interface{}, p unsafe.Pointer) FastRWer {
	v := reflect.Indirect(reflect.ValueOf(obj))
	//fmt.Println(v)
	t := v.Type()
	//pt := reflect.ValueOf(obj)
	objAddr := uintptr(p)
	numField := t.NumField()

	meta := StructMeta{}
	meta.IsFixedSize = true
	meta.FieldOffsetsByName = make(map[string]uintptr)
	meta.FieldOffsetsByIndex = make([]uintptr, numField, numField)
	meta.FieldNamesByIndex = make([]string, numField, numField)
	for i := 0; i < t.NumField(); i++ {
		fType := t.Field(i)
		f := v.Field(i)
		meta.FieldOffsetsByIndex[i] = f.UnsafeAddr() - objAddr
		meta.FieldOffsetsByName[fType.Name] = meta.FieldOffsetsByIndex[i]
		meta.FieldNamesByIndex[i] = fType.Name
		//v := f.Interface()
	}

	return newFastRWImpl(meta,
		func(this defaultFastRWImpl, obj unsafe.Pointer, i int) unsafe.Pointer {
			return unsafe.Pointer(uintptr(obj) + meta.FieldOffsetsByIndex[i])
		},
		func(this defaultFastRWImpl, obj unsafe.Pointer, fieldName string) unsafe.Pointer {
			return unsafe.Pointer(uintptr(obj) + meta.FieldOffsetsByName[fieldName])
		},
		func(this defaultFastRWImpl, obj unsafe.Pointer, fieldName string, value interface{}) {
			//p := unsafe.Pointer(uintptr(unsafe.Pointer((*interface{})(obj))) + meta.FieldOffsetsByName[fieldName])
			//&((*interface{})(p)) = value
		})
}
