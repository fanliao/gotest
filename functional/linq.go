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
	ErrUnsupportSource = errors.New("unsupport dataSource")
)

func init() {
	numCPU = runtime.NumCPU()
	fmt.Println("go linq")
}

// the struct and interface about data dataSource---------------------------------------------------
type chunk struct {
	data  []interface{}
	order int
}

const (
	SOURCE_BLOCK int = iota
	SOURCE_CHUNK
)

type dataSource interface {
	Typ() int //block or chan?
	ToSlice(bool) []interface{}
	ToChan() chan interface{}
}

type listSource struct {
	data interface{}
}

func (this listSource) Typ() int {
	return SOURCE_BLOCK
}

func (this listSource) ToSlice(keepOrder bool) []interface{} {
	switch data := this.data.(type) {
	case []interface{}:
		return data
	case map[interface{}]interface{}:
		i := 0
		results := make([]interface{}, len(data), len(data))
		for k, v := range data {
			results[i] = &keyValue{k, v}
			i++
		}
		return results
	default:
		value := reflect.ValueOf(this.data)
		switch value.Kind() {
		case reflect.Slice:
			l := value.Len()
			results := make([]interface{}, l, l)
			for i := 0; i < l; i++ {
				results[i] = value.Index(i).Interface()
			}
			return results
		case reflect.Map:
			l := value.Len()
			results := make([]interface{}, l, l)
			for i, k := range value.MapKeys() {
				results[i] = &keyValue{k.Interface(), value.MapIndex(k).Interface()}
			}
			return results
		}
		return nil

	}
	return nil
}

func (this listSource) ToChan() chan interface{} {
	out := make(chan interface{})
	go func() {
		for _, v := range this.ToSlice(true) {
			out <- v
		}
		close(out)
	}()
	return out
}

type chanSource struct {
	data chan *chunk
}

func (this chanSource) Typ() int {
	return SOURCE_CHUNK
}

func (this chanSource) Itr() func() (*chunk, bool) {
	ch := this.data
	return func() (*chunk, bool) {
		c, ok := <-ch
		return c, ok
	}
}

func (this chanSource) Close() {
	close(this.data)
}

func (this chanSource) ToSlice(keepOrder bool) []interface{} {
	//chunks := make([]*chunk, 0, 4)
	//for c := range this.data {
	//	chunks = appendChunkSlice(chunks, c)
	//}

	//count := 0
	//for _, c := range chunks {
	//	count = count + len(c.data)
	//}

	//result := make([]interface{}, 0, count)
	//start := 0
	//for _, c := range chunks {
	//	copy(result[start:start+len(c.data)], c.data)
	//}
	//return result

	chunks := make([]interface{}, 0, 2)
	for c := range this.data {
		chunks = appendSlice(chunks, c)
	}
	return expandChunks(chunks, keepOrder)
}

