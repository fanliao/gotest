package main

import (
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

type sizeTester1 interface {
	Get() string
}

type sizeTester2 interface {
	Set(i int)
}

type sizeTester1Impl struct {
}

func (this sizeTester1Impl) Get() string {
	return "hello"
}

type sizeTester2Impl struct {
	i1 int
	i2 int
	i3 int
}

func (this sizeTester2Impl) Set(i int) {
	fmt.Println(i)
}

type types struct {
	Boolean      bool
	Byte         byte
	Rune         rune
	Char         string    `orm:"size(50)"`
	Text         string    `orm:"type(text)"`
	Date         time.Time `orm:"type(date)"`
	Int          int
	Int8         int8
	Int16        int16
	Int32        int32
	Int64        int64
	Uint         uint
	Uint8        uint8
	Uint16       uint16
	Uint32       uint32
	Uint64       uint64
	Float32      float32
	Float64      float64
	Decimal      float64 `orm:"digits(8);decimals(4)"`
	Complex64    complex64
	Complex128   complex128
	Uintptr      uintptr
	String       string
	SliceInt     []int
	SliceString  []string
	ArrayInt1    [2]int
	ArrayInt2    [200]int
	ArrayString1 [4]string
	ArrayString2 [199]string
	MapIntString map[int]string
	Channel1     chan int
	Channel2     chan string
	Function1    func()
	Function2    func(int)
	Interface1   sizeTester1
	Interface2   sizeTester2
	Struct2      sizeTester2Impl
	End          bool
}

type RWTestStruct struct {
	Id   int
	Name string
}

func getFieldOffset(p interface{}, t *testing.T) map[string]uintptr {
	v := reflect.ValueOf(p).Elem()
	tp := v.Type()
	result := make(map[string]uintptr, tp.NumField()-1)

	for i := 0; i < tp.NumField()-1; i++ {
		f := tp.Field(i)
		fv := v.Field(i)
		t.Log(fv.UnsafeAddr()-v.UnsafeAddr(), f.Name, f.Type, v.Field(i+1).UnsafeAddr()-fv.UnsafeAddr())
		result[f.Name] = v.Field(i+1).UnsafeAddr() - fv.UnsafeAddr()
		t.Log(f.Name, fv.Type().Size())
	}
	return result
}

func isSameMap(m1 map[string]uintptr, m2 map[string]uintptr) (string, uintptr, uintptr, bool) {
	for k, _ := range m1 {
		if m1[k] != m2[k] {
			return k, m1[k], m2[k], false
		}
	}
	for k1, _ := range m2 {
		if m2[k1] != m1[k1] {
			return k1, m2[k1], m1[k1], false
		}
	}
	return "", 0, 0, true
}

func TestReflect(t *testing.T) {
	o := &RWTestStruct{}
	p := interface{}(o)
	//Below failed
	//p1 := *interface{}(o)
	t.Log(o, p)
}

func TestTypeSize(t *testing.T) {
	t1 := types{}
	sizes1 := getFieldOffset(&t1, t)

	t.Log("")
	t1 = types{}
	t1.String = "abbbbbcccc"
	t1.Function1 = func() {
		fmt.Println("func")
	}
	t1.Interface1 = sizeTester1Impl{}
	s2 := sizeTester2Impl{1, 2, 3}
	t1.Interface2 = s2
	t1.MapIntString = make(map[int]string)
	t.Log("size of sizeTester2Impl is ", unsafe.Sizeof(s2))
	t.Log("size of Interface2 is ", unsafe.Sizeof(t1.Interface2))
	t1.SliceInt = make([]int, 2)

	sizes2 := getFieldOffset(&t1, t)

	t.Log("")
	t1 = types{}
	b := true
	bp := &b
	t.Log("size of pointer of bool is ", unsafe.Sizeof(bp))
	//fail
	//t1.Boolean = bp
	t1.String = "keisikeie"
	t1.Function1 = func() {
		i := 1
		fmt.Println(i)
		fmt.Println("func")
	}
	sp := &sizeTester2Impl{}
	t.Log("size of pointer of struct is ", unsafe.Sizeof(sp))
	t1.Interface2 = sp
	t1.MapIntString = make(map[int]string)
	t1.SliceInt = make([]int, 2)

	sizes3 := getFieldOffset(&t1, t)
	//1个struct定义好以后，每个实例的内存布局都是相同的
	name, size1, size2, ok := isSameMap(sizes1, sizes2)
	if !ok {
		t.Log(name, size1, size2)
		t.FailNow()
	}

	//1个struct定义好以后，每个实例的内存布局应该都是相同的
	name, size1, size2, ok = isSameMap(sizes2, sizes3)
	if !ok {
		t.Log(name, size1, size2)
		t.FailNow()
	}
}

//func BenchmarkTypeCast(b *testing.B) {
//	o := &RWTestStruct{}
//	//var o2 *RWTestStruct
//	for i := 0; i < b.N; i++ {
//		var o1 interface{} = o
//		_ = o1.(*RWTestStruct)
//	}
//	b.StopTimer()
//	//b.Log(o2)
//}

func BenchmarkFastRWerGetByName(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = *((*int)(rw.GetPtrByName(p, "Id")))
		_ = *((*string)(rw.GetPtrByName(p, "Name")))
	}
	//b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkFastRWerGet(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = *((*int)(rw.GetPtr(p, 0)))
		_ = *((*string)(rw.GetPtr(p, 1)))
	}
	b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkFastRWerGetValue(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	var id interface{}
	var name interface{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id = rw.GetValue(p, 0)
		name = rw.GetValue(p, 1)
	}
	b.StopTimer()
	b.Log(id, name)
}

