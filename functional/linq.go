package main

import (
	"fmt"
	//"time"
	"errors"
	"github.com/fanliao/go-promise"
	"reflect"
	"runtime"
	"sort"
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
	SOURCE_MAP
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
	chunks := make([]*chunk, 0, 10)
	for c := range this.data {
		chunks = append(chunks, c)
	}

	count := 0
	for _, c := range chunks {
		count = count + len(c.data)
	}

	result := make([]interface{}, 0, count)
	start := 0
	for _, c := range chunks {
		copy(result[start:start+len(c.data)], c.data)
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

type keyValue struct {
	key   interface{}
	value interface{}
}

type MapSource struct {
	data map[interface{}]interface{}
}

func (this MapSource) Typ() int {
	return SOURCE_MAP
}

func (this MapSource) ToSlice() []interface{} {
	i := 0
	results := make([]interface{}, len(this.data), len(this.data))
	for k, v := range this.data {
		results[i] = &keyValue{k, v}
		i++
	}
	return results
}

func (this MapSource) ToChan() chan interface{} {
	out := make(chan interface{})
	go func() {
		for k, v := range this.data {
			out <- &keyValue{k, v}
		}
		close(out)
	}()
	return out
}

//the queryable struct-------------------------------------------------------------------------
type Queryable struct {
	data      source
	steps     []step
	keepOrder bool
}

func From(src interface{}) (q Queryable) {
	q = Queryable{}
	q.keepOrder = true
	q.steps = make([]step, 0, 4)

	if s, ok := src.([]interface{}); ok {
		q.data = &blockSource{data: s}
	} else if s, ok := src.(chan *chunk); ok {
		q.data = &chunkSource{data: s}
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

func (this Queryable) get() source {
	data := this.data
	for _, step := range this.steps {
		data, this.keepOrder, _ = step.stepAction()(data, this.keepOrder)
	}
	return data
}

func (this Queryable) Results() []interface{} {
	return this.get().ToSlice()
}

func (this Queryable) Where(sure func(interface{}) bool) Queryable {
	this.steps = append(this.steps, commonStep{ACT_WHERE, sure, numCPU})
	return this
}

func (this Queryable) Select(selectFunc func(interface{}) interface{}) Queryable {
	this.steps = append(this.steps, commonStep{ACT_SELECT, selectFunc, numCPU})
	return this
}

func (this Queryable) Distinct(distinctFunc func(interface{}) interface{}) Queryable {
	this.steps = append(this.steps, commonStep{ACT_DISTINCT, distinctFunc, numCPU})
	return this
}

func (this Queryable) Order(compare func(interface{}, interface{}) int) Queryable {
	this.steps = append(this.steps, commonStep{ACT_ORDERBY, compare, numCPU})
	return this
}

func (this Queryable) GroupBy(keySelector func(interface{}) interface{}) Queryable {
	this.steps = append(this.steps, commonStep{ACT_GROUPBY, keySelector, numCPU})
	return this
}

func (this Queryable) Join(inner interface{},
	outerKeySelector func(interface{}) interface{},
	innerKeySelector func(interface{}) interface{},
	resultSelector func(interface{}, interface{}) interface{}) Queryable {
	this.steps = append(this.steps, joinStep{commonStep{ACT_JOIN, inner, numCPU}, outerKeySelector, innerKeySelector, resultSelector})
	return this
}

func (this Queryable) KeepOrder(keep bool) Queryable {
	this.keepOrder = keep
	return this
}

//the struct and functions of step-------------------------------------------------------------------------
type stepAction func(source, bool) (source, bool, error)
type step interface {
	stepAction() stepAction
}

type commonStep struct {
	typ    int
	act    interface{}
	degree int
}

type joinStep struct {
	commonStep
	outerKeySelector func(interface{}) interface{}
	innerKeySelector func(interface{}) interface{}
	resultSelector   func(interface{}, interface{}) interface{}
}

const (
	ACT_SELECT int = iota
	ACT_WHERE
	ACT_GROUPBY
	ACT_ORDERBY
	ACT_DISTINCT
	ACT_JOIN
)

func (this commonStep) stepAction() (act stepAction) {
	switch this.typ {
	case ACT_SELECT:
		act = getSelect(this.act.(func(interface{}) interface{}), this.degree)
	case ACT_WHERE:
		act = getWhere(this.act.(func(interface{}) bool), this.degree)
	case ACT_DISTINCT:
		act = getDistinct(this.act.(func(interface{}) interface{}), this.degree)
	case ACT_ORDERBY:
		act = getOrder(this.act.(func(interface{}, interface{}) int))
	case ACT_GROUPBY:
		act = getGroupBy(this.act.(func(interface{}) interface{}), this.degree)
	}
	return
}

func (this joinStep) stepAction() (act stepAction) {
	act = getJoin(this.act, this.outerKeySelector, this.innerKeySelector, this.resultSelector, this.degree)
	return
}

func getWhere(sure func(interface{}) bool, degree int) stepAction {
	return stepAction(func(src source, keepOrder bool) (dst source, keep bool, e error) {
		var f *promise.Future

		switch s := src.(type) {
		case *blockSource:
			f = makeBlockTasks(s, func(c *chunk) *chunk {
				result := mapChunk(c, whereAction(sure))
				//fmt.Println("src=", c, "result=", result)
				return result
			}, degree)
		case *chunkSource:
			out := make(chan *chunk)

			f1 := makeChanTasks(s, func(c *chunk) {
				out <- mapChunk(c, whereAction(sure))
			}, degree)

			f = makeReduceTask(f1.GetChan(), out, func(v interface{}, result *[]interface{}) {
				*result = append(*result, v)
			})
		}

		dst, e = sourceFromFuture(f, func(results []interface{}) source {
			result := expandChunks(results, keepOrder)
			return &blockSource{result}
		})
		keep = keepOrder
		return
	})
}

func getSelect(selectFunc func(interface{}) interface{}, degree int) stepAction {
	return stepAction(func(src source, keepOrder bool) (dst source, keep bool, e error) {
		var f *promise.Future
		keep = keepOrder

		switch s := src.(type) {
		case *blockSource:
			results := make([]interface{}, len(s.data), len(s.data))
			f = makeBlockTasks(s, func(c *chunk) *chunk {
				out := results[c.order : c.order+len(c.data)]
				mapSlice2(c.data, selectAction(selectFunc), &out)
				return nil
			}, degree)
			dst, e = sourceFromFuture(f, func(r []interface{}) source {
				//fmt.Println("results=", results)
				return &blockSource{results}
			})
			return
		case *chunkSource:
			out := make(chan *chunk)

			_ = makeChanTasks(s, func(c *chunk) {
				result := make([]interface{}, 0, len(c.data)) //c.end-c.start+2)
				mapSlice2(c.data, selectAction(selectFunc), &result)
				out <- &chunk{result, c.order}
			}, degree)

			//todo: how to handle error in promise?
			dst, e = &chunkSource{out}, nil
			return
		}

		panic(ErrUnsupportSource)
	})

}

func getOrder(compare func(interface{}, interface{}) int) stepAction {
	return stepAction(func(src source, keepOrder bool) (dst source, keep bool, e error) {
		switch s := src.(type) {
		case *blockSource:
			sorteds := sortSlice(s.data, func(this, that interface{}) bool {
				return compare(this, that) == -1
			})
			//sortable := sortable{}
			//sortable.less = func(this, that interface{}) bool {
			//	return compare(this, that) == -1
			//}
			//sortable.values = make([]interface{}, len(s.data))
			//_ = copy(sortable.values, s.data)
			//sort.Sort(sortable)
			return &blockSource{sorteds}, true, nil
		case *chunkSource:
			avl := NewAvlTree(compare)
			f := makeChanTasks(s, func(c *chunk) {
				for _, v := range c.data {
					avl.Insert(v)
				}
			}, 1)

			dst, e = sourceFromFuture(f, func(r []interface{}) source {
				return &blockSource{avl.ToSlice()}
			})
			keep = true
			return
		}
		panic(ErrUnsupportSource)
	})
}

func getDistinct(distinctFunc func(interface{}) interface{}, degree int) stepAction {
	return stepAction(func(src source, keepOrder bool) (source, bool, error) {
		out := make(chan *chunk)

		//get all values and keys
		var f *promise.Future
		switch s := src.(type) {
		case *blockSource:
			f = makeBlockTasks(s, func(c *chunk) (r *chunk) {
				out <- &chunk{getKeyValues(c, distinctFunc, nil), c.order}
				return
			}, degree)
		case *chunkSource:
			f = makeChanTasks(s, func(c *chunk) {
				out <- &chunk{getKeyValues(c, distinctFunc, nil), c.order}
			}, degree)
		}

		//get distinct values
		distKvs := make(map[interface{}]int)
		chunks := make([]interface{}, 0, degree)
	L1:
		for {
			select {
			case <-f.GetChan():
				break L1
			case c := <-out:
				chunks = append(chunks, c)
				result := make([]interface{}, 0, len(c.data))
				for _, v := range c.data {
					kv := v.(*keyValue)
					if _, ok := distKvs[kv.key]; !ok {
						distKvs[kv.key] = 1
						result = append(result, kv.value)
					}
				}
				c.data = result
			}
		}

		//get distinct values
		result := expandChunks(chunks, keepOrder)
		return &blockSource{result}, keepOrder, nil
		//i := 0
		//results := make([]interface{}, len(distKvs), len(distKvs))
		//for _, v := range distKvs {
		//	results[i] = v
		//	i++
		//}
		//return &blockSource{results}, nil
	})
}

//note the groupby cannot keep order because the map cannot keep order
func getGroupBy(groupFunc func(interface{}) interface{}, degree int) stepAction {
	return stepAction(func(src source, keepOrder bool) (source, bool, error) {
		out := make(chan *chunk)

		//get all values and keys
		var f *promise.Future
		switch s := src.(type) {
		case *blockSource:
			f = makeBlockTasks(s, func(c *chunk) (r *chunk) {
				out <- &chunk{getKeyValues(c, groupFunc, nil), c.order}
				return
			}, degree)
		case *chunkSource:
			f = makeChanTasks(s, func(c *chunk) {
				out <- &chunk{getKeyValues(c, groupFunc, nil), c.order}
			}, degree)
		}

		//get key with group values values
		groupKvs := make(map[interface{}]interface{})
	L1:
		for {
			select {
			case <-f.GetChan():
				break L1
			case c := <-out:
				for _, v := range c.data {
					kv := v.(*keyValue)
					if v, ok := groupKvs[kv.key]; !ok {
						groupKvs[kv.key] = []interface{}{kv.value}
					} else {
						list := v.([]interface{})
						groupKvs[kv.key] = append(list, kv.value)
					}
				}
			}
		}

		return &MapSource{groupKvs}, keepOrder, nil
	})
}

func getJoin(inner interface{},
	outerKeySelector func(interface{}) interface{},
	innerKeySelector func(interface{}) interface{},
	resultSelector func(interface{}, interface{}) interface{}, degree int) stepAction {
	return stepAction(func(src source, keepOrder bool) (dst source, keep bool, e error) {
		innerKVtask := promise.Start(func() []interface{} {
			innerKvs := From(inner).GroupBy(innerKeySelector).get().(*MapSource).data
			return []interface{}{innerKvs, true}
		})
		switch s := src.(type) {
		case *blockSource:
			outerKeySelectorFuture := makeBlockTasks(s, func(c *chunk) (r *chunk) {
				outerKvs := getKeyValues(c, outerKeySelector, nil)
				results := make([]interface{}, 0, 10)

				if r, ok := innerKVtask.Get(); ok != promise.RESULT_SUCCESS {
					//todo:

				} else {
					innerKvs := r[0].(map[interface{}]interface{})

					for _, o := range outerKvs {
						outerkv := o.(*keyValue)
						if innerList, ok := innerKvs[outerkv.key]; ok {
							innerList1 := innerList.([]interface{})
							for _, iv := range innerList1 {
								results = append(results, resultSelector(outerkv.value, iv))
							}

						}
					}
				}

				//j := 0
				//if c.order == 0 {
				//	for i := 0; i < 10000000; i++ {
				//		j = j + i*i
				//	}
				//}
				//fmt.Println("return order", c.order, j)
				return &chunk{results, c.order}
			}, degree)
			dst, e = sourceFromFuture(outerKeySelectorFuture, func(results []interface{}) source {
				result := expandChunks(results, keepOrder)
				return &blockSource{result}
			})
			keep = keepOrder
			return
		case *chunkSource:
			_ = makeChanTasks(s, func(c *chunk) {
				_ = getKeyValues(c, outerKeySelector, nil)
			}, degree)

		}

		return nil, keep, nil
	})
}

func getKeyValues(c *chunk, keyFunc func(v interface{}) interface{}, keyValues *[]interface{}) []interface{} {
	if keyValues == nil {
		list := (make([]interface{}, len(c.data), len(c.data)))
		keyValues = &list
	}
	mapSlice2(c.data, selectAction(func(v interface{}) interface{} {
		return &keyValue{keyFunc(v), v}
	}), keyValues)
	return *keyValues
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
		if i >= len(*out) {
			fmt.Println("out of", i, len(*out))
		}
		(*out)[i] = s(v)
	}
}

//util funcs------------------------------------------
func makeChanTasks(src *chunkSource, task func(*chunk), degree int) *promise.Future {
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
					task(c)
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

func makeBlockTasks(src source, task func(*chunk) *chunk, degree int) *promise.Future {
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
			return []interface{}{r, true}
		})
		fs[i] = f
		j++
	}
	f := promise.WhenAll(fs[0:j]...)

	return f
}

func makeReduceTask(chEndFlag chan *promise.PromiseResult, out chan *chunk,
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

func mapChunk(c *chunk, f func(interface{}, *[]interface{})) *chunk {
	result := make([]interface{}, 0, len(c.data)+1) //c.end-c.start+2)
	mapSlice(c.data, f, &result)
	//fmt.Println("c=", c)
	//fmt.Println("result=", result)
	return &chunk{result[0:len(result)], c.order}
}

func mapSlice(src []interface{}, f func(interface{}, *[]interface{}), out *[]interface{}) {
	for _, v := range src {
		f(v, out)
	}
}

func mapSlice2(src []interface{}, f func(interface{}, *[]interface{}, int), out *[]interface{}) {
	for i, v := range src {
		f(v, out, i)
	}
}

func expandChunks(src []interface{}, keepOrder bool) []interface{} {
	if src == nil {
		return nil
	}

	if keepOrder {
		src = sortSlice(src, func(a interface{}, b interface{}) bool {
			var (
				a1, b1 *chunk
			)
			switch v := a.(type) {
			case []interface{}:
				a1, b1 = v[0].(*chunk), b.([]interface{})[0].(*chunk)
			case *chunk:
				a1, b1 = v, b.(*chunk)
			}
			//a1, b1 := a.([]interface{})[0].(*chunk), b.([]interface{})[0].(*chunk)
			return a1.order < b1.order
		})
	}

	chunks := make([]*chunk, len(src), len(src))
	for i, c := range src {
		switch v := c.(type) {
		case []interface{}:
			chunks[i] = v[0].(*chunk)
		case *chunk:
			chunks[i] = v
		}
		//fmt.Println("keepOrder", keepOrder, chunks[i].order)
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
		return a/b + 1
	} else {
		return a / b
	}
}

//sort util func-------------------------------------------------------------------------------------------
type sortable struct {
	values []interface{}
	less   func(this, that interface{}) bool
}

func (q sortable) Len() int           { return len(q.values) }
func (q sortable) Swap(i, j int)      { q.values[i], q.values[j] = q.values[j], q.values[i] }
func (q sortable) Less(i, j int) bool { return q.less(q.values[i], q.values[j]) }

func sortSlice(data []interface{}, less func(interface{}, interface{}) bool) []interface{} {
	sortable := sortable{}
	sortable.less = less
	sortable.values = make([]interface{}, len(data))
	_ = copy(sortable.values, data)
	sort.Sort(sortable)
	return sortable.values

}

//AVL----------------------------------------------------
type avlNode struct {
	data           interface{}
	sameList       []interface{}
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
	//fmt.Println("rBalance, r=", *r)
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
		node := avlNode{e, nil, EH, nil, nil}
		*root = &node
		*taller = true
		//fmt.Println("insert to node,node=", *root)
	} else {
		i := compare1(e, (*root).data)
		if e == (*root).data || i == 0 {
			if (*root).sameList == nil {
				(*root).sameList = make([]interface{}, 0, 4)
			}

			(*root).sameList = append((*root).sameList, e)
			return false
		}

		if i == -1 {
			//lchild := (avlTree)((*root).lchild)
			//fmt.Println("will insert to lchild,lchild=", ((*root).lchild), " ,root=", *root, " ,e=", e)
			if !InsertAVL(&((*root).lchild), e, taller, compare1) {
				return false
			}
			//fmt.Println("insert to lchild,lchild=", ((*root).lchild), " ,root=", *root, " ,e=", e)
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
		} else if i == 1 {
			//rchild := (avlTree)((*root).rchild)
			//fmt.Println("will insert to rchild,rchild=", ((*root).rchild), " ,root=", *root, " ,e=", e)
			if !InsertAVL(&((*root).rchild), e, taller, compare1) {
				return false
			}
			//fmt.Println("insert to rchild,rchild=", ((*root).lchild), " ,root=", *root, " ,e=", e)
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

//func compareOriType(a interface{}, b interface{}) int {
//	if a < b {
//		return -1
//	} else if a == b {
//		return 0
//	} else {
//		return 1
//	}
//}

type avlTree struct {
	root    *avlNode
	count   int
	compare func(a interface{}, b interface{}) int
}

func (this *avlTree) Insert(node interface{}) {
	var taller bool
	InsertAVL(&(this.root), node, &taller, this.compare)
	this.count++

}

func (this *avlTree) ToSlice() []interface{} {
	result := (make([]interface{}, 0, this.count))
	avlToSlice(this.root, &result)
	return result
}

func NewAvlTree(compare func(a interface{}, b interface{}) int) *avlTree {
	return &avlTree{nil, 0, compare}
}

func avlToSlice(root *avlNode, result *[]interface{}) []interface{} {
	if result == nil {
		r := make([]interface{}, 0, 10)
		result = &r
	}

	if root == nil {
		return *result
	}

	if (root).lchild != nil {
		l := root.lchild
		avlToSlice(l, result)
	}
	*result = append(*result, root.data)
	if root.sameList != nil {
		for _, v := range root.sameList {
			*result = append(*result, v)
		}
	}
	if (root).rchild != nil {
		r := (root.rchild)
		avlToSlice(r, result)
	}
	return *result
}
