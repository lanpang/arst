### golang1.12 sync.Mutex源码解析

sync.Mutex是go标准库中常用的一个排外锁。当一个goroutine获得这个锁的拥有权
以后，其他请求锁的goroutine会阻塞在`Lock`方法调用上，直到锁被释放。

sync.Mutex的实现是经过多次的演化，增加公平处理和饥饿机制。

互斥锁有两种状态，正常状态和饥饿状态。

在正常状态下，所有等待的goroutine按照FIFO顺序等待。唤醒的goroutine不会直接拥有锁，
而是会和新请求锁的goroutine竞争锁的拥有。新请求锁的goroutine具有优势:他正在cpu上执行，而且可能有好几个，所以刚刚唤醒的goroutine有很大可能在锁竞争中失败。在这种情况下，这个
被唤醒的goroutine会加入到等待队列的前面。如果一个等待的goroutine超过1ms没有获取锁，那么它将会把锁转变为饥饿模式.  

在饥饿模式下，锁的所有权将从unlock的goroutine直接交给等待队列中的第一个。新来的
goroutine讲不会尝试去获得锁，即使锁看起来是unlock状态，也不会去尝试自旋操作，而是放在
等待队列的尾部。  

如果一个等待的goroutine获取锁，并满足一下其中任何一个条件:  
(1) 它是队列中的最后一个；
(2) 它等待的时候小于1ms  
它会将锁的状态转换为正常状态。  

正常状态有说明锁的性能很好，饥饿模式可以改善尾部延迟，更公平。

```
type Mutex struct{
    state int32
    sema uint32
}
```

state是一个共用字段，
* 第0个bit标记这个mutex是否已被某个goroutine所拥有
* 第1个bit标记这个mutex是否已唤醒
* 第2个bit标记这个mutex是否处于饥饿

尝试获取mutex的goroutine也是有状态的  
* 可能是新来的goroutine.  
* 可能是刚唤醒的goroutine.  
* 可能是处于饥饿状态的goroutine.  

