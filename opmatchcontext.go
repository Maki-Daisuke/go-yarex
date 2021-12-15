package yarex

type ContextKey struct {
	Kind  rune
	Index uint
}

type opStackFrame struct {
	Key ContextKey
	Pos int
}

type MatchContext struct {
	Str      string
	getStack func() []opStackFrame
	setStack func([]opStackFrame)
	stackTop int
}

func makeOpMatchContext(str string, getter func() []opStackFrame, setter func([]opStackFrame)) MatchContext {
	return MatchContext{str, getter, setter, 0}
}

func (c MatchContext) Push(k ContextKey, p int) MatchContext {
	st := c.getStack()
	sf := opStackFrame{k, p}
	if len(st) <= c.stackTop {
		st = append(st, sf)
		c.setStack(st)
	} else {
		st[c.stackTop] = sf
	}
	c.stackTop++
	return c
}

func (c MatchContext) GetCaptured(k ContextKey) (string, bool) {
	var start, end int
	st := c.getStack()
	i := c.stackTop - 1
	for ; ; i-- {
		if i == 0 {
			return "", false
		}
		if st[i].Key == k {
			end = st[i].Pos
			break
		}
	}
	i--
	for ; i >= 0; i-- {
		if st[i].Key == k {
			start = st[i].Pos
			return c.Str[start:end], true
		}
	}
	// This should not happen.
	panic("Undetermined capture")
}

func (c MatchContext) FindVal(k ContextKey) int {
	st := c.getStack()
	for i := c.stackTop - 1; i >= 0; i-- {
		if st[i].Key == k {
			return st[i].Pos
		}
	}
	return -1
}
