## Go类型系统
    Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan          引用类型
	Func          实际上是一个地址？
	Interface     接口类型（可能包含引用或者值）
	Map           引用类型
	Ptr           指针
	Slice         引用类型
	String
	Struct
	UnsafePointer 内存布局是怎样的？
    
## Go内存布局
### 所有的值类型都是直接在变量的地址内存储变量的内容
### 引用类型不代表只是一个简单的指针
* Map，Slice实际上是一个包含实际数据地址的结构体。 Map的内存布局不清楚
* Func就是函数的地址？
* Chan不清楚

### interface是的数据结构比较特殊，interface{}具有特殊的内存布局
#### 定义了方法的interface
    // nonEmptyInterface is the header for a interface value with methods. From type.go
    type nonEmptyInterface struct {
        // see ../runtime/iface.c:/Itab
    	itab *struct {
    		ityp   *rtype // static interface type
    		typ    *rtype // dynamic concrete type
    		link   unsafe.Pointer
    		bad    int32
    		unused int32
    		fun    [100000]unsafe.Pointer // method table
    	}
    	word iword
    }

#### interface{}，相当于Java和C#的Object
    // emptyInterface is the header for an interface{} value. From type.go
    type emptyInterface struct {
        typ  *rtype
	    word iword
    }

#### 注意emptyInterface中的iword，通常情况下iword保存的是指向数据的指针，但如果数据的长度小于等于一个word，那iword将直接保存数据

#### 进一步探寻rtype

    type rtype struct {
        size          uintptr        // size in bytes
	    hash          uint32         // hash of type; avoids computation in hash tables
	    _             uint8          // unused/padding
	    align         uint8          // alignment of variable with this type
	    fieldAlign    uint8          // alignment of struct field with this type
	    kind          uint8          // enumeration for C
	    alg           *uintptr       // algorithm table (../runtime/runtime.h:/Alg)
	    gc            unsafe.Pointer // garbage collection data
	    string        *string        // string form; unnecessary but undeniably useful
	    *uncommonType                // (relatively) uncommon fields
	    ptrToThis     *rtype         // type for pointer to this type, if used in binary or has methods
    }

#### 值得一提的是，rtype只相当于type的基类型，很多具体类型都进一步扩展了此类型
    // arrayType represents a fixed array type.
	type arrayType struct {
		rtype `reflect:"array"`
		elem  *rtype // array element type
		slice *rtype // slice type
		len   uintptr
	}

	// chanType represents a channel type.
	type chanType struct {
		rtype `reflect:"chan"`
		elem  *rtype  // channel element type
		dir   uintptr // channel direction (ChanDir)
	}

	// funcType represents a function type.
	type funcType struct {
		rtype     `reflect:"func"`
		dotdotdot bool     // last input parameter is ...
		in        []*rtype // input parameter types
		out       []*rtype // output parameter types
	}

	// imethod represents a method on an interface type
	type imethod struct {
		name    *string // name of method
		pkgPath *string // nil for exported Names; otherwise import path
		typ     *rtype  // .(*FuncType) underneath
	}

	// interfaceType represents an interface type.
	type interfaceType struct {
		rtype   `reflect:"interface"`
		methods []imethod // sorted by hash
	}

	// mapType represents a map type.
	type mapType struct {
		rtype  `reflect:"map"`
		key    *rtype // map key type
		elem   *rtype // map element (value) type
		bucket *rtype // internal bucket structure
		hmap   *rtype // internal map header
	}

	// ptrType represents a pointer type.
	type ptrType struct {
		rtype `reflect:"ptr"`
		elem  *rtype // pointer element (pointed at) type
	}

	// sliceType represents a slice type.
	type sliceType struct {
		rtype `reflect:"slice"`
		elem  *rtype // slice element type
	}

	// Struct field
	type structField struct {
		name    *string // nil for embedded fields
		pkgPath *string // nil for exported Names; otherwise import path
		typ     *rtype  // type of field
		tag     *string // nil if no tag
		offset  uintptr // byte offset of field within struct
	}

	// structType represents a struct type.
	type structType struct {
		rtype  `reflect:"struct"`
		fields []structField // sorted by offset
	}

### go的指针操作

利用unsafe.Pointer，可以将任意指针转化为类似C中*Void的万用指针类型

### interface{}与指针

前面提高过interface{}的内存布局如下：

    type emptyInterface struct {
        typ  *rtype
        word iword
    }

所以只需自己定义一个与emptyInterface一样的struct，就可以分别得到interface{}的数据与类型。reflect.ValueOf与TypeOf已经实现了这个功能，但结合指针操作与前面提到的type类型，可以突破reflect的一些限制（比如读写struct的未公开字段)

#### 首先实现emptyInterface结构与interface{}的互相转换

此处不能直接将一个emptyInterface赋给一个interface{}变量，因为这相当于使用1个interface{}包装另一个interface{}

##### interface{}转为emptyInterface

    var i interface{}
    s := *((*emptyInterface)(unsafe.Pointer(&i)))

##### emptyInterface转为interface{}
    
    eface = emptyInterface{..., ...}
    var face inteface = *(*interface{})(unsafe.Pointer(&eface))
    
#### interface{}与nil

在雨痕的《Go学习笔记》中提到，只有一个interface{}的类型与数据都为nil时才为nil，我们可以从内存布局的角度来分析这个问题：

    func printInterfaceLayout(a interface{}) {
        fmt.Println("printInterfaceLayout", a, "isnil?", a == nil)
	    s := *((*emptyInterface)(unsafe.Pointer(&a)))
        fmt.Println(s)
    }
    var ptr *RWTestStruct2 = nil
    
    //printInterfaceLayout <nil> isnil? true
    //{<nil> 0}
    printInterfaceLayout(nil)
    
    //printInterfaceLayout <nil> isnil? false
    //{0x483120 0} 
	printInterfaceLayout(ptr)
    
    func faceAreEqual(a interface{}, b interface{}) {
        fmt.Println("faceAreEqual?", a == b)
    }
    faceAreEqual(nil, o.Ptr) //false
	fmt.Println("nil == ptr of nil?", o.Ptr == nil) //true

#### 获取对象的[]byte

#### 直接修改作为参数传递的interface{}的内容（如果内容超过一个word的长度)

#### 读写struct的非公开字段


