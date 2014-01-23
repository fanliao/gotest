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
	Cash float32
	Date time.Time
	Ptr  *RWTestStruct
}

func TestNil(t *testing.T) {
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

func TestFastRWerValue(t *testing.T) {
	//f, _ := os.Create("profile_file")
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	var id interface{}
	var name interface{}
	var date interface{}
	var cash interface{}
	var ptr *RWTestStruct
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()
	for i := 0; i < 1; i++ {
		id = rw.Value(p, 0)
		name = rw.Value(p, 1)
		cash = rw.Value(p, 2)
		date = rw.Value(p, 3)
		ptr1 := unsafe.Pointer(&ptr)
		rw.CopyPtr(p, 4, ptr1)
	}
	t.Log(id, name, cash, date, ptr)
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
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = *((*int)(rw.PtrByName(p, "Id")))
		_ = *((*string)(rw.PtrByName(p, "Name")))
		_ = *((*float32)(rw.PtrByName(p, "Cash")))
		_ = *((*time.Time)(rw.PtrByName(p, "Date")))
		_ = *((*RWTestStruct)(rw.PtrByName(p, "Ptr")))
	}
	//b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkFastRWerGet(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = *((*int)(rw.Ptr(p, 0)))
		_ = *((*string)(rw.Ptr(p, 1)))
		_ = *((*float32)(rw.Ptr(p, 2)))
		_ = *((*time.Time)(rw.Ptr(p, 3)))
		_ = *((*RWTestStruct)(rw.Ptr(p, 4)))
	}
	b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkFastRWerValue(b *testing.B) {
	//f, _ := os.Create("profile_file")
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	var id interface{}
	var name interface{}
	var date interface{}
	var cash interface{}
	//var ptr interface{}
	var ptr *RWTestStruct
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id = rw.Value(p, 0)
		name = rw.Value(p, 1)
		cash = rw.Value(p, 2)
		date = rw.Value(p, 3)
		//ptr = rw.Value(p, 4)
		ptr1 := unsafe.Pointer(&ptr)
		rw.CopyPtr(p, 4, ptr1)
	}
	b.StopTimer()
	b.Log(id, name, cash, date, ptr)
}

func BenchmarkFastGet(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	//meta := rw.GetStructMeta()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = *((*int)(FastGet(p, rw, 0)))
		_ = *((*string)(FastGet(p, rw, 1)))
		_ = *((*float32)(FastGet(p, rw, 2)))
		_ = *((*time.Time)(FastGet(p, rw, 3)))
		_ = *((*RWTestStruct)(FastGet(p, rw, 4)))
	}
}

//func BenchmarkFastGetInline(b *testing.B) {
//	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
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
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = o.Id
		_ = o.Name
		_ = o.Cash
		_ = o.Date
		_ = o.Ptr
	}
	//b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkReflectGet(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	v := reflect.ValueOf(o).Elem()
	//t := v.Type()
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.FieldByName("Id").Interface().(int)
		_ = v.FieldByName("Name").Interface().(string)
		_ = v.FieldByName("Cash").Interface().(float32)
		_ = v.FieldByName("Date").Interface().(time.Time)
		_ = v.FieldByName("Ptr").Interface().(*RWTestStruct)
	}
	//b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkReflectGetByIndex(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	v := reflect.ValueOf(o).Elem()
	//t := v.Type()
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Field(0).Interface().(int)
		_ = v.Field(1).Interface().(string)
		_ = v.Field(2).Interface().(float32)
		_ = v.Field(3).Interface().(time.Time)
		_ = v.Field(4).Interface().(*RWTestStruct)
	}
	//b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkReflectGetAddrByIndex(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	v := reflect.ValueOf(o).Elem()
	//t := v.Type()
	//var id int
	//var name string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Field(0).UnsafeAddr()
		_ = v.Field(1).UnsafeAddr()
		_ = v.Field(2).UnsafeAddr()
		_ = v.Field(3).UnsafeAddr()
		_ = v.Field(4).UnsafeAddr()
	}
	//b.StopTimer()
	//b.Log(id, name)
}

func BenchmarkFastRWerSetPtrByName(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	//id := 1111111
	//name := "test unsafe set, great!"
	//cash := 22.22
	//date := time.Now()
	//ptr := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	id, name, cash, date, ptr := testStruct()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idAddr := uintptr(unsafe.Pointer(&id))
		nameAddr := uintptr(unsafe.Pointer(&name))
		cashAddr := uintptr(unsafe.Pointer(&cash))
		dateAddr := uintptr(unsafe.Pointer(&date))
		ptrAddr := uintptr(unsafe.Pointer(&ptr))

		rw.SetPtrByName(p, "Id", idAddr)
		rw.SetPtrByName(p, "Name", nameAddr)
		rw.SetPtrByName(p, "Cash", cashAddr)
		rw.SetPtrByName(p, "Date", dateAddr)
		rw.SetPtrByName(p, "Ptr", ptrAddr)
	}
}

