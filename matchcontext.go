package reaot

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
type matchContext struct {
	parent *matchContext
	index  uint
	pos    int
	str    string
}

func (c *matchContext) with(i uint, p int) *matchContext {
	return &matchContext{c, i, p, c.str}
}

// GetOffset returns (-1, -1) when it cannot find specified index.
func (c *matchContext) GetOffset(i uint) (start int, end int) {
	for ; ; c = c.parent {
		if c == nil {
			return -1, -1
		}
		if c.index == i {
			end = c.pos
			break
		}
	}
	c = c.parent
	for ; ; c = c.parent {
		if c == nil {
			// This should not happen.
			panic("Undetermined capture")
		}
		if c.index == i {
			start = c.pos
			return
		}
	}
}

func (c *matchContext) GetCaptured(i uint) (string, bool) {
	start, end := c.GetOffset(i)
	if start < 0 {
		return "", false
	}
	return c.str[start:end], true
}
