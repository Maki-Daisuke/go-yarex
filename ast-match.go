package reaot

import (
	"strings"
	"unicode/utf8"
	"unsafe"
)

func Match(re Ast, s string) bool {
	c := matchContext{nil, 0, 0, s}
	if re.match(uintptr(unsafe.Pointer(&c)), 0, func(c uintptr, _ int) *matchContext { return (*matchContext)(unsafe.Pointer(c)) }) != nil {
		return true
	}
	if canOnlyMatchAtBegining(re) {
		return false
	}
	for i := 1; i < len(s); i++ {
		c.pos = i
		if re.match(uintptr(unsafe.Pointer(&c)), i, func(c uintptr, _ int) *matchContext { return (*matchContext)(unsafe.Pointer(c)) }) != nil {
			return true
		}
	}
	return false
}

func (re AstLit) match(c uintptr, p int, k Continuation) *matchContext {
	str := (*matchContext)(unsafe.Pointer(c)).str
	lit := string(re)
	if !strings.HasPrefix(str[p:], lit) {
		return nil
	}
	return k(c, p+len(lit))
}

func (re AstNotNewline) match(c uintptr, p int, k Continuation) *matchContext {
	str := (*matchContext)(unsafe.Pointer(c)).str
	if !(p < len(str)) || str[0] == '\n' {
		return nil
	}
	return k(c, p+1)
}

func (r *AstSeq) match(c uintptr, p int, k Continuation) *matchContext {
	seq := r.seq
	var loop func(int) Continuation
	loop = func(i int) func(c uintptr, p int) *matchContext {
		return func(c uintptr, p int) *matchContext {
			if i < len(seq) {
				return seq[i].match(c, p, loop(i+1))
			}
			return k(c, p)
		}
	}
	return loop(0)(c, p)
}

func (r *AstAlt) match(c uintptr, p int, k Continuation) *matchContext {
	for _, re := range r.opts {
		if c1 := re.match(c, p, k); c1 != nil {
			return c1
		}
	}
	return nil
}

func (r *AstRepeat) match(c uintptr, p int, k Continuation) *matchContext {
	str := (*matchContext)(unsafe.Pointer(c)).str
	switch re := r.re.(type) {
	case AstLit:
		s := string(re)
		width := len(s)
		if width == 0 {
			return k(c, p)
		}
		p1 := p
		i := 0
		if r.max < 0 {
			for strings.HasPrefix(str[p1:], s) {
				i++
				p1 += width
			}
		} else {
			for i < r.max && strings.HasPrefix(str[p1:], s) {
				i++
				p1 += width
			}
		}
		for i >= r.min { // Try backtrack
			if ret := k(c, p1); ret != nil {
				return ret
			}
			p1 -= width
			i--
		}
		return nil
	case AstCharClass:
		cc := re.CharClass
		stack := make([]int, 0, 64)
		stack = append(stack, p)
		p1 := p
		i := 0
		if r.max < 0 {
			for p1 < len(str) {
				r, size := utf8.DecodeRuneInString(str[p1:])
				if cc.Contains(r) {
					p1 += size
					i++
					stack = append(stack, p1)
				} else {
					break
				}
			}
		} else {
			for i < r.max && p1 < len(str) {
				r, size := utf8.DecodeRuneInString(str[p1:])
				if cc.Contains(r) {
					p1 += size
					i++
					stack = append(stack, p1)
				} else {
					break
				}
			}
		}
		for i >= r.min { // Try backtrack
			if ret := k(c, stack[i]); ret != nil {
				return ret
			}
			i--
		}
		return nil
	default:
		prev := -1 // initial value must be a number which never equal to any position (i.e. positive integer)
		var loop func(count int) Continuation
		loop = func(count int) Continuation {
			return func(c uintptr, p int) *matchContext {
				if prev == p { // Matched zero-length assertion. So, move ahead the next pattern.
					return k(c, p)
				}
				prev = p
				if count < r.min {
					return re.match(c, p, loop(count+1))
				}
				if count == r.max {
					return k(c, p)
				}
				c1 := re.match(c, p, loop(count+1))
				if c1 != nil {
					return c1
				}
				return k(c, p)
			}
		}
		return loop(0)(c, p)
	}
}

func (r *AstCap) match(c uintptr, p int, k Continuation) *matchContext {
	ctx := (*matchContext)(unsafe.Pointer(c))
	ctx = ctx.with(r.index, p)
	return r.re.match(uintptr(unsafe.Pointer(ctx)), p, func(c uintptr, p1 int) *matchContext {
		ctx := (*matchContext)(unsafe.Pointer(c))
		ctx = ctx.with(r.index, p1)
		return k(uintptr(unsafe.Pointer(ctx)), p1)
	})
}

func (r AstBackRef) match(c uintptr, p int, k Continuation) *matchContext {
	ctx := (*matchContext)(unsafe.Pointer(c))
	cap, ok := (ctx).GetCaptured(uint(r))
	if !ok {
		return nil
	}
	return AstLit(cap).match(c, p, k)
}

func (re AstAssertBegin) match(c uintptr, p int, k Continuation) *matchContext {
	if p != 0 {
		return nil
	}
	return k(c, p)
}

func (re AstAssertEnd) match(c uintptr, p int, k Continuation) *matchContext {
	str := (*matchContext)(unsafe.Pointer(c)).str
	if p != len(str) {
		return nil
	}
	return k(c, p)
}

func (re AstCharClass) match(c uintptr, p int, k Continuation) *matchContext {
	str := (*matchContext)(unsafe.Pointer(c)).str
	if len(str) < p+1 {
		return nil
	}
	r, size := utf8.DecodeRuneInString(str[p:])
	if !re.Contains(r) {
		return nil
	}
	return k(c, p+size)
}
