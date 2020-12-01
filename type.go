package reaot

type Regexp interface {
	//Compile()
	String() string
	match(matchContext, []rune, int, func(matchContext, int) *matchContext) *matchContext
}
