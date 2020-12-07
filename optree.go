package reaot

type OpTree interface {
	minimumReq() int
	match(*matchContext, int) *matchContext
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

type OpClass struct {
	OpBase
	cls CharClass
}

type OpNotNewLine struct {
	OpBase
}

type OpCaptureStart struct {
	OpBase
	index uint
}

type OpCaptureEnd struct { // actuallly, this is identical to OpCaptureStart
	OpBase
	index uint
}

type OpBackRef struct {
	OpBase
	index uint
}

type OpAssertBegin struct {
	OpBase
}

type OpAssertEnd struct {
	OpBase
}
