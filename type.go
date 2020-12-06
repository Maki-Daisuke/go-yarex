package reaot

// Here, we use uintpointer to pass *matchContext
// to avoid from allocating the parameter in heap
type Continuation = func(uintptr, int) *matchContext

type Regexp interface {
	//Compile()
	String() string
	match(uintptr, int, Continuation) *matchContext
}
