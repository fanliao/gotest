package main

import (
	//"fmt"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

type metaCache struct {
	lock  *sync.RWMutex
	metas map[reflect.Type]*FastRW
}

type structMeta struct {
	//IsFixedSize         bool
	FieldIndexsByName   map[string]int
	FieldOffsetsByIndex []uintptr
	FieldNamesByIndex   []string
	FieldSizeByIndex    []uintptr
	FieldTypesByIndex   []reflect.Type
}

//FastRWer implement class
type FastRW struct {
	structMeta
}

var dateType reflect.Type = reflect.TypeOf(time.Now())
var mc *metaCache

func init() {
	mc = &metaCache{}
	mc.lock = new(sync.RWMutex)
	mc.metas = make(map[reflect.Type]*FastRW)
}

func (this *metaCache) get(typ reflect.Type) *FastRW {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if val, ok := this.metas[typ]; ok {
		return val
	}
	return nil
}

func (this *metaCache) set(typ reflect.Type, rw *FastRW) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if val, ok := this.metas[typ]; !ok {
		this.metas[typ] = rw
	} else if val != rw {
		this.metas[typ] = rw
	}
}

func (this *FastRW) Ptr(obj unsafe.Pointer, i int) unsafe.Pointer {
	return FastGet(obj, this, i)
}

func (this *FastRW) Value(obj unsafe.Pointer, i int) interface{} {
	typ, ptr := this.FieldTypesByIndex[i], FastGet(obj, this, i)
	return getValue(typ, ptr)
}

func (this *FastRW) CopyPtr(obj unsafe.Pointer, i int, target unsafe.Pointer) {
	ptr := FastGet(obj, this, i)
	copyVar(uintptr(target), uintptr(ptr), this.FieldSizeByIndex[i])
}

func (this *FastRW) PtrByName(obj unsafe.Pointer, fieldName string) unsafe.Pointer {
	return this.Ptr(obj, this.FieldIndexsByName[fieldName])
}

func (this *FastRW) GetValueByName(obj unsafe.Pointer, fieldName string) interface{} {
	return this.Value(obj, this.FieldIndexsByName[fieldName])
}

func (this *FastRW) SetPtr(obj unsafe.Pointer, i int, source uintptr) {
	target := uintptr(obj) + this.FieldOffsetsByIndex[i]
	copyVar(target, source, this.FieldSizeByIndex[i])
}

func (this *FastRW) SetPtrByName(obj unsafe.Pointer, fieldName string, source uintptr) {
	i := this.FieldIndexsByName[fieldName]
	target := uintptr(obj) + this.FieldOffsetsByIndex[i]
	copyVar(target, source, this.FieldSizeByIndex[i])
}

func (this *FastRW) SetValue(obj unsafe.Pointer, i int, source interface{}) {
	target := uintptr(obj) + this.FieldOffsetsByIndex[i]
	addr := uintptr(unsafe.Pointer(&source))
	copyVar(target, addr, this.FieldSizeByIndex[i])
}

func (this *FastRW) SetValueByName(obj unsafe.Pointer, fieldName string, source interface{}) {
	i := this.FieldIndexsByName[fieldName]
	target := uintptr(obj) + this.FieldOffsetsByIndex[i]
	addr := uintptr(unsafe.Pointer(&source))
	copyVar(target, addr, this.FieldSizeByIndex[i])
}

func FastGet(obj unsafe.Pointer, this *FastRW, i int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(obj) + this.FieldOffsetsByIndex[i])
}

func FastSet(obj unsafe.Pointer, this *FastRW, i int, source uintptr) {
	target := uintptr(obj) + this.FieldOffsetsByIndex[i]
	copyVar(target, source, this.FieldSizeByIndex[i])
}

//Get a FastRWer implement class by a pointer of struct
//obj must be a pointer of struct value
func GetFastRWer(obj interface{}) *FastRW {
	v := reflect.Indirect(reflect.ValueOf(obj))
	t := v.Type()

	if fastRW := mc.get(t); fastRW != nil {
		return fastRW
	} else {

		meta := structMeta{}

		objAddr := v.UnsafeAddr()
		numField := t.NumField()

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

		fastRW = newFastRWImpl(meta)
		mc.set(t, fastRW)
		return fastRW
	}

}

//factory function
func newFastRWImpl(meta structMeta) *FastRW {
	return &FastRW{meta}
}

