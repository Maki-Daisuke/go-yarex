package yarex

import (
	"runtime"
	"strings"
	"sync/atomic"
	"unicode/utf8"
	"unsafe"
)

var maxConcurrency uint32 = uint32(runtime.NumCPU()) - 1
var availableThreads uint32 = 0

type task struct {
	ome    opMatchEngine
	cancel uintptr
	op     OpTree
	ctx    uintptr
	pos    int
	res    chan<- *opMatchContext
}

var taskChan = make(chan task, 1)

func init() {
	// Prepare worker threads shared in the process.
	for i := uint32(0); i < maxConcurrency; i++ {
		atomic.AddUint32(&availableThreads, 1)
		go func() {
			for {
				t := <-taskChan
				t.res <- t.ome.exec(t.cancel, t.op, t.ctx, t.pos)
				atomic.AddUint32(&availableThreads, 1)
			}
		}()
	}
}

func tryAcquireThread() bool {
	n := atomic.LoadUint32(&availableThreads)
	if n == 0 {
		return false
	}
	return atomic.CompareAndSwapUint32(&availableThreads, n, n-1)
}

func matchOpTree(op OpTree, s string) bool {
	ome := opMatchEngine{func(_ *opMatchContext) {}}
	ctx := &opMatchContext{nil, s, contextKey{'c', 0}, 0}
	var cancel uint32 = 0
	if ome.exec(uintptr(unsafe.Pointer(&cancel)), op, uintptr(unsafe.Pointer(ctx)), 0) != nil {
		return true
	}
	if _, ok := op.(*OpAssertBegin); ok {
		return false
	}
	minReq := op.minimumReq()
	for i := 1; minReq <= len(s)-i; i++ {
		ctx = &opMatchContext{nil, s, contextKey{'c', 0}, i}
		if ome.exec(uintptr(unsafe.Pointer(&cancel)), op, uintptr(unsafe.Pointer(ctx)), i) != nil {
			return true
		}
	}
	return false
}

type opMatchEngine struct {
	onSuccess func(*opMatchContext)
}

func isCancelled(c uintptr) bool {
	p := (*uint32)(unsafe.Pointer(c))
	return atomic.LoadUint32(p) != 0
}

func (ome opMatchEngine) exec(cancelled uintptr, next OpTree, c uintptr, p int) *opMatchContext {
	ctx := (*opMatchContext)(unsafe.Pointer(c))
	str := ctx.str
	for {
		if isCancelled(cancelled) {
			return nil
		}
		switch op := next.(type) {
		case OpSuccess:
			c := ctx.with(contextKey{'c', 0}, p)
			ome.onSuccess(&c)
			return &c
		case *OpStr:
			if len(str)-p < op.minReq {
				return nil
			}
			for i := 0; i < len(op.str); i++ {
				if str[p+i] != op.str[i] {
					return nil
				}
			}
			next = op.follower
			p += len(op.str)
		case *OpAlt:
			if tryAcquireThread() {
				// If there is an available thread (goroutine), speculatively execute opt.alt in parallel
				var cancel uint32 = 0
				res := make(chan *opMatchContext, 1)
				taskChan <- task{ome, uintptr(unsafe.Pointer(&cancel)), op.alt, c, p, res}
				if r := ome.exec(cancelled, op.follower, c, p); r != nil {
					atomic.StoreUint32(&cancel, 1) // Notify cancel to the speculative goroutine
					<-res                          // Wait for the goroutine completes
					// Here, if we don't wait for the groutine competes, the following return-statement rewind the call-stack.
					// That will break data on the stack, which is still in use by the groutine.
					return r
				} else {
					if isCancelled(cancelled) {
						// If the task was cancelled by the parent task, notify cancel to the child
						atomic.StoreUint32(&cancel, 1)
					}
					// Wait for the result from the goroutine
					return <-res
				}
			} else {
				// Execute sequentially
				if r := ome.exec(cancelled, op.follower, c, p); r != nil {
					return r
				}
				next = op.alt
			}
		case *OpRepeat:
			prev := ctx.findVal(op.key)
			if prev == p { // This means zero-width matching occurs.
				next = op.alt // So, terminate repeating.
				continue
			}
			if tryAcquireThread() {
				// If there is a available thread (goroutine), speculatively execute opt.alt in parallel
				var cancel uint32 = 0
				res := make(chan *opMatchContext, 1)
				taskChan <- task{ome, uintptr(unsafe.Pointer(&cancel)), op.alt, c, p, res}
				ctx2 := ctx.with(op.key, p)
				if r := ome.exec(cancelled, op.follower, uintptr(unsafe.Pointer(&ctx2)), p); r != nil {
					atomic.StoreUint32(&cancel, 1) // Notify cancel to the speculative goroutine
					<-res                          // Wait for the goroutine completes
					// Here, if we don't wait for the groutine competes, the following return-statement rewind the call-stack.
					// That will break data on the stack, which is still in use by the groutine.
					return r
				} else {
					if isCancelled(cancelled) {
						// If the task was cancelled by the parent task, notify cancel to the child
						atomic.StoreUint32(&cancel, 1)
					}
					// Wait for the result from the goroutine
					return <-res
				}
			} else {
				// Execute sequentially
				ctx2 := ctx.with(op.key, p)
				if r := ome.exec(cancelled, op.follower, uintptr(unsafe.Pointer(&ctx2)), p); r != nil {
					return r
				}
				next = op.alt
			}
		case *OpClass:
			if len(str)-p < op.minReq {
				return nil
			}
			r, size := utf8.DecodeRuneInString(str[p:])
			if size == 0 || r == utf8.RuneError {
				return nil
			}
			if !op.cls.Contains(r) {
				return nil
			}
			next = op.follower
			p += size
		case *OpNotNewLine:
			if len(str)-p < op.minReq {
				return nil
			}
			r, size := utf8.DecodeRuneInString(str[p:])
			if size == 0 || r == utf8.RuneError {
				return nil
			}
			if r == '\n' {
				return nil
			}
			next = op.follower
			p += size
		case *OpCaptureStart:
			ctx2 := ctx.with(op.key, p)
			return ome.exec(cancelled, op.follower, uintptr(unsafe.Pointer(&ctx2)), p)
		case *OpCaptureEnd:
			ctx2 := ctx.with(op.key, p)
			return ome.exec(cancelled, op.follower, uintptr(unsafe.Pointer(&ctx2)), p)
		case *OpBackRef:
			s, ok := ctx.GetCaptured(op.key)
			if !ok || !strings.HasPrefix(str[p:], s) {
				return nil
			}
			next = op.follower
			p += len(s)
		case *OpAssertBegin:
			if p != 0 {
				return nil
			}
			next = op.follower
		case *OpAssertEnd:
			if p != len(str) {
				return nil
			}
			next = op.follower
		}
	}
}
