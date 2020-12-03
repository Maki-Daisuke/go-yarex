package reaot

import (
	"fmt"
	"strconv"
	"unicode"
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
		case '[':
			re, str = p.parseClass(str)
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
	return p.parseEscapeAux(str, false)
}

func (p *parser) parseEscapeAux(str []rune, inClass bool) (Regexp, []rune) {
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
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if str[2] < '0' || '9' < str[2] {
			if str[1] == '0' {
				return ReLit([]rune{0}), str[2:]
			}
			if !inClass {
				return ReBackRef(str[1] - '0'), str[2:]
			}
			panic(fmt.Errorf("invalid character %q in octal escape: %q", str[2], str))
		}
		if str[3] < '0' || '9' < str[3] {
			panic(fmt.Errorf("invalid character %q in octal escape: %q", str[3], str))
		}
		oct, err := strconv.ParseUint(string(str[1:4]), 8, 8)
		if err != nil {
			panic(fmt.Errorf("can't parse octal escape in %q: %w", str, err))
		}
		return ReLit([]rune{rune(oct)}), str[4:]
	default:
		panic(fmt.Errorf("Unknown escape sequence: %q", str))
	}
}

func (p *parser) parseClass(str []rune) (Regexp, []rune) {
	if str[0] != '[' {
		panic(fmt.Errorf("'[' is expected, but cannot find: %q", string(str)))
	}
	origStr := str
	str = str[1:]
	rangeTable := &unicode.RangeTable{}
	ccs := []CharClass{}
	isNegate := false
	if str[0] == '^' {
		isNegate = true
		str = str[1:]
	}
	if str[0] == ']' || str[0] == '-' {
		rangeTable = mergeRangeTable(rangeTable, rangeTableFromTo(str[0], str[0]))
		str = str[1:]
	}
LOOP:
	for {
		if len(str) == 0 {
			panic(fmt.Errorf("unmatched '[' here: %q", string(origStr)))
		}
		if str[0] == ']' {
			str = str[1:]
			break LOOP
		}
		var from rune
		if str[0] == '\\' {
			var re Regexp
			re, str = p.parseEscapeAux(str, true)
			from = ([]rune)(re.(ReLit))[0] // This must work at least now
		} else {
			from = str[0]
			str = str[1:]
		}
		if str[0] != '-' {
			rangeTable = mergeRangeTable(rangeTable, rangeTableFromTo(from, from))
			continue LOOP
		}
		switch str[1] { // In the case of character range, i.e. "X-Y"
		case ']':
			rangeTable = mergeRangeTable(rangeTable, rangeTableFromTo(from, from))
			rangeTable = mergeRangeTable(rangeTable, rangeTableFromTo('-', '-'))
			str = str[2:]
			break LOOP
		case '\\':
			var re Regexp
			re, str = p.parseEscapeAux(str[1:], true)
			to := ([]rune)(re.(ReLit))[0] // This must work at least now
			rangeTable = mergeRangeTable(rangeTable, rangeTableFromTo(from, to))
			break
		default:
			rangeTable = mergeRangeTable(rangeTable, rangeTableFromTo(from, str[1]))
			str = str[2:]
		}
	}
	if rangeTable.R16 != nil || rangeTable.R32 != nil {
		ccs = append(ccs, (*rangeTableClass)(rangeTable))
	}
	var out CharClass
	if len(ccs) == 1 {
		out = ccs[0]
	} else {
		out = compositeClass(ccs)
	}
	if isNegate {
		out = NegateCharClass(out)
	}
	return ReCharClass{out}, str
}
