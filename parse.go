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
	i := 0
LOOP:
	for ; i < len(str); i++ {
		switch str[i] {
		case '$', '^', '*', '(', ')', '+', '[', ']', '{', '}', '|', '\\', '.', '?':
			break LOOP
		}
	}
	if i == 0 {
		panic(fmt.Errorf("Litetal is expected, but cannot find: %q", str))
	}
	return &ReLit{str[0:i]}, str[i:]
}

func parseSeq(str string) (Regexp, string) {
	seq := make([]Regexp, 0, 8)
LOOP:
	for len(str) > 0 {
		switch str[0] {
		case '(':
			var re Regexp
			re, str = parseGroup(str)
			seq = append(seq, re)
		case ')', '|':
			break LOOP
		default:
			var re Regexp
			re, str = parseLit(str)
			seq = append(seq, re)
		}
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
