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
	re, remain := (&(parser{})).parseAlt([]rune(s))
	if len(remain) > 0 {
		return nil, fmt.Errorf("Unknown context: %q", remain)
	}
	return re, nil
}

type parser struct {
	openCaptures  uint
	closeCaptures uint
}

func (*parser) parseLit(str []rune) (Regexp, []rune) {
	if len(str) == 0 {
		panic(fmt.Errorf("Litetal is expected, but reached end-of-string unexpectedly"))
	}
	switch str[0] {
	case '$', '^', '*', '(', ')', '+', '[', ']', '{', '}', '|', '\\', '.', '?':
		panic(fmt.Errorf("Litetal is expected, but cannot find: %q", str))
	}
	return ReLit(str[0:1]), str[1:]
}

func (p *parser) parseSeq(str []rune) (Regexp, []rune) {
	seq := make([]Regexp, 0, 8)
LOOP:
	for len(str) > 0 {
		var re Regexp
		switch str[0] {
		case '^':
			re = ReAssertBegin{}
			str = str[1:]
		case '.':
			re = ReNotNewline{}
			str = str[1:]
		case '\\':
			re, str = p.parseEscape(str)
		case '(':
			re, str = p.parseGroup(str)
		case ')', '|':
			break LOOP
		default:
			re, str = p.parseLit(str)
		}
		re, str = p.parseQuantifier(str, re)
		seq = append(seq, re)
	}
	if len(seq) == 1 {
		return seq[0], str
	} else {
		return &ReSeq{seq}, str
	}
}

func (p *parser) parseAlt(str []rune) (Regexp, []rune) {
	re, str := p.parseSeq(str)
	opts := []Regexp{re}
LOOP:
	for len(str) > 0 {
		switch str[0] {
		case '|':
			var re Regexp
			re, str = p.parseAlt(str[1:])
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

func (p *parser) parseGroup(str []rune) (Regexp, []rune) {
	if str[0] != '(' {
		panic(fmt.Errorf("'(' is expected, but cannot find: %q", str))
	}
	if len(str) < 2 {
		panic(fmt.Errorf("Unmatched '(' : %q", str))
	}
	if str[1] != '?' {
		return p.parseCapture(str[1:])
	}
	if str[2] != ':' {
		panic(fmt.Errorf("Unknown extended pattern syntzx: %q", str))
	}
	re, remain := p.parseAlt(str[1:])
	if remain[0] != ')' {
		panic(fmt.Errorf("Unmatched '(' : %q", str))
	}
	return re, remain[1:]
}

func (p *parser) parseCapture(str []rune) (Regexp, []rune) {
	p.openCaptures++
	index := p.openCaptures
	re, remain := p.parseAlt(str)
	if remain[0] != ')' {
		panic(fmt.Errorf("Unmatched '(' : %q", str))
	}
	p.closeCaptures++
	return &ReCap{index, re}, remain[1:]
}

func (p *parser) parseQuantifier(str []rune, re Regexp) (Regexp, []rune) {
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

func (p *parser) parseEscape(str []rune) (Regexp, []rune) {
	if str[0] != '\\' {
		panic(fmt.Errorf("'\\' is expected, but cannot find: %q", str))
	}
	if len(str) < 2 {
		panic(fmt.Errorf("Trailing '\\' in regex: %q", str))
	}
	switch str[1] {
	case ' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/', ':', ';',
		'<', '=', '>', '?', '@', '[', '\\', ']', '^', '_', '`', '{', '|', '}', '~':
		return ReLit(str[1:2]), str[2:]
	case '0':
		return ReLit([]rune{0}), str[2:]
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return ReBackRef(str[1] - '0'), str[2:]
	default:
		panic(fmt.Errorf("Unknown escape sequence: %q", str))
	}
}
