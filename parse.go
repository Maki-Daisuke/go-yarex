package reaot

import (
	"fmt"
	"strconv"
	"unicode"
)

func parse(s string) (re Ast, err error) {
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

func (*parser) parseLit(str []rune) (Ast, []rune) {
	if len(str) == 0 {
		panic(fmt.Errorf("Literal is expected, but reached end-of-string unexpectedly"))
	}
	switch str[0] {
	case '$', '^', '*', '(', ')', '+', '[', ']', '{', '}', '|', '\\', '.', '?':
		panic(fmt.Errorf("Literal is expected, but cannot find: %q", string(str)))
	}
	return AstLit(str[0:1]), str[1:]
}

func (p *parser) parseSeq(str []rune) (Ast, []rune) {
	seq := make([]Ast, 0, 8)
LOOP:
	for len(str) > 0 {
		var re Ast
		switch str[0] {
		case '^':
			re = AstAssertBegin{}
			str = str[1:]
		case '$':
			re = AstAssertEnd{}
			str = str[1:]
		case '.':
			re = AstNotNewline{}
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
		return &AstSeq{seq}, str
	}
}

func (p *parser) parseAlt(str []rune) (Ast, []rune) {
	re, str := p.parseSeq(str)
	opts := []Ast{re}
LOOP:
	for len(str) > 0 {
		switch str[0] {
		case '|':
			var re Ast
			re, str = p.parseAlt(str[1:])
			opts = append(opts, re)
		case ')':
			break LOOP
		default:
			panic(fmt.Errorf("Unknown context: %q", string(str)))
		}
	}
	if len(opts) == 1 {
		return opts[0], str
	} else {
		return &AstAlt{opts}, str
	}
}

func (p *parser) parseGroup(str []rune) (Ast, []rune) {
	if str[0] != '(' {
		panic(fmt.Errorf("'(' is expected, but cannot find: %q", string(str)))
	}
	if len(str) < 2 {
		panic(fmt.Errorf("Unmatched '(' : %q", string(str)))
	}
	if str[1] != '?' {
		return p.parseCapture(str[1:])
	}
	if str[2] != ':' {
		panic(fmt.Errorf("Unknown extended pattern syntzx: %q", string(str)))
	}
	re, remain := p.parseAlt(str[1:])
	if remain[0] != ')' {
		panic(fmt.Errorf("Unmatched '(' : %q", string(str)))
	}
	return re, remain[1:]
}

func (p *parser) parseCapture(str []rune) (Ast, []rune) {
	p.openCaptures++
	index := p.openCaptures
	re, remain := p.parseAlt(str)
	if remain[0] != ')' {
		panic(fmt.Errorf("Unmatched '(' : %q", string(str)))
	}
	p.closeCaptures++
	return &AstCap{index, re}, remain[1:]
}

func (p *parser) parseQuantifier(str []rune, re Ast) (Ast, []rune) {
	if len(str) == 0 {
		return re, str
	}
	switch str[0] {
	case '*':
		return &AstRepeat{re, 0, -1}, str[1:]
	case '+':
		return &AstRepeat{re, 1, -1}, str[1:]
	case '?':
		return &AstRepeat{re, 0, 1}, str[1:]
	case '{':
		start, remain := p.parseInt(str[1:])
		if remain == nil {
			panic(fmt.Errorf(`Invalid quantifier: %q`, string(str)))
		}
		switch remain[0] {
		case '}':
			return &AstRepeat{re, start, start}, remain[1:]
		case ',':
			end, remain := p.parseInt(remain[1:])
			if remain == nil {
				panic(fmt.Errorf(`Invalid quantifier: %q`, string(str)))
			}
			if remain[0] != '}' {
				panic(fmt.Errorf("Unmatched '{' : %q", string(str)))
			}
			return &AstRepeat{re, start, end}, remain[1:]
		default:
			panic(fmt.Errorf("Unmatched '{' : %q", string(str)))
		}
	}
	return re, str
}

// parseInt returns (0, nil) if it cannot find any integer at the head of str
func (p *parser) parseInt(str []rune) (int, []rune) {
	i := 0
	for ; i < len(str); i++ {
		if str[i] < '0' || '9' < str[i] {
			break
		}
	}
	if i == 0 {
		return 0, nil
	}
	x, err := strconv.ParseInt(string(str[0:i]), 10, 32)
	if err != nil {
		panic(fmt.Errorf(`(THIS SHOULD NOT HAPPEN) can't parse int: %q`, string(str[0:i])))
	}
	return int(x), str[i:]
}

func (p *parser) parseEscape(str []rune) (Ast, []rune) {
	return p.parseEscapeAux(str, false)
}

func (p *parser) parseEscapeAux(str []rune, inClass bool) (Ast, []rune) {
	if str[0] != '\\' {
		panic(fmt.Errorf("'\\' is expected, but cannot find: %q", string(str)))
	}
	if len(str) < 2 {
		panic(fmt.Errorf("Trailing '\\' in regex: %q", string(str)))
	}
	switch str[1] {
	case ' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/', ':', ';',
		'<', '=', '>', '?', '@', '[', '\\', ']', '^', '_', '`', '{', '|', '}', '~':
		return AstLit(str[1:2]), str[2:]
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if str[2] < '0' || '9' < str[2] {
			if str[1] == '0' {
				return AstLit([]rune{0}), str[2:]
			}
			if !inClass {
				return AstBackRef(str[1] - '0'), str[2:]
			}
			panic(fmt.Errorf("invalid character %q in octal escape: %q", str[2], string(str)))
		}
		if str[3] < '0' || '9' < str[3] {
			panic(fmt.Errorf("invalid character %q in octal escape: %q", str[3], string(str)))
		}
		oct, err := strconv.ParseUint(string(str[1:4]), 8, 8)
		if err != nil {
			panic(fmt.Errorf("can't parse octal escape in %q: %w", string(str), err))
		}
		return AstLit([]rune{rune(oct)}), str[4:]
	default:
		panic(fmt.Errorf("Unknown escape sequence: %q", string(str)))
	}
}

func (p *parser) parseClass(str []rune) (Ast, []rune) {
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
			var re Ast
			re, str = p.parseEscapeAux(str, true)
			from = ([]rune)(re.(AstLit))[0] // This must work at least now
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
			var re Ast
			re, str = p.parseEscapeAux(str[1:], true)
			to := ([]rune)(re.(AstLit))[0] // This must work at least now
			rangeTable = mergeRangeTable(rangeTable, rangeTableFromTo(from, to))
			break
		default:
			rangeTable = mergeRangeTable(rangeTable, rangeTableFromTo(from, str[1]))
			str = str[2:]
		}
	}
	if rangeTable.R16 != nil || rangeTable.R32 != nil {
		ccs = append(ccs, (*RangeTableClass)(rangeTable))
	}
	var out CharClass
	if len(ccs) == 1 {
		out = ccs[0]
	} else {
		out = CompositeClass(ccs)
	}
	out = toAsciiMaskClass(out) // this returns the input as-is if impossible to convert to asciiMaskClass
	if isNegate {
		out = NegateCharClass(out)
	}
	return AstCharClass{out}, str
}