func copyVar(target uintptr, source uintptr, size uintptr) {
	//fmt.Println("target=", target, " source=", source)
	switch size {
	case 1:
		*((*[1]byte)(unsafe.Pointer(target))) = *((*[1]byte)(unsafe.Pointer(source)))
	case 2:
		*((*[2]byte)(unsafe.Pointer(target))) = *((*[2]byte)(unsafe.Pointer(source)))
	case 3:
		*((*[3]byte)(unsafe.Pointer(target))) = *((*[3]byte)(unsafe.Pointer(source)))
	case 4:
		*((*[4]byte)(unsafe.Pointer(target))) = *((*[4]byte)(unsafe.Pointer(source)))
	case 5:
		*((*[5]byte)(unsafe.Pointer(target))) = *((*[5]byte)(unsafe.Pointer(source)))
	case 6:
		*((*[6]byte)(unsafe.Pointer(target))) = *((*[6]byte)(unsafe.Pointer(source)))
	case 7:
		*((*[7]byte)(unsafe.Pointer(target))) = *((*[7]byte)(unsafe.Pointer(source)))
	case 8:
		*((*[8]byte)(unsafe.Pointer(target))) = *((*[8]byte)(unsafe.Pointer(source)))
	case 9:
		*((*[9]byte)(unsafe.Pointer(target))) = *((*[9]byte)(unsafe.Pointer(source)))
	case 10:
		*((*[10]byte)(unsafe.Pointer(target))) = *((*[10]byte)(unsafe.Pointer(source)))
	case 11:
		*((*[11]byte)(unsafe.Pointer(target))) = *((*[11]byte)(unsafe.Pointer(source)))
	case 12:
		*((*[12]byte)(unsafe.Pointer(target))) = *((*[12]byte)(unsafe.Pointer(source)))
	case 13:
		*((*[13]byte)(unsafe.Pointer(target))) = *((*[13]byte)(unsafe.Pointer(source)))
	case 14:
		*((*[14]byte)(unsafe.Pointer(target))) = *((*[14]byte)(unsafe.Pointer(source)))
	case 15:
		*((*[15]byte)(unsafe.Pointer(target))) = *((*[15]byte)(unsafe.Pointer(source)))
	case 16:
		*((*[16]byte)(unsafe.Pointer(target))) = *((*[16]byte)(unsafe.Pointer(source)))
	case 24:
		*((*[24]byte)(unsafe.Pointer(target))) = *((*[24]byte)(unsafe.Pointer(source)))
	default:
		unWriteSize := size
		targetAddr := target
		sourceAddr := source
		for {
			if unWriteSize <= 16 || unWriteSize == 24 {
				copyVar(targetAddr, sourceAddr, unWriteSize)
				return
			} else {
				*((*[16]byte)(unsafe.Pointer(targetAddr))) = *((*[16]byte)(unsafe.Pointer(sourceAddr)))
				targetAddr += 16
				sourceAddr += 16
				unWriteSize -= 16
			}
		}
	}
}

func getValue(typ reflect.Type, ptr unsafe.Pointer) interface{} {
	switch typ.Kind() {
	case reflect.Bool:
		return *((*bool)(ptr))
	case reflect.Int:
		return *((*int)(ptr))
	case reflect.Int8:
		return *((*int8)(ptr))
	case reflect.Int16:
		return *((*int16)(ptr))
	case reflect.Int32:
		return *((*int32)(ptr))
	case reflect.Int64:
		return *((*int64)(ptr))
	case reflect.Uint:
		return *((*uint)(ptr))
	case reflect.Uint8:
		return *((*uint8)(ptr))
	case reflect.Uint16:
		return *((*uint16)(ptr))
	case reflect.Uint32:
		return *((*uint32)(ptr))
	case reflect.Uint64:
		return *((*uint64)(ptr))
	case reflect.Float32:
		return *((*float32)(ptr))
	case reflect.Float64:
		return *((*float64)(ptr))
	case reflect.Complex64:
		return *((*complex64)(ptr))
	case reflect.Complex128:
		return *((*complex128)(ptr))
	case reflect.String:
		return *((*string)(ptr))
	case reflect.Struct:
		if typ == dateType {
			//fmt.Println("use *time")
			return *((*time.Time)(ptr))
		} else {
			return reflect.NewAt(typ, ptr).Elem().Interface()
		}
	default:
		return reflect.NewAt(typ, ptr).Elem().Interface()
	}
}
