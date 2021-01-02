package yarex

import "unsafe"

var compiledRegexps = map[string]*Regexp{}

func RegisterCompiledRegexp(s string, h bool, m int, f func(int, uintptr, int, func(*MatchContext)) bool) bool {
	compiledRegexps[s] = &Regexp{s, &compiledExecer{f, h, m}}
	return true
}

type compiledExecer struct {
	fun      func(int, uintptr, int, func(*MatchContext)) bool
	headOnly bool
	minReq   int
}

func (exe *compiledExecer) exec(str string, pos int, onSuccess func(*MatchContext)) bool {
	headOnly := exe.headOnly
	minReq := exe.minReq
	if headOnly && pos != 0 {
		return false
	}
	if minReq > len(str)-pos {
		return false
	}
	ctx := MatchContext{nil, str, ContextKey{'c', 0}, pos}
	if exe.fun(0, uintptr(unsafe.Pointer(&ctx)), pos, onSuccess) {
		return true
	}
	if headOnly {
		return false
	}
	for i := pos + 1; minReq <= len(str)-i; i++ {
		ctx := MatchContext{nil, str, ContextKey{'c', 0}, i}
		if exe.fun(0, uintptr(unsafe.Pointer(&ctx)), i, onSuccess) {
			return true
		}
	}
	return false
}
