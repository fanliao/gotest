package main

import (
	"errors"
	//"fmt"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

const ptrSize = unsafe.Sizeof((*byte)(nil))

var bigEndian bool

func init() {
	x := 0x1234
	p := unsafe.Pointer(&x)
	p2 := (*[ptrSize]byte)(p)
	if p2[0] == 0 {
		//fmt.Println("本机器：大端")
		bigEndian = true
	} else {
		//fmt.Println("本机器：小端")
		bigEndian = false
	}

}

type metaCache struct {
	lock  *sync.RWMutex
	metas map[reflect.Type]*FastRW
}

type structMeta struct {
	//IsFixedSize         bool
	FieldIndexsByName map[string]int
	//FieldOffsetsByIndex []uintptr
	FieldNamesByIndex []string
	//FieldSizeByIndex    []uintptr
	//FieldTypesByIndex   []*rtype
	FieldsByIndex []*structField
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
	fld := this.FieldsByIndex[i]
	return getValue(fld.typ, unsafe.Pointer(uintptr(obj)+fld.offset))
}

func (this *FastRW) CopyPtr(obj unsafe.Pointer, i int, target unsafe.Pointer) {
	fld := this.FieldsByIndex[i]
	ptr := unsafe.Pointer(uintptr(obj) + fld.offset)
	copyVar(uintptr(target), uintptr(ptr), fld.typ.size)
}

func (this *FastRW) PtrByName(obj unsafe.Pointer, fieldName string) unsafe.Pointer {
	return this.Ptr(obj, this.FieldIndexsByName[fieldName])
}

func (this *FastRW) GetValueByName(obj unsafe.Pointer, fieldName string) interface{} {
	return this.Value(obj, this.FieldIndexsByName[fieldName])
}

func (this *FastRW) SetPtr(obj unsafe.Pointer, i int, source uintptr) {
	fld := this.FieldsByIndex[i]
	target := uintptr(obj) + fld.offset   // this.FieldOffsetsByIndex[i]
	copyVar(target, source, fld.typ.size) // this.FieldSizeByIndex[i])
}

func (this *FastRW) SetPtrByName(obj unsafe.Pointer, fieldName string, source uintptr) {
	if i, isExist := this.FieldIndexsByName[fieldName]; isExist {
		fld := this.FieldsByIndex[i]
		target := uintptr(obj) + fld.offset   // this.FieldOffsetsByIndex[i]
		copyVar(target, source, fld.typ.size) //this.FieldSizeByIndex[i])
	}
}

func (this *FastRW) SetValue(obj unsafe.Pointer, i int, source interface{}) (err error) {
	//fmt.Println("SetValue", source)
	//fmt.Println(*((*interfaceHeader)(unsafe.Pointer(&source))))
	fld := this.FieldsByIndex[i]
	target := uintptr(obj) + fld.offset //this.FieldOffsetsByIndex[i]
	size := fld.typ.size                // this.FieldSizeByIndex[i]

	var dataPtr uintptr = 0
	//if source != nil {
	s := *((*interfaceHeader)(unsafe.Pointer(&source)))
	t := s.typ
	//if t.Kind() == reflect.Ptr || t.size > ptrSize {
	dataPtr = s.word
	//} else {
	//	dataPtr = s.word
	//}
	//fmt.Println(t.Kind(), size, t.size, dataPtr)
	fldTyp, valTyp := fld.typ, t.Kind() // this.FieldTypesByIndex[i], t.Kind()
	if fldTyp.Kind() == reflect.Ptr {
		if valTyp != reflect.Ptr {
			err = errors.New("expect a pointer, but actual is " + valTyp.String())
		} else {
			//如果两个变量都是指针类型，则直接copy原指针指向的地址
			copyUint(target, dataPtr, size)
		}
	} else {
		if (valTyp == reflect.Ptr && dataPtr != 0) || t.size > ptrSize {
			//如果i是指针或者i的长度超过了一个字的长度，则s.word是一个指向数据的指针
			copyVar(target, dataPtr, size)
		} else {
			//copyVar(target, uintptr(unsafe.Pointer(&dataPtr)), size)
			copyUint(target, dataPtr, size)
		}
	}
	return
	//_, _, _ = dataPtr, target, size
}

func (this *FastRW) SetValueByName(obj unsafe.Pointer, fieldName string, source interface{}) {
	if i, isExist := this.FieldIndexsByName[fieldName]; isExist {
		//fmt.Println(*((*interfaceHeader)(unsafe.Pointer(&source))))
		this.SetValue(obj, i, source)
	}
}

func FastGet(obj unsafe.Pointer, this *FastRW, i int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(obj) + this.FieldsByIndex[i].offset) // this.FieldOffsetsByIndex[i])
}