```
func (m *Mutex) Lock() {
    // Fast path: grab unlocked mutex.
// 如果mutex的state没有被加锁，也没有等待/唤醒的goroutine,本goroutine直接获得锁
//mutex饥饿？ 如何处理
    if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
        if race.Enabled {
            race.Acquire(unsafe.Pointer(m))
        }
        return
    }   
// 计算本goroutine的等待时间
    var waitStartTime int64
// 本goroutine是否处于饥饿状态
    starving := false
// 本goroutine是否已唤醒
    awoke := false
// 本goroutine自旋次数
    iter := 0
// 复制锁的当前状态
    old := m.state
    for {
        // Don't spin in starvation mode, ownership is handed off to waiters
        // so we won't be able to acquire the mutex anyway.
        // 第一个条件是state已被锁，但是不是饥饿状态。
        // 如果时饥饿状态，自旋时没有用的，锁的拥有权直接交给了等待队列的第一个。
        // 可以自旋，多核，压力不大并且在一定次数内可以自旋，
        // 如果满足这两个条件，不断自旋来等待锁被释放、或者进入饥饿状态、
        // 或者不能再自旋。
        if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
            // Active spinning makes sense.
            // Try to set mutexWoken flag to inform Unlock
            // to not wake other blocked goroutines.
            // 自旋过程中如果发现state还没有设置woken标识，设置，并标记自己为被唤醒
// 第一次掉用lock函数，woken没有被标识
            if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
                atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
                awoke = true
            }
            runtime_doSpin()
            iter++
            old = m.state
            continue
        }
// 到这里，state的状态可能是:
//1. 锁还没有被释放，锁处于正常状态
//2. 锁还没有被释放，锁处于饥饿状态
//3. 锁还已经被释放，锁处于正常状态
//4. 锁还已经被释放，锁处于饥饿状态

// 本goroutine的awoke可能是true,也可能是false(其他goroutine已经设置了state的awoken)

// new 复制state的当前状态，用来设置新的状态
// old 是锁当前的状态
        new := old 
        // Don't try to acquire starving mutex, new arriving goroutines must queue.

// 如果old state状态不是饥饿状态，new state设置锁，尝试通过cas获取锁
// 如果old state状态是饥饿状态，则不设置new state的锁，因为饥饿状态,unlock以后，
//锁直接 等待队列的第一个goroutine
        if old&mutexStarving == 0 { 
            new |= mutexLocked
        }
        // 将等待队列数量加1
        if old&(mutexLocked|mutexStarving) != 0 { 
            new += 1 << mutexWaiterShift
        }
        // The current goroutine switches mutex to starvation mode.
        // But if the mutex is currently unlocked, don't do the switch.
        // Unlock expects that starving mutex has waiters, which will not
        // be true in this case.
// 如果当前goroutine处于饥饿状态，并且old state已被加锁
// 将new state的状态标记为饥饿状态,将锁转变为饥饿状态
// 如何将锁状态转换为饥饿状态呢？？？
        if starving && old&mutexLocked != 0 { 
            new |= mutexStarving
        }
// 如果本goroutine已经设置为唤醒状态, 需要清除new state的唤醒标记, 
// 因为本goroutine要么获得了锁，要么进入休眠，
        // 总之state的新状态不再是woken状态.
        if awoke {
            // The goroutine has been woken from sleep,
            // so we need to reset the flag in either case.
            if new&mutexWoken == 0 {
                throw("sync: inconsistent mutex state")
            }
            new &^= mutexWoken
        }

        // 通过CAS设置new state值.
        // 注意new的锁标记不一定是true, 也可能只是标记一下锁的state是饥饿状态.

        if atomic.CompareAndSwapInt32(&m.state, old, new) {
            // 如果old state的状态是未被锁状态，并且锁不处于饥饿状态,
            // 那么当前goroutine已经获取了锁的拥有权，返回
            if old&(mutexLocked|mutexStarving) == 0 {
                break // locked the mutex with CAS
            }
            // If we were already waiting before, queue at the front of the queue.
            // 设置/计算本goroutine的等待时间
            queueLifo := waitStartTime != 0
            if waitStartTime == 0 {
                waitStartTime = runtime_nanotime()
            }
            // 既然未能获取到锁， 那么就使用sleep原语阻塞本goroutine
            // 如果是新来的goroutine,queueLifo=false, 加入到等待队列的尾部，耐心等待
            // 如果是唤醒的goroutine, queueLifo=true, 加入到等待队列的头部
            runtime_SemacquireMutex(&m.sema, queueLifo)
//runtime_SemacquireMutex会调用gopark,使goroutine sleep
//在runtime_ReleaseMutex 会将某个goroutine唤醒
//唤醒的goroutine接着在从这里开始执行

// 计算当前goroutine是否处于饥饿状态
            starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
//得到锁当前的状态
            old = m.state

// 如果当前的state已经是饥饿状态
// 那么锁应该处于Unlock状态，那么锁应该直接被交给了本goroutine
            if old&mutexStarving != 0 {
// 如果当前的state已经是饥饿状态
// 那么锁应该处于 Unlock状态，那么应该是锁被直接交给本goroutine
                if old&(mutexLocked|mutexWoken) != 0 || old>>mutexWaiterShift == 0 {
                    throw("sync: inconsistent mutex state")
                }
// 当前goroutine用来设置锁，并将等待的goroutine数减1
                delta := int32(mutexLocked - 1<<mutexWaiterShift)
                if !starving || old>>mutexWaiterShift == 1 {
// 退出饥饿模式
                    delta -= mutexStarving
                }
// 设置新state,因为已经获得了锁，退出、返回
                atomic.AddInt32(&m.state, delta)
                break
            }

// 如果当前的锁是正常模式，本goroutine被唤醒，自旋次数清零，从for循环开始处重新开始
            awoke = true
            iter = 0
        } else { // 如果cas不成功，重新获取锁的state，从for循环处重新开始
            old = m.state
        }
    }

    if race.Enabled {
        race.Acquire(unsafe.Pointer(m))
    }
}

```


####  2 Unlock

```
func (m *Mutex) Unlock() {
    if race.Enabled {
        _ = m.state
        race.Release(unsafe.Pointer(m))
    }   

    // Fast path: drop lock bit.
// 如果state不是处于锁的状态，那么就是Unlock根本没有加锁mutex,panic
    new := atomic.AddInt32(&m.state, -mutexLocked)
    if (new+mutexLocked)&mutexLocked == 0 { 
        throw("sync: unlock of unlocked mutex")
    }   

// 释放了锁，还的需要通知其他等待者
// 锁处于饥饿状态，直接交给等待队列的第一个，唤醒它，让它获取锁
    if new&mutexStarving == 0 { 
        old := new 
        for {
// 如果没有等待的goroutine,或者锁不处于空闲的状态，直接返回
            if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 { 
                return
            }
            // 将等待的goroutine数减一，设置woken标识 
            new = (old - 1<<mutexWaiterShift) | mutexWoken

// 设置新的state,这里通过信号量唤醒阻塞的goroutine去获取锁
            if atomic.CompareAndSwapInt32(&m.state, old, new) {
                runtime_Semrelease(&m.sema, false)
                return
            }
            old = m.state
        }
    } else {
//饥饿模式下，直接将锁的拥有权传给等待队列的第一个
// 注意此时的state的mutexlocked还没有加锁，唤醒的goroutine会设置它
// 在此期间，如果所有的goroutine来请求锁，因为mutex处于饥饿状态，mutex还是被人认为
// 处于锁状态 新来的goroutine不会把锁抢过去
        runtime_Semrelease(&m.sema, true)
    }   
}

```
