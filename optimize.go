package reaot

import "unicode"

func optimize(re Regexp) Regexp {
	re = optimizeSingleCharacterClass(re)
	re = optimizeUnwrapSingletonSeqAndAlt(re)
	return re
}

// Join adjacent literals and unwrap seq and alt containing a single re as much as possible
func optimizeUnwrapSingletonSeqAndAlt(re Regexp) Regexp {
	switch v := re.(type) {
	case *ReSeq:
		out := make([]Regexp, 0, len(v.seq))
		var acc *string = nil
		for _, r := range v.seq {
			r = optimizeUnwrapSingletonSeqAndAlt(r)
			if lit, ok := r.(ReLit); ok {
				if acc == nil {
					s := string(lit)
					acc = &s
				} else {
					*acc = *acc + string(lit)
				}
			} else {
				if acc != nil {
					out = append(out, ReLit(*acc))
					acc = nil
				}
				out = append(out, r)
			}
		}
		if acc != nil {
			out = append(out, ReLit(*acc))
		}
		switch len(out) {
		case 0:
			return ReLit("")
		case 1:
			return out[0]
		}
		return &ReSeq{out}
	case *ReAlt:
		out := make([]Regexp, len(v.opts), len(v.opts))
		for i, r := range v.opts {
			out[i] = optimizeUnwrapSingletonSeqAndAlt(r)
		}
		switch len(out) {
		case 0:
			return ReLit("")
		case 1:
			return out[0]
		}
		return &ReAlt{out}
	case *ReRepeat:
		out := *v
		out.re = optimizeUnwrapSingletonSeqAndAlt(v.re)
		return &out
	case *ReCap:
		out := *v
		out.re = optimizeUnwrapSingletonSeqAndAlt(v.re)
		return &out
	default:
		return v
	}
}

// optimizeSingleCharacterClass replaces ReCharClass containing a single codepoint with ReLit
func optimizeSingleCharacterClass(re Regexp) Regexp {
	switch v := re.(type) {
	case *ReSeq:
		out := make([]Regexp, len(v.seq), len(v.seq))
		for i, r := range v.seq {
			out[i] = optimizeSingleCharacterClass(r)
		}
		return &ReSeq{out}
	case *ReAlt:
		out := make([]Regexp, len(v.opts), len(v.opts))
		for i, r := range v.opts {
			out[i] = optimizeSingleCharacterClass(r)
		}
		return &ReAlt{out}
	case *ReRepeat:
		out := *v
		out.re = optimizeSingleCharacterClass(v.re)
		return &out
	case *ReCap:
		out := *v
		out.re = optimizeSingleCharacterClass(v.re)
		return &out
	case ReCharClass:
		if rtc, ok := v.CharClass.(*rangeTableClass); ok {
			rt := (*unicode.RangeTable)(rtc)
			if len(rt.R16) == 1 && len(rt.R32) == 0 && rt.R16[0].Lo == rt.R16[0].Hi {
				return ReLit(string(rune(rt.R16[0].Lo)))
			}
			if len(rt.R16) == 0 && len(rt.R32) == 1 && rt.R32[0].Lo == rt.R32[0].Hi {
				return ReLit(string(rune(rt.R32[0].Lo)))
			}
		}
		return v
	default:
		return v
	}
}

func canOnlyMatchAtBegining(re Regexp) bool {
	switch v := re.(type) {
	case ReAssertBegin:
		return true
	case *ReSeq:
		if len(v.seq) == 0 {
			return false
		}
		return canOnlyMatchAtBegining(v.seq[0])
	case *ReAlt:
		if len(v.opts) == 0 {
			return false
		}
		for _, r := range v.opts {
			if !canOnlyMatchAtBegining(r) {
				return false
			}
		}
		return true
	case *ReRepeat:
		if v.min == 0 {
			return false
		}
		return canOnlyMatchAtBegining(v.re)
	case *ReCap:
		return canOnlyMatchAtBegining(v.re)
	default:
		return false
	}
}
