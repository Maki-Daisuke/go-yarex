package yarex

type execer interface {
	exec(str string, pos int, onSuccess func(MatchContext)) bool
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
	return re.exe.exec(s, 0, func(_ MatchContext) {})
}

func (re Regexp) FindString(s string) string {
	matched := ""
	re.exe.exec(s, 0, func(c MatchContext) {
		matched, _ = c.GetCaptured(ContextKey{'c', 0})
	})
	return matched
}

func (re Regexp) FindStringIndex(s string) (loc []int) {
	re.exe.exec(s, 0, func(c MatchContext) {
		loc = c.GetCapturedIndex(ContextKey{'c', 0})
	})
	return loc
}
