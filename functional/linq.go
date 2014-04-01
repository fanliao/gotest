package main

import (
	//"fmt"
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
}

func (this blockSource) Typ() int {
	return SOURCE_BLOCK
}

func (this blockSource) Itr() func() (*chunk, bool) {
	i := 0
	len := len(this.data)
	return func() (c *chunk, ok bool) {
		if i < this.numberOfBlock {
			size := ceilSplitSize(len, this.numberOfBlock)
			end := (i+1)*size - 1
			if end >= len {
				end = len
			}
			c, ok = &chunk{this.data, i * size, end}, true
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
		var f *promise.Future

		switch s := src.(type) {
		case blockSource:
			f = makeBlockTasks(s, func(c *chunk) []interface{} {
				result := forChunk(c, whereAction(sure))
				return result
			})
		case chunkSource:
			out := make(chan interface{})

			f1 := makeTasks(s, func(itr func() (*chunk, bool)) []interface{} {
				for {
					if chunk, ok := itr(); ok {
						if reflect.ValueOf(chunk).IsNil() {
							s.Close()
							break
						}
						out <- forChunk(chunk, whereAction(sure))
					} else {
						break
					}
				}
				return nil
			})

			f = makeSummaryTask(f1.GetChan(), out, func(v interface{}, result *[]interface{}) {
				*result = append(*result, (v.([]interface{}))...)
			})
		}
		if results, typ := f.Get(); typ != promise.RESULT_SUCCESS {
			//todo
			return nil
		} else {
			result := expandSlice(results)
			return blockSource{result, src.NumberOfBlock()}
		}
	})

}

//actions---------------------------------------------
func whereAction(sure func(interface{}) bool) func(v interface{}, out *[]interface{}) {
	return func(v interface{}, out *[]interface{}) {
		if sure(v) {
			*out = append(*out, v)
		}
	}
}

//util funcs------------------------------------------
func makeTasks(src source, task func(func() (*chunk, bool)) []interface{}) *promise.Future {
	itr := src.Itr()
	degree := src.NumberOfBlock()
	fs := make([]*promise.Future, degree, degree)
	for i := 0; i < degree; i++ {
		f := promise.Start(func() []interface{} {
			return task(itr)
		})
		fs[i] = f
	}
	f := promise.WhenAll(fs...)

	return f
}

func makeBlockTasks(src source, task func(*chunk) []interface{}) *promise.Future {
	degree := src.NumberOfBlock()

	fs := make([]*promise.Future, degree, degree)
	data := src.(blockSource).data
	len := len(data)
	size := ceilSplitSize(len, src.NumberOfBlock())
	for i := 0; i < degree; i++ {
		end := (i+1)*size - 1
		if end >= len-1 {
			end = len - 1
		}
		//fmt.Println("block, i=", i, ", size=", size, "end=", end)
		c := &chunk{data, i * size, end}
		//fmt.Println("block, c=", c.start, c.end)
		f := promise.Start(func() []interface{} {
			return task(c)
		})
		fs[i] = f
	}
	f := promise.WhenAll(fs...)

	return f
}

func makeSummaryTask(chEndFlag chan *promise.PromiseResult, out chan interface{},
	summary func(interface{}, *[]interface{}),
) *promise.Future {
	f := promise.Start(func() []interface{} {
		//todo
		//need to modify the hardcode 10
		result := make([]interface{}, 0, 10)
		for {
			select {
			case <-chEndFlag:
				return result
			case v, _ := <-out:
				//todo
				//need improve the append()
				summary(v, &result)
			}
		}

		return result
	})

	return f
}

func forChunk(c *chunk, f func(interface{}, *[]interface{})) []interface{} {
	result := make([]interface{}, 0, c.end-c.start+2)
	forSlice(c.data[c.start:c.end+1], f, &result)
	//fmt.Println("c=", c)
	//fmt.Println("result=", result)
	return result[0:len(result)]

}

func forSlice(src []interface{}, f func(interface{}, *[]interface{}), out *[]interface{}) {
	for _, v := range src {
		f(v, out)
	}
}

func expandSlice(src []interface{}) []interface{} {
	if src == nil {
		return nil
	}
	if _, hasSubSlice := src[0].([]interface{}); !hasSubSlice {
		return src
	}

	count := 0
	for _, sub := range src {
		count += len(sub.([]interface{}))
	}

	//fmt.Println("count", count)
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
		return (a + (b - (a % b))) / b
	} else {
		return a / b
	}
}
