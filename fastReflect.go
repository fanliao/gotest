package main

import (
	//"fmt"
	"reflect"
	"unsafe"
)

type StructMeta struct {
	//IsFixedSize         bool
	FieldIndexsByName   map[string]int
	FieldOffsetsByIndex []uintptr
	FieldNamesByIndex   []string
	FieldSizeByIndex    []uintptr
	FieldTypesByIndex   []reflect.Type
}

//FastRWer implement class
type FastRW struct {
	StructMeta
}

func (this *FastRW) GetPtr(obj unsafe.Pointer, i int) unsafe.Pointer {
	return FastGet(obj, this, i)
	//return unsafe.Pointer(uintptr(obj) + this.FieldOffsetsByIndex[i])
}

func (this *FastRW) GetValue(obj unsafe.Pointer, i int) interface{} {
	ptr := FastGet(obj, this, i)
	return reflect.NewAt(this.FieldTypesByIndex[i], ptr).Elem().Interface()
	//return unsafe.Pointer(uintptr(obj) + this.FieldOffsetsByIndex[i])
}

func (this *FastRW) GetPtrByName(obj unsafe.Pointer, fieldName string) unsafe.Pointer {
	return this.GetPtr(obj, this.FieldIndexsByName[fieldName])
}

func (this *FastRW) Set(obj unsafe.Pointer, i int, source uintptr) {
	target := uintptr(obj) + this.FieldOffsetsByIndex[i]
	copyVar(target, source, this.FieldSizeByIndex[i])
}

func (this *FastRW) SetByName(obj unsafe.Pointer, fieldName string, source uintptr) {
	this.Set(obj, this.FieldIndexsByName[fieldName], source)
}

func FastGet(obj unsafe.Pointer, this *FastRW, i int) unsafe.Pointer {
	//return this.GetbyIndexImpl(this, obj, i)
	return unsafe.Pointer(uintptr(obj) + this.FieldOffsetsByIndex[i])
}

func FastSet(obj unsafe.Pointer, this *FastRW, i int, source uintptr) {
	target := uintptr(obj) + this.FieldOffsetsByIndex[i]
	copyVar(target, source, this.FieldSizeByIndex[i])
}

//factory function
func newFastRWImpl(structMeta StructMeta) *FastRW {
	return &FastRW{structMeta}
}

//Get a FastRWer implement class by a pointer of struct
//obj must be a pointer of struct value
func GetFastRWer(obj interface{}) *FastRW {
	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()
	objAddr := v.UnsafeAddr()
	numField := t.NumField()

	meta := StructMeta{}
	meta.FieldIndexsByName = make(map[string]int, numField)
	meta.FieldOffsetsByIndex = make([]uintptr, numField, numField)
	meta.FieldNamesByIndex = make([]string, numField, numField)
	meta.FieldSizeByIndex = make([]uintptr, numField, numField)
	meta.FieldTypesByIndex = make([]reflect.Type, numField, numField)
	for i := 0; i < t.NumField(); i++ {
		fType := t.Field(i)
		f := v.Field(i)
		meta.FieldOffsetsByIndex[i] = f.UnsafeAddr() - objAddr
		meta.FieldIndexsByName[fType.Name] = i
		meta.FieldNamesByIndex[i] = fType.Name
		meta.FieldSizeByIndex[i] = f.Type().Size()
		meta.FieldTypesByIndex[i] = f.Type()
		//v := f.Interface()
	}

	return newFastRWImpl(meta)
}

func fastGetByOffset(obj unsafe.Pointer, offset uintptr) unsafe.Pointer {
	//return this.GetbyIndexImpl(this, obj, i)
	return unsafe.Pointer(uintptr(obj) + offset)
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
