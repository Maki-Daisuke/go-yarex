package yarex

import "unsafe"

const initialStackSize = 128

type stackFrame struct {
	index uint
	pos   int
}

type matchContext struct {
	str      uintptr // *string              // string being matched
	getStack uintptr // *func() []stackFrame // Accessors to stack to record capturing positions.
	setStack uintptr // *func([]stackFrame)  // We use uintptr to avoid leaking param.
	stackTop int     // stack top
}

func makeContext(str *string, getter *func() []stackFrame, setter *func([]stackFrame)) matchContext {
	return matchContext{uintptr(unsafe.Pointer(str)), uintptr(unsafe.Pointer(getter)), uintptr(unsafe.Pointer(setter)), 0}
}

func (c matchContext) push(i uint, p int) matchContext {
	st := (*(*func() []stackFrame)(unsafe.Pointer(c.getStack)))() // == c.getStack()
	sf := stackFrame{i, p}
	if len(st) <= c.stackTop {
		st = append(st, sf)
		(*(*func([]stackFrame))(unsafe.Pointer(c.setStack)))(st) // == c.setStack(st)
	} else {
		st[c.stackTop] = sf
	}
	c.stackTop++
	return c
}

// GetOffset returns (-1, -1) when it cannot find specified index.
func (c matchContext) GetOffset(idx uint) (start int, end int) {
	st := (*(*func() []stackFrame)(unsafe.Pointer(c.getStack)))() // == c.getStack()
	i := c.stackTop - 1
	for ; ; i-- {
		if i == 0 {
			return -1, -1
		}
		if st[i].index == idx {
			end = st[i].pos
			break
		}
	}
	i--
	for ; i >= 0; i-- {
		if st[i].index == idx {
			start = st[i].pos
			return
		}
	}
	// This should not happen.
	panic("Undetermined capture")
}

func (c matchContext) GetCaptured(i uint) (string, bool) {
	start, end := c.GetOffset(i)
	if start < 0 {
		return "", false
	}
	str := *(*string)(unsafe.Pointer(c.str))
	return str[start:end], true
}
