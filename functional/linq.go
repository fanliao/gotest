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
	Typ() int //block or chunk?
	ToSlice() []interface{}
	ToChan() chan interface{}
}

type blockSource struct {
	data []interface{}
}

func (this blockSource) Typ() int {
	return SOURCE_BLOCK
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
	data chan *chunk
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

func (this chunkSource) Close() {
	close(this.data)
}

func (this chunkSource) ToSlice() []interface{} {
	result := make([]interface{}, 0, 10)
	for c := range this.data {
		for _, v := range c.data {
			result = append(result, v)
		}
	}
	return result
}

func (this chunkSource) ToChan() chan interface{} {
	out := make(chan interface{})
	go func() {
		for c := range this.data {
			for _, v := range c.data {
				out <- v
			}
		}
	}()
	return out

}

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
		data, _ = step.stepAction()(data)
	}
	return data.ToSlice()
}

func (this Queryable) Where(sure func(interface{}) bool) Queryable {
	this.steps = append(this.steps, step{ACT_WHERE, sure, numCPU})
	return this
}

func (this Queryable) Select(selectFunc func(interface{}) interface{}) Queryable {
	this.steps = append(this.steps, step{ACT_SELECT, selectFunc, numCPU})
	return this
}

//the struct and functions of step-------------------------------------------------------------------------
type step struct {
	typ    int
	act    interface{}
	degree int
}

type stepAction func(source) (source, error)

func (this step) stepAction() (act stepAction) {
	switch this.typ {
	case ACT_SELECT:
		act = getSelect(this.act.(func(interface{}) interface{}), this.degree)
	case ACT_WHERE:
		act = getWhere(this.act.(func(interface{}) bool), this.degree)
	}
	return
}

func getWhere(sure func(interface{}) bool, degree int) stepAction {
	return stepAction(func(src source) (source, error) {
		var f *promise.Future

		switch s := src.(type) {
		case *blockSource:
			f = makeBlockTasks(s, func(c *chunk) []interface{} {
				result := forChunk(c, whereAction(sure))
				//fmt.Println("src=", c, "result=", result)
				return []interface{}{result, true}
			}, degree)
		case *chunkSource:
			out := make(chan *chunk)

			f1 := makeChanTasks(s, out, func(c *chunk) *chunk {
				return forChunk(c, whereAction(sure))
			}, degree)

			f = makeSummaryTask(f1.GetChan(), out, func(v interface{}, result *[]interface{}) {
				*result = append(*result, (v.(*chunk).data)...)
			})
		}

		return sourceFromFuture(f, func(results []interface{}) source {
			result := expandSlice(results)
			return &blockSource{result}
		})
	})
}

func getSelect(selectFunc func(interface{}) interface{}, degree int) stepAction {
	return stepAction(func(src source) (source, error) {
		var f *promise.Future

		switch s := src.(type) {
		case *blockSource:
			results := make([]interface{}, len(s.data), len(s.data))
			f = makeBlockTasks(s, func(c *chunk) []interface{} {
				out := results[c.order : c.order+len(c.data)]
				forSlice2(c.data, selectAction(selectFunc), &out)
				//fmt.Println("(out)=", (out))
				return nil
			}, degree)
			return sourceFromFuture(f, func(results []interface{}) source {
				return &blockSource{results}
			})
		case *chunkSource:
			out := make(chan *chunk)

			_ = makeChanTasks(s, out, func(c *chunk) *chunk {
				result := make([]interface{}, 0, len(c.data)) //c.end-c.start+2)
				forSlice2(c.data, selectAction(selectFunc), &result)
				return &chunk{result, c.order}
			}, degree)

			//todo: how to handle error in promise?
			return &chunkSource{out}, nil
		}

		panic(ErrUnsupportSource)
	})

}

