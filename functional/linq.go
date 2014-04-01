package main

import (
	"fmt"
	//"time"
	"github.com/fanliao/go-promise"
	"reflect"
)

// the struct and interface about data source---------------------------------------------------
type chunk struct {
	data  []interface{}
	start int
	end   int
}

const (
	SOURCE_BLOCK int = iota
	SOURCE_CHUNK
)

type source interface {
	Typ() int                   //block or chunk?
	Itr() func() (*chunk, bool) //return a itr function
	NumberOfBlock() int         //get the degree of parallel
}

type blockSource struct {
	data          []interface{}
	numberOfBlock int //the count of block
	sizeOfBlock   int //the size of block
}

func (this blockSource) Typ() int {
	return SOURCE_BLOCK
}

func (this blockSource) Itr() func() (*chunk, bool) {
	i := 0
	len := len(this.data)
	return func() (c *chunk, ok bool) {
		if i < this.numberOfBlock {
			end := (i+1)*this.sizeOfBlock - 1
			if end >= len {
				end = len
			}
			c, ok = &chunk{this.data, i * this.sizeOfBlock, end}, true
			i++
			return
		} else {
			return nil, false
		}
	}
}

func (this blockSource) NumberOfBlock() int {
	return this.numberOfBlock
}

type chunkSource struct {
	data          chan *chunk
	numberOfBlock int
}

func (this chunkSource) Typ() int {
	return SOURCE_CHUNK
}

func (this chunkSource) Itr() func() (*chunk, bool) {
	ch := this.data
	return func() (*chunk, bool) {
		c, ok := <-ch
		return c, ok
	}
}

func (this chunkSource) NumberOfBlock() int {
	return this.numberOfBlock
}

func (this chunkSource) Close() {
	close(this.data)
}

type stepAction func(source) source

//the function of step-------------------------------------------------------------------------
type step struct {
	src    source
	act    stepAction
	result source
	degree int
}

func where(sure func(interface{}) bool) stepAction {
	return stepAction(func(src source) source {
		switch s := src.(type) {
		case blockSource:
			itr := s.Itr()
			f := makeTasks(s, func() []interface{} {
				chunk, _ := itr()
				result := make([]interface{}, 0, chunk.end-chunk.start+2)
				forSlice(chunk.data[chunk.start:chunk.end+1], func(v interface{}) {
					if sure(v) {
						result = append(result, v)
					}
				})
				result = append(result, true)
				return result
			})

			if results, typ := f.Get(); typ != promise.RESULT_SUCCESS {
				//todo
				return nil
			} else {
				result := expandSlice(results)
				return blockSource{result, s.numberOfBlock, ceilSplitSize(len(result), s.numberOfBlock)}
			}
		case chunkSource:
			itr := s.Itr()
			out := make(chan interface{})

			f := makeTasks(s, func() []interface{} {
				for {
					if chunk, ok := itr(); ok {
						if reflect.ValueOf(chunk).IsNil() {
							s.Close()
							break
						}
						result := make([]interface{}, 0, chunk.end-chunk.start+2)
						forSlice(chunk.data[chunk.start:chunk.end+1], func(v interface{}) {
							if sure(v) {
								result = append(result, v)
							}
						})
						//fmt.Println("get result...", result)
						out <- result[0:len(result)]
					} else {
						break
					}
				}
				return nil

			})

			//todo
			//need to modify the hardcode 10
			result := make([]interface{}, 0, 10)

			reduce := promise.Start(func() []interface{} {
				for {
					select {
					case <-f.GetChan():
						return nil
					case v, _ := <-out:
						//todo
						//need improve the append()
						result = append(result, (v.([]interface{}))...)
						//fmt.Println("result", result)
					}
				}
				return nil
			})

			if _, typ := reduce.Get(); typ != promise.RESULT_SUCCESS {
				//todo
				return nil
			} else {
				return blockSource{result, s.numberOfBlock, ceilSplitSize(len(result), s.numberOfBlock)}
			}
		}
		return nil
	})

}

//util funcs------------------------------------------
func makeTasks(src source, task func() []interface{}) *promise.Future {
	//itr := src.Itr()
	degree := src.NumberOfBlock()
	fs := make([]*promise.Future, degree, degree)
	for i := 0; i < degree; i++ {
		f := promise.Start(task)
		fs[i] = f
	}
	f := promise.WhenAll(fs...)
	return f
}

//func forChunk(c *chunk, f func(interface{}, out (*[]interface{}))) []interface{}{
//	result := make([]interface{}, 0, chunk.end-chunk.start+2)
//	forSlice(chunk.data[chunk.start:chunk.end+1], f)
//	return result[0:len(result)]

//}

func forSlice(src []interface{}, f func(interface{})) {
	for _, v := range src {
		f(v)
	}
}

func expandSlice(src []interface{}) []interface{} {
	count := 0
	for _, sub := range src {
		count += len(sub.([]interface{}))
	}

	fmt.Println("count", count)
	result := make([]interface{}, count, count)
	start := 0
	for _, sub := range src {
		size := len(sub.([]interface{}))
		copy(result[start:start+size], sub.([]interface{}))
		start += size
	}
	return result
}

func ceilSplitSize(a int, b int) int {
	if a%b != 0 {
		return a + (b - (a % b))
	} else {
		return a / b
	}
}
