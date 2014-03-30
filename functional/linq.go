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
		fmt.Println("begin from chan", this.data)
		c, ok := <-ch
		fmt.Println("from chan", c.start, c.end, ok)
		return c, ok
	}
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
			fs := make([]*promise.Future, s.numberOfBlock, s.numberOfBlock)
			for i := 0; i < s.numberOfBlock; i++ {
				chunk, _ := itr()
				f := promise.Start(func() []interface{} {
					result := make([]interface{}, 0, chunk.end-chunk.start+2)
					forSlice(chunk.data[chunk.start:chunk.end+1], func(v interface{}) {
						if sure(v) {
							result = append(result, v)
						}
					})
					result = append(result, true)
					return result
				})
				fs[i] = f
			}
			f := promise.WhenAll(fs...)
			if results, typ := f.Get(); typ != promise.RESULT_SUCCESS {
				//todo
				return nil
			} else {
				result := expandSlice(results)
				return blockSource{result, s.numberOfBlock, ceilSplitSize(len(result), s.numberOfBlock)}
			}
		case chunkSource:
			//itr := s.Itr()
			fs := make([]*promise.Future, s.numberOfBlock, s.numberOfBlock)
			out := make(chan interface{})

			for i := 0; i < s.numberOfBlock; i++ {
				j := i
				fs[j] = promise.Start(func() []interface{} {
					//fmt.Println("start", j)
					//for {
					//if chunk, ok := itr(); ok {
					for chunk := range s.data {
						//fmt.Println("get chunk...", chunk, reflect.ValueOf(chunk).IsNil())
						if reflect.ValueOf(chunk).IsNil() {
							//fmt.Println("close with nil", j)
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
					}
					//} else {
					//break
					//}
					//}
					//fmt.Println("return ", j)
					return nil
				})
			}

			f := promise.WhenAll(fs...)
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
