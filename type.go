package reaot

type Regexp interface {
	//Compile()
	String() string
	match(matchContext, int, func(matchContext, int)) // This throws *matchContext via panic when it successfully match.
}
