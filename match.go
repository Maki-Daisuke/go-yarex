package reaot

import "strings"

func Match(re Regexp, s string) bool {
	for i := 0; i < len(s); i++ {
		if re.match(s[i:], func(_ string) bool { return true }) {
			return true
		}
	}
	return false
}

func (re *ReLit) match(s string, k func(string) bool) bool {
	if !strings.HasPrefix(s, re.str) {
		return false
	}
	return k(s[len(re.str):])
}

func (re ReNotNewline) match(s string, k func(string) bool) bool {
	if len(s) == 0 || s[0] == '\n' {
		return false
	}
	return k(s[1:])
}

func (r *ReSeq) match(s string, k func(string) bool) bool {
	if len(r.seq) == 0 {
		return k(s)
	}
	return r.seq[0].match(s, func(t string) bool {
		return (&ReSeq{r.seq[1:]}).match(t, k)
	})
}

func (r *ReAlt) match(s string, k func(string) bool) bool {
	for _, re := range r.opts {
		if re.match(s, k) {
			return true
		}
	}
	return false
}

func (r *ReZeroOrMore) match(s string, k func(string) bool) bool {
	re := r.re
	if re.match(s, func(s string) bool { return r.match(s, k) }) {
		return true
	}
	return k(s)
}

func (r *ReOneOrMore) match(s string, k func(string) bool) bool {
	re := r.re
	if re.match(s, func(s string) bool { return (&ReZeroOrMore{re}).match(s, k) }) {
		return true
	}
	return false
}

func (r *ReOpt) match(s string, k func(string) bool) bool {
	re := r.re
	if re.match(s, k) {
		return true
	}
	return k(s)
}
