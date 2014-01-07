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
	FieldSizeByIndex    []uintptr
}

type FastRWer interface {
	GetStructMeta() StructMeta
	GetbyIndex(obj unsafe.Pointer, i int) unsafe.Pointer
	Get(obj unsafe.Pointer, fieldName string) unsafe.Pointer
	Set(obj unsafe.Pointer, filedName string, value interface{})
}

//FastRWer implement class
type defaultFastRWImpl struct {
	StructMeta
	GetbyIndexImpl func(this defaultFastRWImpl, obj unsafe.Pointer, i int) unsafe.Pointer
	GetImpl        func(this defaultFastRWImpl, obj unsafe.Pointer, fieldName string) unsafe.Pointer
	SetImpl        func(this defaultFastRWImpl, obj unsafe.Pointer, fieldName string, value interface{})
}

func (this defaultFastRWImpl) GetbyIndex(obj unsafe.Pointer, i int) unsafe.Pointer {
	//return this.GetbyIndexImpl(this, obj, i)
	return unsafe.Pointer(uintptr(obj) + this.FieldOffsetsByIndex[i])
}

func (this defaultFastRWImpl) Get(obj unsafe.Pointer, fieldName string) unsafe.Pointer {
	return this.GetImpl(this, obj, fieldName)
}

func (this defaultFastRWImpl) Set(obj unsafe.Pointer, fieldName string, value interface{}) {
	this.SetImpl(this, obj, fieldName, value)
}

func (this defaultFastRWImpl) GetStructMeta() StructMeta {
	return this.StructMeta
}

//factory function
func newFastRWImpl(structMeta StructMeta,
	getByIndex func(this defaultFastRWImpl, obj unsafe.Pointer, i int) unsafe.Pointer,
	get func(this defaultFastRWImpl, obj unsafe.Pointer, fieldName string) unsafe.Pointer,
	set func(this defaultFastRWImpl, obj unsafe.Pointer, fieldName string, value interface{})) *defaultFastRWImpl {
	return &defaultFastRWImpl{structMeta, getByIndex, get, set}
}

//Get a FastRWer implement class by a pointer of struct
//obj must be a pointer of struct value
func getFastRWer(obj interface{}, p unsafe.Pointer) *defaultFastRWImpl {
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
	meta.FieldSizeByIndex = make([]uintptr, numField, numField)
	for i := 0; i < t.NumField(); i++ {
		fType := t.Field(i)
		f := v.Field(i)
		meta.FieldOffsetsByIndex[i] = f.UnsafeAddr() - objAddr
		meta.FieldOffsetsByName[fType.Name] = meta.FieldOffsetsByIndex[i]
		meta.FieldNamesByIndex[i] = fType.Name
		meta.FieldSizeByIndex[i] = f.Type().Size()
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

func fastGetByOffset(obj unsafe.Pointer, offset uintptr) unsafe.Pointer {
	//return this.GetbyIndexImpl(this, obj, i)
	return unsafe.Pointer(uintptr(obj) + offset)
}

func FastGet(obj unsafe.Pointer, this *defaultFastRWImpl, i int) unsafe.Pointer {
	//return this.GetbyIndexImpl(this, obj, i)
	return unsafe.Pointer(uintptr(obj) + this.FieldOffsetsByIndex[i])
}

func FastSet(obj unsafe.Pointer, this *defaultFastRWImpl, i int, source uintptr) {
	size := this.FieldSizeByIndex[i]
	target := uintptr(obj) + this.FieldOffsetsByIndex[i]
	copyVar(target, source, size)
}

func copyVar(target uintptr, source uintptr, size uintptr) {
	switch size {
	case 1:
		*((*[1]byte)(unsafe.Pointer(target))) = *((*[1]byte)(unsafe.Pointer(source)))
	case 2:
		*((*[2]byte)(unsafe.Pointer(target))) = *((*[2]byte)(unsafe.Pointer(source)))
	case 4:
		*((*[4]byte)(unsafe.Pointer(target))) = *((*[4]byte)(unsafe.Pointer(source)))
	case 8:
		*((*[8]byte)(unsafe.Pointer(target))) = *((*[8]byte)(unsafe.Pointer(source)))
	case 12:
		*((*[12]byte)(unsafe.Pointer(target))) = *((*[12]byte)(unsafe.Pointer(source)))
	case 16:
		*((*[16]byte)(unsafe.Pointer(target))) = *((*[16]byte)(unsafe.Pointer(source)))
	default:
		unWriteSize := size
		targetAddr := target
		sourceAddr := source
		for {
			if unWriteSize <= 16 {
				copyVar(targetAddr, sourceAddr, unWriteSize)
			}
			*((*[16]byte)(unsafe.Pointer(targetAddr))) = *((*[16]byte)(unsafe.Pointer(sourceAddr)))
			targetAddr += 16
			sourceAddr += 16
			unWriteSize -= 16
		}
	}

}
