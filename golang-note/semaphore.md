### Semaphore
在学习golang 的mutex过程中，发现`mutex.Lock`使用了`runtime_SemacquireMutex`,  
学习了一下golang 的`runtime_SemacquireMutex`的实现方式。本文主要讲述`runtime_SemacquireMutex`  
具体实现以及一些自己的思考.  

`runtime_SemacquireMutex`在`runtime/sema.go` 中  
#### 数据结构
 Go 语言中暴露的 semaphore 实现
 具体的用法是提供 sleep 和 wakeup 原语
 以使其能够在其它同步原语中的竞争情况下使用
 因此这里的 semaphore 和 Linux 中的 futex 目标是一致的
 只不过语义上更简单一些

 也就是说，不要认为这些是信号量
 把这里的东西看作 sleep 和 wakeup 实现的一种方式
 每一个 sleep 都会和一个 wakeup 配对
 即使在发生 race 时，wakeup 在 sleep 之前时也是如此

 See Mullender and Cox, ``Semaphores in Plan 9,''
 http://swtch.com/semaphore.pdf

 为 sync.Mutex 准备的异步信号量

 semaRoot 持有一棵 地址各不相同的 sudog(s.elem) 的平衡树
 每一个 sudog 都反过来指向(通过 s.waitlink)一个在同一个地址上等待的其它 sudog 们
 同一地址的 sudog 的内部列表上的操作时间复杂度都是 O(1)。顶层 semaRoot 列表的扫描
 的时间复杂度是 O(log n)，n 是被哈希到同一个 semaRoot 的不同地址的总数，每一个地址上都会有一些 goroutine 被阻塞。
 访问 golang.org/issue/17953 来查看一个在引入二级列表之前性能较差的程序样例，test/locklinear.go
 中有一个复现这个样例的测试

```
type semaRoot struct {
    lock mutex
    treap *sudog //
    nwait uint32 // 等待者数量，read w/o the lock
}

const semTabSize = 251

var semtable [semTableSize]struct {
    root semaRoot
//确保semtable的一个元素占用一个cacheline,不会跨行，保证高性能
    pad [sys.CacheLineSize - unsafe.Sizeof(semaRoot{})]byte
}

func semroot(addr *uint32) *semaRoot {
// addr是一个指针，指针8bytes对齐的,所以直接右移3位，减少运算。amd64
    return &semtable[(uintptr(unsafe.Pointer(addr))>>3)%semTabSize].root
}
```
