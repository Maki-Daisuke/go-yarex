package yarex

import "fmt"

func optimizeAst(re Ast) Ast {
	re = optimizeAstFlattenSeqAndAlt(re)
	re = optimizeAstUnwrapSingletonSeqAndAlt(re)
	return re
}

// optimizeAstFlattenSeqAndAlt flattens nested AstSeqs and AstAlts
func optimizeAstFlattenSeqAndAlt(re Ast) Ast {
	switch v := re.(type) {
	case *AstSeq:
		out := make([]Ast, 0, len(v.seq))
		for _, r := range v.seq {
			r = optimizeAstFlattenSeqAndAlt(r)
			if as, ok := r.(*AstSeq); ok {
				out = append(out, as.seq...)
			} else {
				out = append(out, r)
			}
		}
		return &AstSeq{out}
	case *AstAlt:
		out := make([]Ast, 0, len(v.opts))
		for _, r := range v.opts {
			r = optimizeAstFlattenSeqAndAlt(r)
			if aa, ok := r.(*AstAlt); ok {
				out = append(out, aa.opts...)
			} else {
				out = append(out, r)
			}
		}
		return &AstAlt{out}
	case *AstRepeat:
		out := *v
		out.re = optimizeAstFlattenSeqAndAlt(v.re)
		return &out
	case *AstCap:
		out := *v
		out.re = optimizeAstFlattenSeqAndAlt(v.re)
		return &out
	case AstLit, AstNotNewline, AstAssertBegin, AstAssertEnd, AstBackRef, AstCharClass:
		return v
	default:
		panic(fmt.Errorf("IMPLEMENT optimizeAstFlattenSeqAndAlt for %T", re))
	}
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
