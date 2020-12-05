package reaot

import (
	"strings"
	"unicode/utf8"
)

func Match(re Regexp, s string) bool {
	if re.match(matchContext{nil, 0, 0, s}, 0, func(c matchContext, _ int) *matchContext { return &c }) != nil {
		return true
	}
	if canOnlyMatchAtBegining(re) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if re.match(matchContext{nil, 0, i, s}, i, func(c matchContext, _ int) *matchContext { return &c }) != nil {
			return true
		}
	}
	return false
}

func (re ReLit) match(c matchContext, p int, k Continuation) *matchContext {
	lit := string(re)
	if !strings.HasPrefix(c.str[p:], lit) {
		return nil
	}
	return k(c, p+len(lit))
}

func (re ReNotNewline) match(c matchContext, p int, k Continuation) *matchContext {
	if !(p < len(c.str)) || c.str[0] == '\n' {
		return nil
	}
	return k(c, p+1)
}

func (r *ReSeq) match(c matchContext, p int, k Continuation) *matchContext {
	seq := r.seq
	var loop func(int) Continuation
	loop = func(i int) func(c matchContext, p int) *matchContext {
		return func(c matchContext, p int) *matchContext {
			if i < len(seq) {
				return seq[i].match(c, p, loop(i+1))
			}
			return k(c, p)
		}
	}
	return loop(0)(c, p)
}

func (r *ReAlt) match(c matchContext, p int, k Continuation) *matchContext {
	for _, re := range r.opts {
		if c1 := re.match(c, p, k); c1 != nil {
			return c1
		}
	}
	return nil
}

func (r *ReRepeat) match(c matchContext, p int, k Continuation) *matchContext {
	switch re := r.re.(type) {
	case ReLit:
		s := string(re)
		width := len(s)
		if width == 0 {
			return k(c, p)
		}
		p1 := p
		i := 0
		if r.max < 0 {
			for strings.HasPrefix(c.str[p1:], s) {
				i++
				p1 += width
			}
		} else {
			for i < r.max && strings.HasPrefix(c.str[p1:], s) {
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
	case ReCharClass:
		cc := re.CharClass
		stack := make([]int, 0, 64)
		stack = append(stack, p)
		p1 := p
		i := 0
		if r.max < 0 {
			for p1 < len(c.str) {
				r, size := utf8.DecodeRuneInString(c.str[p1:])
				if cc.Contains(r) {
					p1 += size
					i++
					stack = append(stack, p1)
				} else {
					break
				}
			}
		} else {
			for i < r.max && p1 < len(c.str) {
				r, size := utf8.DecodeRuneInString(c.str[p1:])
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
			return func(c matchContext, p int) *matchContext {
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

func (r *ReCap) match(c matchContext, p int, k Continuation) *matchContext {
	return r.re.match(c.with(r.index, p), p, func(c matchContext, p1 int) *matchContext {
		return k((&c).with(r.index, p1), p1)
	})
}

func (r ReBackRef) match(c matchContext, p int, k Continuation) *matchContext {
	cap, ok := (&c).GetCaptured(uint(r))
	if !ok {
		return nil
	}
	return ReLit(cap).match(c, p, k)
}

func (re ReAssertBegin) match(c matchContext, p int, k Continuation) *matchContext {
	if p != 0 {
		return nil
	}
	return k(c, p)
}

func (re ReAssertEnd) match(c matchContext, p int, k Continuation) *matchContext {
	if p != len(c.str) {
		return nil
	}
	return k(c, p)
}

func (re ReCharClass) match(c matchContext, p int, k Continuation) *matchContext {
	if len(c.str) < p+1 {
		return nil
	}
	r, size := utf8.DecodeRuneInString(c.str[p:])
	if !re.Contains(r) {
		return nil
	}
	return k(c, p+size)
}
