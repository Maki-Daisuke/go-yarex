package reaot

type Regexp interface {
	//Compile()
	String() string
	match(string, func(string) bool) bool
}
