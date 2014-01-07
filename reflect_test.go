package main

//import (
//	"reflect"
//	"testing"
//	"unsafe"
//)

//type User struct {
//	Id   int
//	Name string
//}

//type User1 struct {
//	Id   int64
//	Name string
//}

////func testPerformance(o interface{}, b *testing.B) {
////	b.Log("test reflect ", b.N)

////	v := reflect.ValueOf(o)
////	fieldId := v.Elem().FieldByName("Id")
////	fieldName := v.Elem().FieldByName("Name")

////	b.ReportAllocs()
////	b.ResetTimer()

////	for i := 0; i < b.N; i++ {
////		fieldId.SetInt(int64(i))
////		fieldName.SetString("reflect")
////	}

////	//u := o.(*User)
////	//t1 = time.Now()
////	//for i = 0; i < b.N; i++ {
////	//	u.Id = 3000 + i
////	//	u.Name = "normal"
////	//}
////	//d = time.Since(t1)
////	//fmt.Println(time.Now())
////	//fmt.Println(d.Nanoseconds())
////	//fmt.Println(t1)
////	//fmt.Println(u.Id)
////	//fmt.Println("")
////}

//func BenchmarkReflectSet1(b *testing.B) {
//	u := User{1, "aaa"}
//	//b.Log("test reflect set ", b.N)

//	v := reflect.ValueOf(&u)
//	fieldId := v.Elem().FieldByName("Id")
//	fieldName := v.Elem().FieldByName("Name")

//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		fieldId.SetInt(int64(i))
//		fieldName.SetString("reflect")
//	}
//	//testPerformance(&u, b)
//}

//func BenchmarkSet(b *testing.B) {
//	u := User1{1, "aaa"}
//	//b.Log("test set ", b.N)

//	b.ResetTimer()
//	var i int
//	for i = 0; i < b.N; i++ {
//		u.Id = int64(i)
//		u.Name = "normal"
//	}
//}

//func BenchmarkReflectGet(b *testing.B) {
//	u := User{1, "aaa"}
//	//b.Log("test reflect set ", b.N)

//	v := reflect.ValueOf(&u)
//	fieldId := v.Elem().FieldByName("Id")
//	fieldName := v.Elem().FieldByName("Name")

//	var id int
//	var name string
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		id = int(fieldId.Int())
//		name = fieldName.String()
//	}
//	b.StopTimer()
//	b.Log(id, name)
//	//testPerformance(&u, b)
//}

//func BenchmarkGet(b *testing.B) {
//	u := &User1{1, "aaa"}
//	//b.Log("test set ", b.N)

//	var id int
//	var name string
//	b.ResetTimer()
//	var i int
//	for i = 0; i < b.N; i++ {
//		id = int(u.Id)
//		name = u.Name
//	}
//	b.StopTimer()
//	b.Log(id, name)
//}

//func BenchmarkGetByPointer(b *testing.B) {
//	u := &User1{11, "aaa_pointer"}
//	//b.Log("test set ", b.N)

//	var id int
//	var name string
//	v := reflect.ValueOf(u).Elem()
//	fieldId := v.FieldByName("Id")
//	fieldName := v.FieldByName("Name")
//	//b.Log("id", unsafe.Offsetof(u.Id), fieldId.UnsafeAddr()-v.UnsafeAddr())
//	//idp := uintptr(unsafe.Pointer(u)) + fieldId.UnsafeAddr()
//	//b.Log("name", unsafe.Offsetof(u.Name), fieldName.UnsafeAddr()-v.UnsafeAddr())
//	//namep := uintptr(unsafe.Pointer(u)) + fieldName.UnsafeAddr()
//	b.ResetTimer()
//	var i int
//	for i = 0; i < b.N; i++ {
//		//id = int(u.Id)
//		//name = u.Name
//		idp := uintptr(unsafe.Pointer(u)) + (fieldId.UnsafeAddr() - v.UnsafeAddr())
//		namep := uintptr(unsafe.Pointer(u)) + (fieldName.UnsafeAddr() - v.UnsafeAddr())
//		id = int(*(*int64)(unsafe.Pointer(idp)))
//		name = *(*string)(unsafe.Pointer(namep))
//	}
//	b.StopTimer()
//	b.Log("result", id, name)
//}
