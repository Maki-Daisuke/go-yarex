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

type ReZeroOrMore struct {
	re Regexp
}

func (re *ReZeroOrMore) String() string {
	return re.re.String() + "*"
}

type ReOneOrMore struct {
	re Regexp
}

func (re *ReOneOrMore) String() string {
	return re.re.String() + "+"
}

type ReOpt struct {
	re Regexp
}

func (re *ReOpt) String() string {
	return re.re.String() + "?"
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

type ReCharClass struct {
	CharClass
}

func (re ReCharClass) String() string {
	return "[" + re.CharClass.String() + "]"
}