func sourceFromFuture(f *promise.Future, sourceFunc func([]interface{}) source) (source, error) {
	if results, typ := f.Get(); typ != promise.RESULT_SUCCESS {
		//todo
		return nil, nil
	} else {
		//fmt.Println("(results)=", (results))
		return sourceFunc(results), nil
	}
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
func makeChanTasks(src *chunkSource, out chan *chunk, task func(*chunk) *chunk, degree int) *promise.Future {
	itr := src.Itr()
	fs := make([]*promise.Future, degree, degree)
	for i := 0; i < degree; i++ {
		f := promise.Start(func() []interface{} {
			for {
				if c, ok := itr(); ok {
					if reflect.ValueOf(c).IsNil() {
						src.Close()
						break
					}
					out <- task(c)
				} else {
					break
				}
			}
			return nil
			//fmt.Println("r=", r)
		})
		fs[i] = f
	}
	f := promise.WhenAll(fs...)

	return f
}

func makeChanTasks1(src *chunkSource, task func(func() (*chunk, bool)) []interface{}, degree int) *promise.Future {
	itr := src.Itr()
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

func makeBlockTasks(src source, task func(*chunk) []interface{}, degree int) *promise.Future {
	fs := make([]*promise.Future, degree, degree)
	data := src.(*blockSource).data
	len := len(data)
	size := ceilSplitSize(len, degree)
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

//AVL----------------------------------------------------
type avlNode struct {
	data           interface{}
	bf             int
	lchild, rchild *avlNode
}

func rRotate(node **avlNode) {
	l := (*node).lchild
	(*node).lchild = l.rchild
	l.rchild = *node
	*node = l
}

func lRotate(node **avlNode) {
	r := (*node).rchild
	(*node).rchild = r.lchild
	r.lchild = *node
	*node = r
}

const (
	LH int = 1
	EH     = 0
	RH     = -1
)

func lBalance(root **avlNode) {
	var lr *avlNode
	l := (*root).lchild
	switch l.bf {
	case LH:
		(*root).bf = EH
		l.bf = EH
		rRotate(root)
	case RH:
		lr = l.rchild
		switch lr.bf {
		case LH:
			(*root).bf = RH
			l.bf = EH
		case EH:
			(*root).bf = EH
			l.bf = EH
		case RH:
			(*root).bf = EH
			l.bf = LH
		}
		lr.bf = EH
		//pLchild := (avlTree)((*root).lchild)
		lRotate(&((*root).lchild))
		rRotate(root)
	}
}

func rBalance(root **avlNode) {
	var rl *avlNode
	r := (*root).rchild
	fmt.Println("rBalance, r=", *r)
	switch r.bf {
	case RH:
		(*root).bf = EH
		r.bf = EH
		lRotate(root)
	case LH:
		rl = r.lchild
		switch rl.bf {
		case LH:
			(*root).bf = RH
			r.bf = EH
		case EH:
			(*root).bf = EH
			r.bf = EH
		case RH:
			(*root).bf = EH
			r.bf = LH
		}
		rl.bf = EH
		//pRchild := (avlTree)((*root).rchild)
		rRotate(&((*root).rchild))
		lRotate(root)
	}
}

func InsertAVL(root **avlNode, e interface{}, taller *bool, compare1 func(interface{}, interface{}) int) bool {
	if *root == nil {
		node := avlNode{e, EH, nil, nil}
		*root = &node
		*taller = true
		fmt.Println("insert to node,node=", *root)
	} else {
		if e == (*root).data {
			return false
		}

		if compare1(e, (*root).data) == -1 {
			//lchild := (avlTree)((*root).lchild)
			fmt.Println("will insert to lchild,lchild=", ((*root).lchild), " ,root=", *root, " ,e=", e)
			if !InsertAVL(&((*root).lchild), e, taller, compare1) {
				return false
			}
			fmt.Println("insert to lchild,lchild=", ((*root).lchild), " ,root=", *root, " ,e=", e)
			if *taller {
				switch (*root).bf {
				case LH:
					lBalance(root)
					*taller = false
				case EH:
					(*root).bf = LH
					*taller = true
				case RH:
					(*root).bf = EH
					*taller = false
				}
			}
		} else {
			//rchild := (avlTree)((*root).rchild)
			fmt.Println("will insert to rchild,rchild=", ((*root).rchild), " ,root=", *root, " ,e=", e)
			if !InsertAVL(&((*root).rchild), e, taller, compare1) {
				return false
			}
			fmt.Println("insert to rchild,rchild=", ((*root).lchild), " ,root=", *root, " ,e=", e)
			if *taller {
				switch (*root).bf {
				case RH:
					rBalance(root)
					*taller = false
				case EH:
					(*root).bf = RH
					*taller = true
				case LH:
					(*root).bf = EH
					*taller = false
				}
			}
		}
	}
	return true
}

type avlTree struct {
	root    *avlNode
	count   int
	compare func(a interface{}, b interface{}) int
}

func (this *avlTree) Insert(node interface{}) {
	var taller bool
	if InsertAVL(&(this.root), node, &taller, this.compare) {
		this.count++
	}
}

func (this *avlTree) ToSlice() []interface{} {
	result := (make([]interface{}, 0, this.count))
	avlToSlice(this.root, &result)
	return result
}

func avlToSlice(root *avlNode, result *[]interface{}) []interface{} {
	if result == nil {
		r := make([]interface{}, 0, 10)
		result = &r
	}

	if (root).lchild != nil {
		l := root.lchild
		avlToSlice(l, result)
	}
	*result = append(*result, root.data)
	if (root).rchild != nil {
		r := (root.rchild)
		avlToSlice(r, result)
	}
	return *result
}
