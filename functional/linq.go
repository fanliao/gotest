package main

import (
	"fmt"
	//"time"
	"github.com/fanliao/go-promise"
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
	return func() (*chunk, bool) {
		if i < this.numberOfBlock {
			end := (i+1)*this.sizeOfBlock - 1
			if end >= len {
				end = len
			}
			return &chunk{this.data, i * this.numberOfBlock, end}, true
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
	return func() (*chunk, bool) {
		c, ok := <-this.data
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
				fmt.Println("add", i)
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
			itr := s.Itr()
			fs := make([]*promise.Future, s.numberOfBlock, s.numberOfBlock)
			out := make(chan interface{})

			for i := 0; i < s.numberOfBlock; i++ {
				fs[i] = promise.Start(func() []interface{} {
					for {
						if chunk, ok := itr(); ok {
							if chunk.end < chunk.start {
								s.Close()
								break
							}
							result := make([]interface{}, 0, chunk.end-chunk.start+2)
							forSlice(chunk.data[chunk.start:chunk.end+1], func(v interface{}) {
								if sure(v) {
									result = append(result, v)
								}
							})
							out <- result[0:len(result)]
						} else {
							break
						}
					}
					return nil
				})
			}

			f := promise.WhenAll(fs...)
			//todo
			//need to modify the hardcode 10
			result := make([]interface{}, 0, 10)

			reduce := promise.Start(func() []interface{} {
				select {
				case <-f.GetChan():
					break
				case v, _ := <-out:
					//todo
					//need improve the append()
					result = append(result, (v.([]interface{}))...)
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

	result := make([]interface{}, 0, count)
	start := 0
	for _, sub := range src {
		size := len(sub.([]interface{}))
		copy(result[start:start+size], sub.([]interface{}))
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
