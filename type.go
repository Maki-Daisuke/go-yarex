package reaot

type Regexp interface {
	//Compile()
	String() string
	Match(string) (string, bool)
}

