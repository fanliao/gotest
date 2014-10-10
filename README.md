gotest
======

Test code for study go

Go标准库中提供了Sync.Once来实现“只执行一次”的功能。学习了一下源代码，里面用的是经典的双重检查的模式：

对并发编程的良好支持是Golang的最大卖点之一。通过轻量级的Goroutines和Channel，我们可以较容易地实现代码的异步执行，Goroutines之间的同步与协调、多路复用等并发功能。但直接用Goroutines和Channel来实现大规模的异步程序仍然不是一件轻松的事情。Promise/Future模式作为一种经典的并发模式，为此提供了更高层次的抽象来简化并发编程。

### Promise/Future
Promise/Future模式的介绍可以参考 [Wiki](http://en.wikipedia.org/wiki/Futures_and_promises) ，或者参考Scala的这份文档[Future and Promise](https://code.csdn.net/DOC_Scala/chinese_scala_offical_document/file/Futures-and-Promises-cn.md) 。

简单来说，Promise和Future代表了一个未完成的异步任务。我们可以为任务实现下面的操作：

* 添加回调函数实现非阻塞的操作
* 也可以阻塞式地等待任务的执行结果
* 建立Future的管道，一个Future可以在完成后开始多个后续Future的执行，也可以让多个Future全部或者任意一个完成后开始另一个Future
* 取消Future的执行

Promise与Future的区别在于，Future是Promise的一个只读的视图，也就是说Future没有设置任务结果的方法，只能获取任务执行结果或者为Future添加回调函数。

### 开始Go-Promise
项目地址在https://github.com/fanliao/go-promise

