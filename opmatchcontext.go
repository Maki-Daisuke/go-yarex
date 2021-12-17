package yarex

import (
	"sync"
	"unsafe"
)

type ContextKey struct {
	Kind  rune
	Index uint
}

type opStackFrame struct {
	Key ContextKey
	Pos int
}

var opStackPool = sync.Pool{
	New: func() interface{} {
		b := make([]opStackFrame, initialStackSize)
		return &b
	},
}

type MatchContext struct {
	Str      uintptr // *string                // string being matched
	getStack uintptr // *func() []opStackFrame // Accessors to stack to record capturing positions.
	setStack uintptr // *func([]opStackFrame)  // We use uintptr to avoid leaking param.
	stackTop int     // stack top
}

func makeOpMatchContext(str *string, getter *func() []opStackFrame, setter *func([]opStackFrame)) MatchContext {
	return MatchContext{uintptr(unsafe.Pointer(str)), uintptr(unsafe.Pointer(getter)), uintptr(unsafe.Pointer(setter)), 0}
}

func (c MatchContext) Push(k ContextKey, p int) MatchContext {
	st := (*(*func() []opStackFrame)(unsafe.Pointer(c.getStack)))() // c.getStack()
	sf := opStackFrame{k, p}
	if len(st) <= c.stackTop {
		st = append(st, sf)
		st = st[:cap(st)]
		(*(*func([]opStackFrame))(unsafe.Pointer(c.setStack)))(st) // c.setStack(st)
	} else {
		st[c.stackTop] = sf
	}
	c.stackTop++
	return c
}

func (c MatchContext) GetCaptured(k ContextKey) (string, bool) {
	var start, end int
	st := (*(*func() []opStackFrame)(unsafe.Pointer(c.getStack)))() // c.getStack()
	i := c.stackTop - 1
	for ; ; i-- {
		if i == 0 {
			return "", false
		}
		if st[i].Key == k {
			end = st[i].Pos
			break
		}
	}
	i--
	for ; i >= 0; i-- {
		if st[i].Key == k {
			start = st[i].Pos
			str := *(*string)(unsafe.Pointer(c.Str))
			return str[start:end], true
		}
	}
	// This should not happen.
	panic("Undetermined capture")
}

func (c MatchContext) FindVal(k ContextKey) int {
	st := (*(*func() []opStackFrame)(unsafe.Pointer(c.getStack)))() // c.getStack()
	for i := c.stackTop - 1; i >= 0; i-- {
		if st[i].Key == k {
			return st[i].Pos
		}
	}
	return -1
}
