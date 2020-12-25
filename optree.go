package yarex

type OpTree interface {
	minimumReq() int
}

type OpBase struct {
	minReq   int
	follower OpTree
}

func (op OpBase) minimumReq() int {
	return op.minReq
}

type OpSuccess struct{}

func (_ OpSuccess) minimumReq() int {
	return 0
}

type OpStr struct {
	OpBase
	str string
}

type OpAlt struct {
	OpBase
	alt OpTree
}

type OpRepeat struct {
	OpBase
	alt OpTree
	key ContextKey
}

type OpClass struct {
	OpBase
	cls CharClass
}

type OpNotNewLine struct {
	OpBase
}

type OpCaptureStart struct {
	OpBase
	key ContextKey
}

type OpCaptureEnd struct { // actuallly, this is identical to OpCaptureStart
	OpBase
	key ContextKey
}

type OpBackRef struct {
	OpBase
	key ContextKey
}

type OpAssertBegin struct {
	OpBase
}

type OpAssertEnd struct {
	OpBase
}
