package yarex

type ContextKey struct {
	Kind  rune
	Index uint
}

// MatchContext forms linked-list using reference to its parent node.
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
type MatchContext struct {
	Parent *MatchContext
	Str    string
	Key    ContextKey
	Pos    int
}

func (c *MatchContext) With(k ContextKey, p int) MatchContext {
	return MatchContext{
		Parent: c,
		Str:    c.Str,
		Key:    k,
		Pos:    p,
	}
}

func (c *MatchContext) GetCaptured(k ContextKey) (string, bool) {
	var start, end int
	for ; ; c = c.Parent {
		if c == nil {
			return "", false
		}
		if c.Key == k {
			end = c.Pos
			break
		}
	}
	c = c.Parent
	for ; ; c = c.Parent {
		if c == nil {
			// This should not happen.
			panic("Undetermined capture")
		}
		if c.Key == k {
			start = c.Pos
			return c.Str[start:end], true
		}
	}
}

func (c *MatchContext) FindVal(k ContextKey) int {
	for ; ; c = c.Parent {
		if c == nil {
			return -1
		}
		if c.Key == k {
			return c.Pos
		}
	}
}
