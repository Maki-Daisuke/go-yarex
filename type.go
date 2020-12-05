package reaot

type Continuation = func(matchContext, int) *matchContext

type Regexp interface {
	//Compile()
	String() string
	match(matchContext, int, Continuation) *matchContext
}
