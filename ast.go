package reaot

import (
	"bytes"
	"fmt"
)

type ReLit struct {
	str string
}

func (re *ReLit) String() string {
	return re.str
}

type ReSeq struct {
	seq []Regexp
}

func (re *ReSeq) String() string {
	b := bytes.NewBufferString("(")
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
	b := bytes.NewBufferString("(")
	fmt.Fprint(b, re.opts[0].String())
	for _, r := range re.opts[1:] {
		fmt.Fprint(b, "|")
		fmt.Fprint(b, r.String())
	}
	fmt.Fprint(b, ")")
	return b.String()
}
