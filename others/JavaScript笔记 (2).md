#JavaScript笔记
###jQuery选择器
* jQuery最常见的操作就是选择DOM元素
* 常见的选择器包括：

1. 基本选择器
1. 层次选择器
1. 过滤选择器
    
###jQuery中的DOM操作

1. 查找节点
1. 创建/插入/删除/替换/复制节点
1. 读写HTML、文本和值
1. 操作节点属性
1. 遍历节点
1. 样式操作

###注意jQuery对象和Dom对象的区别


* JavaScript可以使用Unicode转义序列来代表Unicode字符，格式为`\uXXXX`，XXXX为4位十六进制数，该格式可以用于字符串直接量、正则表达式直接量与标识符中。在处理html、脚本、sql注入时，要注意这一点。
* 直接在程序中使用的数据称为直接量。除了字符串、数字、布尔外，Javascript还支持对象与数组直接量

        { x:1, y:2 } //对象
        [1,2,3] //数组
* 标识符与保留字规则与Java类似
* 定义变量名与函数名时，除了避免Javascript的保留字，还要避免使用Javascript的全局变量和函数名
* 虽然Javascript的语句结尾可以不加分号，但强烈建议总是用分号来标记语句的结束。否则Javascript的自动填补分号的规则可能造成意料之外的结果(参见2.5）。例如

        return
	    true；
	    x
	    ++
	    y
    会被解析为
    
        return；true；
	    x;++y;

###类型，值与变量
Javascript的类型分为2大类：

1. 原始类型：布尔、数字、字符串

2. 对象类型

注意Javascript中的2个特殊的原始值null与undefined，typeof(null)结果为object,typeof(undefined)结果为undefined.

Javascript的对象是属性的集合，属性是由名/值对组成。数组和函数都是Javascript的特殊对象。此外Javascipt语言核心还定义了Date、RegExp、Error三类对象。

Javascript变量是无类型的。

**注意：Javascript中的函数也是对象，因此可以将一个函数赋值给变量、作为参数传递、为函数设置属性和调用函数的方法**

####数字
Javascript中没有整数，所有数字均用浮点数表示。

* Javascript支持十进制与十六进制（0x开头）的整型直接量
* 浮点型直接量
* 算术运算
    Javascript的算术运算在溢出或被0除时返回Infinity或-Infinity，其表现与数学中的无穷大一致.  
0/0，给负数开方或与不是数字的值进行算术运算时将返回NaN（表示非数字）.NaN有个特殊的地方：**NaN不等于任何值包括自身**。即NaN != NaN，只有NaN有此特征。有一个函数isNaN()作用一样。  
注意0 == -0。
* 二进制浮点数与四舍五入
    Javascript的浮点数表示法无法精确表示0.1这样的十进制分数，所以请注意**0.3-0.2 != 0.2 - 0.1**。在进行金融运算时要特别注意此点
                
####文本
Javascript的字符串是由16位值组成的有序序列，大部分Unicode字符由16位内码表示，但**有些不能表示为16位的Unicode字符将表示为2个16位值的序列**。  
Javascript字符串的长度表示所含16位值的个数，字符串的各方法均基于16位值而不是字符。所以Javascript字符串的长度（length）不一定等于字符数。例如：

    "𝑒".length == 2

与HTML一样，Javascript的字符串可以使用"也可以使用'来包含。因此当Javascript与HTML混杂时，最好各自约定使用不同的引号。

ECMAScript5中，Javascript的字符串可以当作只读数组。在ECMAScript5和除IE外的其他浏览器中，可以使用[]来访问字符串中的单个字符。例如：

    s[0] == s.charAt(0)

####布尔
Javascript的任何值都可以转为布尔值。会被转为false的有六个值：

    undefined
    null
    0
    -0
    NaN
    "" //空字符串

这六个值称为“假值”，其他皆为“真值”。

####null与undefined

####全局对象

在Javascript中，下来属性都是属于全局对象的初始属性： 

* 全局属性，比如undefined, Infinity与NaN   
* 全局函数， 比如IsNaN()  
* Javascript语言定义的构造函数，如Date(), RegExp(), String(), Object(), Array()  
* 全局对象，如Math与JSON

在顶级的Javascript代码(不属于任何函数的Javascript代码)中, this表示全局对象  
在客户端Javascript中，Window对象表示全局对象  
在Javascript中定义的全局变量和全局函数，同样也是全局对象的一个属性

####包装对象
与Java和c#类似，Javascript可以为数字、布尔、字符串等原始类型的值创建包装对象。如果对原始对象调用属性或方法，Javascript将为每次调用自动创建一个临时的包装对象。

    var s = "test";
    s.len = 4;        //为包装对象的len属性赋值，此对象在调用完成将不会保留
    var t = s.len;    //t的值为undifined

==操作符会将原始值与包装对象视为相等

    new Number(1) == 1     //true

####类型转换
Javascript的类型转换规则见《javascript权威指南》3.8，值得注意的有这样一些转换：  
undifined -> NaN, null -> 0, "" -> 0, [] -> "", [] -> 0, [9] -> "9", [9] -> 9, null -> false, undifined -> false

* 转换与相等性  
    Javascript的“==”运算符会自动进行一些类型转换，比如下面均为true：

        [] == 0
        [9] == "9"
        "" == 0
        "0" == 0
        0 == true

    但要注意, JavaScript在进行==运算时不会自动将值转为布尔值，例如下列结果均为false：

        undifined = false
        null == false
        null == 0
        [] == true

* 显式类型转换  
    常见的显式类型转换方法：  
    使用Boolean()、Number()、String()或Object()  
    使用运算符进行类型转换：
     
        x + ""    //等价于String(x)
        +x        //等价于将对象转为数字
        !!x       //等价于将对象转为boolean类型
        !!null == false     //结果为true
        +[]       //0

    使用toString()
    使用Number的toFixed(), toExponential(), toPrecision()
    使用全局函数parseInt()与parseFloat()

* 对象转为原始值  
    对象->布尔值的规则：所有对象均转为true，所以有  

        !!(new Boolean(false))     //true
        !![]    //true
        !!{}    //true

    对象->字符串的规则： 
    
    1. 如果对象有toString(）方法并且toString()返回一个原始值，Javascript将这个值转为字符串返回  
    2. 如果第一条不成立，则尝试以toValue()函数进行同样处理  
    3. 如果1.和2.都失败，则抛出类型错误异常  
        
    对象->数字的规则与对象转换为字符串的规则类似，但尝试顺序是先toValue()，再toString()  
    Date的转换有些特殊：  

        var now = new Date();
        typeof(now + 1);      //string
        typeof(now - 1);      //number

####变量声明与作用域

使用var声明变量  
Javascript的变量作用域包括全局作用域和函数作用域，没有Java和c#中的块级作用域的概念。  
Javascript局部变量不受声明语句位置的影响，在整个函数体内都可见，下面两段代码是等价的：

        function f(){
            alert(scope);
            var scope = "a";
        }

        function f(){
            var scope;
            alert(scope);
            scope = "a";
        }

* 变量即属性

    使用var定义一个Javascript全局变量实际是定义了全局变量的一个属性，与this.方式定义的不同在于这个属性不可配置，无法删除。

        var i = 1;
        this.j = 2;
        delete i;    //false, 无法删除
        delete j;    //true, 可以删除

    局部变量与全局变量类似，同样是与函数调用相关的一个对象的属性（不像全局对象可以通过this引用，这个存放局部变量的对象是不可引用的）。  

* 作用域链

    每段Javascript代码都有一个与之对应的作用域链，当Javascript查找一个变量的值时，会从链中的第一个对象开始查找，直到找到对应于变量的属性被找到为止。  
    Javascript的顶层代码中，作用链域由1个全局对象组成。从顶层代码开始，每个函数都有1个对象保存其局部变量，每次调用都会产生一个新对象来保存局部变量并附加到上级函数的作用域链，由此得到该次函数调用的作用域链。

        var i = 1;
        //对嵌套函数g()来说，其作用域链有3个对象，  
        //分别对应全局变量，f函数的参数和局部变量，g函数自己的参数和局部变量
        function f(a) { 
            return function g() { 
                var b = 2; return i+a+b;
            };
        }   
    注意JavaScript代码关联的作用域链中保存的是变量的引用，而不是变量的值。
    
        var i = 1;

        function f(a){
            return {
            	get: function() { 
        			return i + a;
        		},
        		set: function(v){
        			a = v;
        		}
        	};
        }
        
        var o = f(1);
        o.get();   // 返回2
        i = 2;
        o.get();   //返回3
        o.set(2);
        o.get();   //返回4
    
    理解作用域链的概率是理解闭包的关键。

###表达式和运算符
####原始表达式
####对象与数组初始化表达式
注意JavaScript的数组直接量运行这样写`[1,,,5]`，省略的元素将是undifined。
####函数定义表达式

    var square = function(x) { 
        return x * x; 
    };

####属性访问表达式
因为JavaScript的对象也是属性的集合，所以支持两种属性表达式：

    expression.identifier
    expression[ expression ]
    
####调用表达式
注意JavaScript中方法调用与函数调用的区别，如果被调用的函数是一个对象的属性，则称为“方法调用”。  
在方法调用中，this是调用的对象，而在函数调用中，this是全局对象或undifined（根据是否采用ECMAScript 5的严格模式）
####对象创建表达式
对象创建表达式的形式为`new Point(2,3)`。  
对象创建表达式的流程为：  
1. 创建空对象
1. JavaScript传入指定参数并以新对象为this来调用指定函数1. 如果函数没有return一个对象值，则新对象就是对象创建表达式的值
1. 如果函数有return对象值，则刚才创建的新对象将被废弃，而返回值将作为对象创建表达式的值

####运算符
#####算术运算符
注意"/"，与Java不同，JavaScript没有整型，所以/的结果总是浮点数。  
"+"运算符的行为：如果有一个操作数为对象，首先转换对象为原始值（参考类型转换），然后如果有任一原始值为字符串，则执行字符串连接，否则都转为数组进行加法操作
#####关系运算符
* “====”严格相等运算符
* “=="相等运算符
    两者的区别在于“==”会自动进行类型转换  
    注意字符串的比较方式是比较对应位的16位值，而不是显示出的字符。
* 比较运算符 （> < <= >=）
    注意比较运算符的类型转换规则是优先使用数字比较
* instanceof运算符
    JavaScript中的类是通过构造函数定义的，因此instanceof的右操作数是一个函数。对于 o instanceof f，如果f.prototype在o的原型链中存在，则instanceof运算符的结果为true。

#####逻辑运算符
与Java语言不同的是，在JavaScript中，任何类型的值都可以当作假值或真值，因此JavaScript中的&&与||并非只返回布尔值的结果。准确地说：

* 对于a1 && a2 && a3，如果a1, a2, a3全部为真值，表达式将返回a3的值，否则返回左起第一个假值，后面的值不会计算（此时计算后面的值不会影响&&的结果）。

        var o = { x: 1};
        var p = null;
        p && o; //返回null
        o && p && o.x; //返回null
        o && o.x; //返回1
    
* 对a1 || a2 || a3，如果a1, a3, a3全部为假值，表达式将返回a3的值，否则返回左起第一个真值，后面的值不会计算。

        var o = { x: 1};  
        var p = null;  
        var u = undifined;  
        p || u; //返回undifned  
        p || o.x || u; //返回1
        p || u || o.x; //返回1
        
* !运算符与&&, ||不同，总是将操作数转换为布尔后运算，因此总是返回true或false。

#####赋值运算符
#####eval运算符
* eval()将一个字符串作为JavaScript源代码进行解释执行，并返回字符串中最后一个语句或表达式的值，实际上eval()是一个被当作运算符对待的函数。  
* eval()使用调用它的变量作用域，如果在顶层调用，将使用全局作用域。
* 直接eval与全局eval
    在ES3中，不允许对eval()赋予别名。而在ES5中，对eval()赋予别名并调用将使用全局对象作为上下文作用域，而无法读写和定义局部变量和函数。不使用别名而直接使用eval()在ES5中称为“直接eval”。  
        
        var geval = eval;
        var x = "g", y = "g";
        function f(){
            var x = "l";
            eval("x += 'changed';");
            return x;
        }
        function g(){
           var y = "l";
           geval(y += 'changed';");
           return y;
        }
        console.log(f(), x);  //输出”lchanged g"
        console.log(g(), y);  //输出"l gchanged, "

* 严格eval()
    如果在ES5严格模式下调用eval()，eval执行的代码段将不能在局部作用域中定义新的变量和函数。并且eval将成为保留字，无法取别名。

#####其他运算符
* typeof
* delete
* void
* ,

###语句
注意JavaScript允许空语句;，结合JavaScript自动插入;的特性，可能会带来一些隐蔽的问题：

    if ( a == 0 )
        o = null;

等价于

    if (a == 0);
    o = null;
    
####函数定义语句与函数定义表达式
函数定义语句与函数定义表达式都可以定义新的函数对象。但在标准中，JavaScript定义语句不能出现在if、while等块内。并且函数定义表达式的变量声明和函数体是分开的，当JavaScript对变量定义和函数定义进行“提前”的时候，函数定义表达式不会将函数体提前。

####switch语句
注意switch语句中对case语句进行匹配操作时使用的是“====”操作符而不是“==”，因此不会进行任何类型转换。
####for/in循环
    for( var in object)
        statement

####for循环语句
for循环枚举对象的“可枚举”属性，并将属性名赋给var表达式。注意var除了可以是变量名，还可以是任何“左值”的表达式，例如：

    var a = [], i = 0;
    var o = { x:1, y:2};
    for (a[i++] in o); //将所有对象属性名复制到一个数组
####try/catch/finally语句
注意，JavaScript中如果在finally块中使用return语句返回，将导致被抛出的异常被忽略。

    var foo = function(){　
        try{ 
            throw new Error('aa'); 
        } 
        finally{ 
            return 1; 
        } 
        return 3;
    };
    foo();//返回1
    
    var foo = function(){　
        try{ 
            throw new Error('aa'); 
        } 
        finally{ 
        } 
        return 3;
    };
    foo();//抛出异常aa

    var foo = function(){　
        try{ 
            throw new Error('aa'); 
        } 
        finally{ 
            throw new Error('bb');
            return 1; 
        } 
        return 3;
    };
    foo();//抛出异常bb
    
####with语句
    with(object)
    statement
with语句作用是将object添加到作用域链的头部，然后执行statement。在JavaScript中不建议使用with语句。

####debugger语句
debugger语句什么也不做。当开启了JavaScript调试器后，程序将在debugger处暂停执行，相当与一个断点。

####"use strict"严格模式

###对象
JavaScript的对象可以看作属性的无序集合，每个属性都是一个名/值对。 

与普通字典不同的是，JavsScript的对象还可以从原型对象（prototype)继承属性。这种原型式继承是JavaScript的对象与Java、.Net之类的面向对象最大的不同之处。 

对象的属性除了名/值，还拥有一些相关的特性值（property attribute）：

* 可读
* 可写
* 可配置

除了属性，对象还拥有3个特性：

* 原型对象
* 对象的类
* 对象的扩展标记(extensible flag)指明在ES5中是否可以为对象添加新属性

####类和原型
与Java等面向对象的语言不同，JavaScript中的对象没有父类的概念。绝大多数JavaScript对象都与1个原型对象相关联，并从原型对象继承属性。  
没有原型的对象包括null, Object.prototype(类似与Java中的Object，所有对象的原型对象最终都继承自此对象），以及用Object.create(null)创建的对象。  
每个对象都有原型对象，原型对象又有自己的原型对象，一直上溯到Object.prototype，构成了对象的“原型链”。

####创建对象的方法
* 对象直接量
    
        var point = { x:1, y:1 };

用这种方式创建的对象，原型对象为Object.prototype。

* 通过new创建对象

        var o = new Object(); //与var o = {}一样
        
new后面是类的构造函数（所以在JavaScript中，Object、Date都是函数）。对象的原型对象就是构造函数的prototype属性对象。

* 通过Object.create()

        var o = Object.create({ x:1 });
        
Object.create是在ES5中定义的。通过Object.create创建的对象将使用obj作为自己的原型对象，可以使用null来作为新对象的原型对象，但这样的对象没有原型，不会继承任何的属性和方法。

        var o = Object.create(null);
        
####继承
在JavaScript中，当访问对象o的属性x时，首先在对象o自身中查找，如果不存在，则到o的原型对象中查找，如果没有则继续道原型对象的原型中去查找。。。。。。直到遍历所有的原型对象或找到为止。以此实现与Java等面向对象类似的效果。  

查找原型链只限于读操作，如果对属性x赋值而属性名在对象中不存在，JavaScript会直接在对象上创建新属性x并赋值而不会去修改原型对象。

    var x1 = { x:1 };
    var x2 = Object.create(x1);
    x2.x = 2;
    x2.x;   //2，x2上会创建属性x并赋值
    x1.x;   //1，x1的x值没有被覆盖
    
上面的描述有例外情况，如果属性x在原型对象中存在并且只读，那么不会创建新属性，而是赋值失败。

####检测属性
* for/in循环 循环对象的属性，包括从原型链继承的属性
* hasOwnProperty() 循环对象的自有属性
* propertyIsEnumrable() 循环对象自有的可枚举属性

####原型属性
ES5中可以使用Object.getPrototypeOf()可以查询对象的原型属性，但在ES3中没有。在ES3中如果没有修改prototype对象，JavaScript通常会在prototype中设置一个constructor属性来指向对象的构造函数，所以通常可以通过o.constructor.prototype来查询对象的原型属性（使用Object.create的不行）。

####对象的可扩展性
在ES5中，可以利用Object.preventExtensions(), Object.seal(), Object.freeze()来设置对象的可扩展性。

####Json序列化
ES5中可以使用JSON.stringify()和JSON.parse()来序列化和还原JavaScript对象。注意并非对象的所有属性都可以还原。  
NaN, Infinity和-Infinity序列化的结果是null, 函数，RegExp，Error和undefined值不能序列化和还原。  
Date序列化为ISO格式的字符串，但JSON.parse()不会将此字符串还原为Date对象。

###函数
####函数定义
注意函数定义语句与函数定义表达式的区别。

####函数调用的4种方法
* 函数调用
* 方法调用
    方法调用和函数调用最大的区别是this关键字的含义不同，函数调用中的this或者为全局对象（非严格模式）或为undefined（严格模式），方法调用中的this为调用函数体的上下文对象。  
    请注意由于JavaScript的函数也是值，所以任意结果值为函数的表达式，都可以直接作为函数调用：

        function add(x, y) { 
            console.log(this);
            return x + y; 
        }
        var a1 = add( 1, 2); //函数调用
        var a2 = { add : add );
        a2.add(1, 2);    //方法调用
        a2['add'](1, 2); //与a2.add(1, 2)相同
        
        (function () { 
            return 1;}
        )();             //直接调用匿名函数
        
        var f1 = function() { 
            return function(){
                return 1;
            }
        };
        f1()();          //f1返回一个函数，该函数的返回值又是一个函数
        
    在JQuery中，常见一种名为链式方法的设计风格：
    
        $(':header').map(function() { return this.id }).get().sort();
        
    这种风格的API调用比较直观。实现的方式是如果方法不需要返回值，则直接返回this即可。
        
* 构造函数调用
    构造函数的执行过程可以参照前面对象创建表达式的描述，另外要注意构造函数的调用上下文始终是新创建的对象。即使如下面的代码也是如此：

      var newObj = new o.m();

* 通过call()和apply()调用
    JavaScript中的函数对象包含2个方法call()和apply()可以用来调用函数，这两个方法可以显式指定调用的this对象。

        function f(a) { 
            if (this) {
                this.a = a; //如果this存在，则为this的a属性赋值
            } 
            return a + 1; 
        }
        
        var o = {};      //新对象o
        o.f(1);          //TypeError: Object #<Object> has no method 'f'
        f.call(o, 1);    //return 2
        o.a;             //o有了属性a
        f.call(o, [1]);  //apply接受数组作为参数，与f.call(o, 1)作用相同
        
####函数参数
与Java不同的是，JavaScript的函数可以接受任意数量的参数，也不检查参数类型，因此在JavaScript中没有函数重载的概念。

* JavaScript允许调用时指定的参数数量小于函数声明的形参数量，此时未指定的参数值为undefined。
    如果需要给未赋值的参数赋予默认值，最简单的方法是：

        a = a || defaultValue; //这里的defaultValue代表你想赋予的默认值
        
* JavaScript中可以使用arguments来获得指向实参对象数组的引用。

        function f(x,y) { 
            console.log(arguments.length);
            console.log(arguments[0]);
        }
        
        f(1, 3);  //输出2 1
        
    值得注意的是，在非严格模式中，arguments对象中保存的是实参的一个别名，也就是说，无论修改实参还是arguments对象中的相应值，两者都会获取到更新后的值。
    
        function f(x){
			arguments[0] = 1;
			console.log(x);
			x= 2;
			console.log(arguments[0])
		}; 
        
    arguments对象还具有callee和caller属性，前者代表当前正在执行的函数，后者代表调用当前函数的函数( 非标准，在chrome上测试失败)。
        
        function t() {
			console.log(arguments.caller);
		}
		(function() { t(1) })(); //输出undefined
		
		function t() {
			console.log(arguments.callee);
		}
		(function() { t(1) })();  //输出function t() {console.log(arguments.callee);}
        
* 如果JavaScript函数的参数很多，调用者将很难记住参数顺序。可以利用对象属性作为参数来改善这个问题，例如JQuery的Ajax函数：

        $.ajax({
            type: "GET",
            url: “test.js",
            dataType: "script"
        });
        
#### 使用函数作为命名空间
#### 闭包

使用闭包时要注意this和arguments,他们会随着函数上下文的改变而自动改变，因此不能在闭包内使用this和arguments得到闭包外的this和arguments值。
    
    function f(){
        var self = this;
        return function(){ 
            //return this.x;   //这里的this将指向全局对象
            return self.x;     //使用self才能得到f函数对应的this
        }
    }
    
####函数的属性和方法
* length属性
    函数的length属性表示函数形参的个数
* prototype属性
* call()和apply()方法
* bind()方法（ES5)
* toString()方法

###类和模块
####原型和构造函数

调用构造函数创建对象时，构造函数的prototype属性会被当作新对象的原型。  
不同的构造函数可以使用相同的原型对象。  
在ES3中，对象无法知道自己的原型对象，但对象将从构造函数的prototype对象继承属性，因此可以采取下面的方法让对象获得自己的原型对象：  
    1. 为构造函数的prototype对象增加constructor属性，指向构造函数自身（JavaScript默认）  
    2. 通过构造函数创建对象，对象从构造函数的prototype对象继承constructor属性  
    3. 由于constructor属性指向构造函数，因此可以通过obj.constructor.prototype获得对象的原型对象
    
        var o = {x:1};
        o.constructor;
        o.constructor.prototype;
        
        function Point( x, y ){
            this.x = x;
            this.y = y;
        }
        o = new Point(1,2);
        o.constructor; //function Point( x, y ){...}
        o.constructor.prototype;

        //为原型对象添加方法，不会覆盖默认的constructor属性
        Point.prototype.getX() = function{
            return this.x;
        }
        o.constructor; //function Point( x, y ){...}

        //直接覆盖prototype属性，constructor属性从原型中消失了
        Point.prototype = {
            z : 1
        };
        o = new Point(1,2);
        o.constructor;   //function Object() { [native code] }

        //手工补上constructor属性
        Point.prototype = {
            constructor : Point,
            z : 1
        };   
        o = new Point(1,2);
        o.constructor;   //function Object() { [native code] }


####检测对象的类型
* instanceof运算符
    instanceof运算符并不是检测对象是否从指定的函数构造而来，而是检测函数的prototype属性是否存在于对象的原型链上。  

        var p = {x:1};
        function C1() {}
        function C2() {}
        C1.prototype = p;
        C2.prototype = p;
        
        var o = new C1();
        o instanceof C1;   //true
        o instanceof C2;   //true

* constructor属性
    利用constrctor属性，可以检测对象是否从指定函数构造而来，但请注意，在前面的一些例子中可以发现某些情况下对象并没有constructor属性或者constructor属性并不指向真正的构造函数。

* 构造函数名称
    JavaScript中函数的toString()函数默认返回函数的内容，因此可以通过分析这个字符串得到构造函数的名称，但与是哟娜constructor属性一样，不是所有对象都有construtor属性，也不是所有函数都有名称（比如匿名函数）。

####鸭子类型
“像鸭子一样走路，游泳并且嘎嘎叫的鸟就是鸭子”

####函数式编程
#####使用函数处理数组
ES5中的一些数组方法：

* forEach() //循环数组，并为每个元素调用指定函数  
* map()     //针对每个元素调用指定函数，并组成一个新数组返回 
            var a = [2,3,4];
            a.map( function ( x ) {
                return x * x;
            });   //[4, 9, 16]

* filter()  //过滤符合条件的元素得到一个新数组  
* reduce()  //对数组进行组合操作，得到一个单一值  
            var a = [2,3,4];
            a.reduce( function ( x, y ) {
                return x * y;
            });   //24

#####高阶函数
高阶函数是操作函数的函数，它以函数为参数并返回新函数。

    function operateArray( arrFunc, f){
        return function (a) { 
            return arrFunc.call(a, f);
        };
    }
    
    var mapperSquare = operateArray(
        Array.prototype.map, 
        function(x) { return x*x;}
    );
    
    var reduceSquare = operateArray(
        Array.prototype.reduce, 
        function(x, y) { return x*y;}
    );
    
    var a = [2,3,4];
    mapper(a); //[4,9,16]
    reduceSquare(a); //24
    
#####一个AOP函数的例子

    function aop(f, before, after){
	    return function(){
		    var b = before.apply(this, arguments);
            if (b && !(b.canRun)) return b.value;
		
		    var returnValue = f.apply(this, arguments);
		
            //arguments不能添加元素，所以将其转为一个数组，将返回值添加到数组中。注意arguments并不是真正的数组，但可以调用数组的方法，请参考鸭子类型的概念		
			var args = Array.prototype.map.call(arguments, function(x) { return x; });
            args.push(returnValue);
		    after.apply(this, args);
		    return returnValue;
	    }
    }

    var square = function(x) { console.log("square"); return x*x; }
    var before = function() { console.log("before", arguments); }
    var after = function() { console.log("after", arguments); }

    var aopSquare = aop(square, before, after);
    aopSquare(3);
    /*输出下面内容：
    before [3]
    square
    after [3, 9]
    9
    */
    
    //下面利用aop函数实现一个缓存的aop函数
    //定义简单的缓存函数
    function aopCache() {
	    var cache = {};
	    return {
		    before: function(key){
	            var key = arguments.length + Array.prototype.join.call(arguments, ",");
			    if (key in cache) {
				    return { canRun: false, value: cache[key] };
    		    }
			    else{
				    return null;
			    }
		    },
		    after : function(key, value){
    	        var returnValue = Array.prototype.pop.call(arguments);
	            var key = arguments.length + Array.prototype.join.call(arguments, ",");
			    cache[key] = returnValue;
	    	}
	    };
    }
	var cache = aopCache();
    var cacheSquare = aop(square, cache.before, cache.after);

    cacheSquare(4);
    /* 输出
    square
    16
    */
    cacheSquare(4);
    /* 输出
    16
    */    
    
    //测试aop能否正确处理this
    var o1 = { 
        x : 1, 
        square: function() { 
            console.log("o1 square"); 
            return this.x * this.x}
    };
    o1.square = aop(o1.square, cache.before, cache.after);
    o1.x = 2;
    o1.square();
    /* 输出
    o1 square
    4
    */
    o1.square();
    /* 输出
    4
    */
    
    


        
        

        
    



