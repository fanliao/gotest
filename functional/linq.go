package main

import (
	"fmt"
	//"time"
	"errors"
	"github.com/fanliao/go-promise"
	"reflect"
	"runtime"
)

var (
	numCPU             int
	ErrUnsupportSource = errors.New("unsupport source")
)

func init() {
	numCPU = runtime.NumCPU()
	fmt.Println("go linq")
}

// the struct and interface about data source---------------------------------------------------
type chunk struct {
	data  []interface{}
	order int
	//start int
	//end   int
}

const (
	SOURCE_BLOCK int = iota
	SOURCE_CHUNK
)

const (
	ACT_SELECT int = iota
	ACT_WHERE
	ACT_GROUPBY
	ACT_ORDERBY
)

type source interface {
	Typ() int           //block or chunk?
	NumberOfBlock() int //get the degree of parallel
	SetNumberOfBlock(int)
	ToSlice() []interface{}
	ToChan() chan interface{}
}

type blockSource struct {
	data          []interface{}
	numberOfBlock int //the count of block
}

func (this blockSource) Typ() int {
	return SOURCE_BLOCK
}

func (this *blockSource) SetNumberOfBlock(n int) {
	this.numberOfBlock = n
}

func (this blockSource) NumberOfBlock() int {
	return this.numberOfBlock
}

func (this blockSource) ToSlice() []interface{} {
	return this.data
}

func (this blockSource) ToChan() chan interface{} {
	out := make(chan interface{})
	go func() {
		for _, v := range this.data {
			out <- v
		}
		close(out)
	}()
	return out
}

type chunkSource struct {
	data          chan *chunk
	numberOfBlock int
}

func (this chunkSource) Typ() int {
	return SOURCE_CHUNK
}