func (this chanSource) ToChan() chan interface{} {
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

//the queryable struct-------------------------------------------------------------------------
type Queryable struct {
	data      dataSource
	steps     []step
	keepOrder bool
}

func From(src interface{}) (q Queryable) {
	q = Queryable{}
	q.keepOrder = true
	q.steps = make([]step, 0, 4)

	if k := reflect.ValueOf(src).Kind(); k == reflect.Slice || k == reflect.Map {
		q.data = &listSource{data: src}
	} else if s, ok := src.(chan *chunk); ok {
		q.data = &chanSource{data: s}
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

func (this Queryable) get() dataSource {
	data := this.data
	for _, step := range this.steps {
		data, this.keepOrder, _ = step.stepAction()(data, this.keepOrder)
	}
	return data
}

func (this Queryable) Results() []interface{} {
	return this.get().ToSlice(this.keepOrder)
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
type stepAction func(dataSource, bool) (dataSource, bool, error)
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
	act = getJoin(this.act, this.outerKeySelector, this.innerKeySelector, this.resultSelector, false, this.degree)
	return
}

func getWhere(sure func(interface{}) bool, degree int) stepAction {
	return stepAction(func(src dataSource, keepOrder bool) (dst dataSource, keep bool, e error) {
		var f *promise.Future
		mapChunk := func(c *chunk) *chunk {
			return filterChunk(c, sure)
			//fmt.Println("src=", c, "result=", result)
		}

		switch s := src.(type) {
		case *listSource:
			f = parallelMapList(s, mapChunk, degree)
		case *chanSource:
			reduceSrc := make(chan *chunk)

			f = parallelMapChan(s, reduceSrc, mapChunk, degree)

			results := make([]interface{}, 0, 1)
			if keepOrder {
				avl := newChunkAvlTree()
				reduceChan(f.GetChan(), reduceSrc, func(v *chunk) { avl.Insert(v) })
				results = avl.ToSlice()
				keepOrder = false
			} else {
				reduceChan(f.GetChan(), reduceSrc, func(v *chunk) { results = appendSlice(results, v) })
			}

			f = promise.Wrap(results)
		}

		dst, e = getSource(f, func(results []interface{}) dataSource {
			result := expandChunks(results, false)
			return &listSource{result}
		})
		keep = keepOrder
		return
	})
}

func getSelect(selectFunc func(interface{}) interface{}, degree int) stepAction {
	return stepAction(func(src dataSource, keepOrder bool) (dst dataSource, keep bool, e error) {
		var f *promise.Future
		keep = keepOrder

		switch s := src.(type) {
		case *listSource:
			l := len(s.ToSlice(false))
			results := make([]interface{}, l, l)
			f = parallelMapList(s, func(c *chunk) *chunk {
				out := results[c.order : c.order+len(c.data)]
				mapSlice(c.data, selectFunc, &out)
				return nil
			}, degree)
			dst, e = getSource(f, func(r []interface{}) dataSource {
				//fmt.Println("results=", results)
				return &listSource{results}
			})
			return
		case *chanSource:
			out := make(chan *chunk)

			_ = parallelMapChan(s, out, func(c *chunk) *chunk {
				result := make([]interface{}, 0, len(c.data)) //c.end-c.start+2)
				mapSlice(c.data, selectFunc, &result)
				return &chunk{result, c.order}
			}, degree)

			//todo: how to handle error in promise?
			dst, e = &chanSource{out}, nil
			return
		}

		panic(ErrUnsupportSource)
	})

}

func getOrder(compare func(interface{}, interface{}) int) stepAction {
	return stepAction(func(src dataSource, keepOrder bool) (dst dataSource, keep bool, e error) {
		switch s := src.(type) {
		case *listSource:
			sorteds := sortSlice(s.ToSlice(false), func(this, that interface{}) bool {
				return compare(this, that) == -1
			})
			return &listSource{sorteds}, true, nil
		case *chanSource:
			avl := NewAvlTree(compare)
			f := parallelMapChan(s, nil, func(c *chunk) *chunk {
				for _, v := range c.data {
					avl.Insert(v)
				}
				return nil
			}, 1)

			dst, e = getSource(f, func(r []interface{}) dataSource {
				return &listSource{avl.ToSlice()}
			})
			keep = true
			return
		}
		panic(ErrUnsupportSource)
	})
}

func getDistinct(distinctFunc func(interface{}) interface{}, degree int) stepAction {
	return stepAction(func(src dataSource, keepOrder bool) (dataSource, bool, error) {
		reduceSrc := make(chan *chunk)
		mapChunk := func(c *chunk) (r *chunk) {
			reduceSrc <- &chunk{getKeyValues(c, distinctFunc, nil), c.order}
			return
		}

		//get all values and keys
		var f *promise.Future
		switch s := src.(type) {
		case *listSource:
			f = parallelMapList(s, mapChunk, degree)
		case *chanSource:
			f = parallelMapChan(s, nil, mapChunk, degree)
		}

		//get distinct values
		distKvs := make(map[interface{}]int)
		chunks := make([]interface{}, 0, degree)
		reduceChan(f.GetChan(), reduceSrc, func(c *chunk) {
			chunks = appendSlice(chunks, c)
			result := make([]interface{}, 0, 2)
			for _, v := range c.data {
				kv := v.(*keyValue)
				if _, ok := distKvs[kv.key]; !ok {
					distKvs[kv.key] = 1
					result = appendSlice(result, kv.value)
				}
			}
			c.data = result
		})

		//get distinct values
		result := expandChunks(chunks, false)
		return &listSource{result}, keepOrder, nil
	})
}

//note the groupby cannot keep order because the map cannot keep order
func getGroupBy(groupFunc func(interface{}) interface{}, degree int) stepAction {
	return stepAction(func(src dataSource, keepOrder bool) (dataSource, bool, error) {
		reduceSrc := make(chan *chunk)
		mapChunk := func(c *chunk) (r *chunk) {
			reduceSrc <- &chunk{getKeyValues(c, groupFunc, nil), c.order}
			return
		}

		//get all values and keys
		var f *promise.Future
		switch s := src.(type) {
		case *listSource:
			f = parallelMapList(s, mapChunk, degree)
		case *chanSource:
			f = parallelMapChan(s, nil, mapChunk, degree)
		}

		//get key with group values values
		groupKvs := make(map[interface{}]interface{})
		reduceChan(f.GetChan(), reduceSrc, func(c *chunk) {
			for _, v := range c.data {
				kv := v.(*keyValue)
				if v, ok := groupKvs[kv.key]; !ok {
					groupKvs[kv.key] = []interface{}{kv.value}
				} else {
					list := v.([]interface{})
					groupKvs[kv.key] = appendSlice(list, kv.value)
				}
			}
		})

		return &listSource{groupKvs}, keepOrder, nil
	})
}

func getJoin(inner interface{},
	outerKeySelector func(interface{}) interface{},
	innerKeySelector func(interface{}) interface{},
	resultSelector func(interface{}, interface{}) interface{}, isLeftJoin bool, degree int) stepAction {
	return stepAction(func(src dataSource, keepOrder bool) (dst dataSource, keep bool, e error) {
		keep = keepOrder
		innerKVtask := promise.Start(func() []interface{} {
			innerKvs := From(inner).GroupBy(innerKeySelector).get().(*listSource).data
			return []interface{}{innerKvs, true}
		})

		mapChunk := func(c *chunk) (r *chunk) {
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
							results = appendSlice(results, resultSelector(outerkv.value, iv))
						}
					} else if isLeftJoin {
						results = appendSlice(results, resultSelector(outerkv.value, nil))
					}
				}
			}

			return &chunk{results, c.order}
		}

		switch s := src.(type) {
		case *listSource:
			outerKeySelectorFuture := parallelMapList(s, mapChunk, degree)
			dst, e = getSource(outerKeySelectorFuture, func(results []interface{}) dataSource {
				result := expandChunks(results, false)
				return &listSource{result}
			})
			return
		case *chanSource:
			out := make(chan *chunk)
			_ = parallelMapChan(s, nil, mapChunk, degree)
			dst, e = &chanSource{out}, nil
			return
		}

		return nil, keep, nil
	})
}

