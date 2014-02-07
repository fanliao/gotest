package main

import (
	"reflect"
	"testing"
)

type test1er interface {
	Get() string
}

type testerImpl struct {
	GetImpl func(this testerImpl) string
}

func (this testerImpl) Get() string {
	//return "default"
	return this.GetImpl(this)
}

func GetTesterImpl() *testerImpl {
	t := &testerImpl{}
	t.GetImpl = func(this testerImpl) string {
		return "defaulImpl"
	}
	return t
}

func TestIoc(t *testing.T) {
	//newGet用于替换原来的Get()实现，但go要求的参数和返回值极不友好，难以使用
	newGet := func(in []reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf("ioc")}
	}
	tImpl := &testerImpl{}
	tType := reflect.ValueOf(tImpl).Elem()
	for i := 0; i < tType.Type().NumMethod(); i++ {
		m := tType.Type().Method(i)
		t.Log(m.Func)
		v := reflect.MakeFunc(m.Func.Type(), newGet)
		t.Log(v)
		//failed
		//tType.Method(i).Set(v)
		t.Log(tType.Method(i).Kind())
		t.Log(tType.Method(i).CanSet())
	}
}

//测试运行时修改方法的实现
func TestIoc2(t *testing.T) {
	tImpl := GetTesterImpl()
	if tImpl.Get() != "defaulImpl" {
		t.Error("except defaulImpl, but " + tImpl.Get())
		t.FailNow()
	}

	tImpl.GetImpl = func(this testerImpl) string {
		return "IOCImpl"
	}
	if tImpl.Get() != "IOCImpl" {
		t.Error("except IOCImpl, but " + tImpl.Get())
		t.FailNow()
	}
}
