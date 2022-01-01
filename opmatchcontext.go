package yarex

import (
	"sync"
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
	Str      *string                // string being matched
	getStack *func() []opStackFrame // Accessors to stack to record capturing positions.
	setStack *func([]opStackFrame)  // We use uintptr to avoid leaking param.
	stackTop int                    // stack top
}

func makeOpMatchContext(str *string, getter *func() []opStackFrame, setter *func([]opStackFrame)) MatchContext {
	return MatchContext{str, getter, setter, 0}
}

func (c MatchContext) Push(k ContextKey, p int) MatchContext {
	st := (*c.getStack)()
	sf := opStackFrame{k, p}
	if len(st) <= c.stackTop {
		st = append(st, sf)
		st = st[:cap(st)]
		(*c.setStack)(st)
	} else {
		st[c.stackTop] = sf
	}
	c.stackTop++
	return c
}

func (c MatchContext) GetCaptured(k ContextKey) (string, bool) {
	loc := c.GetCapturedIndex(k)
	if loc == nil {
		return "", false
	}
	return (*c.Str)[loc[0]:loc[1]], true
}

func (c MatchContext) GetCapturedIndex(k ContextKey) []int {
	var start, end int
	st := (*c.getStack)()
	i := c.stackTop - 1
	for ; ; i-- {
		if i == 0 {
			return nil
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
			return []int{start, end}
		}
	}
	// This should not happen.
	panic("Undetermined capture")
}

func (c MatchContext) FindVal(k ContextKey) int {
	st := (*c.getStack)()
	for i := c.stackTop - 1; i >= 0; i-- {
		if st[i].Key == k {
			return st[i].Pos
		}
	}
	return -1
}
