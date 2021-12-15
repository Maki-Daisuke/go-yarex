package yarex

// matchContext forms linked-list using link to its parent node.
// It holds an index number of a capturing group and a position in a string
// being matched.
// The two positions with the identical index represents the end and the start
// position of the string captured by the index.
// For example, when the following data eixsts:
//
//   +-------------+    +-------------+    +-------------+    +---------------+
//   | position: +-+--->| position: +-+--->| position: +-+--->| position: nil |
//   | index   : 0 |    | index   : 1 |    | index   : 1 |    | index   :   0 |
//   | pos     : 9 |    | pos     : 5 |    | pos     : 2 |    | pos     :   1 |
//   +-------------+    +-------------+    +-------------+    +---------------+
//
// This means, captured string by the first "()" is indexed as str[2:5] and
// whole matched string is indexed as str[1:9].

const initialStackSize = 128

type stackFrame struct {
	index uint
	pos   int
}

type matchContext struct {
	str      string       // string being matched
	stack    []stackFrame // stack to record capturing position, etc.
	stackTop int          // stack top
}

func makeContext(s string) matchContext {
	return matchContext{s, make([]stackFrame, initialStackSize, initialStackSize), 0}
}

func (c matchContext) push(i uint, p int) matchContext {
	st := c.stack
	sf := stackFrame{i, p}
	if len(st) <= c.stackTop {
		st = append(st, sf)
	} else {
		st[c.stackTop] = sf
	}
	return matchContext{c.str, st, c.stackTop + 1}
}

// GetOffset returns (-1, -1) when it cannot find specified index.
func (c matchContext) GetOffset(idx uint) (start int, end int) {
	st := c.stack
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
	return c.str[start:end], true
}