func (this chunkSource) SetNumberOfBlock(n int) {
	this.numberOfBlock = n
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

func (this chunkSource) ToSlice() []interface{} {
	return nil
}

func (this chunkSource) ToChan() chan interface{} {
	return nil
}

type stepAction func(source) source

//the queryable struct-------------------------------------------------------------------------
type Queryable struct {
	data  source
	steps []step
}

func From(src interface{}) (q Queryable) {
	q = Queryable{}
	q.steps = make([]step, 0, 4)

	if s, ok := src.([]interface{}); ok {
		q.data = &blockSource{data: s}
	} else {
		typ := reflect.TypeOf(src)
		switch typ.Kind() {
		case reflect.Slice:

		case reflect.Chan:
		case reflect.Map:
		default:
		}
		panic(ErrUnsupportSource)
	}
	return
}

func (this Queryable) Results() []interface{} {
	data := this.data
	for _, step := range this.steps {
		switch step.typ {
		case ACT_SELECT:
			data.SetNumberOfBlock(numCPU)
			whereAct := getSelect(step.act.(func(interface{}) interface{}))
			data = whereAct(data)
		case ACT_WHERE:
			data.SetNumberOfBlock(numCPU)
			whereAct := where(step.act.(func(interface{}) bool))
			data = whereAct(data)

		}
	}
	return data.ToSlice()
}

func (this Queryable) Where(sure func(interface{}) bool) Queryable {
	this.steps = append(this.steps, step{ACT_WHERE, sure, 0})
	return this
}

func (this Queryable) Select(selectFunc func(interface{}) interface{}) Queryable {
	this.steps = append(this.steps, step{ACT_SELECT, selectFunc, 0})
	return this
}

//the function of step-------------------------------------------------------------------------
type step struct {
	typ    int
	act    interface{}
	degree int
}

func where(sure func(interface{}) bool) stepAction {
	return stepAction(func(src source) source {
		var f *promise.Future

		switch s := src.(type) {
		case *blockSource:
			f = makeBlockTasks(s, func(c *chunk) []interface{} {
				result := forChunk(c, whereAction(sure))
				//fmt.Println("src=", c, "result=", result)
				return []interface{}{result, true}
			})
		case *chunkSource:
			out := make(chan *chunk)

			f1 := makeTasks(s, func(itr func() (*chunk, bool)) []interface{} {
				for {
					if c, ok := itr(); ok {
						if reflect.ValueOf(c).IsNil() {
							s.Close()
							break
						}
						out <- forChunk(c, whereAction(sure))
					} else {
						break
					}
				}
				return nil
			})

			f = makeSummaryTask(f1.GetChan(), out, func(v interface{}, result *[]interface{}) {
				*result = append(*result, (v.(*chunk).data)...)
			})
		}
		if results, typ := f.Get(); typ != promise.RESULT_SUCCESS {
			//todo
			return nil
		} else {
			result := expandSlice(results)
			return &blockSource{result, src.NumberOfBlock()}
		}
	})

}

func getSelect(selectFunc func(interface{}) interface{}) stepAction {
	return stepAction(func(src source) source {
		var f *promise.Future

		switch s := src.(type) {
		case *blockSource:
			results := make([]interface{}, len(s.data), len(s.data))
			f = makeBlockTasks(s, func(c *chunk) []interface{} {
				//result := forChunk(c, selectAction(selectFunc))
				out := results[c.order : c.order+len(c.data)]
				//fmt.Println("(c)=", (c), c.order, c.order+len(c.data), "out=", out)
				forSlice2(c.data, selectAction(selectFunc), &out)
				//fmt.Println("(out)=", (out))
				//copy(results[c.order:c.order+len(c.data)], result.data)
				return nil
			})
			if _, typ := f.Get(); typ != promise.RESULT_SUCCESS {
				//todo
				return nil
			} else {
				//fmt.Println("(results)=", (results))
				return &blockSource{results, src.NumberOfBlock()}
			}
		case *chunkSource:
			out := make(chan *chunk)

			_ = makeTasks(s, func(itr func() (*chunk, bool)) []interface{} {
				for {
					if c, ok := itr(); ok {
						if reflect.ValueOf(c).IsNil() {
							s.Close()
							break
						}
						result := make([]interface{}, 0, len(c.data)) //c.end-c.start+2)
						forSlice2(c.data, selectAction(selectFunc), &result)
						//fmt.Println("c=", c)
						//fmt.Println("result=", result)
						out <- &chunk{result, c.order}
					} else {
						break
					}
				}
				return nil
			})

			//todo: how to handle error in promise?
			return &chunkSource{out, src.NumberOfBlock()}
		}

		panic(ErrUnsupportSource)
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

func selectAction(s func(interface{}) interface{}) func(v interface{}, out *[]interface{}, i int) {
	return func(v interface{}, out *[]interface{}, i int) {
		(*out)[i] = s(v)
	}
}

//util funcs------------------------------------------
func makeTasks(src *chunkSource, task func(func() (*chunk, bool)) []interface{}) *promise.Future {
	itr := src.Itr()
	degree := src.NumberOfBlock()
	fs := make([]*promise.Future, degree, degree)
	for i := 0; i < degree; i++ {
		f := promise.Start(func() []interface{} {
			r := task(itr)
			//fmt.Println("r=", r)
			return r
		})
		fs[i] = f
	}
	f := promise.WhenAll(fs...)

	return f
}

func makeBlockTasks(src source, task func(*chunk) []interface{}) *promise.Future {
	degree := src.NumberOfBlock()

	fs := make([]*promise.Future, degree, degree)
	data := src.(*blockSource).data
	len := len(data)
	size := ceilSplitSize(len, src.NumberOfBlock())
	j := 0
	for i := 0; i < degree && i*size < len; i++ {
		end := (i + 1) * size
		if end >= len {
			end = len
		}
		c := &chunk{data[i*size : end], i * size} //, end}
		f := promise.Start(func() []interface{} {
			r := task(c)
			return r
		})
		fs[i] = f
		j++
	}
	f := promise.WhenAll(fs[0:j]...)

	return f
}

func makeSummaryTask(chEndFlag chan *promise.PromiseResult, out chan *chunk,
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

func forChunk(c *chunk, f func(interface{}, *[]interface{})) *chunk {
	result := make([]interface{}, 0, len(c.data)+1) //c.end-c.start+2)
	forSlice(c.data, f, &result)
	//fmt.Println("c=", c)
	//fmt.Println("result=", result)
	return &chunk{result[0:len(result)], c.order}
}

func forSlice(src []interface{}, f func(interface{}, *[]interface{}), out *[]interface{}) {
	for _, v := range src {
		f(v, out)
	}
}

func forSlice2(src []interface{}, f func(interface{}, *[]interface{}, int), out *[]interface{}) {
	for i, v := range src {
		f(v, out, i)
	}
}

func expandSlice(src []interface{}) []interface{} {
	if src == nil {
		return nil
	}

	chunks := make([]*chunk, len(src), len(src))
	for i, c := range src {
		chunks[i] = c.([]interface{})[0].(*chunk)
		//fmt.Println("chunks[i]", i, "=", chunks[i])
	}

	count := 0
	for _, c := range chunks {
		count += len(c.data)
	}

	//fmt.Println("count", count)
	result := make([]interface{}, count, count)
	start := 0
	for _, c := range chunks {
		size := len(c.data)
		copy(result[start:start+size], c.data)
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
