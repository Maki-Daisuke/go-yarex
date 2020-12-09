package reaot

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
	key contextKey
}

type OpAltZeroWidthCheck struct {
	OpBase
	alt OpTree
	key contextKey
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
	key contextKey
}

type OpCaptureEnd struct { // actuallly, this is identical to OpCaptureStart
	OpBase
	key contextKey
}

type OpBackRef struct {
	OpBase
	key contextKey
}

type OpAssertBegin struct {
	OpBase
}

type OpAssertEnd struct {
	OpBase
}
