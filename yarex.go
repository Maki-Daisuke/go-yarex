package yarex

type execer interface {
	exec(str string, pos int, onSuccess func(*MatchContext)) bool
}

type Regexp struct {
	str string
	exe execer
}

func Compile(ptn string) (*Regexp, error) {
	if r, ok := compiledRegexps[ptn]; ok {
		return r, nil
	}
	ast, err := parse(ptn)
	if err != nil {
		return nil, err
	}
	ast = optimizeAst(ast)
	op := opCompile(ast)
	return &Regexp{ptn, opExecer{op}}, nil
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
	return re.exe.exec(s, 0, func(_ *MatchContext) {})
}