func FastSet(obj unsafe.Pointer, this *FastRW, i int, source uintptr) {
	fld := this.FieldsByIndex[i]
	target := uintptr(obj) + fld.offset   // this.FieldOffsetsByIndex[i]
	copyVar(target, source, fld.typ.size) // this.FieldSizeByIndex[i])
}

func fastGet(obj unsafe.Pointer, this *FastRW, i int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(obj) + this.FieldsByIndex[i].offset) // this.FieldOffsetsByIndex[i])
}

////Get a FastRWer implement class by a pointer of struct
////obj must be a pointer of struct value
//func GetFastRWer_bak(obj interface{}) *FastRW {
//	v := reflect.Indirect(reflect.ValueOf(obj))
//	t := v.Type()

//	if fastRW := mc.get(t); fastRW != nil {
//		return fastRW
//	} else {

//		meta := structMeta{}

//		objAddr := v.UnsafeAddr()
//		numField := t.NumField()

//		meta.FieldIndexsByName = make(map[string]int, numField)
//		meta.FieldOffsetsByIndex = make([]uintptr, numField, numField)
//		meta.FieldNamesByIndex = make([]string, numField, numField)
//		meta.FieldSizeByIndex = make([]uintptr, numField, numField)
//		//meta.FieldTypesByIndex = make([]*reflect.Type, numField, numField)
//		for i := 0; i < t.NumField(); i++ {
//			fType := t.Field(i)
//			f := v.Field(i)
//			if !f.CanInterface() {
//				continue
//			}
//			meta.FieldOffsetsByIndex[i] = f.UnsafeAddr() - objAddr
//			meta.FieldIndexsByName[fType.Name] = i
//			meta.FieldNamesByIndex[i] = fType.Name
//			meta.FieldSizeByIndex[i] = f.Type().Size()
//			//t := f.Type()
//			//meta.FieldTypesByIndex[i] = &t
//		}

//		fastRW = newFastRWImpl(meta)
//		mc.set(t, fastRW)
//		return fastRW
//	}

//}

//Get a FastRWer implement class by a pointer of struct
//obj must be a pointer of struct value
func GetFastRWer(obj interface{}) *FastRW {
	fs1 := faceToStruct(obj)
	var fs interfaceHeader
	if fs1.typ.Kind() == reflect.Ptr {
		elemTyp := (*rtype)(unsafe.Pointer(((*ptrType)(unsafe.Pointer(fs1.typ))).elem))
		if elemTyp.Kind() == reflect.Struct {
			fs = interfaceHeader{elemTyp, fs1.word}
		} else {
			return nil
		}
	} else if fs1.typ.Kind() != reflect.Struct {
		return nil
	} else {
		fs = fs1
	}

	s := *((*structType)(unsafe.Pointer(fs.typ)))

	//if fastRW := mc.get(t); fastRW != nil {
	//	return fastRW
	//} else {

	meta := structMeta{}

	numField := len(s.fields)
	//fmt.Println("numField", len(s.fields))

	meta.FieldIndexsByName = make(map[string]int, numField)
	//meta.FieldOffsetsByIndex = make([]uintptr, numField, numField)
	meta.FieldNamesByIndex = make([]string, numField, numField)
	//meta.FieldSizeByIndex = make([]uintptr, numField, numField)
	//meta.FieldTypesByIndex = make([]*rtype, numField, numField)
	meta.FieldsByIndex = make([]*structField, numField, numField)
	for i := 0; i < numField; i++ {
		fld := s.fields[i]
		//fmt.Println(i)
		var name string
		if fld.name == nil {
			name = *(fld.typ.string)
		} else {
			name = *(fld.name)
		}
		//fmt.Println(i, name, fld.typ.size, *fld.typ.string)
		//meta.FieldOffsetsByIndex[i] = fld.offset
		meta.FieldIndexsByName[name] = i
		meta.FieldNamesByIndex[i] = name
		//meta.FieldSizeByIndex[i] = fld.typ.size
		//fmt.Println(fld.typ.size, *fld.typ.string)
		//t := f.Type()
		//meta.FieldTypesByIndex[i] = fld.typ
		meta.FieldsByIndex[i] = &(s.fields[i])
		//v := f.Interface()
	}

	fastRW := newFastRWImpl(meta)
	//mc.set(t, fastRW)
	return fastRW
	//}

}

