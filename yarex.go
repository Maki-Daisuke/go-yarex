package yarex

type Regexp struct {
	str string
	op  OpTree
}

func Compile(ptn string) (*Regexp, error) {
	ast, err := parse(ptn)
	if err != nil {
		return nil, err
	}
	ast = optimizeAst(ast)
	return &Regexp{
		str: ptn,
		op:  opCompile(ast),
	}, nil
}

func MustCompile(ptn string) *Regexp {
	r, err := Compile(ptn)
	if err != nil {
		panic(err)
	}
	return r
}

func (re Regexp) String() string {
	return re.str
}

func (re Regexp) MatchString(s string) bool {
	return matchOpTree(re.op, s)
}