func getGroupJoin(inner interface{},
	outerKeySelector func(interface{}) interface{},
	innerKeySelector func(interface{}) interface{},
	resultSelector func(interface{}, []interface{}) interface{}, isLeftJoin bool, degree int) stepAction {
	return stepAction(func(src dataSource, keepOrder bool) (dst dataSource, keep bool, e error) {
		keep = keepOrder
		innerKVtask := promise.Start(func() []interface{} {
			innerKvs := From(inner).GroupBy(innerKeySelector).get().(*listSource).data
			return []interface{}{innerKvs, true}
		})

		mapChunk := func(c *chunk) (r *chunk) {
			outerKvs := getKeyValues(c, outerKeySelector, nil)
			results := make([]interface{}, 0, 10)

			if r, ok := innerKVtask.Get(); ok != promise.RESULT_SUCCESS {
				//todo:

			} else {
				innerKvs := r[0].(map[interface{}]interface{})

				for _, o := range outerKvs {
					outerkv := o.(*keyValue)
					if innerList, ok := innerKvs[outerkv.key]; ok {
						results = appendSlice(results, resultSelector(outerkv.value, innerList.([]interface{})))
					} else if isLeftJoin {
						results = appendSlice(results, resultSelector(outerkv.value, nil))
					}
				}
			}

			return &chunk{results, c.order}
		}

		switch s := src.(type) {
		case *listSource:
			outerKeySelectorFuture := parallelMapList(s, mapChunk, degree)
			dst, e = getSource(outerKeySelectorFuture, func(results []interface{}) dataSource {
				result := expandChunks(results, false)
				return &listSource{result}
			})
			return
		case *chanSource:
			out := make(chan *chunk)
			_ = parallelMapChan(s, nil, mapChunk, degree)
			dst, e = &chanSource{out}, nil
			return
		}

		return nil, keep, nil
	})
}

