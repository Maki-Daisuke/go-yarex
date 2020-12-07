package reaot

import (
	"unsafe"
)

func MatchOpTree(op OpTree, s string) bool {
	rs := []rune(s)
	ome := opMatchEngine{func(_ *opMatchContext) {}}
	ctx := &opMatchContext{nil, rs, "c0", 0}
	if ome.exec(op, uintptr(unsafe.Pointer(ctx)), 0) != nil {
		return true
	}
	if _, ok := op.(*OpAssertBegin); ok {
		return false
	}
	minReq := op.minimumReq()
	for i := 1; minReq <= len(s)-i; i++ {
		ctx = &opMatchContext{nil, rs, "c0", i}
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
			c := ctx.with("c0", p)
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
			if !op.cls.Contains(str[p]) {
				return nil
			}
			next = op.follower
			p++
		case *OpNotNewLine:
			if len(str)-p < op.minReq {
				return nil
			}
			if str[p] == '\n' {
				return nil
			}
			next = op.follower
			p++
		case *OpCaptureStart:
			ctx2 := ctx.with(op.key, p)
			return ome.exec(op.follower, uintptr(unsafe.Pointer(&ctx2)), p)
		case *OpCaptureEnd:
			ctx2 := ctx.with(op.key, p)
			return ome.exec(op.follower, uintptr(unsafe.Pointer(&ctx2)), p)
		case *OpBackRef:
			s := ctx.GetCaptured(op.key)
			if s == nil { // When the corresponding capture group didn't match anything, this backref fails according to Perl's regex.
				return nil
			}
			if len(str)-p < len(s) {
				return nil
			}
			for i, r := range s {
				if str[p+i] != r {
					return nil
				}
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
