package yarex

import (
	"bytes"
	"fmt"
)

// Here, we use uintpointer to pass *matchContext
// to avoid from allocating the parameter in heap
type Continuation = func(uintptr, int) *matchContext

type Ast interface {
	//Compile()
	String() string
	match(uintptr, int, Continuation) *matchContext // This implements an interpreter implementation.
}

type AstLit string

func (re AstLit) String() string {
	return string(re)
}

type AstSeq struct {
	seq []Ast
}

func (re *AstSeq) String() string {
	b := bytes.NewBufferString("(?:")
	for _, r := range re.seq {
		fmt.Fprint(b, r.String())
	}
	fmt.Fprint(b, ")")
	return b.String()
}

type AstAlt struct {
	opts []Ast
}

func (re *AstAlt) String() string {
	b := bytes.NewBufferString("(?:")
	fmt.Fprint(b, re.opts[0].String())
	for _, r := range re.opts[1:] {
		fmt.Fprint(b, "|")
		fmt.Fprint(b, r.String())
	}
	fmt.Fprint(b, ")")
	return b.String()
}

type AstNotNewline struct{}

func (re AstNotNewline) String() string {
	return "."
}

type AstRepeat struct {
	re       Ast
	min, max int // -1 means unlimited
}

func (re *AstRepeat) String() string {
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

type AstCap struct {
	index uint
	re    Ast
}

func (re *AstCap) String() string {
	return fmt.Sprintf("(%s)", re.re)
}

type AstBackRef uint

func (re AstBackRef) String() string {
	return fmt.Sprintf("\\%d", uint(re))
}

type AstAssertBegin struct{}

func (re AstAssertBegin) String() string {
	return "^"
}

type AstAssertEnd struct{}

func (re AstAssertEnd) String() string {
	return "$"
}

type AstCharClass struct {
	CharClass
	str string
}

func (re AstCharClass) String() string {
	return "[" + re.CharClass.String() + "]"
}