func getJoin(inner interface{},
	outerKeySelector func(interface{}) interface{},
	innerKeySelector func(interface{}) interface{},
	resultSelector func(interface{}, interface{}) interface{}, isLeftJoin bool, degree int) stepAction {
	return stepAction(func(src dataSource, keepOrder bool) (dst dataSource, keep bool, e error) {
		keep = keepOrder
		innerKVtask := promise.Start(func() []interface{} {
			innerKvs := From(inner).GroupBy(innerKeySelector).get().(*listSource).data
			return []interface{}{innerKvs, true}
		})

		mapChunk := func(c *chunk) (r *chunk) {
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
							results = appendSlice(results, resultSelector(outerkv.value, iv))
						}
					} else if isLeftJoin {
						results = appendSlice(results, resultSelector(outerkv.value, nil))
					}
				}
			}

			return &chunk{results, c.order}
		}

		switch s := src.(type) {
		case *listSource:
			outerKeySelectorFuture := parallelMapList(s, mapChunk, degree)
			dst, e = getSource(outerKeySelectorFuture, func(results []interface{}) dataSource {
				result := expandChunks(results, false)
				return &listSource{result}
			})
			return
		case *chanSource:
			out := make(chan *chunk)
			_ = parallelMapChan(s, nil, mapChunk, degree)
			dst, e = &chanSource{out}, nil
			return
		}

		return nil, keep, nil
	})
}

func getKeyValues(c *chunk, keyFunc func(v interface{}) interface{}, keyValues *[]interface{}) []interface{} {
	if keyValues == nil {
		list := (make([]interface{}, len(c.data), len(c.data)))
		keyValues = &list
	}
	mapSlice(c.data, func(v interface{}) interface{} {
		return &keyValue{keyFunc(v), v}
	}, keyValues)
	return *keyValues
}

//util funcs------------------------------------------
func parallelMapChan(src *chanSource, out chan *chunk, task func(*chunk) *chunk, degree int) *promise.Future {
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
					d := task(c)
					if out != nil {
						out <- d
					}
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

func parallelMapList(src dataSource, task func(*chunk) *chunk, degree int) *promise.Future {
	fs := make([]*promise.Future, degree, degree)
	data := src.ToSlice(false)
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

func reduceChan(chEndFlag chan *promise.PromiseResult, src chan *chunk, reduce func(*chunk)) {
	for {
		select {
		case <-chEndFlag:
			return
		case v, _ := <-src:
			reduce(v)
		}
	}
}

func getSource(f *promise.Future, dataSourceFunc func([]interface{}) dataSource) (dataSource, error) {
	if results, typ := f.Get(); typ != promise.RESULT_SUCCESS {
		//todo
		return nil, nil
	} else {
		//fmt.Println("(results)=", (results))
		return dataSourceFunc(results), nil
	}
}

func filterChunk(c *chunk, f func(interface{}) bool) *chunk {
	result := filterSlice(c.data, f)
	//fmt.Println("c=", c)
	//fmt.Println("result=", result)
	return &chunk{result, c.order}
}

func filterSlice(src []interface{}, f func(interface{}) bool) []interface{} {
	dst := make([]interface{}, 0, 10)

	for _, v := range src {
		if f(v) {
			dst = append(dst, v)
		}
	}
	return dst
}

func mapSlice(src []interface{}, f func(interface{}) interface{}, out *[]interface{}) []interface{} {
	var dst []interface{}
	if out == nil {
		dst = make([]interface{}, len(src), len(src))
	} else {
		dst = *out
	}

	for i, v := range src {
		dst[i] = f(v)
	}
	return dst
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

	count := 0
	chunks := make([]*chunk, len(src), len(src))
	for i, c := range src {
		switch v := c.(type) {
		case []interface{}:
			chunks[i] = v[0].(*chunk)
		case *chunk:
			chunks[i] = v
		}
		count += len(chunks[i].data)
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

//
func appendSlice(src []interface{}, v interface{}) []interface{} {
	c, l := cap(src), len(src)
	if c >= l+1 {
		return append(src, v)
	} else {
		//reslice
		newSlice := make([]interface{}, l, 2*c)
		_ = copy(newSlice[0:l], src)
		return append(newSlice, v)
	}
}

func appendChunkSlice(src []*chunk, v *chunk) []*chunk {
	c, l := cap(src), len(src)
	if c >= l+1 {
		return append(src, v)
	} else {
		//reslice
		newSlice := make([]*chunk, l+1, 2*c)
		_ = copy(newSlice[0:l], src)
		return append(newSlice, v)
	}
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

func newChunkAvlTree() *avlTree {
	return NewAvlTree(func(a interface{}, b interface{}) int {
		c1, c2 := a.(*chunk), b.(*chunk)
		if c1.order < c2.order {
			return -1
		} else if c1.order == c2.order {
			return 0
		} else {
			return 1
		}
	})
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
				(*root).sameList = make([]interface{}, 0, 2)
			}

			(*root).sameList = appendSlice((*root).sameList, e)
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
