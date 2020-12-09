package reaot

// opMatchContext forms linked-list using reference to its parent node.
// It holds a key and an integer value, which are used to represent both capturing groups
// and start position of repeat for checking zero-width matcing.
// A pair integers with the identical "cN" key represents the end and the start
// position of the string captured by the key.
// For example, when the following data eixsts:
//
//   +-----------+    +-----------+    +-----------+    +-----------+    +-------------+
//   | parent: +-+--->| parent: +-+--->| parent: +-+--->| parent: +-+--->| parent: nil |
//   | key: "c0" |    | key: "c1" |    | key: "r1" |    | key: "c1" |    | key:   "c0" |
//   | pos: 9    |    | pos: 5    |    | pos: 2    |    | pos: 2    |    | pos:   1    |
//   +-----------+    +-----------+    +-----------+    +-----------+    +-------------+
//
// This means, captured string by the first "()" is indexed as str[2:5] and
// whole matched string is indexed as str[1:9].
type contextKey struct {
	kind  rune
	index uint
}

type opMatchContext struct {
	parent *opMatchContext
	str    string
	key    contextKey
	pos    int
}

func (c *opMatchContext) with(k contextKey, p int) opMatchContext {
	return opMatchContext{
		parent: c,
		str:    c.str,
		key:    k,
		pos:    p,
	}
}

func (c *opMatchContext) GetCaptured(k contextKey) (string, bool) {
	var start, end int
	for ; ; c = c.parent {
		if c == nil {
			return "", false
		}
		if c.key == k {
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
		if c.key == k {
			start = c.pos
			return c.str[start:end], true
		}
	}
}

func (c *opMatchContext) findVal(k contextKey) int {
	for ; ; c = c.parent {
		if c == nil {
			return -1
		}
		if c.key == k {
			return c.pos
		}
	}
}
