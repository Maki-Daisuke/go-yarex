package reaot

import (
	"strings"
	"unicode/utf8"
)

func MatchOpTree(op OpTree, s string) bool {
	if op.match(&matchContext{nil, 0, 0, s}, 0) != nil {
		return true
	}
	if _, ok := op.(*OpAssertBegin); ok {
		return false
	}
	minReq := op.minimumReq()
	for i := 1; minReq <= len(s)-i; i++ {
		if op.match(&matchContext{nil, 0, i, s}, i) != nil {
			return true
		}
	}
	return false
}

func (_ OpSuccess) match(c *matchContext, p int) *matchContext {
	return c
}

func (op *OpStr) match(c *matchContext, p int) *matchContext {
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

func (op *OpAlt) match(c *matchContext, p int) *matchContext {
	if r := op.follower.match(c, p); r != nil {
		return r
	}
	return op.alt.match(c, p)
}

func (op *OpClass) match(c *matchContext, p int) *matchContext {
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

func (op *OpNotNewLine) match(c *matchContext, p int) *matchContext {
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

func (op *OpCaptureStart) match(c *matchContext, p int) *matchContext {
	return op.follower.match(c.with(op.index, p), p)
}

func (op *OpCaptureEnd) match(c *matchContext, p int) *matchContext {
	return op.follower.match(c.with(op.index, p), p)
}

func (op *OpBackRef) match(c *matchContext, p int) *matchContext {
	s, ok := c.GetCaptured(op.index)
	if !ok || !strings.HasPrefix(c.str, s) {
		return nil
	}
	return op.follower.match(c, p+len(s))
}

func (op *OpAssertBegin) match(c *matchContext, p int) *matchContext {
	if p != 0 {
		return nil
	}
	return op.follower.match(c, p)
}

func (op *OpAssertEnd) match(c *matchContext, p int) *matchContext {
	if p != len(c.str) {
		return nil
	}
	return op.follower.match(c, p)
}
