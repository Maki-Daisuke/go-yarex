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
}

func (c *matchContext) with(i uint, p int) matchContext {
	return matchContext{c, i, p}
}

func (c *matchContext) get(i uint) []int {
	p := make([]int, 2, 2)
	for ; ; c = c.parent {
		if c == nil {
			return nil
		}
		if c.index == i {
			p[1] = c.pos
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
			p[0] = c.pos
			return p
		}
	}
}