//factory function
func newFastRWImpl(meta structMeta) *FastRW {
	return &FastRW{meta}
}

func copyUint(target uintptr, source uintptr, size uintptr) {
	//sourceAddr := uintptr()
	//fmt.Printf("十六进制：%X\n", source)
	//fmt.Printf("十进制：%d\n", source)
	var i uintptr
	for i = 0; i < size; i++ {
		//fmt.Printf("十六进制：%X\n", (byte)((source>>((size-1-i)*8))&0xff))
		if bigEndian {
			*((*byte)(unsafe.Pointer(target + i))) = (byte)((source >> ((size - 1 - i) * 8)) & 0xff)
		} else {
			*((*byte)(unsafe.Pointer(target + i))) = (byte)((source >> ((i) * 8)) & 0xff)
		}
	}
	//fmt.Printf("十六进制结果：：%X\n", *((*int32)(unsafe.Pointer(target))))
	//fmt.Printf("十进制结果：%d\n", *((*int32)(unsafe.Pointer(target))))
}

func copyVar(target uintptr, source uintptr, size uintptr) {
	//fmt.Println("target=", target, " source=", source)
	switch size {
	case 1:
		*((*byte)(unsafe.Pointer(target))) = *((*byte)(unsafe.Pointer(source)))
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
	case 16:
		*((*[16]byte)(unsafe.Pointer(target))) = *((*[16]byte)(unsafe.Pointer(source)))
	case 24:
		*((*[24]byte)(unsafe.Pointer(target))) = *((*[24]byte)(unsafe.Pointer(source)))
	default:
		unWriteSize := size
		targetAddr := target
		sourceAddr := source
		for {
			if unWriteSize <= 8 || unWriteSize == 16 || unWriteSize == 24 {
				copyVar(targetAddr, sourceAddr, unWriteSize)
				return
			} else {
				*((*[8]byte)(unsafe.Pointer(targetAddr))) = *((*[8]byte)(unsafe.Pointer(sourceAddr)))
				targetAddr += 8
				sourceAddr += 8
				unWriteSize -= 8
			}
		}
	}
	//for i := 0; uintptr(i) < size; i++ {
	//	*((*byte)(unsafe.Pointer(target + uintptr(i)))) = *((*byte)(unsafe.Pointer(source + uintptr(i))))
	//}
}

func getValue(typ *rtype, ptr unsafe.Pointer) interface{} {
	t := typ
	switch t.Kind() {
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
		if *(t.string) == dateType.Name() {
			//fmt.Println("use *time")
			return *((*time.Time)(ptr))
		} else {
			return toFace(toFaceHeader(typ, ptr)) //reflect.NewAt(t, ptr).Elem().Interface()
		}
	default:
		return toFace(toFaceHeader(typ, ptr)) //reflect.NewAt(t, ptr).Elem().Interface()
	}
}

//the below code is copy from go source code

type flag1 uintptr

const (
	flagRO flag1 = 1 << iota
	flagIndir
	flagAddr
	flagMethod
	flagKindShift         = iota
	flagKindWidth         = 5 // there are 27 kinds
	flagKindMask    flag1 = 1<<flagKindWidth - 1
	flagMethodShift       = flagKindShift + flagKindWidth
)

