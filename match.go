package reaot

import "strings"

func Match(re Regexp, s string) bool {
	for i := 0; i < len(s); i++ {
		if re.match(s, i, func(_ int) bool { return true }) {
			return true
		}
	}
	return false
}

func (re *ReLit) match(s string, p int, k func(int) bool) bool {
	if !strings.HasPrefix(s[p:], re.str) {
		return false
	}
	return k(p + len(re.str))
}

func (re ReNotNewline) match(s string, p int, k func(int) bool) bool {
	if len(s) <= p || s[0] == '\n' {
		return false
	}
	return k(p + 1)
}

func (r *ReSeq) match(s string, p int, k func(int) bool) bool {
	if len(r.seq) == 0 {
		return k(p)
	}
	return r.seq[0].match(s, p, func(p1 int) bool {
		return (&ReSeq{r.seq[1:]}).match(s, p1, k)
	})
}

func (r *ReAlt) match(s string, p int, k func(int) bool) bool {
	for _, re := range r.opts {
		if re.match(s, p, k) {
			return true
		}
	}
	return false
}

func (r *ReZeroOrMore) match(s string, p int, k func(int) bool) bool {
	re := r.re
	matched := re.match(s, p, func(p1 int) bool {
		if p1 == p { // This means zero-length assertion pattern matched.
			return k(p1) // So, move ahead to the following pattern.
		}
		return r.match(s, p1, k)
	})
	if matched {
		return matched
	}
	return k(p)
}

func (r *ReOneOrMore) match(s string, p int, k func(int) bool) bool {
	re := r.re
	if re.match(s, p, func(p1 int) bool { return (&ReZeroOrMore{re}).match(s, p1, k) }) {
		return true
	}
	return false
}

func (r *ReOpt) match(s string, p int, k func(int) bool) bool {
	re := r.re
	if re.match(s, p, k) {
		return true
	}
	return k(p)
}

func (re ReAssertBegin) match(s string, p int, k func(int) bool) bool {
	if p != 0 {
		return false
	}
	return k(p)
}
