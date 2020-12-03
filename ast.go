package reaot

import (
	"bytes"
	"fmt"
)

type ReLit []rune

func (re ReLit) String() string {
	return string(re)
}

type ReSeq struct {
	seq []Regexp
}

func (re *ReSeq) String() string {
	b := bytes.NewBufferString("(?:")
	for _, r := range re.seq {
		fmt.Fprint(b, r.String())
	}
	fmt.Fprint(b, ")")
	return b.String()
}

type ReAlt struct {
	opts []Regexp
}

func (re *ReAlt) String() string {
	b := bytes.NewBufferString("(?:")
	fmt.Fprint(b, re.opts[0].String())
	for _, r := range re.opts[1:] {
		fmt.Fprint(b, "|")
		fmt.Fprint(b, r.String())
	}
	fmt.Fprint(b, ")")
	return b.String()
}

type ReNotNewline struct{}

func (re ReNotNewline) String() string {
	return "."
}

type ReRepeat struct {
	re       Regexp
	min, max int // -1 means unlimited
}

func (re *ReRepeat) String() string {
	if re.min == 0 && re.max == 1 {
		return re.re.String() + "?"
	}
	if re.min == 0 && re.max < 0 {
		return re.re.String() + "*"
	}
	if re.min == 1 && re.max < 0 {
		return re.re.String() + "+"
	}
	if re.min == re.max {
		return fmt.Sprintf("%s{%d}", re.re.String(), re.min)
	}
	if re.max < 0 {
		return fmt.Sprintf("%s{%d,}", re.re.String(), re.min)
	}
	return fmt.Sprintf("%s{%d,%d}", re.re.String(), re.min, re.max)
}

type ReCap struct {
	index uint
	re    Regexp
}

func (re *ReCap) String() string {
	return fmt.Sprintf("(%s)", re.re)
}

type ReBackRef uint

func (re ReBackRef) String() string {
	return fmt.Sprintf("\\%d", uint(re))
}

type ReAssertBegin struct{}

func (re ReAssertBegin) String() string {
	return "^"
}

type ReAssertEnd struct{}

func (re ReAssertEnd) String() string {
	return "$"
}

type ReCharClass struct {
	CharClass
}

func (re ReCharClass) String() string {
	return "[" + re.CharClass.String() + "]"
}
