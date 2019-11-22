package reaot

type Regexp interface {
	//Compile()
	String() string
	match(string, int, func(int) bool) bool
}
