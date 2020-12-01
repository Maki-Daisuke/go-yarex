package reaot

func Match(re Regexp, s string) bool {
	for i := 0; i < len(s); i++ {
		if re.match(matchContext{nil, 0, i}, []rune(s), i, func(c matchContext, _ int) *matchContext { return &c }) != nil {
			return true
		}
	}
	return false
}

func (re ReLit) match(c matchContext, s []rune, p int, k func(matchContext, int) *matchContext) *matchContext {
	if len(s)-p < len(re) {
		return nil
	}
	for i := 0; i < len(re); i++ {
		if s[p+i] != re[i] {
			return nil
		}
	}
	return k(c, p+len(re))
}

func (re ReNotNewline) match(c matchContext, s []rune, p int, k func(matchContext, int) *matchContext) *matchContext {
	if len(s) <= p || s[0] == '\n' {
		return nil
	}
	return k(c, p+1)
}

func (r *ReSeq) match(c matchContext, s []rune, p int, k func(matchContext, int) *matchContext) *matchContext {
	if len(r.seq) == 0 {
		return k(c, p)
	}
	return r.seq[0].match(c, s, p, func(c matchContext, p1 int) *matchContext {
		return (&ReSeq{r.seq[1:]}).match(c, s, p1, k)
	})
}

func (r *ReAlt) match(c matchContext, s []rune, p int, k func(matchContext, int) *matchContext) *matchContext {
	for _, re := range r.opts {
		if c1 := re.match(c, s, p, k); c1 != nil {
			return c1
		}
	}
	return nil
}

func (r *ReZeroOrMore) match(c matchContext, s []rune, p int, k func(matchContext, int) *matchContext) *matchContext {
	re := r.re
	c1 := re.match(c, s, p, func(c matchContext, p1 int) *matchContext {
		if p1 == p { // This means zero-length assertion pattern matched.
			return k(c, p1) // So, move ahead to the following pattern.
		}
		return r.match(c, s, p1, k)
	})
	if c1 != nil {
		return c1
	}
	return k(c, p)
}

func (r *ReOneOrMore) match(c matchContext, s []rune, p int, k func(matchContext, int) *matchContext) *matchContext {
	re := r.re
	if c1 := re.match(c, s, p, func(c matchContext, p1 int) *matchContext { return (&ReZeroOrMore{re}).match(c, s, p1, k) }); c1 != nil {
		return c1
	}
	return nil
}

func (r *ReOpt) match(c matchContext, s []rune, p int, k func(matchContext, int) *matchContext) *matchContext {
	re := r.re
	if c1 := re.match(c, s, p, k); c1 != nil {
		return c1
	}
	return k(c, p)
}

func (r *ReCap) match(c matchContext, s []rune, p int, k func(matchContext, int) *matchContext) *matchContext {
	return r.re.match((&c).with(r.index, p), s, p, func(c matchContext, p1 int) *matchContext {
		return k((&c).with(r.index, p1), p1)
	})
}

func (r ReBackRef) match(c matchContext, s []rune, p int, k func(matchContext, int) *matchContext) *matchContext {
	ref := (&c).get(uint(r))
	if ref == nil {
		// This means that the specified capture groups haven't matched any substring.
		// It always fails (according to Perl regex).
		return nil
	}
	return ReLit(s[ref[0]:ref[1]]).match(c, s, p, k)
}

func (re ReAssertBegin) match(c matchContext, s []rune, p int, k func(matchContext, int) *matchContext) *matchContext {
	if p != 0 {
		return nil
	}
	return k(c, p)
}