//from type.go
// High bit says whether type has
// embedded pointers,to help garbage collector.
const (
	kindMask       = 0x7f
	kindNoPointers = 0x80
)

// interfaceHeader is the header for an interface{} value. it is copied from unsafe.emptyInterface
type interfaceHeader struct {
	typ  *rtype
	word uintptr
}

func faceToDataPtr(i interface{}) uintptr {
	s := *((*interfaceHeader)(unsafe.Pointer(&i)))
	return s.word
}

func faceToStruct(i interface{}) interfaceHeader {
	s := *((*interfaceHeader)(unsafe.Pointer(&i)))
	return s
}

func toFaceHeader(typ *rtype, ptr unsafe.Pointer) *interfaceHeader {
	if typ.Kind() == reflect.Ptr {
		return &interfaceHeader{((*ptrType)(unsafe.Pointer(typ))).elem, uintptr(ptr)}
	}

	if typ.size > ptrSize {
		return &interfaceHeader{typ, uintptr(ptr)}
	} else {
		var word uintptr
		copyVar(uintptr(unsafe.Pointer(&word)), uintptr(ptr), typ.size)
		return &interfaceHeader{typ, word}
	}
}

func toFace(eface *interfaceHeader) interface{} {
	return *((*interface{})(unsafe.Pointer(eface)))
}

// rtype is the common implementation of most values.
// It is embedded in other, public struct types, but always
// with a unique tag like `reflect:"array"` or `reflect:"ptr"`
// so that code cannot convert from, say, *arrayType to *ptrType.
type rtype struct {
	size              uintptr        // size in bytes
	hash              uint32         // hash of type; avoids computation in hash tables
	_                 uint8          // unused/padding
	align             uint8          // alignment of variable with this type
	fieldAlign        uint8          // alignment of struct field with this type
	kind              uint8          // enumeration for C
	alg               *uintptr       // algorithm table (../runtime/runtime.h:/Alg)
	gc                unsafe.Pointer // garbage collection data
	string            *string        // string form; unnecessary but undeniably useful
	ptrToUncommonType uintptr        // (relatively) uncommon fields
	ptrToThis         *rtype         // type for pointer to this type, if used in binary or has methods
}

func (t *rtype) Kind() reflect.Kind { return reflect.Kind(t.kind & kindMask) }

// arrayType represents a fixed array type.
type arrayType struct {
	rtype `reflect:"array"`
	elem  *rtype // array element type
	slice *rtype // slice type
	len   uintptr
}

// chanType represents a channel type.
type chanType struct {
	rtype `reflect:"chan"`
	elem  *rtype  // channel element type
	dir   uintptr // channel direction (ChanDir)
}

// funcType represents a function type.
type funcType struct {
	rtype     `reflect:"func"`
	dotdotdot bool     // last input parameter is ...
	in        []*rtype // input parameter types
	out       []*rtype // output parameter types
}

// imethod represents a method on an interface type
type imethod struct {
	name    *string // name of method
	pkgPath *string // nil for exported Names; otherwise import path
	typ     *rtype  // .(*FuncType) underneath
}

// interfaceType represents an interface type.
type interfaceType struct {
	rtype   `reflect:"interface"`
	methods []imethod // sorted by hash
}

// mapType represents a map type.
type mapType struct {
	rtype  `reflect:"map"`
	key    *rtype // map key type
	elem   *rtype // map element (value) type
	bucket *rtype // internal bucket structure
	hmap   *rtype // internal map header
}

// ptrType represents a pointer type.
type ptrType struct {
	rtype `reflect:"ptr"`
	elem  *rtype // pointer element (pointed at) type
}

// sliceType represents a slice type.
type sliceType struct {
	rtype `reflect:"slice"`
	elem  *rtype // slice element type
}

// Struct field
type structField struct {
	name    *string // nil for embedded fields
	pkgPath *string // nil for exported Names; otherwise import path
	typ     *rtype  // type of field
	tag     *string // nil if no tag
	offset  uintptr // byte offset of field within struct
}

// structType represents a struct type.
type structType struct {
	rtype  `reflect:"struct"`
	fields []structField // sorted by offset
}
