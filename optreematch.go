package reaot

import (
	"strings"
	"unicode/utf8"
)

func MatchOpTree(op OpTree, s string) bool {
	if op.match(&opMatchContext{nil, s, "c0", 0}, 0) != nil {
		return true
	}
	if _, ok := op.(*OpAssertBegin); ok {
		return false
	}
	minReq := op.minimumReq()
	for i := 1; minReq <= len(s)-i; i++ {
		if op.match(&opMatchContext{nil, s, "c0", i}, i) != nil {
			return true
		}
	}
	return false
}

func (_ OpSuccess) match(c *opMatchContext, p int) *opMatchContext {
	return c.with("c0", p)
}

func (op *OpStr) match(c *opMatchContext, p int) *opMatchContext {
	str := c.str
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

func (op *OpAlt) match(c *opMatchContext, p int) *opMatchContext {
	if r := op.follower.match(c, p); r != nil {
		return r
	}
	return op.alt.match(c, p)
}

func (op *OpRepeat) match(c *opMatchContext, p int) *opMatchContext {
	prev := c.findVal(op.key)
	if prev == p { // This means zero-width matching occurs.
		return op.alt.match(c, p) // So, terminate repeating.
	}
	if r := op.follower.match(c.with(op.key, p), p); r != nil {
		return r
	}
	return op.alt.match(c, p)
}

func (op *OpClass) match(c *opMatchContext, p int) *opMatchContext {
	str := c.str
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

func (op *OpNotNewLine) match(c *opMatchContext, p int) *opMatchContext {
	str := c.str
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

func (op *OpCaptureStart) match(c *opMatchContext, p int) *opMatchContext {
	return op.follower.match(c.with(op.key, p), p)
}

func (op *OpCaptureEnd) match(c *opMatchContext, p int) *opMatchContext {
	return op.follower.match(c.with(op.key, p), p)
}

func (op *OpBackRef) match(c *opMatchContext, p int) *opMatchContext {
	s, ok := c.GetCaptured(op.key)
	if !ok || !strings.HasPrefix(c.str[p:], s) {
		return nil
	}
	return op.follower.match(c, p+len(s))
}

func (op *OpAssertBegin) match(c *opMatchContext, p int) *opMatchContext {
	if p != 0 {
		return nil
	}
	return op.follower.match(c, p)
}

func (op *OpAssertEnd) match(c *opMatchContext, p int) *opMatchContext {
	if p != len(c.str) {
		return nil
	}
	return op.follower.match(c, p)
}
