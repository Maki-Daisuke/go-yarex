package reaot

func optimize(re Regexp) Regexp {
	return optimizeJoinLiterals(re)
}

func optimizeJoinLiterals(re Regexp) Regexp {
	switch v := re.(type) {
	case *ReSeq:
		out := make([]Regexp, 0, len(v.seq))
		var acc *string = nil
		for _, r := range v.seq {
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
		return &ReSeq{out}
	case *ReAlt:
		out := make([]Regexp, len(v.opts), len(v.opts))
		for i, r := range v.opts {
			out[i] = optimizeJoinLiterals(r)
		}
		return &ReAlt{out}
	case *ReRepeat:
		out := *v
		out.re = optimizeJoinLiterals(v.re)
		return &out
	case *ReCap:
		out := *v
		out.re = optimizeJoinLiterals(v.re)
		return &out
	default:
		return v
	}
}