func BenchmarkFastGet(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	//meta := rw.GetStructMeta()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = *((*int)(FastGet(p, rw, 0)))
		_ = *((*string)(FastGet(p, rw, 1)))
	}
}

//func BenchmarkFastGetInline(b *testing.B) {
//	o := &RWTestStruct{1, "test"}
//	p := unsafe.Pointer(o)
//	rw := GetFastRWer(o, p)
//	//var id int
//	//var name string
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		_ = *((*int)(unsafe.Pointer(uintptr(p) + rw.FieldOffsetsByIndex[0])))
//		_ = *((*string)(unsafe.Pointer(uintptr(p) + rw.FieldOffsetsByIndex[1])))
//	}
//	//b.StopTimer()
//	//b.Log(id, name)
//}

func BenchmarkOriginalGet(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = o.Id
		_ = o.Name
	}
	//b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkReflectGet(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	v := reflect.ValueOf(o).Elem()
	//t := v.Type()
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.FieldByName("Id").Interface().(int)
		_ = v.FieldByName("Name").Interface().(string)
	}
	//b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkReflectGetByIndex(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	v := reflect.ValueOf(o).Elem()
	//t := v.Type()
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Field(0).Interface().(int)
		_ = v.Field(1).Interface().(string)
	}
	//b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkReflectGetAddrByIndex(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	v := reflect.ValueOf(o).Elem()
	//t := v.Type()
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Field(0).UnsafeAddr()
		_ = v.Field(1).UnsafeAddr()
	}
	//b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkFastRWerSetByName(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	id := 1111111
	name := "test unsafe set, great!"
	idAddr := uintptr(unsafe.Pointer(&id))
	nameAddr := uintptr(unsafe.Pointer(&name))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.SetByName(p, "Id", idAddr)
		rw.SetByName(p, "Name", nameAddr)
	}
}

func BenchmarkFastRWerSet(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	id := 1111111
	name := "test unsafe set, great!"
	idAddr := uintptr(unsafe.Pointer(&id))
	nameAddr := uintptr(unsafe.Pointer(&name))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.Set(p, 0, idAddr)
		rw.Set(p, 1, nameAddr)
	}
}

func BenchmarkFastSet(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	b.ResetTimer()
	id := 1111111
	name := "test unsafe set, great!"
	idAddr := uintptr(unsafe.Pointer(&id))
	nameAddr := uintptr(unsafe.Pointer(&name))

	for i := 0; i < b.N; i++ {
		//bs := []bytes
		FastSet(p, rw, 0, idAddr)
		FastSet(p, rw, 1, nameAddr)
	}
	b.StopTimer()
	//b.Log(o.Id, o.Name)
}

func BenchmarkCopyVar(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	b.ResetTimer()
	id := 1111111
	name := "test unsafe set, great!"
	idAddr := uintptr(unsafe.Pointer(&id))
	nameAddr := uintptr(unsafe.Pointer(&name))

	for i := 0; i < b.N; i++ {
		size1 := (rw.FieldSizeByIndex[0])
		offset1 := uintptr(p) + rw.FieldOffsetsByIndex[0]
		//bs := []bytes
		copyVar(offset1, idAddr, size1)

		size2 := (rw.FieldSizeByIndex[1])
		offset2 := uintptr(p) + rw.FieldOffsetsByIndex[1]
		copyVar(offset2, nameAddr, size2)
	}
	b.StopTimer()
	//b.Log(o.Id, o.Name)
}

func BenchmarkOriginalSet(b *testing.B) {
	o := &RWTestStruct{1, "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o.Id = 11
		o.Name = "test"
	}
	b.StopTimer()
	//b.Log(o.Id, o.Name)
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
	//b.Log(o.Id, o.Name)
}
