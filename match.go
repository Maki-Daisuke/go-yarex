package reaot

func Match(re Regexp, s string) (ret bool) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if _, ok := e.(*matchContext); ok {
			ret = true
			return
		}
		panic(e) // rethrow error
	}()
	for i := 0; i < len(s); i++ {
		re.match(matchContext{nil, 0, i, []rune(s)}, i, func(c matchContext, _ int) { panic(&c) })
		// If you are here, matching failed. So, try next.
	}
	return false
}

func (re ReLit) match(c matchContext, p int, k func(matchContext, int)) {
	if len(c.str)-p < len(re) {
		return
	}
	for i := 0; i < len(re); i++ {
		if c.str[p+i] != re[i] {
			return
		}
	}
	k(c, p+len(re))
}

func (re ReNotNewline) match(c matchContext, p int, k func(matchContext, int)) {
	if len(c.str) <= p || c.str[0] == '\n' {
		return
	}
	k(c, p+1)
}

func (r *ReSeq) match(c matchContext, p int, k func(matchContext, int)) {
	if len(r.seq) == 0 {
		k(c, p)
		return
	}
	r.seq[0].match(c, p, func(c matchContext, p1 int) {
		(&ReSeq{r.seq[1:]}).match(c, p1, k)
	})
}

func (r *ReAlt) match(c matchContext, p int, k func(matchContext, int)) {
	for _, re := range r.opts {
		re.match(c, p, k)
	}
}

func (r *ReRepeat) match(c matchContext, p int, k func(matchContext, int)) {
	re := r.re
	prev := -1 // initial value must be a number which never equal to any position (i.e. positive integer)
	var loop func(count int) func(matchContext, int)
	loop = func(count int) func(matchContext, int) {
		return func(c matchContext, p int) {
			if prev == p { // Matched zero-length assertion. So, move ahead the next pattern.
				k(c, p)
				return
			}
			prev = p
			if count < r.min {
				re.match(c, p, loop(count+1))
				return
			}
			if count == r.max {
				k(c, p)
				return
			}
			re.match(c, p, loop(count+1))
			k(c, p)
		}
	}
	loop(0)(c, p)
}

func (r *ReCap) match(c matchContext, p int, k func(matchContext, int)) {
	r.re.match(c.with(r.index, p), p, func(c matchContext, p1 int) {
		k((&c).with(r.index, p1), p1)
	})
}

func (r ReBackRef) match(c matchContext, p int, k func(matchContext, int)) {
	cap := (&c).GetCaptured(uint(r))
	if cap == nil {
		// This means that the specified capture groups haven't matched any substring.
		// It always fails (according to Perl regex).
		return
	}
	ReLit(cap).match(c, p, k)
}

func (re ReAssertBegin) match(c matchContext, p int, k func(matchContext, int)) {
	if p != 0 {
		return
	}
	k(c, p)
}

func (re ReAssertEnd) match(c matchContext, p int, k func(matchContext, int)) {
	if p != len(c.str) {
		return
	}
	k(c, p)
}

func (re ReCharClass) match(c matchContext, p int, k func(matchContext, int)) {
	if len(c.str) < p+1 {
		return
	}
	if !re.Contains(c.str[p]) {
		return
	}
	k(c, p+1)
}
