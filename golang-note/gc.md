### Garbage Collection

#### 大致流程

```
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector (GC).
//
// The GC runs concurrently with mutator threads, is type accurate (aka precise), allows multiple
// GC thread to run in parallel. It is a concurrent mark and sweep that uses a write barrier. It is
// non-generational and non-compacting. Allocation is done using size segregated per P allocation
// areas to minimize fragmentation while eliminating locks in the common case.
//
// The algorithm decomposes into several steps.
// This is a high level description of the algorithm being used. For an overview of GC a good
// place to start is Richard Jones' gchandbook.org.
//
// The algorithm's intellectual heritage includes Dijkstra's on-the-fly algorithm, see
// Edsger W. Dijkstra, Leslie Lamport, A. J. Martin, C. S. Scholten, and E. F. M. Steffens. 1978.
// On-the-fly garbage collection: an exercise in cooperation. Commun. ACM 21, 11 (November 1978),
// 966-975.
// For journal quality proofs that these steps are complete, correct, and terminate see
// Hudson, R., and Moss, J.E.B. Copying Garbage Collection without stopping the world.
// Concurrency and Computation: Practice and Experience 15(3-5), 2003.
//
// 1. GC performs sweep termination.
//
//    a. Stop the world. This causes all Ps to reach a GC safe-point.
//
//    b. Sweep any unswept spans. There will only be unswept spans if
//    this GC cycle was forced before the expected time.
//
// 2. GC performs the mark phase.
//
//    a. Prepare for the mark phase by setting gcphase to _GCmark
//    (from _GCoff), enabling the write barrier, enabling mutator
//    assists, and enqueueing root mark jobs. No objects may be
//    scanned until all Ps have enabled the write barrier, which is
//    accomplished using STW.
//
//    b. Start the world. From this point, GC work is done by mark
//    workers started by the scheduler and by assists performed as
//    part of allocation. The write barrier shades both the
//    overwritten pointer and the new pointer value for any pointer
//    writes (see mbarrier.go for details). Newly allocated objects
//    are immediately marked black.
//
//    c. GC performs root marking jobs. This includes scanning all
//    stacks, shading all globals, and shading any heap pointers in
//    off-heap runtime data structures. Scanning a stack stops a
//    goroutine, shades any pointers found on its stack, and then
//    resumes the goroutine.
//
//    d. GC drains the work queue of grey objects, scanning each grey
//    object to black and shading all pointers found in the object
//    (which in turn may add those pointers to the work queue).
//
//    e. Because GC work is spread across local caches, GC uses a
//    distributed termination algorithm to detect when there are no
//    more root marking jobs or grey objects (see gcMarkDone). At this
//    point, GC transitions to mark termination.
//
// 3. GC performs mark termination.
//
//    a. Stop the world.
//
//    b. Set gcphase to _GCmarktermination, and disable workers and
//    assists.
//
//    c. Perform housekeeping like flushing mcaches.
//
// 4. GC performs the sweep phase.
//
//    a. Prepare for the sweep phase by setting gcphase to _GCoff,
//    setting up sweep state and disabling the write barrier.
//
//    b. Start the world. From this point on, newly allocated objects
//    are white, and allocating sweeps spans before use if necessary.
//
//    c. GC does concurrent sweeping in the background and in response
//    to allocation. See description below.
//
// 5. When sufficient allocation has taken place, replay the sequence
// starting with 1 above. See discussion of GC rate below.

// Concurrent sweep.
//
// The sweep phase proceeds concurrently with normal program execution.
// The heap is swept span-by-span both lazily (when a goroutine needs another span)
// and concurrently in a background goroutine (this helps programs that are not CPU bound).
// At the end of STW mark termination all spans are marked as "needs sweeping".
//
// The background sweeper goroutine simply sweeps spans one-by-one.
//
// To avoid requesting more OS memory while there are unswept spans, when a
// goroutine needs another span, it first attempts to reclaim that much memory
// by sweeping. When a goroutine needs to allocate a new small-object span, it
// sweeps small-object spans for the same object size until it frees at least
// one object. When a goroutine needs to allocate large-object span from heap,
// it sweeps spans until it frees at least that many pages into heap. There is
// one case where this may not suffice: if a goroutine sweeps and frees two
// nonadjacent one-page spans to the heap, it will allocate a new two-page
// span, but there can still be other one-page unswept spans which could be
// combined into a two-page span.
//
// It's critical to ensure that no operations proceed on unswept spans (that would corrupt
// mark bits in GC bitmap). During GC all mcaches are flushed into the central cache,
// so they are empty. When a goroutine grabs a new span into mcache, it sweeps it.
// When a goroutine explicitly frees an object or sets a finalizer, it ensures that
// the span is swept (either by sweeping it, or by waiting for the concurrent sweep to finish).
// The finalizer goroutine is kicked off only when all spans are swept.
// When the next GC starts, it sweeps all not-yet-swept spans (if any).

// GC rate.
// Next GC is after we've allocated an extra amount of memory proportional to
// the amount already in use. The proportion is controlled by GOGC environment variable
// (100 by default). If GOGC=100 and we're using 4M, we'll GC again when we get to 8M
// (this mark is tracked in next_gc variable). This keeps the GC cost in linear
// proportion to the allocation cost. Adjusting GOGC just changes the linear constant
// (and also the amount of extra memory used).

// Oblets
//
// In order to prevent long pauses while scanning large objects and to
// improve parallelism, the garbage collector breaks up scan jobs for
// objects larger than maxObletBytes into "oblets" of at most
// maxObletBytes. When scanning encounters the beginning of a large
// object, it scans only the first oblet and enqueues the remaining
// oblets as new scan jobs.

```

