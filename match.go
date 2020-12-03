package reaot

func Match(re Regexp, s string) bool {
	for i := 0; i < len(s); i++ {
		if re.match(matchContext{nil, 0, i, []rune(s)}, i, func(c matchContext, _ int) *matchContext { return &c }) != nil {
			return true
		}
	}
	return false
}

func (re ReLit) match(c matchContext, p int, k func(matchContext, int) *matchContext) *matchContext {
	if len(c.str)-p < len(re) {
		return nil
	}
	for i := 0; i < len(re); i++ {
		if c.str[p+i] != re[i] {
			return nil
		}
	}
	return k(c, p+len(re))
}

func (re ReNotNewline) match(c matchContext, p int, k func(matchContext, int) *matchContext) *matchContext {
	if len(c.str) <= p || c.str[0] == '\n' {
		return nil
	}
	return k(c, p+1)
}

func (r *ReSeq) match(c matchContext, p int, k func(matchContext, int) *matchContext) *matchContext {
	if len(r.seq) == 0 {
		return k(c, p)
	}
	return r.seq[0].match(c, p, func(c matchContext, p1 int) *matchContext {
		return (&ReSeq{r.seq[1:]}).match(c, p1, k)
	})
}

func (r *ReAlt) match(c matchContext, p int, k func(matchContext, int) *matchContext) *matchContext {
	for _, re := range r.opts {
		if c1 := re.match(c, p, k); c1 != nil {
			return c1
		}
	}
	return nil
}

func (r *ReRepeat) match(c matchContext, p int, k func(matchContext, int) *matchContext) *matchContext {
	re := r.re
	prev := -1 // initial value must be a number which never equal to any position (i.e. positive integer)
	var loop func(count int) func(matchContext, int) *matchContext
	loop = func(count int) func(matchContext, int) *matchContext {
		return func(c matchContext, p int) *matchContext {
			if prev == p { // Matched zero-length assertion. So, move ahead the next pattern.
				return k(c, p)
			}
			prev = p
			if count < r.min {
				return re.match(c, p, loop(count+1))
			}
			if count == r.max {
				return k(c, p)
			}
			c1 := re.match(c, p, loop(count+1))
			if c1 != nil {
				return c1
			}
			return k(c, p)
		}
	}
	return loop(0)(c, p)
}

func (r *ReCap) match(c matchContext, p int, k func(matchContext, int) *matchContext) *matchContext {
	return r.re.match(c.with(r.index, p), p, func(c matchContext, p1 int) *matchContext {
		return k((&c).with(r.index, p1), p1)
	})
}

func (r ReBackRef) match(c matchContext, p int, k func(matchContext, int) *matchContext) *matchContext {
	cap := (&c).GetCaptured(uint(r))
	if cap == nil {
		// This means that the specified capture groups haven't matched any substring.
		// It always fails (according to Perl regex).
		return nil
	}
	return ReLit(cap).match(c, p, k)
}

func (re ReAssertBegin) match(c matchContext, p int, k func(matchContext, int) *matchContext) *matchContext {
	if p != 0 {
		return nil
	}
	return k(c, p)
}

func (re ReAssertEnd) match(c matchContext, p int, k func(matchContext, int) *matchContext) *matchContext {
	if p != len(c.str) {
		return nil
	}
	return k(c, p)
}

func (re ReCharClass) match(c matchContext, p int, k func(matchContext, int) *matchContext) *matchContext {
	if len(c.str) < p+1 {
		return nil
	}
	if !re.Contains(c.str[p]) {
		return nil
	}
	return k(c, p+1)
}
