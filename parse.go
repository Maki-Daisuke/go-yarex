package reaot

import (
	"fmt"
)

func parse(s string) (re Regexp, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			re = nil
		}
	}()
	re, remain := parseAlt(s)
	if remain != "" {
		return nil, fmt.Errorf("Unknown context: %q", remain)
	}
	return re, nil
}

func parseLit(str string) (Regexp, string) {
	if len(str) == 0 {
		panic(fmt.Errorf("Litetal is expected, but reached end-of-string unexpectedly"))
	}
	switch str[0] {
	case '$', '^', '*', '(', ')', '+', '[', ']', '{', '}', '|', '\\', '.', '?':
		panic(fmt.Errorf("Litetal is expected, but cannot find: %q", str))
	}
	return &ReLit{str[0:1]}, str[1:]
}

func parseSeq(str string) (Regexp, string) {
	seq := make([]Regexp, 0, 8)
LOOP:
	for len(str) > 0 {
		var re Regexp
		switch str[0] {
		case '.':
			re = ReNotNewline{}
			str = str[1:]
		case '(':
			re, str = parseGroup(str)
		case ')', '|':
			break LOOP
		default:
			re, str = parseLit(str)
		}
		re, str = parseQuantifier(str, re)
		seq = append(seq, re)
	}
	if len(seq) == 1 {
		return seq[0], str
	} else {
		return &ReSeq{seq}, str
	}
}

func parseAlt(str string) (Regexp, string) {
	re, str := parseSeq(str)
	opts := []Regexp{re}
LOOP:
	for len(str) > 0 {
		switch str[0] {
		case '|':
			var re Regexp
			re, str = parseAlt(str[1:])
			opts = append(opts, re)
		case ')':
			break LOOP
		default:
			panic(fmt.Errorf("Unknown context: %q", str))
		}
	}
	if len(opts) == 1 {
		return opts[0], str
	} else {
		return &ReAlt{opts}, str
	}
}

func parseGroup(str string) (Regexp, string) {
	if str[0] != '(' {
		panic(fmt.Errorf("'(' is expected, but cannot find: %q", str))
	}
	re, remain := parseAlt(str[1:])
	if remain[0] != ')' {
		panic(fmt.Errorf("Unmatched '(' : %q", str))
	}
	return re, remain[1:]
}

func parseQuantifier(str string, re Regexp) (Regexp, string) {
	if len(str) == 0 {
		return re, str
	}
	switch str[0] {
	case '*':
		re = &ReZeroOrMore{re}
		str = str[1:]
	case '+':
		re = &ReOneOrMore{re}
		str = str[1:]
	case '?':
		re = &ReOpt{re}
		str = str[1:]
	}
	return re, str
}
