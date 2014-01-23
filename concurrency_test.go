package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func testChan() {
	ch := make(chan string)
	sendData(ch)
	go getData(ch)
	time.Sleep(1e9)
}

func sendData(ch chan string) {
	ch <- "One"
	ch <- "Two"
	ch <- "Three"
	ch <- "Four"
	ch <- "Five2"
}

func getData(ch chan string) {
	var input string
	for {
		input = <-ch
		fmt.Printf("%s\n", input)
	}
}

type mapChan struct {
	m map[string]int
	c chan func()
}

func newMapChan() *mapChan {
	mc := &mapChan{make(map[string]int), make(chan func())}
	go mc.backend()
	return mc
}

func (this mapChan) backend() {
	for f := range this.c {
		f()
	}
}

func (this mapChan) get(key string) int {
	fc := make(chan int)
	this.c <- func() { fc <- this.m[key] }
	return <-fc
}

func (this mapChan) set(key string, value int) {
	this.c <- func() { this.m[key] = value }
}

func BenchmarkMapChan(b *testing.B) {
	k := "testkey"
	v := 100
	r := 1
	b.ResetTimer()
	mapC := newMapChan()
	for i := 0; i < b.N; i++ {
		mapC.set(k, v)
		r = mapC.get(k)
	}
	b.Log(r)
}

type mapLock struct {
	lock *sync.RWMutex
	m    map[string]int
}

func newMapLock() *mapLock {
	mc := &mapLock{}
	mc.lock = new(sync.RWMutex)
	mc.m = make(map[string]int)
	return mc
}

func (this *mapLock) get(typ string) int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	val, _ := this.m[typ]
	return val
}

func (this *mapLock) set(typ string, rw int) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.m[typ] = rw
}

func BenchmarkMapLock(b *testing.B) {
	k := "testkey"
	v := 100
	r := 1
	b.ResetTimer()
	mapC := newMapLock()
	for i := 0; i < b.N; i++ {
		mapC.set(k, v)
		r = mapC.get(k)
	}
	b.Log(r)
}
