package yarex

import (
	"strings"
	"unicode/utf8"
	"unsafe"
)

type opExecer struct {
	op OpTree
}

func (oe opExecer) exec(str string, pos int, onSuccess func(*MatchContext)) bool {
	op := oe.op
	_, headOnly := op.(*OpAssertBegin)
	if headOnly && pos != 0 {
		return false
	}
	minReq := op.minimumReq()
	if minReq > len(str)-pos {
		return false
	}
	ctx := MatchContext{nil, str, ContextKey{'c', 0}, pos}
	if opTreeExec(op, uintptr(unsafe.Pointer(&ctx)), pos, onSuccess) {
		return true
	}
	if headOnly {
		return false
	}
	for i := pos + 1; minReq <= len(str)-i; i++ {
		ctx := MatchContext{nil, str, ContextKey{'c', 0}, i}
		if opTreeExec(op, uintptr(unsafe.Pointer(&ctx)), i, onSuccess) {
			return true
		}
	}
	return false
}

func opTreeExec(next OpTree, c uintptr, p int, onSuccess func(*MatchContext)) bool {
	ctx := (*MatchContext)(unsafe.Pointer(c))
	str := ctx.Str
	for {
		switch op := next.(type) {
		case OpSuccess:
			c := ctx.With(ContextKey{'c', 0}, p)
			onSuccess(&c)
			return true
		case *OpStr:
			if len(str)-p < op.minReq {
				return false
			}
			for i := 0; i < len(op.str); i++ {
				if str[p+i] != op.str[i] {
					return false
				}
			}
			next = op.follower
			p += len(op.str)
		case *OpAlt:
			if opTreeExec(op.follower, c, p, onSuccess) {
				return true
			}
			next = op.alt
		case *OpRepeat:
			prev := ctx.FindVal(op.key)
			if prev == p { // This means zero-width matching occurs.
				next = op.alt // So, terminate repeating.
				continue
			}
			ctx2 := ctx.With(op.key, p)
			if opTreeExec(op.follower, uintptr(unsafe.Pointer(&ctx2)), p, onSuccess) {
				return true
			}
			next = op.alt
		case *OpClass:
			if len(str)-p < op.minReq {
				return false
			}
			r, size := utf8.DecodeRuneInString(str[p:])
			if size == 0 || r == utf8.RuneError {
				return false
			}
			if !op.cls.Contains(r) {
				return false
			}
			next = op.follower
			p += size
		case *OpNotNewLine:
			if len(str)-p < op.minReq {
				return false
			}
			r, size := utf8.DecodeRuneInString(str[p:])
			if size == 0 || r == utf8.RuneError {
				return false
			}
			if r == '\n' {
				return false
			}
			next = op.follower
			p += size
		case *OpCaptureStart:
			ctx2 := ctx.With(op.key, p)
			return opTreeExec(op.follower, uintptr(unsafe.Pointer(&ctx2)), p, onSuccess)
		case *OpCaptureEnd:
			ctx2 := ctx.With(op.key, p)
			return opTreeExec(op.follower, uintptr(unsafe.Pointer(&ctx2)), p, onSuccess)
		case *OpBackRef:
			s, ok := ctx.GetCaptured(op.key)
			if !ok || !strings.HasPrefix(str[p:], s) {
				return false
			}
			next = op.follower
			p += len(s)
		case *OpAssertBegin:
			if p != 0 {
				return false
			}
			next = op.follower
		case *OpAssertEnd:
			if p != len(str) {
				return false
			}
			next = op.follower
		}
	}
}
