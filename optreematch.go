package reaot

import (
	"unicode/utf8"
)

func MatchOpTree(op OpTree, s string) bool {
	ome := newOpMatchEngine(func(_ *opMatchEngine) {})
	if ome.exec(op, s, 0) {
		return true
	}
	if _, ok := op.(*OpAssertBegin); ok {
		return false
	}
	minReq := op.minimumReq()
	for i := 1; minReq <= len(s)-i; i++ {
		if ome.exec(op, s, i) {
			return true
		}
	}
	return false
}

type opStackFrame struct {
	next OpTree
	pos  int
	key  string
	val  int
}

type opMatchEngine struct {
	onSuccess func(*opMatchEngine)
	stack     []opStackFrame
	stacktop  int
}

func newOpMatchEngine(onSuccess func(*opMatchEngine)) *opMatchEngine {
	return &opMatchEngine{
		onSuccess: onSuccess,
		stack:     make([]opStackFrame, 512),
		stacktop:  0,
	}
}

func (ome *opMatchEngine) push(f opStackFrame) {
	if ome.stacktop >= len(ome.stack) {
		ome.stack = append(ome.stack, make([]opStackFrame, len(ome.stack))...)
	}
	ome.stack[ome.stacktop] = f
	ome.stacktop++
}

func (ome *opMatchEngine) pop() (OpTree, int) {
	for {
		if ome.stacktop == 0 {
			return nil, 0
		}
		ome.stacktop--
		f := ome.stack[ome.stacktop]
		if f.next != nil {
			return f.next, f.pos
		}
	}
	panic("SHOULD NOT REACH HERE")
}

func (ome *opMatchEngine) findVal(k string) int {
	for i := ome.stacktop - 1; i >= 0; i-- {
		if ome.stack[i].key == k {
			return ome.stack[i].val
		}
	}
	return -1
}

func (ome opMatchEngine) exec(next OpTree, str string, p int) bool {
LOOP:
	for {
		switch op := next.(type) {
		case OpSuccess:
			return true
		case *OpStr:
			if len(str)-p < op.minReq {
				next, p = ome.pop()
				continue LOOP
			}
			for i := 0; i < len(op.str); i++ {
				if str[p+i] != op.str[i] {
					next, p = ome.pop()
					continue LOOP
				}
			}
			next = op.follower
			p += len(op.str)
		case *OpAlt:
			ome.push(opStackFrame{op.alt, p, "", 0})
			next = op.follower
		case *OpRepeat:
			prev := ome.findVal(op.key)
			if prev == p { // This means zero-width matching occurs.
				next = op.alt // So, terminate repeating.
				continue LOOP
			}
			ome.push(opStackFrame{op.alt, p, op.key, p})
			next = op.follower
		case *OpClass:
			if len(str)-p < op.minReq {
				next, p = ome.pop()
				continue LOOP
			}
			r, size := utf8.DecodeRuneInString(str[p:])
			if size == 0 || r == utf8.RuneError {
				next, p = ome.pop()
				continue LOOP
			}
			if !op.cls.Contains(r) {
				next, p = ome.pop()
				continue LOOP
			}
			next = op.follower
			p += size
		case *OpNotNewLine:
			if len(str)-p < op.minReq {
				next, p = ome.pop()
				continue LOOP
			}
			r, size := utf8.DecodeRuneInString(str[p:])
			if size == 0 || r == utf8.RuneError {
				next, p = ome.pop()
				continue LOOP
			}
			if r == '\n' {
				next, p = ome.pop()
				continue LOOP
			}
			next = op.follower
			p += size
		case *OpCaptureStart:
			ome.push(opStackFrame{nil, 0, op.key, p})
			next = op.follower
		case *OpCaptureEnd:
			ome.push(opStackFrame{nil, 0, op.key, p})
			next = op.follower
		case *OpBackRef:
			//s, ok := ctx.GetCaptured(op.key)
			//if !ok || !strings.HasPrefix(str[p:], s) {
			//	return nil
			//}
			//next = op.follower
			//p += len(s)
			// Not imlemented yet
			next, p = ome.pop()
		case *OpAssertBegin:
			if p != 0 {
				next, p = ome.pop()
				continue LOOP
			}
			next = op.follower
		case *OpAssertEnd:
			if p != len(str) {
				next, p = ome.pop()
				continue LOOP
			}
			next = op.follower
		default:
			return false
		}
	}
}
