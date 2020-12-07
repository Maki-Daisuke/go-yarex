package reaot

type OpTree interface {
	minimumReq() int
	match(uintptr, int) *opMatchContext // Here, we use uintptr to pass *opMatchContext to avpid heap-allocation
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
	key string
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
	key string
}

type OpCaptureEnd struct { // actuallly, this is identical to OpCaptureStart
	OpBase
	key string
}

type OpBackRef struct {
	OpBase
	key string
}

type OpAssertBegin struct {
	OpBase
}

type OpAssertEnd struct {
	OpBase
}
