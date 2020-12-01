package reaot

type Regexp interface {
	//Compile()
	String() string
	match(matchContext, int, func(matchContext, int) *matchContext) *matchContext
}
