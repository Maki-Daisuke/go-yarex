package reaot

type Regexp interface {
	//Compile()
	String() string
	match(matchContext, string, int, func(matchContext, int) *matchContext) *matchContext
}
