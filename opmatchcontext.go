package reaot

// opMatchContext forms linked-list using reference to its parent node.
// It holds an index number of a capturing group and a position in a string
// being matched.
// The two positions with the identical index represents the end and the start
// position of the string captured by the index.
// For example, when the following data eixsts:
//
//   +-----------+    +-----------+    +-----------+    +-------------+
//   | parent: +-+--->| parent: +-+--->| parent: +-+--->| parent: nil |
//   | capIdx: 0 |    | capIdx: 1 |    | capIdx: 1 |    | capIdx:   0 |
//   | capPos: 9 |    | capPos: 5 |    | capPos: 2 |    | capPos:   1 |
//   +-----------+    +-----------+    +-----------+    +-------------+
//
// This means, captured string by the first "()" is indexed as str[2:5] and
// whole matched string is indexed as str[1:9].
type opMatchContext struct {
	parent *opMatchContext
	str    string
	capIdx uint // Capture index
	capPos int  // Capture position
	repIdx uint // Repeat index for zero-width-check index
	repPos int  // Repeat position for repIndex
}

func (c *opMatchContext) withCap(i uint, p int) *opMatchContext {
	new := new(opMatchContext)
	*new = *c
	new.parent = c
	new.capIdx = i
	new.capPos = p
	return new
}

func (c *opMatchContext) GetCaptured(i uint) (string, bool) {
	var start, end int
	for ; ; c = c.parent {
		if c == nil {
			return "", false
		}
		if c.capIdx == i {
			end = c.capPos
			break
		}
	}
	c = c.parent
	for ; ; c = c.parent {
		if c == nil {
			// This should not happen.
			panic("Undetermined capture")
		}
		if c.capIdx == i {
			start = c.capPos
			return c.str[start:end], true
		}
	}
}

func (c *opMatchContext) withRep(i uint, p int) *opMatchContext {
	new := new(opMatchContext)
	*new = *c
	new.parent = c
	new.repIdx = i
	new.repPos = p
	return new
}

func (c *opMatchContext) findRepeatStart(i uint) int {
	for ; ; c = c.parent {
		if c == nil {
			return -1
		}
		if c.repIdx == i {
			return c.repPos
		}
	}
}
