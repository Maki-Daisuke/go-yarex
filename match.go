package reaot

import "strings"

func Match(re Regexp, s string) bool {
	for i := 0; i < len(s); i++ {
		if _, ok := re.Match(s[i:]); ok {
			return true
		}
	}
	return false
}

func (re *ReLit) Match(s string) (string, bool) {
	if strings.HasPrefix(s, re.str) {
		return s[len(re.str):], true
	}
	return "", false
}

func (r *ReSeq) Match(s string) (string, bool) {
	for _, re := range r.seq {
		var ok bool
		if s, ok = re.Match(s); !ok {
			return "", false
		}
	}
	return s, true
}

func (r *ReAlt) Match(s string) (string, bool) {
	for _, re := range r.opts {
		if t, ok := re.Match(s); ok {
			return t, true
		}
	}
	return "", false
}
