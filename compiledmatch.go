package yarex

import "sync"

// IntStackPool is accessed by compiled matchers to reuse int stacks.
// Do not use this for any other purposes.
var IntStackPool = sync.Pool{
	New: func() interface{} {
		b := make([]int, 256)
		return &b
	},
}

var compiledRegexps = map[string]*Regexp{}

func RegisterCompiledRegexp(s string, h bool, m int, f func(int, MatchContext, int, func(MatchContext)) bool) bool {
	compiledRegexps[s] = &Regexp{s, &compiledExecer{f, h, m}}
	return true
}

type compiledExecer struct {
	fun      func(int, MatchContext, int, func(MatchContext)) bool
	headOnly bool
	minReq   int
}

func (exe *compiledExecer) exec(str string, pos int, onSuccess func(MatchContext)) bool {
	headOnly := exe.headOnly
	minReq := exe.minReq
	if headOnly && pos != 0 {
		return false
	}
	if minReq > len(str)-pos {
		return false
	}
	stack := *(opStackPool.Get().(*[]opStackFrame))
	defer func() { opStackPool.Put(&stack) }()
	getter := func() []opStackFrame { return stack }
	setter := func(s []opStackFrame) { stack = s }
	ctx0 := makeOpMatchContext(str, getter, setter)
	if exe.fun(0, ctx0.Push(ContextKey{'c', 0}, pos), pos, onSuccess) {
		return true
	}
	if headOnly {
		return false
	}
	for i := pos + 1; minReq <= len(str)-i; i++ {
		if exe.fun(0, ctx0.Push(ContextKey{'c', 0}, i), i, onSuccess) {
			return true
		}
	}
	return false
}