func BenchmarkFastRWerSetValueByName(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	id, name, cash, date, ptr := testStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.SetValueByName(p, "Id", id)
		rw.SetValueByName(p, "Name", name)
		rw.SetValueByName(p, "Cash", cash)
		rw.SetValueByName(p, "Date", date)
		rw.SetValueByName(p, "Ptr", ptr)
	}
}

func BenchmarkFastRWerSetPtr(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	id, name, cash, date, ptr := testStruct()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idAddr := uintptr(unsafe.Pointer(&id))
		nameAddr := uintptr(unsafe.Pointer(&name))
		cashAddr := uintptr(unsafe.Pointer(&cash))
		dateAddr := uintptr(unsafe.Pointer(&date))
		ptrAddr := uintptr(unsafe.Pointer(&ptr))

		rw.SetPtr(p, 0, idAddr)
		rw.SetPtr(p, 1, nameAddr)
		rw.SetPtr(p, 2, cashAddr)
		rw.SetPtr(p, 3, dateAddr)
		rw.SetPtr(p, 4, ptrAddr)
	}
}

func BenchmarkFastRWerSetValue(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)

	id, name, cash, date, ptr := testStruct()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.SetValue(p, 0, id)
		rw.SetValue(p, 1, name)
		rw.SetValue(p, 2, cash)
		rw.SetValue(p, 3, date)
		rw.SetValue(p, 4, ptr)
	}
}

func BenchmarkFastSet(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	b.ResetTimer()

	id, name, cash, date, ptr := testStruct()
	for i := 0; i < b.N; i++ {
		idAddr := uintptr(unsafe.Pointer(&id))
		nameAddr := uintptr(unsafe.Pointer(&name))
		cashAddr := uintptr(unsafe.Pointer(&cash))
		dateAddr := uintptr(unsafe.Pointer(&date))
		ptrAddr := uintptr(unsafe.Pointer(&ptr))

		FastSet(p, rw, 0, idAddr)
		FastSet(p, rw, 1, nameAddr)
		FastSet(p, rw, 2, cashAddr)
		FastSet(p, rw, 3, dateAddr)
		FastSet(p, rw, 4, ptrAddr)
	}
	b.StopTimer()
	//b.Log(o.Id, o.Name)
}

func BenchmarkCopyVar(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	p := unsafe.Pointer(o)
	rw := GetFastRWer(o)
	b.ResetTimer()

	id, name, cash, date, ptr := testStruct()
	for i := 0; i < b.N; i++ {
		idAddr := uintptr(unsafe.Pointer(&id))
		nameAddr := uintptr(unsafe.Pointer(&name))
		cashAddr := uintptr(unsafe.Pointer(&cash))
		dateAddr := uintptr(unsafe.Pointer(&date))
		ptrAddr := uintptr(unsafe.Pointer(&ptr))

		offset1 := uintptr(p) + rw.FieldOffsetsByIndex[0]
		copyVar(offset1, idAddr, rw.FieldSizeByIndex[0])

		offset2 := uintptr(p) + rw.FieldOffsetsByIndex[1]
		copyVar(offset2, nameAddr, rw.FieldSizeByIndex[1])

		offset3 := uintptr(p) + rw.FieldOffsetsByIndex[2]
		copyVar(offset3, cashAddr, rw.FieldSizeByIndex[2])

		offset4 := uintptr(p) + rw.FieldOffsetsByIndex[3]
		copyVar(offset4, dateAddr, rw.FieldSizeByIndex[3])

		offset5 := uintptr(p) + rw.FieldOffsetsByIndex[4]
		copyVar(offset5, ptrAddr, rw.FieldSizeByIndex[4])
	}
	b.StopTimer()
	//b.Log(o.Id, o.Name)
}

func BenchmarkOriginalSet(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	b.ResetTimer()
	id, name, cash, date, ptr := testStruct()
	for i := 0; i < b.N; i++ {
		o.Id = id
		o.Name = name
		o.Cash = cash
		o.Date = date
		o.Ptr = ptr
	}
	b.StopTimer()
	//b.Log(o.Id, o.Name)
}

func BenchmarkReflectSet(b *testing.B) {
	o := &RWTestStruct{1, "test", 1.1, time.Now(), nil}
	v := reflect.ValueOf(o).Elem()
	//t := v.Type()
	b.ResetTimer()
	id, name, cash, date, ptr := testStruct()
	for i := 0; i < b.N; i++ {
		v.Field(0).Set(reflect.ValueOf(id))
		v.Field(1).Set(reflect.ValueOf(name))
		v.Field(2).Set(reflect.ValueOf(cash))
		v.Field(3).Set(reflect.ValueOf(date))
		v.Field(4).Set(reflect.ValueOf(ptr))
	}
	b.StopTimer()
	//b.Log(o.Id, o.Name)
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

func testStruct() (int, string, float32, time.Time, *RWTestStruct) {
	id := 1111111
	name := "test unsafe set, great!"
	var cash float32 = 22.22
	date := time.Now()
	var ptr *RWTestStruct = nil //&RWTestStruct{1, "test", 1.1, time.Now(), nil}

	return id, name, cash, date, ptr
}
