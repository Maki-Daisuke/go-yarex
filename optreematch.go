package yarex

import (
	"strings"
	"unicode/utf8"
	"unsafe"
)

func matchOpTree(op OpTree, s string) bool {
	ome := opMatchEngine{func(_ *opMatchContext) {}}
	ctx := &opMatchContext{nil, s, contextKey{'c', 0}, 0}
	if ome.exec(op, uintptr(unsafe.Pointer(ctx)), 0) != nil {
		return true
	}
	if _, ok := op.(*OpAssertBegin); ok {
		return false
	}
	minReq := op.minimumReq()
	for i := 1; minReq <= len(s)-i; i++ {
		ctx = &opMatchContext{nil, s, contextKey{'c', 0}, i}
		if ome.exec(op, uintptr(unsafe.Pointer(ctx)), i) != nil {
			return true
		}
	}
	return false
}

type opMatchEngine struct {
	onSuccess func(*opMatchContext)
}

func (ome opMatchEngine) exec(next OpTree, c uintptr, p int) *opMatchContext {
	ctx := (*opMatchContext)(unsafe.Pointer(c))
	str := ctx.str
	for {
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
			if r := ome.exec(op.follower, c, p); r != nil {
				return r
			}
			next = op.alt
		case *OpRepeat:
			prev := ctx.findVal(op.key)
			if prev == p { // This means zero-width matching occurs.
				next = op.alt // So, terminate repeating.
				continue
			}
			ctx2 := ctx.with(op.key, p)
			if r := ome.exec(op.follower, uintptr(unsafe.Pointer(&ctx2)), p); r != nil {
				return r
			}
			next = op.alt
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
			return ome.exec(op.follower, uintptr(unsafe.Pointer(&ctx2)), p)
		case *OpCaptureEnd:
			ctx2 := ctx.with(op.key, p)
			return ome.exec(op.follower, uintptr(unsafe.Pointer(&ctx2)), p)
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
