package reaot

import "unicode"

func optimizeAst(re Ast) Ast {
	re = optimizeAstSingleCharacterClass(re)
	re = optimizeAstUnwrapSingletonSeqAndAlt(re)
	return re
}

// optimizeAstUnwrapSingletonSeqAndAlt joins adjacent literals, and unwrap seqs and alts
// containing a single re as much as possible
func optimizeAstUnwrapSingletonSeqAndAlt(re Ast) Ast {
	switch v := re.(type) {
	case *AstSeq:
		out := make([]Ast, 0, len(v.seq))
		var acc *string = nil
		for _, r := range v.seq {
			r = optimizeAstUnwrapSingletonSeqAndAlt(r)
			if lit, ok := r.(AstLit); ok {
				if acc == nil {
					s := string(lit)
					acc = &s
				} else {
					*acc = *acc + string(lit)
				}
			} else {
				if acc != nil {
					out = append(out, AstLit(*acc))
					acc = nil
				}
				out = append(out, r)
			}
		}
		if acc != nil {
			out = append(out, AstLit(*acc))
		}
		switch len(out) {
		case 0:
			return AstLit("")
		case 1:
			return out[0]
		}
		return &AstSeq{out}
	case *AstAlt:
		out := make([]Ast, len(v.opts), len(v.opts))
		for i, r := range v.opts {
			out[i] = optimizeAstUnwrapSingletonSeqAndAlt(r)
		}
		switch len(out) {
		case 0:
			return AstLit("")
		case 1:
			return out[0]
		}
		return &AstAlt{out}
	case *AstRepeat:
		out := *v
		out.re = optimizeAstUnwrapSingletonSeqAndAlt(v.re)
		return &out
	case *AstCap:
		out := *v
		out.re = optimizeAstUnwrapSingletonSeqAndAlt(v.re)
		return &out
	default:
		return v
	}
}

// optimizeAstSingleCharacterClass replaces ReCharClass containing a single codepoint with ReLit
func optimizeAstSingleCharacterClass(re Ast) Ast {
	switch v := re.(type) {
	case *AstSeq:
		out := make([]Ast, len(v.seq), len(v.seq))
		for i, r := range v.seq {
			out[i] = optimizeAstSingleCharacterClass(r)
		}
		return &AstSeq{out}
	case *AstAlt:
		out := make([]Ast, len(v.opts), len(v.opts))
		for i, r := range v.opts {
			out[i] = optimizeAstSingleCharacterClass(r)
		}
		return &AstAlt{out}
	case *AstRepeat:
		out := *v
		out.re = optimizeAstSingleCharacterClass(v.re)
		return &out
	case *AstCap:
		out := *v
		out.re = optimizeAstSingleCharacterClass(v.re)
		return &out
	case AstCharClass:
		if rtc, ok := v.CharClass.(*RangeTableClass); ok {
			rt := (*unicode.RangeTable)(rtc)
			if len(rt.R16) == 1 && len(rt.R32) == 0 && rt.R16[0].Lo == rt.R16[0].Hi {
				return AstLit(string(rune(rt.R16[0].Lo)))
			}
			if len(rt.R16) == 0 && len(rt.R32) == 1 && rt.R32[0].Lo == rt.R32[0].Hi {
				return AstLit(string(rune(rt.R32[0].Lo)))
			}
		}
		return v
	default:
		return v
	}
}

func canOnlyMatchAtBegining(re Ast) bool {
	switch v := re.(type) {
	case AstAssertBegin:
		return true
	case *AstSeq:
		if len(v.seq) == 0 {
			return false
		}
		return canOnlyMatchAtBegining(v.seq[0])
	case *AstAlt:
		if len(v.opts) == 0 {
			return false
		}
		for _, r := range v.opts {
			if !canOnlyMatchAtBegining(r) {
				return false
			}
		}
		return true
	case *AstRepeat:
		if v.min == 0 {
			return false
		}
		return canOnlyMatchAtBegining(v.re)
	case *AstCap:
		return canOnlyMatchAtBegining(v.re)
	default:
		return false
	}
}