### 初始化

```
// Initialized from $GOGC.  GOGC=off means no GC.
var gcpercent int32

func gcinit() {
    if unsafe.Sizeof(workbuf{}) != _WorkbufSize {
        throw("size of Workbuf is suboptimal")
    }

    // No sweep on the first cycle.
    mheap_.sweepdone = 1

    // Set a reasonable initial GC trigger.
    memstats.triggerRatio = 7 / 8.0

    // Fake a heap_marked value so it looks like a trigger at
    // heapminimum is the appropriate growth from heap_marked.
    // This will go into computing the initial GC goal.
    memstats.heap_marked = uint64(float64(heapminimum) / (1 + memstats.triggerRatio))

    // Set gcpercent from the environment. This will also compute
    // and set the GC trigger and goal.
    _ = setGCPercent(readgogc())

    work.startSema = 1
    work.markDoneSema = 1
}

func readgogc() int32 {
    p := gogetenv("GOGC")
    if p == "off" {
        return -1
    }
    if n, ok := atoi32(p); ok {
        return n
    }
    return 100
}

// gcenable is called after the bulk of the runtime initialization,
// just before we're about to start letting user code run.
// It kicks off the background sweeper goroutine and enables GC.
func gcenable() {
    c := make(chan int, 1)
    go bgsweep(c)
    <-c
    memstats.enablegc = true // now that runtime is initialized, GC is okay
}

//go:linkname setGCPercent runtime/debug.setGCPercent
func setGCPercent(in int32) (out int32) {
    lock(&mheap_.lock)
    out = gcpercent
    if in < 0 {
        in = -1
    }
    gcpercent = in
    heapminimum = defaultHeapMinimum * uint64(gcpercent) / 100
    // Update pacing in response to gcpercent change.
    gcSetTriggerRatio(memstats.triggerRatio)
    unlock(&mheap_.lock)

    // If we just disabled GC, wait for any concurrent GC to
    // finish so we always return with no GC running.
    if in < 0 {
        // Disable phase transitions.
        lock(&work.sweepWaiters.lock)
        if gcphase == _GCmark {
            // GC is active. Wait until we reach sweeping.
            gp := getg()
            gp.schedlink = work.sweepWaiters.head
            work.sweepWaiters.head.set(gp)
            goparkunlock(&work.sweepWaiters.lock, "wait for GC cycle", traceEvGoBlock, 1)
        } else {
            // GC isn't active.
            unlock(&work.sweepWaiters.lock)
        }
    }

    return out
}
```

### gc的调用时机
整个gc流程的入口是gcStart,gcStart的调用方为:  

```
graph LR
mallocgc --> shouldhelpgc
shouldhelpgc --> |no|return
shouldhelpgc --> |yes|gc_condition_satisfied
gc_condition_satisfied --> |no|return
gc_condition_satisfied --> |yes|gcStart
runtime.GC --> gcStart
init --> forcegchelper
forcegchelper --> gcStart
```

其实就是`mallocgc,forcegchelper,runtime.GC` 这三个入口.  
* mallogc,分配堆内存时触发，
* forcegchelper,在forcegchelper中，会把forcegc.g这个全局对象的运行g挂起.  
sysmon会调用test检查上次触发gc的时间到当前时间是否已经经过了forcegcperiod长的时间,  
如果已经超过，那么就会将`forcegc.g`注入到globrunq.这样会在该g被调度到的时候触发gc.
* `runtime.GC`是由用户主动触发的,相当于强制触发gc.  

### gcTrigger和gc条件检查


