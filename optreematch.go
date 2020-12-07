package reaot

import (
	"strings"
	"unicode/utf8"
	"unsafe"
)

func MatchOpTree(op OpTree, s string) bool {
	ctx := &opMatchContext{nil, s, "c0", 0}
	if op.match(uintptr(unsafe.Pointer(ctx)), 0) != nil {
		return true
	}
	if _, ok := op.(*OpAssertBegin); ok {
		return false
	}
	minReq := op.minimumReq()
	for i := 1; minReq <= len(s)-i; i++ {
		ctx = &opMatchContext{nil, s, "c0", i}
		if op.match(uintptr(unsafe.Pointer(ctx)), i) != nil {
			return true
		}
	}
	return false
}

func (_ OpSuccess) match(c uintptr, p int) *opMatchContext {
	return (*opMatchContext)(unsafe.Pointer(c)).with("c0", p)
}

func (op *OpStr) match(c uintptr, p int) *opMatchContext {
	ctx := (*opMatchContext)(unsafe.Pointer(c))
	str := ctx.str
	if len(str)-p < op.minReq {
		return nil
	}
	for i := 0; i < len(op.str); i++ {
		if str[p+i] != op.str[i] {
			return nil
		}
	}
	return op.follower.match(c, p+len(op.str))
}

func (op *OpAlt) match(c uintptr, p int) *opMatchContext {
	if r := op.follower.match(c, p); r != nil {
		return r
	}
	return op.alt.match(c, p)
}

func (op *OpRepeat) match(c uintptr, p int) *opMatchContext {
	ctx := (*opMatchContext)(unsafe.Pointer(c))
	prev := ctx.findVal(op.key)
	if prev == p { // This means zero-width matching occurs.
		return op.alt.match(c, p) // So, terminate repeating.
	}
	ctx2 := ctx.with(op.key, p)
	if r := op.follower.match(uintptr(unsafe.Pointer(ctx2)), p); r != nil {
		return r
	}
	return op.alt.match(c, p)
}

func (op *OpClass) match(c uintptr, p int) *opMatchContext {
	ctx := (*opMatchContext)(unsafe.Pointer(c))
	str := ctx.str
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
	return op.follower.match(c, p+size)
}

func (op *OpNotNewLine) match(c uintptr, p int) *opMatchContext {
	ctx := (*opMatchContext)(unsafe.Pointer(c))
	str := ctx.str
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
	return op.follower.match(c, p+size)
}

func (op *OpCaptureStart) match(c uintptr, p int) *opMatchContext {
	ctx := (*opMatchContext)(unsafe.Pointer(c))
	ctx = ctx.with(op.key, p)
	return op.follower.match(uintptr(unsafe.Pointer(ctx)), p)
}

func (op *OpCaptureEnd) match(c uintptr, p int) *opMatchContext {
	ctx := (*opMatchContext)(unsafe.Pointer(c))
	ctx = ctx.with(op.key, p)
	return op.follower.match(uintptr(unsafe.Pointer(ctx)), p)
}

func (op *OpBackRef) match(c uintptr, p int) *opMatchContext {
	ctx := (*opMatchContext)(unsafe.Pointer(c))
	s, ok := ctx.GetCaptured(op.key)
	if !ok || !strings.HasPrefix(ctx.str[p:], s) {
		return nil
	}
	return op.follower.match(c, p+len(s))
}

func (op *OpAssertBegin) match(c uintptr, p int) *opMatchContext {
	if p != 0 {
		return nil
	}
	return op.follower.match(c, p)
}

func (op *OpAssertEnd) match(c uintptr, p int) *opMatchContext {
	ctx := (*opMatchContext)(unsafe.Pointer(c))
	if p != len(ctx.str) {
		return nil
	}
	return op.follower.match(c, p)
}
